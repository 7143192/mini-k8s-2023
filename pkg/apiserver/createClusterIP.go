package apiserver

import (
	"bytes"
	"encoding/json"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v3"
	"log"
	"mini-k8s/pkg/config"
	"mini-k8s/pkg/defines"
	"mini-k8s/pkg/etcd"
	"net/http"
	"time"
)

func CreateClusterIPSvc(cli *clientv3.Client, cipSvc *defines.ClusterIPService) *defines.EtcdService {
	// check service with duplicate name here.
	originList := make([]string, 0)
	oldListKey := defines.AllServiceSetPrefix + "/"
	oldKv := etcd.Get(cli, oldListKey).Kvs
	if len(oldKv) != 0 {
		_ = yaml.Unmarshal(oldKv[0].Value, &originList)
		for _, item := range originList {
			if item == (defines.ServicePrefix + "/" + cipSvc.Metadata.Name) {
				log.Printf("[apiserver] the service %v has already been creted!\n", cipSvc.Metadata.Name)
				oldSvcKey := item
				oldSvcKv := etcd.Get(cli, oldSvcKey).Kvs
				tmp := &defines.EtcdService{}
				_ = yaml.Unmarshal(oldSvcKv[0].Value, tmp)
				// if already been created, return the old one.
				return tmp
			}
		}
	}
	// first create an etcd service object
	res := &defines.EtcdService{}
	res.SvcName = cipSvc.Metadata.Name
	res.SvcType = "ClusterIP"
	res.SvcClusterIP = cipSvc.Spec.ClusterIP
	res.SvcSelector = cipSvc.Spec.Selector
	res.SvcNodePort = 0 // 0 means the type is ClusterIP.
	res.SvcPorts = cipSvc.Spec.Ports
	// get all pods' info that satisfy the service's pod selector.
	ServiceSelectPods(cli, res)
	res.SvcStartTime = time.Now()
	// then store the new object into etcd.
	svcKey := defines.ServicePrefix + "/" + res.SvcName
	resByte, _ := yaml.Marshal(res)
	etcd.Put(cli, svcKey, string(resByte))
	// then store the new service list.(to watch changes)
	svcListKey := defines.AllServiceSetPrefix + "/"
	oldList := make([]string, 0)
	kv := etcd.Get(cli, svcListKey).Kvs
	if len(kv) == 0 {
		// the first service case.
		newList := append(oldList, svcKey)
		newListByte, _ := yaml.Marshal(&newList)
		etcd.Put(cli, svcListKey, string(newListByte))
	} else {
		// normal case.
		_ = yaml.Unmarshal(kv[0].Value, &oldList)
		newList := append(oldList, svcKey)
		newListByte, _ := yaml.Marshal(&newList)
		etcd.Put(cli, svcListKey, string(newListByte))
	}
	// then do some other jobs according to Kube proxy(TODO: not sure what to do for now)
	_ = SendOutCreateService(cli, res)
	return res
}

func SendOutCreateService(cli *clientv3.Client, service *defines.EtcdService) error {
	prefixKey := defines.NodePrefix + "/"
	kvs := etcd.GetWithPrefix(cli, prefixKey).Kvs
	nodes := make([]*defines.NodeInfo, 0)
	if len(kvs) != 0 {
		for _, kv := range kvs {
			oneNode := &defines.NodeInfo{}
			err0 := yaml.Unmarshal(kv.Value, oneNode)
			if err0 != nil {
				log.Printf("[apiserver] an error occurs when unmarshal one node in SendOutCreateService func: %v\n", err0)
				return nil
			}
			nodes = append(nodes, oneNode)
		}
	}
	masterNode := &defines.NodeInfo{}
	masterNode.NodeData.NodeSpec.Metadata.Ip = "192.168.1.8"
	nodes = append(nodes, masterNode)
	for _, nodeInfo := range nodes {
		svcInfo := service
		body, err := json.Marshal(svcInfo)
		if err != nil {
			return err
		}
		nodeIP := nodeInfo.NodeData.NodeSpec.Metadata.Ip
		url := ""
		if svcInfo.SvcType == "ClusterIP" {
			url = "http://" + nodeIP + ":" + config.ProxyPort + "/objectAPI/createClusterIPService"
		}
		if svcInfo.SvcType == "NodePort" {
			url = "http://" + nodeIP + ":" + config.ProxyPort + "/objectAPI/createNodePortService"
		}
		log.Printf("[apiserver] srvice add pod url = %v\n", url)
		request, err := http.NewRequest("POST", url, bytes.NewReader(body))
		if err != nil {
			return err
		}
		request.Header.Add("Content-Type", "application/json")
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			return err
		}
		if response.StatusCode == http.StatusOK {
			log.Println("[apiserver] service create successfully!")
		} else {
			log.Println("[apiserver] service create fails!")
		}
	}

	//nodeIdMap := make(map[string][]*defines.Pod)
	//for _, pod := range service.SvcPods {
	//	nodeId := pod.NodeId
	//	_, exist := nodeIdMap[nodeId]
	//	if !exist {
	//		nodeIdMap[nodeId] = []*defines.Pod{}
	//		nodeIdMap[nodeId] = append(nodeIdMap[nodeId], pod)
	//	} else {
	//		nodeIdMap[nodeId] = append(nodeIdMap[nodeId], pod)
	//	}
	//}
	//for nodeId, pods := range nodeIdMap {
	//	svcInfoForNode := service
	//	svcInfoForNode.SvcPods = pods
	//	body, err := json.Marshal(svcInfoForNode)
	//	if err != nil {
	//		return err
	//	}
	//	nodeKey := defines.NodePrefix + "/" + nodeId
	//	kv := etcd.Get(cli, nodeKey).Kvs
	//	nodeInfo := &defines.NodeInfo{}
	//	_ = yaml.Unmarshal(kv[0].Value, nodeInfo)
	//	// get nodeIP here to send out request.
	//	nodeIP := nodeInfo.NodeData.NodeSpec.Metadata.Ip
	//
	//	url := ""
	//	if service.SvcType == "ClusterIP" {
	//		url = "http://" + nodeIP + ":" + config.ProxyPort + "/objectAPI/createClusterIPService"
	//	}
	//	if service.SvcType == "NodePort" {
	//		url = "http://" + nodeIP + ":" + config.ProxyPort + "/objectAPI/createNodePortService"
	//	}
	//	fmt.Printf("srvice add pod url = %v\n", url)
	//	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	//	if err != nil {
	//		return err
	//	}
	//	request.Header.Add("Content-Type", "application/json")
	//	response, err := http.DefaultClient.Do(request)
	//	if err != nil {
	//		return err
	//	}
	//	if response.StatusCode == http.StatusOK {
	//		fmt.Println("service create successfully!")
	//	} else {
	//		fmt.Println("service create fails!")
	//	}
	//}

	return nil
}

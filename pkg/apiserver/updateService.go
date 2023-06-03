package apiserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v3"
	"log"
	"mini-k8s/pkg/config"
	"mini-k8s/pkg/defines"
	"mini-k8s/pkg/etcd"
	"net/http"
)

func SendOutServiceAddPod(cli *clientv3.Client, pod *defines.Pod, svc *defines.EtcdService) error {
	svcPods := []*defines.Pod{pod}
	svcInfo := svc
	svcInfo.SvcPods = svcPods
	body, err := json.Marshal(svcInfo)
	if err != nil {
		return err
	}
	prefixKey := defines.NodePrefix + "/"
	kvs := etcd.GetWithPrefix(cli, prefixKey).Kvs
	nodes := make([]*defines.NodeInfo, 0)
	if len(kvs) != 0 {
		for _, kv := range kvs {
			oneNode := &defines.NodeInfo{}
			err0 := yaml.Unmarshal(kv.Value, oneNode)
			if err0 != nil {
				log.Printf("an error occurs when unmarshal one node in SendOutCreateService func: %v\n", err0)
				return nil
			}
			nodes = append(nodes, oneNode)
		}
	}
	masterNode := &defines.NodeInfo{}
	masterNode.NodeData.NodeSpec.Metadata.Ip = "192.168.1.8"
	nodes = append(nodes, masterNode)
	for _, nodeInfo := range nodes {
		url := ""
		nodeIP := nodeInfo.NodeData.NodeSpec.Metadata.Ip
		if svcInfo.SvcType == "ClusterIP" {
			url = "http://" + nodeIP + ":" + config.ProxyPort + "/objectAPI/addClusterIPRule"
		}
		if svcInfo.SvcType == "NodePort" {
			url = "http://" + nodeIP + ":" + config.ProxyPort + "/objectAPI/addNodePortRule"
		}
		log.Printf("srvice add pod url = %v\n", url)
		request, err := http.NewRequest("POST", url, bytes.NewReader(body))
		if err != nil {
			return err
		}
		request.Header.Add("Content-Type", "application/json")
		response, err2 := http.DefaultClient.Do(request)
		if err2 != nil {
			return err
		}
		if response.StatusCode == http.StatusOK {
			log.Println("service add new pod successfully!")
		} else {
			log.Println("service add pod fails!")
		}
	}

	return nil
}

func SendOutServiceDelPod(cli *clientv3.Client, pod *defines.Pod, svc *defines.EtcdService) error {
	svcPods := []*defines.Pod{pod}
	svcInfo := svc
	svcInfo.SvcPods = svcPods
	body, err := json.Marshal(svcInfo)
	if err != nil {
		return err
	}
	prefixKey := defines.NodePrefix + "/"
	kvs := etcd.GetWithPrefix(cli, prefixKey).Kvs
	nodes := make([]*defines.NodeInfo, 0)
	if len(kvs) != 0 {
		for _, kv := range kvs {
			oneNode := &defines.NodeInfo{}
			err0 := yaml.Unmarshal(kv.Value, oneNode)
			if err0 != nil {
				fmt.Printf("an error occurs when unmarshal one node in SendOutCreateService func: %v\n", err0)
				return nil
			}
			nodes = append(nodes, oneNode)
		}
	}
	masterNode := &defines.NodeInfo{}
	masterNode.NodeData.NodeSpec.Metadata.Ip = "192.168.1.8"
	nodes = append(nodes, masterNode)
	for _, nodeInfo := range nodes {
		// get nodeIP here to send out request.
		nodeIP := nodeInfo.NodeData.NodeSpec.Metadata.Ip
		url := ""
		if svcInfo.SvcType == "ClusterIP" {
			url = "http://" + nodeIP + ":" + config.ProxyPort + "/objectAPI/delClusterIPRule"
		}
		if svcInfo.SvcType == "NodePort" {
			url = "http://" + nodeIP + ":" + config.ProxyPort + "/objectAPI/delNodePortRule"
		}
		fmt.Printf("srvice del pod url = %v\n", url)
		request, err1 := http.NewRequest("POST", url, bytes.NewReader(body))
		if err1 != nil {
			return err
		}
		request.Header.Add("Content-Type", "application/json")
		response, err2 := http.DefaultClient.Do(request)
		if err2 != nil {
			return err
		}
		if response.StatusCode == http.StatusOK {
			fmt.Println("service del old pod successfully!")
		} else {
			fmt.Println("service del old pod fails!")
		}
	}
	return nil
}

func CheckPodMatchService(svc *defines.EtcdService, pod *defines.Pod) bool {
	selector := svc.SvcSelector
	app := selector.App
	env := selector.Env
	if app == "" && env == "" {
		return true
	}
	if app == "" && env != "" {
		if pod.Metadata.Label.Env == env {
			return true
		}
	}
	if app != "" && env == "" {
		if pod.Metadata.Label.App == app {
			return true
		}
	}
	if app != "" && env != "" {
		if pod.Metadata.Label.Env == env && pod.Metadata.Label.App == app {
			return true
		}
	}
	return false
}

//func DiffTwoPodsList(oldList []*defines.Pod, newList []*defines.Pod) bool {
//	if len(oldList) != len(newList) {
//		return true
//	}
//	for _, oldItem := range oldList {
//		exist := false
//		for _, newItem := range newList {
//			if newItem.Metadata.Name == oldItem.Metadata.Name {
//				exist = true
//				break
//			}
//		}
//		if exist == false {
//			return true
//		}
//	}
//	for _, newItem := range newList {
//		exist := false
//		for _, oldItem := range oldList {
//			if oldItem.Metadata.Name == newItem.Metadata.Name {
//				exist = true
//				break
//			}
//		}
//		if exist == false {
//			return true
//		}
//	}
//	return false
//}
//
//// UpdateService used to check whether service(s) is(are) required to be updated after a pod is added/ deleted.
//func UpdateService(cli *clientv3.Client) {
//	prefixKey := defines.ServicePrefix + "/"
//	allServices := make([]*defines.EtcdService, 0)
//	kvs := etcd.GetWithPrefix(cli, prefixKey).Kvs
//	// no service for now.
//	if len(kvs) == 0 {
//		fmt.Println("No service in the current system, so no need to update service!")
//		return
//	}
//	// get all services first.
//	for _, kv := range kvs {
//		tmp := &defines.EtcdService{}
//		_ = yaml.Unmarshal(kv.Value, tmp)
//		allServices = append(allServices, tmp)
//	}
//	for _, svc := range allServices {
//		updated := false
//		oldPodsList := svc.SvcPods
//		// then re-select pods for every service.
//		// NOTE: as this op will check all podInstances in the ETCD, so this update func should be called after the pod instance has been updated into ETCD.
//		ServiceSelectPods(cli, svc)
//		newPodsList := svc.SvcPods
//		updated = DiffTwoPodsList(oldPodsList, newPodsList)
//		if updated == true {
//			fmt.Printf("service %v is updated!\n", svc.SvcName)
//			// if this svc is updated, update its corresponding info in etcd.
//			key := defines.ServicePrefix + "/" + svc.SvcName
//			svcByte, _ := yaml.Marshal(svc)
//			etcd.Put(cli, key, string(svcByte))
//			// TODO: if this service is updated, should do some works related to kube-proxy(not sure for now)
//
//		}
//	}
//	return
//}

func CheckPodAddInService(cli *clientv3.Client, pod *defines.Pod) {
	prefixKey := defines.ServicePrefix + "/"
	allServices := make([]*defines.EtcdService, 0)
	kvs := etcd.GetWithPrefix(cli, prefixKey).Kvs
	// no service for now.
	if len(kvs) == 0 {
		fmt.Println("No service in the current system, so no need to add pod to service!")
		return
	}
	// get all services first.
	for _, kv := range kvs {
		tmp := &defines.EtcdService{}
		_ = yaml.Unmarshal(kv.Value, tmp)
		allServices = append(allServices, tmp)
	}
	for _, svc := range allServices {
		matched := CheckPodMatchService(svc, pod)
		if matched == true {
			// if matched, add this pod to this service.
			svc.SvcPods = append(svc.SvcPods, pod)
			key := defines.ServicePrefix + "/" + svc.SvcName
			svcByte, _ := yaml.Marshal(svc)
			etcd.Put(cli, key, string(svcByte))
			// send out http request to update iptables.
			_ = SendOutServiceAddPod(cli, pod, svc)
		}
	}
}

func CheckPodDelInService(cli *clientv3.Client, pod *defines.Pod) {
	prefixKey := defines.ServicePrefix + "/"
	allServices := make([]*defines.EtcdService, 0)
	kvs := etcd.GetWithPrefix(cli, prefixKey).Kvs
	// no service for now.
	if len(kvs) == 0 {
		fmt.Println("No service in the current system, so no need to delete pod in service!")
		return
	}
	// get all services first.
	for _, kv := range kvs {
		tmp := &defines.EtcdService{}
		_ = yaml.Unmarshal(kv.Value, tmp)
		allServices = append(allServices, tmp)
	}
	for _, svc := range allServices {
		matched := CheckPodMatchService(svc, pod)
		if matched == true {
			// if matched, add this pod to this service.
			// svc.SvcPods = append(svc.SvcPods, pod)
			idx := -1
			for id, _ := range svc.SvcPods {
				if pod.PodId == svc.SvcPods[id].PodId {
					idx = id
					break
				}
			}
			if idx != -1 {
				svc.SvcPods = append(svc.SvcPods[0:idx], svc.SvcPods[idx+1:]...)
			}
			key := defines.ServicePrefix + "/" + svc.SvcName
			svcByte, _ := yaml.Marshal(svc)
			etcd.Put(cli, key, string(svcByte))
			// send out http request to update iptables.
			_ = SendOutServiceDelPod(cli, pod, svc)
		}
	}
}

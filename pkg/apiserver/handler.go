package apiserver

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v2"
	"log"
	"math"
	"mini-k8s/pkg/cadvisor"
	"mini-k8s/pkg/config"
	"mini-k8s/pkg/container"
	"mini-k8s/pkg/defines"
	"mini-k8s/pkg/etcd"
	"mini-k8s/pkg/kubeproxy"
	"mini-k8s/pkg/node"
	"mini-k8s/pkg/pod"
	"mini-k8s/utils"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func (s *Server) ServerCreatePod(context *gin.Context) {
	yamlPod := &defines.YamlPod{}
	// get yamlPod data from request body.
	err := json.NewDecoder(context.Request.Body).Decode(yamlPod)
	if err != nil {
		log.Printf("ServerCreatePod error: %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Wrong yaml file when create pod!"})
		return
	}
	// then create the new pod in the api-server layer.
	newPod := CreatePod(s.es.Client, yamlPod)
	if newPod == nil {
		// error occurs.
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Error in creating a new pod!"})
	} else {
		context.JSON(http.StatusOK, gin.H{"INFO": "The pod has been created successfully!"})
	}

	//// test
	//context.JSON(http.StatusOK, gin.H{"INFO": "The pod has been created successfully!"})
}

func (s *Server) ServerDeletePod(context *gin.Context) {
	delName := ""
	err := json.NewDecoder(context.Request.Body).Decode(&delName)
	if err != nil {
		log.Printf("ServerDeletePod error: %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "error when getting del pod name!"})
		return
	}
	err = DeletePod(s.es.Client, delName)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "error when deleting a pod!"})
	} else {
		context.JSON(http.StatusOK, gin.H{"INFO": "The pod has been deleted successfully!"})
	}
}

func (s *Server) ServerDescribePod(context *gin.Context) {
	podName := ""
	err := json.NewDecoder(context.Request.Body).Decode(&podName)
	if err != nil {
		log.Printf("ServerDescribePod: an error occurs when getting describing pod name from request: %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "an error occurs when getting describing pod name from request body"})
		return
	}
	res := DescribePod(s.es.Client, podName, false)
	// resByte, _ := json.Marshal(&res)
	test := &defines.PodResourceSend{}
	_ = json.Unmarshal([]byte(res), test)
	log.Printf("[handler] object of pod resource before sending: %v\n", test)
	context.JSON(http.StatusOK, test)
}

func (s *Server) ServerDescribeNode(context *gin.Context) {
	nodeName := ""
	err := json.NewDecoder(context.Request.Body).Decode(&nodeName)
	if err != nil {
		log.Printf("ServerDescribeNode: an error occurs when getting describing node name from request: %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "an error occurs when getting describing node name from request body"})
		return
	}
	res := DescribeNode(s.es.Client, nodeName)
	// resByte, _ := json.Marshal(&res)
	test := &defines.NodeResourceSend{}
	_ = json.Unmarshal([]byte(res), test)
	log.Printf("[handler] object of node resource before sending: %v\n", test)
	context.JSON(http.StatusOK, test)
}

func (s *Server) ServerDescribeService(context *gin.Context) {
	svcName := ""
	err := json.NewDecoder(context.Request.Body).Decode(&svcName)
	if err != nil {
		log.Printf("ServerDescribeService: an error occurs when describing service: %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "an error occurs when describing service"})
		return
	}
	res := DescribeService(s.es.Client, svcName)
	tmp := &defines.ServiceInfo{}
	_ = json.Unmarshal([]byte(res), tmp)
	context.JSON(http.StatusOK, tmp)
}

func (s *Server) ServerDescribeDNS(context *gin.Context) {
	dnsName := ""
	err := json.NewDecoder(context.Request.Body).Decode(&dnsName)
	if err != nil {
		log.Printf("ServerDescribeDNS: an error occurs when describing dns: %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "an error occurs when describing dns"})
		return
	}
	res := DescribeDNS(s.es.Client, dnsName)
	tmp := &defines.EtcdDNS{}
	_ = json.Unmarshal([]byte(res), tmp)
	context.JSON(http.StatusOK, tmp)
}

func (s *Server) ServerGetPods(context *gin.Context) {
	podName := ""
	err := json.NewDecoder(context.Request.Body).Decode(&podName)
	if err != nil {
		log.Printf("ServerGetPods: an error occurs when getting pods: %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "an error occurs when getting pods"})
		return
	}
	res := GetPod(s.es.Client)
	tmp := &defines.GetPods{}
	_ = json.Unmarshal([]byte(res), tmp)
	context.JSON(http.StatusOK, tmp)
}

func (s *Server) ServerGetNodes(context *gin.Context) {
	nodeName := ""
	err := json.NewDecoder(context.Request.Body).Decode(&nodeName)
	if err != nil {
		log.Printf("ServerGetNodes: an error occurs when getting nodes: %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "an error occurs when getting nodes"})
		return
	}
	res := GetNode(s.es.Client)
	tmp := &defines.GetNodesResource{}
	_ = json.Unmarshal([]byte(res), &tmp)
	context.JSON(http.StatusOK, tmp)
}

func (s *Server) ServerGetServices(context *gin.Context) {
	svcName := ""
	err := json.NewDecoder(context.Request.Body).Decode(&svcName)
	if err != nil {
		log.Printf("ServerGetServices: an error occurs when getting services: %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "an error occurs when getting services"})
		return
	}
	res := GetService(s.es.Client)
	tmp := &defines.ServiceInfoSend{}
	_ = json.Unmarshal([]byte(res), tmp)
	context.JSON(http.StatusOK, tmp)
}

func (s *Server) ServerGetDNSs(context *gin.Context) {
	dnsName := ""
	err := json.NewDecoder(context.Request.Body).Decode(&dnsName)
	if err != nil {
		log.Printf("ServerGetDNSs: an error occurs when getting dnss: %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "an error occurs when getting dnss"})
		return
	}
	res := GetDNS(s.es.Client)
	tmp := &defines.DNSInfoSend{}
	_ = json.Unmarshal([]byte(res), tmp)
	context.JSON(http.StatusOK, tmp)
}

//func (s *Server) ServerDescribePod(context *gin.Context) {
//	name := ""
//	err := json.NewDecoder(context.Request.Body).Decode(&name)
//	if err != nil {
//		fmt.Printf("%v\n", err)
//		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "error when getting del pod name!"})
//		return
//	}
//	DescribePod(s.es.Client, name, false)
//	context.JSON(http.StatusOK, gin.H{"INFO": "describe pod successfully!"})
//}

func (s *Server) ServerRegisterNewNode(context *gin.Context) {
	yamlNode := &defines.NodeYaml{}
	err := json.NewDecoder(context.Request.Body).Decode(yamlNode)
	if err != nil {
		log.Printf("ServerRegisterNewNode Error: %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Wrong yaml file when register node!"})
		return
	}
	nodeInfo := node.RegisterNodeToMaster(s.es.Client, yamlNode)
	// nodeInfoJson, _ := json.Marshal(nodeInfo)
	// context.JSON(http.StatusOK, gin.H{"INFO": nodeInfoJson})
	context.JSON(http.StatusOK, nodeInfo)
}

func (s *Server) ServerReceiveNodeHeartBeat(context *gin.Context) {
	heartBeat := &defines.NodeHeartBeat{}
	err := json.NewDecoder(context.Request.Body).Decode(heartBeat)
	if err != nil {
		log.Printf("ServerReceiveNodeHeartBeat error: %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Wrong data when receiving node heartbeat!"})
		return
	}
	key := defines.NodeHeartBeatPrefix + "/" + heartBeat.NodeId
	// fmt.Printf("heart beat key = %v\n", key)
	heartBeatYaml, _ := yaml.Marshal(heartBeat)
	etcd.Put(s.es.Client, key, string(heartBeatYaml))
	context.JSON(http.StatusOK, "OK")
}

func (s *Server) ServerUpdateNodeHealthState(context *gin.Context) {
	nodeInfo := &defines.NodeInfo{}
	err := json.NewDecoder(context.Request.Body).Decode(nodeInfo)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Wrong data when receiving node health state!"})
		return
	}
	key := defines.NodePrefix + "/" + nodeInfo.NodeData.NodeSpec.Metadata.Name
	nodeInfoYaml, _ := yaml.Marshal(nodeInfo)
	etcd.Put(s.es.Client, key, string(nodeInfoYaml))
	context.JSON(http.StatusOK, "OK")
}

// ServerEtcdWatcher
// in version 2.0, we may change the watcher location from kubelet to api server.(but not sure.)
// NOTE: try to remove this watcher and just handle the changes after creation.
func (s *Server) ServerEtcdWatcher() {
	serverWatchPrefix := defines.NodePodsListPrefix + "/"
	log.Println("ServerEtcdWatcher: server is ready to watch etcd changes!")
	serverWatchChan := etcd.WatchWithPrefix(s.es.Client, serverWatchPrefix)
	for resp := range serverWatchChan {
		for _, ev := range resp.Events {
			log.Printf("[handler] Type: %s Key:%s Value:%s\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
			newSet := make([]string, 0)
			err := yaml.Unmarshal(ev.Kv.Value, &newSet)
			if err != nil {
				log.Printf("ServerEtcdWatcher: an error occurs when unmarshal json in func NewEtcdWatcher: %v\n", err)
				return
			}
			// TODO: send changes to the corresponding node kubelet here.
			changeKey := string(ev.Kv.Key)
			idx := strings.Index(changeKey, "/")
			// get changed node name first.
			changeNodeName := changeKey[idx+1:]
			nodeKey := defines.NodePrefix + "/" + changeNodeName
			kv := etcd.Get(s.es.Client, nodeKey).Kvs
			nodeInfo := &defines.NodeInfo{}
			_ = yaml.Unmarshal(kv[0].Value, nodeInfo)
			// get node ip here. this IP should be configured by node YAML file.
			nodeIp := nodeInfo.NodeData.NodeSpec.Metadata.Ip
			url := "http://" + nodeIp + ":" + config.WorkerPort + "/objectAPI/changeNodePod"
			newSetJson, _ := json.Marshal(newSet)
			response, err := http.Post(url, "application/json", strings.NewReader(string(newSetJson)))
			if err != nil {
				log.Printf("[handler] error when changeNodePod in func ServerEtcdWatcher: %v\n", err)
			}

			gotData := &defines.HandlePodResult{}
			bodyReader := response.Body
			buf := new(bytes.Buffer)
			_, _ = buf.ReadFrom(bodyReader)
			_ = json.Unmarshal(buf.Bytes(), gotData)
			log.Printf("[handler] got changed pods data from response = %v\n", *gotData)
			// handle service changes here.
			for _, del := range gotData.Del {
				CheckPodDelInService(s.es.Client, del)
			}
			for _, add := range gotData.Add {
				CheckPodAddInService(s.es.Client, add)
			}
		}
	}
}

func (s *Server) KubeletServerHandlePodChange(context *gin.Context) {
	// add lock here.
	config.PodMutex.Lock()
	log.Printf("\nKubeletServerHandlePodChange: kubelet handle pod changes function holds the mutex lock!\n\n")
	newSet := make([]string, 0)
	err := json.NewDecoder(context.Request.Body).Decode(&newSet)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Wrong data when handling node pods changes!"})
		return
	}
	res := pod.HandlePodChanges(newSet, s.NodeInfo)
	context.JSON(http.StatusOK, res)
	log.Printf("\nKubeletServerHandlePodChange: kubelet handle pod changes function gives up the mutex lock!\n\n")
	config.PodMutex.Unlock()
}

func (s *Server) KubeletServerHandlePodChangeNew(context *gin.Context) {
	// add lock here.
	config.PodMutex.Lock()
	log.Printf("\nKubeletServerHandlePodChangesNew: kubelet handle pod changes function holds the mutex lock!\n\n")
	newSet := make([]string, 0)
	err := json.NewDecoder(context.Request.Body).Decode(&newSet)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Wrong data when handling node pods changes!"})
		return
	}
	res := pod.HandlePodChangesNew(newSet, s.NodeInfo)
	context.JSON(http.StatusOK, res)
	log.Printf("\nKubeletServerHandlePodChangeNew: kubelet handle pod changes function gives up the mutex lock!\n\n")
	config.PodMutex.Unlock()
}

func (s *Server) KubeletServerHandleRealPodChanges(context *gin.Context) {
	res := &defines.HandlePodResult{}
	err := json.NewDecoder(context.Request.Body).Decode(res)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "error occurs when kubeletServer handles real pod changes"})
		return
	}
	pod.HandleCurNodeRealChanges(res, s.NodeInfo)
	// send back new node info to apiServer.
	context.JSON(http.StatusOK, s.NodeInfo)
}

func (s *Server) ServerCreateClusterIPSvc(context *gin.Context) {
	cipSvc := &defines.ClusterIPService{}
	err := json.NewDecoder(context.Request.Body).Decode(cipSvc)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, "Wrong data when creating a ClusterIP service!")
		return
	}
	CreateClusterIPSvc(s.es.Client, cipSvc)
	context.JSON(http.StatusOK, "OK")
}

func (s *Server) ServerCreateNodePortSvc(context *gin.Context) {
	npSvc := &defines.NodePortService{}
	err := json.NewDecoder(context.Request.Body).Decode(npSvc)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, "Wrong data when creating a NodePort service!")
		return
	}
	CreateNodePortSvc(s.es.Client, npSvc)
	context.JSON(http.StatusOK, "OK")
}

func (s *Server) ServerCreateAutoScaler(context *gin.Context) {
	auto := &defines.AutoScaler{}
	err := json.NewDecoder(context.Request.Body).Decode(auto)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, "Wrong data when creating a new auto Scaler!")
		return
	}
	CreateAutoScaler(s.es.Client, auto)
	context.JSON(http.StatusOK, "OK")
}

func (s *Server) ServerCreateClusterIPService(context *gin.Context) {
	svcInfo := &defines.EtcdService{}
	err := json.NewDecoder(context.Request.Body).Decode(svcInfo)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, "Wrong data when adding a new pod in CIP service!")
		return
	}
	svcChainName := config.KubeSvcChainPrefix + svcInfo.SvcName
	if !kubeproxy.IsChainExist(svcChainName, "nat") {
		kubeproxy.CreateChain("nat", svcChainName)
		kubeproxy.AddSvcMatchRuleToChain(config.KubeSvcMainChainName, "nat", svcInfo.SvcClusterIP,
			strconv.Itoa(svcInfo.SvcPorts[0].Port), svcInfo.SvcPorts[0].Protocol, svcChainName)
	}
	num := len(svcInfo.SvcPods)
	probability := 1.0 / float64(num)
	probabilityStr := strconv.FormatFloat(probability, 'f', -1, 64)
	for _, svcPod := range svcInfo.SvcPods {
		podDNATChainName := config.PodDNATChainPrefix + svcPod.PodId
		kubeproxy.CreateChain("nat", podDNATChainName)
		kubeproxy.AddSvcForwardRuleToChain(svcChainName, "nat", podDNATChainName, probabilityStr)
		kubeproxy.AddSvcDNATRuleToChain(podDNATChainName, svcPod.PodIp+":"+strconv.Itoa(svcInfo.SvcPorts[0].TargetPort))
	}
	log.Printf("[handler] create new CIP service = %v\n", svcInfo)
	context.JSON(http.StatusOK, "OK")
}

func (s *Server) ServerCreateNodePortService(context *gin.Context) {
	serviceInfo := &defines.EtcdService{}
	err := json.NewDecoder(context.Request.Body).Decode(serviceInfo)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, "Wrong data when adding a new pod in CIP service!")
		return
	}

	log.Printf("[handler] create new NP service = %v\n", serviceInfo)
	context.JSON(http.StatusOK, "OK")
}

//	func (s *Server) ServerCreateActualDNS(context *gin.Context) {
//		dnsInfo := &defines.EtcdDNS{}
//		err := json.NewDecoder(context.Request.Body).Decode(dnsInfo)
//		if err != nil {
//			context.AbortWithStatusJSON(http.StatusBadRequest, "Wrong data when create new dns!")
//			return
//		}
//		nginxServerIp := dns.StartNewNginxServer(dnsInfo)
//		dns.AddHost(dnsInfo.DNSHost, nginxServerIp)
//		context.JSON(http.StatusOK, "OK")
//	}
func (s *Server) ServerAddClusterIPRule(context *gin.Context) {
	svcInfo := &defines.EtcdService{}
	err := json.NewDecoder(context.Request.Body).Decode(svcInfo)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, "Wrong data when adding a new pod in CIP service!")
		return
	}
	//fmt.Printf("pod received by CIP add = %v\n", podInfo)
	// add logic of kube-proxy here.
	svcChainName := config.KubeSvcChainPrefix + svcInfo.SvcName
	targets := kubeproxy.GetChainRuleTarget(svcChainName)
	probabilityStr := "1.0"
	if len(targets) > 0 {
		numFloat := float64(len(targets))
		probability := 1.0 / (1.0 + numFloat)
		probabilityStr = strconv.FormatFloat(probability, 'f', -1, 64)
		kubeproxy.DeleteAllRuleInChain(svcChainName)
		for _, target := range targets {
			kubeproxy.AddSvcForwardRuleToChain(svcChainName, "nat", target, probabilityStr)
		}
	}
	for _, svcPod := range svcInfo.SvcPods {
		podDNATChainName := config.PodDNATChainPrefix + svcPod.PodId
		kubeproxy.CreateChain("nat", podDNATChainName)
		kubeproxy.AddSvcForwardRuleToChain(svcChainName, "nat", podDNATChainName, probabilityStr)
		kubeproxy.AddSvcDNATRuleToChain(podDNATChainName, svcPod.PodIp+":"+strconv.Itoa(svcInfo.SvcPorts[0].TargetPort))
	}
	context.JSON(http.StatusOK, "OK")
}

func (s *Server) ServerAddNodePortRule(context *gin.Context) {
	podInfo := &defines.Pod{}
	err := json.NewDecoder(context.Request.Body).Decode(podInfo)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, "Wrong data when adding a new pod in NP service!")
		return
	}
	log.Printf("[handler] pod received by NP add = %v\n", podInfo)
	// add logic of kube-proxy here.

	context.JSON(http.StatusOK, "OK")
}

func (s *Server) ServerDelClusterIPRule(context *gin.Context) {
	svcInfo := &defines.EtcdService{}
	err := json.NewDecoder(context.Request.Body).Decode(svcInfo)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, "Wrong data when deleting a pod in CIP service!")
		return
	}
	//fmt.Printf("pod received by CIP del = %v\n", podInfo)
	// add logic of kube-proxy here.
	svcChainName := config.KubeSvcChainPrefix + svcInfo.SvcName
	targets := kubeproxy.GetChainRuleTarget(svcChainName)
	probabilityStr := "1.0"
	if len(targets) > 1 {
		numFloat := float64(len(targets))
		probability := 1.0 / (numFloat - 1.0)
		probabilityStr = strconv.FormatFloat(probability, 'f', -1, 64)
		kubeproxy.DeleteAllRuleInChain(svcChainName)
		for _, target := range targets {
			kubeproxy.AddSvcForwardRuleToChain(svcChainName, "nat", target, probabilityStr)
		}
	}
	for _, svcPod := range svcInfo.SvcPods {
		podDNATChainName := config.PodDNATChainPrefix + svcPod.PodId
		kubeproxy.DeleteSvcDNATRuleFromChain(podDNATChainName, svcPod.PodIp+":"+strconv.Itoa(svcInfo.SvcPorts[0].TargetPort))
		kubeproxy.DeleteSvcForwardRuleFromChain(svcChainName, "nat", podDNATChainName, probabilityStr)
		kubeproxy.DeleteChain("nat", podDNATChainName)
	}
	context.JSON(http.StatusOK, "OK")
}

func (s *Server) ServerDelNodePortRule(context *gin.Context) {
	svcInfo := &defines.EtcdService{}
	err := json.NewDecoder(context.Request.Body).Decode(svcInfo)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, "Wrong data when deleting a pod in NP service!")
		return
	}
	//fmt.Printf("pod received by NP del = %v\n", podInfo)
	// add logic of kube-proxy here.

	context.JSON(http.StatusOK, "OK")
}

func (s *Server) ServerGetAllNodes(context *gin.Context) {
	// fmt.Printf("get into ServerGetAllNode function!\n")
	prefixKey := defines.NodePrefix + "/"
	kvs := etcd.GetWithPrefix(s.es.Client, prefixKey).Kvs
	if len(kvs) == 0 {
		context.JSON(http.StatusOK, gin.H{"INFO": ""})
	} else {
		nodeNames := ""
		for _, kv := range kvs {
			nodeInfo := &defines.NodeInfo{}
			_ = yaml.Unmarshal(kv.Value, nodeInfo)
			// skip not-ready nodes.
			if nodeInfo.NodeData.NodeState == defines.NodeNotReady {
				continue
			}
			nodeNames = nodeNames + nodeInfo.NodeData.NodeSpec.Metadata.Name + " "
		}
		log.Printf("[handler] all node names = %v\n", nodeNames)
		context.JSON(http.StatusOK, gin.H{"INFO": nodeNames})
	}
}

func (s *Server) ServerGetUnhandledPods(context *gin.Context) {
	yamlPods := make([]*defines.YamlPod, 0)
	kvs := etcd.GetWithPrefix(s.es.Client, defines.PodInstancePrefix+"/").Kvs
	for _, kv := range kvs {
		p := &defines.Pod{}
		err := yaml.Unmarshal(kv.Value, p)
		if err != nil {
			log.Printf("ServerGetUnhandlePods error: %v\n", err)
			continue
		}
		if p.NodeId == "" {
			yamlPods = append(yamlPods, &p.YamlPod)
		}
	}

	yamlPodsInfo, err := json.Marshal(yamlPods)
	if err != nil {
		log.Printf("ServerGetUnhandlePods error: %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Something went wrong in wrapping yamlPods!"})
	} else {
		context.JSON(http.StatusOK, gin.H{"INFO": yamlPodsInfo})
	}
}

func (s *Server) ServerUpdatePodNodeId(context *gin.Context) {
	result := make(map[string]string)
	err := json.NewDecoder(context.Request.Body).Decode(&result)
	if err != nil {
		log.Printf("ServerUpdatePodId error: %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Wrong pod!"})
		return
	}
	pods := s.es.Get(defines.PodInstancePrefix + "/" + result["podName"]).Kvs
	if len(pods) == 0 {
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "The pod doesn't exist anymore!"})
		return
	}
	podInfo := &defines.Pod{}
	err = yaml.Unmarshal(pods[0].Value, podInfo)
	if err != nil {
		log.Printf("ServerUpdatePodId error: %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Error in saving pod config!"})
		return
	}
	podInfo.NodeId = result["nodeID"]
	podVal, err := yaml.Marshal(podInfo)
	if err != nil {
		log.Printf("ServerUpdatePodId error: %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Error in saving pod config!"})
		return
	}
	s.es.Put(defines.PodInstancePrefix+"/"+podInfo.Metadata.Name, string(podVal))
	nodePodListKey := defines.NodePodsListPrefix + "/" + result["nodeID"]
	kv := s.es.Get(nodePodListKey).Kvs
	if len(kv) == 0 {
		return
	}
	newList := make([]string, 0)
	_ = yaml.Unmarshal(kv[0].Value, &newList)
	newList = append(newList, podInfo.PodId)
	newListByte, _ := yaml.Marshal(&newList)
	s.es.Put(nodePodListKey, string(newListByte))
	HandlePodUpdatesForNodeNew(s.es.Client, result["NodeID"], newList)
	context.JSON(http.StatusOK, gin.H{"INFO": "The pod has been updated successfully!"})
}

// HandlePodUpdateForNode @nodeId: target node to manage this changed pod.@newSet: new pod name list after changes.
func HandlePodUpdateForNode(cli *clientv3.Client, nodeId string, newSet []string) {
	fmt.Printf("new pods set before sending HTTP: %v\n", newSet)
	nodeKey := defines.NodePrefix + "/" + nodeId
	kv := etcd.Get(cli, nodeKey).Kvs
	nodeInfo := &defines.NodeInfo{}
	_ = yaml.Unmarshal(kv[0].Value, nodeInfo)
	// get node ip here. this IP should be configured by node YAML file.
	nodeIp := nodeInfo.NodeData.NodeSpec.Metadata.Ip
	url := "http://" + nodeIp + ":" + config.WorkerPort + "/objectAPI/changeNodePod"
	newSetJson, _ := json.Marshal(&newSet)
	response, err := http.Post(url, "application/json", strings.NewReader(string(newSetJson)))
	if err != nil {
		log.Printf("HandlePodUpdateForNode error: %v\n", err)
		return
	}
	gotData := &defines.HandlePodResult{}
	bodyReader := response.Body
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(bodyReader)
	_ = json.Unmarshal(buf.Bytes(), gotData)
	log.Printf("[handler] got changed pods data from response = %v\n", *gotData)
	// handle service changes here.
	for _, del := range gotData.Del {
		CheckPodDelInService(cli, del)
	}
	for _, add := range gotData.Add {
		CheckPodAddInService(cli, add)
	}
}

// HandlePodUpdatesForNodeNew TODO: New version of function "HnaldePodUpdateForNode"
func HandlePodUpdatesForNodeNew(cli *clientv3.Client, nodeId string, newSet []string) {
	log.Printf("[handler] new pods set before sending HTTP: %v\n", newSet)
	nodeKey := defines.NodePrefix + "/" + nodeId
	kv := etcd.Get(cli, nodeKey).Kvs
	nodeInfo := &defines.NodeInfo{}
	_ = yaml.Unmarshal(kv[0].Value, nodeInfo)
	// get node ip here. this IP should be configured by node YAML file.
	nodeIp := nodeInfo.NodeData.NodeSpec.Metadata.Ip
	url := "http://" + nodeIp + ":" + config.WorkerPort + "/objectAPI/changeNodePodNew"
	newSetJson, _ := json.Marshal(&newSet)
	response, err := http.Post(url, "application/json", strings.NewReader(string(newSetJson)))
	if err != nil {
		log.Printf("HandlePodUpdateForNode error: %v\n", err)
		return
	}
	gotData := &defines.HandlePodNameResult{}
	bodyReader := response.Body
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(bodyReader)
	_ = json.Unmarshal(buf.Bytes(), gotData)
	log.Printf("[handler] got changed pods names from response = %v\n", *gotData)
	err = HandleRealPodUpdateForNode(cli, gotData, nodeIp)
	if err != nil {
		log.Printf("HandlePodUpdateForNode error: %v\n", err)
	}
	//// handle service changes here.
	//for _, del := range gotData.Del {
	//	CheckPodDelInService(cli, del)
	//}
	//for _, add := range gotData.Add {
	//	CheckPodAddInService(cli, add)
	//}
}

func HandleRealPodUpdateForNode(cli *clientv3.Client, changedNames *defines.HandlePodNameResult, nodeIp string) error {
	res := &defines.HandlePodResult{}
	res.Del = make([]*defines.Pod, 0)
	res.Add = make([]*defines.Pod, 0)
	for _, addName := range changedNames.Add {
		addPod := pod.CreateRealPod(cli, addName)
		res.Add = append(res.Add, addPod)
		// handle service changes here.
		CheckPodAddInService(cli, addPod)
	}
	for _, delName := range changedNames.Del {
		delPod := pod.DelRealPod(cli, delName)
		res.Del = append(res.Del, delPod)
		// handle service changes here.
		CheckPodDelInService(cli, delPod)
	}
	body, err := json.Marshal(res)
	if err != nil {
		return err
	}
	url := "http://" + nodeIp + ":" + config.WorkerPort + "/objectAPI/changeNodeRealPod"
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Printf("an error occurs when handling real pod changes for node!")
		return err
	}
	if response.StatusCode != http.StatusOK {
		return errors.New("error when getting response of HandleRealPodUpdateForNode")
	}
	gotData := &defines.NodeInfo{}
	bodyReader := response.Body
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(bodyReader)
	_ = json.Unmarshal(buf.Bytes(), gotData)
	log.Printf("[handler] new nodeInfo after updates is: %v\n", gotData)
	// put the new node data into etcd.
	nodeKey := defines.NodePrefix + "/" + gotData.NodeData.NodeSpec.Metadata.Name
	gotDataByte, _ := yaml.Marshal(gotData)
	etcd.Put(cli, nodeKey, string(gotDataByte))
	return nil
}

func (s *Server) ServerScheduleNodeForPod(context *gin.Context) {
	podInfo := &defines.Pod{}
	err := json.NewDecoder(context.Request.Body).Decode(podInfo)
	if err != nil {
		log.Printf("an error occurs when receiving podInfo in scheduler request handler!")
	}
	targetNodeName := s.Scheduler.ChooseNodeForPod(podInfo.Metadata.Name)
	log.Printf("[handler] targetNodeName in handler: %v\n", targetNodeName)
	context.JSON(http.StatusOK, targetNodeName)
}

func SendOutUpdateAutoController(controller *defines.AutoScalerController) {
	body, err := json.Marshal(controller)
	if err != nil {
		return
	}
	url := "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/updateEtcdAutoControllerList"
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return
	}
	request.Header.Add("Content-Type", "application/json")
	reponse, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Printf("an error occurs when updating controller list!")
		return
	}
	if reponse.StatusCode == http.StatusOK {
		log.Println("SendOutUpdateAutoController success!")
	} else {
		log.Println("SendOutUpdateAutoController error occurs!")
	}
	return
}

// ServerHandleAddAutoScaler ???
func (s *Server) ServerHandleAddAutoScaler(context *gin.Context) {
	etcdAuto := &defines.EtcdAutoScaler{}
	err := json.NewDecoder(context.Request.Body).Decode(etcdAuto)
	if err != nil {
		log.Printf("ServerHandleAddAutoScaler error: %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Error in handle add autoScaler!"})
		return
	}
	// add the auto object into etcd first.
	key := defines.AutoScalerControllerPrefix + "/"
	old := &defines.AutoScalerController{}
	kv := etcd.Get(s.es.Client, key).Kvs
	if len(kv) == 0 {
		old.AutoScalers = make([]*defines.EtcdAutoScaler, 0)
		old.AutoScalers = append(old.AutoScalers, etcdAuto)
		SendOutUpdateAutoController(old)
	} else {
		_ = yaml.Unmarshal(kv[0].Value, old)
		old.AutoScalers = append(old.AutoScalers, etcdAuto)
		SendOutUpdateAutoController(old)
	}
	// fmt.Printf("old autoScalers list = %v\n", s.AutoScalerController.AutoScalers)
	s.AutoScalerController.AutoScalers = append(s.AutoScalerController.AutoScalers, etcdAuto)
	// fmt.Printf("old autoScalers list = %v\n", s.AutoScalerController.AutoScalers)
}

func (s *Server) ServerHandleRemoveAutoScaler(context *gin.Context) {
	etcdAuto := &defines.EtcdAutoScaler{}
	err := json.NewDecoder(context.Request.Body).Decode(etcdAuto)
	if err != nil {
		log.Printf("ServerHandleRemoveAutoScaler error: %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Error in handle removing autoScaler!"})
		return
	}
	// update the auto object list info in etcd.
	key := defines.AutoScalerControllerPrefix + "/"
	old := &defines.AutoScalerController{}
	kv := etcd.Get(s.es.Client, key).Kvs
	if len(kv) == 0 {
		log.Printf("the controller info not stored in the etcd!\n")
		return
	} else {
		_ = yaml.Unmarshal(kv[0].Value, old)
		oldId := -1
		for idx, item := range old.AutoScalers {
			if item.AutoName == etcdAuto.AutoName {
				// old.AutoScalers = append(old.AutoScalers[0:idx], old.AutoScalers[idx+1:]...)
				oldId = idx
				break
			}
		}
		if oldId != -1 {
			if len(old.AutoScalers) == 1 {
				old.AutoScalers = make([]*defines.EtcdAutoScaler, 0)
			} else {
				if oldId == 0 {
					old.AutoScalers = old.AutoScalers[1:]
				} else {
					if oldId == len(old.AutoScalers)-1 {
						old.AutoScalers = old.AutoScalers[0 : len(old.AutoScalers)-1]
					} else {
						old.AutoScalers = append(old.AutoScalers[0:oldId], old.AutoScalers[oldId+1:]...)
					}
				}
			}
		}
		SendOutUpdateAutoController(old)
	}
	Id := -1
	for idx, item := range s.AutoScalerController.AutoScalers {
		if item.AutoName == etcdAuto.AutoName {
			// s.AutoScalerController.AutoScalers = append(s.AutoScalerController.AutoScalers[0:idx], s.AutoScalerController.AutoScalers[idx+1:]...)
			Id = idx
			break
		}
	}
	if Id != -1 {
		if len(s.AutoScalerController.AutoScalers) == 1 {
			s.AutoScalerController.AutoScalers = make([]*defines.EtcdAutoScaler, 0)
		} else {
			if Id == 0 {
				s.AutoScalerController.AutoScalers = s.AutoScalerController.AutoScalers[1:]
			} else {
				if Id == len(s.AutoScalerController.AutoScalers)-1 {
					s.AutoScalerController.AutoScalers = s.AutoScalerController.AutoScalers[0 : len(s.AutoScalerController.AutoScalers)-1]
				} else {
					s.AutoScalerController.AutoScalers = append(s.AutoScalerController.AutoScalers[0:Id], s.AutoScalerController.AutoScalers[Id+1:]...)
				}
			}
		}
	}
}

func (s *Server) ServerDeleteAutoScaler(context *gin.Context) {
	delName := ""
	err := json.NewDecoder(context.Request.Body).Decode(&delName)
	if err != nil {
		log.Printf("an error occurs when deleting an autoScaler!")
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Error in delete autoScaler!"})
		return
	}
	err = DeleteAutoScaler(s.es.Client, delName)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Error in handle deleting autoScaler!"})
		return
	}
	context.JSON(http.StatusOK, "OK")
}

func (s *Server) ServerStorePodResource(context *gin.Context) {
	podResource := &defines.PodResourceUsed{}
	err := json.NewDecoder(context.Request.Body).Decode(podResource)
	log.Printf("got podResourceInfo from kubelet cadvisor: %v\n", *podResource)
	if err != nil {
		log.Printf("an error occurs when storing a new podResource!")
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Error in store podResource!"})
		return
	}
	key := defines.PodResourceStatePrefix + "/" + podResource.PodName
	podResourceByte, _ := yaml.Marshal(podResource)
	etcd.Put(s.es.Client, key, string(podResourceByte))
	context.JSON(http.StatusOK, "OK")
}

func (s *Server) ServerStoreNodeResource(context *gin.Context) {
	nodeResource := &defines.NodeResourceUsed{}
	err := json.NewDecoder(context.Request.Body).Decode(nodeResource)
	log.Printf("got nodeResourceInfo from kubelet cadvisor: %v\n", *nodeResource)
	if err != nil {
		log.Printf("an error occurs when storing a new nodeResource!")
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Error in store nodeResource!"})
		return
	}
	key := defines.NodeResourcePrefix + "/" + nodeResource.NodeId
	nodeResourceByte, _ := yaml.Marshal(nodeResource)
	etcd.Put(s.es.Client, key, string(nodeResourceByte))
	context.JSON(http.StatusOK, "OK")
}

// CheckAutoPodsResource used to check all replicas pods resources usage and do pod instances shrink and extend.
func (s *Server) CheckAutoPodsResource() {
	// log.Printf("all autoScalers held by controller = %v\n", s.AutoScalerController.AutoScalers)
	for _, auto := range s.AutoScalerController.AutoScalers {
		// get all pod replicas names for one autoScaler pod.
		replicaListKey := defines.AutoPodReplicasNamesPrefix + "/" + auto.AutoPodId
		kv := etcd.Get(s.es.Client, replicaListKey).Kvs
		if len(kv) == 0 {
			continue
		}
		nameList := make([]string, 0)
		_ = yaml.Unmarshal(kv[0].Value, &nameList)
		totalMem := uint64(0)
		totalCPU := uint64(0)
		curMinMem := uint64(math.MaxInt64)
		curMinCPU := uint64(math.MaxInt64)
		minMemId := ""
		minCPUId := ""
		// then traverse every pod (through its name) to check its resource information.
		healthyNum := 0
		for _, podName := range nameList {
			podInstanceKey := defines.PodInstancePrefix + "/" + podName
			instanceKV := etcd.Get(s.es.Client, podInstanceKey).Kvs
			if len(instanceKV) > 0 {
				tmpPod := &defines.Pod{}
				_ = yaml.Unmarshal(instanceKV[0].Value, tmpPod)
				if tmpPod.PodState == defines.Failed {
					_ = DeleteAutoPodReplica(tmpPod.Metadata.Name, auto.AutoName)
					continue
				} else {
					healthyNum++
				}
			} else {
				// this pod does not exist in the current system.
				continue
			}
			podResourceKey := defines.PodResourceStatePrefix + "/" + podName
			kv1 := etcd.Get(s.es.Client, podResourceKey).Kvs
			if len(kv1) == 0 {
				// resource of this pod has not been collected by cadvisor yet.
				continue
			}
			// get this replicated pod resource info.
			podResource := &defines.PodResourceUsed{}
			_ = yaml.Unmarshal(kv1[0].Value, podResource)
			if curMinMem > podResource.MemUsed {
				curMinMem = podResource.MemUsed
				minMemId = podResource.PodName
			}
			if curMinCPU > podResource.CpuUsed {
				curMinCPU = podResource.CpuUsed
				minCPUId = podResource.PodName
			}
			totalCPU += podResource.CpuUsed
			totalMem += podResource.MemUsed
			log.Printf("[handler] one autoPod info: cpuUsed: %v, memUsed: %v\n", podResource.CpuUsed, podResource.MemUsed)
		}
		if healthyNum < auto.MinReplicas {
			// running pod number is less than minReplicas number, extend to min number.
			initPodName := defines.AutoPodNamePrefix + auto.AutoName
			yamlKey := defines.AutoScalerYamlPodPrefix + "/" + initPodName
			// log.Println("Add a replica as too high mem usage for pod: ", initPodName)
			kv := etcd.Get(s.es.Client, yamlKey).Kvs
			yamlPod := &defines.YamlPod{}
			_ = yaml.Unmarshal(kv[0].Value, yamlPod)
			// if there are too few alive pod replicas, just extend to min number directly.
			_ = CreateAutoPodReplicas(s.es.Client, yamlPod, auto.MinReplicas-healthyNum)
			continue
		}
		// get average cpu and memory usage.
		avgCPU := totalCPU / uint64(len(nameList))
		avgMem := totalMem / uint64(len(nameList))
		log.Printf("[handler] average cpu usage: %v\n", avgCPU)
		log.Printf("[handler] average memory usage: %v\n", avgMem)
		// NOTE: change edge conditions here.
		if avgCPU == 0 && avgMem == 0 {
			continue
		}
		minMem := uint64(utils.ConvertMemStrToInt(auto.MinMem))
		maxMem := uint64(utils.ConvertMemStrToInt(auto.MaxMem))
		minCPU := auto.MinCPU
		maxCPU := auto.MaxCPU
		// log.Printf("minCPU = %v, maxCPU = %v\n", minCPU, maxCPU)
		// consider CPU usage first.
		if avgCPU < minCPU {
			// cpu usage is too little, shrink
			// remove the pod replica with the least cpu usage.
			if len(nameList) > auto.MinReplicas {
				log.Printf("[handler] remove pod %v as the least CPU usage\n", minCPUId)
				_ = DeleteAutoPodReplica(minCPUId, auto.AutoName)
				// for normal shrink and extend, every time only consider ONE pod change.
				continue
			}
		}
		if avgCPU > maxCPU {
			if len(nameList) < auto.MaxReplicas {
				// cpu usage is too high, extend
				initPodName := defines.AutoPodNamePrefix + auto.AutoName
				yamlKey := defines.AutoScalerYamlPodPrefix + "/" + initPodName
				log.Println("[handler] Add a replica as too high cpu usage for pod: ", initPodName)
				kv := etcd.Get(s.es.Client, yamlKey).Kvs
				yamlPod := &defines.YamlPod{}
				_ = yaml.Unmarshal(kv[0].Value, yamlPod)
				// if we want to do extension, every time just need to extend ONE replica.
				_ = CreateAutoPodReplicas(s.es.Client, yamlPod, 1)
				// for normal shrink and extend, every time only consider ONE pod change.
				continue
			}
		}
		// consider memory usage next.
		if avgMem < minMem {
			// mem usage is too little, shrink
			// remove the pod replica with the least memory usage when there are at least min-number replicas.
			if len(nameList) > auto.MinReplicas {
				log.Printf("[handler] remove pod %v as the least mem usage\n", minMemId)
				_ = DeleteAutoPodReplica(minMemId, auto.AutoName)
				// for normal shrink and extend, every time only consider ONE pod change.
				continue
			}
		}
		if avgMem > maxMem {
			if len(nameList) < auto.MaxReplicas {
				// mem usage is too high, extend
				initPodName := defines.AutoPodNamePrefix + auto.AutoName
				yamlKey := defines.AutoScalerYamlPodPrefix + "/" + initPodName
				log.Println("[handler] Add a replica as too high mem usage for pod: ", initPodName)
				kv := etcd.Get(s.es.Client, yamlKey).Kvs
				yamlPod := &defines.YamlPod{}
				_ = yaml.Unmarshal(kv[0].Value, yamlPod)
				// if we want to do extension, every time just need to extend ONE replica.
				_ = CreateAutoPodReplicas(s.es.Client, yamlPod, 1)
				// for normal shrink and extend, every time only consider ONE pod change.
				continue
			}
		}
	}
}

// CheckShrinkAndExtend TODO: in version 2.0, check pods resources every 15 seconds.
func (s *Server) CheckShrinkAndExtend() {
	for {
		// log.Println("[handler] begin to check shrink and extend!")
		s.CheckAutoPodsResource()
		time.Sleep(15 * time.Second)
	}
}

func (s *Server) ServerCreateAutoPodReplica(context *gin.Context) {
	res := &defines.AutoCreateReplicaSend{}
	yamlPod := &defines.YamlPod{}
	_ = json.NewDecoder(context.Request.Body).Decode(res)
	yamlPod = res.YamlInfo
	// initPodName := yamlPod.Metadata.Name
	// then create a new pod.
	CreatePod(s.es.Client, yamlPod)
	// then update this pod-related-replicasName list.
	listKey := defines.AutoPodReplicasNamesPrefix + "/" + res.PodName
	kv := etcd.Get(s.es.Client, listKey).Kvs
	if len(kv) == 0 {
		// first replica case.
		newList := make([]string, 0)
		newList = append(newList, yamlPod.Metadata.Name)
		newListByte, _ := yaml.Marshal(&newList)
		etcd.Put(s.es.Client, listKey, string(newListByte))
	} else {
		oldList := make([]string, 0)
		_ = yaml.Unmarshal(kv[0].Value, &oldList)
		newList := append(oldList, yamlPod.Metadata.Name)
		newListByte, _ := yaml.Marshal(&newList)
		etcd.Put(s.es.Client, listKey, string(newListByte))
	}
}

func (s *Server) ServerDeleteAutoPodReplica(context *gin.Context) {
	nameSend := &defines.AutoNameSend{}
	_ = json.NewDecoder(context.Request.Body).Decode(nameSend)
	// remove list item first.
	listKey := defines.AutoPodReplicasNamesPrefix + "/" + defines.AutoPodNamePrefix + nameSend.AutoName
	kv := etcd.Get(s.es.Client, listKey).Kvs
	oldList := make([]string, 0)
	_ = yaml.Unmarshal(kv[0].Value, &oldList)
	for idx, name := range oldList {
		if name == nameSend.PodName {
			oldList = append(oldList[0:idx], oldList[idx+1:]...)
		}
	}
	newListByte, _ := yaml.Marshal(&oldList)
	etcd.Put(s.es.Client, listKey, string(newListByte))
	// then remove the real pod obj from etcd.
	_ = DeletePod(s.es.Client, nameSend.PodName)
}

func (s *Server) ServerGetAutoScalers(context *gin.Context) {
	autoName := ""
	_ = json.NewDecoder(context.Request.Body).Decode(&autoName)
	got := GetAutoScaler(s.es.Client)
	context.JSON(http.StatusOK, got)
}

func (s *Server) ServerDescribeAutoScaler(context *gin.Context) {
	autoName := ""
	_ = json.NewDecoder(context.Request.Body).Decode(&autoName)
	got := DescribeAutoScaler(s.es.Client, autoName)
	context.JSON(http.StatusOK, got)
}

func (s *Server) ServerUpdateEtcdAutoControllerList(context *gin.Context) {
	key := defines.AutoScalerControllerPrefix + "/"
	controller := &defines.AutoScalerController{}
	err := json.NewDecoder(context.Request.Body).Decode(controller)
	if err != nil {
		log.Printf("an error occurs when update etcd auto controller info: %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Error in updating etcd controller list!"})
		return
	}
	controllerByte, _ := yaml.Marshal(controller)
	etcd.Put(s.es.Client, key, string(controllerByte))
	context.JSON(http.StatusOK, "OK")
}

func (s *Server) ServerUpdatePodRestartInfo(context *gin.Context) {
	podInfo := &defines.Pod{}
	err := json.NewDecoder(context.Request.Body).Decode(podInfo)
	if err != nil {
		log.Printf("an error occurs when updating pod restart info: %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Error in ServerUpdatePodInfo!"})
		return
	}
	key := defines.PodInstancePrefix + "/" + podInfo.Metadata.Name
	podInfoByte, _ := yaml.Marshal(podInfo)
	etcd.Put(s.es.Client, key, string(podInfoByte))
	context.JSON(http.StatusOK, "OK")
}

func (s *Server) ServerCreateDeployment(context *gin.Context) {
	deploy := &defines.Deployment{}
	err := json.NewDecoder(context.Request.Body).Decode(deploy)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Error in ServerCreateDeployment!"})
		return
	}
	res := CreateDeployment(s.es.Client, deploy)
	if res != nil {
		context.JSON(http.StatusOK, "OK")
		return
	} else {
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "failt to create a new deployment!"})
		return
	}
}

func (s *Server) ServerCreateReplicaSet(context *gin.Context) {
	rs := &defines.ReplicaSet{}
	err := json.NewDecoder(context.Request.Body).Decode(rs)
	if err != nil {
		log.Printf("ServerCreateReplicaSet error: %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Wrong yaml file when create replicaSet!"})
		return
	}
	err = CreateReplicaSet(s.es.Client, rs)
	if err != nil {
		log.Printf("ServerCreateReplicaSet error: %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": err.Error()})
	} else {
		context.JSON(http.StatusOK, gin.H{"INFO": "The replicaSet has been created successfully!"})
	}
}

func (s *Server) ServerDeleteReplicaSet(context *gin.Context) {
	rsName := ""
	err := json.NewDecoder(context.Request.Body).Decode(&rsName)
	if err != nil {
		log.Printf("ServerDeleteReplicaSet error: %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Wrong replicaSet name when delete one!"})
		return
	}
	err = DeleteReplicaSet(s.es.Client, rsName)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": err.Error()})
	} else {
		context.JSON(http.StatusOK, gin.H{"INFO": "The replicaSet has been deleted successfully!"})
	}
}

func (s *Server) ServerGetReplicaSet(context *gin.Context) {
	rss, err := GetReplicaSets(s.es.Client)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"ERROR": err.Error()})
	} else {
		context.JSON(http.StatusOK, gin.H{"INFO": rss})
	}
}

func (s *Server) ServerDesReplicaSet(context *gin.Context) {
	rsName := ""
	err := json.NewDecoder(context.Request.Body).Decode(&rsName)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": err.Error()})
		return
	}
	rsInfo, err := DesReplicaSet(s.es.Client, rsName)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": err.Error()})
	} else {
		if err != nil {
			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": err.Error()})
			return
		}
		context.JSON(http.StatusOK, gin.H{"INFO": rsInfo})
	}
}

func (s *Server) ServerHandleNewReplica(context *gin.Context) {
	rs := &defines.ReplicaSet{}
	err := json.NewDecoder(context.Request.Body).Decode(rs)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Wrong yaml format!"})
		return
	}
	err = s.ReplicaController.AddSet(rs)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Error in joining set!"})
		return
	}
	context.JSON(http.StatusOK, gin.H{"INFO": "The replicaSet has been handled successfully!"})
}

func (s *Server) ServerHandleDelReplica(context *gin.Context) {
	rsName := ""
	err := json.NewDecoder(context.Request.Body).Decode(&rsName)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": err.Error()})
		return
	}
	err = s.ReplicaController.DeleteSet(rsName)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": err.Error()})
	} else {
		context.JSON(http.StatusOK, gin.H{"INFO": "Delete replicaSet successfully!"})
	}
}

func (s *Server) ServerHandleDesReplicaSet(context *gin.Context) {
	rsName := ""
	err := json.NewDecoder(context.Request.Body).Decode(&rsName)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": err.Error()})
		return
	}
	rsi := s.ReplicaController.DesReplicaSet(rsName)
	if rsi == nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "The replicaSet doesn't exist in RC!"})
		return
	}
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": err.Error()})
		return
	} else {
		context.JSON(http.StatusOK, gin.H{"INFO": rsi})
	}
}

func (s *Server) ServerGetAllPods(context *gin.Context) {
	pods := make([]*defines.Pod, 0)
	kvs := s.es.GetWithPrefix(defines.NodePrefix + "/").Kvs
	for _, kv := range kvs {
		nodeInstance := &defines.NodeInfo{}
		err := yaml.Unmarshal(kv.Value, nodeInstance)
		if err != nil {
			log.Printf("ServerGetAllPods error: %v\n", err)
			continue
		}
		// if node is not ready ,DO NOT consider the pods manged ny this node !!!
		if nodeInstance.NodeData.NodeState == defines.NodeNotReady {
			continue
		}
		pods = append(pods, nodeInstance.Pods...)
	}
	podsVal, err := json.Marshal(pods)
	if err != err {
		log.Printf("ServerGetAllPods error: %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Error in marshal pod list!"})
		return
	}
	context.JSON(http.StatusOK, gin.H{"INFO": podsVal})
}

func (s *Server) ServerGetAllReplicaSets(context *gin.Context) {
	replicaSets := make([]*defines.ReplicaSet, 0)
	KVS := etcd.GetWithPrefix(s.es.Client, defines.RSInstancePrefix+"/").Kvs
	for _, kv := range KVS {
		replicaSet := &defines.ReplicaSet{}
		err := yaml.Unmarshal(kv.Value, replicaSet)
		if err != nil {
			log.Printf("ServerGetAllReplicaSets error: %v\n", err)
			etcd.Del(s.es.Client, string(kv.Key))
			continue
		}
		replicaSets = append(replicaSets, replicaSet)
	}

	replicaSetsInfo, err := json.Marshal(replicaSets)
	if err != nil {
		log.Printf("ServerGetAllReplicaSets error: %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Something went wrong in marshal result!"})
	} else {
		context.JSON(http.StatusOK, gin.H{"INFO": replicaSetsInfo})
	}
}

func (s *Server) ServerGetNodeResourceUsage(context *gin.Context) {
	nodeName := ""
	err := json.NewDecoder(context.Request.Body).Decode(&nodeName)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Error in getting node resource usage info!"})
		return
	}
	got := GetNodeResourceInfo()
	context.JSON(http.StatusOK, got)
}

func (s *Server) ServerSchedulePolicy(context *gin.Context) {
	yamlPod := &defines.YamlPod{}
	_ = json.NewDecoder(context.Request.Body).Decode(yamlPod)
	res := s.Scheduler.SchedulePolicy(yamlPod)
	context.JSON(http.StatusOK, res)
}

func (s *Server) ServerCreateDNS(context *gin.Context) {
	DNSYaml := &defines.DNSYaml{}
	err := json.NewDecoder(context.Request.Body).Decode(DNSYaml)
	if err != nil {
		fmt.Printf("%v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Wrong yaml file when create pod!"})
		return
	}
	// then create the new pod in the api-server layer.
	newDNS := CreateDNS(s.es.Client, DNSYaml)
	if newDNS == nil {
		// error occurs.
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Error in creating a new dns!"})
	} else {
		context.JSON(http.StatusOK, gin.H{"INFO": "The dns has been created successfully!"})
	}
}

func (s *Server) ServerGetAllServices(context *gin.Context) {
	res := make([]*defines.EtcdService, 0)
	all := &defines.AllSvc{}
	key := defines.ServicePrefix + "/"
	kv := etcd.GetWithPrefix(s.es.Client, key).Kvs
	if len(kv) == 0 {
		context.JSON(http.StatusOK, res)
	} else {
		for _, oneKV := range kv {
			tmp := &defines.EtcdService{}
			_ = yaml.Unmarshal(oneKV.Value, tmp)
			res = append(res, tmp)
		}
		log.Printf("[handler] all svc = %v\n", res)
		all.Svc = res
		context.JSON(http.StatusOK, all)
	}
}

func (s *Server) ServerReInitAllServicesChains() {
	for _, svcInfo := range s.Services {
		svcChainName := config.KubeSvcChainPrefix + svcInfo.SvcName
		if !kubeproxy.IsChainExist(svcChainName, "nat") {
			kubeproxy.CreateChain("nat", svcChainName)
			kubeproxy.AddSvcMatchRuleToChain(config.KubeSvcMainChainName, "nat", svcInfo.SvcClusterIP,
				strconv.Itoa(svcInfo.SvcPorts[0].Port), svcInfo.SvcPorts[0].Protocol, svcChainName)
		}
		num := len(svcInfo.SvcPods)
		probability := 1.0 / float64(num)
		probabilityStr := strconv.FormatFloat(probability, 'f', -1, 64)
		for _, svcPod := range svcInfo.SvcPods {
			podDNATChainName := config.PodDNATChainPrefix + svcPod.PodId
			kubeproxy.CreateChain("nat", podDNATChainName)
			kubeproxy.AddSvcForwardRuleToChain(svcChainName, "nat", podDNATChainName, probabilityStr)
			kubeproxy.AddSvcDNATRuleToChain(podDNATChainName, svcPod.PodIp+":"+strconv.Itoa(svcInfo.SvcPorts[0].TargetPort))
		}
		log.Printf("[handler] Success to re-init all chains for svc %v\n", svcInfo.SvcName)
	}
	log.Printf("[handler] re-init all chain rules after kubeproxy starts successfully!\n")
}

func (s *Server) ServerCreateGPUJob(context *gin.Context) {
	job := &defines.GPUJob{}
	err := json.NewDecoder(context.Request.Body).Decode(job)
	if err != nil {
		log.Printf("error when decode gpuJob yaml info: %v\n", err)
		return
	}
	res := CreateGPUJob(s.es.Client, job)
	if res == nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "error in creating a new GPUJob object"})
		return
	}
	context.JSON(http.StatusOK, "OK")
}

func (s *Server) ServerUpdatePodInfoAfterRealOp(context *gin.Context) {
	podInfo := &defines.Pod{}
	err := json.NewDecoder(context.Request.Body).Decode(podInfo)
	if err != nil {
		log.Printf("ServerUpdatePodInfoAfterRealOp error: %v\n", err)
		return
	}
	key := defines.PodInstancePrefix + "/" + podInfo.Metadata.Name
	podInfoByte, _ := yaml.Marshal(podInfo)
	etcd.Put(s.es.Client, key, string(podInfoByte))
	context.JSON(http.StatusOK, "OK")
}

func (s *Server) ServerUpdateNodeInfoAfterRealOp(context *gin.Context) {
	nodeInfo := &defines.NodeInfo{}
	err := json.NewDecoder(context.Request.Body).Decode(nodeInfo)
	if err != nil {
		log.Printf("ServerUpdateNodeInfoAfterRealOp error: %v\n", err)
		return
	}
	key := defines.NodePrefix + "/" + nodeInfo.NodeData.NodeSpec.Metadata.Name
	nodeInfoByte, _ := yaml.Marshal(nodeInfo)
	etcd.Put(s.es.Client, key, string(nodeInfoByte))
	context.JSON(http.StatusOK, "OK")
}

//	func (s *Server) ServerRemoveOneImage(context *gin.Context) {
//		name := ""
//		err := json.NewDecoder(context.Request.Body).Decode(&name)
//		if err != nil {
//			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "fail to unmarshal target image name to remove!"})
//			return
//		}
//		exist := container.CheckImgInList(name)
//		if exist == false {
//			return
//		}
//		// exist in the current node.
//		err = container.RemoveImage(name)
//		if err != nil {
//			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "fail to remove an image!"})
//			return
//		}
//		return
//	}
func (s *Server) ServerInitFunc(context *gin.Context) {
	function := ""
	err := json.NewDecoder(context.Request.Body).Decode(&function)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Can not parse the function name!"})
		return
	}
	etcd.Put(s.es.Client, defines.FunctionPrefix+"/"+function, function)
	context.JSON(http.StatusOK, gin.H{"INFO": "Initial function successfully!"})
}

func (s *Server) ServerGetServerlessObject(context *gin.Context) {
	kind := context.Param("type")
	if kind == "F" {
		functions := make([]string, 0)
		KVS := etcd.GetWithPrefix(s.es.Client, defines.FunctionPrefix+"/").Kvs
		for _, kv := range KVS {
			functions = append(functions, string(kv.Value))
		}

		content, _ := json.Marshal(functions)
		context.JSON(http.StatusOK, gin.H{"INFO": content})
	} else if kind == "W" {
		workflows := make([]string, 0)
		KVS := etcd.GetWithPrefix(s.es.Client, defines.WorkFlowPrefix+"/").Kvs
		for _, kv := range KVS {
			workflows = append(workflows, string(kv.Key)[len(defines.WorkFlowPrefix)+1:])
		}
		content, _ := json.Marshal(workflows)
		context.JSON(http.StatusOK, gin.H{"INFO": content})
	} else {
		log.Printf("[ERROR] Unknown type %s!\n", kind)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Unknown type!"})
	}
}

func (s *Server) ServerDelOldWorkflows(context *gin.Context) {
	log.Println("[WARNING] Try to delete all workflows, should only be used when serverless-controller restart!")
	etcd.DelWithPrefix(s.es.Client, defines.WorkFlowPrefix+"/")
	context.JSON(http.StatusOK, gin.H{"INFO": "Initial operation on workflow success!"})
}

func (s *Server) ServerNewFunc(context *gin.Context) {
	op := context.Param("op")
	if op == "" || (op != "add" && op != "update") {
		log.Printf("[ERROR] Unknown op type: %s.\n", op)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Unknown op type!"})
		return
	}

	file, err := context.FormFile("file")
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "can't parsing the uploaded file!"})
		return
	}

	if op == "add" {
		_, err = os.Stat(filepath.Join(config.ServerlessFileDir, file.Filename[0:len(file.Filename)-4]))
		if !os.IsNotExist(err) {
			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "The function has existed!"})
			return
		}
	}

	filePath := filepath.Join(config.ServerlessTmpFileDir, file.Filename)
	err = context.SaveUploadedFile(file, filePath)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "can't save the uploaded file!"})
		return
	}
	defer func(name string) {
		_ = os.Remove(name)
	}(filePath)

	err = utils.UnzipFile(filePath, config.ServerlessFileDir)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "can't unzip the uploaded file!"})
		return
	}

	url := ""
	if op == "add" {
		url = "http://" + config.MasterIP + ":" + config.SLPort + "/objectAPI/newFunction/add"
	} else {
		url = "http://" + config.MasterIP + ":" + config.SLPort + "/objectAPI/newFunction/update"
	}

	body, _ := json.Marshal(file.Filename[0 : len(file.Filename)-4])
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Something went wrong when communicate with serverless server!"})
		return
	}
	request.Header.Add("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Something went wrong when communicate with serverless server!"})
		return
	}
	result := make(map[string]string)
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Something went wrong when parsing serverless server res!"})
		return
	}
	if res.StatusCode == http.StatusOK {
		etcd.Put(s.es.Client, defines.FunctionPrefix+"/"+file.Filename[0:len(file.Filename)-4], file.Filename[0:len(file.Filename)-4])
		context.JSON(http.StatusOK, gin.H{"INFO": result["INFO"]})
	} else {
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": result["ERROR"]})
	}
}

func (s *Server) ServerNewWorkFlow(context *gin.Context) {
	op := context.Param("op")
	if op == "" || (op != "add" && op != "update") {
		log.Printf("[ERROR] Unknown op type: %s.\n", op)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Unknown op type!"})
		return
	}

	workflow := &defines.WorkFlow{}

	err := json.NewDecoder(context.Request.Body).Decode(workflow)

	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Can not parse request body!"})
		return
	}

	body, err := json.Marshal(workflow)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Can not marshal the workflow!"})
		return
	}

	KVS := etcd.Get(s.es.Client, defines.WorkFlowPrefix+"/"+workflow.Name).Kvs
	if len(KVS) != 0 {
		if op == "add" {
			log.Printf("[ERROR] The workflow %s has existed!\n", workflow.Name)
			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "The workflow has existed!"})
			return
		}
	} else {
		if op == "update" {
			log.Printf("[ERROR] The workflow %s doesn't exist!\n", workflow.Name)
			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "The workflow doesn't exist!"})
			return
		}
	}

	url := ""
	if op == "add" {
		url = "http://" + config.MasterIP + ":" + config.SLPort + "/objectAPI/newWorkFlow/add"
	} else {
		url = "http://" + config.MasterIP + ":" + config.SLPort + "/objectAPI/newWorkFlow/update"
	}

	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Something went wrong when communicate with serverless server!"})
		return
	}
	request.Header.Add("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Something went wrong when communicate with serverless server!"})
		return
	}

	result := make(map[string]string)
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Can not parse response from serverless server!"})
		return
	}

	if res.StatusCode == http.StatusOK {
		etcd.Put(s.es.Client, defines.WorkFlowPrefix+"/"+workflow.Name, string(body))
		log.Printf("[INFO] %v\n", result["INFO"])
		context.JSON(http.StatusOK, gin.H{"INFO": "Add workflow successfully!"})
	} else {
		log.Printf("[ERROR] %v\n", result["ERROR"])
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": result["ERROR"]})
	}

}

func (s *Server) ServerGetFuncPods(context *gin.Context) {
	name := context.Param("name")
	pods := make([]*defines.Pod, 0)

	KVS := etcd.GetWithPrefix(s.es.Client, defines.PodInstancePrefix+"/function-"+name).Kvs

	for _, kv := range KVS {
		p := &defines.Pod{}
		err := yaml.Unmarshal(kv.Value, p)
		if err != nil || p.PodState != defines.Running {
			continue
		}
		pods = append(pods, p)
	}

	body, err := json.Marshal(pods)

	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Something went wrong when marshal pods!"})
	}

	context.JSON(http.StatusOK, gin.H{"INFO": body})
}

func (s *Server) ServerDelImage(context *gin.Context) {
	name := ""
	err := json.NewDecoder(context.Request.Body).Decode(&name)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Can not parse image name!"})
		return
	}
	if name == "" {
		log.Println("[WARNING] No image name specified!")
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "No image name specified!"})
		return
	}

	body, _ := json.Marshal(name)
	flag := true

	kvs := etcd.GetWithPrefix(s.es.Client, defines.NodePrefix+"/").Kvs
	for _, kv := range kvs {
		nodeInfo := &defines.NodeInfo{}
		_ = yaml.Unmarshal(kv.Value, nodeInfo)
		url := fmt.Sprintf("http://%v:%v/objectAPI/removeOneImage",
			nodeInfo.NodeData.NodeSpec.Metadata.Ip, nodeInfo.NodeData.NodeSpec.Metadata.Port)
		req, err := http.NewRequest("POST", url, bytes.NewReader(body))
		if err != nil {
			log.Printf("[ERROR] %v\n", err)
			flag = false
			continue
		}
		req.Header.Add("Content-Type", "application/json")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("[ERROR] %v\n", err)
			flag = false
			continue
		}

		if res.StatusCode != http.StatusOK {
			flag = false
			result := make(map[string]string)
			err = json.NewDecoder(res.Body).Decode(&result)
			if err != nil {
				log.Printf("[ERROR] The del-image request has failed, but the response can not be parsed!(error: %v)\n", err)
				continue
			}
			log.Printf("[ERROR] %v\n", result["ERROR"])
		}
	}

	if flag {
		context.JSON(http.StatusOK, gin.H{"INFO": "The image has been deleted successfully!"})
	} else {
		context.JSON(http.StatusOK, gin.H{"INFO": "[WARNING] Something may be incorrect during deleting the image!"})
	}
}

//func (s *Server) ServerDelFuncPod(context *gin.Context) {
//	name := context.Param("name")
//
//	KVS := etcd.GetWithPrefix(s.es.Client, defines.PodInstancePrefix+"/function-"+name).Kvs
//
//	p := &defines.Pod{}
//	for _, kv := range KVS {
//		err := yaml.Unmarshal(kv.Value, p)
//		if err != nil || p.PodState != defines.Running {
//			p = &defines.Pod{}
//			continue
//		} else {
//			break
//		}
//	}
//	if len(KVS) == 0 || p.Metadata.Name == "" {
//		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "The function doesn't have running replica!"})
//	} else {
//		err := DeletePod(s.es.Client, p.Metadata.Name)
//		if err != nil {
//			log.Printf("[ERROR] %v\n", err)
//			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Something went wrong in deleting replica!"})
//		} else {
//			context.JSON(http.StatusOK, gin.H{"INFO": "Delete replica successfully!"})
//		}
//	}
//}

func (s *Server) ServerTrigger(context *gin.Context) {
	kind := context.Param("type")
	if kind != "F" && kind != "W" {
		log.Printf("[ERROR] Unsupported type %s.\n", kind)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Unknown type!"})
		return
	}
	name := context.Param("name")
	params := make(map[string]string)

	err := json.NewDecoder(context.Request.Body).Decode(&params)

	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Can not parse function parameters!"})
		return
	}

	url := ""

	if kind == "F" {
		url = "http://" + config.MasterIP + ":" + config.SLPort + "/objectAPI/trigger/" + name
	} else {
		url = "http://" + config.MasterIP + ":" + config.SLPort + "/objectAPI/triggerWorkFlow/" + name
	}

	body, _ := json.Marshal(params)

	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Something went wrong when communicate with serverless server!"})
		return
	}
	request.Header.Add("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Something went wrong when communicate with serverless server!"})
		return
	}

	if res.StatusCode == http.StatusOK {
		result := make(map[string][]byte)
		err = json.NewDecoder(res.Body).Decode(&result)
		if err != nil {
			log.Printf("[ERROR] %v\n", err)
			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Can not parse result map!"})
			return
		}
		context.JSON(http.StatusOK, gin.H{"INFO": result["INFO"]})
	} else {
		result := make(map[string]string)
		err = json.NewDecoder(res.Body).Decode(&result)
		if err != nil {
			log.Printf("[ERROR] %v\n", err)
			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Can not parse result map!"})
			return
		}
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": result["ERROR"]})
	}
}

func (s *Server) ServerDeleteServerlessObject(context *gin.Context) {
	kind := context.Param("type")
	name := context.Param("name")
	if kind != "F" && kind != "W" {
		log.Printf("[ERROR] Unknown type %s!\n", kind)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Unknown type!"})
		return
	} else if name == "" {
		log.Println("[ERROR] Name should not be empty!")
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Name should not be empty!"})
		return
	}

	defer func() {
		if kind == "F" {
			etcd.Del(s.es.Client, defines.FunctionPrefix+"/"+name)
		} else {
			etcd.Del(s.es.Client, defines.WorkFlowPrefix+"/"+name)
		}
	}()

	url := fmt.Sprintf("http://%v:%v/objectAPI/del/%v/%v", config.MasterIP, config.SLPort, kind, name)
	request, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Something went wrong when communicate with serverless server!"})
		return
	}
	request.Header.Add("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Something went wrong when communicate with serverless server!"})
		return
	}

	result := make(map[string]string)
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Can not parse result map!"})
		return
	}

	if res.StatusCode == http.StatusOK {
		context.JSON(http.StatusOK, gin.H{"INFO": result["INFO"]})
	} else {
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": result["ERROR"]})
	}
}

func (s *Server) ServerHandleNewFunction(context *gin.Context) {
	op := context.Param("op")
	if op == "" || (op != "add" && op != "update") {
		log.Printf("[ERROR] Unknown op type: %s.\n", op)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Unknown op type!"})
		return
	}

	fileName := ""
	err := json.NewDecoder(context.Request.Body).Decode(&fileName)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Something went wrong when serverless controller tries to parsing the fileName!"})
		return
	}
	if op == "add" {
		err = s.ServerlessController.AddFunction(fileName)
	} else {
		err = s.ServerlessController.UpdateFunction(fileName)
	}
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": err.Error()})
	} else {
		context.JSON(http.StatusOK, gin.H{"INFO": "Add function successfully!"})
	}
}

func (s *Server) ServerHandleTrigger(context *gin.Context) {
	function := context.Param("function")
	params := make(map[string]string)
	err := json.NewDecoder(context.Request.Body).Decode(&params)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "There is something wrong in params parsing!"})
		return
	}

	result, err := s.ServerlessController.TriggerFunction(function, params)

	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": err.Error()})
	} else {
		value, err := json.Marshal(result)
		if err != nil {
			log.Printf("[ERROR] %v\n", err)
			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Can't parse result string!"})
			return
		}
		context.JSON(http.StatusOK, gin.H{"INFO": value})
	}
}

func (s *Server) ServerDescribeGPUJob(context *gin.Context) {
	name := ""
	err := json.NewDecoder(context.Request.Body).Decode(&name)
	if err != nil {
		log.Printf("[handler] fail to decode describe job name!\n")
		return
	}
	job := &defines.EtcdGPUJob{}
	key := defines.GPUJobPrefix + "/" + name
	kv := etcd.Get(s.es.Client, key).Kvs
	if len(kv) == 0 {
		context.JSON(http.StatusOK, job)
	} else {
		_ = yaml.Unmarshal(kv[0].Value, job)
		context.JSON(http.StatusOK, job)
	}
}

func (s *Server) ServerGetAllGPUJobs(context *gin.Context) {
	name := ""
	err := json.NewDecoder(context.Request.Body).Decode(&name)
	if err != nil {
		log.Printf("[handler] fail to decode describe job name!\n")
		return
	}
	prefixKey := defines.GPUJobPrefix + "/"
	kvs := etcd.GetWithPrefix(s.es.Client, prefixKey).Kvs
	jobs := make([]*defines.EtcdGPUJob, 0)
	res := &defines.AllJobs{}
	for _, kv := range kvs {
		tmp := &defines.EtcdGPUJob{}
		_ = yaml.Unmarshal(kv.Value, tmp)
		jobs = append(jobs, tmp)
	}
	res.Jobs = jobs
	context.JSON(http.StatusOK, res)
}

func (s *Server) ServerUpdateGPUJobState(context *gin.Context) {
	job := &defines.EtcdGPUJob{}
	err := json.NewDecoder(context.Request.Body).Decode(job)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Can't parse update GPUJob info!"})
		return
	}
	key := defines.GPUJobPrefix + "/" + job.JobInfo.Name
	jobByte, _ := yaml.Marshal(job)
	etcd.Put(s.es.Client, key, string(jobByte))
	context.JSON(http.StatusOK, "OK")
}

func (s *Server) ServerDeleteGPUJob(context *gin.Context) {
	name := ""
	err := json.NewDecoder(context.Request.Body).Decode(&name)
	if err != nil {
		log.Printf("[handler] fail to decode describe job name!\n")
		return
	}
	res := DeleteGPUJob(s.es.Client, name)
	if res == nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "fail to delete GPUJob!"})
		return
	}
	context.JSON(http.StatusOK, res)
}

func (s *Server) ServerRemoveOneImage(context *gin.Context) {
	name := ""
	err := json.NewDecoder(context.Request.Body).Decode(&name)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "fail to unmarshal target image name to remove!"})
		return
	}
	exist := container.CheckImgInList(name)
	if exist == false {
		return
	}
	// exist in the current node.
	err = container.RemoveImage(name)
	if err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "fail to remove an image!"})
		return
	}
	return
}

func (s *Server) ServerHandleNewWorkFlow(context *gin.Context) {
	op := context.Param("op")
	if op == "" || (op != "add" && op != "update") {
		log.Printf("[ERROR] Unknown op type: %s.\n", op)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Unknown op type!"})
		return
	}

	workflow := &defines.WorkFlow{}

	err := json.NewDecoder(context.Request.Body).Decode(workflow)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Something went wrong in parsing request body!"})
		return
	}
	if op == "add" {
		err = s.ServerlessController.AddWorkFlow(workflow)
	} else {
		err = s.ServerlessController.UpdateWorkFlow(workflow)
	}
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": err.Error()})
	} else {
		context.JSON(http.StatusOK, gin.H{"INFO": "Add workflow successfully!"})
	}
}

func (s *Server) ServerHandleTriggerWorkFlow(context *gin.Context) {
	workflow := context.Param("workflow")

	params := make(map[string]string)

	err := json.NewDecoder(context.Request.Body).Decode(&params)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Can not parse function parameters!"})
		return
	}

	result, err := s.ServerlessController.TriggerWorkFlow(workflow, params)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": err.Error()})
		return
	}

	resultVal, err := json.Marshal(result)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Something went wrong when parse result map!"})
	} else {
		context.JSON(http.StatusOK, gin.H{"INFO": resultVal})
	}
}

func (s *Server) ServerGetAllControllerScalers(context *gin.Context) {
	name := ""
	_ = json.NewDecoder(context.Request.Body).Decode(&name)
	key := defines.AutoScalerControllerPrefix + "/"
	res := &defines.AutoScalerController{}
	kv := etcd.Get(s.es.Client, key).Kvs
	if len(kv) == 0 {
		context.JSON(http.StatusOK, res)
		return
	} else {
		_ = yaml.Unmarshal(kv[0].Value, res)
		context.JSON(http.StatusOK, res)
		return
	}
}

func (s *Server) ServerHandleDelServerlessObject(context *gin.Context) {
	kind := context.Param("type")
	name := context.Param("name")

	if kind == "F" {
		err := s.ServerlessController.DelFunction(name)
		if err != nil {
			log.Printf("[ERROR] %v\n", err)
			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": err.Error()})
		} else {
			context.JSON(http.StatusOK, gin.H{"INFO": "delete function successfully!"})
		}
	} else if kind == "W" {
		err := s.ServerlessController.DelWorkFlow(name)
		if err != nil {
			log.Printf("[ERROR] %v\n", err)
			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": err.Error()})
		} else {
			context.JSON(http.StatusOK, gin.H{"INFO": "delete workflow successfully!"})
		}
	} else {
		log.Printf("[ERROR] Unknown type %s!\n", kind)
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Unknown type!"})
	}
}

func (s *Server) ServerUpdatePodHealthState(context *gin.Context) {
	podInfo := &defines.Pod{}
	err := json.NewDecoder(context.Request.Body).Decode(podInfo)
	if err != nil {
		log.Printf("[handler] fail to parse podInfo in func ServerUpdatePodHealthState!\n")
		return
	}
	key := defines.PodInstancePrefix + "/" + podInfo.Metadata.Name
	podInfoByte, _ := yaml.Marshal(podInfo)
	etcd.Put(s.es.Client, key, string(podInfoByte))
	context.JSON(http.StatusOK, "OK")
}

func (s *Server) ServerGetAllReplicaPodNames(context *gin.Context) {
	name := ""
	err := json.NewDecoder(context.Request.Body).Decode(&name)
	if err != nil {
		log.Printf("[handler] error in func ServerGetAllReplicaPodNames: %v!\n", err)
		return
	}
	allNames := &defines.AllReplicaPodNames{}
	allNames.Names = make([]string, 0)
	key := defines.RSInstancePrefix + "/"
	kvs := etcd.GetWithPrefix(s.es.Client, key).Kvs
	if len(kvs) != 0 {
		for _, kv := range kvs {
			rs := &defines.ReplicaSet{}
			_ = yaml.Unmarshal(kv.Value, rs)
			nameKey := defines.PodInstancePrefix + "/" + rs.Metadata.Name + "-"
			nameKVs := etcd.GetWithPrefix(s.es.Client, nameKey).Kvs
			for _, nameKV := range nameKVs {
				podInfo := &defines.Pod{}
				_ = yaml.Unmarshal(nameKV.Value, podInfo)
				allNames.Names = append(allNames.Names, podInfo.Metadata.Name)
			}
		}
	}
	context.JSON(http.StatusOK, allNames)
}

func (s *Server) ServerCheckRCPodState(context *gin.Context) {
	podInfo := &defines.Pod{}
	err := json.NewDecoder(context.Request.Body).Decode(podInfo)
	if err != nil {
		log.Printf("[handler] error in func ServerCheckRCPodState: %v\n", err)
		return
	}
	res := cadvisor.CheckReplicaSetPodContainersState(podInfo)
	tmp := &defines.ReplicaPodState{}
	tmp.Live = res
	context.JSON(http.StatusOK, tmp)
}

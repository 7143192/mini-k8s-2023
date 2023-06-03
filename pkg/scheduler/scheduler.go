package scheduler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"math"
	"mini-k8s/pkg/config"
	"mini-k8s/pkg/defines"
	"mini-k8s/pkg/etcd"
	"mini-k8s/utils"
	"mini-k8s/utils/client"
	"mini-k8s/utils/queue"
	"net/http"
	"strings"
	"time"
)

type Scheduler struct {
	//msgList   *message.MsgList
	podQueue  *queue.Queue
	nodeList  map[string]int
	client    *client.Client
	NodeNames []string
}

// This function is used to tolerate fault, such as scheduler fails.
func (s *Scheduler) scheduleWhenStart(pods []*defines.YamlPod) {
	for _, yamlPod := range pods {
		for {
			nodeID := "node1" /*s.SchedulePolicy(yamlPod)*/
			result := make(map[string]string)
			result["podName"] = yamlPod.Metadata.Name
			result["nodeID"] = nodeID
			content, err := json.Marshal(result)
			if err != nil {
				log.Printf("[ERROR]: %v\n", err)
				break
			}
			//TODO(lyh): I send nodeID to "/objectAPI/updatePod", and it need some changes to adapt to pod info change.
			res, err := s.client.PostRequest(bytes.NewReader(content), "/objectAPI/updatePod")
			if err == nil {
				log.Println(res)
				break
			} else {
				log.Printf("[ERROR] %v\n", err)
			}
			time.Sleep(2 * time.Second)
		}
	}
}

func NewScheduler( /*queueName string*/ ) *Scheduler {
	//msgList := message.NewMsgList()
	////
	//err := msgList.Put(defines.PodIdSetPrefix, queueName)
	//if err != nil {
	//	log.Printf("[ERROR]: %v\n", err)
	//	return nil
	//}
	////
	//err = msgList.Put(defines.AllNodeSetPrefix, defines.AllNodeSetPrefix)
	//if err != nil {
	//	log.Printf("[ERROR]: %v\n", err)
	//	return nil
	//}
	//podQueue := &queue.Queue{}
	cli := client.NewClient(config.MasterIP, config.MasterPort)
	s := &Scheduler{
		nodeList: make(map[string]int),
		client:   cli,
	}
	status, res, err := cli.GetRequest("/objectAPI/getUnhandledPods")
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
	} else if status != http.StatusOK {
		log.Printf("[ERROR] %s\n", string(res))
	} else {
		result := make([]*defines.YamlPod, 0)
		err = json.Unmarshal(res, &result)
		if err == nil {
			go s.scheduleWhenStart(result)
		} else {
			log.Printf("[ERROR] %v\n", err)
		}
	}
	return s
}

func (s *Scheduler) watchNewPod(content []byte) {
	//NOTE: If the pod has been given a nodeID? We just send it to server and let server handle it!
	podName := string(content)
	s.podQueue.Enqueue(podName)
}

func (s *Scheduler) watchNode(nodeName []byte) {
	msg := strings.Fields(string(nodeName))
	if msg[0] == "del" {
		if _, exist := s.nodeList[msg[1]]; exist {
			delete(s.nodeList, msg[1])
		} else {
			log.Println("[ERROR] The del-node isn't in the list, which should not happen!")
		}
	} else {
		if _, exist := s.nodeList[msg[0]]; exist {
			log.Printf("[ERROR] The node %s has been in the list, which should not happen!", string(nodeName))
		} else {
			s.nodeList[msg[0]] = 0
		}
	}
}

func (s *Scheduler) dispatch() { // A simple round-robin scheduler
	for {
		podName := s.podQueue.Dequeue()
		if podName == nil {
			time.Sleep(10 * time.Millisecond)
			continue
		}
		//Race-condition: the only node is removed right after the if-statement
		if len(s.nodeList) == 0 {
			s.podQueue.Enqueue(podName)
			log.Println("[INFO]: Now there is no node!")
			time.Sleep(5 * time.Millisecond)
		} else {
			index := ""
			min := 2147483647
			for i, busy := range s.nodeList {
				if busy < min {
					index = i
					min = busy
				}
			}
			if index == "" {
				log.Println("[ERROR]: Racing condition happens in nodeList!")
				s.podQueue.Enqueue(podName)
				time.Sleep(5 * time.Millisecond)
				continue
			}
			s.nodeList[index]++

			result := make(map[string]string)
			result["podName"] = fmt.Sprintf("%v", podName)
			result["nodeID"] = index
			content, err := json.Marshal(result)
			if err != nil {
				log.Fatalln("[ERROR]: Error in marshal sent pod!")
			}

			response, err := s.client.PostRequest(bytes.NewReader(content), "/objectAPI/schePod")
			if err != nil {
				log.Printf("[ERROR]: %v\n", err)
			} else {
				log.Print(response)
			}
		}
	}
}

// ChooseNodeForPod
// @podName may be used to get pod resource info to help the scheduler to do better choice.
// TODO: in version 2.0, this is the simplest RR version.
func (s *Scheduler) ChooseNodeForPod(podName string) string {
	// init node list here.
	s.getInitNode()
	if len(s.nodeList) == 0 {
		// s.podQueue.Enqueue(podName)
		log.Println("[INFO]: Now there is no node!")
		return ""
	} else {
		index := ""
		min := 2147483647
		for i, busy := range s.nodeList {
			if busy < min {
				index = i
				min = busy
			}
		}
		if index == "" {
			log.Println("[ERROR]: Racing condition happens in nodeList!")
			// s.podQueue.Enqueue(podName)
			// time.Sleep(5 * time.Millisecond)
			return ""
		}
		s.nodeList[index]++
		return index
	}
}

func (s *Scheduler) getInitNode() {
	//// test only !!!
	//names := make([]string, 0)
	//names = append(names, "node1")
	//names = append(names, "node2")
	//names = append(names, "node3")

	status, info, err := s.client.GetRequestStr("/objectAPI/getNodes")
	fmt.Printf("all nodes names got = %v\n", info)
	if err != nil {
		log.Fatalf("[ERROR]: %v\n", err)
	}
	if status == http.StatusOK {
		nodes := strings.Fields(string(info))

		//// test only !!!
		//nodes := names

		for _, node := range nodes {
			if _, exist := s.nodeList[node]; !exist {
				s.nodeList[node] = 0
			}
		}
		log.Println("[INFO]: Node Init successfully!")
	} else {
		log.Fatal("[ERROR]: Init Node Failed!\n")
	}
}

func (s *Scheduler) watch() {
	s.getInitNode()
	//go s.msgList.Get(defines.PodIdSetPrefix).ConsumeSimple(s.watchNewPod)
	//go s.msgList.Get(defines.AllNodeSetPrefix).ConsumeSimple(s.watchNode)
	select {}
}

func (s *Scheduler) Run() {
	go s.watch()
	go s.dispatch()
	select {}
}

func SelectNodesByNodeSelector(names []string, yamlPod *defines.YamlPod) []*defines.NodeInfo {
	res := make([]*defines.NodeInfo, 0)
	cli := etcd.EtcdStart()
	defer cli.Close()
	infos := make([]*defines.NodeInfo, 0)
	for _, name := range names {
		key := defines.NodePrefix + "/" + name
		kv := etcd.Get(cli, key).Kvs
		tmp := &defines.NodeInfo{}
		_ = yaml.Unmarshal(kv[0].Value, tmp)
		infos = append(infos, tmp)
	}
	for _, info := range infos {
		if info.NodeData.NodeState == defines.NodeNotReady {
			// skip not-ready node(s).
			continue
		}
		if yamlPod.NodeSelector.Gpu == info.NodeData.Selector.Gpu {
			res = append(res, info)
		}
	}
	// res = infos
	return res
}

func SendOutNodeUsageRequest(nodeInfo *defines.NodeInfo) *defines.NodeResourceInfo {
	body, _ := json.Marshal(nodeInfo.NodeData.NodeSpec.Metadata.Name)
	url := "http://" + nodeInfo.NodeData.NodeSpec.Metadata.Ip + ":" + config.WorkerPort + "/objectAPI/getNodeResourceUsage"
	fmt.Printf("scheduler get node resource request = %v\n", url)
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil
	}
	bodyReader := response.Body
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(bodyReader)
	res := &defines.NodeResourceInfo{}
	_ = json.Unmarshal(buf.Bytes(), res)
	return res
}

func SelectNodesByResourceInfo(selectedNodes []*defines.NodeInfo, yamlPod *defines.YamlPod) []*defines.NodeInfo {
	res := make([]*defines.NodeInfo, 0)
	totalMem := 0
	// totalCPU := 0
	for _, con := range yamlPod.Spec.Containers {
		totalMem += utils.ConvertMemStrToInt(con.Resource.ResourceLimit.Memory)
	}
	for _, oneNode := range selectedNodes {
		got := SendOutNodeUsageRequest(oneNode)
		fmt.Printf("got resource info for node %v is: %v\n", oneNode.NodeData.NodeSpec.Metadata.Name, *got)
		freeMem := got.Total - got.Used
		if freeMem <= uint64(totalMem) {
			// res = append(res, oneNode)
			continue
		} else {
			res = append(res, oneNode)
		}
	}
	return res
}

// SchedulePolicy :
// use "NodeSelector" to select nodes with detailed label to decide schedule target.
// consider the pod resource limits and every node's real states.
// consider number of running pods in every selected node.
func (s *Scheduler) SchedulePolicy(yamlPod *defines.YamlPod) string {
	fmt.Printf("get into SchedulePolicy function!\n")
	// get all nodes in the current system.
	s.getInitNode()
	s.NodeNames = make([]string, 0)
	for name, _ := range s.nodeList {
		s.NodeNames = append(s.NodeNames, name)
	}
	fmt.Printf("s.NodeNames = %v\n", s.NodeNames)
	// then select nodes through pod's NodeSelector and node's selector.
	// TODO: in version 2.0, only Gpu label is considered.
	selectedNodes := SelectNodesByNodeSelector(s.NodeNames, yamlPod)
	for _, node := range selectedNodes {
		fmt.Printf("selected1 = %v\n", node.NodeData.NodeSpec.Metadata.Name)
	}
	if len(selectedNodes) == 0 {
		// no node is selected, return "" directly.
		fmt.Println("no satisfied nodes in stage 1!")
		return ""
	}
	// then select again by memory resources.
	selectedNodes1 := SelectNodesByResourceInfo(selectedNodes, yamlPod)
	for _, node := range selectedNodes1 {
		fmt.Printf("selected2 = %v\n", node.NodeData.NodeSpec.Metadata.Name)
	}
	if len(selectedNodes1) == 0 {
		return ""
	}
	// then select a node with the least number of running pod.
	min := math.MaxInt
	minId := 0
	for idx, _ := range selectedNodes1 {
		num := len(selectedNodes1[idx].Pods)
		if num < min {
			min = num
			minId = idx
		}
	}
	return selectedNodes1[minId].NodeData.NodeSpec.Metadata.Name
}

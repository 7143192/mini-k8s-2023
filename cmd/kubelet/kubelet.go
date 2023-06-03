package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"mini-k8s/pkg/apiserver"
	"mini-k8s/pkg/cadvisor"
	"mini-k8s/pkg/config"
	"mini-k8s/pkg/container"
	"mini-k8s/pkg/defines"
	etcd2 "mini-k8s/pkg/etcd"
	"mini-k8s/pkg/pod"
	"mini-k8s/utils"
	yamlTool "mini-k8s/utils/yaml"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

// KubeNode used to store the local node information in one working machine.
// NOTE: this var can NOT be stored into etcd, as only master can interact with etcd.
var KubeNode = defines.NodeInfo{}
var jobs = make([]*defines.EtcdGPUJob, 0)

// node configuration file path.
var path = flag.String("f", "./utils/templates/node_template.yaml", "path of node configuration file.")

//// NewEtcdWatcher used to start a new etcdWatcher to watch node status changes.
//func NewEtcdWatcher() {
//	//points := make([]string, 3)
//	//points[0] = config.IP + ":" + config.EtcdPort1
//	//points[1] = config.IP + ":" + config.EtcdPort2
//	//points[2] = config.IP + ":" + config.EtcdPort3
//	//cli, err := clientv3.New(clientv3.Config{
//	//	Endpoints:   points,
//	//	DialTimeout: 5 * time.Second,
//	//})
//	//if err != nil {
//	//	log.Printf("an error occurs when start an etcd client instance in StartPodWatcher: %v\n", err)
//	//	return
//	//}
//	//defer cli.Close()
//	//log.Println("get here!")
//	//// watchChan := cli.Watch(context.Background(), defines.PodIdSetPrefix+"/")
//	//watchChan := etcd.WatchNew(cli, defines.PodIdSetPrefix+"/")
//	//for resp := range watchChan {
//	//	for _, ev := range resp.Events {
//	//		log.Printf("Type: %s Key:%s Value:%s\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
//	//		newSet := make([]string, 0)
//	//		err := yaml.Unmarshal(ev.Kv.Value, &newSet)
//	//		if err != nil {
//	//			log.Printf("an error occurs when unmarshal json in func NewEtcdWatcher: %v\n", err)
//	//			return
//	//		}
//	//		// do some operations for pods to sync new changes
//	//		// log.Println("get here!")
//	//		pod.HandlePodChanges(newSet, &KubeNode)
//	//	}
//	//}
//
//	// TODO: tmp sol for version 2.0!
//	// KubeNode.EtcdClient = etcd.EtcdStart()
//	// KubeNode.CadvisorClient, _ = cadvisor.CadStart()
//	// KubeNode.NodeData.NodeSpec.Metadata.Name = "node1"
//
//	// cli := KubeNode.EtcdClient
//	cli := etcd2.EtcdStart()
//	defer cli.Close()
//	log.Println("ready to start a new etcdWatcher!")
//	// NOTE: change here in version 2.0: change listening target from "allID/" to "NodePodsList/NODE_NAME"
//	watchKey := defines.NodePodsListPrefix + "/" + KubeNode.NodeData.NodeSpec.Metadata.Name
//	watchChan := etcd2.WatchNew(cli, watchKey)
//	for resp := range watchChan {
//		for _, ev := range resp.Events {
//			log.Printf("Type: %s Key:%s Value:%s\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
//			newSet := make([]string, 0)
//			err := yaml.Unmarshal(ev.Kv.Value, &newSet)
//			if err != nil {
//				log.Printf("an error occurs when unmarshal json in func NewEtcdWatcher: %v\n", err)
//				return
//			}
//			// do some operations for pods to sync new changes
//			// log.Println("get here!")
//			pod.HandlePodChanges(newSet, &KubeNode)
//		}
//	}
//}

func CollectResourceInfo() {
	for {
		//TODO: tmp sol for test in version 2.0!
		etcdClient := etcd2.EtcdStart()
		nodeName := KubeNode.NodeData.NodeSpec.Metadata.Name
		nodeKey := defines.NodePrefix + "/" + nodeName
		kv := etcd2.Get(etcdClient, nodeKey).Kvs
		if len(kv) == 0 {
			log.Printf("[kubelet] the node %v does not exist in the system !\n", nodeName)
			time.Sleep(30 * time.Second)
			continue
		} else {
			_ = yaml.Unmarshal(kv[0].Value, &KubeNode)
			// KubeNode.EtcdClient = etcdClient
			// cadVisorClient, _ := cadvisor.CadStart()
			// KubeNode.CadvisorClient = cadVisorClient
		}

		// cli := KubeNode.CadvisorClient
		cli, _ := cadvisor.CadStart()
		// log.Println("ready to collect resource information!")
		cadvisor.RecordInstanceResource(cli, &KubeNode)
		etcdClient.Close()
		time.Sleep(30 * time.Second)
	}
}

func SendOutNodeState(nodeInfo *defines.NodeInfo) error {
	body, err := json.Marshal(nodeInfo)
	if err != nil {
		return err
	}
	url := "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/updateNodeHealthState"
	log.Printf("[kubelet] update node state url = %v\n", url)
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	//result := make(map[string]string, 0)
	//err = json.NewDecoder(response.Body).Decode(&result)
	//if err != nil {
	//	return err
	//}
	if response.StatusCode != http.StatusOK {
		return errors.New("error when sending new node state")
	} else {
		log.Println("[kubelet] update node state successfully!")
	}
	return nil
}

func SendOutPodState(podInfo *defines.Pod) error {
	body, err := json.Marshal(podInfo)
	if err != nil {
		return err
	}
	url := "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/updatePodHealthState"
	log.Printf("[kubelet] update pod state url = %v\n", url)
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	//result := make(map[string]string, 0)
	//err = json.NewDecoder(response.Body).Decode(&result)
	//if err != nil {
	//	return err
	//}
	if response.StatusCode != http.StatusOK {
		return errors.New("error when sending new pod state")
	} else {
		log.Println("[kubelet] update pod state successfully!")
	}
	return nil
}

func SendOutGetAllRsNames() *defines.AllReplicaPodNames {
	allNames := &defines.AllReplicaPodNames{}
	allNames.Names = make([]string, 0)
	name := "allReplicaPodNames"
	body, err := json.Marshal(&name)
	if err != nil {
		log.Printf("[kubelet] error: %v\n", err)
		return nil
	}
	url := "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/getAllReplicaPodsNames"
	log.Printf("[kubelet] get all replicasNames url = %v\n", url)
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		log.Printf("[kubelet] error: %v\n", err)
		return nil
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Printf("[kubelet] error: %v\n", err)
		return nil
	}
	//result := make(map[string]string, 0)
	//err = json.NewDecoder(response.Body).Decode(&result)
	//if err != nil {
	//	return err
	//}
	if response.StatusCode != http.StatusOK {
		return allNames
	} else {
		// log.Println("[kubelet] update pod state successfully!")
		bodyReader := response.Body
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(bodyReader)
		_ = json.Unmarshal(buf.Bytes(), allNames)
	}
	return allNames
}

func CheckNodePodsStates() {
	// TODO: in version 2.0, just use the node1 as the current node.
	// so there will be some logics about getting related info from etcd.kubectl get
	nodeName := KubeNode.NodeData.NodeSpec.Metadata.Name
	key := defines.NodePrefix + "/" + nodeName
	cli := etcd2.EtcdStart()
	defer cli.Close()
	for {
		kv := etcd2.Get(cli, key).Kvs
		if len(kv) == 0 {
			log.Printf("[kubelet] node %v does not exist in the current system!\n", nodeName)
			time.Sleep(30 * time.Second)
			continue
		} else {
			_ = yaml.Unmarshal(kv[0].Value, &KubeNode)
			// KubeNode.EtcdClient = cli
			// KubeNode.CadvisorClient, _ = cadvisor.CadStart()
		}
		nodeHealth := true

		config.PodMutex.Lock()
		log.Printf("\n[kubelet] kubelet holds the mutex lock here!\n\n")
		// NOTE: get all replicaPods names here.
		allNames := SendOutGetAllRsNames()
		fmt.Printf("allNames = %v\n", allNames.Names)
		for _, podInfo := range KubeNode.Pods {
			//// NOTE: change here change the order to restart a pod when there are something wrong with this pod's containers.
			//fmt.Printf("podInfo state = %v\n", podInfo.PodState)
			//if podInfo.PodState == defines.Failed {
			//	pod.RestartPod(podInfo)
			//}
			if podInfo == nil {
				continue
			}
			if utils.CheckStrInStrList(podInfo.YamlPod.Metadata.Name, allNames.Names) {
				continue
			}
			res := cadvisor.CheckPodContainersStates(podInfo)
			if res == false {
				podInfo.PodState = defines.Failed
				nodeHealth = false
			} else {
				podInfo.PodState = defines.Running
			}
			err1 := SendOutPodState(podInfo)
			if err1 != nil {
				log.Printf("[kubelet] fail to send out new pod state for pod %v!\n", podInfo.Metadata.Name)
			}
		}

		log.Printf("\n[kubelet] kubelet gives up the mutex lock here!\n\n")
		config.PodMutex.Unlock()

		if nodeHealth == true {
			log.Printf("[kubelet] node %v is healthy!\n", nodeName)
		} else {
			log.Printf("[kubelet] node %v is not healthy at all !\n", nodeName)
		}
		//// then update the node info in etcd.
		//KubeNode.EtcdClient = nil
		//KubeNode.CadvisorClient = nil

		//newNodeByte, _ := yaml.Marshal(&KubeNode)
		//etcd2.Put(cli, key, string(newNodeByte))
		_ = SendOutNodeState(&KubeNode)
		time.Sleep(15 * time.Second)
	}
}

func GetNodeConfig() *defines.NodeYaml {
	flag.Parse()
	// log.Printf("path = %v\n", *path)
	// TODO: in version 2.0, we just need to send the yamlNode info to api-server directly?
	yamlNode, _ := yamlTool.ParseNodeConfig(*path)
	// log.Println(yamlNode)
	return yamlNode
}

func PostNodeToMaster(yamlNode *defines.NodeYaml) *http.Response {
	yamlNodeJson, _ := json.Marshal(yamlNode)
	url := "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/registerNewNode"
	resp, err := http.Post(url, "application/json", strings.NewReader(string(yamlNodeJson)))
	if err != nil {
		log.Printf("[kubelet] an error occurs when post yamlNode info to api server: %v\n", err)
		return nil
	}
	return resp
}

// SendOneHeartBeat used to send a new heartBeat.
func SendOneHeartBeat() {
	heartBeat := &defines.NodeHeartBeat{}
	heartBeat.NodeId = KubeNode.NodeData.NodeSpec.Metadata.Name
	key := defines.NodeHeartBeatPrefix + "/" + KubeNode.NodeData.NodeSpec.Metadata.Name
	cli := etcd2.EtcdStart()
	defer cli.Close()
	kv := etcd2.Get(cli, key).Kvs
	if len(kv) == 0 {
		log.Printf("[kubelet] no previous heart beat for node %v is stored in etcd!\n", KubeNode.NodeData.NodeSpec.Metadata.Name)
		heartBeat.CurTime = time.Now()
		heartBeat.LastTime = time.Now()
	} else {
		oldHeartBeat := &defines.NodeHeartBeat{}
		_ = yaml.Unmarshal(kv[0].Value, oldHeartBeat)
		heartBeat.LastTime = oldHeartBeat.CurTime
		heartBeat.CurTime = time.Now()
	}
	// send out a HTTP POST request.
	heartBeatJson, _ := json.Marshal(heartBeat)
	url := "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/heartBeat"
	resp, err := http.Post(url, "application/json", strings.NewReader(string(heartBeatJson)))
	if err != nil {
		log.Printf("[kubelet] an error occurs when post heartBeat info to api server: %v\n", err)
		return
	}
	bodyReader := resp.Body
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(bodyReader)
	log.Printf("[kubelet] response of send heart beat = %v\n", buf.String())
}

// SendHeartBeat used to send heartBeat periodically.
func SendHeartBeat() {
	for {
		log.Println("[kubelet] ready to send a new heart beat!")
		SendOneHeartBeat()
		// send heart beat every 90 seconds.
		time.Sleep(60 * time.Second)
	}
}

func KubeletHandlePodsChange(newSet []string) {
	pod.HandlePodChanges(newSet, &KubeNode)
}

// KubeletServerRestartInit used to restore node info (from etcd) such as pod infos and container states (through docker API).
func KubeletServerRestartInit() {
	for _, singlePod := range KubeNode.Pods {
		// check containers states of this pod and restart exited containers.
		for _, state := range singlePod.ContainerStates {
			inspect := container.InspectContainer(state.Id)
			// if a pod container is not running, just restart this container.
			if inspect.State.Running == false {
				err := container.RestartContainer(state.Id)
				if err != nil {
					log.Printf("[kubelet] can not restart container %v when restart kubelet!\n", state.Id)
					panic(err)
				}
			}
			state.State = defines.Running
		}
		singlePod.PodState = defines.Running
	}
	KubeNode.NodeData.NodeState = defines.NodeReady
	// then send the new node states to apiServer to update etcd info.
	err := SendOutNodeState(&KubeNode)
	if err != nil {
		log.Printf("[kubelet] can not send out new node state to apiServer!\n")
		panic(err)
	}
	log.Printf("[kubelet] re-init kubelet node successfully!\n")
}

func SendOutGetAllJobs() error {
	name := "gpuJobs"
	body, _ := json.Marshal(&name)
	url := "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/getAllGPUJobs"
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	gotData := &defines.AllJobs{}
	bodyReader := response.Body
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(bodyReader)
	_ = json.Unmarshal(buf.Bytes(), gotData)
	fmt.Printf("got all jobs from response = %v\n", *gotData)
	jobs = gotData.Jobs
	return nil
}

func SendOutUpdateOneJobState(job *defines.EtcdGPUJob) error {
	body, _ := json.Marshal(job)
	url := "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/updateGPUJobState"
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		log.Printf("[kubelet] fail tp update GPUJob state!")
		return errors.New("[kubelet] fail tp update GPUJob state")
	}
	return nil
}

func WatchOneJobResult(job *defines.EtcdGPUJob) {
	err := os.Chdir("/home/gpuResults")
	if err != nil {
		log.Printf("dir /home/gpuResults not exist!\n")
		return
	}
	watchDir := "/home/gpuResults/" + job.JobInfo.Name
	err = os.Chdir(watchDir)
	if err != nil {
		log.Printf("dir %v not exist!\n", watchDir)
		return
	}
	cmd := exec.Command("ls")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("[kubelet] error when ls dir %v: %v\n", watchDir, err)
		return
	}
	got := string(output)
	parts := strings.Split(got, "\n")
	//log.Printf("got =\n%v\n", got)
	//log.Printf("parts = %v\n", parts)
	//log.Printf("len of parts = %v\n", len(parts))
	if strings.Contains(got, ".out") && strings.Contains(got, ".err") {
		// a job is finished, check its error file first.
		for _, part := range parts {
			if part == "" {
				continue
			}
			if strings.Contains(part, ".err") {
				cmd1 := exec.Command("cat", part)
				out, err := cmd1.Output()
				if err != nil {
					log.Printf("fail to cat file %v: %v\n", part, err)
					return
				}
				gotContent := string(out)
				if gotContent != "" {
					job.JobState = defines.Failed
				} else {
					job.JobState = defines.Succeed
				}
			}
		}
		// job.JobState = defines.Succeed
		// update state.
		_ = SendOutUpdateOneJobState(job)
	}
}

func WatchAllJobsResult() {
	for {
		err := SendOutGetAllJobs()
		if err != nil {
			log.Printf("fail to get all jobs from apiServer: %v\n", err)
			return
		}
		// check every job's states.
		for _, job := range jobs {
			WatchOneJobResult(job)
		}
		// sleep for 10 seconds.
		time.Sleep(10 * time.Second)
	}
}

// a main function to start a kubelet in a NODE object.
func main() {
	// NOTE: add iptables chain init here !
	//kubeproxy.InitSvcMainChain()

	// Change in version 2.0: combine collectResource, checkNodePodsStates and etcdWatcher together in kubelet layer.
	// create a new node here.
	// TODO: in version 2.0, how to know running states of the node in this local machine ? (done: through HeartBeat. )
	// get new node Info / restore node Info (after restart or register a node repeatedly. )
	yamlNode := GetNodeConfig()
	resp := PostNodeToMaster(yamlNode)
	log.Printf("%v\n", resp)
	// handle response body here.
	bodyReader := resp.Body
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(bodyReader)
	nodeInfoJson := buf.Bytes()
	_ = json.Unmarshal(nodeInfoJson, &KubeNode)
	log.Printf("[kubelet] kubenode after register = %v\n", KubeNode)
	// try to re-init kubelet node (after restart)
	if KubeNode.Registered == true {
		KubeletServerRestartInit()
	}
	// send node's heart beat.
	go SendHeartBeat()
	// collect node's resource info.
	go CollectResourceInfo()
	// NOTE: comment this line to do bug-test!(2023-5-6 by lyh)
	// check all pods' states of this node.
	go CheckNodePodsStates()

	// a go routine to watch all jobs' running states.
	go WatchAllJobsResult()

	// in version 2.0, start a http server in kubelet layer to communicate with api server.
	kubeletServer := apiserver.KubeletServerInit()
	// store th current node info into server to simplify server handler.
	kubeletServer.NodeInfo = &KubeNode
	err := kubeletServer.KubeletServerRun()
	if err != nil {
		panic(err)
	}
}

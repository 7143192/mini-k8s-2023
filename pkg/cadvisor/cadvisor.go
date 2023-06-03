package cadvisor

import (
	"bytes"
	"encoding/json"
	"github.com/google/cadvisor/client"
	v1 "github.com/google/cadvisor/info/v1"
	"log"
	"mini-k8s/pkg/config"
	"mini-k8s/pkg/container"
	defines2 "mini-k8s/pkg/defines"
	etcd2 "mini-k8s/pkg/etcd"
	"mini-k8s/pkg/pod"
	"mini-k8s/utils"
	"net/http"
	"time"
)

func CadStart() (*client.Client, error) {
	addr := "http://" + config.IP + ":" + config.CadvisorPort + "/"
	cli, err := client.NewClient(addr)
	return cli, err
}

//func Test(cli *client.Client) []v1.ContainerInfo {
//	req := v1.DefaultContainerInfoRequest()
//	allInfo, err := cli.AllDockerContainers(&req)
//	if err != nil {
//		fmt.Printf("an error occurs when getting all containers info with cadvisor: %v\n", err)
//		return nil
//	}
//	return allInfo
//}

func GetOneContainerInfo(cli *client.Client, containerID string) *v1.ContainerInfo {
	req := v1.DefaultContainerInfoRequest()
	// fmt.Printf("cli = %v\n", cli)
	allInfo, err := cli.AllDockerContainers(&req)
	if err != nil {
		log.Printf("[cadvisor] an error occurs when getting all containers info with cadvisor: %v\n", err)
		return nil
	}
	for _, info := range allInfo {
		if containerID == info.Id {
			return &info
		}
	}
	return nil
}

func CadEnd() {
	return
}

// RecordInstanceResource used to collect all instances resource info in one NODE.
// change in version 2.0: only record pods belong to the current node.
func RecordInstanceResource(cadCli *client.Client, nodeInfo *defines2.NodeInfo) {
	//// version 1.0: get all pods according to allID/ val.
	//cli := etcd.EtcdStart()
	//all := etcd.Get(cli, defines.PodIdSetPrefix+"/").Kvs
	//if len(all) == 0 {
	//	return
	//}
	//allId := make([]string, 0)
	//err := yaml.Unmarshal(all[0].Value, &allId)
	//if err != nil {
	//	fmt.Printf("an error occurs when unmarshal all ids in func RecordInstanceResource:%v \n", err)
	//	return
	//}
	//allPod := make([]defines.Pod, len(allId))
	//for _, id := range allId {
	//	single := etcd.Get(cli, id).Kvs
	//	pod := defines.Pod{}
	//	err = yaml.Unmarshal(single[0].Value, &pod)
	//	allPod = append(allPod, pod)
	//}
	//for _, singlePod := range allPod {
	//	// record every pod's resource.
	//	RecordPodInstanceResource(&singlePod, cadCli)
	//}
	if nodeInfo.NodeData.NodeSpec.Metadata.Name == "" {
		log.Println("[cadvisor] WARN: nodeName is empty!")
		return
	}
	all := nodeInfo.Pods
	if len(all) == 0 {
		log.Println("[cadvisor] current node does not have a pod!")
		// NOTE: in version 3.0, do not return here directly!
		// If return here, the init state of one node CAN NOT be gotten!

		// return
	}
	podsResources := make([]*defines2.PodResourceUsed, 0)
	for _, podInfo := range all {
		resourceInfo := RecordPodInstanceResource(podInfo, cadCli)
		podsResources = append(podsResources, resourceInfo)
	}
	RecordNodeInstanceResource(podsResources, nodeInfo)
}

// RecordPodInstanceResource used to record resource of one single pod.
func RecordPodInstanceResource(pod *defines2.Pod, cadVisorCli *client.Client) *defines2.PodResourceUsed {
	// NOTE: the Stats recorded by cadVisor is serial, similar to time-serial DB.
	if pod.PodState != defines2.Running {
		// no need to record information for pod instance that is not running.
		return nil
	}
	podResourceUsed := &defines2.PodResourceUsed{}
	podResourceUsed.PodName = pod.Metadata.Name
	memUsed := uint64(0)
	cpuUsed := uint64(0)
	// not sure here: don't need to consider pause container?
	containerStates := make([]*defines2.ContainerResourceUsed, 0)
	for idx, con := range pod.ContainerStates {
		containerState := &defines2.ContainerResourceUsed{}
		// fmt.Printf("con = %v\n", con)
		info := GetOneContainerInfo(cadVisorCli, con.Id)
		// fmt.Printf("one container info = %v\n", *info)
		if info == nil {
			continue
		}
		length := len(info.Stats)
		if length == 0 || length == 1 {
			continue
		}
		mem := info.Stats[length-1].Memory.Usage
		// spentTime := utils.GetTime(info.Stats[length-1].Timestamp) - utils.GetTime(info.Stats[length-2].Timestamp)
		spentTime := uint64(info.Stats[length-1].Timestamp.Unix() - info.Stats[length-2].Timestamp.Unix())
		cpuTime := info.Stats[length-1].Cpu.Usage.Total - info.Stats[length-2].Cpu.Usage.Total
		curTime := time.Now()
		containerState.MemUsed = mem
		percent := float64(cpuTime*100) / float64(spentTime*1e9)
		// debug
		log.Printf("cpuTime = %v, spentTime = %v\n", cpuTime, spentTime)
		log.Printf("percent of CPU: %v\n", percent)
		totalCPUofCon := 0.0
		if idx > 0 {
			containerInfo := pod.YamlPod.Spec.Containers[idx-1]
			cpuLimit := utils.ParseCPUInfo(containerInfo.Resource.ResourceLimit.Cpu)
			totalCPUofCon = float64(cpuLimit/100) * percent
		}
		// containerState.CpuUsed = cpuTime / spentTime
		containerState.CpuUsed = uint64(totalCPUofCon)
		containerState.TimeUsed = curTime
		// store single container resource information.
		containerStates = append(containerStates, containerState)
		memUsed += mem
		cpuUsed += containerState.CpuUsed
	}
	// store th information for the whole pod instance.
	podResourceUsed.CpuUsed = cpuUsed
	podResourceUsed.MemUsed = memUsed
	podResourceUsed.TimeUsed = time.Now()
	// test
	log.Printf("[cadvisor] Pod ResourceInfo:\n")
	log.Printf("[cadvisor] cpuUsed = %v, memUsed = %v, time = %v\n", cpuUsed, memUsed, podResourceUsed.TimeUsed)
	// then put these information into etcd storage.
	etcdCli := etcd2.EtcdStart()
	defer etcdCli.Close()
	//podResourceByte, _ := yaml.Marshal(podResourceUsed)
	//containersResourceByte, _ := yaml.Marshal(containerStates)
	//totalKey := defines2.PodResourceStatePrefix + "/" + pod.Metadata.Name
	//containersKey := defines2.PodContainersResourcePrefix + "/" + pod.Metadata.Name
	//etcd2.Put(etcdCli, totalKey, string(podResourceByte))
	//etcd2.Put(etcdCli, containersKey, string(containersResourceByte))

	// NOTE: in version 2.0, move the cadvisor etcd operations to api-server through HTTP.
	_ = SendOutStorePodResource(podResourceUsed)
	return podResourceUsed
}

func RecordNodeInstanceResource(podsResources []*defines2.PodResourceUsed, nodeInfo *defines2.NodeInfo) {
	cli := etcd2.EtcdStart()
	defer cli.Close()
	nodeResource := &defines2.NodeResourceUsed{}
	nodeResource.NodeId = nodeInfo.NodeData.NodeSpec.Metadata.Name
	totalMem := uint64(0)
	totalCpu := uint64(0)
	usedMem := uint64(0)
	usedCpu := uint64(0)
	for _, resourceInfo := range podsResources {
		if resourceInfo == nil {
			continue
		}
		usedMem += resourceInfo.MemUsed
		usedCpu += resourceInfo.CpuUsed
	}
	for _, podInfo := range nodeInfo.Pods {
		if podInfo == nil {
			continue
		}
		_, limCpu := utils.GetTotalCPU(podInfo)
		totalCpu += uint64(limCpu * 100)
		_, limMem := utils.GetTotalMem(podInfo)
		totalMem += uint64(utils.ConvertMemStrToInt(limMem))
	}
	nodeResource.CpuUsed = usedCpu
	nodeResource.MemUsed = usedMem
	nodeResource.CpuTotal = totalCpu
	nodeResource.MemTotal = totalMem
	nodeResource.Time = time.Now()
	log.Printf("[cadvisor] Node resourceInfo: \n")
	log.Printf("[cadvisor] cpuUsed: %v, memUsed: %v, cpuTotal: %v, memTotal: %v\n", nodeResource.CpuUsed,
		nodeResource.MemUsed, nodeResource.CpuTotal, nodeResource.MemTotal)
	//key := defines2.NodeResourcePrefix + "/" + nodeInfo.NodeData.NodeSpec.Metadata.Name
	//val, _ := yaml.Marshal(nodeResource)
	//// store resource info into etcd.
	//etcd2.Put(cli, key, string(val))

	// NOTE: in version 2.0, move the cadvisor etcd operations to api-server through HTTP.
	_ = SendOutNodeResource(nodeResource)
}

// CheckPodContainersStates just simply make every stopped
// container as a Failed one? But what if a container is
// exited normally?
// If return True, this pod runs well(or some containers finish
// successfully), else this pod crashes(maybe some containers exits with error.)
func CheckPodContainersStates(podInfo *defines2.Pod) bool {
	//cadCli, _ := CadStart()
	//req := v1.DefaultContainerInfoRequest()
	//allInfo, _ := cadCli.AllDockerContainers(&req)
	//infos := make([]string, 0)
	//for _, info := range allInfo {
	//	infos = append(infos, info.Id)
	//}
	healthy := true
	for idx, con := range podInfo.ContainerStates {
		// hack here ???
		curId := con.Id
		if curId == "" {
			// not finish creating.
			log.Println("[cadvisor] this container does not finish creating process!")
			continue
		}
		status := container.InspectContainer(curId)
		if status.ContainerJSONBase == nil {
			log.Printf("get nothing when inspect container %v\n!", curId)
			// continue
			// NOTE: change here to "return" directly? (not sure)
			return true
		}
		if status.State.Status == "running" {
			// a well-running container
			podInfo.ContainerStates[idx].State = defines2.Running
			continue
		} else {
			if status.State.ExitCode == 0 {
				// a container that stops successfully.
				podInfo.ContainerStates[idx].State = defines2.Succeed
				continue
			}
			if status.State.ExitCode != 0 {
				log.Printf("[cadvisor] pod %v failed! the error container is: %v!\n", podInfo.Metadata.Name, con.Name)
				podInfo.ContainerStates[idx].State = defines2.Failed
				healthy = false
				continue
			}
		}
	}
	if healthy == false {
		// if this pod runs wrong, restart this pod directly.
		pod.RestartPod(podInfo)
	}
	return healthy
}

func CheckReplicaSetPodContainersState(podInfo *defines2.Pod) bool {
	//cadCli, _ := CadStart()
	//req := v1.DefaultContainerInfoRequest()
	//allInfo, _ := cadCli.AllDockerContainers(&req)
	//infos := make([]string, 0)
	//for _, info := range allInfo {
	//	infos = append(infos, info.Id)
	//}
	healthy := true
	for idx, con := range podInfo.ContainerStates {
		// hack here ???
		curId := con.Id
		if curId == "" {
			// not finish creating.
			log.Println("[cadvisor] this container does not finish creating process!")
			continue
		}
		status := container.InspectContainer(curId)
		if status.ContainerJSONBase == nil {
			log.Printf("get nothing when inspect container %v\n!", curId)
			// continue
			// NOTE: change here to "return" directly? (not sure)
			return true
		}
		if status.State.Status == "running" {
			// a well-running container
			podInfo.ContainerStates[idx].State = defines2.Running
			continue
		} else {
			if status.State.ExitCode == 0 {
				// a container that stops successfully.
				podInfo.ContainerStates[idx].State = defines2.Succeed
				continue
			}
			if status.State.ExitCode != 0 {
				log.Printf("[cadvisor] pod %v failed! the error container is: %v!\n", podInfo.Metadata.Name, con.Name)
				podInfo.ContainerStates[idx].State = defines2.Failed
				healthy = false
				continue
			}
		}
	}
	return healthy
}

func SendOutStorePodResource(podResource *defines2.PodResourceUsed) error {
	body, _ := json.Marshal(podResource)
	url := "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/storePodResource"
	log.Printf("[cadvisor] store pod resource request = %v\n", url)
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Printf("[cadvisor] error when sendOutStorePodResource: %v\n", err)
		return err
	}
	if response.StatusCode == http.StatusOK {
		log.Println("[cadvisor] Store pod resource info successfully!")
		return nil
	}
	return nil
}

func SendOutNodeResource(nodeResource *defines2.NodeResourceUsed) error {
	body, _ := json.Marshal(nodeResource)
	url := "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/storeNodeResource"
	log.Printf("[cadvisor] store node resource request = %v\n", url)
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Printf("[cadvisor] error when sendOutNodeResource: %v\n", err)
		return err
	}
	if response.StatusCode == http.StatusOK {
		log.Println("[cadvisor] Store node resource info successfully!")
		return nil
	}
	return nil
}

//func CheckPodContainersStates(podInfo *defines2.Pod) bool {
//	cadCli, _ := CadStart()
//	req := v1.DefaultContainerInfoRequest()
//	allInfo, _ := cadCli.AllDockerContainers(&req)
//	infos := make([]string, 0)
//	for _, info := range allInfo {
//		infos = append(infos, info.Id)
//	}
//	healthy := true
//	for idx, con := range podInfo.ContainerStates {
//		checkRes := utils.CheckStrInStrList(con.Name, infos)
//		if checkRes == true {
//			podInfo.ContainerStates[idx].State = defines2.Running
//		} else {
//			podInfo.ContainerStates[idx].State = defines2.Failed
//			healthy = false
//		}
//	}
//	if healthy == false {
//		// if this pod runs wrong, restart this pod directly.
//		pod.RestartPod(podInfo)
//	}
//	return healthy
//}

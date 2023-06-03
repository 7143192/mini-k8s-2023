package apiserver

import (
	"encoding/json"
	"fmt"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v3"
	"mini-k8s/pkg/defines"
	"mini-k8s/pkg/etcd"
	"mini-k8s/utils"
	"time"
)

func DescribeNode(client *clientv3.Client, nodeName string) string {
	finalRes := &defines.NodeResourceSend{}
	// get node info from etcd first.
	key := defines.NodePrefix + "/" + nodeName
	nodeInfo := &defines.NodeInfo{}
	kv := etcd.Get(client, key).Kvs
	if len(kv) == 0 {
		// fmt.Printf("node %v does not exist in the system!\n", nodeName)
		return fmt.Sprintf("node %v does not exist in the system!\n", nodeName)
	}
	_ = yaml.Unmarshal(kv[0].Value, nodeInfo)
	fmt.Println("Node Information:")
	// then get resource info from etcd.
	resourceKey := defines.NodeResourcePrefix + "/" + nodeName
	kv = etcd.Get(client, resourceKey).Kvs
	if len(kv) == 0 {
		fmt.Printf("resource information for node %v does not be collected!\n", nodeName)
		return fmt.Sprintf("resource information for node %v does not be collected!\n", nodeName)
	}
	nodeResource := &defines.NodeResourceUsed{}
	_ = yaml.Unmarshal(kv[0].Value, nodeResource)
	// then describe this node itself.
	readyNum := 0
	if nodeInfo.NodeData.NodeState == defines.NodeReady {
		readyNum = 1
	} else {
		readyNum = 0
	}
	nodeId := nodeInfo.NodeData.NodeId
	memUsed := utils.ConvertMemTotalToStr(int(nodeResource.MemUsed))
	memTotal := utils.ConvertMemTotalToStr(int(nodeResource.MemTotal))
	cpuUsed := float64(nodeResource.CpuUsed)
	cpuTotal := float64(nodeResource.CpuTotal)
	// consider the case that there is only one pause container(no mem and cpu limit provided.)
	totalConNum := 0
	for _, pod := range nodeInfo.Pods {
		totalConNum += len(pod.ContainerStates)
	}
	if totalConNum == len(nodeInfo.Pods) {
		memUsed = "0B"
		memTotal = "0B"
		cpuUsed = 0.0
		cpuTotal = 0.0
	}
	// fmt.Printf("NAME\t\t\tID\t\t\tREADY\t\t\tMEMORY\t\t\tCPU\t\t\tIP\n")
	// fmt.Printf("%v\t\t\t%v\t\t\t%v/1\t\t\t%v/%v\t\t\t%v/%v\t\t\t%v\n", nodeName, nodeId, readyNum, memUsed, memTotal, cpuUsed,
	// 	cpuTotal, nodeInfo.NodeData.NodeSpec.Metadata.Ip)
	finalRes.NodeName = nodeName
	finalRes.NodeId = nodeId
	finalRes.ReadyNum = readyNum
	finalRes.MemUsed = memUsed
	finalRes.MemTotal = memTotal
	finalRes.CpuUsed = cpuUsed
	finalRes.CpuTotal = cpuTotal
	finalRes.NodeIP = nodeInfo.NodeData.NodeSpec.Metadata.Ip
	// then describe the pods that belong to this node.
	fmt.Println("Pods Information:")
	for _, pod := range nodeInfo.Pods {
		gotStr := DescribePod(client, pod.Metadata.Name, true)
		gotPodInfo := &defines.PodResourceSend{}
		_ = json.Unmarshal([]byte(gotStr), gotPodInfo)
		finalRes.PodsResourcesSend = append(finalRes.PodsResourcesSend, gotPodInfo)
	}
	finalResByte, _ := json.Marshal(finalRes)
	return string(finalResByte)
}

func GetNodeResourceInfo() *defines.NodeResourceInfo {
	res := &defines.NodeResourceInfo{}
	v, _ := mem.VirtualMemory()
	res.Total = v.Total
	res.Used = v.Used
	percents, _ := cpu.Percent(10*time.Second, false)
	res.CpuPercent = percents[0]
	return res
}

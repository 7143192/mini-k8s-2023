package apiserver

import (
	"encoding/json"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v3"
	defines2 "mini-k8s/pkg/defines"
	"mini-k8s/pkg/etcd"
	"mini-k8s/utils"
	"strings"
	"time"
)

func DescribeContainers(client *clientv3.Client, name string) []*defines2.ContainerResourceSend {
	finalRes := make([]*defines2.ContainerResourceSend, 0)
	key := "PodInstance/" + name
	kv := etcd.Get(client, key).Kvs
	if len(kv) == 0 {
		fmt.Printf("No pod named %s !\n", name)
		return finalRes
	}
	val := kv[0].Value
	pod := &defines2.Pod{}
	err := yaml.Unmarshal(val, pod)
	if err != nil {
		fmt.Printf("an error occurs when describe pod %s!\n", name)
		return finalRes
	}

	// fmt.Println("NAME\t\t\tREADY\t\t\tSTATUS\t\t\tRESTART\t\t\tTIME")

	// TODO: in version 1.0, restart num is always 0 as we don't implement corresponding functions.
	for idx, con := range pod.ContainerStates {
		res := &defines2.ContainerResourceSend{}
		containerStatus := pod.ContainerStates[idx]
		conName := con.Name
		if strings.Index(conName, "pause") >= 0 {
			conName = "pause"
		}
		readyNum := utils.GetReadyNumber(containerStatus)
		curState := utils.ConvertContainerStateToStr(containerStatus)
		// curState := "Pending"
		var totalTime uint64 = 0
		if curState == "Running" {
			totalTime = utils.GetTime(time.Now()) - utils.GetTime(pod.Start)
			res.ConName = conName
			res.ReadyNum = readyNum
			res.CurState = curState
			res.TotalTime = totalTime
			// fmt.Printf("%s\t\t\t%d/1\t\t\t%s\t\t\t0\t\t\t%vs\n", conName, readyNum, curState, totalTime)

		} else {
			// fmt.Printf("%s\t\t\t%d/1\t\t\t%s\t\t\t0\t\t\t---\n", conName, readyNum, curState)
			res.ConName = conName
			res.ReadyNum = readyNum
			res.CurState = curState
			res.TotalTime = 0
		}
		finalRes = append(finalRes, res)
	}
	return finalRes
}

func DescribePod(client *clientv3.Client, name string, isNode bool) string {
	finalRes := &defines2.PodResourceSend{}
	key := defines2.PodInstancePrefix + "/" + name
	kv := etcd.Get(client, key).Kvs
	resourceKey := defines2.PodResourceStatePrefix + "/" + name
	resource := etcd.Get(client, resourceKey).Kvs
	if len(kv) == 0 {
		// fmt.Printf("No pod named %s!\n", name)
		return fmt.Sprintf("No pod named %s!\n", name)
	}
	// corner case.
	if len(resource) == 0 {
		// fmt.Printf("Pod %s doesn't collect resource for now!\n", name)
		return fmt.Sprintf("Pod %s doesn't collect resource for now!\n", name)
	}
	resourceVal := resource[0].Value
	val := kv[0].Value
	pod := &defines2.Pod{}
	podResource := &defines2.PodResourceUsed{}
	err := yaml.Unmarshal(val, pod)
	if err != nil {
		// fmt.Printf("an error occurs when describing pod %s!\n", name)
		return fmt.Sprintf("an error occurs when describing pod %s!\n", name)
	}
	err = yaml.Unmarshal(resourceVal, podResource)
	if err != nil {
		// fmt.Printf("an error occurs when unmarshaling pod resource %s!\n", name)
		return fmt.Sprintf("an error occurs when unmarshaling pod resource %s!\n", name)
	}
	// pod-info header line
	if isNode == false {
		// fmt.Println("Pod Information:")
	}
	// fmt.Println("NAME\t\tPOD IP\tMEM REQ\t\tMEM LIMIT\tCPU REQ\t\tCPU LIMIT\tAGE")

	// change here.
	_, limMem := utils.GetTotalMem(pod)
	// _, limCPU := utils.GetTotalCPU(pod)
	_, limCPU := utils.GetDescribeCPU(pod)

	if pod.PodState == defines2.Running {
		continueTime := utils.GetTime(time.Now()) - utils.GetTime(pod.Start)
		memUsed := utils.ConvertMemTotalToStr(int(podResource.MemUsed))
		cpuUsed := podResource.CpuUsed
		// fmt.Printf("%v\t\t%v\t\t%v\t\t%v\t\t%v\t\t%v\t\t%vs\n", name, pod.PodIp, memUsed, limMem, cpuUsed, limCPU, continueTime)
		finalRes.PodName = name
		finalRes.PodIP = pod.PodIp
		finalRes.MemUsed = memUsed
		finalRes.LimMem = limMem
		finalRes.CpuUsed = cpuUsed
		finalRes.LimCpu = limCPU
		finalRes.ContinueTime = continueTime
		finalRes.RestartNum = pod.RestartNum
	} else {
		// fmt.Printf("%v\t\t%v\t\t---\t\t%v\t\t---\t\t%v\t\t---s\n", name, pod.PodIp, limMem, limCPU)
		finalRes.PodName = name
		finalRes.PodIP = "--.--.--.--"
		finalRes.MemUsed = "---"
		finalRes.LimMem = limMem
		finalRes.CpuUsed = 0
		finalRes.LimCpu = limCPU
		finalRes.ContinueTime = 0
		finalRes.RestartNum = pod.RestartNum
	}
	if isNode == false {
		// fmt.Println("Container information of this pod:")
		got := DescribeContainers(client, name)
		finalRes.ContainerResourcesSend = got
	}
	finalResByte, _ := json.Marshal(finalRes)
	fmt.Printf("resource of one pod before sending: %v\n", string(finalResByte))
	return string(finalResByte)
}

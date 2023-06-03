package apiserver

import (
	"encoding/json"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v3"
	defines2 "mini-k8s/pkg/defines"
	"mini-k8s/pkg/etcd"
	"mini-k8s/utils"
	"time"
)

func GetPod(cli *clientv3.Client) string {
	res := &defines2.GetPods{}
	kv := etcd.GetWithPrefix(cli, "PodInstance/").Kvs
	if len(kv) == 0 {
		// fmt.Println("No pods in the current mini-k8s system!")
		return "No pods in the current mini-k8s system!"
	}
	// fmt.Println("all pods brief information:")
	// fmt.Println("NAME\t\tREADY\t\tSTATUS\t\tRESTART\t\tAGE")
	for _, oneKv := range kv {
		tmp := &defines2.GetPodsSend{}
		val := oneKv.Value
		pod := &defines2.Pod{}
		err := yaml.Unmarshal(val, pod)
		if err != nil {
			// fmt.Printf("an error occurs when unmarshal pod in GetPod function:%v\n", err)
			return fmt.Sprintf("an error occurs when unmarshal pod in GetPod function:%v\n", err)
		}
		curState := utils.ConvertIntToState(pod.PodState)
		readyNum := utils.GetReadyNumberFromInt(pod.PodState)
		// NOTE: in version 2.0, use the real restart number to return to kubectl.
		restartNum := pod.RestartNum
		if pod.PodState == defines2.Running {
			continueTime := utils.GetTime(time.Now()) - utils.GetTime(pod.Start)
			// fmt.Printf("%s\t\t%v/1\t\t%s\t\t%v\t\t%vs\n", pod.Metadata.Name, readyNum, curState, restartNum, continueTime)
			tmp.Name = pod.Metadata.Name
			tmp.CurState = curState
			tmp.ReadyNum = readyNum
			tmp.RestartNum = restartNum
			tmp.ContinueTime = continueTime
			tmp.Ip = pod.PodIp
		} else {
			// fmt.Printf("%s\t\t%v/1\t\t%s\t\t%v\t\t---\n", pod.Metadata.Name, readyNum, curState, restartNum)
			tmp.Name = pod.Metadata.Name
			tmp.CurState = curState
			tmp.ReadyNum = readyNum
			tmp.RestartNum = restartNum
			tmp.ContinueTime = 0
			tmp.Ip = pod.PodIp
		}
		res.PodsSend = append(res.PodsSend, tmp)
	}
	resByte, _ := json.Marshal(res)
	return string(resByte)
}

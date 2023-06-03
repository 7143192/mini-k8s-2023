package pod

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"mini-k8s/pkg/config"
	"mini-k8s/pkg/defines"
	etcd2 "mini-k8s/pkg/etcd"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

// MakeNewHostResultDir NOTE: add this function to check hostResultDir status.
func MakeNewHostResultDir(name string) error {
	// cd to /home dir first.
	//cmd0 := exec.Command("cd", "~")
	//err := cmd0.Run()
	err := os.Chdir("/home")
	if err != nil {
		log.Printf("fail to cd to /home dir: %v\n", err)
		return err
	}
	// then use ls to check whether the rootResultDir has already existed.
	cmd1 := exec.Command("ls")
	output, err := cmd1.Output()
	if err != nil {
		log.Printf("fail to ls !\n")
		return err
	}
	got1 := string(output)
	if strings.Contains(got1, "gpuResults") == false {
		cmd2 := exec.Command("mkdir", "gpuResults")
		err = cmd2.Run()
		if err != nil {
			log.Printf("fail to mkdir gpuResults!\n")
			return err
		}
	}
	// then cd to gpuResults dir.
	//cmd3 := exec.Command("cd", "gpuResults")
	//err = cmd3.Run()
	err = os.Chdir("gpuResults")
	if err != nil {
		log.Printf("fail to cd to gpuResults dir!\n")
		return err
	}
	// every should have a private dir to contain its own results and errors.
	dirName := name
	// ls again to check existence.
	cmd4 := exec.Command("ls")
	output1, err := cmd4.Output()
	if err != nil {
		log.Printf("fail to ls in dir gpuResults!\n")
		return err
	}
	got2 := string(output1)
	if strings.Contains(got2, dirName) == true {
		// if exists....
		//cmd5 := exec.Command("cd", dirName)
		//err = cmd5.Run()
		err = os.Chdir(dirName)
		if err != nil {
			log.Printf("fail to cd to job private dir !\n")
			return err
		}
		// then remove everything under this job-private directory.
		cmd6 := exec.Command("rm", "-rf", "./*")
		err = cmd6.Run()
		if err != nil {
			log.Printf("fail to clean every thing under a job-pivate dir!\n")
			return err
		}
	} else {
		// if not exists...
		cmd7 := exec.Command("mkdir", dirName)
		err = cmd7.Run()
		if err != nil {
			log.Printf("fail to mkdir for job %v!\n", name)
			return err
		}
	}
	return nil
}

func SendOutNodeInfoAfterRealOp(nodeInfo *defines.NodeInfo) error {
	body, _ := json.Marshal(nodeInfo)
	url := "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/updateNodeInfoAfterRealOp"
	log.Printf("[apiserver] update node info request after real op = %v\n", url)
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
		return nil
	}
	return errors.New("[realPod] error when update node info after real ops are made")
}

// HandlePodChanges used to do changes to the pods belong to current node to sync with the state in etcd.
func HandlePodChanges(newSet []string, nodeInfo *defines.NodeInfo) *defines.HandlePodResult {
	log.Printf("old node info = %v\n", nodeInfo)
	// version 2.0: add node, so here old pod can be got from the un-updated node info of the current worker machine.
	changedPods := &defines.HandlePodResult{}
	//nodeInfo.EtcdClient = etcd2.EtcdStart()
	//// nodeInfo.CadvisorClient, _ = cadvisor.CadStart()
	//cli := nodeInfo.EtcdClient
	cli := etcd2.EtcdStart()
	defer cli.Close()
	oldSet := make([]string, 0)
	for _, instance := range nodeInfo.Pods {
		oldName := defines.PodInstancePrefix + "/" + instance.Metadata.Name
		oldSet = append(oldSet, oldName)
	}
	newAdd, newDel := etcd2.GetPodsChanges(newSet, oldSet)
	log.Printf("newAdd = %v\n", newAdd)
	log.Printf("newDel = %v\n", newDel)
	if len(newAdd) == 0 && len(newDel) == 0 {
		fmt.Println("no new pod is added or old pod is deleted!")
		return changedPods
	}
	// TODO: as we don't have node object for now, we just create or del pod and don't need tpo update node information.
	for _, del := range newDel {
		log.Printf("a old pod %s is deleted!\n", del)
		delPod := DelRealPod(cli, del)
		changedPods.Del = append(changedPods.Del, delPod)
	}
	for _, add := range newAdd {
		log.Printf("a new pod %s is added!\n", add)
		// get pod info here to check whether this pod is gpu-related pod.
		kv := etcd2.Get(cli, add).Kvs
		if len(kv) != 0 {
			tmp := &defines.Pod{}
			_ = yaml.Unmarshal(kv[0].Value, tmp)
			// gpu-related pods should have "GPUJobPod_" name prefix.
			// make corresponding dir first.
			// NOTE: hack here !!!
			if strings.Contains(tmp.YamlPod.Metadata.Name, defines.GPUJobPodNamePrefix+"_") == true {
				i := strings.Index(tmp.YamlPod.Metadata.Name, "_")
				jobName := tmp.YamlPod.Metadata.Name[i+1:]
				_ = MakeNewHostResultDir(jobName)
			}
		}
		addedPod := CreateRealPod(cli, add)
		changedPods.Add = append(changedPods.Add, addedPod)
	}
	// then really add / remove these instance from local node's pods list.
	for _, realDel := range newDel {
		id := -1
		for idx, one := range nodeInfo.Pods {
			targetName := defines.PodInstancePrefix + "/" + one.Metadata.Name
			log.Printf("target del pod name = %s\n", targetName)
			if realDel == targetName {
				id = idx
				break
			}
		}
		log.Printf("id that try del = %v\n", id)
		if id != -1 {
			if id == 0 {
				nodeInfo.Pods = nodeInfo.Pods[1:]
			} else {
				if id == len(nodeInfo.Pods)-1 {
					nodeInfo.Pods = nodeInfo.Pods[0 : len(nodeInfo.Pods)-1]
				} else {
					nodeInfo.Pods = append(nodeInfo.Pods[0:id], nodeInfo.Pods[id+1:]...)
				}
			}
		}
	}
	for _, realAdd := range newAdd {
		// get newly added instance from etcd.
		addPod := etcd2.Get(cli, realAdd).Kvs
		addPodVal := &defines.Pod{}
		_ = yaml.Unmarshal(addPod[0].Value, addPodVal)
		nodeInfo.Pods = append(nodeInfo.Pods, addPodVal)
	}
	// the store the new nodeInfo into etcd.
	//// NOTE: hack here.
	//oldCadCli := nodeInfo.CadvisorClient
	//oldEtcdCli := nodeInfo.EtcdClient
	//nodeInfo.CadvisorClient = nil
	//nodeInfo.EtcdClient = nil

	//// TODO: move the etcd o from kubelet to api-server.
	//newCli := etcd2.EtcdStart()
	//newKey := defines.NodePrefix + "/" + nodeInfo.NodeData.NodeSpec.Metadata.Name
	//newNodeByte, _ := yaml.Marshal(nodeInfo)
	//etcd2.Put(newCli, newKey, string(newNodeByte))
	////nodeInfo.CadvisorClient = oldCadCli
	////nodeInfo.EtcdClient = oldEtcdCli
	//newCli.Close()
	_ = SendOutNodeInfoAfterRealOp(nodeInfo)
	// as this etcd client can be shared, DO NOT close it here !!!
	return changedPods
}

func HandlePodChangesNew(newSet []string, nodeInfo *defines.NodeInfo) *defines.HandlePodNameResult {
	oldSet := make([]string, 0)
	for _, instance := range nodeInfo.Pods {
		oldName := defines.PodInstancePrefix + "/" + instance.Metadata.Name
		oldSet = append(oldSet, oldName)
	}
	newAdd, newDel := etcd2.GetPodsChanges(newSet, oldSet)
	res := &defines.HandlePodNameResult{}
	res.Del = newDel
	res.Add = newAdd
	return res
}

func HandleCurNodeRealChanges(result *defines.HandlePodResult, nodeInfo *defines.NodeInfo) {
	// handle real-deleted pods.
	for _, newDel := range result.Del {
		targetName := defines.PodInstancePrefix + "/" + newDel.Metadata.Name
		delId := -1
		for idx, _ := range nodeInfo.Pods {
			curName := defines.PodInstancePrefix + "/" + nodeInfo.Pods[idx].Metadata.Name
			if curName == targetName {
				delId = idx
				break
			}
		}
		if delId != -1 {
			if delId == 0 {
				nodeInfo.Pods = nodeInfo.Pods[1:]
			} else {
				if delId == len(nodeInfo.Pods)-1 {
					nodeInfo.Pods = nodeInfo.Pods[0 : len(nodeInfo.Pods)-1]
				} else {
					nodeInfo.Pods = append(nodeInfo.Pods[0:delId], nodeInfo.Pods[delId+1:]...)
				}
			}
		}
	}
	// handle real-added pods.
	for _, newAdd := range result.Add {
		nodeInfo.Pods = append(nodeInfo.Pods, newAdd)
	}
}

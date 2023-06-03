package pod

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/docker/docker/client"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v3"
	"log"
	"mini-k8s/pkg/config"
	container2 "mini-k8s/pkg/container"
	defines2 "mini-k8s/pkg/defines"
	etcd2 "mini-k8s/pkg/etcd"
	network2 "mini-k8s/pkg/network"
	"net/http"
	"time"
)

func SendOutNewPodInfoAfterRealOp(podInfo *defines2.Pod) error {
	body, _ := json.Marshal(podInfo)
	url := "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/updatePodInfoAfterRealOp"
	log.Printf("[apiserver] update pod info request after real op = %v\n", url)
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
	return errors.New("[realPod] error when update pod info after real ops are made")
}

// CreateRealPod used to really create a pod instance according to new added etcd object.
func CreateRealPod(cli *clientv3.Client, id string) *defines2.Pod {
	// create a new docker client first.
	podClient, err := client.NewClientWithOpts(client.WithVersion(config.DockerApiVersion))
	if err != nil {
		log.Printf("an error occurs when creating a docker client: %v\n", err)
		return nil
	}
	// get the pod instance with placeholders from etcd.
	val := etcd2.Get(cli, id).Kvs
	if len(val) == 0 {
		log.Printf("the key %s is not sotred in etcd!\n", id)
		return nil
	}
	res := &defines2.Pod{}
	err = yaml.Unmarshal(val[0].Value, res)
	if err != nil {
		log.Printf("a error occurs when unmarshal in createRealPod func: %v\n", err)
		return nil
	}
	// set current time to Start.
	res.Start = time.Now()
	// set node name here.

	//// NOTE: change here in version 2.0
	//nodeName := "node1"
	//res.NodeId = nodeName

	//// get node info from etcd here to get its cniIP as the network namespace fot the new pod.
	//nodeKey := defines.NodePrefix + "/" + res.NodeId
	//gotNodeKV := etcd.Get(cli, nodeKey).Kvs
	//nodeInfo := &defines.NodeInfo{}
	//_ = yaml.Unmarshal(gotNodeKV[0].Value, nodeInfo)

	// check whether the same-name old ns has exists, if so, delete the old one here.
	oldNsPath := config.NsNamePrefix + res.Metadata.Name
	checkRes := network2.CheckCreateNamespaceSuccess(oldNsPath)
	if checkRes == true {
		// exist the same-name old ns.
		delNsPath := defines2.PodNsPathPrefix + oldNsPath
		network2.DelNsFile(delNsPath)
	}

	//NOTE: change here in version 2.0 .
	//// create a new ns for the new pod.
	//newPodNsName := network2.CreateNamespace(res.Metadata.Name)
	//netNameSpace := defines2.PodNsPathPrefix + newPodNsName
	//fmt.Printf("namspace for new pod %v is : %v\n", id, netNameSpace)
	//index := strings.Index(id, "/")
	//realPodName := id[index+1:]
	//allocatedPodIp := network2.AssignPodIp(netNameSpace, realPodName)
	//res.PodIp = allocatedPodIp
	//fmt.Printf("new allocated pod ip = %v\n", res.PodIp)

	//// TODO: in version 2.0, tmp solution here.
	//netName := defines2.PodNetworkIDPrefix + res.Metadata.Name
	//subNetIP := container2.CreateNetwork(res.PodIp, netName)
	//fmt.Printf("a new network for pod %v is: %v\n", res.Metadata.Name, subNetIP)

	// set pod containers information.
	length := len(res.YamlPod.Spec.Containers)
	res.ContainerStates = make([]defines2.ContainerState, length+1)
	// the change the placeholder to real value and create real containers.
	ret0 := container2.CreatePauseContainer(res, "", podClient)

	// fmt.Printf("ret0 = %s\n", ret0)
	if ret0 == nil {
		log.Println("fail to create a pause container in  CreatePod function!")
	} else {
		// fmt.Println("TEST: successfully create a pause container !")
	}

	pauseState := defines2.ContainerState{}
	pauseState.Id = ret0.Id
	pauseState.State = ret0.State
	pauseState.Name = ret0.Name
	res.ContainerStates[0] = pauseState
	log.Printf("pauseContainer info after CreatePauseContainer: %v\n", pauseState)
	// in version 2.0, change here to get pause container's IP as the pod IP.
	inspectRes := container2.InspectContainer(ret0.Id)
	res.PodIp = inspectRes.NetworkSettings.IPAddress
	mode := "container:" + string(ret0.Id)
	success := true
	failId := -1
	for idx, con := range res.YamlPod.Spec.Containers {
		ret1 := container2.CreateNormalContainer(podClient, &con, mode, res)
		if ret1 == "-1" {
			// log.Println("get into CreateNormalContainer return -1 case!")
			log.Printf("an error occurs when creating a new normal container in CreatePod function! (return -1)\n")
			// TODO: add error handle logic here??
			success = false
			failId = idx
			break
		}
		// start the newly-created normal container here.
		// TODO: in version 1.0, try to start this container directly but in the future this start should be managed.
		// fmt.Println("ready to start a normal container in func CreatePod!")
		// fmt.Printf("new normal container id = %s\n", string(ret1))
		container2.StartNormalContainer(string(ret1))
		state0 := defines2.ContainerState{}
		state0.Id = string(ret1)
		state0.Name = con.Name
		/* TODO:(lyh) not sure here : just set the newly created pod's state to RUNNING ??? And what about IP addr??? */
		state0.State = defines2.Running
		log.Printf("normal container states info after creating: %v\n", state0)
		// the first pos is for pause container.
		res.ContainerStates[idx+1] = state0
	}

	// NOTE: add error handler here. (If a normal podContainer fails to be created, delete all other successfully-created containers of this pod. )
	if success == false && failId != -1 {
		// remove all running normal containers.
		for i := 0; i < failId; i++ {
			container2.RemoveForceContainer(res.ContainerStates[i+1].Id)
		}
		// remove pause container.
		container2.RemoveForceContainer(res.ContainerStates[0].Id)
		// then clear the states recorded by res.
		res.ContainerStates = make([]defines2.ContainerState, 0)
	}

	// mark current pod running.
	res.PodState = defines2.Running

	// error handler.
	if success == false {
		res.PodState = defines2.Failed
	}

	// next store new pod instance info back to etcd.
	//newKey := id
	//newVal, err := yaml.Marshal(res)
	//// TODO: in version 3.0, move the etcd-write operations from kubelet to api-server.
	// etcd2.Put(cli, newKey, string(newVal))
	_ = SendOutNewPodInfoAfterRealOp(res)
	return res
}

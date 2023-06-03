package apiserver

import (
	"bytes"
	"encoding/json"
	"github.com/docker/docker/client"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v3"
	"log"
	"mini-k8s/pkg/config"
	"mini-k8s/pkg/defines"
	"mini-k8s/pkg/etcd"
	"net/http"
	"time"
)

const PodInstancePrefix = "PodInstance"

func SendOutNewScheduleRequest(podInfo *defines.Pod) string {
	body, _ := json.Marshal(podInfo.YamlPod)
	url := "http://" + config.MasterIP + ":" + config.SchedulerPort + "/objectAPI/schedulePolicy"
	log.Printf("[apiserver] schedule policy request = %v\n", url)
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return ""
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Printf("[apiserver] an error occurs when send out new shcedule request!\n")
		return ""
	}
	bodyReader := response.Body
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(bodyReader)
	return buf.String()
}

// SendOutScheduleRequest should return nodeId for this pod.
func SendOutScheduleRequest(podInfo *defines.Pod) string {
	body, _ := json.Marshal(podInfo)
	url := "http://" + config.MasterIP + ":" + config.SchedulerPort + "/objectAPI/schedulePodNode"
	log.Printf("[apiserver] schedule request = %v\n", url)
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return ""
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return ""
	}
	bodyReader := response.Body
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(bodyReader)
	return buf.String()
}

func CreatePod(cli *clientv3.Client, yamlPod *defines.YamlPod) *defines.Pod {
	// may add mutex here to avoid race ???
	config.ApiServerMutex.Lock()
	log.Printf("[apiserver] createPod %v holds this lock!\n", yamlPod.Metadata.Name)
	defer func() {
		log.Printf("[apiserver] createPod %v give up lock!\n", yamlPod.Metadata.Name)
		config.ApiServerMutex.Unlock()
	}()

	// Change in version 2.0: add duplicate check here first.
	prefixKey := defines.PodInstancePrefix + "/"
	oldPodsKvs := etcd.GetWithPrefix(cli, prefixKey).Kvs
	if len(oldPodsKvs) != 0 {
		// don't need to consider the duplicate case for the first pod instance.
		for _, oldPodsKv := range oldPodsKvs {
			oldPod := &defines.Pod{}
			_ = yaml.Unmarshal(oldPodsKv.Value, oldPod)
			if oldPod.Metadata.Name == yamlPod.Metadata.Name {
				log.Println("[apiserver] Pod with this name has already been created in the current system!")
				return oldPod
			}
		}
	}
	res := &defines.Pod{}
	// means nothing.
	res.PodIp = "--.--.--.--"
	// first add the metadata into etcd.
	key := defines.PodPrefix + "/" + yamlPod.Metadata.Name
	res.YamlPod = *yamlPod

	// change here.
	if res.NodeSelector.Gpu == "" {
		// give this yamlPod a default value.
		res.YamlPod.NodeSelector.Gpu = "nvidia"
	}

	// res.Start = time.Now()
	res.PodId = PodInstancePrefix + "/" + yamlPod.Metadata.Name
	podClient, err := client.NewClientWithOpts(client.WithVersion(config.DockerApiVersion))
	if err != nil {
		log.Printf("[apiserver] an error occurs when creating a docker client: %v\n", err)
		return nil
	}

	// nodeName := SendOutScheduleRequest(res)
	nodeName := SendOutNewScheduleRequest(res)
	nodeName = nodeName[1 : len(nodeName)-1]
	// nodeName := ""
	log.Printf("[apiserver] received target node name = %v\n", nodeName)
	//// NOTE: test only !!!
	//nodeName := "node1"

	res.NodeId = nodeName

	// then persist these data into etcd.
	val, err0 := yaml.Marshal(*yamlPod)
	if err0 != nil {
		log.Printf("[apiserver] an error occurs when marshal yamlPod in CreatePod function : %v\n", err0)
		return nil
	}
	// persist pod yaml data.
	etcd.Put(cli, key, string(val))
	val, err0 = yaml.Marshal(res)
	if err0 != nil {
		log.Printf("[apiserver] an error occurs when marshal Pod instance in CreatePod function : %v\n", err0)
		return nil
	}
	// persist pod instance data.
	etcd.Put(cli, res.PodId, string(val))
	// then try to update all-pod-set data in etcd.
	allKey := defines.PodIdSetPrefix + "/"
	// oldAllKey := defines.OldPodIdSetPrefix + "/"
	got := etcd.Get(cli, allKey)
	kv := got.Kvs
	newSet := make([]string, 0)
	if len(kv) == 0 {
		// .Println("get into size = 0 case!")
		//// store old val here.
		//old, err := yaml.Marshal(newSet)
		//if err != nil {
		//	fmt.Printf("an error occurs when marshal old val in size = 0 case: %v\n", err)
		//	return nil
		//}
		//etcd.Put(cli, oldAllKey, string(old))

		newSet = append(newSet, res.PodId)
		newVal, err0 := yaml.Marshal(newSet)
		if err0 != nil {
			log.Printf("[apiserver] an error occurs when marshal allSet when size = 0 :%v\n", err0)
			return nil
		}
		etcd.Put(cli, allKey, string(newVal))
		// return res
	} else {
		oldVal := kv[0].Value
		//// store old value here.
		//etcd.Put(cli, oldAllKey, string(oldVal))

		err = yaml.Unmarshal(oldVal, &newSet)
		if err != nil {
			log.Printf("[apiserver] an error occurs when unmarshal allSet data: %v\n", err)
			return nil
		}
		// add check-duplicate logic.
		for _, singleId := range newSet {
			if singleId == res.PodId {
				log.Printf("[apiserver] id %s has already been stored in etcd!\n", res.PodId)
				// no need to store again.
				return res
			}
		}
		newSet = append(newSet, res.PodId)
		newVal1, err2 := yaml.Marshal(newSet)
		if err2 != nil {
			log.Printf("[apiserver] an error occurs when marshal allSet when size > 0:%v\n", err2)
			return nil
		}
		etcd.Put(cli, allKey, string(newVal1))
	}

	// NOTE: in version 2.0, add scheduler failure logic here.
	if nodeName == "" {
		// retry for 10 times before give up.
		succeed := false
		for i := 0; i < 10; i++ {
			gotName := SendOutNewScheduleRequest(res)
			gotName = gotName[1 : len(gotName)-1]
			if gotName != "" {
				res.NodeId = gotName
				nodeName = gotName
				succeed = true
				break
			}
			time.Sleep(5 * time.Second)
		}
		if succeed == false {
			errClose := podClient.Close()
			if errClose != nil {
				log.Printf("[apiserver] an error occurs when close podClient in createPod function: %v\n", errClose)
				return nil
			}
			log.Printf("[apiserver] create a new pod %s successfully, but not be scheduled to a node!\n", res.Metadata.Name)
			return res
		}
	}

	// BUG: here we can not put the pod into node directly as there are many placeholders!!!(have been removed)

	// then put the new pod name into node's pod list in etcd.

	// NOTE: change here in version 2.0, the node pod list should not be updated here, and should be updated by pod-change-watcher and scheduler.
	nodePodListKey := defines.NodePodsListPrefix + "/" + nodeName
	// fmt.Printf("nodPodListKey = %v\n", nodePodListKey)
	kv = etcd.Get(cli, nodePodListKey).Kvs
	if len(kv) == 0 {
		// the first pod of related node.
		list := make([]string, 0)
		list = append(list, res.PodId)
		listByte, _ := yaml.Marshal(&list)
		etcd.Put(cli, nodePodListKey, string(listByte))
	} else {
		gotList := make([]string, 0)
		err = yaml.Unmarshal(kv[0].Value, &gotList)
		if err != nil {
			log.Printf("[apiserver] an error occurs when unmarshal ols node pods list in CreatePod func: %v\n", err)
			return nil
		}
		newList := append(gotList, res.PodId)
		newListByte, _ := yaml.Marshal(&newList)
		etcd.Put(cli, nodePodListKey, string(newListByte))
	}

	// change in version 2.0: add handleNodePod logic here !!!(no watch anymore!)
	newNodePodSet := make([]string, 0)
	kv = etcd.Get(cli, nodePodListKey).Kvs
	err = yaml.Unmarshal(kv[0].Value, &newNodePodSet)
	HandlePodUpdateForNode(cli, res.NodeId, newNodePodSet)
	////TODO: change here to new version of HandlePodUpdateForNode in version 2.0
	//HandlePodUpdatesForNodeNew(cli, res.NodeId, newNodePodSet)

	// in version 2.0, change here to support service update checking after a new pod is added.
	// UpdateService(cli)
	// CheckPodAddInService(cli, res)

	errClose := podClient.Close()
	if errClose != nil {
		log.Printf("[apiserver] an error occurs when close podClient in createPod function: %v\n", errClose)
		return nil
	}
	log.Printf("[apiserver] create a new pod %s successfully!\n", res.Metadata.Name)
	return res
}

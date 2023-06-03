package apiserver

import (
	"errors"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v3"
	"log"
	"mini-k8s/pkg/config"
	defines2 "mini-k8s/pkg/defines"
	etcd "mini-k8s/pkg/etcd"
	"strings"
)

func CheckPodInAutoAndDel(cli *clientv3.Client, podName string) {
	if len(podName) < 8 {
		return
	}
	prefix := podName[0:8]
	if prefix == defines2.AutoPodNamePrefix {
		// MAY be managed by a autoScaler.
		partName := podName[8:]
		index := strings.Index(partName, "_")
		if index <= 0 {
			return
		} else {
			autoName := partName[0:index]
			key := defines2.AutoPodReplicasNamesPrefix + "/" + defines2.AutoPodNamePrefix + autoName
			kv := etcd.Get(cli, key).Kvs
			if len(kv) == 0 {
				// this autoScaler does not exist in the current system.
				return
			}
			gotList := make([]string, 0)
			_ = yaml.Unmarshal(kv[0].Value, &gotList)
			for idx, name := range gotList {
				if name == podName {
					gotList = append(gotList[0:idx], gotList[idx+1:]...)
					break
				}
			}
			gotListByte, _ := yaml.Marshal(&gotList)
			etcd.Put(cli, key, string(gotListByte))
			return
		}
	}
}

// DeletePod used tp delete a pod instance info in etcd according to param pod name @podName.
// NOTE: in version 2.0, add logic to delete related node info in etcd.
func DeletePod(client *clientv3.Client, podName string) error {
	// may add mutex here ??? the same mutex as create or a new one ???
	config.ApiServerMutex.Lock()
	log.Printf("[apiserver] deletePod %v holds this lock!\n", podName)
	defer func() {
		log.Printf("[apiserver] deletePod %v give up lock!\n", podName)
		config.ApiServerMutex.Unlock()
	}()

	res := &defines2.Pod{}
	key := "PodInstance/" + podName

	// change here in version 2.0 .
	nodeListPrefixKey := defines2.NodePodsListPrefix + "/"

	got := etcd.Get(client, key)
	if got == nil {
		log.Printf("[apiserver] an error occurs when getting pod %s info from etcd\n", podName)
		return errors.New("error when getting pod from etcd in func DeletePod")
	}
	kv := got.Kvs
	if len(kv) == 0 {
		log.Printf("[apiserver] information for pod %s is empty!\n", podName)
		return errors.New("error, the pod is empty in function DeletePod")
	}

	// add logic for deleting pod that does not be scheduled to a node.
	tmpPod := &defines2.Pod{}
	_ = yaml.Unmarshal(kv[0].Value, tmpPod)
	if tmpPod.NodeId == "" {
		// no node is managing this pod.
		instanceKey := defines2.PodInstancePrefix + "/" + tmpPod.Metadata.Name
		podKey := defines2.PodPrefix + "/" + tmpPod.Metadata.Name
		// delete this pod instance from podsList first.
		oldListKey := defines2.PodIdSetPrefix + "/"
		kv0 := etcd.Get(client, oldListKey).Kvs
		oldSet := make([]string, 0)
		_ = yaml.Unmarshal(kv0[0].Value, &oldSet)
		for id, name := range oldSet {
			if name == instanceKey {
				oldSet = append(oldSet[0:id], oldSet[id+1:]...)
				break
			}
		}
		newSetByte, _ := yaml.Marshal(&oldSet)
		etcd.Put(client, oldListKey, string(newSetByte))
		// then delete the pod instance and store the oldPodInstance.
		oldInstanceKey := defines2.OldPodInstancePrefix + "/" + tmpPod.Metadata.Name
		podByte, _ := yaml.Marshal(tmpPod)
		etcd.Del(client, instanceKey)
		etcd.Del(client, podKey)
		etcd.Put(client, oldInstanceKey, string(podByte))
		log.Printf("[apiserver] delete pod %s successfully!\n", res.Metadata.Name)
		return nil
	}

	// change here.
	kvs := etcd.GetWithPrefix(client, nodeListPrefixKey).Kvs
	if len(kvs) == 0 {
		log.Println("[apiserver] Error! There is no node in the current system!")
		return errors.New("error! There is no node in the current system")
	}
	nodeName := ""
	found := false
	targetPodName := defines2.PodInstancePrefix + "/" + podName
	for _, singleKv := range kvs {
		thisList := make([]string, 0)
		_ = yaml.Unmarshal(singleKv.Value, &thisList)
		for _, i := range thisList {
			if i == targetPodName {
				_ = yaml.Unmarshal(singleKv.Key, &nodeName)
				found = true
				break
			}
		}
		if found == true {
			break
		}
	}

	// fmt.Printf("nodeName = %s\n", nodeName)
	if nodeName != "" {
		index := strings.Index(nodeName, "/")
		nodeName = nodeName[index+1:]
	}

	// remove this pod instance name from this node-related NodePodsList.
	targetNodeListKey := defines2.NodePodsListPrefix + "/" + nodeName
	targetNodeListKv := etcd.Get(client, targetNodeListKey).Kvs
	if len(targetNodeListKv) == 0 {
		log.Println("[apiserver] Error! target node doesn't exist in the related node's podsList!")
		return errors.New("error! target node doesn't exist in the related node's podsList")
	}
	targetList := make([]string, 0)
	_ = yaml.Unmarshal(targetNodeListKv[0].Value, &targetList)
	targetPodListId := -1
	for idx, element := range targetList {
		if element == targetPodName {
			targetPodListId = idx
			break
		}
	}
	if targetPodListId == -1 {
		log.Println("[apiserver] Error! this pod doesn't be stored be the related node's podsList!")
		return errors.New("error! this pod doesn't be stored be the related node's podsList")
	}
	targetList = append(targetList[:targetPodListId], targetList[targetPodListId+1:]...)
	//// store the new node info into etcd.
	//nodeInfoByte, _ := yaml.Marshal(nodeInfo)
	//etcd.Put(client, nodeKey, string(nodeInfoByte))

	info := kv[0].Value
	err := yaml.Unmarshal(info, res)
	if err != nil {
		log.Printf("[apiserver] an error occurs when unmarshal in func DeletePod %v\n", err)
		return errors.New("an error occurs when unmarshal in func DeletePod")
	}
	podId := res.PodId
	podId = strings.Replace(podId, "/", "-", -1)

	//// then stop all containers in this pod first.
	//for _, conState := range res.ContainerStates {
	//	// name := podId + "-" + con.Name
	//	tmp := container.StopContainer(conState.Id)
	//	if tmp == false {
	//		fmt.Printf("an error occurs when stopping container %s\n", conState.Name)
	//		return
	//	}
	//}
	//// next remove all containers in this pod.
	//for _, conState := range res.ContainerStates {
	//	// name := podId + "-" + con.Name
	//	tmp := container.RemoveContainer(conState.Id)
	//	if tmp == false {
	//		fmt.Printf("an error occurs when removing container %s\n", conState.Name)
	//		return
	//	}
	//}

	// then remove the persisted data in etcd of this pod.
	podKey := defines2.PodPrefix + "/" + res.Metadata.Name
	podInstanceKey := "PodInstance/" + res.Metadata.Name
	oldPodInstanceKey := defines2.OldPodInstancePrefix + "/" + res.Metadata.Name

	// store old podInstance value here(TODO: tmp sol for version 1.0)
	oldPod, err := yaml.Marshal(res)
	etcd.Put(client, oldPodInstanceKey, string(oldPod))

	// no need to store old yamlPod info.
	etcd.Del(client, podKey)
	etcd.Del(client, podInstanceKey)
	// then remove the pod id from all id set stored in etcd.
	allKey := defines2.PodIdSetPrefix + "/"
	// oldAllKey := defines.OldPodIdSetPrefix + "/"
	all := etcd.Get(client, allKey).Kvs
	if len(all) == 0 {
		fmt.Println("the allID list should not be empty when deleting a pod is not finished!")
		return errors.New("the allID list should not be empty when deleting a pod is not finished")
	}
	allStr := make([]string, 0)

	// store old value here.
	etcd.Put(client, defines2.OldPodIdSetPrefix+"/", string(all[0].Value))

	err = yaml.Unmarshal(all[0].Value, &allStr)
	if err != nil {
		log.Printf("[apiserver] an error occurs when unmarshal allID list in deletePod func: %v\n", err)
		return errors.New("an error occurs when unmarshal allID list in deletePod func")
	}
	idx := -1
	for id, str := range allStr {
		if str == podInstanceKey {
			idx = id
			break
		}
	}
	if idx == -1 {
		log.Println("[apiserver] not found the podInstance name from allID list whn deleting!")
		return errors.New("not found the podInstance name from allID list whn deleting")
	}
	// fix a BUG here.
	if idx == 0 {
		allStr = allStr[1:]
	} else {
		if idx == len(allStr)-1 {
			allStr = allStr[0 : len(allStr)-1]
		} else {
			if idx > 0 && idx < len(allStr)-1 {
				// remove the idx element with slice operations.
				left := allStr[0:idx]
				right := allStr[idx+1:]
				// concat two slice together.
				allStr = append(left, right...)
			}
		}
	}
	newSet, err := yaml.Marshal(&allStr)
	if err != nil {
		log.Printf("[apiserver] an error occurs when marshal new set val in DeletePod func: %v\n", err)
		return errors.New("an error occurs when marshal new set val in DeletePod func")
	}
	etcd.Put(client, allKey, string(newSet))

	// IMPORTANT: Note the order of operations! the change to the watch target of watcher should be put at the end of one function!!!
	// store the new node pods-list info into etcd.
	newListByte, _ := yaml.Marshal(&targetList)
	etcd.Put(client, targetNodeListKey, string(newListByte))

	// check whether this pod is managed by an autoScaler.
	CheckPodInAutoAndDel(client, podName)

	// change in version 2.0: add handleNodePod logic here !!!(no watch anymore!)
	newNodePodSet := make([]string, 0)
	kv = etcd.Get(client, targetNodeListKey).Kvs
	err = yaml.Unmarshal(kv[0].Value, &newNodePodSet)
	log.Printf("[apiserver] res nodeId = %v\n", res.NodeId)

	HandlePodUpdateForNode(client, res.NodeId, newNodePodSet)
	//// TODO: change here to new version of HandlePodUpdateForNode in version 2.0
	//HandlePodUpdatesForNodeNew(client, res.NodeId, newNodePodSet)

	// in version 2.0, we change here to add service-update-check after a pod is deleted.
	// UpdateService(client)
	// CheckPodDelInService(client, res)

	log.Printf("[apiserver] delete pod %s successfully!\n", res.Metadata.Name)
	return nil
}

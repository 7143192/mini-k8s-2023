package pod

import (
	clientv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v3"
	"log"
	container2 "mini-k8s/pkg/container"
	"mini-k8s/pkg/defines"
	"mini-k8s/pkg/etcd"
	network2 "mini-k8s/pkg/network"
	"strings"
)

// DelRealPod used to really create a pod instance according to new added etcd object.
func DelRealPod(cli *clientv3.Client, id string) *defines.Pod {
	// TODO: tmp solution for version 1.0(use "OldPodInstance/" prefix) as there is no node object to store all pod information.
	// get oldVal stored in etcd first.
	idx := strings.Index(id, "/")
	name := id[idx+1:]
	oldVal := etcd.Get(cli, defines.OldPodInstancePrefix+"/"+name).Kvs
	if len(oldVal) == 0 {
		log.Println("list should not be empty(in function DelRealPod)!")
		return nil
	}
	res := &defines.Pod{}
	err := yaml.Unmarshal(oldVal[0].Value, res)

	if err != nil {
		log.Printf("an error occurs when unmarshal oldVal in func DelRealPod: %v\n", err)
		return nil
	}

	// then remove the ns created for this pod.
	podNsName := network2.GetNamespace(name)
	//network.DelNamespace(podNsName, name)

	//// then stop all containers in this pod first.
	//for _, conState := range res.ContainerStates {
	//	// name := podId + "-" + con.Name
	//	tmp := container2.StopContainer(conState.Id)
	//	if tmp == false {
	//		fmt.Printf("an error occurs when stopping container %s\n", conState.Name)
	//		return nil
	//	}
	//}
	// next remove all containers in this pod.
	for _, conState := range res.ContainerStates {
		// name := podId + "-" + con.Name
		tmp := container2.RemoveForceContainer(conState.Id)
		if tmp == false {
			log.Printf("an error occurs when removing container %s\n", conState.Name)
			return nil
		}
	}
	delResult := network2.DelNsFile(podNsName)
	if delResult == true {
		log.Printf("delete pod ns path %v successfully!\n", podNsName)
	}
	return res
}

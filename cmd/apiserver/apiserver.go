package main

import (
	"gopkg.in/yaml.v3"
	"log"
	"mini-k8s/pkg/apiserver"
	defines2 "mini-k8s/pkg/defines"
	etcd2 "mini-k8s/pkg/etcd"
	"strings"
	"time"
)

func CheckNodesStatesOneTime() {
	// get all nodes info from etcd first.
	allNodeKey := defines2.AllNodeSetPrefix + "/"
	cli := etcd2.EtcdStart()
	defer cli.Close()
	kv := etcd2.Get(cli, allNodeKey).Kvs
	if len(kv) == 0 {
		log.Println("[apiserver]: No node exists in the current system!")
		return
	}
	heartBeats := make([]*defines2.NodeHeartBeat, 0)
	nodeNames := make([]string, 0)
	_ = yaml.Unmarshal(kv[0].Value, &nodeNames)
	for i, name := range nodeNames {
		// heartBeats = append(heartBeats, nil)
		id := strings.Index(name, "/")
		name = name[id+1:]
		nodeNames[i] = name
	}
	for _, name := range nodeNames {
		key := defines2.NodeHeartBeatPrefix + "/" + name
		kv = etcd2.Get(cli, key).Kvs
		if len(kv) == 0 {
			// should not get here!
			heartBeats = append(heartBeats, nil)
			continue
		} else {
			heartBeat := &defines2.NodeHeartBeat{}
			_ = yaml.Unmarshal(kv[0].Value, heartBeat)
			heartBeats = append(heartBeats, heartBeat)
		}
	}
	// log.Printf("heartBeats = %v\n", heartBeats)
	// check the diff between last heartBeat time and current time
	// nodeRealName := ""
	for _, heartBeat := range heartBeats {
		healthy := true
		if heartBeat == nil {
			continue
		}
		diff := time.Now().Sub(heartBeat.CurTime).Seconds()
		log.Printf("[apiserver] diff from last heart beat time = %v\n", diff)
		if diff >= 120 {
			// not a healthy node.
			healthy = false
			log.Printf("[apiserver] node %v is not ready for now!\n", heartBeat.NodeId)
			// reset some states of this node stored in etcd.
			// nodeRealName = heartBeat.NodeId
			nodeKey := defines2.NodePrefix + "/" + heartBeat.NodeId
			// log.Printf("nodeKey = %v\n", nodeKey)
			kv = etcd2.Get(cli, nodeKey).Kvs
			if len(kv) == 0 {
				// should never reach here!
				log.Printf("[apiserver] not-ready node %v does not exist in the system at all!\n", heartBeat.NodeId)
				return
			} else {
				nodeInfo := &defines2.NodeInfo{}
				_ = yaml.Unmarshal(kv[0].Value, nodeInfo)
				// set this node state as NOT READY.
				nodeInfo.NodeData.NodeState = defines2.NodeNotReady
				// set all pods in this node as UNKNOWN.
				for _, pod := range nodeInfo.Pods {
					pod.PodState = defines2.Unknown
				}
				// set all containers in every pod as UNKNOWN.
				for _, pod := range nodeInfo.Pods {
					for idx, _ := range pod.ContainerStates {
						pod.ContainerStates[idx].State = defines2.Unknown
					}
				}
				nodeInfoByte, _ := yaml.Marshal(nodeInfo)
				etcd2.Put(cli, nodeKey, string(nodeInfoByte))
			}
		}
		if healthy == true {
			// if the new state is well, check whether the old state is NOT READY.
			oldKey := defines2.NodePrefix + "/" + heartBeat.NodeId
			// log.Printf("oldKey = %v\n", oldKey)
			oldKv := etcd2.Get(cli, oldKey).Kvs
			if len(oldKv) != 0 {
				// log.Println("get into len != 0 and re-healthy case!")
				oldNodeInfo := &defines2.NodeInfo{}
				_ = yaml.Unmarshal(oldKv[0].Value, oldNodeInfo)
				oldNodeInfo.NodeData.NodeState = defines2.NodeReady
				for _, pod := range oldNodeInfo.Pods {
					pod.PodState = defines2.Running
					for idx := range pod.ContainerStates {
						pod.ContainerStates[idx].State = defines2.Running
					}
				}
				newNodeInfoByte, _ := yaml.Marshal(oldNodeInfo)
				etcd2.Put(cli, oldKey, string(newNodeInfoByte))
			}
		}
	}
}

func CheckNodesStates() {
	for {
		// log.Println("ready to check nodes states!")
		CheckNodesStatesOneTime()
		time.Sleep(30 * time.Second)
	}
}

func main() {
	//http.HandleFunc("/hello", apiserver.Hello)
	//http.HandleFunc("/registerNewNode", apiserver.RegisterNewNode)
	//http.HandleFunc("/heartBeat", apiserver.ReceiveNodeHeartBeat)

	// in version 2.0, make kubectl to send corresponding request to api-server,
	// not call the api-server's functions directly.

	// TODO: in version 2.0, start Flannel here.(tmp solution)
	cli := etcd2.EtcdStart()
	//network.FlannelInit(cli)
	cli.Close()

	go CheckNodesStates()
	// http.ListenAndServe(":"+config.MasterPort, nil)
	// in version 2.0, change to the new server.
	server := apiserver.ServerInit()

	// NOTE: change here: DO NOT watch etcd anymore, and all changes are sent through HTTP directly.

	//// in version 2.0, watch etcd in api server and not in kubelet.
	//go server.ServerEtcdWatcher()

	err := server.Run()
	if err != nil {
		panic(err)
	}
}

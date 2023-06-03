package apiserver

import (
	"encoding/json"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v3"
	"mini-k8s/pkg/defines"
	"mini-k8s/pkg/etcd"
)

func GetNode(client *clientv3.Client) string {
	res := &defines.GetNodesResource{}
	prefixKey := defines.NodePrefix + "/"
	kvs := etcd.GetWithPrefix(client, prefixKey).Kvs
	if len(kvs) == 0 {
		// fmt.Println("No node exists in the current system!")
		return "No node exists in the current system!"
	}
	// fmt.Println("all nodes brief information:")
	// fmt.Println("NAME\t\tID\t\tLABEL\t\tIP\t\t\tSTATE")
	for _, kv := range kvs {
		tmp := &defines.GetNodeResourceSend{}
		state := ""
		nodeInfo := &defines.NodeInfo{}
		_ = yaml.Unmarshal(kv.Value, nodeInfo)
		if nodeInfo.NodeData.NodeState == defines.NodeReady {
			state = "READY"
		} else {
			state = "NOT READY"
		}
		//fmt.Printf("%v\t\t%v\t\t%v\t\t%v\t\t%v\n",
		//	nodeInfo.NodeData.NodeSpec.Metadata.Name,
		//	nodeInfo.NodeData.NodeId,
		//	nodeInfo.NodeData.NodeSpec.Metadata.Label,
		//	nodeInfo.NodeData.NodeSpec.Metadata.Ip,
		//	state)
		tmp.Name = nodeInfo.NodeData.NodeSpec.Metadata.Name
		tmp.Id = nodeInfo.NodeData.NodeId
		tmp.Ip = nodeInfo.NodeData.NodeSpec.Metadata.Ip
		tmp.Label = nodeInfo.NodeData.NodeSpec.Metadata.Label
		tmp.State = state
		res.NodesSend = append(res.NodesSend, tmp)
	}
	resByte, _ := json.Marshal(res)
	return string(resByte)
}

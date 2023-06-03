package node

import (
	clientv3 "go.etcd.io/etcd/client/v3"
	defines2 "mini-k8s/pkg/defines"
)

func CreateNode(cli *clientv3.Client, yamlNode *defines2.NodeYaml) *defines2.NodeInfo {
	res := &defines2.NodeInfo{}
	node := &defines2.Node{}
	node.NodeSpec = *yamlNode
	// similar to pod, this id should be filled when watcher monitors this creation and create real node.
	node.NodeId = ""
	node.NodeState = defines2.NodeNotReady
	res.NodeData = *node
	res.NodeData.Selector.Gpu = yamlNode.Metadata.Selector.Gpu
	// these data should be filled when the node is really created in worker layer.
	//// change: init two clients here!!
	//res.EtcdClient = nil
	//res.CadvisorClient = nil
	res.Pods = make([]*defines2.Pod, 0)
	res.Registered = false
	// no need to store into etcd here.
	//// then store the data with placeholders into etcd.
	//nodeByte, _ := yaml.Marshal(node)
	//nodeKey := defines.NodePrefix + "/" + yamlNode.Metadata.Name
	//etcd.Put(cli, nodeKey, string(nodeByte))
	return res
}

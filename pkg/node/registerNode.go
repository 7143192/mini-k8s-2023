package node

import (
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v3"
	"log"
	"mini-k8s/pkg/defines"
	etcd2 "mini-k8s/pkg/etcd"
	"mini-k8s/utils"
)

// RegisterNodeToMaster used to register a new node into master according to the YAML configuration file.
func RegisterNodeToMaster(cli *clientv3.Client, yamlNode *defines.NodeYaml) *defines.NodeInfo {
	//// NOTE: in version 2.0, add kube-proxy ip chain init here.
	//kubeproxy.InitSvcMainChain()

	// parse flannel ip to get first two parts of the IP.
	a, b := utils.ParseFlannelIP()

	res := CreateNode(cli, yamlNode)
	// node name and ID can not be duplicated!!!
	// check whether node name and IP are duplicated.
	// get all nodes here to check.
	prefixKey := defines.NodePrefix + "/"
	kvs := etcd2.GetWithPrefix(cli, prefixKey).Kvs
	nodes := make([]*defines.NodeInfo, 0)
	if len(kvs) != 0 {
		for _, kv := range kvs {
			oneNode := &defines.NodeInfo{}
			err0 := yaml.Unmarshal(kv.Value, oneNode)
			if err0 != nil {
				log.Printf("an error occurs when unmarshal one node in RegisterNodeToMaster func: %v\n", err0)
				return nil
			}
			nodes = append(nodes, oneNode)
		}
	}
	for _, singleNode := range nodes {
		if yamlNode.Metadata.Ip == singleNode.NodeData.NodeSpec.Metadata.Ip ||
			yamlNode.Metadata.Name == singleNode.NodeData.NodeSpec.Metadata.Name {
			log.Println("Node with this IP/name has already been registered in the current system!")
			singleNode.Registered = true // is registered before!
			// when detecting duplication, return the old node directly.
			return singleNode
		}
	}
	// if not duplicated, allocate a new id to this new node.
	curIdKey := defines.CurNodeIdPrefix + "/"
	newId := 0
	gotId := etcd2.Get(cli, curIdKey).Kvs
	if len(gotId) == 0 {
		// this is the first node that try to register to master.
		newId = 1
	} else {
		// not the first node.
		err := yaml.Unmarshal(gotId[0].Value, &newId)
		if err != nil {
			log.Printf("an error occurs when unmarshal oldId in RegisterNodeToMaster func: %v\n", err)
			return nil
		}
		newId = newId + 1
	}
	// reset some states of NodeInfo object.
	res.NodeData.NodeId = fmt.Sprintf("%d", newId)

	// TODO: in version 2.0, alloc a subnetwork IP as Cni IP for the new node.
	gotIP := utils.AllocateNewSubIp(a, b, res)
	res.NodeData.CniIp = gotIP

	res.NodeData.NodeState = defines.NodeReady
	// then store new current node id into etcd.
	newIdByte, _ := yaml.Marshal(&newId)
	etcd2.Put(cli, curIdKey, string(newIdByte))
	// store the new object into etcd.
	newNodeKey := defines.NodePrefix + "/" + res.NodeData.NodeSpec.Metadata.Name
	nodeByte, _ := yaml.Marshal(res)
	etcd2.Put(cli, newNodeKey, string(nodeByte))
	// then update all node set info in etcd.
	setKey := defines.AllNodeSetPrefix + "/"
	kvs = etcd2.Get(cli, setKey).Kvs
	if len(kvs) == 0 {
		// this is the first node.
		nodeSet := make([]string, 0)
		nodeSet = append(nodeSet, newNodeKey)
		nodeSetByte, _ := yaml.Marshal(&nodeSet)
		etcd2.Put(cli, setKey, string(nodeSetByte))
		return res
	}
	oldSet := make([]string, 0)
	_ = yaml.Unmarshal(kvs[0].Value, &oldSet)
	newSet := append(oldSet, newNodeKey)
	newSetByte, _ := yaml.Marshal(&newSet)
	etcd2.Put(cli, setKey, string(newSetByte))

	return res
}

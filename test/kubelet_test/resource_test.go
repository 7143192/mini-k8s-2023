package kubelet_test

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"mini-k8s/pkg/cadvisor"
	"mini-k8s/pkg/defines"
	"mini-k8s/pkg/etcd"
	"testing"
	"time"
)

var times int = 0

func CollectResourceInfoTest(KubeNode *defines.NodeInfo) {
	for {
		//TODO: tmp sol for test in version 2.0!
		etcdClient := etcd.EtcdStart()
		nodeName := KubeNode.NodeData.NodeSpec.Metadata.Name
		nodeKey := defines.NodePrefix + "/" + nodeName
		kv := etcd.Get(etcdClient, nodeKey).Kvs
		if len(kv) == 0 {
			fmt.Printf("the node %v does not exist in the system !\n", nodeName)
			time.Sleep(30 * time.Second)
			continue
		} else {
			_ = yaml.Unmarshal(kv[0].Value, KubeNode)
			// KubeNode.EtcdClient = etcdClient
			// cadVisorClient, _ := cadvisor.CadStart()
			// KubeNode.CadvisorClient = cadVisorClient
		}

		// cli := KubeNode.CadvisorClient
		cli, _ := cadvisor.CadStart()
		// fmt.Println("ready to collect resource information!")
		cadvisor.RecordInstanceResource(cli, KubeNode)
		etcdClient.Close()
		times++
		if times == 5 {
			return
		}
		time.Sleep(30 * time.Second)
	}
}

func TestResource(t *testing.T) {
	// we only need to test the correctness of the resource collection functions.
	cli := etcd.EtcdStart()
	nodeInfo := &defines.NodeInfo{}
	kv := etcd.Get(cli, "Node/node1").Kvs
	_ = yaml.Unmarshal(kv[0].Value, nodeInfo)
	CollectResourceInfoTest(nodeInfo)
	if times == 5 {
		t.Log("Pass resource test!")
	} else {
		t.Error("Fail resource test!")
	}
}

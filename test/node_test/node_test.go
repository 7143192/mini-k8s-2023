package node_test

import (
	"mini-k8s/pkg/etcd"
	"mini-k8s/pkg/node"
	"mini-k8s/utils/yaml"
	"testing"
)

func TestNode(t *testing.T) {
	// start etcd client first.
	cli := etcd.EtcdStart()
	defer cli.Close()
	// get nodeYaml first.
	yamlNode, _ := yaml.ParseNodeConfig("../../utils/templates/node_template.yaml")
	// create test.
	nodeInfo := node.RegisterNodeToMaster(cli, yamlNode)
	if nodeInfo.NodeData.NodeId != "1" {
		t.Error("Fail create Node Test!")
		return
	}
	yamlNode1, _ := yaml.ParseNodeConfig("../../utils/templates/node_template_1.yaml")
	nodeInfo1 := node.RegisterNodeToMaster(cli, yamlNode1)
	if nodeInfo1.NodeData.NodeId != "2" {
		t.Error("Fail create node Test!")
		return
	}
	t.Log("Pass create Node Test!")
	//// get test.
	//got := apiserver.GetNode(cli)
	//// fmt.Printf("get nodes string = %v\n", got)
	//res := &defines.GetNodesResource{}
	//_ = json.Unmarshal([]byte(got), res)
	//// fmt.Printf("getNodes res = %v\n", res)
	//if len(res.NodesSend) != 2 {
	//	t.Error("Fail get Node Test!")
	//	return
	//} else {
	//	t.Log("Pass get Node Test!")
	//}
	//// describe test.
	//got1 := apiserver.DescribeNode(cli, "node1")
	//res1 := &defines.NodeResourceSend{}
	//_ = json.Unmarshal([]byte(got1), res1)
	//if res1.NodeId != "1" || res1.ReadyNum != 1 || res1.NodeName != "node1" {
	//	t.Error("Fail describe Node Test!")
	//	return
	//} else {
	//	t.Log("Pass describe Node Test!")
	//}
	// then delete node2 from etcd here.
	etcd.Del(cli, "Node/node2")
}

package node_test

import (
	yamlv3 "gopkg.in/yaml.v3"
	"mini-k8s/pkg/apiserver"
	"mini-k8s/pkg/defines"
	"mini-k8s/pkg/etcd"
	"mini-k8s/utils/yaml"
	"testing"
)

// this test should be made after pod_test is finished.

func TestNode1(t *testing.T) {
	// start etcd client first.
	cli := etcd.EtcdStart()
	defer cli.Close()
	// add pod test.
	yamlPod, _ := yaml.ParsePodConfig("../../utils/templates/pod_template.yaml")
	podInfo := apiserver.CreatePod(cli, yamlPod)
	if podInfo.NodeId != "node1" {
		t.Error("Fail add pod Node Test!")
		return
	} else {
		kv := etcd.Get(cli, "Node/node1").Kvs
		if len(kv) == 0 {
			t.Error("Fail add pod Node Test!")
			return
		} else {
			nodeInfo1 := &defines.NodeInfo{}
			_ = yamlv3.Unmarshal(kv[0].Value, nodeInfo1)
			if len(nodeInfo1.Pods) != 1 {
				t.Error("Fail add pod NOde Test!")
				return
			} else {
				if nodeInfo1.Pods[0].Metadata.Name != "s" {
					t.Error("Fail add pod Node Test!")
					return
				} else {
					t.Log("Pass add pod Node Test!")
				}
			}
		}
	}
	// then delete the created pod.
	// delete pod node test.
	_ = apiserver.DeletePod(cli, "s")
	kv1 := etcd.Get(cli, "Node/node1").Kvs
	if len(kv1) == 0 {
		t.Error("Fail delete pod Node Test!")
		return
	} else {
		nodeInfo2 := &defines.NodeInfo{}
		_ = yamlv3.Unmarshal(kv1[0].Value, nodeInfo2)
		if len(nodeInfo2.Pods) == 0 {
			t.Log("Pass delete pod Node Test!")
		} else {
			t.Error("Fail delete pod Node Test!")
			return
		}
	}
}

package replicaSet_test

import (
	"mini-k8s/pkg/apiserver"
	"mini-k8s/pkg/etcd"
	"mini-k8s/utils/yaml"
	"testing"
)

func TestReplicaBasic(t *testing.T) {
	cli := etcd.EtcdStart()
	defer cli.Close()
	rsYaml, _ := yaml.ParseReplicaSetConfig("../../utils/templates/replicaSet_template.yaml")
	err := apiserver.CreateReplicaSet(cli, rsYaml)
	if err != nil {
		t.Error("Fail createReplicaSet test!")
		return
	} else {
		t.Log("Pass createReplicaSet test!")
	}
	err = apiserver.DeleteReplicaSet(cli, "rs-example")
	if err != nil {
		t.Error("Fail deleteReplicaSet test!")
		return
	} else {
		t.Log("Pass deleteReplicaSet test!")
	}
}

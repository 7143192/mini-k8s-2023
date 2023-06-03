package etcd_test

import (
	yamlv3 "gopkg.in/yaml.v3"
	"mini-k8s/pkg/defines"
	"mini-k8s/pkg/etcd"
	"mini-k8s/utils/yaml"
	"testing"
)

func TestEtcd(t *testing.T) {
	cli := etcd.EtcdStart()
	defer cli.Close()
	if cli == nil {
		t.Error("fil etcd start!")
		return
	}
	// pod Test
	yamlPod, _ := yaml.ParsePodConfig("../../utils/templates/pod_template.yaml")
	testPodKey := "testPod"
	yamlPodByte, _ := yamlv3.Marshal(yamlPod)
	etcd.Put(cli, testPodKey, string(yamlPodByte))
	t.Log("Pass etcd put!")
	kv := etcd.Get(cli, testPodKey).Kvs
	if len(kv) == 0 {
		t.Error("Fail get test!")
		return
	}
	tmp := &defines.YamlPod{}
	_ = yamlv3.Unmarshal(kv[0].Value, tmp)
	if tmp.Metadata.Name == "s" && len(tmp.Spec.Containers) == 2 {
		t.Log("Pass Get Test!")
	} else {
		t.Error("Fail Get Test!")
	}
	etcd.Del(cli, testPodKey)
	kv = etcd.Get(cli, testPodKey).Kvs
	if len(kv) == 0 {
		t.Log("Pass Del Test!")
	} else {
		t.Error("Fail Del Test!")
	}
}

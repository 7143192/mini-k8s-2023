package auto

import (
	"encoding/json"
	"mini-k8s/pkg/apiserver"
	"mini-k8s/pkg/defines"
	"mini-k8s/pkg/etcd"
	"mini-k8s/utils/yaml"
	"testing"
	"time"
)

func TestAutoScaler1(t *testing.T) {
	cli := etcd.EtcdStart()
	defer cli.Close()
	// create auto test.
	autoYaml, _ := yaml.ParseAutoScalerConfig("../../utils/templates/auto_template_1.yaml")
	res := apiserver.CreateAutoScaler(cli, autoYaml)
	if res == nil {
		t.Error("Fail create AutoScaler Test!")
		return
	} else {
		t.Log("Pass create autoScaler Test!")
	}
	// auto status test.
	podInfo := &defines.GetPods{}
	infoStr := apiserver.GetPod(cli)
	_ = json.Unmarshal([]byte(infoStr), podInfo)
	if len(podInfo.PodsSend) != 2 {
		t.Error("Fail get autoScaler test!")
		return
	} else {
		t.Log("Pass get autoScaler test!")
	}
	time.Sleep(60 * time.Second)
	// get again to check shrink status.
	podInfo1 := &defines.GetPods{}
	infoStr1 := apiserver.GetPod(cli)
	_ = json.Unmarshal([]byte(infoStr1), podInfo1)
	if len(podInfo1.PodsSend) != 4 {
		t.Error("Fail auto-extend test!")
		return
	} else {
		t.Log("Pass auto-extend test!")
	}
	// delete test.
	err := apiserver.DeleteAutoScaler(cli, "auto2")
	if err != nil {
		t.Error("Fail delete AutoScaler Test!")
	}
	t.Log("Pass delete autoScaler Test!")
}

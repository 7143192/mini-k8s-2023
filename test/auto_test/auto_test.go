package auto_test

import (
	"mini-k8s/pkg/apiserver"
	"mini-k8s/pkg/etcd"
	"mini-k8s/utils/yaml"
	"testing"
)

//func TestAutoScaler(t *testing.T) {
//	cli := etcd.EtcdStart()
//	defer cli.Close()
//	// create auto test.
//	autoYaml, _ := yaml.ParseAutoScalerConfig("../../utils/templates/auto_template.yaml")
//	res := apiserver.CreateAutoScaler(cli, autoYaml)
//	if res == nil {
//		t.Error("Fail create AutoScaler Test!")
//		return
//	} else {
//		t.Log("Pass create autoScaler Test!")
//	}
//	// auto status test.
//	podInfo := &defines.GetPods{}
//	infoStr := apiserver.GetPod(cli)
//	_ = json.Unmarshal([]byte(infoStr), podInfo)
//	if len(podInfo.PodsSend) != 3 {
//		t.Error("Fail get autoScaler test!")
//		return
//	} else {
//		t.Log("Pass get autoScaler test!")
//	}
//	time.Sleep(45 * time.Second)
//	// get again to check shrink status.
//	podInfo1 := &defines.GetPods{}
//	infoStr1 := apiserver.GetPod(cli)
//	_ = json.Unmarshal([]byte(infoStr1), podInfo1)
//	if len(podInfo1.PodsSend) != 2 {
//		t.Error("Fail auto-shrink test!")
//		return
//	} else {
//		t.Log("Pass auto-shrink test!")
//	}
//	// delete test.
//	err := apiserver.DeleteAutoScaler(cli, "auto1")
//	if err != nil {
//		t.Error("Fail delete AutoScaler Test!")
//	}
//	t.Log("Pass delete autoScaler Test!")
//}

func TestAutoScalerBasic(t *testing.T) {
	cli := etcd.EtcdStart()
	defer cli.Close()
	autoYaml, _ := yaml.ParseAutoScalerConfig("../../utils/templates/auto_template.yaml")
	res := apiserver.CreateAutoScaler(cli, autoYaml)
	if res == nil {
		t.Error("Fail createAutoScaler test!")
		return
	} else {
		t.Log("Pass createAutoScaler test!")
	}
	err := apiserver.DeleteAutoScaler(cli, "auto1")
	if err != nil {
		t.Error("Fail deleteAutoScaler test!")
		return
	} else {
		t.Log("Pass deleteAutoScaler test!")
	}
}

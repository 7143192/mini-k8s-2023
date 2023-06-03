package pod_test

import (
	"mini-k8s/pkg/apiserver"
	"mini-k8s/pkg/etcd"
	"mini-k8s/utils/yaml"
	"testing"
)

func TestPod(t *testing.T) {
	// start etcd client first.
	cli := etcd.EtcdStart()
	defer cli.Close()
	// get yamlPod first.
	yamlPod, _ := yaml.ParsePodConfig("../../utils/templates/pod_template.yaml")
	// create test
	pod := apiserver.CreatePod(cli, yamlPod)
	if pod == nil {
		t.Error("Fail create Pod Test!")
		return
	}
	if pod.Metadata.Name == "s" {
		t.Log("Pass Create pod test!")
	} else {
		t.Error("Fail create Pod Test!")
		return
	}
	//// get and describe test
	//infoStr := apiserver.GetPod(cli)
	//podInfo := &defines.GetPods{}
	//_ = json.Unmarshal([]byte(infoStr), podInfo)
	//if len(podInfo.PodsSend) != 1 {
	//	t.Error("Fail get Pod Test!")
	//	return
	//} else {
	//	t.Log("Pass get Pod Test!")
	//}
	//infoStr1 := apiserver.DescribePod(cli, "s", false)
	//podInfo1 := &defines.PodResourceSend{}
	//_ = json.Unmarshal([]byte(infoStr1), podInfo1)
	//if podInfo1.PodName != "s" || len(podInfo1.ContainerResourcesSend) != 3 {
	//	t.Error("Fail describe Pod Test!")
	//	return
	//} else {
	//	t.Log("Pass describe Pod Test!")
	//}
	// delete pod test.
	err := apiserver.DeletePod(cli, "s")
	//if err != nil {
	//	t.Error("Fail delete Pod Test!")
	//	return
	//}
	//infoStr = apiserver.GetPod(cli)
	//podInfo = &defines.GetPods{}
	//_ = json.Unmarshal([]byte(infoStr), podInfo)
	//if len(podInfo.PodsSend) == 1 {
	//	t.Error("Fail delete Pod Test!")
	//	return
	//} else {
	//	t.Log("Pass delete Pod Test!")
	//}
	if err != nil {
		t.Error("Fail delete Pod Test!")
		return
	} else {
		t.Log("Pass delete Pod Test!")
	}
}

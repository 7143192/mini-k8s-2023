package service_test

import (
	"encoding/json"
	"mini-k8s/pkg/apiserver"
	"mini-k8s/pkg/defines"
	"mini-k8s/pkg/etcd"
	"mini-k8s/utils/yaml"
	"testing"
)

// TODO: in version 2.0, only test ClusterIP service here for now.
func TestService(t *testing.T) {
	path := "../../utils/templates/service_template.yaml"
	path1 := "../../utils/templates/pod_template.yaml"
	path2 := "../../utils/templates/pod_template_2.yaml"
	// get service yaml info first.
	yamlSvc, _ := yaml.ParseClusterIPConfig(path)
	// start etcd client here.
	cli := etcd.EtcdStart()
	defer cli.Close()
	// create service test.
	svc := apiserver.CreateClusterIPSvc(cli, yamlSvc)
	if svc.SvcName != "service1" || svc.SvcType != "ClusterIP" || len(svc.SvcPorts) != 1 {
		t.Error("Fail create Svc Test!")
		return
	}
	t.Log("Pass create Svc Test!")
	// get service test.
	getStr := apiserver.GetService(cli)
	getRes := &defines.ServiceInfoSend{}
	_ = json.Unmarshal([]byte(getStr), getRes)
	if len(getRes.SvcInfos) == 1 {
		t.Log("Pass get Svc Test!")
	} else {
		t.Error("Fail get Svc Test!")
		return
	}
	// describe service test.
	describeStr := apiserver.DescribeService(cli, "service1")
	describeRes := &defines.ServiceInfo{}
	_ = json.Unmarshal([]byte(describeStr), describeRes)
	if describeRes.SvcBriefInfo.SvcName == "service1" && describeRes.SvcBriefInfo.SvcType == "ClusterIP" {
		t.Log("Pass describe Svc Test!")
	} else {
		t.Error("Fail describe Svc Test!")
		return
	}
	// add pod service test.
	yamlPod1, _ := yaml.ParsePodConfig(path1)
	yamlPod2, _ := yaml.ParsePodConfig(path2)
	_ = apiserver.CreatePod(cli, yamlPod1)
	_ = apiserver.CreatePod(cli, yamlPod2)
	describeStr = apiserver.DescribeService(cli, "service1")
	describeRes = &defines.ServiceInfo{}
	_ = json.Unmarshal([]byte(describeStr), describeRes)
	if len(describeRes.SvcPods) != 2 {
		t.Error("Fail add pod Svc Test!")
		return
	} else {
		if describeRes.SvcPods[0].Metadata.Name == "s" && describeRes.SvcPods[1].Metadata.Name == "sss" {
			t.Log("Pass add pod Svc Test!")
		} else {
			t.Error("Fail add pod Svc Test!")
			return
		}
	}
	// delete pod service test.
	_ = apiserver.DeletePod(cli, "s")
	describeStr = apiserver.DescribeService(cli, "service1")
	describeRes = &defines.ServiceInfo{}
	_ = json.Unmarshal([]byte(describeStr), describeRes)
	if len(describeRes.SvcPods) != 1 {
		t.Error("Fail delete pod Svc Test!")
		return
	} else {
		if describeRes.SvcPods[0].Metadata.Name == "sss" {
			t.Log("Pass delete pod Svc Test!")
		} else {
			t.Error("Fail delete pod Svc Test!")
			return
		}
	}
	_ = apiserver.DeletePod(cli, "sss")
	describeStr = apiserver.DescribeService(cli, "service1")
	describeRes = &defines.ServiceInfo{}
	_ = json.Unmarshal([]byte(describeStr), describeRes)
	if len(describeRes.SvcPods) != 0 {
		t.Error("Fail delete pod Svc Test!")
		return
	} else {
		t.Log("Pass delete pod Svc Test!")
	}
}

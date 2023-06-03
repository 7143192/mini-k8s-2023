package yaml_test

import (
	"mini-k8s/pkg/defines"
	"mini-k8s/utils/yaml"
	"testing"
)

func Test(t *testing.T) {
	path1 := "../../utils/templates/pod_template.yaml"
	path2 := "../../utils/templates/node_template.yaml"
	path3 := "../../utils/templates/service_template.yaml"
	// test reading pod yaml.
	t1, err := yaml.ParseYamlKind(path1)
	if err != nil {
		t.Error("error occurs when parsing yaml kind!")
		return
	}
	if t1 != defines.POD {
		t.Error("error occurs when getting yaml pod type!")
	} else {
		t.Log("Pass Pod Type!")
	}
	pod, err := yaml.ParsePodConfig(path1)
	if err != nil {
		t.Error("error occurs when reading pod yaml config!")
		return
	}
	if pod.Metadata.Name != "s" || len(pod.Spec.Containers) != 2 {
		t.Error("fail pod config test!")
	} else {
		t.Log("Pass pod config test!")
	}
	// test reading node yaml.
	t2, err := yaml.ParseYamlKind(path2)
	if err != nil {
		t.Error("error occurs when parsing yaml kind!")
		return
	}
	if t2 != defines.NODE {
		t.Error("error occurs when getting yaml node type!")
	} else {
		t.Log("Pass Node Type!")
	}
	node, err := yaml.ParseNodeConfig(path2)
	if err != nil {
		t.Error("error occurs when reading node yaml config!")
		return
	}
	if node.Metadata.Name != "node1" || node.Metadata.Label != "node1" {
		t.Error("fail node config test!")
	} else {
		t.Log("Pass node config test!")
	}
	// test reading service yaml.
	t3, err := yaml.ParseYamlKind(path3)
	if err != nil {
		t.Error("error occurs when parsing yaml kind!")
		return
	}
	if t3 != defines.SERVICE {
		t.Error("error occurs when getting yaml service type!")
	} else {
		t.Log("Pass Service Type!")
	}
	svc, err := yaml.ParseClusterIPConfig(path3)
	if err != nil {
		t.Error("error occurs when reading service yaml config!")
		return
	}
	if svc.Metadata.Name != "service1" || svc.Spec.Type != "ClusterIP" {
		t.Error("fail service config test!")
	} else {
		t.Log("Pass service config test!")
	}
}

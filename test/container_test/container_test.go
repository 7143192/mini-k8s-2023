package container_test

import (
	"github.com/docker/docker/client"
	"mini-k8s/pkg/config"
	"mini-k8s/pkg/container"
	"mini-k8s/pkg/defines"
	"mini-k8s/pkg/etcd"
	"mini-k8s/utils/yaml"
	"testing"
	"time"
)

func TestContainer(t *testing.T) {
	// create docker client first.
	cli, _ := client.NewClientWithOpts(client.WithVersion(config.DockerApiVersion))
	etcdCli := etcd.EtcdStart()
	defer etcdCli.Close()
	yamlPod, _ := yaml.ParsePodConfig("../../utils/templates/pod_template_2.yaml")
	pod := &defines.Pod{
		PodId:           "PodInstance/sss",
		PodIp:           "10.0.0.1",
		PodState:        defines.Running,
		ContainerStates: make([]defines.ContainerState, 0),
		Start:           time.Now(),
		NodeId:          "node1",
		YamlPod:         *yamlPod,
	}
	// create container test.
	res := container.CreatePauseContainer(pod, "", cli)
	if res == nil {
		t.Error("Fail create Container Test!")
		return
	} else {
		t.Log("Pass create Container Test!")
	}
	id := res.Id
	// inspect test.
	inspect := container.InspectContainer(id)
	if inspect.Name == "" {
		t.Error("Fail inspect Container Test!")
		return
	} else {
		t.Log("Pass inspect Container Test!")
	}
	// stop test.
	stopped := container.StopContainer(id)
	if stopped == false {
		t.Error("Fail stop Container test!")
		return
	} else {
		t.Log("Pass stop Container Test!")
	}
	// start test.
	err := container.StartPauseContainer(id)
	if err != nil {
		t.Error("Fail start Container Test!")
		return
	} else {
		t.Log("Pass start Container Test!")
	}
	// stop again to do restart test.
	_ = container.StopContainer(id)
	err = container.RestartContainer(id)
	if err != nil {
		t.Error("Fail restart Container Test!")
		return
	} else {
		t.Log("Pass restart Container Test!")
	}
	// stop again to do remove test.
	_ = container.StopContainer(id)
	removed := container.RemoveContainer(id)
	if removed == false {
		t.Error("Fail remove Container Test!")
		return
	} else {
		t.Log("Pass remove Container Test!")
	}
}

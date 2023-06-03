package race_test

import (
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"mini-k8s/pkg/apiserver"
	"mini-k8s/pkg/defines"
	"mini-k8s/pkg/etcd"
	"mini-k8s/utils/yaml"
	"testing"
	"time"
)

func ParallelCreate(cli *clientv3.Client, yamlPod *defines.YamlPod, i int) {
	// fmt.Printf("i = %v\n", i)
	tmpYaml := *yamlPod
	for j := 0; j < i+1; j++ {
		tmpYaml.Metadata.Name += "s"
	}
	fmt.Printf("pod name: %v\n", tmpYaml.Metadata.Name)
	apiserver.CreatePod(cli, &tmpYaml)
}

func ParallelDelete(cli *clientv3.Client, yamlPod *defines.YamlPod, i int) {
	// fmt.Printf("i = %v\n", i)
	tmpYaml := *yamlPod
	for j := 0; j < i+1; j++ {
		tmpYaml.Metadata.Name += "s"
	}
	fmt.Printf("pod name: %v\n", tmpYaml.Metadata.Name)
	apiserver.DeletePod(cli, tmpYaml.Metadata.Name)
}

func TestRace(t *testing.T) {
	cli := etcd.EtcdStart()
	yamlPod, _ := yaml.ParsePodConfig("../../utils/templates/pod_template_1.yaml")
	for i := 0; i < 20; i++ {
		// go ParallelCreate(cli, yamlPod, i)
		go ParallelDelete(cli, yamlPod, i)
	}
	for {
		time.Sleep(10 * time.Second)
	}
}

package main

import (
	"encoding/json"
	"gopkg.in/yaml.v3"
	"log"
	"mini-k8s/pkg/defines"
	"os"
)

func main() {
	//cli := etcd.EtcdStart()
	file, err := os.Open("./utils/templates/replicaSet_template.yaml")
	if err != nil {
		log.Fatalf("[ERROR] %v\n", err)
	}
	//yamlPod := &defines.YamlPod{}
	//err = yaml.NewDecoder(file).Decode(yamlPod)
	replicaSet := &defines.ReplicaSet{}
	err = yaml.NewDecoder(file).Decode(replicaSet)
	if err != nil {
		log.Fatalf("[ERROR] %v\n", err)
	}
	//pod := &defines.Pod{
	//	YamlPod:  *yamlPod,
	//	PodState: 1,
	//}
	write, err := os.Create("./test/json/replicaSet_template.json")
	if err != nil {
		log.Fatalf("[ERROR] %v\n", err)
	}
	replicaSetVal, err := json.Marshal(replicaSet)
	//podVal, err := json.Marshal(pod)
	if err != nil {
		log.Fatalf("[ERROR] %v\n", err)
	}
	_, err = write.Write(replicaSetVal)
	//_, err = write.Write(podVal)
	if err != nil {
		log.Fatalf("[ERROR] %v\n", err)
	}

	//etcd.Put(cli, "PodInstance/"+pod.Metadata.Name, string(podVal))
	//_ = etcd.EtcdEnd(cli)
}

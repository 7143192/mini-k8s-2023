package main

import (
	"mini-k8s/pkg/apiserver"
	"mini-k8s/pkg/controller/replicaSet"
)

func main() {
	rc := replicaSet.NewReplicaController()
	server := apiserver.RCServerInit()
	server.ReplicaController = rc

	go rc.Watch()

	err := server.RCServerRun()
	if err != nil {
		panic(err)
	}
}

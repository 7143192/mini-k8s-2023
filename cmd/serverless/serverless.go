package main

import (
	"log"
	"mini-k8s/pkg/apiserver"
	"mini-k8s/pkg/serverless"
)

func main() {
	SLController := serverless.SLControllerInit()

	server := apiserver.SLServerInit()
	server.ServerlessController = SLController

	go SLController.Watch()

	err := server.SLServerRun()
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
	}
}

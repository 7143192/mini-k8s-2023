package main

import "mini-k8s/pkg/apiserver"

func main() {
	server := apiserver.TestServerInit()
	err := server.TestServerRun()
	if err != nil {
		panic(err)
	}
}

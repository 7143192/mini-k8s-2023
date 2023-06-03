package container

import (
	"context"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"log"
	"mini-k8s/pkg/config"
)

func StopContainer(id string) bool {
	cli, err := client.NewClientWithOpts(client.WithVersion(config.DockerApiVersion))
	defer cli.Close()
	if err != nil {
		log.Printf("an error occurs when creating a client in StopContainer func :%v\n", err)
		return false
	}
	err = cli.ContainerStop(context.Background(), id, container.StopOptions{})
	if err != nil {
		log.Printf("an error occurs when stop container %s: %v\n", id, err)
		return false
	}
	log.Printf("stop container %s successfully!\n", id)
	return true
}

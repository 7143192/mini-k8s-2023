package container

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"log"
	"mini-k8s/pkg/config"
)

func RemoveContainer(id string) bool {
	cli, err := client.NewClientWithOpts(client.WithVersion(config.DockerApiVersion))
	defer cli.Close()
	if err != nil {
		log.Printf("an error occurs when creating client in RemoveContainer func :%v\n", err)
		return false
	}
	err = cli.ContainerRemove(context.Background(), id, types.ContainerRemoveOptions{})
	if err != nil {
		log.Printf("an error occurs when stop container %s: %v\n", id, err)
		return false
	}
	log.Printf("remove container %s successfully!\n", id)
	return true
}

func RemoveForceContainer(id string) bool {
	cli, err := client.NewClientWithOpts(client.WithVersion(config.DockerApiVersion))
	defer cli.Close()
	if err != nil {
		log.Printf("an error occurs when creating client in RemoveContainer func :%v\n", err)
		return false
	}
	err = cli.ContainerRemove(context.Background(), id, types.ContainerRemoveOptions{Force: true})
	if err != nil {
		log.Printf("an error occurs when stop container %s: %v\n", id, err)
		return false
	}
	log.Printf("remove container %s successfully!\n", id)
	return true
}

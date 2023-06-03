package container

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"log"
	"mini-k8s/pkg/config"
)

// StartPauseContainer is used to start a pause container.
func StartPauseContainer(containerID string) error {
	cli, err := client.NewClientWithOpts(client.WithVersion(config.DockerApiVersion))
	defer cli.Close()
	if err != nil {
		log.Printf("an error occurs when starting a pause container(1):%v\n", err)
		return err
	}
	containerStart := types.ContainerStartOptions{}
	err1 := cli.ContainerStart(context.Background(), containerID, containerStart)
	if err1 != nil {
		log.Printf("an error occurs when starting pause container(2):%v\n", err1)
		return err1
	}
	log.Println("successfully start a new pause container !")
	return nil
}

// StartNormalContainer is used to start a normal container.
func StartNormalContainer(containerID string) {
	cli, err := client.NewClientWithOpts(client.WithVersion(config.DockerApiVersion))
	defer cli.Close()
	if err != nil {
		log.Printf("an error occurs when starting a normal container(1):%v\n", err)
		return
	}
	containerStart := types.ContainerStartOptions{}
	err1 := cli.ContainerStart(context.Background(), containerID, containerStart)
	if err1 != nil {
		log.Printf("an error occurs when starting normal container(2): %v\n", err1)
		return
	}
	return
}

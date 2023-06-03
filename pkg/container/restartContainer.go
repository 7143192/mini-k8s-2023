package container

import (
	"context"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"log"
	"mini-k8s/pkg/config"
)

func RestartContainer(containerID string) error {
	cli, err := client.NewClientWithOpts(client.WithVersion(config.DockerApiVersion))
	defer cli.Close()
	if err != nil {
		log.Printf("an error occurs when creating a docker client in func RestartContainer: %v\n", err)
		return err
	}
	// NOTE: in version 2.0, add race check here.
	names := ListContainer()
	found := false
	for _, name := range names {
		if name == containerID {
			found = true
			break
		}
	}
	if found == false {
		// the container that want to be restarted is not in the system (maybe caused by race conditions. )
		log.Printf("the restart container %v does not exist in the current system!\n", containerID)
		return nil
	}
	err = cli.ContainerRestart(context.Background(), containerID, container.StopOptions{})
	if err != nil {
		log.Printf("an error occurs when restart container %v: %v\n", containerID, err)
		return err
	}
	log.Printf("restart container %v successfully!\n", containerID)
	return nil
}

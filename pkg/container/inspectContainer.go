package container

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"log"
	"mini-k8s/pkg/config"
)

// InspectContainer docker inspect containerID
func InspectContainer(containerID string) types.ContainerJSON {
	cli, err := client.NewClientWithOpts(client.WithVersion(config.DockerApiVersion))
	if err != nil {
		log.Printf("an error occurs when creating a docker client in func inspectContainer: %v\n", err)
		return types.ContainerJSON{}
	}
	defer cli.Close()
	res, err := cli.ContainerInspect(context.Background(), containerID)
	if err != nil {
		log.Printf("an error occurs when getting docker info in func InspectContainer: %v\n", err)
		return types.ContainerJSON{}
	}
	return res
}

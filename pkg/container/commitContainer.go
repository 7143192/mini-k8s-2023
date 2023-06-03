package container

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"log"
	"mini-k8s/pkg/config"
)

func CommitContainer(containerID string) string {
	cli, _ := client.NewClientWithOpts(client.WithVersion(config.DockerApiVersion))
	defer cli.Close()
	info := InspectContainer(containerID)
	response, err := cli.ContainerCommit(context.Background(), containerID, types.ContainerCommitOptions{Config: info.Config})
	if err != nil {
		log.Printf("[container] fail to commit a container: %v\n", err)
		return response.ID
	}
	// fmt.Printf("error when commit a container: %v\n", err)
	return response.ID
}

package container

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"log"
	"mini-k8s/pkg/config"
)

func ListContainer() []string {
	cli, err := client.NewClientWithOpts(client.WithVersion(config.DockerApiVersion))
	defer cli.Close()
	if err != nil {
		log.Printf("error when creating a new docker client in function ListContainer: %v\n", err)
		panic(err)
	}
	res, err := cli.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		log.Printf("error when listing all containers info: %v\n", err)
		panic(err)
	}
	names := make([]string, 0)
	for _, con := range res {
		names = append(names, con.ID)
	}
	return names
}

package container

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"log"
	"mini-k8s/pkg/config"
)

func ListNetwork() []types.NetworkResource {
	cli, err := client.NewClientWithOpts(client.WithVersion(config.DockerApiVersion))
	if err != nil {
		log.Printf("%v\n", err)
		panic(err)
	}
	defer cli.Close()
	networks, err := cli.NetworkList(context.Background(), types.NetworkListOptions{})
	if err != nil {
		log.Printf("%v\n", err)
		panic(err)
	}
	return networks
}

package container

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"log"
	"mini-k8s/pkg/config"
	"strings"
)

// CreateNetwork receives a cni IP as param, and return the new network ID.
func CreateNetwork(IP string, networkName string) string {
	cli, err := client.NewClientWithOpts(client.WithVersion(config.DockerApiVersion))
	if err != nil {
		log.Printf("createNetwork: %v\n", err)
		panic(err)
	}
	defer cli.Close()
	networks := ListNetwork()
	// try to find duplicate network name.
	for _, net := range networks {
		if net.Name == networkName {
			fmt.Printf("the network %v has already been created!\n", networkName)
			return net.ID
		}
	}
	createOptions := types.NetworkCreate{}
	ipamConfigs := make([]network.IPAMConfig, 0)
	ipamConfig := network.IPAMConfig{}
	idx := strings.LastIndex(IP, ".")
	tmpIP := IP[0:idx]
	gateway := tmpIP + ".1"
	ipamConfig.Subnet = IP + "/24"
	ipamConfig.Gateway = gateway
	ipamConfigs = append(ipamConfigs, ipamConfig)
	Ipam := &network.IPAM{}
	Ipam.Config = ipamConfigs
	options := make(map[string]string)
	options["com.docker.network.bridge.name"] = networkName
	createOptions.Options = options
	createOptions.IPAM = Ipam
	resp, err := cli.NetworkCreate(context.Background(), networkName, createOptions)
	if err != nil {
		fmt.Printf("an error occurs when create network: %v\n", err)
		return ""
	}
	fmt.Printf("a new network %v (with name %v) has been created successfully!\n", resp.ID, networkName)
	return resp.ID
}

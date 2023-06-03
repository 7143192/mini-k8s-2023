package container

import (
	"context"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	uuid "github.com/satori/go.uuid"
	"log"
	"mini-k8s/pkg/config"
	defines2 "mini-k8s/pkg/defines"
	"strings"
)

// CreateNormalContainer used to create a normal(not pause container) container. return container id in this function.
func CreateNormalContainer(cli *client.Client, podContainer *defines2.PodContainer, mode string, pod *defines2.Pod) defines2.ContainerID {
	UUID := uuid.NewV4()
	prefix := pod.PodId
	prefix = strings.Replace(prefix, "/", "-", -1)
	// generate a UUID for the newly-created normal container.
	// unique := uuid.NewV4()
	name := prefix + "-" + podContainer.Name + "-" + UUID.String()
	// fmt.Printf("new normal container name = %s\n", name)
	if CheckImageExist(podContainer.Image) == false {
		// fmt.Printf("target image = %s\n", podContainer.Image)
		// fmt.Println("ready to pull a image for normal container!")
		PullImage(podContainer.Image)
	}
	containerConfig := GetContainerConfig(podContainer)
	containerHostConfig := GetContainerHostConfig(podContainer, mode, pod.Spec.Volumes)
	SetNetNamespaceShareLocal(mode, containerHostConfig)
	containerNetworkConfig := &network.NetworkingConfig{}
	response, err := cli.ContainerCreate(context.Background(), containerConfig, containerHostConfig, containerNetworkConfig, nil, name)
	if err != nil {
		log.Printf("an error occurs when creating a new normal container : %v\n", err)
		return "-1"
	}
	// fmt.Printf("response id = %s\n", response.ID)
	//// TODO: in version 1.0, try to start this container directly but in the future this start should be managed by controller.
	//StartNormalContainer(response.ID)
	return defines2.ContainerID(response.ID)
}

// CreatePauseContainer used to create a pause container, this container should be used to
// allocate network to every container in this pod and should act as the "leader" when working.
func CreatePauseContainer(pod *defines2.Pod, networkID string, cli *client.Client) *defines2.ContainerState {
	// generate a new UUID for the name of pause container.
	UUID := uuid.NewV4()
	id := pod.PodId
	// fmt.Printf("old id = %s\n", id)
	id = strings.Replace(id, "/", "-", -1)
	// fmt.Printf("new id = %s\n", id)
	pauseContainerName := id + "-pause-" + UUID.String()
	containerConfig := &container.Config{}
	hostConfig := &container.HostConfig{}
	portMap, portSet := GetContainerPortsInfo(pod)
	containerConfig.Image = defines2.PauseContainerImage
	// add image check and pull logic.
	if CheckImageExist(containerConfig.Image) == false {
		// fmt.Println("ready to pull a new image!")
		ans := PullImage(containerConfig.Image)
		if ans == false {
			log.Printf("an error occurs when pulling a image %s\n", containerConfig.Image)
			return nil
		}
	}
	// fmt.Printf("image %s has already existed locally!\n", config.Image)
	containerConfig.ExposedPorts = portSet
	hostConfig.PortBindings = portMap
	//inspect, err := cli.ContainerInspect(context.Background(), config.CoreDNSServerName)
	//if err != nil {
	//	panic(err)
	//}
	//hostConfig.DNS = []string{inspect.NetworkSettings.IPAddress}
	hostConfig.DNS = []string{config.CoreDNSServerIP}
	// in version 1.0, do not consider the network.
	// TODO: in version 2.0, add networking config here to support network.
	// networkID := defines2.PodNetworkIDPrefix + pod.Metadata.Name
	networkingConfig := &network.NetworkingConfig{}
	// networkingConfig := GetContainerNetworkingConfig(networkID)
	response, err := cli.ContainerCreate(context.Background(), containerConfig, hostConfig, networkingConfig, nil, pauseContainerName)
	if err != nil {
		log.Printf("an error occurs when creating pause cntainer :%v\n", err)
		return nil
	}
	// TODO: in version 2.0, add a new network here to support network.
	// podNamespace := defines2.PodNsPathPrefix + config1.NsNamePrefix + pod.Metadata.Name
	// network1.AddNewNetwork(networkID, podNamespace, response.ID)
	// should start pause container here directly.
	// fmt.Println("ready to start a pause container!")
	StartPauseContainer(response.ID)
	ans := &defines2.ContainerState{}
	ans.Id = response.ID
	ans.Name = pauseContainerName
	ans.State = defines2.Running
	return ans
}

func CreateGPUJobContainer(conName string, con *defines2.PodContainer) defines2.ContainerID {
	cli, _ := client.NewClientWithOpts(client.WithVersion(config.DockerApiVersion))
	containerConfig := GetContainerConfig(con)
	hostConfig := &container.HostConfig{}
	networkConfig := &network.NetworkingConfig{}
	id, _ := cli.ContainerCreate(context.Background(), containerConfig, hostConfig, networkConfig, nil, conName)
	return defines2.ContainerID(id.ID)
}

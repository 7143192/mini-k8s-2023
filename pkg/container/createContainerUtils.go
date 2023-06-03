package container

import (
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/go-connections/nat"
	"log"
	defines2 "mini-k8s/pkg/defines"
	"strconv"
	"strings"
)

func GetContainerConfig(podContainer *defines2.PodContainer) *container.Config {
	res := &container.Config{}
	res.Image = podContainer.Image
	res.WorkingDir = podContainer.WorkingDir
	// set tty = true to make this container stay "up" continuously.
	res.Tty = true
	if len(podContainer.Command) != 0 {
		res.Entrypoint = podContainer.Command
	}
	if len(podContainer.Args) != 0 {
		res.Cmd = podContainer.Args
	}
	return res
}

// SetNetNamespaceShareLocal this function is used to make normal containers to share the network namespace of pause container.
func SetNetNamespaceShareLocal(mode string, hostConfig *container.HostConfig) {
	hostConfig.NetworkMode = container.NetworkMode(mode)
}

func GetContainerHostConfig(podContainer *defines2.PodContainer, namespace string, volumes []defines2.PodVolume) *container.HostConfig {
	// TODO(lyh): How to present a namespace ???
	res := &container.HostConfig{}
	resource := &container.Resources{}
	cpu := podContainer.Resource.ResourceLimit.Cpu
	memory := podContainer.Resource.ResourceLimit.Memory
	if len(cpu) != 0 {
		// in version 2.0, support configured CPU < 1.00 .
		cpuLen := len(cpu)
		cpuLimit := 0
		//if cpu[0] == '0' {
		//	idx := strings.Index(cpu, ".")
		//	intCpu := strings.Replace(cpu, ".", "", -1)
		//	gotIntCpu, _ := strconv.Atoi(intCpu)
		//	cpuLimit = gotIntCpu * 10e9
		//	backLen := cpuLen - idx - 1
		//	for i := 0; i < backLen; i++ {
		//		cpuLimit = cpuLimit / 10
		//	}
		//} else {
		//	gotCpu, _ := strconv.Atoi(cpu)
		//	cpuLimit = gotCpu * 10e9
		//}
		if cpu[cpuLen-1] == 'm' {
			// micro-CPU
			val := cpu[0 : cpuLen-1]
			gotVal, _ := strconv.Atoi(val)
			cpuLimit = gotVal * 1e6
		} else {
			// 0.XX
			if cpu[0] == '0' {
				idx := strings.Index(cpu, ".")
				intCpu := strings.Replace(cpu, ".", "", -1)
				gotIntCpu, _ := strconv.Atoi(intCpu)
				cpuLimit = gotIntCpu * 10e9
				backLen := cpuLen - idx - 1
				for i := 0; i < backLen; i++ {
					cpuLimit = cpuLimit / 10
				}
			} else {
				// a.XX (a != 0)
				gotCpu, _ := strconv.Atoi(cpu)
				cpuLimit = gotCpu * 1e9
			}
		}
		resource.NanoCPUs = int64(cpuLimit)
	}
	if len(memory) != 0 {
		memType := memory[len(memory)-2:]
		memVal := memory[:len(memory)-2]
		gotVal := 0
		switch memType {
		case "KB":
			got, _ := strconv.Atoi(memVal)
			gotVal = got << 10
		case "MB":
			got1, _ := strconv.Atoi(memVal)
			gotVal = got1 << 20
		case "GB":
			got2, _ := strconv.Atoi(memVal)
			gotVal = got2 << 30
		}
		resource.Memory = int64(gotVal)
	}
	mounts := make([]mount.Mount, 0)
	vol := make([]*defines2.PodVolume, 0)
	for _, volume := range volumes {
		for _, volume1 := range podContainer.VolumeMounts {
			if volume.Name == volume1.Name {
				//
				tmp := &defines2.PodVolume{Name: volume.HostPath, HostPath: volume1.MountPath}
				vol = append(vol, tmp)
			}
		}
	}
	// use "bind" type to support file sharing between host and containers.
	for _, v := range vol {
		m := mount.Mount{
			Type:   "bind",
			Source: v.Name,
			Target: v.HostPath,
		}
		mounts = append(mounts, m)
	}
	// set the res hostConfig.
	res.Resources = *resource
	res.Mounts = mounts
	// ???
	//res.IpcMode = container.IpcMode(namespace)
	res.IpcMode = "shareable"
	// Through setting networkMode to "container:pauseContainer-ID" to share pod local network namespace.
	// res.NetworkMode = container.NetworkMode(namespace)
	res.PidMode = container.PidMode(namespace)
	return res
}

//// GetContainerNetworkingConfig TODO: in version 1.0, don't consider the network.
//func GetContainerNetworkingConfig(networkID string) *network.NetworkingConfig {
//	return &network.NetworkingConfig{
//		EndpointsConfig: map[string]*network.EndpointSettings{
//			networkID: &network.EndpointSettings{
//				NetworkID: networkID,
//				// NetworkDriver:
//			},
//		},
//	}
//}

func GetContainerPortsInfo(pod *defines2.Pod) (nat.PortMap, nat.PortSet) {
	containers := pod.YamlPod.Spec.Containers
	portMap := make(nat.PortMap)
	portSet := make(nat.PortSet)
	for _, con := range containers {
		ports := con.Ports
		for _, port := range ports {
			// fmt.Println(port)
			containerPort := port.ContainerPort
			// hostPort := port.HostPort
			// default container port is 8080.
			if containerPort == 0 {
				containerPort = 8080
			}
			//// default hostPort is 80.
			//if hostPort == 0 {
			//	hostPort = 80
			//}
			newPort, err := nat.NewPort(port.Protocol, strconv.Itoa(containerPort))
			if err != nil {
				log.Printf("an error occurs when getting container port information: %v\n", err)
				return nil, nil
			}
			bindings := make([]nat.PortBinding, 1)
			// binding := nat.PortBinding{HostIP: pod.PodIp, HostPort: strconv.Itoa(hostPort)}
			// in version 2.0, change here!!
			// binding := nat.PortBinding{HostIP: "127.0.0.1", HostPort: strconv.Itoa(hostPort)}
			binding := nat.PortBinding{}
			bindings[0] = binding
			portMap[newPort] = bindings
			portSet[newPort] = struct{}{}
		}
	}
	return portMap, portSet
}

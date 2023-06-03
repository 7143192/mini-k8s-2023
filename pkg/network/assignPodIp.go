package network

import (
	"mini-k8s/pkg/config"
)

// AssignPodIp kubelet should call the function when create a new pod to assign its ip addr
// TODO: kubelet may need to create a new network namespace
func AssignPodIp(Netns string, id string) string {
	pluginDir := []string{config.FlannelPluginDir}
	cni := CNIStart(pluginDir, config.FlannelConfDir)
	// remove old one first.
	RemoveNetns(cni, Netns, id)
	Ip := SetupNetns(cni, Netns, id)
	return Ip
}

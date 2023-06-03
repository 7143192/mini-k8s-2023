package network

import (
	"mini-k8s/pkg/config"
	"mini-k8s/pkg/defines"
)

func GetNamespace(id string) string {
	res := defines.PodNsPathPrefix + config.NsNamePrefix + id
	return res
}

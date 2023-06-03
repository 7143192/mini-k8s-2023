package network

import (
	clientv3 "go.etcd.io/etcd/client/v3"
	"mini-k8s/pkg/config"
	"mini-k8s/pkg/etcd"
)

// FlannelInit call this function when mini-k8s start
func FlannelInit(cli *clientv3.Client) {
	key := config.FlannelEtcdPrefix + "/config"
	// config := "{ \"Network\": \"10.5.0.0/16\", \"Backend\": {\"Type\": \"vxlan\"}}"
	configName := "{ \"Network\": " + "\"" + config.FlannelIP + "\", \"Backend\": {\"Type\": \"vxlan\"}}"
	etcd.Put(cli, key, configName)
}

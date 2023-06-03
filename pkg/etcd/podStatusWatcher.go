package etcd

import (
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"mini-k8s/pkg/config"
	"mini-k8s/pkg/defines"
	"time"
)

/* this file is put here as pod status watcher can be implemented by etcd APIs. */

// StartPodWatcher is used to start a watcher for a pod.
// TODO: this function watches pod directly because we do not implement NODE object for now.
// TODO: the real usage of watcher should be that a node starts Node-level watcher to watch all pod states in this node!
func StartPodWatcher(pod *defines.Pod) {
	id := "PodInstance/" + pod.PodId
	points := make([]string, 3)
	points[0] = config.IP + ":" + config.EtcdPort1
	points[1] = config.IP + ":" + config.EtcdPort2
	points[2] = config.IP + ":" + config.EtcdPort3
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   points,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Printf("an error occurs when start an etcd client instance in StartPodWatcher: %v\n", err)
		return
	}
	defer cli.Close()
	watchChan := WatchNew(cli, id)
	for resp := range watchChan {
		for _, ev := range resp.Events {
			fmt.Printf("Type: %s Key:%s Value:%s\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
		}
	}
}

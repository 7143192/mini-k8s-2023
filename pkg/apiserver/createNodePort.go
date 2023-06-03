package apiserver

import (
	clientv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v3"
	"log"
	"mini-k8s/pkg/defines"
	"mini-k8s/pkg/etcd"
	"time"
)

func CreateNodePortSvc(cli *clientv3.Client, npSvc *defines.NodePortService) *defines.EtcdService {
	// check service with duplicate name here.
	originList := make([]string, 0)
	oldListKey := defines.AllServiceSetPrefix + "/"
	oldKv := etcd.Get(cli, oldListKey).Kvs
	if len(oldKv) != 0 {
		_ = yaml.Unmarshal(oldKv[0].Value, &originList)
		for _, item := range originList {
			if item == (defines.ServicePrefix + "/" + npSvc.Metadata.Name) {
				log.Printf("[apiserver] the service %v has already been creted!\n", npSvc.Metadata.Name)
				oldSvcKey := item
				oldSvcKv := etcd.Get(cli, oldSvcKey).Kvs
				tmp := &defines.EtcdService{}
				_ = yaml.Unmarshal(oldSvcKv[0].Value, tmp)
				// if already been created, return the old one.
				return tmp
			}
		}
	}
	// first create an etcd service object
	res := &defines.EtcdService{}
	res.SvcName = npSvc.Metadata.Name
	res.SvcType = "NodePort"
	res.SvcClusterIP = npSvc.Spec.ClusterIP
	res.SvcSelector = npSvc.Spec.Selector
	res.SvcNodePort = npSvc.Spec.Ports[0].NodePort
	res.SvcPorts = npSvc.Spec.Ports
	// get all pods' info that satisfy the service's pod selector.
	ServiceSelectPods(cli, res)
	res.SvcStartTime = time.Now()
	// then store the new object into etcd.
	svcKey := defines.ServicePrefix + "/" + res.SvcName
	resByte, _ := yaml.Marshal(res)
	etcd.Put(cli, svcKey, string(resByte))
	// then store the new service list.(to watch changes)
	svcListKey := defines.AllServiceSetPrefix + "/"
	oldList := make([]string, 0)
	kv := etcd.Get(cli, svcListKey).Kvs
	if len(kv) == 0 {
		// the first service case.
		newList := append(oldList, svcKey)
		newListByte, _ := yaml.Marshal(&newList)
		etcd.Put(cli, svcListKey, string(newListByte))
	} else {
		// normal case.
		_ = yaml.Unmarshal(kv[0].Value, &oldList)
		newList := append(oldList, svcKey)
		newListByte, _ := yaml.Marshal(&newList)
		etcd.Put(cli, svcListKey, string(newListByte))
	}
	// then do some other jobs according to Kube proxy(TODO: not sure what to do for now)
	_ = SendOutCreateService(cli, res)
	return res
}

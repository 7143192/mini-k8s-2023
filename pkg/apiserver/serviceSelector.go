package apiserver

import (
	clientv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v3"
	"log"
	"mini-k8s/pkg/defines"
	"mini-k8s/pkg/etcd"
)

func ServiceSelectPods(cli *clientv3.Client, svc *defines.EtcdService) {
	selector := svc.SvcSelector
	allPods := make([]*defines.Pod, 0)
	selectedPods := make([]*defines.Pod, 0)
	// NOTE: in version 2.0, we only select pods according to app and env labels,if the label value is "", ignore this label selector.
	// get all pods in the etcd first.
	prefixKey := defines.PodInstancePrefix + "/"
	kvs := etcd.GetWithPrefix(cli, prefixKey).Kvs
	if len(kvs) == 0 {
		log.Println("[serviceSelector] no pods in the current system!")
		svc.SvcPods = selectedPods // empty array.
		return
	}
	for _, kv := range kvs {
		pod := &defines.Pod{}
		_ = yaml.Unmarshal(kv.Value, pod)
		allPods = append(allPods, pod)
	}
	// if not selector is legal(all ""), then return directly.
	if selector.App == "" && selector.Env == "" {
		svc.SvcPods = allPods
		return
	}
	// then begin to select required pods.
	for _, pod := range allPods {
		if selector.App == "" && selector.Env != "" {
			if pod.Metadata.Label.Env == selector.Env {
				selectedPods = append(selectedPods, pod)
			}
		} else {
			if selector.App != "" && selector.Env == "" {
				if pod.Metadata.Label.App == selector.App {
					selectedPods = append(selectedPods, pod)
				}
			} else {
				if selector.App != "" && selector.Env != "" {
					if pod.Metadata.Label.App == selector.App && pod.Metadata.Label.Env == selector.Env {
						selectedPods = append(selectedPods, pod)
					}
				}
			}
		}
	}
	svc.SvcPods = selectedPods
	return
}

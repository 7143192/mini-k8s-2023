package apiserver

import (
	clientv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v3"
	"log"
	"mini-k8s/pkg/defines"
	"mini-k8s/pkg/etcd"
)

func DeleteGPUJob(cli *clientv3.Client, jobName string) *defines.EtcdGPUJob {
	key := defines.GPUJobPrefix + "/" + jobName
	kv := etcd.Get(cli, key).Kvs
	if len(kv) == 0 {
		log.Printf("job %v not exist in the current system!\n", jobName)
		return nil
	}
	res := &defines.EtcdGPUJob{}
	_ = yaml.Unmarshal(kv[0].Value, res)
	if res.JobState == defines.Pending {
		// DO NOT delete the pending GPUJob.
		return &defines.EtcdGPUJob{}
	}
	// update jobs name list first.
	listKey := defines.GPUJobListPrefix + "/"
	listKV := etcd.Get(cli, listKey).Kvs
	names := make([]string, 0)
	_ = yaml.Unmarshal(listKV[0].Value, &names)
	for idx, name := range names {
		if name == jobName {
			names = append(names[0:idx], names[idx+1:]...)
			break
		}
	}
	nameByte, _ := yaml.Marshal(&names)
	// update list name info.
	etcd.Put(cli, listKey, string(nameByte))
	// then remove this etcdGPUJob object from etcd.
	etcd.Del(cli, key)
	// then remove the related pod.
	podName := res.PodInfo.Metadata.Name
	if podName != "" {
		// name is legal, delete this pod.
		err := DeletePod(cli, podName)
		if err != nil {
			log.Printf("[apiserver] fail to delete pod %v: %v\n", podName, err)
		}
	}
	return res
}

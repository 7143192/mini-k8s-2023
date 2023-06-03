package apiserver

import (
	clientv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v3"
	"mini-k8s/pkg/defines"
	"mini-k8s/pkg/etcd"
)

func GetAutoScaler(cli *clientv3.Client) *defines.GetAutoScalerSend {
	res := &defines.GetAutoScalerSend{}
	scalers := make([]*defines.AutoScalerBriefInfo, 0)
	prefixKey := defines.AutoScalerPrefix + "/"
	kvs := etcd.GetWithPrefix(cli, prefixKey).Kvs
	if len(kvs) == 0 {
		res.AutoScalerBriefs = scalers
		return res
	}
	for _, kv := range kvs {
		tmp := &defines.EtcdAutoScaler{}
		_ = yaml.Unmarshal(kv.Value, tmp)
		info := &defines.AutoScalerBriefInfo{}
		info.AutoName = tmp.AutoName
		info.MinReplicas = tmp.MinReplicas
		info.MaxReplicas = tmp.MaxReplicas
		nameKey := defines.AutoPodReplicasNamesPrefix + "/" + defines.AutoPodNamePrefix + tmp.AutoName
		listKv := etcd.Get(cli, nameKey).Kvs
		names := make([]string, 0)
		_ = yaml.Unmarshal(listKv[0].Value, &names)
		info.CurReplicas = len(names)
		info.Age = tmp.StartTime
		scalers = append(scalers, info)
	}
	res.AutoScalerBriefs = scalers
	return res
}

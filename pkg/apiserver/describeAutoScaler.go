package apiserver

import (
	clientv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v3"
	"mini-k8s/pkg/defines"
	"mini-k8s/pkg/etcd"
)

func DescribeAutoScaler(cli *clientv3.Client, autoName string) *defines.DescribeAutoScalerSend {
	res := &defines.DescribeAutoScalerSend{}
	autoKey := defines.AutoScalerPrefix + "/" + autoName
	kv := etcd.Get(cli, autoKey).Kvs
	if len(kv) == 0 {
		return res
	}
	etcdAuto := &defines.EtcdAutoScaler{}
	_ = yaml.Unmarshal(kv[0].Value, etcdAuto)
	briefInfo := &defines.AutoScalerBriefInfo{}
	briefInfo.AutoName = autoName
	briefInfo.MinReplicas = etcdAuto.MinReplicas
	briefInfo.MaxReplicas = etcdAuto.MaxReplicas
	briefInfo.Age = etcdAuto.StartTime
	namesKey := defines.AutoPodReplicasNamesPrefix + "/" + defines.AutoPodNamePrefix + autoName
	kv = etcd.Get(cli, namesKey).Kvs
	names := make([]string, 0)
	_ = yaml.Unmarshal(kv[0].Value, &names)
	briefInfo.CurReplicas = len(names)
	res.AutoScalerBrief = briefInfo
	res.PodReplicasName = names
	return res
}

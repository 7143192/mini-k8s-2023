package apiserver

import (
	"encoding/json"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v3"
	"log"
	"mini-k8s/pkg/defines"
	"mini-k8s/pkg/etcd"
)

func GetReplicaSets(cli *clientv3.Client) ([]byte, error) {
	rss := make([]*defines.ReplicaSetInfo, 0)
	KVs := etcd.GetWithPrefix(cli, defines.RSInstancePrefix+"/").Kvs
	for _, kv := range KVs {
		rs := &defines.ReplicaSet{}
		err := yaml.Unmarshal(kv.Value, rs)
		if err != nil {
			continue
		}
		rsi := &defines.ReplicaSetInfo{
			Name:      rs.Metadata.Name,
			Replicas:  rs.Spec.Replicas,
			StartTime: rs.StartTime,
		}
		rss = append(rss, rsi)
	}
	rssVal, err := json.Marshal(rss)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		return rssVal, err
	}
	return rssVal, nil
}

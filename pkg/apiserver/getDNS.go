package apiserver

import (
	"encoding/json"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v3"
	"mini-k8s/pkg/defines"
	"mini-k8s/pkg/etcd"
)

func GetDNS(cli *clientv3.Client) string {
	finalRes := &defines.DNSInfoSend{}
	res := make([]*defines.EtcdDNS, 0)
	prefixKey := defines.DNSPrefix + "/"
	kvs := etcd.GetWithPrefix(cli, prefixKey).Kvs
	if len(kvs) == 0 {
		fmt.Println("no dns in the current system for now!")
		resByte, _ := json.Marshal(res)
		return string(resByte)
	}
	for _, kv := range kvs {
		tmp := &defines.EtcdDNS{}
		dnsInfo := &defines.EtcdDNS{}
		_ = yaml.Unmarshal(kv.Value, dnsInfo)
		tmp.DNSName = dnsInfo.DNSName
		tmp.DNSHost = dnsInfo.DNSHost
		tmp.DNSPaths = dnsInfo.DNSPaths
		res = append(res, tmp)
	}
	finalRes.DNSInfos = res
	finalResByte, _ := json.Marshal(finalRes)
	return string(finalResByte)
}

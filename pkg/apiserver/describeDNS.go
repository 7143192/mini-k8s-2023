package apiserver

import (
	"encoding/json"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v3"
	"mini-k8s/pkg/defines"
	"mini-k8s/pkg/etcd"
)

func DescribeDNS(cli *clientv3.Client, dnsName string) string {
	dnsInfo := &defines.EtcdDNS{}
	dnsKey := defines.DNSPrefix + "/" + dnsName
	res := &defines.EtcdDNS{}
	kv := etcd.Get(cli, dnsKey).Kvs
	if len(kv) == 0 {
		return fmt.Sprintf("dns %v does not exist in the system!\n", dnsName)
	} else {
		_ = yaml.Unmarshal(kv[0].Value, dnsInfo)
		res.DNSName = dnsInfo.DNSName
		res.DNSHost = dnsInfo.DNSHost
		res.DNSPaths = dnsInfo.DNSPaths
	}
	resByte, _ := json.Marshal(res)
	return string(resByte)
}

package apiserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v3"
	"log"
	"mini-k8s/pkg/config"
	"mini-k8s/pkg/defines"
	"mini-k8s/pkg/dns"
	"mini-k8s/pkg/etcd"
	"net/http"
)

func CreateDNS(cli *clientv3.Client, dnsYaml *defines.DNSYaml) *defines.EtcdDNS {
	originList := make([]string, 0)
	oldListKey := defines.AllDNSSetPrefix + "/"
	oldKv := etcd.Get(cli, oldListKey).Kvs
	if len(oldKv) != 0 {
		_ = yaml.Unmarshal(oldKv[0].Value, &originList)
		for _, item := range originList {
			if item == (defines.DNSPrefix + "/" + dnsYaml.Metadata.Name) {
				log.Printf("[apiserver] the service %v has already been creted!\n", dnsYaml.Metadata.Name)
				oldSvcKey := item
				oldSvcKv := etcd.Get(cli, oldSvcKey).Kvs
				tmp := &defines.EtcdDNS{}
				_ = yaml.Unmarshal(oldSvcKv[0].Value, tmp)
				// if already been created, return the old one.
				return tmp
			}
		}
	}
	res := &defines.EtcdDNS{}
	res.DNSName = dnsYaml.Metadata.Name
	res.DNSHost = dnsYaml.Spec.Host
	res.DNSPaths = []*defines.EtcdDNSPath{}
	for _, path := range dnsYaml.Spec.Paths {
		svcInfo := &defines.EtcdService{}
		svcKey := defines.ServicePrefix + "/" + path.ServiceName
		kv := etcd.Get(cli, svcKey).Kvs
		if len(kv) == 0 {
			log.Printf("[apiserver] dns create fail, service %v does not exist in the system!\n", path.ServiceName)
			return nil
		} else {
			_ = yaml.Unmarshal(kv[0].Value, svcInfo)
		}
		etcdPath := &defines.EtcdDNSPath{}
		etcdPath.PathAddr = path.PathAddr
		etcdPath.ServiceName = path.ServiceName
		etcdPath.ServiceIp = svcInfo.SvcClusterIP
		etcdPath.Port = path.Port
		res.DNSPaths = append(res.DNSPaths, etcdPath)
	}
	// store the new object into etcd.
	svcKey := defines.DNSPrefix + "/" + res.DNSName
	resByte, _ := yaml.Marshal(res)
	etcd.Put(cli, svcKey, string(resByte))
	svcListKey := defines.AllDNSSetPrefix + "/"
	oldList := make([]string, 0)
	kv := etcd.Get(cli, svcListKey).Kvs
	if len(kv) == 0 {
		newList := append(oldList, svcKey)
		newListByte, _ := yaml.Marshal(&newList)
		etcd.Put(cli, svcListKey, string(newListByte))
	} else {
		_ = yaml.Unmarshal(kv[0].Value, &oldList)
		newList := append(oldList, svcKey)
		newListByte, _ := yaml.Marshal(&newList)
		etcd.Put(cli, svcListKey, string(newListByte))
	}

	//_ = SendOutCreateDNS(cli, res)
	nginxServerIp := dns.StartNewNginxServer(res)
	dns.AddHost(res.DNSHost, nginxServerIp)
	return res
}

func SendOutCreateDNS(cli *clientv3.Client, dns *defines.EtcdDNS) error {
	prefixKey := defines.NodePrefix + "/"
	kvs := etcd.GetWithPrefix(cli, prefixKey).Kvs
	nodes := make([]*defines.NodeInfo, 0)
	if len(kvs) != 0 {
		for _, kv := range kvs {
			oneNode := &defines.NodeInfo{}
			err0 := yaml.Unmarshal(kv.Value, oneNode)
			if err0 != nil {
				fmt.Printf("an error occurs when unmarshal one node in SendOutCreateDNS func: %v\n", err0)
				return nil
			}
			nodes = append(nodes, oneNode)
		}
	}
	for _, nodeInfo := range nodes {
		dnsInfo := dns
		body, err := json.Marshal(dnsInfo)
		if err != nil {
			return err
		}
		nodeIP := nodeInfo.NodeData.NodeSpec.Metadata.Ip
		url := "http://" + nodeIP + ":" + config.ProxyPort + "/objectAPI/createActualDNS"
		fmt.Printf("dns create url = %v\n", url)
		request, err := http.NewRequest("POST", url, bytes.NewReader(body))
		if err != nil {
			return err
		}
		request.Header.Add("Content-Type", "application/json")
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			return err
		}
		if response.StatusCode == http.StatusOK {
			fmt.Println("dns create successfully!")
		} else {
			fmt.Println("dns create fails!")
		}

	}
	return nil
}

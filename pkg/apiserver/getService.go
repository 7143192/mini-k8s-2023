package apiserver

import (
	"encoding/json"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v3"
	"mini-k8s/pkg/defines"
	"mini-k8s/pkg/etcd"
	"mini-k8s/utils"
	"time"
)

func GetService(cli *clientv3.Client) string {
	finalRes := &defines.ServiceInfoSend{}
	res := make([]*defines.ServiceBriefInfo, 0)
	prefixKey := defines.ServicePrefix + "/"
	kvs := etcd.GetWithPrefix(cli, prefixKey).Kvs
	if len(kvs) == 0 {
		fmt.Println("no service in the current system for now!")
		resByte, _ := json.Marshal(res)
		return string(resByte)
	}
	for _, kv := range kvs {
		tmp := &defines.ServiceBriefInfo{}
		svcInfo := &defines.EtcdService{}
		_ = yaml.Unmarshal(kv.Value, svcInfo)
		tmp.SvcName = svcInfo.SvcName
		tmp.SvcType = svcInfo.SvcType
		tmp.SvcPorts = svcInfo.SvcPorts
		tmp.SvcClusterIP = svcInfo.SvcClusterIP
		tmp.SvcExternalIP = "empty"
		tmp.SvcAge = utils.GetTime(time.Now()) - utils.GetTime(svcInfo.SvcStartTime)
		res = append(res, tmp)
	}
	finalRes.SvcInfos = res
	finalResByte, _ := json.Marshal(finalRes)
	return string(finalResByte)
}

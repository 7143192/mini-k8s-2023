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

func DescribeService(cli *clientv3.Client, svcName string) string {
	svcInfo := &defines.EtcdService{}
	svcKey := defines.ServicePrefix + "/" + svcName
	res := &defines.ServiceInfo{}
	kv := etcd.Get(cli, svcKey).Kvs
	if len(kv) == 0 {
		return fmt.Sprintf("service %v does not exist in the system!\n", svcName)
	} else {
		_ = yaml.Unmarshal(kv[0].Value, svcInfo)
		res.SvcBriefInfo = &defines.ServiceBriefInfo{}
		res.SvcBriefInfo.SvcName = svcInfo.SvcName
		res.SvcBriefInfo.SvcType = svcInfo.SvcType
		res.SvcBriefInfo.SvcClusterIP = svcInfo.SvcClusterIP
		res.SvcBriefInfo.SvcExternalIP = "empty"
		res.SvcBriefInfo.SvcPorts = svcInfo.SvcPorts
		res.SvcBriefInfo.SvcAge = utils.GetTime(time.Now()) - utils.GetTime(svcInfo.SvcStartTime)
		res.SvcPods = svcInfo.SvcPods
	}
	resByte, _ := json.Marshal(res)
	return string(resByte)
}

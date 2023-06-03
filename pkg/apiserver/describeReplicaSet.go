package apiserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v3"
	"mini-k8s/pkg/config"
	"mini-k8s/pkg/defines"
	"mini-k8s/pkg/etcd"
	"net/http"
)

func DesReplicaSet(cli *clientv3.Client, rsName string) (*defines.DesRSInfo, error) {
	KVS := etcd.Get(cli, defines.RSInstancePrefix+"/"+rsName).Kvs
	if len(KVS) == 0 {
		return nil, fmt.Errorf("no replicaSet named %s", rsName)
	}
	body, err := json.Marshal(rsName)
	if err != nil {
		return nil, err
	}
	url := "http://" + config.MasterIP + ":" + config.RCPort + "/objectAPI/describeReplicaSet"
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		result := make(map[string]string)
		err := json.NewDecoder(res.Body).Decode(&result)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf(result["ERROR"])
	} else {
		result := make(map[string]*defines.DesRSInfo)
		err := json.NewDecoder(res.Body).Decode(&result)
		if err != nil {
			return nil, err
		}
		rs := &defines.ReplicaSet{}
		err = yaml.Unmarshal(KVS[0].Value, rs)
		if err != nil {
			return nil, err
		}
		result["INFO"].Info.StartTime = rs.StartTime
		return result["INFO"], nil
	}
}

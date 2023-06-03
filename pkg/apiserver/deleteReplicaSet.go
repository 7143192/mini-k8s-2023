package apiserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"mini-k8s/pkg/config"
	"mini-k8s/pkg/defines"
	"mini-k8s/pkg/etcd"
	"net/http"
)

func SendDeletedRSToRC(rsName string) error {
	url := "http://" + config.MasterIP + ":" + config.RCPort + "/objectAPI/delete"
	log.Printf("[apiserver] delete replicaSet request = %v\n", url)
	name, _ := json.Marshal(rsName)
	request, err := http.NewRequest("POST", url, bytes.NewReader(name))
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		result := make(map[string]string)
		err = json.NewDecoder(response.Body).Decode(&result)
		if err != nil {
			return err
		}
		return fmt.Errorf("%s", result["ERROR"])
	}
	return nil
}

func DeleteReplicaSet(cli *clientv3.Client, rsName string) error {
	Kvs := etcd.Get(cli, defines.RSInstancePrefix+"/"+rsName).Kvs
	if len(Kvs) == 0 {
		return fmt.Errorf("the replicaSet %s does not exist", rsName)
	}
	err := SendDeletedRSToRC(rsName)
	if err != nil {
		return err
	}
	etcd.Del(cli, defines.RSInstancePrefix+"/"+rsName)
	return nil
}

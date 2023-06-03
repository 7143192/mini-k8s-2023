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
	"mini-k8s/pkg/etcd"
	"net/http"
	"time"
)

func SendReplicaSetToRC(rs *defines.ReplicaSet) error {
	url := "http://" + config.MasterIP + ":" + config.RCPort + "/objectAPI/create"
	log.Printf("[apiserver] replicaSet request = %v\n", url)
	body, err := json.Marshal(rs)
	if err != nil {
		return err
	}
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("something went wrong in sending replicaSet")
	}
	return nil
}

func CreateReplicaSet(cli *clientv3.Client, rs *defines.ReplicaSet) error {
	key := defines.RSInstancePrefix + "/" + rs.Metadata.Name
	if len(etcd.Get(cli, key).Kvs) != 0 {
		return fmt.Errorf("replicaSet with this name has already been created in the current system")
	}
	err := SendReplicaSetToRC(rs)
	if err != nil {
		return err
	}
	rs.StartTime = time.Now()
	rsVal, err := yaml.Marshal(rs)
	if err != nil {
		return err
	}
	etcd.Put(cli, key, string(rsVal))
	return nil
}

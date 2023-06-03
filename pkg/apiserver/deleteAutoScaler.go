package apiserver

import (
	"bytes"
	"encoding/json"
	"errors"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v3"
	"log"
	"mini-k8s/pkg/config"
	"mini-k8s/pkg/defines"
	"mini-k8s/pkg/etcd"
	"net/http"
)

func SendOutDeleteAutoScalerRequest(etcdAuto *defines.EtcdAutoScaler) error {
	body, err := json.Marshal(etcdAuto)
	if err != nil {
		return err
	}
	url := "http://" + config.MasterIP + ":" + config.AutoScalerPort + "/objectAPI/removeAutoScaler"
	log.Printf("[apiserver] delete autoScaler url = %v\n", url)
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	result := make(map[string]string, 0)
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return errors.New("[ERROR]: " + result["ERROR"])
	}
	return nil
}

func DeleteAutoScaler(cli *clientv3.Client, autoName string) error {
	//config.PodMutex.Lock()
	//fmt.Printf("\ndeleteAutoScaler func holds the mutex lock here!\n\n")

	// delete this name from all-scaler-list first.
	allKey := defines.AutoScalerListPrefix + "/"
	kv := etcd.Get(cli, allKey).Kvs
	oldList := make([]string, 0)
	_ = yaml.Unmarshal(kv[0].Value, &oldList)
	for idx, item := range oldList {
		if item == autoName {
			oldList = append(oldList[0:idx], oldList[idx+1:]...)
			break
		}
	}
	// put new list back to etcd.
	oldListByte, _ := yaml.Marshal(&oldList)
	etcd.Put(cli, allKey, string(oldListByte))
	// then delete this obj itself from etcd.
	autoKey := defines.AutoScalerPrefix + "/" + autoName
	kv = etcd.Get(cli, autoKey).Kvs
	tmp := &defines.EtcdAutoScaler{}
	_ = yaml.Unmarshal(kv[0].Value, tmp)
	etcd.Del(cli, autoKey)
	// then send the deleted info to autoScalerController.
	_ = SendOutDeleteAutoScalerRequest(tmp)
	// then delete the yamlPod instance in etcd.
	yamlKey := defines.AutoScalerYamlPodPrefix + "/" + defines.AutoPodNamePrefix + autoName
	etcd.Del(cli, yamlKey)
	// then delete all pod replicas related to this autoScaler.
	nameList := make([]string, 0)
	namesKey := defines.AutoPodReplicasNamesPrefix + "/" + defines.AutoPodNamePrefix + autoName
	kv = etcd.Get(cli, namesKey).Kvs
	_ = yaml.Unmarshal(kv[0].Value, &nameList)

	// then delete pod replicas list kv in etcd.
	etcd.Del(cli, namesKey)

	// then delete every pod replica object.
	log.Printf("[apiserver] names of replicas that try to delete = %v\n", nameList)

	for _, replicaName := range nameList {
		err := DeletePod(cli, replicaName)
		if err != nil {
			log.Printf("[apiserver] an error occurs when deleting a replica of pod: %v\n", err)
			return err
		}
	}
	//// then delete pod replicas list kv in etcd.
	//etcd.Del(cli, namesKey)
	//fmt.Printf("\ndeleteAutoScaler func gives up the mutex lock here!\n\n")
	//config.PodMutex.Unlock()
	return nil
}

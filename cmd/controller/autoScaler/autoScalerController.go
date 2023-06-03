package main

import (
	"bytes"
	"encoding/json"
	"log"
	"mini-k8s/pkg/apiserver"
	"mini-k8s/pkg/config"
	"mini-k8s/pkg/defines"
	"net/http"
)

func SendOutGetAllAutoScalers(autoServer *apiserver.Server) {
	name := "allControllerScalers"
	body, _ := json.Marshal(&name)
	url := "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/getAutoControllerScalers"
	// log.Printf("[apiserver] get all controlerScalers request = %v\n", url)
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Printf("[apiserver] an error occurs when send out new shcedule request!\n")
		return
	}
	bodyReader := response.Body
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(bodyReader)
	autoServer.AutoScalerController = &defines.AutoScalerController{}
	_ = json.Unmarshal(buf.Bytes(), autoServer.AutoScalerController)
}

func main() {
	//asc := &defines.AutoScalerController{}
	//asc.AutoScalers = make([]*defines.EtcdAutoScaler, 0)
	autoServer := apiserver.AutoScalerServerInit()
	//key := defines.AutoScalerControllerPrefix + "/"
	//cli := etcd.EtcdStart()
	//kv := etcd.Get(cli, key).Kvs
	//if len(kv) == 0 {
	//	autoServer.AutoScalerController = &defines.AutoScalerController{}
	//	autoServer.AutoScalerController.AutoScalers = make([]*defines.EtcdAutoScaler, 0)
	//} else {
	//	autoServer.AutoScalerController = &defines.AutoScalerController{}
	//	_ = yaml.Unmarshal(kv[0].Value, autoServer.AutoScalerController)
	//}
	//// close the etcd client.
	//_ = cli.Close()

	// try to restore autoController states.
	SendOutGetAllAutoScalers(autoServer)

	// autoServer.AutoScalerController = asc
	// start a go routine to check replicas resource usage periodically.
	go autoServer.CheckShrinkAndExtend()
	err := autoServer.AutoScalerServerRun()
	if err != nil {
		log.Printf("[autoScaler] an error occurs when start to run an autoServer: %v\n", err)
		panic(err)
	}
}

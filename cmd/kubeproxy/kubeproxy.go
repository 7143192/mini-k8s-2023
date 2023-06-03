package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mini-k8s/pkg/apiserver"
	"mini-k8s/pkg/config"
	"mini-k8s/pkg/defines"
	"net/http"
)

func SendOutGetAllSvc(s *apiserver.Server) error {
	name := "allServices"
	body, _ := json.Marshal(name)
	url := "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/getAllServices"
	log.Printf("[kubeproxy] kubeproxy gets all services request = %v\n", url)
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Printf("[kubeproxy] an error occurs when send out new shcedule request!\n")
		return err
	}
	if response.StatusCode != http.StatusOK {
		return errors.New("[kubeproxy] error when getting all services info after kube-proxy starts")
	}
	bodyReader := response.Body
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(bodyReader)
	log.Printf("got buf = %v\n", buf.String())
	if buf.String() == "[]" {
		return nil
	}
	tmp := &defines.AllSvc{}
	err = json.Unmarshal(buf.Bytes(), tmp)
	s.Services = tmp.Svc
	// fmt.Printf("svc = %v\n", tmp.Svc)
	for _, s := range s.Services {
		fmt.Printf("[kubeproxy] a svc after init: %v\n", *s)
	}
	if err != nil {
		log.Printf("[kubeproxy] error when getting all services from apiServer:%v\n", err)
	}
	return nil
}

func main() {
	// run a server to receive service changes request from api server.
	kubeProxyServer := apiserver.KubeProxyServerInit()
	// try to get all services.
	err := SendOutGetAllSvc(kubeProxyServer)
	if err != nil {
		panic(err)
	}
	// re-init all chains for the old services that are discarded after the node is restarted.
	kubeProxyServer.ServerReInitAllServicesChains()
	err = kubeProxyServer.KubeProxyServerRun()
	if err != nil {
		panic(err)
	}
}

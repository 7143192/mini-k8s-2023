package pod

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"mini-k8s/pkg/config"
	container2 "mini-k8s/pkg/container"
	"mini-k8s/pkg/defines"
	"net/http"
	"time"
)

func SendOutUpdateRestartInfo(podInfo *defines.Pod) error {
	body, _ := json.Marshal(podInfo)
	url := "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/updatePodRestart"
	log.Printf("schedule request = %v\n", url)
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
		return errors.New("error when handling updating pod restart info")
	}
	return nil
}

func RestartPod(podInfo *defines.Pod) {
	// stop all containers first.
	for _, con := range podInfo.ContainerStates {
		container2.StopContainer(con.Id)
	}
	// then restart all containers.
	for _, con := range podInfo.ContainerStates {
		container2.RestartContainer(con.Id)
	}
	// then update every sub-container running state to running.
	for idx, _ := range podInfo.ContainerStates {
		podInfo.ContainerStates[idx].State = defines.Running
	}
	// then update the restart info of podInfo.
	podInfo.RestartNum += len(podInfo.ContainerStates)
	podInfo.RestartTime = time.Now()
	_ = SendOutUpdateRestartInfo(podInfo)
	return
}

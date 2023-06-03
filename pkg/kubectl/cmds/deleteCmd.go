package cmds

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/urfave/cli/v2"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"mini-k8s/pkg/config"
	"mini-k8s/pkg/defines"
	"mini-k8s/utils/yaml"
	"net/http"
	"time"
)

// format: kubectl del TYPE OBJ_NAME

func DeleteCmd() *cli.Command {
	return &cli.Command{
		Name:  "del",
		Usage: "delete a detailed object of detailed type.",
		Flags: []cli.Flag{},
		Action: func(c *cli.Context) error {
			Delete(c)
			return nil
		},
	}
}

func SendOutDeletePod(delName string) error {
	body, err := json.Marshal(delName)
	if err != nil {
		return err
	}
	url := "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/deletePod"
	// fmt.Printf("delete pod url = %v\n", url)
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
	} else {
		// fmt.Println("[INFO]: " + result["INFO"])
		fmt.Printf("delete pod %v successfully!\n", delName)
	}
	return nil
}

func SendOutDeleteAutoScaler(delName string) error {
	body, err := json.Marshal(delName)
	if err != nil {
		return err
	}
	url := "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/deleteAutoScaler"
	// fmt.Printf("delete auto url = %v\n", url)
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
		return errors.New("error when deleting an autoScaler from system")
	} else {
		fmt.Printf("delete autoScaler %v successfully!\n", delName)
	}
	return nil
}

func SendOutDelReplicaSet(delName string) {
	body, err := json.Marshal(delName)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		return
	}
	url := "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/deleteReplicaSet"
	// fmt.Printf("delete replicaSet url = %v\n", url)
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		return
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		return
	}
	result := make(map[string]string, 0)
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		return
	}
	if response.StatusCode != http.StatusOK {
		log.Println("[ERROR]: " + result["ERROR"])
	} else {
		// log.Println("[INFO]: " + result["INFO"])
		fmt.Printf("delete replicaSet %v successfully!\n", delName)
	}
}

func SendOutDelGPUJob(delName string) {
	body, err := json.Marshal(delName)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		return
	}
	url := "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/deleteGPUJob"
	// fmt.Printf("delete GPUJob url = %v\n", url)
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		return
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		return
	}
	if response.StatusCode != http.StatusOK {
		log.Printf("[kubectl] fail to delete GPUJob %v\n", delName)
		return
	}
	gotData := &defines.EtcdGPUJob{}
	bodyReader := response.Body
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(bodyReader)
	_ = json.Unmarshal(buf.Bytes(), gotData)
	if gotData.JobInfo == nil || gotData.JobInfo.Name == "" {
		fmt.Printf("Fail to delete job, as the job %v is pending!\n", delName)
		return
	}
	fmt.Printf("delete GPUJob %v successfully!\n", delName)
	return
}

func Delete(c *cli.Context) {
	args := c.Args()
	if args.Len() < 2 {
		fmt.Println("Too few arguments! delete type and object name required!")
		return
	}
	if args.Len() > 2 {
		fmt.Println("Too many arguments for del command! Only delete type and object name required!")
		return
	}
	delType := args.Get(0)
	delName := args.Get(1)
	// TODO: this client-v3 should not be created here after http is built in the system.
	// there should be a hanging server waiting for http, and in this server's main func we will create the clientv3 object.
	client, _ := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379", "localhost:22379", "localhost:32379"},
		DialTimeout: 5 * time.Second,
	})
	defer client.Close()
	switch delType {
	case "pod":
		fallthrough
	case "pods":
		// TODO: in version 1.0, no network, so call apiserver API directly.
		if len(delName) > 5 && delName[len(delName)-4:] == "yaml" {
			// input object is a yaml file.
			kind, err := yaml.ParseYamlKind(delName)
			if err != nil {
				fmt.Println("unknown type when parsing yaml in delete API!")
				return
			}
			// TODO: in version 1.0, only consider the POD type case.
			if kind == defines.POD {
				// pod
				pod, err1 := yaml.ParsePodConfig(delName)
				if err1 != nil {
					fmt.Println("unknown type when parsing yaml information in delete API!")
					return
				}
				delName = pod.Metadata.Name
			}
		}
		_ = SendOutDeletePod(delName)
	case "autoScaler":
		fallthrough
	case "autoScalers":
		// TODO: in version 1.0, no network, so call apiserver API directly.
		if len(delName) > 5 && delName[len(delName)-4:] == "yaml" {
			// input object is a yaml file.
			kind, err := yaml.ParseYamlKind(delName)
			if err != nil {
				fmt.Println("unknown type when parsing yaml in delete API!")
				return
			}
			// TODO: in version 1.0, only consider the POD type case.
			if kind == defines.AUTO {
				// pod
				auto, err1 := yaml.ParseAutoScalerConfig(delName)
				if err1 != nil {
					fmt.Println("unknown type when parsing yaml information in delete API!")
					return
				}
				delName = auto.Metadata.AutoName
			}
		}
		_ = SendOutDeleteAutoScaler(delName)
	case "replicaSet":
		SendOutDelReplicaSet(delName)
	case "GPUJob":
		SendOutDelGPUJob(delName)
	default:
		fmt.Println("Required instruction: kubectl del OBJ_TYPE OBJ_NAME/OBJ_YAML_PATH")
	}
}

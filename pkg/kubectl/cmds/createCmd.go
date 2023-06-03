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
	"strings"
	"time"
)

// format: kubectl create --file/-f filePath

func TestInit() *cli.App {
	app := cli.NewApp()
	app.Name = "kubectl"
	app.Version = "1.0"
	app.Usage = "one cmd tool"
	app.Flags = []cli.Flag{}
	app.Commands = []*cli.Command{
		HelloWorld(),
		CreateCmd(),
		DescribeCmd(),
		DeleteCmd(),
		GetCmd(),
	}
	return app
}

func TestParseArgs(app *cli.App, cmdStr string) error {
	parts := strings.Split(cmdStr, " ")
	err := app.Run(parts)
	if err != nil {
		log.Fatal("[Fault] ", err)
	}
	return err
}

func CreateCmd() *cli.Command {
	return &cli.Command{
		Name:  "create",
		Usage: "create a object according to the file passed in",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Value:   "",
				Usage:   "pass a file path into kubectl.",
			},
		},
		Action: func(c *cli.Context) error {
			Create(c)
			return nil
		},
	}
}

func SendOutCreatePod(pod *defines.YamlPod) error {
	body, err := json.Marshal(pod)
	if err != nil {
		return err
	}
	url := "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/createPod"
	// fmt.Printf("create pod url = %v\n", url)
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
		fmt.Println("[INFO]: " + result["INFO"])
	}
	return nil
}

func SendOutCreateClusterIPSvc(cipSvc *defines.ClusterIPService) error {
	body, err := json.Marshal(cipSvc)
	if err != nil {
		return err
	}
	url := "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/createClusterIP"
	// fmt.Printf("create ClusterIP url = %v\n", url)
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	if response.StatusCode == http.StatusOK {
		fmt.Printf("create a new ClusterIP service successfully!\n")
	} else {
		fmt.Println("an error occurs when creating a new ClusterIP service!")
	}
	return nil
}

func SendOutCreateNodePortSvc(npSvc *defines.NodePortService) error {
	body, err := json.Marshal(npSvc)
	if err != nil {
		return err
	}
	url := "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/createNodePort"
	// fmt.Printf("create NodePort url = %v\n", url)
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	if response.StatusCode == http.StatusOK {
		fmt.Printf("create a new NodePort service successfully!\n")
	} else {
		fmt.Println("an error occurs when creating a new NodePort service!")
	}
	return nil
}

func SendOutCreateReplicaSet(replicaSet *defines.ReplicaSet) {
	body, err := json.Marshal(replicaSet)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		return
	}
	url := "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/createReplicaSet"
	// fmt.Printf("create replicaSet url = %v\n", url)
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
		log.Println("[ERROR] " + result["ERROR"])
	} else {
		fmt.Println("[INFO]: " + result["INFO"])
	}
}

func SendOutCreateAutoScaler(auto *defines.AutoScaler) error {
	body, err := json.Marshal(auto)
	if err != nil {
		return err
	}
	url := "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/createAutoScaler"
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "application/json")
	reponse, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	if reponse.StatusCode == http.StatusOK {
		fmt.Println("create a new AutoScaler successfully!")
	} else {
		fmt.Println("an error occurs when creating a new AutoScaler!")
	}
	return nil
}

func SendOutCreateDeployment(deploy *defines.Deployment) error {
	body, err := json.Marshal(deploy)
	if err != nil {
		return err
	}
	url := "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/createDeployment"
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil
	}
	if response.StatusCode != http.StatusOK {
		return errors.New("error when creating a new deployment object")
	}
	return nil
}

func SendOutCreateDNS(dns *defines.DNSYaml) error {
	body, err := json.Marshal(dns)
	if err != nil {
		return err
	}
	url := "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/createDNS"
	// fmt.Printf("create DNS url = %v\n", url)
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
		fmt.Println("[INFO]: " + result["INFO"])
	}
	return nil
}

func SendOutCreateGPUJob(gpuJob *defines.GPUJob) error {
	body, err := json.Marshal(gpuJob)
	if err != nil {
		return err
	}
	url := "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/createGPUJob"
	// fmt.Printf("create GPUJob url = %v\n", url)
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
		return errors.New("error when sending out creating a new gpuJob")
	}
	return nil
}

func Create(c *cli.Context) {
	fileName := c.String("file")
	if fileName == "" {
		fmt.Println("You should pass a fileName as a parameter!")
		fmt.Println(c.App.UsageText)
		return
	}
	if len(fileName) <= 5 {
		fmt.Println("Too short file name!")
		return
	}
	fileType := fileName[len(fileName)-4 : len(fileName)]
	if fileType != "yaml" {
		fmt.Println("Configuration file type should be YAML!")
		return
	}
	Type, err := yaml.ParseYamlKind(fileName)
	if err != nil {
		fmt.Println("Can not recognise the object type!")
		return
	}
	// TODO: this clientv3 should not be created here after http is built in the system.
	// there should be a hanging server waiting for http, and in this server's main func we will create the clientv3 object.
	client, _ := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	defer client.Close()
	if Type == defines.POD {
		pod, err1 := yaml.ParsePodConfig(fileName)
		if err1 != nil {
			fmt.Println("Fail to get configuration of pod that want to create!")
			return
		}
		// some tests during the development.
		// fmt.Printf("new yaml pod info = %v\n", pod)
		// apiserver.CreatePod(client, pod)
		_ = SendOutCreatePod(pod)
		return
	}
	if Type == defines.SERVICE {
		svcKind := yaml.GetServiceKind(fileName)
		switch svcKind {
		case "ClusterIP":
			cipSvc, _ := yaml.ParseClusterIPConfig(fileName)
			//fmt.Printf("ClusterIP info = %v\n", cipSvc)
			//fmt.Printf("ClusterIP ports info = %v\n", *(cipSvc.Spec.Ports[0]))
			_ = SendOutCreateClusterIPSvc(cipSvc)
		case "NodePort":
			npSvc, _ := yaml.ParseNodePortConfig(fileName)
			//fmt.Printf("NodePort info = %v\n", npSvc)
			//fmt.Printf("NodePort ports info = %v\n", *(npSvc.Spec.Ports[0]))
			_ = SendOutCreateNodePortSvc(npSvc)
		}
		return
	}
	if Type == defines.AUTO {
		auto, _ := yaml.ParseAutoScalerConfig(fileName)
		// fmt.Printf("auto info = %v\n", auto)
		_ = SendOutCreateAutoScaler(auto)
		return
	}
	if Type == defines.DEPLOYMENT {
		deployment, _ := yaml.ParseDeploymentConfig(fileName)
		// fmt.Printf("deployment info = %v\n", deployment)
		_ = SendOutCreateDeployment(deployment)
		return
	}
	if Type == defines.REPLICASET {
		replicaSet, _ := yaml.ParseReplicaSetConfig(fileName)
		SendOutCreateReplicaSet(replicaSet)
		return
	}
	if Type == defines.DNS {
		dns, err1 := yaml.ParseDNSConfig(fileName)
		if err1 != nil {
			fmt.Println("Fail to get configuration of DNS that want to create!")
			return
		}
		// some tests during the development.
		// fmt.Printf("new yaml dns info = %v\n", dns)
		_ = SendOutCreateDNS(dns)
		return
	}
	if Type == defines.GPU {
		job, err := yaml.ParseGPUJobConfig(fileName)
		if err != nil {
			fmt.Println("Fail to get configuration of GPUJob that want to create!")
			return
		}
		// fmt.Printf("new job info = %v\n", *job)
		_ = SendOutCreateGPUJob(job)
		return
	}

	// wrong format of the command.
	fmt.Printf("Required instruction: kubectl create --file/-f FILE_PATH\n")
	//if Type == defines.NODE {
	//	nodeInfo, err2 := yaml.ParseNodeConfig(fileName)
	//	if err2 != nil {
	//		fmt.Println("Fail to get configuration of node that want to create!")
	//		return
	//	}
	//	res := node.RegisterNodeToMaster(client, nodeInfo)
	//	fmt.Println(res)
	//	return
	//}

	return
}

package cmds

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/urfave/cli/v2"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"mini-k8s/pkg/config"
	"mini-k8s/pkg/defines"
	"mini-k8s/utils"
	"net/http"
	"strconv"
	"time"
)

func GetCmd() *cli.Command {
	return &cli.Command{
		Name:  "get",
		Usage: "get all objects of one detailed type",
		Flags: []cli.Flag{}, // describe don't need options.
		Action: func(c *cli.Context) error {
			Get(c)
			return nil
		},
	}
}

func SendOutGetPod() error {
	name := "pods"
	body, _ := json.Marshal(&name)
	url := "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/getPod"
	// fmt.Printf("describe pod url = %v\n", url)
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	gotData := &defines.GetPods{}
	bodyReader := response.Body
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(bodyReader)
	_ = json.Unmarshal(buf.Bytes(), gotData)
	// fmt.Printf("got pods data from response = %v\n", *gotData)
	if len(gotData.PodsSend) != 0 {
		fmt.Println("all pods brief information:")
		fmt.Println("NAME\t\tIP\t\t\tREADY\t\t\tSTATUS\t\t\tRESTART\t\t\tAGE")
		for _, data := range gotData.PodsSend {
			if data.CurState == "Running" {
				fmt.Printf("%s\t\t%v\t\t%v\t\t\t%s\t\t\t%v\t\t\t%vs\n", data.Name, data.Ip, data.ReadyNum, data.CurState, data.RestartNum, data.ContinueTime)
			} else {
				fmt.Printf("%s\t\t%v\t\t%v\t\t\t%s\t\t\t%v\t\t\t---\n", data.Name, data.Ip, data.ReadyNum, data.CurState, data.RestartNum)
			}
		}
	} else {
		fmt.Println("No pods in the current system!")
	}
	return nil
}

func SendOutGetNode() error {
	name := "nodes"
	body, _ := json.Marshal(&name)
	url := "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/getNode"
	// fmt.Printf("describe pod url = %v\n", url)
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	gotData := &defines.GetNodesResource{}
	bodyReader := response.Body
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(bodyReader)
	_ = json.Unmarshal(buf.Bytes(), gotData)
	// fmt.Printf("got nodes data from response = %v\n", *gotData)
	if len(gotData.NodesSend) != 0 {
		fmt.Println("all nodes brief information:")
		fmt.Println("NAME\t\tID\t\tLABEL\t\tIP\t\t\tSTATE")
		for _, data := range gotData.NodesSend {
			fmt.Printf("%v\t\t%v\t\t%v\t\t%v\t\t%v\n", data.Name, data.Id, data.Label, data.Ip, data.State)
		}
	} else {
		fmt.Println("No nodes in the current system!")
	}
	return nil
}

func SendOutGetService() error {
	name := "services"
	body, _ := json.Marshal(name)
	url := "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/getService"
	// fmt.Printf("get services url = %v\n", url)
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	gotData := &defines.ServiceInfoSend{}
	bodyReader := response.Body
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(bodyReader)
	_ = json.Unmarshal(buf.Bytes(), &gotData)
	if len(gotData.SvcInfos) == 0 {
		fmt.Println("No service in the current system!")
		return nil
	}
	fmt.Println("all services brief information:")
	fmt.Println("NAME\t\t\tTYPE\t\t\tCLUSTER_IP\t\tEXTERNAL_IP\tPORTS\t\t\tAGE")
	for _, data := range gotData.SvcInfos {
		ports := ""
		// TODO: in version 2.0, only consider the only port pair case for now.
		if len(data.SvcPorts) == 1 {
			portStr := strconv.Itoa(data.SvcPorts[0].Port)
			tarPortStr := strconv.Itoa(data.SvcPorts[0].TargetPort)
			ports = portStr + "," + tarPortStr + "/" + data.SvcPorts[0].Protocol
		} else {

		}
		fmt.Printf("%v\t\t%v\t\t%v\t\t%v\t\t%v\t\t%v\n", data.SvcName, data.SvcType, data.SvcClusterIP, data.SvcExternalIP, ports, data.SvcAge)
	}
	return nil
}

func SendOutGetAutoScalers() error {
	name := "services"
	body, _ := json.Marshal(name)
	url := "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/getAutoScaler"
	// fmt.Printf("get autos url = %v\n", url)
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	gotData := &defines.GetAutoScalerSend{}
	bodyReader := response.Body
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(bodyReader)
	_ = json.Unmarshal(buf.Bytes(), &gotData)
	if len(gotData.AutoScalerBriefs) == 0 {
		fmt.Println("No auto scalers in the current system!")
		return nil
	}
	fmt.Println("all auto scalers brief information:")
	fmt.Println("NAME\t\t\tMIN_REPLICA\t\tMAX_REPLICA\t\tCUR_REPLICA\t\tAGE")
	for _, info := range gotData.AutoScalerBriefs {
		fmt.Printf("%v\t\t\t%v\t\t\t%v\t\t\t%v\t\t\t%v\n", info.AutoName, info.MinReplicas, info.MaxReplicas,
			info.CurReplicas, utils.GetTime(time.Now())-utils.GetTime(info.Age))
	}
	return nil
}

func SendOutGetReplicaSet() {
	url := "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/getReplicaSet"
	request, err := http.NewRequest("POST", url, nil)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		return
	}
	res, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		return
	}
	if res.StatusCode != http.StatusOK {
		result := make(map[string]string)
		err = json.NewDecoder(res.Body).Decode(&result)
		if err != nil {
			log.Printf("[ERROR] %v\n", err)
			return
		}
		log.Printf("[ERROR] %s\n", result["ERROR"])
	} else {
		result := make(map[string][]byte)
		err := json.NewDecoder(res.Body).Decode(&result)
		if err != nil {
			log.Printf("[ERROR] %v\n", err)
			return
		}
		rssSet := make([]*defines.ReplicaSetInfo, 0)
		err = json.Unmarshal(result["INFO"], &rssSet)
		if err != nil {
			log.Printf("[ERROR] %v\n", err)
			return
		}
		fmt.Printf("Name\t\tReplicas\t\tAge\n")
		for _, rss := range rssSet {
			fmt.Printf("%v\t\t%v\t\t%v\n", rss.Name, rss.Replicas, utils.GetTime(time.Now())-utils.GetTime(rss.StartTime))
		}
	}
}

func SendOutGetDNS() error {
	name := "dnss"
	body, _ := json.Marshal(name)
	url := "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/getDNS"
	// fmt.Printf("get dnss url = %v\n", url)
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	gotData := &defines.DNSInfoSend{}
	bodyReader := response.Body
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(bodyReader)
	_ = json.Unmarshal(buf.Bytes(), &gotData)
	if len(gotData.DNSInfos) == 0 {
		fmt.Println("No dns in the current system!")
		return nil
	}
	fmt.Println("all dns brief information:")
	fmt.Println("NAME\t\t\tHOST\t\t\tPATHS")
	for _, data := range gotData.DNSInfos {
		paths := ""
		// TODO: in version 2.0, only consider the only port pair case for now.
		if len(data.DNSPaths) == 1 {
			paths = data.DNSPaths[0].PathAddr + "," + data.DNSPaths[0].ServiceName + ":" + strconv.Itoa(data.DNSPaths[0].Port)
		} else {
			fmt.Printf("length is %v\n", strconv.Itoa(len(data.DNSPaths)))
			for _, path := range data.DNSPaths {
				paths += path.PathAddr + "," + path.ServiceName + ":" + strconv.Itoa(path.Port) + "\t"
			}
		}
		fmt.Printf("%v\t\t%v\t\t%v\t\t\n", data.DNSName, data.DNSHost, paths)
	}
	return nil
}

func SendOutGetGPUJobs() error {
	name := "gpuJobs"
	body, _ := json.Marshal(name)
	url := "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/getAllGPUJobs"
	// fmt.Printf("get gpuJobs url = %v\n", url)
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	gotData := &defines.AllJobs{}
	bodyReader := response.Body
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(bodyReader)
	_ = json.Unmarshal(buf.Bytes(), &gotData)
	if len(gotData.Jobs) == 0 {
		fmt.Println("No gpu jobs in the current system!")
		return nil
	}
	fmt.Println("all gpu jobs information:")
	fmt.Println("NAME\t\t\tSTATE\t\t\tCUDA FUNC")
	for _, data := range gotData.Jobs {
		fmt.Printf("%v\t\t\t%v\t\t\t%v\t\t\n", data.JobInfo.Name, utils.ConvertIntToState(data.JobState), data.JobInfo.ImageName)
	}
	return nil
}

func Get(c *cli.Context) {
	args := c.Args()
	if args.Len() < 1 {
		fmt.Println("Too few arguments! get type is required!")
		return
	}
	if args.Len() > 1 {
		fmt.Println("Too many arguments for get command! Only get type is required!")
		return
	}
	getType := args.Get(0)
	// TODO: this clientv3 should not be created here after http is built in the system.
	// there should be a hanging server waiting for http, and in this server's main func we will create the clientv3 object.
	client, _ := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379", "localhost:22379", "localhost:32379"},
		DialTimeout: 5 * time.Second,
	})
	defer client.Close()
	switch getType {
	case "pod":
		fallthrough
	case "pods":
		// apiserver.GetPod(client)
		_ = SendOutGetPod()
	case "node":
		fallthrough
	case "nodes":
		// apiserver.GetNode(client)
		_ = SendOutGetNode()
	case "service":
		fallthrough
	case "svc":
		fallthrough
	case "services":
		_ = SendOutGetService()
	case "autoScaler":
		fallthrough
	case "auto":
		fallthrough
	case "autoScalers":
		_ = SendOutGetAutoScalers()
	case "replicaSet":
		fallthrough
	case "replicaSets":
		SendOutGetReplicaSet()
	case "DNSs":
		_ = SendOutGetDNS()
	case "GPUJob":
		fallthrough
	case "GPUJobs":
		_ = SendOutGetGPUJobs()
	default:
		fmt.Println("Required instruction: kubectl get OBJ_TYPE(s)")
	}
}

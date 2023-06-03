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
	"mini-k8s/utils/yaml"
	"net/http"
	"strconv"
	"time"
)

// format: kubectl describe TYPE OBJ_NAME

func DescribeCmd() *cli.Command {
	return &cli.Command{
		Name:  "describe",
		Usage: "get a detailed object of one detailed type",
		Flags: []cli.Flag{}, // describe don't need options.
		Action: func(c *cli.Context) error {
			Describe(c)
			return nil
		},
	}
}

func SendOutDescribePod(podName string) error {
	body, err := json.Marshal(podName)
	if err != nil {
		return err
	}
	url := "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/describePod"
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
	// handle response body here.
	gotData := &defines.PodResourceSend{}
	bodyReader := response.Body
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(bodyReader)
	_ = json.Unmarshal(buf.Bytes(), gotData)
	// fmt.Printf("got pod resource data from response = %v\n", *gotData)
	if gotData.PodName == "" {
		fmt.Printf("No information about pod %v! \n", podName)
		return nil
	}
	per := 0.0
	if gotData.LimCpu == 0.0 {
		per = 0.0
	} else {
		per = float64(gotData.CpuUsed) * 100.0 / float64(gotData.LimCpu)
	}
	fmt.Println("Pod Information:")
	// fmt.Println("NAME\t\tPOD IP\t\t\tMEM REQ\t\tMEM LIMIT\tCPU REQ\t\tCPU LIMIT\tRESTARTS\t\tAGE")
	// fmt.Println("NAME\t\tPOD IP\t\t\tMEM REQ\t\tMEM LIMIT\tCPU PER\t\tRESTARTS\t\tAGE")
	if gotData.ContinueTime == 0 {
		//fmt.Printf("%v\t\t%v\t\t---\t\t%v\t\t---\t\t%v\t\t%v\t\t---s\n", gotData.PodName, gotData.PodIP,
		//	gotData.LimMem, gotData.LimCpu, gotData.RestartNum)
		fmt.Printf("name: %v\npodIp: %v\nmemRequest: ---\nmemLimit: %v\nCPUPercent: %v%%\nrestartNum: %v\nage: ---s\n",
			gotData.PodName, gotData.PodIP, gotData.LimMem, per, gotData.RestartNum)
	} else {
		//fmt.Printf("%v\t\t%v\t\t%v\t\t%v\t\t%v\t\t%v\t\t%v\t\t\t%vs\n", gotData.PodName, gotData.PodIP, gotData.MemUsed, gotData.LimMem,
		//	gotData.CpuUsed, gotData.LimCpu, gotData.RestartNum, gotData.ContinueTime)
		fmt.Printf("name: %v\npodIp: %v\nmemRequest: %v\nmemLimit: %v\nCPUPercent: %v%%\nrestartNum: %v\nage: %vs\n",
			gotData.PodName, gotData.PodIP, gotData.MemUsed, gotData.LimMem, per, gotData.RestartNum, gotData.ContinueTime)
	}
	if len(gotData.ContainerResourcesSend) == 0 {
		fmt.Printf("No containers info for pod %v\n", gotData.PodName)
		return nil
	}
	// fmt.Println("Containers Information:")
	// fmt.Println("NAME\t\t\tREADY\t\t\tSTATUS\t\t\tTIME")
	for _, resource := range gotData.ContainerResourcesSend {
		conName := resource.ConName
		readyNum := resource.ReadyNum
		curState := resource.CurState
		totalTime := resource.TotalTime
		if resource.CurState == "Running" {
			fmt.Printf("- name: %s\n\tready: %d/1\n\tstatus: %s\n\ttime: %vs\n", conName, readyNum, curState, totalTime)
		} else {
			fmt.Printf("- name: %s\n\tready: %d/1\n\tstatus: %s\n\ttime: ---s\n", conName, readyNum, curState)
		}
	}
	return nil
}

func SendOutDescribeNode(nodeName string) error {
	body, err := json.Marshal(nodeName)
	if err != nil {
		return err
	}
	url := "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/describeNode"
	// fmt.Printf("describe node url = %v\n", url)
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	gotData := &defines.NodeResourceSend{}
	bodyReader := response.Body
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(bodyReader)
	_ = json.Unmarshal(buf.Bytes(), gotData)
	// fmt.Printf("got node resource data from response = %v\n", *gotData)
	if gotData.NodeName == "" {
		fmt.Printf("No information for node %v!\n", nodeName)
		return nil
	}
	// per := gotData.CpuUsed * 100 / gotData.CpuTotal
	per := 0.0
	if gotData.CpuTotal == 0.0 {
		per = 0.0
	} else {
		per = gotData.CpuUsed * 100.0 / gotData.CpuTotal
	}
	fmt.Println("Node Information:")
	// fmt.Printf("NAME\t\t\tID\t\t\tREADY\t\t\tMEMORY\t\t\tCPU\t\t\tIP\n")
	//fmt.Printf("%v\t\t\t%v\t\t\t%v/1\t\t\t%v/%v\t\t\t%v/%v\t\t\t%v\n", nodeName, gotData.NodeId, gotData.ReadyNum, gotData.MemUsed,
	//	gotData.MemTotal, gotData.CpuUsed, gotData.CpuTotal, gotData.NodeIP)
	fmt.Printf("name: %v\nid: %v\nready: %v/1\nmemory: %v/%v\nCPUPercent: %v%%\nip: %v\n",
		nodeName, gotData.NodeId, gotData.ReadyNum, gotData.MemUsed, gotData.MemTotal, per, gotData.NodeIP)
	if len(gotData.PodsResourcesSend) == 0 {
		return nil
	}
	// fmt.Println("Pods Information:")
	//fmt.Println("NAME\t\tPOD IP\t\t\tMEM REQ\t\tMEM LIMIT\tCPU REQ\t\tCPU LIMIT\tAGE")
	// fmt.Println("NAME\t\tPOD IP\t\t\tMEM REQ\t\tMEM LIMIT\tCPU PER\t\tAGE")
	for _, Data := range gotData.PodsResourcesSend {
		podPer := 0.0
		if Data.LimCpu == 0.0 {
			podPer = 0.0
		} else {
			podPer = float64(Data.CpuUsed) * 100.0 / Data.LimCpu
		}
		if Data.ContinueTime == 0 {
			//fmt.Printf("%v\t\t%v\t\t---\t\t%v\t\t---\t\t%v\t\t---s\n", Data.PodName, Data.PodIP, Data.LimMem, Data.LimCpu)
			fmt.Printf("- name: %v\n\tpodIp: %v\n\tmemRequest: ---\n\tmemLimit: %v\n\tCPUPercent: ---%%\n\tage: ---s\n",
				Data.PodName, Data.PodIP, Data.LimMem)
		} else {
			//fmt.Printf("%v\t\t%v\t\t%v\t\t%v\t\t%v\t\t%v\t\t%vs\n", Data.PodName, Data.PodIP, Data.MemUsed, Data.LimMem,
			//	Data.CpuUsed, Data.LimCpu, Data.ContinueTime)
			fmt.Printf("- name: %v\n\tpodIp: %v\n\tmemRequest: %v\n\tmemLimit: %v\n\tCPUPercent: %v%%\n\tage: %vs\n",
				Data.PodName, Data.PodIP, Data.MemUsed, Data.LimMem, podPer, Data.ContinueTime)
		}
	}
	return nil
}

func SendOutDescribeService(svcName string) error {
	body, err := json.Marshal(svcName)
	if err != nil {
		return err
	}
	url := "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/describeService"
	// fmt.Printf("describe service url = %v\n", url)
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	gotData := &defines.ServiceInfo{}
	bodyReader := response.Body
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(bodyReader)
	_ = json.Unmarshal(buf.Bytes(), gotData)
	if gotData == nil || gotData.SvcBriefInfo == nil || gotData.SvcBriefInfo.SvcName == "" {
		fmt.Printf("No information for service %v!\n", svcName)
		return nil
	}
	fmt.Println("Service information:")
	// fmt.Println("NAME\t\t\tTYPE\t\t\tCLUSTER_IP\t\tEXTERNAL_IP\tPORTS\t\t\tAGE")
	data := gotData.SvcBriefInfo
	ports := ""
	// TODO: in version 2.0, only consider the only port pair case for now.
	if len(data.SvcPorts) == 1 {
		portStr := strconv.Itoa(data.SvcPorts[0].Port)
		tarPortStr := strconv.Itoa(data.SvcPorts[0].TargetPort)
		ports = portStr + "," + tarPortStr + "/" + data.SvcPorts[0].Protocol
	} else {

	}
	fmt.Printf("name: %v\ntype: %v\nclusterIP: %v\nexternalIP: %v\nports: %v\nage: %v\n",
		data.SvcName, data.SvcType, data.SvcClusterIP, data.SvcExternalIP, ports, data.SvcAge)
	if len(gotData.SvcPods) == 0 {
		// fmt.Printf("No pods in service %v for now.\n", svcName)
		return nil
	}
	fmt.Printf("podsName: ")
	podNames := ""
	for _, pod := range gotData.SvcPods {
		podNames = podNames + pod.Metadata.Name + "  "
	}
	fmt.Printf(podNames + "\n")
	return nil
}

func SendOutDescribeAutoScaler(autoName string) error {
	body, err := json.Marshal(autoName)
	if err != nil {
		return err
	}
	url := "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/describeAutoScaler"
	// fmt.Printf("describe autoScaler url = %v\n", url)
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	gotData := &defines.DescribeAutoScalerSend{}
	bodyReader := response.Body
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(bodyReader)
	_ = json.Unmarshal(buf.Bytes(), gotData)
	if gotData == nil || gotData.AutoScalerBrief == nil || gotData.AutoScalerBrief.AutoName == "" {
		fmt.Printf("No information for autoScaler %v!\n", autoName)
		return nil
	}
	fmt.Println("Detailed autoScaler information:")
	// fmt.Println("NAME\t\t\tMIN_REPLICA\t\tMAX_REPLICA\t\tCUR_REPLICA\t\tAGE")
	data := gotData.AutoScalerBrief
	fmt.Printf("name: %v\nminReplicas: %v\nmaxReplicas: %v\ncurReplicas: %v\nage: %v\n",
		data.AutoName, data.MinReplicas, data.MaxReplicas, data.CurReplicas, utils.GetTime(time.Now())-utils.GetTime(data.Age))
	if len(gotData.PodReplicasName) == 0 {
		fmt.Printf("No pod replicas in autoScaler %v for now.\n", autoName)
		return nil
	}
	fmt.Printf("podsName: ")
	podNames := ""
	for _, pod := range gotData.PodReplicasName {
		podNames = podNames + pod + "  "
	}
	fmt.Printf(podNames + "\n")
	return nil
}

func SendOutDesReplicaSet(rsName string) {
	body, err := json.Marshal(rsName)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		return
	}
	url := "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/describeReplicaSet"
	// fmt.Printf("describe replicaSet url = %v\n", url)
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		return
	}
	request.Header.Add("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		return
	}
	if res.StatusCode != http.StatusOK {
		result := make(map[string]string)
		err := json.NewDecoder(res.Body).Decode(&result)
		if err != nil {
			log.Printf("[ERROR] %v\n", err)
			return
		}
		log.Printf("[ERROR] %s\n", result["ERROR"])
	} else {
		result := make(map[string]*defines.DesRSInfo)
		err := json.NewDecoder(res.Body).Decode(&result)
		if err != nil {
			log.Printf("[ERROR] %v\n", err)
			return
		}
		rsi := result["INFO"]
		fmt.Printf("replicaSet-name: %s\n", rsi.Info.Name)
		fmt.Printf("replicas-number: %d/%d\n", len(rsi.PodName), rsi.Info.Replicas)
		fmt.Printf("replicaSet-age: %v\n", utils.GetTime(time.Now())-utils.GetTime(rsi.Info.StartTime))
		fmt.Printf("replicaSet-podName: ")
		for _, name := range rsi.PodName {
			fmt.Printf("%s ", name)
		}
		fmt.Println()
	}
}

func SendOutDescribeDNS(dnsName string) error {
	body, err := json.Marshal(dnsName)
	if err != nil {
		return err
	}
	url := "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/describeDNS"
	// fmt.Printf("describe dns url = %v\n", url)
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	gotData := &defines.EtcdDNS{}
	bodyReader := response.Body
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(bodyReader)
	_ = json.Unmarshal(buf.Bytes(), gotData)
	if gotData == nil || gotData.DNSName == "" {
		fmt.Printf("No information for dns %v!\n", dnsName)
		return nil
	}
	fmt.Println("DNS information:")
	// fmt.Println("NAME\t\t\tHOST\t\t\tPATHS")
	data := gotData
	paths := ""
	// TODO: in version 2.0, only consider the only port pair case for now.
	if len(data.DNSPaths) == 1 {
		paths = data.DNSPaths[0].PathAddr + "," + data.DNSPaths[0].ServiceName + ":" + strconv.Itoa(data.DNSPaths[0].Port)
	} else {
		for _, path := range data.DNSPaths {
			paths += path.PathAddr + "," + path.ServiceName + ":" + strconv.Itoa(path.Port) + "\t"
		}
	}
	fmt.Printf("name: %v\nhost: %v\npaths: %v\n", data.DNSName, data.DNSHost, paths)
	return nil
}

func SendOutDescribeGPUJob(jobName string) error {
	body, err := json.Marshal(&jobName)
	if err != nil {
		return err
	}
	url := "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/describeGPUJob"
	// fmt.Printf("describe gpuJob url = %v\n", url)
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	gotData := &defines.EtcdGPUJob{}
	bodyReader := response.Body
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(bodyReader)
	_ = json.Unmarshal(buf.Bytes(), gotData)
	if gotData == nil || gotData.JobInfo.Name == "" {
		fmt.Printf("No information for gpuJob %v!\n", gotData.JobInfo.Name)
		return nil
	}
	fmt.Println("GPUJob information:")
	// fmt.Println("NAME\t\t\tSTATE\t\t\tCUDA FUNC")
	fmt.Printf("name: %v\nstate: %v\ncudaFunc: %v\n", gotData.JobInfo.Name, utils.ConvertIntToState(gotData.JobState), gotData.JobInfo.ImageName)
	return nil
}

func Describe(c *cli.Context) {
	args := c.Args()
	if args.Len() < 2 {
		fmt.Println("Too few arguments! describe type and object name required!")
		return
	}
	if args.Len() > 2 {
		fmt.Println("Too many arguments for describe command! Only delete type and object name required!")
		return
	}
	describeType := args.Get(0)
	describeName := args.Get(1)
	// TODO: this clientv3 should not be created here after http is built in the system.
	// there should be a hanging server waiting for http, and in this server's main func we will create the clientv3 object.
	client, _ := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379", "localhost:22379", "localhost:32379"},
		DialTimeout: 5 * time.Second,
	})
	defer client.Close()
	switch describeType {
	case "pod":
		fallthrough
	case "pods":
		if len(describeName) > 5 && describeName[len(describeName)-4:] == "yaml" {
			// input object is a yaml file.
			kind, err := yaml.ParseYamlKind(describeName)
			if err != nil {
				fmt.Println("unknown type when parsing yaml in describe API!")
				return
			}
			if kind == defines.POD {
				// pod
				pod, err1 := yaml.ParsePodConfig(describeName)
				if err1 != nil {
					fmt.Println("unknown type when parsing yaml information in describe API!")
					return
				}
				describeName = pod.Metadata.Name
			}
		}
		// apiserver.DescribePod(client, describeName, false)
		// in version 2.0, change here to send http request and do not call api server function directly.
		_ = SendOutDescribePod(describeName)
	case "node":
		fallthrough
	case "nodes":
		if len(describeName) > 5 && describeName[len(describeName)-4:] == "yaml" {
			// input object is a yaml file.
			kind, err := yaml.ParseYamlKind(describeName)
			if err != nil {
				fmt.Println("unknown type when parsing yaml in describe API!")
				return
			}
			if kind == defines.NODE {
				// pod
				nodeData, err1 := yaml.ParseNodeConfig(describeName)
				if err1 != nil {
					fmt.Println("unknown type when parsing yaml information in describe API!")
					return
				}
				describeName = nodeData.Metadata.Name
			}
		}
		// apiserver.DescribeNode(client, describeName)
		// in version 2.0, change here to send http request and do not call api server function directly.
		_ = SendOutDescribeNode(describeName)
	case "service":
		fallthrough
	case "svc":
		fallthrough
	case "services":
		if len(describeName) > 5 && describeName[len(describeName)-4:] == "yaml" {
			// input object is a yaml file.
			kind, err := yaml.ParseYamlKind(describeName)
			if err != nil {
				fmt.Println("unknown type when parsing yaml in describe API!")
				return
			}
			if kind == defines.SERVICE {
				// pod
				svcData, err1 := yaml.ParseNodeConfig(describeName)
				if err1 != nil {
					fmt.Println("unknown type when parsing yaml information in describe API!")
					return
				}
				describeName = svcData.Metadata.Name
			}
		}
		// apiserver.DescribeNode(client, describeName)
		// in version 2.0, change here to send http request and do not call api server function directly.
		_ = SendOutDescribeService(describeName)
	case "autoScaler":
		fallthrough
	case "auto":
		fallthrough
	case "autoScalers":
		if len(describeName) > 5 && describeName[len(describeName)-4:] == "yaml" {
			// input object is a yaml file.
			kind, err := yaml.ParseYamlKind(describeName)
			if err != nil {
				fmt.Println("unknown type when parsing yaml in describe API!")
				return
			}
			if kind == defines.AUTO {
				// pod
				Data, err1 := yaml.ParseAutoScalerConfig(describeName)
				if err1 != nil {
					fmt.Println("unknown type when parsing yaml information in describe API!")
					return
				}
				describeName = Data.Metadata.AutoName
			}
		}
		_ = SendOutDescribeAutoScaler(describeName)
	case "replicaSet":
		SendOutDesReplicaSet(describeName)
	case "DNS":
		fallthrough
	case "DNSs":
		_ = SendOutDescribeDNS(describeName)
	case "GPUJob":
		fallthrough
	case "GPUJobs":
		_ = SendOutDescribeGPUJob(describeName)
	default:
		fmt.Println("Required instruction: kubectl describe OBJ_TYPE OBJ_NAME")
	}
}

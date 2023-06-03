package yaml

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"mini-k8s/pkg/defines"
	"os"
)

func ParseYamlKind(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("fail to read file: ", err)
		return -1, err
	}
	start := defines.YamlStart{}
	err = yaml.Unmarshal(data, &start)
	if err != nil {
		fmt.Println("fail to parse yaml(1): ", err)
		return -1, err
	}
	// in version 1.0, only consider the pod type.
	switch start.Kind {
	case "Pod":
		return defines.POD, nil
	case "Node":
		return defines.NODE, nil
	case "Service":
		return defines.SERVICE, nil
	case "HorizontalPodAutoscaler":
		return defines.AUTO, nil
	case "Deployment":
		return defines.DEPLOYMENT, nil
	case "ReplicaSet":
		return defines.REPLICASET, nil
	case "DNS":
		return defines.DNS, nil
	case "GPUJob":
		return defines.GPU, nil
	}
	return 0, nil
}

func ParsePodConfig(path string) (*defines.YamlPod, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("fail to read file: ", err)
		return nil, err
	}
	pod := defines.YamlPod{}
	err = yaml.Unmarshal(data, &pod)
	if err != nil {
		fmt.Println("fail to parse yaml(1): ", err)
		return nil, err
	}
	// in version 1.0, only consider the pod type.
	return &pod, nil
}

func ParseNodeConfig(path string) (*defines.NodeYaml, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("fail to read file: ", err)
		return nil, err
	}
	node := defines.NodeYaml{}
	err = yaml.Unmarshal(data, &node)
	if err != nil {
		fmt.Println("fail to parse yaml(1): ", err)
		return nil, err
	}
	// in version 1.0, only consider the pod type.
	return &node, nil
}

func GetServiceKind(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("fail to read file: ", err)
		return ""
	}
	nodePortService := &defines.NodePortService{}
	err = yaml.Unmarshal(data, nodePortService)
	if nodePortService.Spec.Type == "ClusterIP" {
		return "ClusterIP"
	}
	if nodePortService.Spec.Type == "NodePort" {
		return "NodePort"
	}
	if nodePortService.Spec.Type == "LoadBalance" {
		return "LoadBalance"
	}
	return ""
}

func ParseClusterIPConfig(path string) (*defines.ClusterIPService, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("fail to read file: ", err)
		return nil, err
	}
	clusterIP := defines.ClusterIPService{}
	err = yaml.Unmarshal(data, &clusterIP)
	return &clusterIP, err
}

func ParseNodePortConfig(path string) (*defines.NodePortService, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("fail to read file: ", err)
		return nil, err
	}
	nodePort := defines.NodePortService{}
	err = yaml.Unmarshal(data, &nodePort)
	return &nodePort, err
}

func ParseAutoScalerConfig(path string) (*defines.AutoScaler, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("fail to read file: ", err)
		return nil, err
	}
	res := &defines.AutoScaler{}
	err = yaml.Unmarshal(data, res)
	return res, nil
}

func ParseDeploymentConfig(path string) (*defines.Deployment, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("fail to read file: ", err)
		return nil, err
	}
	res := &defines.Deployment{}
	err = yaml.Unmarshal(data, res)
	return res, nil
}

func ParseReplicaSetConfig(path string) (*defines.ReplicaSet, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("fail to read file: ", err)
		return nil, err
	}
	replicaSet := defines.ReplicaSet{}
	err = yaml.Unmarshal(data, &replicaSet)
	if err != nil {
		log.Println("fail to parse yaml(1): ", err)
		return nil, err
	}
	return &replicaSet, nil
}

func ParseDNSConfig(path string) (*defines.DNSYaml, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("fail to read file: ", err)
		return nil, err
	}
	dns := defines.DNSYaml{}
	err = yaml.Unmarshal(data, &dns)
	if err != nil {
		fmt.Println("fail to parse yaml(1): ", err)
		return nil, err
	}
	return &dns, nil
}

func ParseGPUJobConfig(path string) (*defines.GPUJob, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("fail to read file: %v\n", err)
		return nil, err
	}
	job := &defines.GPUJob{}
	err = yaml.Unmarshal(data, job)
	if err != nil {
		log.Printf("failt to parse yaml for GPUJob: %v\n", err)
		return nil, err
	}
	return job, nil
}

//func main() {
//	got, err := ParseYamlKind("../templates/pod_template.yaml")
//	if err != nil {
//		return
//	}
//	fmt.Println(got)
//	if got == 1 {
//		pod, err := ParsePodConfig("../templates/pod_template.yaml")
//		if err != nil {
//			return
//		}
//		fmt.Println(pod)
//		fmt.Println(pod.Spec.Containers)
//		fmt.Println(pod.Spec.Volumes)
//	}
//}

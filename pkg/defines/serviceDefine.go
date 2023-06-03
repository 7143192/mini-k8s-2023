package defines

import (
	"time"
)

const ServicePrefix = "Service"

// AllServiceSetPrefix :the watch target about service changes
const AllServiceSetPrefix = "ServiceList"

type ServiceMetadata struct {
	Name string `json:"name" yaml:"name"`
}

// ServiceSelector used to select pods according target labels and pod labels.
type ServiceSelector struct {
	App string `json:"app" yaml:"app"`
	Env string `json:"env" yaml:"env"`
}

type ServicePort struct {
	Port       int    `json:"port" yaml:"port"`
	TargetPort int    `json:"targetPort" yaml:"targetPort"`
	Protocol   string `json:"protocol" yaml:"protocol"`
	NodePort   int    `json:"nodePort" yaml:"nodePort"`
}

type ClusterIPServiceSpec struct {
	Type      string          `json:"type" yaml:"type"`
	ClusterIP string          `json:"clusterIP" yaml:"clusterIP"`
	Ports     []*ServicePort  `json:"ports" yaml:"ports"`
	Selector  ServiceSelector `json:"selector" yaml:"selector"`
}

type ClusterIPService struct {
	ApiVersion string               `json:"apiVersion" yaml:"apiVersion"`
	Kind       string               `json:"kind" yaml:"kind"`
	Metadata   ServiceMetadata      `json:"metadata" yaml:"metadata"`
	Spec       ClusterIPServiceSpec `json:"spec" yaml:"spec"`
}

type NodePortServiceSpec struct {
	Type      string          `json:"type" yaml:"type"`
	ClusterIP string          `json:"clusterIP" yaml:"clusterIP"`
	Ports     []*ServicePort  `json:"ports" yaml:"ports"`
	Selector  ServiceSelector `json:"selector" yaml:"selector"`
}

type NodePortService struct {
	ApiVersion string              `json:"apiVersion" yaml:"apiVersion"`
	Kind       string              `json:"kind" yaml:"kind"`
	Metadata   ServiceMetadata     `json:"metadata" yaml:"metadata"`
	Spec       NodePortServiceSpec `json:"spec" yaml:"spec"`
}

// EtcdService the service struct used to store one service into ETCD.
type EtcdService struct {
	SvcName      string          `json:"svcName" yaml:"svcName"`
	SvcType      string          `json:"svcType" yaml:"svcType"`
	SvcClusterIP string          `json:"svcClusterIP" yaml:"svcClusterIP"`
	SvcSelector  ServiceSelector `json:"svcSelector" yaml:"svcSelector"`
	SvcStartTime time.Time       `json:"svcStartTime" yaml:"svcStartTime"`
	SvcPods      []*Pod          `json:"svcPods" yaml:"svcPods"`
	SvcNodePort  int             `json:"svcNodePort" yaml:"svcNodePort"`
	SvcPorts     []*ServicePort  `json:"svcPorts" yaml:"svcPorts"`
}

// ServiceBriefInfo used in GetService API.
type ServiceBriefInfo struct {
	SvcName       string         `json:"svcName"`
	SvcType       string         `json:"svcType"`
	SvcClusterIP  string         `json:"svcClusterIP"`
	SvcExternalIP string         `json:"svcExternalIP"`
	SvcPorts      []*ServicePort `json:"svcPorts"`
	SvcAge        uint64         `json:"svcAge"`
}

type ServiceInfoSend struct {
	SvcInfos []*ServiceBriefInfo `json:"svcInfos"`
}

// ServiceInfo used in DescribeService API.
type ServiceInfo struct {
	SvcBriefInfo *ServiceBriefInfo `json:"svcBriefInfo"`
	SvcPods      []*Pod            `json:"svcPods"`
}

type AllSvc struct {
	Svc []*EtcdService `json:"svc"`
}

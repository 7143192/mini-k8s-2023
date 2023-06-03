package defines

import (
	"time"
)

// NOTE: we choose CPU and MEMORY as the watching targets.

const AutoScalerPrefix = "AutoScaler"
const AutoScalerListPrefix = "AutoScalerList"
const AutoPodNamePrefix = "AutoPod_"
const AutoScalerYamlPodPrefix = "AutoYamlPod"
const AutoPodReplicasNamesPrefix = "AutoPodList"

type AutoScalerMetadata struct {
	AutoName string `yaml:"name" json:"name"`
}

type AutoScalerCPU struct {
	MinCPU uint64 `yaml:"minCPU" json:"minCPU"`
	MaxCPU uint64 `yaml:"maxCPU" json:"maxCPU"`
}

type AutoScalerMemory struct {
	MinMem string `yaml:"minMem" json:"minMem"`
	MaxMem string `yaml:"maxMem" json:"maxMem"`
}

type AutoScalerMetrics struct {
	//LimitCPU AutoScalerCPU    `yaml:"limitCPU" json:"limitCPU"`
	//LimitMem AutoScalerMemory `yaml:"limitMem" json:"limitMem"`
	Resources []AutoScalerResource `json:"resource" yaml:"resource"`
}

type AutoScalerLabel struct {
	App string `yaml:"app" json:"app"`
	Env string `yaml:"env" json:"env"`
}

type AutoScalerResource struct {
	Name string `json:"name" yaml:"name"`
	Min  string `json:"min" yaml:"min"`
	Max  string `json:"max" yaml:"max"`
}

type AutoScaler struct {
	ApiVersion string `yaml:"apiVersion" json:"apiVersion"`
	// should be "HorizontalPodAutoscaler"
	Kind     string             `yaml:"kind" json:"kind"`
	Metadata AutoScalerMetadata `yaml:"metadata" json:"metadata"`
	// TODO: in version 2.0, only try to support the auto scale for POD type.
	TargetKind  string            `yaml:"targetKind" json:"targetKind"`
	MinReplicas int               `yaml:"minReplicas" json:"minReplicas"`
	MaxReplicas int               `yaml:"maxReplicas" json:"maxReplicas"`
	Metrics     AutoScalerMetrics `yaml:"metrics" json:"metrics"`
	Label       AutoScalerLabel   `yaml:"label" json:"label"`
	// TODO: in version 2.0, just use the pod defines to spec this auto scaler.
	AutoSpec PodSpec `yaml:"spec" json:"spec"`
}

type EtcdAutoScaler struct {
	AutoName    string    `yaml:"autoName" json:"autoName"`
	AutoPodId   string    `yaml:"autoPodId" json:"autoPodId"`
	MinReplicas int       `yaml:"minReplicas" json:"minReplicas"`
	MaxReplicas int       `yaml:"maxReplicas" json:"maxReplicas"`
	MinCPU      uint64    `yaml:"minCPU" json:"minCPU"`
	MaxCPU      uint64    `yaml:"maxCPU" json:"maxCPU"`
	MinMem      string    `yaml:"minMem" json:"minMem"`
	MaxMem      string    `yaml:"maxMem" json:"maxMem"`
	StartTime   time.Time `yaml:"startTime" json:"startTime"`
}

type AutoNameSend struct {
	PodName  string `json:"podName"`
	AutoName string `json:"autoName"`
}

type AutoCreateReplicaSend struct {
	PodName  string   `json:"autoName"`
	YamlInfo *YamlPod `json:"yamlInfo"`
}

type AutoScalerBriefInfo struct {
	AutoName    string    `json:"autoName"`
	MinReplicas int       `json:"minReplicas"`
	MaxReplicas int       `json:"maxReplicas"`
	CurReplicas int       `json:"CurReplicas"`
	Age         time.Time `json:"age"`
}

type GetAutoScalerSend struct {
	AutoScalerBriefs []*AutoScalerBriefInfo `json:"autoScalerBriefs"`
}

type DescribeAutoScalerSend struct {
	AutoScalerBrief *AutoScalerBriefInfo `json:"AutoScalerBrief"`
	PodReplicasName []string             `json:"podReplicasName"`
}

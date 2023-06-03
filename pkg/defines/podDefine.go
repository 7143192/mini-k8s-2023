package defines

import (
	"sync"
	"time"
)

const PodPrefix = "Pod"
const OLdPodPrefix = "OldPod"
const PodInstancePrefix = "PodInstance"
const OldPodInstancePrefix = "OldPodInstance"
const PodIdSetPrefix = "allID"
const OldPodIdSetPrefix = "oldAllID"
const PodResourceStatePrefix = "PodResource"
const PodContainersResourcePrefix = "PodContainersResource"
const PodNsPathPrefix = "/var/run/netns/"
const PodNetworkIDPrefix = "network-"

// some definition used in the configuration of one pod with .yaml file.
// TODO: Version 1.0 by lyh.

type PodLabels struct {
	App string `json:"app" yaml:"app"`
	Env string `json:"env" yaml:"env"`
}

type PodMetadata struct {
	Name  string    `json:"name" yaml:"name"`
	Label PodLabels `json:"labels" yaml:"labels"`
}

type PodNodeSelector struct {
	Gpu string `json:"gpu" yaml:"gpu"`
}

type PodVolumeMount struct {
	Name      string `json:"name" yaml:"name"`
	MountPath string `json:"mountPath" yaml:"mountPath"`
}

type PodPort struct {
	Name          string `json:"name" yaml:"name"`
	ContainerPort int    `json:"containerPort" yaml:"containerPort"`
	// HostPort      int    `json:"hostPort" yaml:"hostPort"`
	Protocol string `json:"protocol" yaml:"protocol"`
}

type PodResourceLimit struct {
	Cpu    string `json:"cpu" yaml:"cpu"`
	Memory string `json:"memory" yaml:"memory"`
}

type PodResourceRequest struct {
	Cpu    string `json:"cpu" yaml:"cpu"`
	Memory string `json:"memory" yaml:"memory"`
}

type PodResource struct {
	ResourceLimit   PodResourceLimit   `json:"limits" yaml:"limits"`
	ResourceRequest PodResourceRequest `json:"requests" yaml:"requests"`
}

type PodVolume struct {
	Name     string `json:"name" yaml:"name"`
	HostPath string `json:"hostPath" yaml:"hostPath"`
}

type PodSpec struct {
	Containers []PodContainer `json:"containers" yaml:"containers"`
	Volumes    []PodVolume    `json:"volumes" yaml:"volumes"`
}

type YamlPod struct {
	ApiVersion   string          `json:"apiVersion" yaml:"apiVersion"`
	Kind         string          `json:"kind" yaml:"kind"`
	Metadata     PodMetadata     `json:"metadata" yaml:"metadata"`
	NodeSelector PodNodeSelector `json:"nodeSelector" yaml:"nodeSelector"`
	Spec         PodSpec         `json:"spec" yaml:"spec"`
}

type Pod struct {
	PodId           string           `json:"podId" yaml:"podId"`
	PodIp           string           `json:"podIp" yaml:"podIp"`
	PodState        int              `json:"podState" yaml:"podState"`
	ContainerStates []ContainerState `json:"containerStates" yaml:"containerStates"`
	Start           time.Time        `json:"start" yaml:"start"`
	NodeId          string           `json:"nodeId" yaml:"nodeId"`
	RestartNum      int              `json:"restartNum" yaml:"restartNum"`
	RestartTime     time.Time        `json:"restartTime" yaml:"restartTime"`
	Mutex           sync.Mutex       `json:"mutex" yaml:"mutex"`
	YamlPod
}

type PodResourceUsed struct {
	PodName  string    `json:"podId" yaml:"podId"`
	CpuUsed  uint64    `json:"cpuUsed" yaml:"cpuUsed"`
	MemUsed  uint64    `json:"memUsed" yaml:"memUsed"`
	TimeUsed time.Time `json:"timeUsed" yaml:"timeUsed"`
}

// PodResourceSend used to send back to kubectl to show describe-pod result to user.
type PodResourceSend struct {
	PodName                string                   `json:"podName"`
	PodIP                  string                   `json:"podIP"`
	MemUsed                string                   `json:"memUsed"`
	LimMem                 string                   `json:"limMem"`
	CpuUsed                uint64                   `json:"cpuUsed"`
	LimCpu                 float64                  `json:"limCpu"`
	ContinueTime           uint64                   `json:"continueTime"`
	RestartNum             int                      `json:"restartNum"`
	ContainerResourcesSend []*ContainerResourceSend `json:"ContainerResourcesSend"`
}

type GetPodsSend struct {
	Name         string `json:"name"`
	ReadyNum     int    `json:"readyNum"`
	CurState     string `json:"curState"`
	RestartNum   int    `json:"restartNum"`
	ContinueTime uint64 `json:"continueTime"`
	Ip           string `json:"ip"`
}

type GetPods struct {
	PodsSend []*GetPodsSend `json:"podsSend"`
}

type HandlePodResult struct {
	Del []*Pod `json:"del"`
	Add []*Pod `json:"add"`
}

type HandlePodNameResult struct {
	Del []string `json:"del"`
	Add []string `json:"add"`
}

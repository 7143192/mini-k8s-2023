package defines

import "time"

// different states of one container.
const (
	Pending = 0
	Running = 1
	Succeed = 2
	Failed  = 3
	Unknown = 4
)

const PauseContainerImage = "registry.aliyuncs.com/google_containers/pause"

type ContainerID string

// ContainerResourceUsed used to coordinate with cadVisor to get container
// detailed resource info, and also can be used when kubectl get / describe.

// ContainerResourceUsed used to coordinate with cadVisor to get container
// detailed resource info, and also can be used when kubectl get / describe.
type ContainerResourceUsed struct {
	CpuUsed  uint64    `json:"cpuUsed" yaml:"cpuUsed"`
	MemUsed  uint64    `json:"memUsed" yaml:"memUsed"`
	TimeUsed time.Time `json:"timeUsed" yaml:"timeUsed"`
}

// ContainerState used to present the current of container <name>
type ContainerState struct {
	Id    string `json:"id" yaml:"id"`
	Name  string `json:"name" yaml:"name"`
	State int    `json:"state" yaml:"state"`
}

type PodContainer struct {
	Name         string           `json:"name" yaml:"name"`
	Image        string           `json:"image" yaml:"image"`
	Command      []string         `json:"command" yaml:"command"`
	Args         []string         `json:"args" yaml:"args"`
	WorkingDir   string           `json:"workingDir" yaml:"workingDir"`
	VolumeMounts []PodVolumeMount `json:"volumeMounts" yaml:"volumeMounts"`
	Ports        []PodPort        `json:"ports" yaml:"ports"`
	Resource     PodResource      `json:"resources" yaml:"resources"`
}

// ContainerResourceSend used to send back container status to show to user.
type ContainerResourceSend struct {
	ConName   string `json:"conName"`
	ReadyNum  int    `json:"readyNum"`
	CurState  string `json:"curState"`
	TotalTime uint64 `json:"totalTime"`
}

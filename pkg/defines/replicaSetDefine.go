package defines

import "time"

const RSInstancePrefix = "ReplicaSet"

//Note that the name of replicaSet should NOT contain "/"

type ReplicaSet struct {
	ApiVersion string         `json:"apiVersion" yaml:"apiVersion"`
	Kind       string         `json:"kind" yaml:"kind"`
	Metadata   ReplicaSetMeta `json:"metadata" yaml:"metadata"`
	Spec       ReplicaSetSpec `json:"spec" yaml:"spec"`
	StartTime  time.Time      `json:"start-time" yaml:"start-time"`
}

type ReplicaSetInfo struct {
	Name      string    `json:"name"`
	Replicas  int32     `json:"replicas"`
	StartTime time.Time `json:"start-time"`
}

type DesRSInfo struct {
	Info    ReplicaSetInfo `json:"rs-info" yaml:"rs-info"`
	PodName []string       `json:"pods" yaml:"pods"`
}

type ReplicaSetMeta struct {
	Name string `json:"name" yaml:"name"`
}

type ReplicaSetSpec struct {
	Replicas int32              `json:"replicas" yaml:"replicas"`
	Selector ReplicaSetSelector `json:"selector" yaml:"selector"`
	Template ReplicaSetTemplate `json:"template" yaml:"template"`
}

type ReplicaSetSelector struct {
	MatchLabels PodLabels `json:"matchLabels" yaml:"matchLabels"`
}

type ReplicaSetTemplate struct {
	Metadata RSMetadata `json:"metadata" yaml:"metadata"`
	Spec     PodSpec    `json:"spec" yaml:"spec"`
}

type RSMetadata struct {
	Label PodLabels `json:"labels" yaml:"labels"`
}

type AllReplicaPodNames struct {
	Names []string `json:"names" yaml:"names"`
}

type ReplicaPodState struct {
	Live bool `json:"live" yaml:"live"`
}

package defines

import "time"

const DeploymentPrefix = "Deployment"
const DeploymentListPrefix = "DeploymentList"
const DeployPodNamesPrefix = "DeployPodNames"
const DeploymentReplicaSetPrefix = "DeployReplicaSet_"

type DeployMetadata struct {
	Name string `yaml:"name" json:"name"`
}

type DeployLabels struct {
	App string `yaml:"app" json:"app"`
	Env string `yaml:"env" json:"env"`
}

type DeploySelector struct {
	Labels DeployLabels `yaml:"matchLabels" json:"matchLabels"`
}

type TemplateMetadata struct {
	Name   string       `yaml:"name" json:"name"`
	Labels DeployLabels `yaml:"labels" json:"labels"`
}

type DeployTemplate struct {
	Metadata TemplateMetadata `yaml:"metadata" json:"metadata"`
	Spec     PodSpec          `yaml:"spec" json:"spec"`
}

type DeploySpec struct {
	Replicas int            `yaml:"replicas" json:"replicas"`
	Selector DeploySelector `yaml:"selector" json:"selector"`
	Template DeployTemplate `yaml:"template" json:"template"`
}

type Deployment struct {
	ApiVersion string         `yaml:"apiVersion" json:"apiVersion"`
	Kind       string         `yaml:"kind" json:"kind"`
	Metadata   DeployMetadata `yaml:"metadata" json:"metadata"`
	Spec       DeploySpec     `yaml:"spec" json:"spec"`
}

type EtcdDeployment struct {
	DeployName string      `yaml:"deployName" json:"deployName"`
	Replicas   int         `yaml:"replicas" json:"replicas"`
	PodName    string      `yaml:"podName" json:"podName"`
	StartTime  time.Time   `yaml:"startTime" json:"startTime"`
	RsInfo     *ReplicaSet `yaml:"rsInfo" json:"rsInfo"`
}

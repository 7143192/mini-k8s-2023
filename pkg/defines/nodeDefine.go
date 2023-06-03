package defines

import (
	"time"
)

const NodePrefix = "Node"
const NodeInfoPrefix = "NodeInstance"
const NodePodsListPrefix = "NodePodsList"
const AllNodeSetPrefix = "allNode"
const CurNodeIdPrefix = "curNodeID"
const NodeResourcePrefix = "nodeResource"
const NodeHeartBeatPrefix = "nodeHeartBeat"

const NodeReady = 0
const NodeNotReady = 1

// NonAliveTime if a node does not send its heartbeat to master for more
// than NonAliveTime period, regard this node as a dead one for now.
const NonAliveTime = 40 * time.Second

// node can also be configured by YAML file, which is similar to pod.

type NodeMetadata struct {
	Name     string       `json:"name" yaml:"name"`
	Label    string       `json:"label" yaml:"label"`
	Port     int          `json:"port" yaml:"port"`
	Ip       string       `json:"ip" yaml:"ip"`
	Selector NodeSelector `json:"selector" yaml:"selector"`
}

// NodeYaml struct used to record info from node configuration file.
type NodeYaml struct {
	ApiVersion string `json:"apiVersion" yaml:"apiVersion"`
	// should be "Node" here!!!
	Kind     string       `json:"kind" yaml:"kind"`
	Metadata NodeMetadata `json:"metadata" yaml:"metadata"`
}

type NodeSelector struct {
	Gpu string `yaml:"gpu" json:"gpu"`
}

// Node is used to record the detailed info for one node
// (including some info not configured in th Yaml file.)
type Node struct {
	NodeId    string   `json:"nodeId" yaml:"nodeId"`
	NodeSpec  NodeYaml `json:"nodeSpec" yaml:"nodeSpec"`
	NodeState int      `json:"nodeState" yaml:"nodeState"`
	// the subnetwork ip should be allocated when registering a new node to master.
	CniIp    string       `json:"cniIp" yaml:"cniIp"`
	Selector NodeSelector `json:"selector" yaml:"selector"`
}

// NodeInfo used to record detailed node information and some other messages required
// in programming. aka., this struct should be used when programming with node object.
type NodeInfo struct {
	NodeData Node `json:"nodeData" yaml:"nodeData"`
	// pods that belong to this node.
	Pods       []*Pod `json:"pods" yaml:"pods"`
	Registered bool   `json:"registered" yaml:"registered"`
	// EtcdClient     *clientv3.Client
	// CadvisorClient *client.Client
}

type NodeResourceUsed struct {
	NodeId   string    `json:"nodeId" yaml:"nodeId"`
	MemUsed  uint64    `json:"memUsed" yaml:"memUsed"`
	MemTotal uint64    `json:"memTotal" yaml:"memTotal"`
	CpuUsed  uint64    `json:"cpuUsed" yaml:"cpuUsed"`
	CpuTotal uint64    `json:"cpuTotal" yaml:"cpuTotal"`
	Time     time.Time `json:"time" yaml:"time"`
}

// NodeHeartBeat is used to tell master this node is still alive.
type NodeHeartBeat struct {
	NodeId string `json:"nodeId" yaml:"nodeId"`
	// these two fields are used to compute whether the node is DEAD.
	// send time for this time.
	CurTime time.Time `json:"curTime" yaml:"curTime"`
	// last send time.
	LastTime time.Time `json:"lastTime" yaml:"lastTime"`
}

type NodeResourceSend struct {
	NodeName          string             `json:"nodeName"`
	NodeId            string             `json:"nodeId"`
	ReadyNum          int                `jon:"readyNum"`
	MemUsed           string             `json:"memUsed"`
	MemTotal          string             `json:"memTotal"`
	CpuUsed           float64            `json:"cpuUsed"`
	CpuTotal          float64            `json:"cpuTotal"`
	NodeIP            string             `json:"nodeIP"`
	PodsResourcesSend []*PodResourceSend `json:"PodsResourcesSend"`
}

type GetNodeResourceSend struct {
	Name  string `json:"name"`
	Id    string `json:"id"`
	Label string `json:"label"`
	Ip    string `json:"ip"`
	State string `jon:"state"`
}

type GetNodesResource struct {
	NodesSend []*GetNodeResourceSend `json:"nodesSend"`
}

type NodeResourceInfo struct {
	Total      uint64  `json:"total"`
	Used       uint64  `json:"used"`
	CpuPercent float64 `json:"cpuPercent"`
}

package config

import "sync"

// IP local client IP.
// NOTE: this IP should be discarded as we can take node's IP as kubelet IP and Proxy IP.
const IP = "localhost"

const WorkerIP = "127.0.0.1"

// MasterIP TODO: single-node version
const MasterIP = "localhost"

const EtcdPort1 = "2379"
const EtcdPort2 = "22379"
const EtcdPort3 = "32379"

// CadvisorPort cadVisor default port is 8080, can be changed by "--port" arg in cmd line.
const CadvisorPort = "8090"

// MasterPort api-server port.
const MasterPort = "8080"

// KubectlPort kubectl running port.
const KubectlPort = "9000"

// WorkerPort kubelet running port.
const WorkerPort = "8000"

// ProxyPort KubeProxy running port.
const ProxyPort = "9090"

// SchedulerPort scheduler running port.
const SchedulerPort = "9999"

// RCPort replicaSet-controller's running port
const RCPort = "8888"

// SLPort serverless-controller's running port
const SLPort = "8081"

// AutoScalerPort autoScaler controller running port.
const AutoScalerPort = "8889"

const DNSPort = "8990"

// FlannelPluginDir the default path of flannel plugin binaries
const FlannelPluginDir = "/opt/cni/bin"

// FlannelConfDir the default path of flannel configuration directory.
const FlannelConfDir = "/etc/cni/net.d"

const PluginConfFilePath = "/etc/cni/net.d/.conf"

const FlannelConfDir1 = "/home/os"

const DefaultInterfacePrefix = "eth"

const FlannelEtcdPrefix = "/coreos.com/network"

const FlannelIP = "10.5.0.0/16"

// NetNameSpace for test
const NetNameSpace = "/var/run/netns/test-3"

// NetNameSpace1 for test
const NetNameSpace1 = "/home/os/ns"

const DockerApiVersion = "1.41"

const NsNamePrefix = "ns-"

const GlobalNetworkName = "test-net"

const KubeSvcChainPrefix = "KUBE-SVC-"

const KubeSvcMainChainName = "KUBE-SVC-CHAIN"

const PodDNATChainPrefix = "KUBE-SEP-"

var PodMutex sync.Mutex

const NginxConfFilePathPrefix = "/home/os/nginx/"

const NginxConfFileDir = "/etc/nginx"

const NginxServerNamePrefix = "nginx-"

const CoreDNSServerName = "coredns_server"

const CoreDNSFilePath = "/home/os/coredns_conf/Corefile"

const GPUWorkDirPrefix = "/lustre/home/acct-stu/stu1643/"

const GPUCompileFileName = "compile.sh"

const ServerlessFileDir = "/home/os/minik8s/serverless/functions"

const ServerlessTmpFileDir = "/home/os/minik8s/serverless/tmp"

const SourceDir = "./utils/serverless"

const ServerlessTemplatePod = "./utils/serverless/pod.yaml"

const Dependency = "Flask==2.0.2\n"

const CoreDNSServerIP = "10.5.98.2"

var ApiServerMutex sync.Mutex

const CharSet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

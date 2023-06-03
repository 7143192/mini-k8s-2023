package defines

const DNSPrefix = "DNS"

const AllDNSSetPrefix = "DNSList"

type DNSMetadata struct {
	Name string `json:"name" yaml:"name"`
}

type DNSPath struct {
	PathAddr    string `json:"pathAddr" yaml:"pathAddr"`
	ServiceName string `json:"serviceName" yaml:"serviceName"`
	Port        int    `json:"port" yaml:"port"`
}

type DNSSpec struct {
	Host  string     `json:"host" yaml:"host"`
	Paths []*DNSPath `json:"paths" yaml:"paths"`
}

type DNSYaml struct {
	ApiVersion string      `json:"apiVersion" yaml:"apiVersion"`
	Kind       string      `json:"kind" yaml:"kind"`
	Metadata   DNSMetadata `json:"metadata" yaml:"metadata"`
	Spec       DNSSpec     `json:"spec" yaml:"spec"`
}

type EtcdDNSPath struct {
	PathAddr    string `json:"pathAddr" yaml:"pathAddr"`
	ServiceName string `json:"serviceName" yaml:"serviceName"`
	Port        int    `json:"port" yaml:"port"`
	ServiceIp   string `json:"serviceIp" yaml:"serviceIp"`
}

type EtcdDNS struct {
	DNSName  string         `json:"dnsName" yaml:"dnsName"`
	DNSHost  string         `json:"dnsHost" yaml:"dnsHost"`
	DNSPaths []*EtcdDNSPath `json:"dnsPaths" yaml:"dnsPaths"`
}

type DNSInfo struct {
}

type DNSInfoSend struct {
	DNSInfos []*EtcdDNS `json:"dnsInfos" yaml:"dnsInfos"`
}

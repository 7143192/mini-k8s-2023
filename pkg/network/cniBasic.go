package network

import (
	gocni "github.com/containerd/go-cni"
	"log"
	"mini-k8s/pkg/config"
)

//var (
//	idF    = flag.Int("id", 2, "")
//	confFF = flag.String("conf", "/etc/cni/net.d/11-flannel.conf", "")
//)
//
//func init() {
//	flag.Parse()
//}

// CNIStart call this function when create a new node to start a new CNI Instance for the node's network
func CNIStart(pluginDir []string, pluginConfDir string) gocni.CNI {
	//l, err := gocni.New(
	//	gocni.WithMinNetworkCount(2),
	//	gocni.WithPluginConfDir(pluginConfDir),
	//	gocni.WithPluginDir(pluginDir),
	//	gocni.WithInterfacePrefix(config.DefaultInterfacePrefix))
	//if err != nil {
	//	log.Fatalf("failed to initialize cni library: %v", err)
	//}
	//
	//if err := l.Load(gocni.WithLoNetwork, gocni.WithDefaultConf); err != nil {
	//	log.Fatalf("failed to load cni configuration: %v", err)
	//}
	//
	//return l

	l, err := gocni.New(
		gocni.WithMinNetworkCount(2),
		gocni.WithPluginConfDir(pluginConfDir),
		gocni.WithPluginDir(pluginDir),
		gocni.WithInterfacePrefix(config.DefaultInterfacePrefix))
	if err != nil {
		log.Fatalf("failed to initialize cni library: %v", err)
	}

	if err := l.Load(gocni.WithLoNetwork, gocni.WithConfFile(config.PluginConfFilePath)); err != nil {
		log.Fatalf("failed to load cni configuration: %v", err)
	}

	return l
}

func CNIEnd() {

}

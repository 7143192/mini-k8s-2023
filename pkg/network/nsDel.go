package network

import (
	"fmt"
	"mini-k8s/pkg/config"
	"os/exec"
)

func DelNamespace(nsPath string, id string) {
	pluginDir := []string{config.FlannelPluginDir}
	cni := CNIStart(pluginDir, config.FlannelConfDir1)
	RemoveNetns(cni, nsPath, id)
}

func DelNsFile(nsPath string) bool {
	arg1 := nsPath
	fmt.Printf("del ns file arg1 = %v\n", arg1)
	cmd := exec.Command("umount", arg1)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("an error occurs when unmount an old ns: %v\n", err)
		return false
	}
	cmd1 := exec.Command("rm", arg1)
	err = cmd1.Run()
	if err != nil {
		fmt.Printf("an error occurs when delete an old ns file: %v\n", err)
		return false
	}
	return true
}

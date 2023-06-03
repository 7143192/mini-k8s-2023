package network

import (
	"fmt"
	"mini-k8s/pkg/config"
	"os/exec"
	"strings"
)

// this file contains methods required to create a new non-duplicated namespace for a new pod.
// name format of the namespace for the new pod : ns-POD_NAME

func CreateNamespace(id string) string {
	nsName := config.NsNamePrefix + id
	cmd := exec.Command("ip", "netns", "add", nsName)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("an error occurs when creating a new namespace %v for pod %v: %v\n", nsName, id, err)
		return ""
	}
	// return new namespace name allocated for the new pod here.
	return nsName
}

func CheckCreateNamespaceSuccess(nsName string) bool {
	cmd := exec.Command("ip", "netns", "list")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("an error occurs when checking correctness of one new namespace: %v: %v\n", nsName, err)
		return false
	}
	got := string(output)
	fmt.Printf("all namespaces =\n %v\n", got)
	if strings.Contains(got, nsName) {
		return true
	}
	return false
}

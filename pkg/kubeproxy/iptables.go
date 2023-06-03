package kubeproxy

import (
	"fmt"
	"mini-k8s/pkg/config"
	"os/exec"
	"strings"
)

func InitSvcMainChain() {
	if !IsChainExist(config.KubeSvcMainChainName, "nat") {
		CreateChain("nat", config.KubeSvcMainChainName)
		cmd := exec.Command("iptables", "-A", "PREROUTING", "-t", "nat", "-j", config.KubeSvcMainChainName)
		err := cmd.Run()
		if err != nil {
			fmt.Println("Error executing command: ", err)
		}
		cmd = exec.Command("iptables", "-A", "OUTPUT", "-t", "nat", "-j", config.KubeSvcMainChainName)
		err = cmd.Run()
		if err != nil {
			fmt.Println("Error executing command: ", err)
		}
	}
}

func CreateChain(tableType string, chainName string) {
	cmd := exec.Command("iptables", "-t", tableType, "-N", chainName)
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error executing command: ", err)
	}
}

func DeleteChain(tableType string, chainName string) {
	cmd := exec.Command("iptables", "-t", tableType, "-X", chainName)
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error executing command: ", err)
	}
}

func AddSvcMatchRuleToChain(chainName string, tableType string, destIp string, destPort string, protocol string, targetChain string) {
	cmd := exec.Command("iptables", "-A", chainName, "-t", tableType, "-d", destIp, "-p", protocol, "--dport", destPort, "-j", targetChain)
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error executing command: ", err)
	}
}

func AddSvcForwardRuleToChain(chainName string, tableType string, targetChain string, probability string) {
	cmd := exec.Command("iptables", "-A", chainName, "-t", tableType, "-m", "statistic", "--mode", "random", "--probability", probability, "-j", targetChain)
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error executing command: ", err)
	}
}

func DeleteSvcForwardRuleFromChain(chainName string, tableType string, targetChain string, probability string) {
	cmd := exec.Command("iptables", "-D", chainName, "-t", tableType, "-m", "statistic", "--mode", "random", "--probability", probability, "-j", targetChain)
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error executing command: ", err)
	}
}

func AddSvcDNATRuleToChain(chainName string, destAddr string) {
	cmd := exec.Command("iptables", "-A", chainName, "-t", "nat", "-p", "tcp", "-j", "DNAT", "--to-destination", destAddr)
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error executing command: ", err)
	}
}

func DeleteSvcDNATRuleFromChain(chainName string, destAddr string) {
	cmd := exec.Command("iptables", "-D", chainName, "-t", "nat", "-p", "tcp", "-j", "DNAT", "--to-destination", destAddr)
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error executing command: ", err)
	}
}

func IsChainExist(chainName string, tableType string) bool {
	cmd := exec.Command("iptables", "-t", tableType, "-nL")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error executing command: ", err)
		return false
	}
	outputStr := string(output)

	// 检查outputStr中是否包含指定的chainName
	if strings.Contains(outputStr, chainName) {
		return true
	} else {
		return false
	}
}

func GetRuleNumFromChain(chainName string) string {
	cmd1 := exec.Command("iptables", "-L", chainName, "-t", "nat", "--line-numbers")
	cmd2 := exec.Command("tail", "-n", "+3")
	cmd3 := exec.Command("wc", "-l")

	cmd2.Stdin, _ = cmd1.StdoutPipe()
	cmd3.Stdin, _ = cmd2.StdoutPipe()
	output, _ := cmd3.Output()

	count := strings.TrimSpace(string(output))
	return count
}

func GetChainRuleTarget(chainName string) []string {
	cmd := exec.Command("bash", "-c", "iptables -L "+chainName+" -t nat --line-numbers | tail -n +3 | awk '{print $2}'")
	outputBytes, _ := cmd.Output()

	output := string(outputBytes)
	fmt.Printf("output is %v\n", output)
	targets := strings.Fields(output)
	return targets
}

func DeleteAllRuleInChain(chainName string) {
	cmd := exec.Command("iptables", "-t", "nat", "-F", chainName)
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error executing command: ", err)
	}
}

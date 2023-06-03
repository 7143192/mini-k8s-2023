package dns

import (
	"fmt"
	"mini-k8s/pkg/config"
	"os"
	"os/exec"
	"strings"
)

//func InitCoreDNSServer() {
//	cmd := exec.Command("docker start coredns_server")
//	err := cmd.Run()
//	if err != nil {
//		fmt.Printf("an error occurs when run cmd: %v\n", err)
//	}
//	cmd = exec.Command("docker exec ")
//}

func AddHost(host string, proxyIp string) {
	//cli, err := client.NewClientWithOpts(client.WithVersion(config.DockerApiVersion))
	//if err != nil {
	//	fmt.Printf("an error occurs when creating a docker client: %v\n", err)
	//}
	//inspect, err := cli.ContainerInspect(context.Background(), config.CoreDNSServerName)
	//if err != nil {
	//	panic(err)
	//}
	//containerID := inspect.ID
	// 读取 Corefile 文件
	corefileBytes, err := os.ReadFile(config.CoreDNSFilePath)
	if err != nil {
		fmt.Println("Error reading Corefile:", err)
		//os.Exit(1)
	}
	// 将 Corefile 转换为字符串
	corefileStr := string(corefileBytes)
	// 查找 hosts 部分的位置
	hostsIndex := strings.Index(corefileStr, "hosts {")
	if hostsIndex == -1 {
		fmt.Println("Error: hosts section not found in Corefile")
		//os.Exit(1)
	}
	// 在 hosts 部分末尾添加新的解析规则
	newRule := "\t" + proxyIp + " " + host
	newCorefile := corefileStr[:hostsIndex+7] + "\n" + newRule + corefileStr[hostsIndex+7:]
	// 将修改后的 Corefile 写回文件
	err = os.WriteFile(config.CoreDNSFilePath, []byte(newCorefile), 0644)
	if err != nil {
		fmt.Println("Error writing Corefile:", err)
		//cd os.Exit(1)
	}
	// 重启 CoreDNS
	//err = executeCommand("systemctl restart coredns")
	//if err != nil {
	//	fmt.Println("Error restarting CoreDNS:", err)
	//	os.Exit(1)
	//}
	//fmt.Println("Successfully added new host rule to Corefile!")
	//err = cli.ContainerKill(context.Background(), containerID, "SIGUSR1")
	//if err != nil {
	//	panic(err)
	//}

}

// 执行命令并返回错误
func executeCommand(command string) error {
	cmd := exec.Command("sh", "-c", command)
	return cmd.Run()
}

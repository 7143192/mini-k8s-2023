package dns

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"mini-k8s/pkg/config"
	"mini-k8s/pkg/defines"
	"os"
	"os/exec"
	"strconv"
)

type Location struct {
	path string
	addr string
}

func StartNewNginxServer(dns *defines.EtcdDNS) string {
	nginxConfFilePath := config.NginxConfFilePathPrefix + dns.DNSName + "/nginx.conf"
	nginxConfFileDir := config.NginxConfFilePathPrefix + dns.DNSName
	nginxServerName := config.NginxServerNamePrefix + dns.DNSName
	err := WriteNginxConfFile(nginxConfFilePath, dns)
	if err != nil {
		fmt.Printf("error when write nginx conf file，%v\n", err)
	}
	cli, err := client.NewClientWithOpts(client.WithVersion(config.DockerApiVersion))
	if err != nil {
		fmt.Printf("an error occurs when creating a docker client: %v\n", err)
	}
	// 配置Nginx容器
	containerConfig := &container.Config{
		Image: "nginx",
	}

	// 配置Nginx容器的主机配置文件
	hostConfig := &container.HostConfig{
		Binds: []string{
			nginxConfFileDir + ":" + config.NginxConfFileDir,
		},
	}

	// 创建Nginx容器
	resp, err := cli.ContainerCreate(context.Background(), containerConfig, hostConfig, nil, nil, nginxServerName)
	if err != nil {
		panic(err)
	}

	// 启动Nginx容器
	if err := cli.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	fmt.Println("Nginx容器已启动")
	inspect, err := cli.ContainerInspect(context.Background(), resp.ID)
	if err != nil {
		panic(err)
	}
	ipAddr := inspect.NetworkSettings.IPAddress
	return ipAddr
}

func ReloadNginx(nginxName string) error {
	err := exec.Command("docker", "exec", nginxName, "nginx", "-s", "reload").Run()
	if err != nil {
		return err
	}
	return nil
}

func WriteNginxConfFile(nginxConfFilePath string, dns *defines.EtcdDNS) error {
	err1 := os.Mkdir(config.NginxConfFilePathPrefix+dns.DNSName, 0755)
	if err1 != nil {
		return err1
	}
	file, err := os.OpenFile(nginxConfFilePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString("events {\n\tworker_connections 1024;\n}\n\nhttp {\n")

	if err != nil {
		return err
	}

	_, err = file.WriteString("\tserver {\n\t\tlisten 80;\n\t\tserver_name " + dns.DNSHost + ";\n")
	if err != nil {
		return err
	}
	for _, path := range dns.DNSPaths {
		_, err = file.WriteString("\t\tlocation = " + path.PathAddr + " {\n\t\t\tproxy_pass http://" + path.ServiceIp + ":" + strconv.Itoa(path.Port) + "/;\n\t\t}\n")
		if err != nil {
			return err
		}
	}
	_, err = file.WriteString("\t}\n\n")
	if err != nil {
		return err
	}

	_, err = file.WriteString("}")
	if err != nil {
		return err
	}
	return nil
}

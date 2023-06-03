package container

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"log"
	"mini-k8s/pkg/config"
	"os"
	"strings"
)

func CopyToContainer(containerID string, srcPath string, dstPath string) error {
	cli, err := client.NewClientWithOpts(client.WithVersion(config.DockerApiVersion))
	defer cli.Close()
	if err != nil {
		log.Printf("error occurs when creating a new docker client\n")
		return err
	}
	contentByte, err := os.ReadFile(srcPath)
	if err != nil {
		log.Printf("error occurs when reading data from path %v\n", srcPath)
	}
	//reader := bytes.NewReader(contentByte)
	//buf := make([]byte, len(contentByte))
	//reader.Read(buf)
	//fmt.Printf("read data = %v\n", string(buf))
	lastIdx := strings.LastIndex(srcPath, "/")
	fileName := ""
	if lastIdx < 0 {
		fileName = srcPath
	} else {
		fileName = srcPath[lastIdx+1:]
	}
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	err = tw.WriteHeader(&tar.Header{
		Name: fileName,
		Mode: 0777,
		Size: int64(len(string(contentByte))),
	})
	if err != nil {
		fmt.Printf("docker copy: %v\n", err)
	}
	tw.Write(contentByte)
	tw.Close()
	err = cli.CopyToContainer(context.Background(), containerID, dstPath, &buf, types.CopyToContainerOptions{
		AllowOverwriteDirWithFile: true,
	})
	if err != nil {
		log.Printf("error occurs when copying file data to container: %v\n", err)
		return err
	}
	return nil
}

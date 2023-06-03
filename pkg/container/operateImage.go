package container

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"io"
	"log"
	"mini-k8s/pkg/config"
	"mini-k8s/pkg/defines"
	"os"
	"os/exec"
)

/* this file is used to implement some methods required when operating on one container image. */

// CheckImageExist used to check whether @image has been pulled to local.
func CheckImageExist(image string) bool {
	cli, err := client.NewClientWithOpts(client.WithVersion(config.DockerApiVersion))
	defer cli.Close()
	if err != nil {
		fmt.Printf("an error occurs when creating container client in CheckImageExist function: %v\n", err)
		return false
	}
	localImages, err := cli.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		fmt.Printf("an error occurs when get all local image in CheckImageExist function: %v\n", err)
		return false
	}
	for _, localImage := range localImages {
		localImageTags := localImage.RepoTags
		for _, localImageTag := range localImageTags {
			if localImageTag == image {
				// find a corresponding local image.
				return true
			}
		}
	}
	return false
}

// PullImage used to pull a @target image to local.
func PullImage(target string) bool {
	cli, err := client.NewClientWithOpts(client.WithVersion(config.DockerApiVersion))
	defer cli.Close()
	if err != nil {
		fmt.Printf("an error occurs when creating container client in CheckImageExist function: %v\n", err)
		return false
	}
	got, err := cli.ImagePull(context.Background(), target, types.ImagePullOptions{})
	_, err1 := io.Copy(io.Discard, got)
	if err1 != nil {
		fmt.Printf("an error occurs when pulling image %s: %v\n", target, err1)
		return false
	}
	// fmt.Printf("successfully pull image %s to local!\n", target)
	return true
}

func PushImage(imageName string) {
	image := imageName
	cli, err := client.NewClientWithOpts(client.WithVersion(config.DockerApiVersion))
	if err != nil {
		fmt.Printf("%v\n", err)
		panic(err)
	}
	defer cli.Close()
	// RegistryAuth is the base64 encoded credentials for the registry
	// So create a types.AuthConfig and translate it into base64
	authConfig := types.AuthConfig{Username: defines.DockerUsername, Password: defines.DockerPwd}
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		panic(err)
	}
	authStr := base64.URLEncoding.EncodeToString(encodedJSON)
	resp, err := cli.ImagePush(context.Background(), image, types.ImagePushOptions{
		All:           false,
		RegistryAuth:  authStr,
		PrivilegeFunc: nil,
	})
	if err != nil {
		fmt.Printf("Pushed imageID failed %s\n %v\n", image, err)
		panic(err)
	}
	_, err = io.Copy(os.Stdout, resp)
	if err != nil {
		fmt.Printf("%v\n", err)
		panic(err)
	} else {
		fmt.Printf("Pushed imageID %s successully\n", image)
	}
}

func TagNewImage(imageName string, tag string) {
	cli, _ := client.NewClientWithOpts(client.WithVersion(config.DockerApiVersion))
	cli.ImageTag(context.Background(), imageName, tag)
}

func RemoveImage(imgName string) error {
	//cli, _ := client.NewClientWithOpts(client.WithVersion(config.DockerApiVersion))
	//options := types.ImageRemoveOptions{}
	//options.Force = true // remove the image by force.
	//_, err := cli.ImageRemove(context.Background(), imgName, options)
	cmd := exec.Command("docker", "rmi", imgName)
	err := cmd.Run()
	if err != nil {
		log.Printf("[container] fail to delete image %v: %v.", imgName, err)
		return err
	}
	return nil
}

func ListImage() ([]types.ImageSummary, error) {
	cli, _ := client.NewClientWithOpts(client.WithVersion(config.DockerApiVersion))
	options := types.ImageListOptions{}
	options.All = true // should list all images
	res, err := cli.ImageList(context.Background(), options)
	if err != nil {
		log.Printf("[container] fail tp list all images: %v\n", err)
	}
	return res, err
}

func CheckImgInList(imgName string) bool {
	res, err := ListImage()
	if err != nil {
		log.Printf("[container] fail tp list all images: %v\n", err)
		return false
	}
	for _, img := range res {
		tags := img.RepoTags
		for _, tag := range tags {
			if tag == imgName {
				log.Printf("[container] found image %v!\n", imgName)
				return true
			}
		}
	}
	return false
}

package container

import (
	"fmt"
	uuid "github.com/satori/go.uuid"
	"mini-k8s/pkg/apiserver"
	"mini-k8s/pkg/container"
	"mini-k8s/pkg/defines"
	"mini-k8s/utils/yaml"
	"testing"
)

func TestContainer1(t *testing.T) {
	gpuJob, _ := yaml.ParseGPUJobConfig("../../utils/templates/gpu_template.yaml")
	apiserver.GenerateSlurmScript(gpuJob)
	// then copy the file into container.
	con := &defines.PodContainer{}
	con.Image = defines.GPUJobBasicImageName
	conName := defines.GPUJobContainerNamePrefix + "_" + gpuJob.Name
	id := string(container.CreateGPUJobContainer(conName, con))
	// a random new image name.
	imageName := defines.GPUJobBasicImageName + ":" + gpuJob.Name + "-" + uuid.NewV4().String()
	gpuJob.ImageName = imageName
	container.CopyToContainer(id, defines.SlurmFileLocalPath, defines.ContainerSlurmPath)
	newId := container.CommitContainer(id)
	fmt.Printf("newId = %v\n", newId)
	imageID := newId[7:]
	container.TagNewImage(imageID, imageName)
	container.PushImage(imageName)
}

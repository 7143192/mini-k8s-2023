package serverless_test

import (
	"fmt"
	"mini-k8s/pkg/container"
	"testing"
)

func TestImage(t *testing.T) {
	res, _ := container.ListImage()
	for _, img := range res {
		fmt.Println(img.ID)
		fmt.Println()
		fmt.Println(img.Labels)
		fmt.Println()
		fmt.Println(img.RepoTags)
		fmt.Println()
		fmt.Println(img.RepoDigests)
	}
	container.RemoveImage("7143192/auto1")
}

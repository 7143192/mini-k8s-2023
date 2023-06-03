package network

import (
	"context"
	"github.com/containerd/go-cni"
	"log"
)

func RemoveNetns(l cni.CNI, netns string, id string) {
	ctx := context.Background()
	if err := l.Remove(ctx, id, netns); err != nil {
		log.Fatalf("failed to teardown network: %v", err)
	}
}

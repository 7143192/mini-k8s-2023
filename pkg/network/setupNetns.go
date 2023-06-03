package network

import (
	"context"
	"fmt"
	gocni "github.com/containerd/go-cni"
	"log"
)

func SetupNetns(l gocni.CNI, netns string, id string) string {
	ctx := context.Background()

	result, err := l.Setup(ctx, id, netns)
	if err != nil {
		log.Fatalf("failed to setup network for namespace: %v", err)
	}

	IP := ""
	for key, iff := range result.Interfaces {
		if len(iff.IPConfigs) > 0 {
			tmpIP := iff.IPConfigs[0].IP.String()
			fmt.Printf("IP of the interface %s:%s\n", key, tmpIP)
			if key == "eth0" {
				// we only need the v ip for the eth0.
				IP = tmpIP
			}
			//IP = iff.IPConfigs[0].IP.String()
			//fmt.Printf("IP of the interface %s:%s\n", key, IP)
		}
	}
	return IP
}

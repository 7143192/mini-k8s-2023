package main

import (
	"mini-k8s/pkg/defines"
	"mini-k8s/pkg/gpu"
	"mini-k8s/utils/ssh"
)

func main() {
	sshClient := ssh.NewSSHClient()
	jobName := gpu.GetJobName("/home/os/src/" + defines.SlurmScriptName)
	gpu.Run(sshClient, jobName)
}

package gpu

import (
	"bufio"
	"fmt"
	"github.com/melbahja/goph"
	"mini-k8s/pkg/config"
	"mini-k8s/pkg/defines"
	"mini-k8s/utils/ssh"
	"os"
	"strings"
	"time"
)

const (
	username = "stu1643"
)

var jobID string

func Compile(cli *goph.Client, compilePath string) {
	_, err := cli.Run(compilePath)
	if err != nil {
		fmt.Printf("an error occurs when run compile file %v\n", err)
	}
	fmt.Printf("finish compile\n")
}

func Mkdir(cli *goph.Client, dirPath string) {
	cmd := "mkdir " + dirPath
	_, err := cli.Run(cmd)
	if err != nil {
		fmt.Printf("an error occurs when run mkdir %v\n", err)
	}
}

func CD(cli *goph.Client, dirPath string) {
	//fmt.Printf("cd to %v\n", dirPath)
	//cmd := "cd " + dirPath
	//respBy, err := cli.Run(cmd)
	//if err != nil {
	//	fmt.Printf("an error occurs when run cd %v\n", err)
	//}
	//fmt.Printf("cd result is %v\n", string(respBy))
	//cmd = "pwd"
	//respBytes, err1 := cli.Run(cmd)
	//if err1 != nil {
	//	fmt.Printf("error pwd\n")
	//}
	//fmt.Printf("pwd is: %v\n", string(respBytes))
	cmd, err := cli.Command("cd", dirPath)
	if err != nil {
	}
	err = cmd.Run()
	if err != nil {
		fmt.Printf("an error occurs when run cd %v\n", err)
	}
	cmd, err = cli.Command("pwd", dirPath)
	if err != nil {
	}
	err = cmd.Run()
	if err != nil {
		fmt.Printf("an error occurs when run cd %v\n", err)
	}
}
func EditPrivilege(cli *goph.Client, filePath string) {
	cmd := "chmod a+rw " + filePath
	_, err := cli.Run(cmd)
	if err != nil {
		fmt.Printf("an error occurs when run chmod %v\n", err)
	}
}

//	func CreateSlurmScript(cli *goph.Client, config defines.SlurmConfig, jobScriptPath string) {
//		template := `#!/bin/bash
//
// #SBATCH --job-name=%s
// #SBATCH --partition=%s
// #SBATCH --output=%s
// #SBATCH --error=%s
// #SBATCH -N %d
// #SBATCH --ntasks-per-node=%d
// #SBATCH --cpus-per-task=%d
// #SBATCH --gres=gpu:%d
//
// %s
// `
//
//	script := fmt.Sprintf(
//		template,
//		config.JobName,
//		config.Partition,
//		config.Output,
//		config.Error,
//		config.NodeNum,
//		config.NTaskPerNode,
//		config.CPUsPerTask,
//		config.GPUNum,
//	)
//
// }
func GetJobName(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("can't open file:", err)
		return ""
	}
	defer file.Close()

	// 创建一个Scanner以逐行读取文件内容
	scanner := bufio.NewScanner(file)

	// 遍历每一行查找JobName
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#SBATCH --job-name=") {
			// 提取JobName
			jobName := strings.TrimPrefix(line, "#SBATCH --job-name=")
			//fmt.Println("JobName:", jobName)
			return jobName
		}
	}

	// 如果没有找到JobName，则输出提示
	fmt.Println("can't find JobName")
	return ""
}

func GetJobState(cli *goph.Client) string {
	cmd := "sacct --format=state -j " + jobID
	respBytes, err := cli.Run(cmd)
	if err != nil {
		fmt.Printf("an error occurs when get job state %v\n", err)
	}
	resp := string(respBytes)
	rows := strings.Split(resp, "\n")
	if len(rows) > 0 {
		state := strings.ReplaceAll(rows[2], " ", "")
		return state
	}
	return ""
}

func SubmitJob(cli *goph.Client, workDir string, jobScriptPath string) {
	cmd := "cd " + workDir + " && sbatch " + jobScriptPath
	respBytes, err := cli.Run(cmd)
	if err != nil {
		fmt.Printf("an error occurs when submit job %v\n", err)
	}
	resp := string(respBytes)
	fmt.Printf("submit job and get response %v", resp)
	_, err1 := fmt.Sscanf(resp, "Submitted batch job %s", &jobID)
	if err != nil {
		fmt.Printf("an error occurs when get jobID %v\n", err1)
	}
}

func Prepare(cli *goph.Client, workDir string) {
	//dataCli := ssh.NewSSHDataClient()
	remoteWorkDir := config.GPUWorkDirPrefix + workDir
	Mkdir(cli, remoteWorkDir)
	//for test
	//ssh.Upload(cli, "/home/os/src/", remoteWorkDir)
	ssh.Upload(cli, defines.ContainerSrcPath, remoteWorkDir)
	ssh.Upload(cli, defines.ContainerCompilePath, remoteWorkDir)
	//CD(cli, remoteWorkDir)

	//EditPrivilege(cli, defines.SlurmScriptName)
}
func WaitJobComplete(cli *goph.Client) {
	state := GetJobState(cli)
	fmt.Printf("job state is %v\n", state)
	if state == "COMPLETED" || state == "FAILED" {
		return
	}
	ticker := time.NewTicker(time.Minute * 5)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			state = GetJobState(cli)
			fmt.Printf("job state is %v\n", state)
			if state == "COMPLETED" || state == "FAILED" {
				return
			}
		}
	}
}

func DownloadResult(cli *goph.Client, workDir string) {
	//for test
	//ssh.Download(cli, config.GPUWorkDirPrefix+workDir+"/"+jobID+".out", "/home/os/result/"+jobID+".out")
	//ssh.Download(cli, config.GPUWorkDirPrefix+workDir+"/"+jobID+".err", "/home/os/result/"+jobID+".err")
	ssh.Download(cli, config.GPUWorkDirPrefix+workDir+"/"+jobID+".out", defines.ContainerResultPath+jobID+".out")
	ssh.Download(cli, config.GPUWorkDirPrefix+workDir+"/"+jobID+".err", defines.ContainerResultPath+jobID+".err")
}

// func CompleteSlurmScripts(slurmPath string) {
//
//	contentBytes, err := os.ReadFile(slurmPath)
//	if err != nil {
//		fmt.Println("Error reading slurmScript:", err)
//		os.Exit(1)
//	}
//
//	content := string(contentBytes)
//	content += "\nmodule load cuda/10.0.130-gcc-4.8.5\n"
//	content += "\nmake run\n"
//
//	file, _ := os.OpenFile(defines.ContainerSrcPath+"job.slurm", os.O_CREATE|os.O_RDWR, 0777)
//	defer file.Close()
//	if _, err1 := file.WriteString(content); err1 != nil {
//		fmt.Println("can't write file", err1)
//	}
//
// }

func Run(cli *goph.Client, workDir string) {
	Prepare(cli, workDir)
	//compilePath := config.GPUWorkDirPrefix + workDir + "/compile/" + config.GPUCompileFileName
	//Compile(cli, compilePath)
	SubmitJob(cli, workDir, defines.SlurmScriptName)
	WaitJobComplete(cli)
	DownloadResult(cli, workDir)
}

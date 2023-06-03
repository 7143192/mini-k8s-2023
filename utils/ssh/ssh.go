package ssh

import (
	"fmt"
	"github.com/melbahja/goph"
	"os"
	"os/exec"
)

const (
	username     = "stu1643"
	password     = "C4wYgg*Y"
	sshLoginAddr = "pilogin.hpc.sjtu.edu.cn"
	sshDataAddr  = "data.hpc.sjtu.edu.cn"
)

func NewSSHClient() *goph.Client {

	cli, err := goph.NewUnknown(username, sshLoginAddr, goph.Password(password))
	if err != nil {
		fmt.Printf("An error occurs when creating new ssh client %v\n", err)
		return nil
	}
	return cli
	//auth, err := goph.Key("/home/os/.ssh/id_rsa", "")
	//if err != nil {
	//	fmt.Printf("An error occurs when creating new auth %v\n", err)
	//	return nil
	//}
	//client, err1 := goph.New(username, sshLoginAddr, auth)
	//if err1 != nil {
	//	fmt.Printf("An error occurs when creating new ssh client %v\n", err1)
	//	return nil
	//}
	//return client
}

func NewSSHDataClient() *goph.Client {
	cli, err := goph.NewUnknown(username, sshDataAddr, goph.Password(password))
	if err != nil {
		fmt.Printf("An error occurs when creating new data ssh client %v\n", err)
		return nil
	}
	return cli
}

func Upload(dataCli *goph.Client, localPath string, remotePath string) {
	files, err := os.ReadDir(localPath)
	if err != nil {
		fmt.Println("an error occurs when read dir", err)
	}

	//remoteAddr := fmt.Sprintf("%s@%s:%s", username, sshDataAddr, remotePath)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		//cmd := exec.Command("scp", "-r", localPath+file.Name(), remoteAddr)
		//err1 := cmd.Run()
		//if err1 != nil {
		//	fmt.Printf("an error occurs when exec scp upload, %v\n", err1)
		//}
		//fmt.Printf("local path: %v, remote path： %v", localPath+file.Name(), remotePath+"/"+file.Name())
		err1 := dataCli.Upload(localPath+file.Name(), remotePath+"/"+file.Name())
		if err1 != nil {
			fmt.Printf("an error occurs when upload file%v\n", err1)
		}
	}
}

func Download(dataCli *goph.Client, remotePath string, localPath string) {
	//remoteAddr := fmt.Sprintf("%s@%s:%s", username, sshDataAddr, remotePath)
	//cmd := exec.Command("scp", remoteAddr, localPath)
	//err := cmd.Run()
	//if err != nil {
	//	fmt.Printf("an error occurs when exec scp download, %v\n", err)
	//}
	err1 := dataCli.Download(remotePath, localPath)
	if err1 != nil {
		fmt.Printf("an error occurs when download file%v\n", err1)
	}
}
func Rsync(username string, localPath string, remotePath string) {
	remoteAddr := fmt.Sprintf("%s@%s:%s", username, sshDataAddr, remotePath)
	cmd := exec.Command("rsync", "--archive", "--partial", "--progress", remoteAddr, localPath)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("an error occurs when exec rsync, %v\n", err)
	}
}

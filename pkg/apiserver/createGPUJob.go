package apiserver

import (
	uuid "github.com/satori/go.uuid"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v3"
	"log"
	"mini-k8s/pkg/container"
	"mini-k8s/pkg/defines"
	"mini-k8s/pkg/etcd"
	"os"
	"strconv"
)

// GenerateSlurmScript generate slurm script file from job info.
func GenerateSlurmScript(gpuJob *defines.GPUJob) string {
	content := "#!/bin/bash\n\n"
	// job name.
	content = content + "#SBATCH --job-name=" + gpuJob.SlurmConfig.JobName + "\n"
	// partition queue name.
	content = content + "#SBATCH --partition=" + gpuJob.SlurmConfig.Partition + "\n"
	// total core number.
	coreNumStr := strconv.Itoa(gpuJob.SlurmConfig.CoreNum)
	content = content + "#SBATCH -n " + coreNumStr + "\n"
	// tasks per node number.
	taskPerNodeStr := strconv.Itoa(gpuJob.SlurmConfig.NTaskPerNode)
	content = content + "#SBATCH --ntasks-per-node=" + taskPerNodeStr + "\n"
	// node number.
	nodeStr := strconv.Itoa(gpuJob.SlurmConfig.NodeNum)
	content = content + "#SBATCH -N " + nodeStr + "\n"
	// cpus per task number.
	cpuPerTaskStr := strconv.Itoa(gpuJob.SlurmConfig.CPUsPerTask)
	content = content + "#SBATCH --cpus-per-task=" + cpuPerTaskStr + "\n"
	// GPU number.
	gpuNumStr := strconv.Itoa(gpuJob.SlurmConfig.GPUNum)
	content = content + "#SBATCH --gres=gpu:" + gpuNumStr + "\n"
	content = content + "#SBATCH --output=%j.out\n"
	content = content + "#SBATCH --error=%j.err\n"
	content = content + "\nmodule load cuda/10.0.130-gcc-4.8.5\n"
	content = content + "\nmake run\n"
	// then write the generated slurm configuration file to local file.
	file, _ := os.OpenFile(defines.SlurmFileLocalPath, os.O_CREATE|os.O_RDWR, 0777)
	defer file.Close()
	_, _ = file.WriteString(content)
	return content
}

func GenerateGPUJobImage(gpuJob *defines.GPUJob) {
	// generate slurm script first.
	GenerateSlurmScript(gpuJob)
	// then copy the file into container.
	con := &defines.PodContainer{}
	con.Image = defines.GPUJobBasicImageName
	conName := defines.GPUJobContainerNamePrefix + "_" + gpuJob.Name
	id := string(container.CreateGPUJobContainer(conName, con))
	// a random new image name.
	imageName := defines.GPUJobBasicImageName + gpuJob.Name + "-" + uuid.NewV4().String()
	gpuJob.ImageName = imageName
	_ = container.CopyToContainer(id, gpuJob.SourcePath, defines.ContainerSrcPath)
	_ = container.CopyToContainer(id, gpuJob.CompilePath, defines.ContainerCompilePath)
	_ = container.CopyToContainer(id, defines.SlurmFileLocalPath, defines.ContainerSrcPath)
	// then commit the new container and image.
	newId := container.CommitContainer(id)
	imageID := newId[7:]
	container.TagNewImage(imageID, imageName)
	container.PushImage(imageName)
	// then remove the container.
	container.RemoveForceContainer(id)
}

//// MakeNewHostResultDir NOTE: add this function to check hostResultDir status.
//func MakeNewHostResultDir(name string) error {
//	// cd to /home dir first.
//	//cmd0 := exec.Command("cd", "~")
//	//err := cmd0.Run()
//	err := os.Chdir("/home")
//	if err != nil {
//		log.Printf("fail to cd to /home dir: %v\n", err)
//		return err
//	}
//	// then use ls to check whether the rootResultDir has already existed.
//	cmd1 := exec.Command("ls")
//	output, err := cmd1.Output()
//	if err != nil {
//		log.Printf("fail to ls !\n")
//		return err
//	}
//	got1 := string(output)
//	if strings.Contains(got1, "gpuResults") == false {
//		cmd2 := exec.Command("mkdir", "gpuResults")
//		err = cmd2.Run()
//		if err != nil {
//			log.Printf("fail to mkdir gpuResults!\n")
//			return err
//		}
//	}
//	// then cd to gpuResults dir.
//	//cmd3 := exec.Command("cd", "gpuResults")
//	//err = cmd3.Run()
//	err = os.Chdir("gpuResults")
//	if err != nil {
//		log.Printf("fail to cd to gpuResults dir!\n")
//		return err
//	}
//	// every should have a private dir to contain its own results and errors.
//	dirName := name
//	// ls again to check existence.
//	cmd4 := exec.Command("ls")
//	output1, err := cmd4.Output()
//	if err != nil {
//		log.Printf("fail to ls in dir gpuResults!\n")
//		return err
//	}
//	got2 := string(output1)
//	if strings.Contains(got2, dirName) == true {
//		// if exists....
//		//cmd5 := exec.Command("cd", dirName)
//		//err = cmd5.Run()
//		err = os.Chdir(dirName)
//		if err != nil {
//			log.Printf("fail to cd to job private dir !\n")
//			return err
//		}
//		// then remove everything under this job-private directory.
//		cmd6 := exec.Command("rm", "-rf", "./*")
//		err = cmd6.Run()
//		if err != nil {
//			log.Printf("fail to clean every thing under a job-pivate dir!\n")
//			return err
//		}
//	} else {
//		// if not exists...
//		cmd7 := exec.Command("mkdir", dirName)
//		err = cmd7.Run()
//		if err != nil {
//			log.Printf("fail to mkdir for job %v!\n", name)
//			return err
//		}
//	}
//	return nil
//}

func GenerateGPUJobYamlPod(gpuJob *defines.GPUJob) *defines.YamlPod {
	res := &defines.YamlPod{}
	res.ApiVersion = "v1"
	res.Kind = "Pod"
	res.Metadata.Name = defines.GPUJobPodNamePrefix + "_" + gpuJob.Name
	res.Metadata.Label.App = gpuJob.Name
	res.Metadata.Label.Env = "product"
	// main part: generate the spec of this pod.

	// generate a new image first.
	GenerateGPUJobImage(gpuJob)

	spec := defines.PodSpec{}
	spec.Volumes = make([]defines.PodVolume, 0)

	// NOTE: change here !
	// add one volume here to act as the hostMountPath.
	vol := defines.PodVolume{}
	vol.Name = defines.VolumeNamePrefix + gpuJob.Name // "volume_JOBNAME"
	vol.HostPath = "/home/gpuResults/" + gpuJob.Name
	spec.Volumes = append(spec.Volumes, vol)
	//// create the corresponding directory.
	//_ = MakeNewHostResultDir(gpuJob)

	// only need one normal container.
	spec.Containers = make([]defines.PodContainer, 0)
	jobContainer := defines.PodContainer{}
	jobContainer.Name = defines.GPUJobContainerNamePrefix + "_" + gpuJob.Name
	jobContainer.VolumeMounts = make([]defines.PodVolumeMount, 0)

	// NOTE: change here!
	// add a volumeMounts here to make mount point.
	volMount := defines.PodVolumeMount{}
	volMount.Name = defines.VolumeNamePrefix + gpuJob.Name // "volume_JOBNAME", the same as the volume name.
	volMount.MountPath = defines.ContainerResultPath
	jobContainer.VolumeMounts = append(jobContainer.VolumeMounts, volMount)

	jobContainer.WorkingDir = ""
	// generate resource info.
	resourceInfo := defines.PodResource{}
	resourceInfo.ResourceLimit = defines.PodResourceLimit{}
	resourceInfo.ResourceLimit.Memory = "128MB"
	resourceInfo.ResourceLimit.Cpu = "100m"
	resourceInfo.ResourceRequest = defines.PodResourceRequest{}
	resourceInfo.ResourceRequest.Memory = "128MB"
	resourceInfo.ResourceRequest.Cpu = "100m"
	jobContainer.Resource = resourceInfo
	// generate pod ports info.
	portsInfo := make([]defines.PodPort, 0)
	port0 := defines.PodPort{}
	port0.ContainerPort = 80
	port0.Name = defines.JobPortName1
	port0.Protocol = "TCP"
	port1 := defines.PodPort{}
	port1.ContainerPort = 22
	port1.Name = defines.JobPortName2
	port1.Protocol = "TCP"
	portsInfo = append(portsInfo, port0)
	portsInfo = append(portsInfo, port1)
	jobContainer.Ports = portsInfo
	// TODO: what command and args are required?
	jobContainer.Args = make([]string, 0)
	jobContainer.Command = make([]string, 0)
	// TODO: what image to use for this cuda program??
	jobContainer.Image = gpuJob.ImageName
	spec.Containers = append(spec.Containers, jobContainer)
	res.Spec = spec
	return res
}

func CreateGPUJob(cli *clientv3.Client, gpuJob *defines.GPUJob) *defines.EtcdGPUJob {
	// check duplication first.(gpu name can not be the same for two job object. )
	prefixKey := defines.GPUJobPrefix + "/"
	kvs := etcd.GetWithPrefix(cli, prefixKey).Kvs
	for _, kv := range kvs {
		tmp := &defines.EtcdGPUJob{}
		_ = yaml.Unmarshal(kv.Value, tmp)
		if tmp.JobInfo.Name == gpuJob.Name {
			log.Printf("job %v has already exists in the current system!\n", gpuJob.Name)
			return tmp
		}
	}
	// generate yaml pod info first.
	yamlPod := GenerateGPUJobYamlPod(gpuJob)
	res := &defines.EtcdGPUJob{}
	// init state should be "pending".
	res.JobState = defines.Pending
	res.JobInfo = gpuJob
	res.PodInfo = &defines.Pod{}
	res.PodInfo.YamlPod = *yamlPod
	// then generate a real pod.
	newPod := CreatePod(cli, yamlPod)
	res.PodInfo = newPod
	resByte, _ := yaml.Marshal(res)
	// store the new gpuJob instance into etcd.
	key := defines.GPUJobPrefix + "/" + gpuJob.Name
	etcd.Put(cli, key, string(resByte))
	// then send the gpuJob to slurm to execute.

	// store the new gpuJob instance into all-jobs-list in etcd.
	listKey := defines.GPUJobListPrefix + "/"
	kv := etcd.Get(cli, listKey).Kvs
	names := make([]string, 0)
	if len(kv) == 0 {
		names = append(names, key)
	} else {
		_ = yaml.Unmarshal(kv[0].Value, &names)
		names = append(names, key)
	}
	namesByte, _ := yaml.Marshal(names)
	etcd.Put(cli, listKey, string(namesByte))
	return res
}

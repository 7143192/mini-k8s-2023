package defines

const GPUJobPrefix = "GPUJob"
const GPUJobListPrefix = "GPUJobList"
const GPUJobPodNamePrefix = "GPUJobPod"
const GPUJobContainerNamePrefix = "GPUJobContainer"
const JobPortName1 = "JobPort80"
const JobPortName2 = "JobPort22"
const ContainerSlurmPath = "/home/"
const ContainerNormalPath = "/home/"
const ContainerSrcPath = "/home/src/"
const ContainerCompilePath = "/home/compile/"
const SlurmFileLocalPath = "/home/os/Desktop/job.slurm"
const SlurmScriptName = "job.slurm"
const ContainerResultPath = "/home/result/"
const GPUJobBasicImageName = "7143192/basic"
const DockerUsername = "7143192"
const DockerPwd = "abc125558"
const VolumeNamePrefix = "volume_"

type SlurmConfig struct {
	JobName string `json:"jobName" yaml:"jobName"`
	// queue name. should be "dgx2".
	Partition    string `json:"partition" yaml:"partition"`
	NTaskPerNode int    `json:"NTaskPerNode" yaml:"NTaskPerNode"`
	NodeNum      int    `json:"nodeNum" yaml:"nodeNum"`
	GPUNum       int    `json:"GPUNum" yaml:"GPUNum"`
	CoreNum      int    `json:"coreNum" yaml:"coreNum"`
	CPUsPerTask  int    `json:"CPUsPerTask" yaml:"CPUsPerTask"`
	Output       string `json:"output" yaml:"output"`
	Error        string `json:"error" yaml:"error"`
}

type GPUJob struct {
	Name string `json:"name" yaml:"name"`
	Kind string `json:"kind" yaml:"kind"`
	// cuda source code path.
	SourcePath string `json:"sourcePath" yaml:"sourcePath"`
	// cuda code compile sh path.
	CompilePath string `json:"compilePath" yaml:"compilePath"`
	// result path of one running cuda program.
	ResultPath  string      `json:"resultPath" yaml:"resultPath"`
	ImageName   string      `json:"imageName" yaml:"imageName"`
	SlurmConfig SlurmConfig `json:"slurmConfig" yaml:"slurmConfig"`
}

type EtcdGPUJob struct {
	JobInfo *GPUJob `json:"jobInfo" yaml:"jobInfo"`
	// podInstance created according to jobInfo.
	PodInfo  *Pod `json:"podInfo" yaml:"podInfo"`
	JobState int  `json:"jobState" yaml:"jobState"`
}

type AllJobs struct {
	Jobs []*EtcdGPUJob `json:"jobs" yaml:"jobs"`
}

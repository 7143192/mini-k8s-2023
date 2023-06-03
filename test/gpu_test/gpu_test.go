package gpu_test

import (
	yamlv3 "gopkg.in/yaml.v3"
	"mini-k8s/pkg/defines"
	"mini-k8s/pkg/etcd"
	"mini-k8s/utils/yaml"
	"testing"
)

func TestGPUJob(t *testing.T) {
	job, _ := yaml.ParseGPUJobConfig("../../utils/templates/gpu_template.yaml")
	job1, _ := yaml.ParseGPUJobConfig("../../utils/templates/gpu_template_1.yaml")
	res1 := &defines.EtcdGPUJob{}
	res2 := &defines.EtcdGPUJob{}
	res1.JobInfo = job
	res1.PodInfo = &defines.Pod{}
	res1.JobState = defines.Pending
	res2.JobInfo = job1
	res2.PodInfo = &defines.Pod{}
	res2.JobState = defines.Pending
	cli := etcd.EtcdStart()
	defer cli.Close()
	key := defines.GPUJobPrefix + "/" + job.Name
	key1 := defines.GPUJobPrefix + "/" + job1.Name
	resByte, _ := yamlv3.Marshal(res1)
	resByte1, _ := yamlv3.Marshal(res2)
	etcd.Put(cli, key, string(resByte))
	etcd.Put(cli, key1, string(resByte1))
}

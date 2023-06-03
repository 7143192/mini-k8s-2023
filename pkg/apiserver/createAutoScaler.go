package apiserver

import (
	"bytes"
	"encoding/json"
	"errors"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v3"
	"log"
	"math/rand"
	"mini-k8s/pkg/config"
	"mini-k8s/pkg/defines"
	"mini-k8s/pkg/etcd"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func GetAutoTotalCPU(auto *defines.AutoScaler) uint64 {
	res := uint64(0)
	for _, con := range auto.AutoSpec.Containers {
		cpu := con.Resource.ResourceLimit.Cpu
		cpuLen := len(cpu)
		cpuLimit := 0
		if cpu[cpuLen-1] == 'm' {
			// micro-CPU
			val := cpu[0 : cpuLen-1]
			gotVal, _ := strconv.Atoi(val)
			cpuLimit = gotVal * 1e6
		} else {
			// 0.XX
			if cpu[0] == '0' {
				idx := strings.Index(cpu, ".")
				intCpu := strings.Replace(cpu, ".", "", -1)
				gotIntCpu, _ := strconv.Atoi(intCpu)
				cpuLimit = gotIntCpu * 10e9
				backLen := cpuLen - idx - 1
				for i := 0; i < backLen; i++ {
					cpuLimit = cpuLimit / 10
				}
			} else {
				// a.XX (a != 0)
				newCPU := strings.Replace(cpu, ".", "", -1)
				gotCpu, _ := strconv.Atoi(newCPU)
				cpuLimit = gotCpu * 1e7
			}
		}
		res = res + uint64(cpuLimit)
	}
	return res
}

func GenerateEtcdAutoScaler(auto *defines.AutoScaler) *defines.EtcdAutoScaler {
	res := &defines.EtcdAutoScaler{}
	res.AutoName = auto.Metadata.AutoName
	podName := defines.AutoPodNamePrefix + auto.Metadata.AutoName
	res.AutoPodId = podName
	res.MinReplicas = auto.MinReplicas
	res.MaxReplicas = auto.MaxReplicas
	totalCPU := GetAutoTotalCPU(auto)
	log.Printf("[debug]: totalCPU: %v\n", totalCPU)
	if len(auto.Metrics.Resources) == 2 {
		//// idx 0 is cpu, idx 1 is memory.
		//minCPU, _ := strconv.Atoi(auto.Metrics.Resources[0].Min)
		//maxCPU, _ := strconv.Atoi(auto.Metrics.Resources[0].Max)
		//res.MinCPU = uint64(minCPU)
		//res.MaxCPU = uint64(maxCPU)
		minCPUStr := auto.Metrics.Resources[0].Min[0 : len(auto.Metrics.Resources[0].Min)-1]
		maxCPUStr := auto.Metrics.Resources[0].Max[0 : len(auto.Metrics.Resources[0].Max)-1]
		minCPUPer, _ := strconv.Atoi(minCPUStr)
		maxCPUPer, _ := strconv.Atoi(maxCPUStr)
		res.MinCPU = uint64(minCPUPer) * totalCPU / uint64(100)
		res.MaxCPU = uint64(maxCPUPer) * totalCPU / uint64(100)
		res.MinMem = auto.Metrics.Resources[1].Min
		res.MaxMem = auto.Metrics.Resources[1].Max
	} else {
		if len(auto.Metrics.Resources) == 1 {
			if auto.Metrics.Resources[0].Name == "cpu" {
				//// cpu is configured.
				//minCPU, _ := strconv.Atoi(auto.Metrics.Resources[0].Min)
				//maxCPU, _ := strconv.Atoi(auto.Metrics.Resources[0].Max)
				//res.MinCPU = uint64(minCPU)
				//res.MaxCPU = uint64(maxCPU)
				minCPUStr := auto.Metrics.Resources[0].Min[0 : len(auto.Metrics.Resources[0].Min)-1]
				maxCPUStr := auto.Metrics.Resources[0].Max[0 : len(auto.Metrics.Resources[0].Max)-1]
				minCPUPer, _ := strconv.Atoi(minCPUStr)
				maxCPUPer, _ := strconv.Atoi(maxCPUStr)
				res.MinCPU = uint64(minCPUPer) * totalCPU / uint64(100)
				res.MaxCPU = uint64(maxCPUPer) * totalCPU / uint64(100)
				// some default values.
				res.MinMem = "5MB"
				res.MaxMem = "1GB"
			}
			if auto.Metrics.Resources[0].Name == "memory" {
				// memory is configured.
				res.MinMem = auto.Metrics.Resources[0].Min
				res.MaxMem = auto.Metrics.Resources[0].Max
				// some default values.
				res.MinCPU = 0
				res.MaxCPU = 1000000
			}
		}
	}
	res.StartTime = time.Now()
	return res
}

func SendOutAddAutoScalerRequest(etcdAuto *defines.EtcdAutoScaler) {
	body, err := json.Marshal(etcdAuto)
	if err != nil {
		return
	}
	url := "http://" + config.MasterIP + ":" + config.AutoScalerPort + "/objectAPI/addAutoScaler"
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return
	}
	request.Header.Add("Content-Type", "application/json")
	reponse, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Printf("[apiserver] an error occurs when adding an autoScaler into list!")
		return
	}
	if reponse.StatusCode == http.StatusOK {
		log.Println("[apiserver] add a new AutoScaler to controller successfully!")
	} else {
		log.Println("[apiserver] an error occurs when adding a new AutoScaler to controller!")
	}
	return
}

func GenerateRandStr(n int) string {
	str := strings.Builder{}
	str.Grow(n)
	r := rand.New(rand.NewSource(time.Now().UnixMicro()))
	for i := 0; i < n; i++ {
		str.WriteByte(config.CharSet[r.Intn(len(config.CharSet))])
	}
	return str.String()
}

func CreateAutoPodReplicas(cli *clientv3.Client, yamlPod *defines.YamlPod, replicaNum int) error {
	initPodName := yamlPod.Metadata.Name
	//initHostPort := make([]int, 0)
	//// collect init host ports information here.
	//// TODO: in version 2.0, we change hostPort for every replica to simplify the single-node test.
	//curNames := make([]string, 0)
	//namesKey := defines.AutoPodReplicasNamesPrefix + "/" + initPodName
	//namesKV := etcd.Get(cli, namesKey).Kvs
	//if len(namesKV) != 0 {
	//	_ = yaml.Unmarshal(namesKV[0].Value, &curNames)
	//}
	//for _, con := range yamlPod.Spec.Containers {
	//	initHostPort = append(initHostPort, con.Ports[0].HostPort+len(curNames))
	//}
	for i := 0; i < replicaNum; i++ {
		tmp := &defines.AutoCreateReplicaSend{}
		tmp.PodName = initPodName
		// NOTE: change name format here!
		nameRandStr := GenerateRandStr(5)
		// podName := initPodName + "_" + uuid.NewV4().String()
		podName := initPodName + "_" + nameRandStr
		yamlPod.Metadata.Name = podName
		//// update port number for every replica.
		//for idx, _ := range yamlPod.Spec.Containers {
		//	yamlPod.Spec.Containers[idx].Ports[0].HostPort = initHostPort[idx] + i
		//}
		tmp.YamlInfo = yamlPod
		log.Printf("[apiserver] new yaml pod info for one replica : %v\n", yamlPod)
		body, _ := json.Marshal(tmp)
		url := "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/addAutoPodReplica"
		log.Printf("[apiserver] add auto pod replica request = %v\n", url)
		request, err := http.NewRequest("POST", url, bytes.NewReader(body))
		if err != nil {
			return err
		}
		request.Header.Add("Content-Type", "application/json")
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			return err
		}
		if response.StatusCode != http.StatusOK {
			return errors.New("[apiserver] fail to create new auto pod replica")
		}
	}
	// restore yamlInfo.
	yamlPod.Metadata.Name = initPodName
	//for idx, _ := range yamlPod.Spec.Containers {
	//	yamlPod.Spec.Containers[idx].Ports[0].HostPort = initHostPort[idx]
	//}
	return nil
}

func DeleteAutoPodReplica(podName string, autoName string) error {
	nameSend := &defines.AutoNameSend{}
	nameSend.PodName = podName
	nameSend.AutoName = autoName
	body, _ := json.Marshal(nameSend)
	url := "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/delAutoPodReplica"
	log.Printf("[apiserver] delete auto pod replica request = %v\n", url)
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return errors.New("[apiserver] fail to create new auto pod replica")
	}
	return nil
}

func CreateAutoScaler(cli *clientv3.Client, auto *defines.AutoScaler) *defines.EtcdAutoScaler {
	// check duplicate autoScaler here.
	prefixKey := defines.AutoScalerPrefix + "/"
	allKv := etcd.GetWithPrefix(cli, prefixKey).Kvs
	if len(allKv) != 0 {
		// check duplication here.
		for _, singleKv := range allKv {
			tmp := &defines.EtcdAutoScaler{}
			_ = yaml.Unmarshal(singleKv.Value, tmp)
			if tmp.AutoName == auto.Metadata.AutoName {
				log.Printf("[apiserver] AutoScaler %v has already been created!\n", auto.Metadata.AutoName)
				// return old obj directly.
				return tmp
			}
		}
	}
	// when targetKind is Pod:
	if auto.TargetKind == "pod" {
		podName := defines.AutoPodNamePrefix + auto.Metadata.AutoName
		// generate etcd autoScaler object.
		etcdAuto := GenerateEtcdAutoScaler(auto)
		// store this object into etcd.
		etcdAutoKey := defines.AutoScalerPrefix + "/" + auto.Metadata.AutoName
		etcdAutoByte, _ := yaml.Marshal(etcdAuto)
		etcd.Put(cli, etcdAutoKey, string(etcdAutoByte))
		// create a new yamlPod according to config information.
		yamlPod := &defines.YamlPod{}
		yamlPod.ApiVersion = auto.ApiVersion
		yamlPod.Kind = "Pod"
		yamlPod.Metadata.Name = podName
		yamlPod.Metadata.Label.App = auto.Label.App
		yamlPod.Metadata.Label.Env = auto.Label.Env
		// TODO:in version 2.0, we just write a CONST here as we do not begin to use this param to do schedule.
		yamlPod.NodeSelector.Gpu = "nvidia"
		yamlPod.Spec = auto.AutoSpec
		// then store this new yamlPod into etcd.
		autoPodKey := defines.AutoScalerYamlPodPrefix + "/" + podName
		yamlPodByte, _ := yaml.Marshal(yamlPod)
		etcd.Put(cli, autoPodKey, string(yamlPodByte))

		// then create the init replica pods set.(according to the MinReplicas)
		_ = CreateAutoPodReplicas(cli, yamlPod, auto.MinReplicas)

		// then update all-auto-scalers-list.
		listKey := defines.AutoScalerListPrefix + "/"
		kv := etcd.Get(cli, listKey).Kvs
		if len(kv) == 0 {
			// the first autoScaler case.
			newList := make([]string, 0)
			newList = append(newList, etcdAuto.AutoName)
			newListByte, _ := yaml.Marshal(&newList)
			etcd.Put(cli, listKey, string(newListByte))
		} else {
			// put into directly here, as we have already checked the duplication of this autoScaler.
			oldList := make([]string, 0)
			_ = yaml.Unmarshal(kv[0].Value, &oldList)
			newList := append(oldList, etcdAuto.AutoName)
			newListByte, _ := yaml.Marshal(&newList)
			etcd.Put(cli, listKey, string(newListByte))
		}
		// then send HTTP to autoScaler controller server.
		SendOutAddAutoScalerRequest(etcdAuto)
		return etcdAuto
	}
	return nil
}

package main

import (
	"github.com/gin-gonic/gin"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
	"mini-k8s/pkg/config"
	"os"
	"path/filepath"
)

type Server struct {
	engine *gin.Engine
	cli    *clientv3.Client
}

//func main() {
//	s := &Server{
//		engine: gin.Default(),
//		cli:    etcd.EtcdStart(),
//	}
//
//	s.engine.POST("/objectAPI/createReplicaSet", func(context *gin.Context) {
//		rs := &defines.ReplicaSet{}
//		err := json.NewDecoder(context.Request.Body).Decode(rs)
//		if err != nil {
//			fmt.Printf("%v\n", err)
//			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Wrong yaml file when create replicaSet!"})
//			return
//		}
//		err = apiserver.CreateReplicaSet(s.cli, rs)
//		if err != nil {
//			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": err.Error()})
//		} else {
//			context.JSON(http.StatusOK, gin.H{"INFO": "The replicaSet has been created successfully!"})
//		}
//	})
//
//	s.engine.POST("/objectAPI/deleteReplicaSet", func(context *gin.Context) {
//		rsName := ""
//		err := json.NewDecoder(context.Request.Body).Decode(&rsName)
//		if err != nil {
//			fmt.Printf("%v\n", err)
//			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Wrong replicaSet name when delete one!"})
//			return
//		}
//		err = apiserver.DeleteReplicaSet(s.cli, rsName)
//		if err != nil {
//			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": err.Error()})
//		} else {
//			context.JSON(http.StatusOK, gin.H{"INFO": "The replicaSet has been deleted successfully!"})
//		}
//	})
//
//	s.engine.GET("/objectAPI/getPods", func(context *gin.Context) {
//		pods := make([]*defines.Pod, 0)
//		Kvs := etcd.GetWithPrefix(s.cli, "PodInstance/").Kvs
//		for _, kv := range Kvs {
//			pod := &defines.Pod{}
//			err := yaml.Unmarshal(kv.Value, pod)
//			if err != nil {
//				log.Printf("[ERROR] %v\n", err)
//				continue
//			}
//			pods = append(pods, pod)
//		}
//		podsVal, _ := json.Marshal(pods)
//		context.JSON(http.StatusOK, gin.H{"INFO": podsVal})
//	})
//
//	s.engine.POST("/objectAPI/createPod", func(context *gin.Context) {
//		yamlPod := &defines.YamlPod{}
//		err := yaml.NewDecoder(context.Request.Body).Decode(yamlPod)
//		if err != nil {
//			log.Printf("[ERROR] %v\n", err)
//			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Wrong yaml format!"})
//			return
//		}
//		//if 7070 <= yamlPod.Spec.Containers[1].Ports[0].HostPort && yamlPod.Spec.Containers[1].Ports[0].HostPort <= 15000 {
//		//	context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "The port has been used!"})
//		//	return
//		//}
//		pod := &defines.Pod{
//			YamlPod:  *yamlPod,
//			PodState: defines.Running,
//		}
//		podVal, err := yaml.Marshal(pod)
//		if err != nil {
//			log.Printf("[ERROR] %v\n", err)
//			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Wrong yaml marshal!"})
//			return
//		}
//		KVs := etcd.Get(s.cli, "PodInstance/"+pod.Metadata.Name).Kvs
//		if len(KVs) != 0 {
//			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "The pod has been created!"})
//			return
//		}
//		etcd.Put(s.cli, "PodInstance/"+pod.Metadata.Name, string(podVal))
//		context.JSON(http.StatusOK, gin.H{"INFO": "Successfully!"})
//	})
//
//	s.engine.POST("/objectAPI/deletePod", func(context *gin.Context) {
//		podName := ""
//		err := json.NewDecoder(context.Request.Body).Decode(&podName)
//		if err != nil {
//			log.Printf("[ERROR] %v\n", err)
//			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Wrong pod name!"})
//			return
//		}
//		etcd.Del(s.cli, "PodInstance/"+podName)
//		context.JSON(http.StatusOK, gin.H{"INFO": "Delete pod successfully!"})
//	})
//
//	s.engine.POST("/objectAPI/changeStatus", func(context *gin.Context) {
//		podName := ""
//		err := json.NewDecoder(context.Request.Body).Decode(&podName)
//		if err != nil {
//			log.Printf("[ERROR] %v\n", err)
//			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Wrong pod name!"})
//			return
//		}
//		kvs := etcd.Get(s.cli, "PodInstance/"+podName).Kvs
//		if len(kvs) == 0 {
//			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Wrong pod name!"})
//			return
//		}
//		pod := &defines.Pod{}
//		err = yaml.Unmarshal(kvs[0].Value, pod)
//		if err != nil {
//			log.Printf("[ERROR] %v\n", err)
//			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Wrong pod struct!"})
//			return
//		}
//		pod.PodState = defines.Failed
//		podVal, err := yaml.Marshal(pod)
//		if err != nil {
//			log.Printf("[ERROR] %v\n", err)
//			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Wrong pod marshal!"})
//			return
//		}
//		etcd.Put(s.cli, "PodInstance/"+podName, string(podVal))
//		context.JSON(http.StatusOK, gin.H{"INFO": "Change pod successfully!"})
//	})
//
//	s.engine.POST("/objectAPI/getReplicaSet", func(context *gin.Context) {
//		rss, err := apiserver.GetReplicaSets(s.cli)
//		if err != nil {
//			context.JSON(http.StatusBadRequest, gin.H{"ERROR": err.Error()})
//		} else {
//			context.JSON(http.StatusOK, gin.H{"INFO": rss})
//		}
//	})
//
//	s.engine.POST("/objectAPI/describeReplicaSet", func(context *gin.Context) {
//		rsName := ""
//		err := json.NewDecoder(context.Request.Body).Decode(&rsName)
//		if err != nil {
//			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": err.Error()})
//			return
//		}
//		rsInfo, err := apiserver.DesReplicaSet(s.cli, rsName)
//		if err != nil {
//			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": err.Error()})
//		} else {
//			if err != nil {
//				context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": err.Error()})
//				return
//			}
//			context.JSON(http.StatusOK, gin.H{"INFO": rsInfo})
//		}
//	})
//
//	s.engine.GET("/objectAPI/getUnhandledPods", func(context *gin.Context) {
//		yamlPods := make([]*defines.YamlPod, 0)
//		kvs := etcd.GetWithPrefix(s.cli, defines.PodInstancePrefix+"/").Kvs
//		for _, kv := range kvs {
//			p := &defines.Pod{}
//			err := yaml.Unmarshal(kv.Value, p)
//			if err != nil {
//				log.Printf("[ERROR] %v\n", err)
//				continue
//			}
//			if p.NodeId == "" {
//				yamlPods = append(yamlPods, &p.YamlPod)
//			}
//		}
//
//		yamlPodsInfo, err := json.Marshal(yamlPods)
//		if err != nil {
//			log.Printf("[ERROR] %v\n", err)
//			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Something went wrong in wrapping yamlPods!"})
//		} else {
//			context.JSON(http.StatusOK, gin.H{"INFO": yamlPodsInfo})
//		}
//	})
//
//	s.engine.POST("/objectAPI/updatePod", func(context *gin.Context) {
//		result := make(map[string]string)
//		err := json.NewDecoder(context.Request.Body).Decode(&result)
//		if err != nil {
//			fmt.Printf("%v\n", err)
//			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Wrong pod!"})
//			return
//		}
//		pods := etcd.Get(s.cli, defines.PodInstancePrefix+"/"+result["podName"]).Kvs
//		if len(pods) == 0 {
//			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "The pod doesn't exist anymore!"})
//			return
//		}
//		podInfo := &defines.Pod{}
//		err = yaml.Unmarshal(pods[0].Value, podInfo)
//		if err != nil {
//			fmt.Printf("%v\n", err)
//			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Error in saving pod config!"})
//			return
//		}
//		podInfo.NodeId = result["nodeID"]
//		podVal, err := yaml.Marshal(podInfo)
//		if err != nil {
//			fmt.Printf("%v\n", err)
//			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Error in saving pod config!"})
//			return
//		}
//		etcd.Put(s.cli, defines.PodInstancePrefix+"/"+podInfo.Metadata.Name, string(podVal))
//		context.JSON(http.StatusOK, gin.H{"INFO": "The pod has been updated successfully!"})
//	})
//
//	s.engine.GET("/objectAPI/getAllReplicaSets", func(context *gin.Context) {
//		replicaSets := make([]*defines.ReplicaSet, 0)
//		KVS := etcd.GetWithPrefix(s.cli, defines.RSInstancePrefix+"/").Kvs
//		for _, kv := range KVS {
//			replicaSet := &defines.ReplicaSet{}
//			err := yaml.Unmarshal(kv.Value, replicaSet)
//			if err != nil {
//				log.Printf("[ERROR] %v\n", err)
//				etcd.Del(s.cli, string(kv.Key))
//				continue
//			}
//			replicaSets = append(replicaSets, replicaSet)
//		}
//
//		replicaSetsInfo, err := json.Marshal(replicaSets)
//		if err != nil {
//			log.Printf("[ERROR] %v\n", err)
//			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Something went wrong in marshal result!"})
//		} else {
//			context.JSON(http.StatusOK, gin.H{"INFO": replicaSetsInfo})
//		}
//	})
//
//	s.engine.POST("/objectAPI/addFunction", func(context *gin.Context) {
//		file, err := context.FormFile("file")
//		if err != nil {
//			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "can't parsing the uploaded file!"})
//			return
//		}
//
//		_, err = os.Stat(filepath.Join(config.ServerlessFileDir, file.Filename[0:len(file.Filename)-4]))
//		if !os.IsNotExist(err) {
//			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "The function has existed!"})
//			return
//		}
//
//		filePath := filepath.Join(config.ServerlessTmpFileDir, file.Filename)
//		err = context.SaveUploadedFile(file, filePath)
//		if err != nil {
//			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "can't save the uploaded file!"})
//			return
//		}
//		defer func(name string) {
//			_ = os.Remove(name)
//		}(filePath)
//
//		err = utils.UnzipFile(filePath, config.ServerlessFileDir)
//		if err != nil {
//			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "can't unzip the uploaded file!"})
//			return
//		}
//
//		url := "http://" + config.MasterIP + ":" + config.SLPort + "/objectAPI/addFunction"
//		body, _ := json.Marshal(file.Filename[0 : len(file.Filename)-4])
//		request, err := http.NewRequest("POST", url, bytes.NewReader(body))
//		if err != nil {
//			log.Printf("[ERROR] %v\n", err)
//			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Something went wrong when communicate with serverless server!"})
//			return
//		}
//		request.Header.Add("Content-Type", "application/json")
//		res, err := http.DefaultClient.Do(request)
//		if err != nil {
//			log.Printf("[ERROR] %v\n", err)
//			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Something went wrong when communicate with serverless server!"})
//			return
//		}
//		result := make(map[string]string)
//		err = json.NewDecoder(res.Body).Decode(&result)
//		if err != nil {
//			log.Printf("[ERROR] %v\n", err)
//			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": "Something went wrong when parsing serverless server res!"})
//			return
//		}
//		if res.StatusCode == http.StatusOK {
//			etcd.Put(s.cli, defines.FunctionPrefix+"/"+file.Filename[0:len(file.Filename)-4], file.Filename[0:len(file.Filename)-4])
//			context.JSON(http.StatusOK, gin.H{"INFO": result["INFO"]})
//		} else {
//			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"ERROR": result["ERROR"]})
//		}
//	})
//
//	err := s.engine.Run()
//	if err != nil {
//		log.Panic(err)
//	}
//}

//func main() {
//	workflow := &defines.WorkFlow{}
//	file, err := os.Open("./utils/templates/workflow_template.yaml")
//
//	if err != nil {
//		log.Printf("[ERROR] %v\n", err)
//		return
//	}
//
//	err = yaml.NewDecoder(file).Decode(workflow)
//
//	if err != nil {
//		log.Printf("[ERROR] %v\n", err)
//		return
//	}
//
//	log.Println(workflow)
//
//}

func main() {
	dirList := make([]string, 0)
	dirErr := filepath.Walk(config.ServerlessFileDir+"/",
		func(path string, f os.FileInfo, err error) error {
			if path == config.ServerlessFileDir+"/" {
				return nil
			}
			if f == nil {
				return err
			}
			if f.IsDir() {
				dirList = append(dirList, f.Name())
				return nil
			}

			return nil
		})

	if dirErr != nil {
		log.Printf("[ERROR] %v\n", dirErr)
		return
	}

	log.Println(dirList)
}

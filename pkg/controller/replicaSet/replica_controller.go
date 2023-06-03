package replicaSet

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"math/rand"
	"mini-k8s/pkg/config"
	"mini-k8s/pkg/defines"
	"mini-k8s/pkg/etcd"
	"mini-k8s/pkg/kubectl/cmds"
	"mini-k8s/utils/client"
	_map "mini-k8s/utils/map"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type Replica struct {
	rs       *defines.ReplicaSet
	pods     map[string]*defines.Pod
	replicas int32
}

type ReplicaController struct {
	setMap     *_map.Map
	client     *client.Client
	lock       sync.Mutex
	handleLock sync.Mutex
}

func NewReplicaController() *ReplicaController {
	setMap := _map.NewMap()
	cli := client.NewClient(config.MasterIP, config.MasterPort)
	status, result, err := cli.GetRequest("/objectAPI/getAllReplicaSets")
	for err != nil {
		log.Printf("[ERROR] %v(Sleep for 1s...)\n", err)
		time.Sleep(time.Second)
		status, result, err = cli.GetRequest("/objectAPI/getAllReplicaSets")
	}
	if status == http.StatusOK {

		replicaSets := make([]*defines.ReplicaSet, 0)
		err = json.Unmarshal(result, &replicaSets)
		if err != nil {
			log.Printf("[ERROR]: %v\n", err)
			return &ReplicaController{
				setMap: setMap,
				client: cli,
			}
		} else {
			for _, replicaSet := range replicaSets {
				replica := &Replica{
					rs:       replicaSet,
					pods:     make(map[string]*defines.Pod),
					replicas: replicaSet.Spec.Replicas,
				}
				_ = setMap.Put(replicaSet.Metadata.Name, replica)
			}
		}

		logFile, logErr := os.OpenFile("./replicaSetLog.txt", os.O_RDWR|os.O_APPEND|os.O_CREATE, 777)
		logContent := ""
		newContent := ""
		if logErr != nil {
			log.Printf("[ERROR] %v\n", logErr)
		} else {
			//Here we simplify the process, support that the log file will not larger than 1G.
			info, _ := logFile.Stat()
			buf := make([]byte, info.Size())
			_, err := logFile.Read(buf)
			if err != nil {
				log.Printf("[ERROR] %v\n", err)
			} else {
				index := strings.LastIndex(string(buf), "BEGIN\n")
				if index == -1 {
					log.Printf("[WARN] Initialize may have problem!\n")
					logContent = string(buf)
				} else {
					logContent = string(buf)[index+6:]
					for logContent == "" && index != -1 {
						lastIndex := index
						index = strings.LastIndex(string(buf)[0:index], "BEGIN\n")
						if index == -1 {
							log.Printf("[WARN] Initialize may have problem!\n")
							logContent = ""
						} else {
							logContent = string(buf)[index+6 : lastIndex]
						}
					}
				}

				_, err := logFile.WriteString("BEGIN\n")
				if err != nil {
					log.Printf("[ERROR] %v\n", err)
				}
				newContent += "BEGIN\n"

				for _, replicaSet := range replicaSets {
					replica := setMap.Get(replicaSet.Metadata.Name).(*Replica)
					if replicaSet.Spec.Selector.MatchLabels.App == "" && replicaSet.Spec.Selector.MatchLabels.Env == "" {
						index := 0
						content := logContent
						name := replicaSet.Metadata.Name
						length := len(name)
						for {
							index = strings.Index(content, name)
							if index == -1 {
								break
							}
							if content[index+length] == '/' {
								content = content[index+length:]
								switch content[1] {
								case ':':
									last := strings.Index(content, "\n")
									if last != -1 {
										tmp := strings.Split(content[3:last], " ")
										for _, name := range tmp {
											replica.pods[name] = &defines.Pod{}
										}
										content = content[last+1:]
									} else {
										log.Printf("[WARN] Log format may have problem!\n")
										goto label
									}
								case '+':
									last := strings.Index(content, "\n")
									if last != -1 {
										replica.pods[content[2:last]] = &defines.Pod{}
										content = content[last+1:]
									} else {
										log.Printf("[WARN] Log format may have problem!\n")
										goto label
									}
								case '-':
									last := strings.Index(content, "\n")
									if last != -1 {
										delete(replica.pods, content[2:last])
										content = content[last+1:]
									} else {
										log.Printf("[WARN] Log format may have problem!\n")
										goto label
									}
								default:
									log.Printf("[ERROR] Log format have problem!\n")
									goto label
								}
							} else {
								index = strings.Index(content, "\n")
								if index == -1 {
									log.Printf("[ERROR] Log format have problem!\n")
									break
								}
								content = content[index+1:]
							}
						}

						_, err = logFile.WriteString(replicaSet.Metadata.Name + "/:")
						if err != nil {
							log.Printf("[ERROR] %v\n", err)
						} else {
							newContent += replicaSet.Metadata.Name + "/:"
							for name, _ := range replica.pods {
								_, err = logFile.WriteString(" " + name)
								if err != nil {
									log.Printf("[ERROR] %v\n", err)
								}
								newContent += " " + name
							}
							_, err = logFile.WriteString("\n")
							if err != nil {
								log.Printf("[ERROR] %v\n", err)
							}
							newContent += "\n"
						}
					label: //When meet wrong log format, the program should go here!
					}
				}

				//Here we compact the log in case of increasing infinitely
				//newInfo, _ := logFile.Stat()
				//size := newInfo.Size() - info.Size()
				//
				//finalBuf := make([]byte, size)
				//_, err = logFile.Read(finalBuf)
				//if err == nil {
				//	_, err = logFile.WriteAt(finalBuf, 0)
				//	if err != nil {
				//		log.Printf("[ERROR] %v\n", err)
				//	} else {
				//		err = logFile.Truncate(size)
				//		if err != nil {
				//			log.Printf("[ERROR] %v\n", err)
				//		}
				//	}
				//} else {
				//	log.Printf("[ERROR] %v\n", err)
				//}
			}

			err = logFile.Close()
			if err != nil {
				log.Printf("[ERROR] %v\n", err)
			} else {
				logFile, err := os.OpenFile("./replicaSetLog.txt", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 777)
				if err != nil {
					log.Printf("[ERROR] %v\n", err)
				} else {
					_, err = logFile.WriteString(newContent)
					if err != nil {
						log.Printf("[ERROR] %v\n", err)
					}
					err = logFile.Close()
					if err != nil {
						log.Printf("[ERROR] %v\n", err)
					}
				}
			}
		}
	}
	return &ReplicaController{
		setMap: setMap,
		client: cli,
	}
}

func (rc *ReplicaController) randomString(n int) string {
	rc.lock.Lock()
	defer rc.lock.Unlock()
	str := strings.Builder{}
	str.Grow(n)
	r := rand.New(rand.NewSource(time.Now().UnixMicro()))
	for i := 0; i < n; i++ {
		str.WriteByte(charset[r.Intn(len(charset))])
	}
	return str.String()
}

func (rc *ReplicaController) AddSet(rs *defines.ReplicaSet) error {
	replicas := rs.Spec.Replicas
	replica := &Replica{
		rs:       rs,
		pods:     make(map[string]*defines.Pod),
		replicas: replicas,
	}

	//May have read-write competing, so we must lock the critical path!
	rc.handleLock.Lock()
	defer rc.handleLock.Unlock()

	//Here we handle the replicaSets that are intended for existing pods
	if rs.Spec.Selector.MatchLabels.App == "" && rs.Spec.Selector.MatchLabels.Env == "" {
		logFile, fileErr := os.OpenFile("./replicaSetLog.txt", os.O_WRONLY|os.O_APPEND, 777)
		if fileErr != nil {
			log.Printf("[ERROR] %v\n", fileErr)
		} else {
			_, err := logFile.WriteString(rs.Metadata.Name + "/:")
			if err != nil {
				log.Printf("[ERROR] %v\n", err)
			}
		}
		//Here we handle replicaSets that are intended for newly created pods.
		yamlPod := &defines.YamlPod{
			ApiVersion: rs.ApiVersion,
			Kind:       "Pod",
			Metadata: defines.PodMetadata{
				//Name:  fmt.Sprintf("%s-%s", rs.Metadata.Name, randomString(5)),
				Label: rs.Spec.Template.Metadata.Label,
			},
			Spec: rs.Spec.Template.Spec,
		}
		//retryTime := 0
		for i := 0; i < int(replicas); /*&& retryTime <= 65535*/ i++ {
			name := fmt.Sprintf("%s-%s", rs.Metadata.Name, rc.randomString(5))
			for _, exist := replica.pods[name]; exist; _, exist = replica.pods[name] {
				name = fmt.Sprintf("%s-%s", rs.Metadata.Name, rc.randomString(5))
			}
			yamlPod.Metadata.Name = name
			replica.pods[name] = &defines.Pod{
				YamlPod: *yamlPod,
			}

			podVal, err := json.Marshal(yamlPod)
			if err != nil {
				log.Println("[ERROR]: The yaml format may have some errors!")
				return err
			}
			if fileErr == nil {
				_, err = logFile.WriteString(" " + name)
				if err != nil {
					log.Printf("[ERROR] %v\n", err)
				}
			}
			//TODO: We may send all pods a time
			res, err := rc.client.PostRequest(bytes.NewReader(podVal), "/objectAPI/createPod")
			if err != nil {
				log.Printf("%v\n", err)
				return err
			}
			////Here we change the hostPort!!
			//for index1, container := range yamlPod.Spec.Containers {
			//	for index2, port := range container.Ports {
			//		yamlPod.Spec.Containers[index1].Ports[index2].HostPort = port.HostPort%65535 + 1
			//	}
			//}
			log.Println(res)
			//if res[0:7] == "[ERROR]" { //Note here we simply think that the wrong response is due to port!
			//	i--
			//	retryTime++
			//}
		}
		if fileErr == nil {
			_, err := logFile.WriteString("\n")
			if err != nil {
				log.Printf("[ERROR] %v\n", err)
			}
			err = logFile.Close()
			if err != nil {
				log.Printf("[ERROR] %v\n", err)
			}
		}
		//rs.Spec.Template.Spec = yamlPod.Spec
	}
	//TODO: What if we have two replicaSet with different name but have the same selector labels?
	//TODO: What if we have two replicaSet with different name but have the same matchLabels?
	err := rc.setMap.Put(rs.Metadata.Name, replica)
	if err != nil {
		log.Printf("[ERROR]: The replicaSet %s have existed!\n", rs.Metadata.Name)
		return err
	}
	return nil
}

func (rc *ReplicaController) DeleteSet(rsName string) error {
	rc.handleLock.Lock()
	defer rc.handleLock.Unlock()

	rsVal := rc.setMap.Get(rsName)
	if rsVal == nil {
		return fmt.Errorf("the replicaSet %s does not exist", rsName)
	}
	rs := rsVal.(*Replica)
	for _, pod := range rs.pods {
		name, _ := json.Marshal(pod.Metadata.Name)
		res, err := rc.client.PostRequest(bytes.NewReader(name), "/objectAPI/deletePod")
		if err != nil {
			return err
		}
		log.Println(res)
	}
	rc.setMap.Del(rsName)

	return nil
}

func (rc *ReplicaController) DesReplicaSet(rsName string) *defines.DesRSInfo {
	val := rc.setMap.Get(rsName)
	if val == nil {
		return nil
	}
	rs := val.(*Replica)
	rsi := &defines.DesRSInfo{
		Info: defines.ReplicaSetInfo{
			Name:      rsName,
			Replicas:  rs.replicas,
			StartTime: rs.rs.StartTime,
		},
		PodName: make([]string, 0),
	}
	for name, _ := range rs.pods {
		rsi.PodName = append(rsi.PodName, name)
	}
	return rsi
}

func RSControllerCheckPodState(podInfo *defines.Pod) (bool, error) {
	nodeId := podInfo.NodeId
	cli := etcd.EtcdStart()
	defer cli.Close()
	nodeKey := defines.NodePrefix + "/" + nodeId
	kv := etcd.Get(cli, nodeKey).Kvs
	if len(kv) == 0 {
		return true, nil
	}
	nodeInfo := &defines.NodeInfo{}
	_ = yaml.Unmarshal(kv[0].Value, nodeInfo)
	ip := nodeInfo.NodeData.NodeSpec.Metadata.Ip
	body, err := json.Marshal(podInfo)
	if err != nil {
		return false, err
	}
	url := "http://" + ip + ":" + config.WorkerPort + "/objectAPI/checkRSPodState"
	log.Printf("[kubelet] check RCPod state url = %v\n", url)
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return false, err
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return false, err
	}
	if response.StatusCode != http.StatusOK {
		return false, errors.New("error when sending new pod state")
	} else {
		state := &defines.ReplicaPodState{}
		bodyReader := response.Body
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(bodyReader)
		_ = json.Unmarshal(buf.Bytes(), state)
		return state.Live, nil
	}
}

func (rc *ReplicaController) Watch() {
	for {
		rc.handleLock.Lock()
		status, result, err := rc.client.GetRequest("/objectAPI/getPods")
		if status == -1 {
			log.Printf("[ERROR]: %v\n", err)
			rc.handleLock.Unlock()
			time.Sleep(10 * time.Second)
			continue
		} else if status == http.StatusBadRequest {
			log.Printf("[ERROR]: %s\n", string(result))
			rc.handleLock.Unlock()
			time.Sleep(10 * time.Second)
			continue
		}

		pods := make([]*defines.Pod, 0)
		err = json.Unmarshal(result, &pods)
		if err != nil {
			log.Printf("[ERROR]: %v\n", err)
			rc.handleLock.Unlock()
			time.Sleep(10 * time.Second)
			continue
		}

		//Here I do think it needs to acquire the big lock!!!

		if rc.setMap.Size() == 0 {
			rc.handleLock.Unlock()
			time.Sleep(10 * time.Second)
			continue
		}
		//shallow-copy
		tmpMap := rc.setMap.Copy()

		//deep-copy
		resultMap := make(map[string]any, 0)
		for key, value := range tmpMap {
			rs := value.(*Replica)
			resultMap[key] = &Replica{
				rs:       rs.rs,
				pods:     make(map[string]*defines.Pod, 0),
				replicas: rs.replicas,
			}
		}

		for _, pod := range pods {
			label, err := json.Marshal(pod.Metadata.Label)
			if err != nil {
				log.Printf("[ERROR]: The pod %s's label may have some problems!\n", pod.Metadata.Name)
				label = []byte("")
			}

			res, _ := RSControllerCheckPodState(pod)
			if res == false {
				_ = cmds.SendOutDeletePod(pod.Metadata.Name)
				continue
			}

			for key, value := range tmpMap {
				rs := value.(*Replica)
				selector, err := json.Marshal(rs.rs.Spec.Selector.MatchLabels)
				if err != nil {
					log.Printf("[ERROR]: The replicaSet %s's label may have some problems!\n", key)
					selector = []byte("")
				}
				result := resultMap[key].(*Replica)
				if _, exist := rs.pods[pod.Metadata.Name]; exist {
					result.pods[pod.Metadata.Name] = pod
				} else if string(selector) == string(label) {
					result.pods[pod.Metadata.Name] = pod
				}
			}
		}

		//Here is ready to do some replica checking, but we first need to check the number of replicaSet
		//rc.handleLock.Lock()
		//rc.setMap.CheckIfAllExist(&resultMap)
		//tmpMap = rc.setMap.Copy()

		logFile, fileErr := os.OpenFile("./replicaSetLog.txt", os.O_WRONLY|os.O_APPEND, 777)
		if fileErr != nil {
			log.Printf("[ERROR] %v\n", fileErr)
		} else {
			for _, value := range resultMap {
				rs := value.(*Replica)
				if rs.rs.Spec.Selector.MatchLabels.App != "" || rs.rs.Spec.Selector.MatchLabels.Env != "" {
					continue
				}
				oldRs := tmpMap[rs.rs.Metadata.Name].(*Replica)
				for name, _ := range oldRs.pods {
					if _, exist := rs.pods[name]; !exist {
						_, err = logFile.WriteString(rs.rs.Metadata.Name + "/-" + name + "\n")
						if err != nil {
							log.Printf("[ERROR] %v\n", err)
						}
					}
				}
			}
		}

		for _, value := range resultMap {
			rs := value.(*Replica)
			yamlPod := &defines.YamlPod{
				ApiVersion: rs.rs.ApiVersion,
				Kind:       "Pod",
				Metadata: defines.PodMetadata{
					//Name:  name,
					Label: rs.rs.Spec.Template.Metadata.Label,
				},
				Spec: rs.rs.Spec.Template.Spec,
			}
			//retryTime := 0
			for int(rs.replicas) > len(rs.pods) /*&& retryTime < 65535*/ {
				name := fmt.Sprintf("%s-%s", rs.rs.Metadata.Name, rc.randomString(5))
				for _, exist := rs.pods[name]; exist; _, exist = rs.pods[name] {
					name = fmt.Sprintf("%s-%s", rs.rs.Metadata.Name, rc.randomString(5))
				}
				yamlPod.Metadata.Name = name
				podVal, err := json.Marshal(yamlPod)
				if err != nil {
					log.Println("[ERROR]: The yaml format may have some errors!")
					continue
				}
				if rs.rs.Spec.Selector.MatchLabels.App == "" && rs.rs.Spec.Selector.MatchLabels.Env == "" && fileErr == nil {
					_, err = logFile.WriteString(rs.rs.Metadata.Name + "/+" + name + "\n")
					if err != nil {
						log.Printf("[ERROR] %v\n", err)
					}
				}
				res, err := rc.client.PostRequest(bytes.NewReader(podVal), "/objectAPI/createPod")
				if err != nil {
					log.Printf("%v\n", err)
					continue
				}
				log.Println(res)
				////Here we change the hostPort!!
				//for index1, container := range yamlPod.Spec.Containers {
				//	for index2, port := range container.Ports {
				//		yamlPod.Spec.Containers[index1].Ports[index2].HostPort = port.HostPort%65535 + 1
				//	}
				//}
				//if res[0:7] == "[ERROR]" {
				//	retryTime++
				//	continue
				//}
				rs.pods[name] = &defines.Pod{
					YamlPod: *yamlPod,
				}
			}
			//rs.rs.Spec.Template.Spec = yamlPod.Spec

			for name := range rs.pods {
				if int(rs.replicas) == len(rs.pods) {
					break
				}
				podName, _ := json.Marshal(name)
				if rs.rs.Spec.Selector.MatchLabels.App == "" && rs.rs.Spec.Selector.MatchLabels.Env == "" && fileErr == nil {
					_, err = logFile.WriteString(rs.rs.Metadata.Name + "/-" + name + "\n")
					if err != nil {
						log.Printf("[ERROR] %v\n", err)
					}
				}
				res, err := rc.client.PostRequest(bytes.NewReader(podName), "/objectAPI/deletePod")
				if err != nil {
					log.Printf("[ERROR]: %v\n", err)
					continue
				} else {
					log.Print(res)
				}
				delete(rs.pods, name)
			}
		}

		err = logFile.Close()
		if err != nil {
			log.Printf("[ERROR] %v\n", err)
		}

		rc.setMap.UpdateMap(&resultMap)

		rc.handleLock.Unlock()

		time.Sleep(30 * time.Second)
	}
}

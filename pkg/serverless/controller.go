package serverless

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	dockerClient "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"gopkg.in/yaml.v3"
	"io"
	"log"
	"math/rand"
	"mini-k8s/pkg/config"
	"mini-k8s/pkg/defines"
	"mini-k8s/utils/client"
	_map "mini-k8s/utils/map"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// retryTime: used to avoid sending request before the pod is created actually!
const retryTime = 5

const MaxCallTime = 65536

type SLController struct {
	fMap         *_map.Map
	wMap         *_map.Map
	client       *client.Client
	functionLock sync.Mutex
	workflowLock sync.Mutex
}

type FuncRS struct {
	Function     string
	Image        string
	Requests     int32
	Replicas     int32
	RecentInvoke int64
	privateLock  sync.Mutex
	Lock         sync.RWMutex
}

func newFuncRS(name, image string) *FuncRS {
	return &FuncRS{
		Function:     name,
		Image:        image,
		Requests:     0,
		Replicas:     0,
		RecentInvoke: time.Now().Unix(),
	}
}

func randString(n int) string {
	str := strings.Builder{}
	str.Grow(n)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < n; i++ {
		str.WriteByte(charset[r.Intn(len(charset))])
	}
	log.Println(str.String())
	return str.String()
}

func (slc *SLController) prepareForFunctionChanging(funcRS *FuncRS) {
	status, result, err := slc.client.GetRequest("/objectAPI/getFuncPods/" + funcRS.Function)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
	} else if status == http.StatusBadRequest {
		log.Println(string(result))
	} else {
		pods := make([]*defines.Pod, 0)
		err = json.Unmarshal(result, &pods)
		if err != nil {
			log.Printf("[ERROR] %v\n", err)
		} else {
			for funcRS.Replicas > 0 {
				body, _ := json.Marshal(pods[0].Metadata.Name)
				res, err := slc.client.PostRequest(bytes.NewReader(body), "/objectAPI/deletePod")
				if err != nil {
					log.Printf("[ERROR] %v\n", err)
				} else if strings.HasPrefix(res, "[ERROR]") {
					log.Print(res)
				}
				funcRS.Replicas--
				pods = pods[1:]
				//Here we do some really simple operation: if the delete operation is failed, we just skip this pod!
				//And we will no longer track these pods, which may be gotten rid of by daemon thread.
			}
		}
	}

	funcRS.Replicas = 0

	body, _ := json.Marshal(funcRS.Image)

	res, err := slc.client.PostRequest(bytes.NewReader(body), "/objectAPI/delImage")
	if err != nil {
		log.Printf("[WARNING] The function changing may meet inconsistent situation!(err during delete image in work node: %v)\n", err)
	} else {
		log.Println(res)
	}
}

func (slc *SLController) scanFunctionDir() {
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

	for _, fun := range dirList {
		err := slc.AddFunction(fun)
		if err != nil {
			log.Printf("[ERROR] %v\n", err)
		}
		retry := 0
		body, _ := json.Marshal(fun)
	send:
		res, err := slc.client.PostRequest(bytes.NewReader(body), "/objectAPI/initFunc")
		if err != nil {
			time.Sleep(time.Second)
			retry++
			if retry < 5 {
				goto send
			} else {
				log.Printf("[ERROR] Failed to initial function %s when send to apiServer!\n", fun)
			}
		} else {
			if strings.HasPrefix(res, "[ERROR]") {
				retry++
				log.Println(res)
				if retry < 5 {
					goto send
				} else {
					log.Printf("[ERROR] Failed to initial function %s when send to apiServer!\n", fun)
				}
			}
		}
	}
}

func SLControllerInit() *SLController {
	slc := &SLController{
		fMap:   _map.NewMap(),
		wMap:   _map.NewMap(),
		client: client.NewClient(config.MasterIP, config.MasterPort),
	}
	for i := 0; i <= 5; i++ {
		if i == 5 {
			log.Println("[WARNING] Can not initialize workflows in apiServer!")
			break
		}
		status, res, err := slc.client.DelRequest("/objectAPI/deleteOldWorkflows")
		if err != nil {
			log.Printf("[ERROR] %v\n", err)
			time.Sleep(time.Second)
		} else if status != http.StatusOK {
			log.Println(res)
			time.Sleep(time.Second)
		} else {
			log.Println("[INFO] Initialize workflows in apiServer successfully!")
			break
		}
	}
	slc.scanFunctionDir()
	return slc
}

func createImage(fileName, imageName, op string) error {
	cli, err := dockerClient.NewClientWithOpts(dockerClient.WithVersion(config.DockerApiVersion))
	if err != nil {
		return err
	}

	images, _ := cli.ImageList(context.Background(), types.ImageListOptions{})

	if op == "new" {
		for _, im := range images {
			for _, name := range im.RepoTags {
				if name == imageName {
					return nil
				}
			}
		}
	} else if op == "update" {
		cmd := exec.Command("docker", "rmi", imageName)
		err = cmd.Run()
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unknown op type %s", op)
	}

	path := filepath.Join(config.ServerlessFileDir, fileName, "/Dockerfile")

	workDir := filepath.Dir(path)

	buildCtx, err := archive.TarWithOptions(workDir, &archive.TarOptions{})
	if err != nil {
		return err
	}

	authConfig := types.AuthConfig{
		Username: defines.DockerUsername,
		Password: defines.DockerPwd,
	}
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		return err
	}

	authStr := base64.URLEncoding.EncodeToString(encodedJSON)

	buildResp, err := cli.ImageBuild(context.Background(), buildCtx, types.ImageBuildOptions{
		Tags:       []string{imageName},
		Dockerfile: filepath.Base(path),
		NoCache:    true,
		Remove:     true,
		BuildArgs: map[string]*string{
			"DOCKER_AUTH_CONFIG": &authStr,
		},
	})
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(buildResp.Body)

	_, err = io.Copy(os.Stdout, buildResp.Body)
	if err != nil {
		return fmt.Errorf("can't construct the docker image")
	}

	pushResp, err := cli.ImagePush(context.Background(), imageName, types.ImagePushOptions{
		RegistryAuth: authStr,
	})

	if err != nil {
		return err
	}
	defer func(pushResp io.ReadCloser) {
		_ = pushResp.Close()

	}(pushResp)

	_, err = io.Copy(os.Stdout, pushResp)
	if err != nil {
		return err
	}

	return nil
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func(srcFile *os.File) {
		_ = srcFile.Close()
	}(srcFile)

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func(dstFile *os.File) {
		_ = dstFile.Close()
	}(dstFile)

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}
	return nil
}

func prepareFile(fileName string) error {
	mainFileSrc := filepath.Join(config.SourceDir, "main.py")
	dockerfileSrc := filepath.Join(config.SourceDir, "Dockerfile")
	mainFileDst := filepath.Join(config.ServerlessFileDir, fileName, "main.py")
	dockerfileDst := filepath.Join(config.ServerlessFileDir, fileName, "Dockerfile")

	err := copyFile(mainFileSrc, mainFileDst)
	if err != nil {
		return err
	}

	err = copyFile(dockerfileSrc, dockerfileDst)
	if err != nil {
		return err
	}

	req, err := os.OpenFile(filepath.Join(config.ServerlessFileDir, fileName, "requirements.txt"), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 664)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(req)

	_, err = req.WriteString(config.Dependency)
	if err != nil {
		return err
	}

	return nil
}

func (slc *SLController) addPod(function, image string) error {
	file, err := os.Open(config.ServerlessTemplatePod)
	defer func(file *os.File) {
		_ = file.Close()
	}(file)
	if err != nil {
		return err
	}
	yamlPod := &defines.YamlPod{}
	err = yaml.NewDecoder(file).Decode(yamlPod)
	if err != nil {
		return err
	}
	yamlPod.Metadata.Name += function + "-" + randString(5)
	yamlPod.Spec.Containers[0].Name = yamlPod.Metadata.Name
	yamlPod.Spec.Containers[0].Image = image

	body, err := json.Marshal(yamlPod)
	if err != nil {
		return err
	}
	res, err := slc.client.PostRequest(bytes.NewReader(body), "/objectAPI/createPod")
	if err != nil {
		return err
	}
	if strings.HasPrefix(res, "[ERROR]") {
		return fmt.Errorf(res[9:])
	} else {
		return nil
	}
}

func (slc *SLController) trigger(funcRS *FuncRS, body []byte) (map[string]string, error) {
	var err error
	for i := 0; i < retryTime; i++ {
		status, reString, err1 := slc.client.GetRequest("/objectAPI/getFuncPods/" + funcRS.Function)
		err = err1
		if err != nil {
			return nil, err
		} else if status == http.StatusBadRequest {
			return nil, fmt.Errorf(string(reString))
		}

		pods := make([]*defines.Pod, 0)
		err = json.Unmarshal(reString, &pods)
		if err != nil {
			return nil, err
		}

		length := len(pods)
		if length == 0 {
			err = fmt.Errorf("no available serverless instance for %s", funcRS.Function)
			time.Sleep(500 * time.Millisecond)
			continue
		}
		r := rand.New(rand.NewSource(time.Now().UnixMicro()))
		id := r.Intn(length)
		url := fmt.Sprintf("http://%s:9090/", pods[id].PodIp)
		request, err := http.NewRequest("POST", url, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		res, err := http.DefaultClient.Do(request)
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		result := make(map[string]string)
		err = json.NewDecoder(res.Body).Decode(&result)
		if err != nil {
			return nil, err
		} else if len(result) == 0 {
			continue
		} else {
			log.Printf("[INFO] The request is handled in ip: %v\n", url)
			return result, nil
		}
	}
	return nil, err
}

func judge(params map[string]string, condition *defines.Condition) (bool, error) {
	L := condition.ParamL.Name
	R := condition.ParamR.Name
	if condition.ParamL.Kind == "VAR" {
		if _, exist := params[condition.ParamL.Name]; !exist {
			return false, fmt.Errorf("var %s does not exist in params", condition.ParamL.Name)
		}
		L = params[condition.ParamL.Name]
	}

	if condition.ParamR.Kind == "VAR" {
		if _, exist := params[condition.ParamR.Name]; !exist {
			return false, fmt.Errorf("var %s does not exist in params", condition.ParamR.Name)
		}
		R = params[condition.ParamR.Name]
	}

	switch condition.Symbol {
	case "EQUAL":
		if L == R {
			return true, nil
		} else {
			return false, nil
		}

	case "NOTEQUAL":
		if L == R {
			return false, nil
		} else {
			return true, nil
		}

	default:
		break
	}

	NL, err := strconv.ParseInt(L, 10, 64)
	if err != nil {
		return false, err
	}

	NR, err := strconv.ParseInt(R, 10, 64)
	if err != nil {
		return false, err
	}

	switch condition.Symbol {
	case "LARGER":
		if NL > NR {
			return true, nil
		} else {
			return false, nil
		}

	case "LARGER-EQUAL":
		if NL >= NR {
			return true, nil
		} else {
			return false, nil
		}

	case "LESS":
		if NL < NR {
			return true, nil
		} else {
			return false, nil
		}

	case "LESS-EQUAL":
		if NL <= NR {
			return true, nil
		} else {
			return false, nil
		}

	default:
		return false, fmt.Errorf("unknown symbol %s", condition.Symbol)
	}
}

func (slc *SLController) AddFunction(fileName string) error {
	imageName := fmt.Sprintf("7143192/%s:latest", fileName)
	//Here the lock is used to protect the file from being overwritten by another request
	slc.functionLock.Lock()
	defer slc.functionLock.Unlock()
	if slc.fMap.Get(fileName) != nil {
		return fmt.Errorf("the function has existed")
	}

	if err := prepareFile(fileName); err != nil {
		return err
	}

	if err := createImage(fileName, imageName, "new"); err != nil {
		return err
	}

	return slc.fMap.Put(fileName, newFuncRS(fileName, imageName))
}

func (slc *SLController) UpdateFunction(fileName string) error {
	imageName := fmt.Sprintf("7143192/%s:latest", fileName)
	//This lock blocks other write-operators like add or another update
	slc.functionLock.Lock()
	defer slc.functionLock.Unlock()
	funcVal := slc.fMap.Get(fileName)
	if funcVal == nil {
		return fmt.Errorf("the function doesn't exist")
	}

	funcRS := funcVal.(*FuncRS)
	funcRS.Lock.Lock()
	defer funcRS.Lock.Unlock()

	//Note that we simplify the process when meet an error, in which we just remove the function instance!
	if err := prepareFile(fileName); err != nil {
		slc.fMap.Del(fileName)
		return err
	}

	//remove all running pods
	slc.prepareForFunctionChanging(funcRS)

	if err := createImage(fileName, imageName, "update"); err != nil {
		slc.fMap.Del(fileName)
		return err
	}

	return nil
}

func (slc *SLController) TriggerFunction(fileName string, data map[string]string) (map[string]string, error) {
	//This lock is used to avoid a request get a function-replicaSet(funcRS) that is removed!
	slc.functionLock.Lock()
	val := slc.fMap.Get(fileName)
	if val == nil {
		slc.functionLock.Unlock()
		return nil, fmt.Errorf("no function named %s", fileName)
	}

	funcRS := val.(*FuncRS)
	funcRS.Lock.RLock()
	//Here we must wait the probable "old" funcRS getting its reader lock, for the same reason as above
	slc.functionLock.Unlock()
	funcRS.privateLock.Lock()
	funcRS.Requests++

	defer func() {
		funcRS.privateLock.Lock()
		funcRS.Requests--
		funcRS.RecentInvoke = time.Now().Unix()
		funcRS.privateLock.Unlock()
		funcRS.Lock.RUnlock()
	}()

	if funcRS.Requests > funcRS.Replicas {
		err := slc.addPod(fileName, funcRS.Image)
		if err != nil {
			funcRS.privateLock.Unlock()
			return nil, err
		} else {
			funcRS.Replicas++
		}
	}

	funcRS.privateLock.Unlock()

	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	result, err := slc.trigger(funcRS, body)
	if err != nil {
		return nil, err
	} else {
		return result, nil
	}
}

func (slc *SLController) DelFunction(fileName string) error {
	slc.functionLock.Lock()
	defer slc.functionLock.Unlock()
	val := slc.fMap.Get(fileName)
	if val == nil {
		return fmt.Errorf("no function named %s", fileName)
	}

	funcRS := val.(*FuncRS)
	funcRS.Lock.Lock()
	defer func() {
		slc.fMap.Del(fileName)
		funcRS.Lock.Unlock()
	}()

	slc.prepareForFunctionChanging(funcRS)

	err := os.RemoveAll(config.ServerlessFileDir + "/" + funcRS.Function)
	if err != nil {
		log.Printf("[ERROR] Remove old function files failed: %v\n", err)
	}

	cmd := exec.Command("docker", "rmi", "7143192/"+fileName+":latest")
	err = cmd.Run()
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		return fmt.Errorf("can not remove the local image")
	}
	return nil
}

func (slc *SLController) AddWorkFlow(workflow *defines.WorkFlow) error {
	//TODO: If we decide to support function-del, we may need lock the "functionLock" here
	for _, task := range workflow.Tasks {
		if slc.fMap.Get(task.Name) == nil {
			return fmt.Errorf("no function named %s", task.Name)
		}
	}
	return slc.wMap.Put(workflow.Name, workflow)
}

func (slc *SLController) UpdateWorkFlow(workflow *defines.WorkFlow) error {
	//This lock is to avoid some requests getting old workflow(If we can del function, the error is fatal)
	slc.workflowLock.Lock()
	defer slc.workflowLock.Unlock()
	oldWorkFlowVal := slc.wMap.Get(workflow.Name)
	if oldWorkFlowVal == nil {
		return fmt.Errorf("workflow %s doesn't exist", workflow.Name)
	}
	oldWorkFlow := oldWorkFlowVal.(*defines.WorkFlow)
	oldWorkFlow.Lock.Lock()
	defer oldWorkFlow.Lock.Unlock()
	slc.wMap.Update(workflow.Name, workflow)
	return nil
}

func (slc *SLController) TriggerWorkFlow(name string, data map[string]string) (map[string]string, error) {
	slc.workflowLock.Lock()
	val := slc.wMap.Get(name)
	if val == nil {
		slc.workflowLock.Unlock()
		return nil, fmt.Errorf("workflow named %s does not exist", name)
	}
	workflow := val.(*defines.WorkFlow)
	//This reader lock is to avoid the workflow being updated or removed before it finishes
	workflow.Lock.RLock()
	slc.workflowLock.Unlock()
	defer workflow.Lock.RUnlock()

	//In case of getting old funcRS
	slc.functionLock.Lock()
	for _, task := range workflow.Tasks {
		taskVal := slc.fMap.Get(task.Name)
		if taskVal == nil {
			slc.functionLock.Unlock()
			return nil, fmt.Errorf("the function %s doesn't exist", task.Name)
		}
		function := taskVal.(*FuncRS)
		//Here to handle concurrent situation...
		function.Lock.RLock()
		defer function.Lock.RUnlock()
	}
	slc.functionLock.Unlock()

	currentFunction := workflow.Start

	result, err := slc.TriggerFunction(currentFunction, data)
	if err != nil {
		return nil, err
	}

	for i := 0; i < MaxCallTime; i++ {
		next := ""
		hasUpdate := false
		for _, relationship := range workflow.Relationships {
			if relationship.Left == currentFunction {
				next = relationship.Right
				break
			}
		}

		if next == "" {
			return result, nil
		}

		for _, choice := range workflow.Choices {
			if choice.Name == next {
				flag, err := judge(result, &choice.Condition)
				if err != nil {
					return result, err
				} else if flag {
					currentFunction = choice.True
				} else {
					currentFunction = choice.False
				}

				hasUpdate = true
			}
		}

		if hasUpdate {
			//Here we handle the situation "choice->choice"
		judgeForChoice:
			for _, choice := range workflow.Choices {
				if choice.Name == currentFunction {
					flag, err := judge(result, &choice.Condition)
					if err != nil {
						return result, err
					} else if flag {
						currentFunction = choice.True
					} else {
						currentFunction = choice.False
					}
					goto judgeForChoice
				}
			}
			result, err = slc.TriggerFunction(currentFunction, result)
			if err != nil {
				return nil, err
			}
			continue
		}

		for _, task := range workflow.Tasks {
			if task.Name == next {
				currentFunction = task.Name
				hasUpdate = true
				break
			}
		}

		if hasUpdate {
			result, err = slc.TriggerFunction(currentFunction, result)
			if err != nil {
				return nil, err
			}
			continue
		}

		return result, fmt.Errorf("it seems the workflow does not end but there is not task that can be scheduled")
	}

	return result, fmt.Errorf("the call time has reached the limit")
}

func (slc *SLController) DelWorkFlow(name string) error {
	slc.workflowLock.Lock()
	defer slc.workflowLock.Unlock()
	workflowVal := slc.wMap.Get(name)
	if workflowVal == nil {
		return fmt.Errorf("workflow %s doesn't exist", name)
	}
	workflow := workflowVal.(*defines.WorkFlow)
	workflow.Lock.Lock()
	defer workflow.Lock.Unlock()
	slc.wMap.Del(name)
	return nil
}

func (slc *SLController) Watch() {
	for {
		cp := slc.fMap.Copy()
		for key, val := range cp {
			funcRS := val.(*FuncRS)
			if funcRS.Lock.TryLock() {
				if funcRS.Replicas > 0 && time.Now().Unix()-funcRS.RecentInvoke > int64(10) {
					log.Printf("[INFO] The %v has replicas %v.\n", key, funcRS.Replicas)
					status, result, err := slc.client.GetRequest("/objectAPI/getFuncPods/" + key)
					if err != nil {
						log.Printf("[ERROR] %v\n", err)
						goto unlock
					} else if status == http.StatusBadRequest {
						log.Println(string(result))
						goto unlock
					}

					pods := make([]*defines.Pod, 0)
					err = json.Unmarshal(result, &pods)
					if err != nil {
						log.Printf("[ERROR] %v\n", err)
						goto unlock
					}

					length := len(pods)

					if funcRS.Replicas != int32(length) {
						log.Printf("[WARNING] The replicas is %v, but the etcd has %v!\n", funcRS.Replicas, length)
					}

					if length == 0 {
						log.Printf("[ERROR] Detect errror! The Function %s should not have no running pods!", key)
						goto unlock
					}

					r := rand.New(rand.NewSource(time.Now().UnixMicro()))
					id := r.Intn(length)
					body, _ := json.Marshal(pods[id].Metadata.Name)
					res, err := slc.client.PostRequest(bytes.NewReader(body), "/objectAPI/deletePod")
					if err != nil {
						log.Printf("[ERROR] %v\n", err)
						goto unlock
					} else if strings.HasPrefix(res, "[ERROR]") {
						log.Print(res)
						goto unlock
					}
					funcRS.Replicas--
					log.Println(res)
				}
			unlock:
				funcRS.Lock.Unlock()
			}
		}
		time.Sleep(10 * time.Second)
	}
}

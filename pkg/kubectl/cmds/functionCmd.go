package cmds

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
	"io"
	"mime/multipart"
	"mini-k8s/pkg/config"
	"mini-k8s/pkg/defines"
	"net/http"
	"os"
	"path/filepath"
)

func FunctionCmd() *cli.Command {
	return &cli.Command{
		Name:  "func",
		Usage: "command about serverless function",
		Subcommands: []*cli.Command{
			{
				Name:  "add",
				Usage: "add a new func to the apiServer",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "type",
						Aliases:  []string{"t"},
						Usage:    "type of the thing that need add to the apiServer",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "path",
						Aliases:  []string{"p"},
						Usage:    "path of the file of the new function project",
						Required: true,
					},
				},
				Action: func(c *cli.Context) error {
					if c.String("type") == "F" {
						newFunction(c, "add")
					} else if c.String("type") == "W" {
						newWorkFlow(c, "add")
					} else {
						fmt.Printf("Unknown type %s, only support 'F'(Function) and 'W'(Workflow)")
					}
					return nil
				},
			},
			{
				Name:  "update",
				Usage: "update a function or workflow",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "type",
						Aliases:  []string{"t"},
						Usage:    "the updated thing's type",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "path",
						Aliases:  []string{"p"},
						Usage:    "the new one's path",
						Required: true,
					},
				},
				Action: func(c *cli.Context) error {
					if c.String("type") == "F" {
						newFunction(c, "update")
					} else if c.String("type") == "W" {
						newWorkFlow(c, "update")
					} else {
						fmt.Printf("Unknown type %s, only support 'F'(Function) and 'W'(Workflow)")
					}
					return nil
				},
			},
			{
				Name:  "trigger",
				Usage: "trigger a function or workflow by http post",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "type",
						Aliases:  []string{"t"},
						Usage:    "the triggered thing's type",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "name",
						Aliases:  []string{"n"},
						Usage:    "the triggered function's name",
						Required: true,
					},
					&cli.StringFlag{
						Name:    "params",
						Aliases: []string{"p"},
						Usage:   "the triggered function's parameters",
					},
				},
				Action: func(c *cli.Context) error {
					trigger(c)
					return nil
				},
			},
			{
				Name:  "get",
				Usage: "get all functions or workflows in the system",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "type",
						Aliases:  []string{"t"},
						Usage:    "the gotten thing's type",
						Required: true,
					},
				},
				Action: func(c *cli.Context) error {
					get(c)
					return nil
				},
			},
			{
				Name:  "del",
				Usage: "delete function or workflow in the system",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "type",
						Aliases:  []string{"t"},
						Usage:    "the deleted thing's type",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "name",
						Aliases:  []string{"n"},
						Usage:    "the deleted thing's name",
						Required: true,
					},
				},
				Action: func(c *cli.Context) error {
					del(c)
					return nil
				},
			},
		},
	}
}

func newFunction(c *cli.Context, op string) {
	path := c.String("path")

	if filepath.Ext(path) != ".zip" {
		fmt.Printf("[ERROR] Only support .zip file!\n")
		return
	}

	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("[ERROR] %v\n", err)
		return
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	fileContent, err := writer.CreateFormFile("file", file.Name())
	if err != nil {
		fmt.Printf("[ERROR] %v\n", err)
		return
	}

	_, err = io.Copy(fileContent, file)
	if err != nil {
		fmt.Printf("[ERROR] %v\n", err)
		return
	}

	_ = writer.Close()

	url := ""

	if op == "add" {
		url = "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/newFunction/add"
	} else if op == "update" {
		url = "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/newFunction/update"
	} else {
		fmt.Printf("Unknown op %s\n", op)
		return
	}

	request, err := http.NewRequest("POST", url, body)
	if err != nil {
		fmt.Printf("[ERROR] %v\n", err)
		return
	}
	request.Header.Add("Content-Type", writer.FormDataContentType())

	res, err := http.DefaultClient.Do(request)

	if err != nil {
		fmt.Printf("[ERROR] %v\n", err)
		return
	}

	result := make(map[string]string)
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		fmt.Printf("[ERROR] %v\n", err)
		return
	}

	if res.StatusCode == http.StatusOK {
		fmt.Printf("[INFO] %v\n", result["INFO"])
	} else {
		fmt.Printf("[ERROR] %v\n", result["ERROR"])
	}
}

func trigger(c *cli.Context) {
	kind := c.String("type")
	if kind != "F" && kind != "W" {
		fmt.Printf("Wrong type! Only 'F'(Function) and 'W'(Workflow) are supported!\n")
		return
	}

	name := c.String("name")
	params := c.String("params")
	data := make(map[string]string)
	var err error
	if params != "" {
		err = json.Unmarshal([]byte(params), &data)
	}

	if err != nil {
		fmt.Printf("The parameters format is wrong!\n")
		return
	}

	body, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Can not marshal the parameters into json objects!\n")
		return
	}

	url := ""
	if kind == "F" {
		url = "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/trigger/" + name + "/F"
	} else {
		url = "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/trigger/" + name + "/W"
	}
	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	request.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(request)

	if err != nil {
		fmt.Println(err)
		return
	}

	if res.StatusCode == http.StatusOK {
		result := make(map[string][]byte)
		err = json.NewDecoder(res.Body).Decode(&result)
		if err != nil {
			fmt.Printf("Can not parse the response body!\n")
			return
		}
		info := make(map[string]string)
		err = json.Unmarshal(result["INFO"], &info)
		if err != nil {
			fmt.Printf("Can not parse the compute result!\n")
			return
		}
		fmt.Printf("The compute result is %v\n", info)
	} else {
		result := make(map[string]string)
		err = json.NewDecoder(res.Body).Decode(&result)
		if err != nil {
			fmt.Printf("Can not parse the response body!\n")
			return
		}
		fmt.Println(string(result["ERROR"]))
	}
}

func newWorkFlow(c *cli.Context, op string) {
	path := c.String("path")
	file, err := os.Open(path)

	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	if err != nil {
		fmt.Printf("[ERROR] %v\n", err)
		return
	}

	workflow := &defines.WorkFlow{}
	err = yaml.NewDecoder(file).Decode(workflow)

	if err != nil {
		fmt.Printf("[ERROR] %v\n", err)
		return
	} else if workflow.Kind != "WorkFlow" {
		fmt.Printf("[ERROR] the file %s does not config a workflow!\n", path)
		return
	}

	//TODO: Here we may add some basic checks to ensure the workflow config error-free

	body, err := json.Marshal(workflow)

	if err != nil {
		fmt.Printf("[ERROR] %v\n", err)
		return
	}

	url := ""

	if op == "add" {
		url = "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/newWorkFlow/add"
	} else if op == "update" {
		url = "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/newWorkFlow/update"
	} else {
		fmt.Printf("Unknown op %s\n", op)
		return
	}

	request, err := http.NewRequest("POST", url, bytes.NewReader(body))

	if err != nil {
		fmt.Printf("[ERROR] %v\n", err)
		return
	}

	request.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(request)

	if err != nil {
		fmt.Printf("[ERROR] %v\n", err)
		return
	}

	result := make(map[string]string)
	err = json.NewDecoder(res.Body).Decode(&result)

	if err != nil {
		fmt.Println("Can not parse response!")
		return
	}

	if res.StatusCode == http.StatusOK {
		fmt.Println(result["INFO"])
	} else {
		fmt.Println(result["ERROR"])
	}
}

func get(c *cli.Context) {
	op := c.String("type")
	if op != "F" && op != "W" {
		fmt.Printf("[ERROR] Unknown type %s!\n", op)
		return
	}
	url := "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/getServerlessObject/" + op

	request, err := http.NewRequest("GET", url, nil)

	if err != nil {
		fmt.Printf("[ERROR] %v\n", err)
		return
	}

	request.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(request)

	if err != nil {
		fmt.Printf("[ERROR] %v\n", err)
		return
	}

	if res.StatusCode == http.StatusOK {
		result := make(map[string][]byte)
		err = json.NewDecoder(res.Body).Decode(&result)
		if err != nil {
			fmt.Println("Can not parse response!")
			return
		}

		data := make([]string, 0)
		err = json.Unmarshal(result["INFO"], &data)
		if err != nil {
			fmt.Println("Can not parse response!")
			return
		}
		if op == "F" {
			fmt.Println("All functions are: ")
		} else {
			fmt.Println("All workflows are: ")
		}
		for _, name := range data {
			fmt.Println(name)
		}
	} else {
		result := make(map[string]string)
		err = json.NewDecoder(res.Body).Decode(&result)
		if err != nil {
			fmt.Println("Can not parse response!")
			return
		}
		fmt.Println(result["ERROR"])
	}
}

func del(c *cli.Context) {
	op := c.String("type")
	if op != "F" && op != "W" {
		fmt.Printf("[ERROR] Unknown type %s!\n", op)
		return
	}
	name := c.String("name")
	url := "http://" + config.MasterIP + ":" + config.MasterPort + "/objectAPI/deleteServerlessObject/" + op + "/" + name

	request, err := http.NewRequest("DELETE", url, nil)

	if err != nil {
		fmt.Printf("[ERROR] %v\n", err)
		return
	}

	request.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(request)

	if err != nil {
		fmt.Printf("[ERROR] %v\n", err)
		return
	}

	result := make(map[string]string)
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		fmt.Println("Can not parse response!")
		return
	}

	if res.StatusCode == http.StatusOK {
		fmt.Println(result["INFO"])
	} else {
		fmt.Println(result["ERROR"])
	}
}

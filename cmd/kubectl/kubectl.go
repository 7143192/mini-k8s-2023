package main

import (
	"bufio"
	"fmt"
	"github.com/urfave/cli/v2"
	"log"
	cmds2 "mini-k8s/pkg/kubectl/cmds"
	"os"
	"strings"
)

func Init() *cli.App {
	app := cli.NewApp()
	app.Name = "kubectl"
	app.Version = "1.0"
	app.Usage = "one cmd tool"
	app.Flags = []cli.Flag{}
	app.Commands = []*cli.Command{
		cmds2.HelloWorld(),
		cmds2.CreateCmd(),
		cmds2.DescribeCmd(),
		cmds2.DeleteCmd(),
		cmds2.GetCmd(),
		cmds2.FunctionCmd(),
	}
	return app
}

func ParseArgs(app *cli.App, cmdStr string) error {
	// remove start and end whitespace of the input cmdLine.
	cmdStr = strings.Trim(cmdStr, " ")
	parts := strings.Split(cmdStr, " ")
	err := app.Run(parts)
	if err != nil {
		log.Fatal("[Fault] ", err)
	}
	return err
}

func main() {
	app := Init()
	for {
		fmt.Printf("kubectl> ")
		cmdReader := bufio.NewReader(os.Stdin)
		cmdStr, _ := cmdReader.ReadString('\n')
		cmdStr = strings.Trim(cmdStr, "\r\n")
		// use "exit" to quit one running kubectl command line.
		if cmdStr == "exit" {
			return
		} else {
			_ = ParseArgs(app, cmdStr)
		}
	}
}

package cmds

import (
	"fmt"
	"github.com/urfave/cli/v2"
)

// usage: kubectl helloworld

func HelloWorld() *cli.Command {
	return &cli.Command{
		Name:  "helloworld",
		Usage: "hello world!",
		Flags: []cli.Flag{},
		Action: func(c *cli.Context) error {
			fmt.Println("Hello World!")
			return nil
		},
	}
}

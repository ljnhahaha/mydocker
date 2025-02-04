package main

import (
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"mydocker/cgroups/subsystems"
	"mydocker/container"
)

var runCommand = cli.Command{
	Name: "run",
	Usage: `Create a container 
	        mydocker run -it [command]`,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "it",
			Usage: "enable tty",
		},
		&cli.StringFlag{
			Name:  "mem",
			Usage: "limit memory, e.g., -mem 100m",
		},
		&cli.IntFlag{
			Name:  "cpu",
			Usage: "limit cpu, e.g., -cpu 100",
		},
		&cli.StringFlag{
			Name:  "cpuset",
			Usage: "limit cpuset, e.g., -cpuset 0,1",
		},
		&cli.StringFlag{
			Name:  "v",
			Usage: "volume, e.g., -v /ect/conf:/etc/conf",
		},
	},
	Action: func(c *cli.Context) error {
		// c.Args() 不包括flag相关参数
		if c.Args().Len() < 1 {
			return errors.New("missing container command")
		}

		// cmd := c.Args().Get(0)
		tty := c.Bool("it")
		resCfg := &subsystems.ResourceConfig{
			MemoryLimit: c.String("mem"),
			CpuSet:      c.String("cpuset"),
			CpuCfsQuota: c.Int("cpu"),
		}

		volume := c.String("v")

		RunCmds(tty, c.Args().Slice(), resCfg, volume)

		return nil
	},
}

var initCommand = cli.Command{
	Name: "init",
	Usage: `Init a container and run user's command, do not use this
			command directly`,
	Action: func(c *cli.Context) error {
		log.Info("Init container ...")
		// cmd := c.Args().Get(0)

		_ = container.RunContainerInitProcess()

		return nil
	},
}

var commitCommand = cli.Command{
	Name:  "commit",
	Usage: "commit container to image",
	Action: func(c *cli.Context) error {
		if len(c.Args().Slice()) < 1 {
			return fmt.Errorf("commit missing image name")
		}
		imageName := c.Args().First()
		container.CommitContainer(imageName)

		return nil
	},
}

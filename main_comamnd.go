package main

import (
	"errors"
	"fmt"
	"os"

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
		&cli.StringFlag{
			Name:  "name",
			Usage: "specify container name",
		},
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
		&cli.BoolFlag{
			Name:  "d",
			Usage: "detach container",
		},
	},
	Action: func(c *cli.Context) error {
		// c.Args() 不包括flag相关参数
		if c.Args().Len() < 1 {
			return errors.New("missing container command")
		}

		tty := c.Bool("it")
		detach := c.Bool("d")
		// 实际运行中只依靠-it来判断是否后台运行
		if tty && detach {
			return fmt.Errorf("it and d parameter can not be both provided")
		}

		resCfg := &subsystems.ResourceConfig{
			MemoryLimit: c.String("mem"),
			CpuSet:      c.String("cpuset"),
			CpuCfsQuota: c.Int("cpu"),
		}

		volume := c.String("v")
		containerName := c.String("name")

		Run(tty, c.Args().Slice(), resCfg, volume, containerName)

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

var listCommand = cli.Command{
	Name:  "ps",
	Usage: "list all containers",
	Action: func(c *cli.Context) error {
		container.ListContainers()
		return nil
	},
}

var logCommand = cli.Command{
	Name:  "logs",
	Usage: "output logs of a container",
	Action: func(c *cli.Context) error {
		if len(c.Args().Slice()) < 1 {
			return errors.New("logs command lacks container ID")
		}
		containerID := c.Args().Get(0)
		container.OutputContainerLog(containerID)
		return nil
	},
}

var execCommand = cli.Command{
	Name:  "exec",
	Usage: "enter a container and exec a command",
	Action: func(c *cli.Context) error {
		if os.Getenv(EnvExecPid) != "" {
			log.Infof("pid callback pid %v", os.Getgid())
			return nil
		}

		if len(c.Args().Slice()) < 2 {
			return errors.New("exec command missing container id or command")
		}
		containerID := c.Args().Get(0)
		var cmdArr []string
		cmdArr = append(cmdArr, c.Args().Tail()...)
		ExecContainer(containerID, cmdArr)
		return nil
	},
}

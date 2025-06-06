package main

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"mydocker/cgroups/resource"
	"mydocker/container"
	"mydocker/network"
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
		&cli.StringSliceFlag{
			Name:  "e",
			Usage: "set environment, e.g., -e name=mydocker -e foo=bar",
		},
		&cli.StringFlag{
			Name:  "net",
			Usage: "container network name, e.g., -net testnet",
		},
		&cli.StringSliceFlag{
			Name:  "p",
			Usage: "port mapping, e.g., -p 8080:80 -p 6000:60",
		},
	},
	Action: func(c *cli.Context) error {
		// c.Args() 不包括flag相关参数
		if c.Args().Len() < 2 {
			return errors.New("missing container command or image")
		}

		tty := c.Bool("it")
		detach := c.Bool("d")
		// 实际运行中只依靠-it来判断是否后台运行
		if tty && detach {
			return fmt.Errorf("it and d parameter can not be both provided")
		}

		resCfg := &resource.ResourceConfig{
			MemoryLimit: c.String("mem"),
			CpuSet:      c.String("cpuset"),
			CpuCfsQuota: c.Int("cpu"),
		}

		volume := c.String("v")
		containerName := c.String("name")
		envSlice := c.StringSlice("e")
		imageName := c.Args().First()
		netName := c.String("net")
		portMapping := c.StringSlice("p")

		Run(tty, c.Args().Tail(), envSlice, resCfg, volume, containerName, imageName, netName, portMapping)

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
		if len(c.Args().Slice()) < 2 {
			return fmt.Errorf("commit missing container id or image name")
		}
		containerID := c.Args().Get(0)
		imageName := c.Args().Get(1)
		container.CommitContainer(containerID, imageName)

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

var stopCommand = cli.Command{
	Name:  "stop",
	Usage: "stop a container, e.g., mydocker stop {containerID}",
	Action: func(c *cli.Context) error {
		if len(c.Args().Slice()) < 1 {
			return errors.New("stop command missing container id")
		}
		containerID := c.Args().Get(0)
		stopContainer(containerID)
		return nil
	},
}

var removeCommand = cli.Command{
	Name:  "rm",
	Usage: "remove container with container ID, e.g., mydocker rm {containerID}",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "f",
			Usage: "force delete running container",
		},
	},
	Action: func(c *cli.Context) error {
		if len(c.Args().Slice()) < 1 {
			return errors.New("rm command missing container id")
		}
		containerID := c.Args().Get(0)
		force := c.Bool("f")
		removeContainer(containerID, force)
		return nil
	},
}

var networkCommand = cli.Command{
	Name:  "network",
	Usage: "container network commands",
	Subcommands: []*cli.Command{
		{
			Name:  "create",
			Usage: "create a network",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "driver",
					Usage: "network driver",
				},
				&cli.StringFlag{
					Name:  "subnet",
					Usage: "subnet CIDR",
				},
			},
			Action: func(c *cli.Context) error {
				if len(c.Args().Slice()) < 1 {
					return errors.New("missing network name")
				}

				driver := c.String("driver")
				subnet := c.String("subnet")
				networkName := c.Args().First()

				err := network.CreateNetwork(driver, subnet, networkName)
				if err != nil {
					return errors.Wrapf(err, "create network %s failed", networkName)
				}

				return nil
			},
		},
		{
			Name:  "list",
			Usage: "list container networks",
			Action: func(c *cli.Context) error {
				network.ListNetwork()
				return nil
			},
		},
		{
			Name:  "remove",
			Usage: "remove container networks",
			Action: func(c *cli.Context) error {
				if len(c.Args().Slice()) < 1 {
					return errors.New("missing network name")
				}
				name := c.Args().First()
				err := network.DeleteNetwork(name)
				if err != nil {
					return errors.WithMessagef(err, "remove network %s failed", name)
				}
				return nil
			},
		},
	},
}

var startCommand = cli.Command{
	Name:  "start",
	Usage: "start a stopped container and run it in background",
	Action: func(c *cli.Context) error {
		if os.Getenv(EnvStart) != "" {
			container.StartContainerInitProcess()
		}

		if len(c.Args().Slice()) < 1 {
			return errors.New("start command missing container id")
		}

		containerID := c.Args().First()
		startContainer(containerID)

		return nil
	},
}

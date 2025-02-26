package main

import (
	"os"
	"strconv"
	"strings"

	"mydocker/cgroups"
	"mydocker/cgroups/resource"
	"mydocker/container"
	"mydocker/network"

	"github.com/sirupsen/logrus"
)

func Run(tty bool, cmdArray, envSlice []string, res *resource.ResourceConfig, volume, containerName string,
	imageName, net string, portMapping []string) {

	containerID := container.GenerateContainerID()

	parent, wPipe := container.NewParentProcessPipe(tty, volume, containerID, imageName, envSlice)
	if err := parent.Start(); err != nil {
		logrus.Error(err.Error())
		return
	}

	// cgroup控制资源
	logrus.Infof("child proc: %d", parent.Process.Pid)
	cgroupManager := cgroups.NewCgroupManager("mydocker-cgroup")
	defer cgroupManager.Destroy()
	_ = cgroupManager.Set(res)
	_ = cgroupManager.Apply(parent.Process.Pid, res)

	var containerIP string
	// 配置网络
	if net != "" {
		containerINFO := &container.Info{
			Id:          containerID,
			Name:        containerName,
			Pid:         strconv.Itoa(parent.Process.Pid),
			PortMapping: portMapping,
		}
		ip, err := network.Connect(net, containerINFO)
		if err != nil {
			logrus.Errorf("connect to net %s failed, %v", net, err)
			container.DelWorkSpace(containerID, volume)
			err := container.DelContainerInfo(containerID)
			if err != nil {
				logrus.Error(err)
			}
			return
		}
		containerIP = ip.String()
	}

	info, err := container.RecordContainerInfo(parent.Process.Pid, cmdArray, containerName, containerID, volume, net, containerIP, portMapping)
	if err != nil {
		logrus.Errorf("record container info failed, err: %v", err)
		return
	}

	// 父进程没有向Pipe输入数据时，子进程会阻塞
	sendInitCmds(cmdArray, wPipe)

	// 如果tty，父进程就需要等到容器进程结束再退出
	// 否则直接退出
	if tty {
		_ = parent.Wait()
		container.DelWorkSpace(containerID, volume)
		if err := container.DelContainerInfo(containerID); err != nil {
			logrus.Error(err)
		}
		if net != "" {
			if err = network.Disconnect(info); err != nil {
				logrus.Errorf("%+v", err)
			}
		}
	}

}

func sendInitCmds(cmdArray []string, writePipe *os.File) {
	cmd := strings.Join(cmdArray, " ")
	logrus.Infof("Container init command: %s", cmd)
	_, _ = writePipe.Write([]byte(cmd))
	_ = writePipe.Close()
}

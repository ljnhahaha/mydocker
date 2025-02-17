package main

import (
	"os"
	"strings"

	"mydocker/cgroups"
	"mydocker/cgroups/subsystems"
	"mydocker/container"

	"github.com/sirupsen/logrus"
)

func Run(tty bool, cmdArray, envSlice []string, res *subsystems.ResourceConfig, volume, containerName, imageName string) {
	containerID := container.GenerateContainerID()

	parent, wPipe := container.NewParentProcessPipe(tty, volume, containerID, imageName, envSlice)
	if err := parent.Start(); err != nil {
		logrus.Error(err.Error())
		return
	}

	err := container.RecordContainerInfo(parent.Process.Pid, cmdArray, containerName, containerID, volume)
	if err != nil {
		logrus.Errorf("record container info failed, err: %v", err)
		return
	}

	// cgroup控制资源
	logrus.Infof("child proc: %d", parent.Process.Pid)
	cgroupManager := cgroups.NewCgroupManager("mydocker-cgroup")
	defer cgroupManager.Destroy()
	_ = cgroupManager.Set(res)
	_ = cgroupManager.Apply(parent.Process.Pid, res)

	// 父进程没有向Pipe输入数据时，子进程会阻塞
	sendInitCmds(cmdArray, wPipe)

	// 如果tty，父进程就需要等到容器进程结束再退出
	// 否则直接退出
	if tty {
		_ = parent.Wait()
		container.DelWorkSpace(containerID, volume)
		err := container.DelContainerInfo(containerID)
		if err != nil {
			logrus.Error(err)
		}
	}

}

func sendInitCmds(cmdArray []string, writePipe *os.File) {
	cmd := strings.Join(cmdArray, " ")
	logrus.Infof("Container init command: %s", cmd)
	_, _ = writePipe.Write([]byte(cmd))
	_ = writePipe.Close()
}

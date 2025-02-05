package main

import (
	"os"
	"strings"

	"mydocker/cgroups"
	"mydocker/cgroups/subsystems"
	"mydocker/container"

	"github.com/sirupsen/logrus"
)

// func Run(tty bool, cmd string) {
// 	parent := container.NewParentProcess(tty, cmd)

// 	if err := parent.Start(); err != nil {
// 		logrus.Error(err.Error())
// 	}

// 	_ = parent.Wait()
// 	os.Exit(-1)
// }

func RunCmds(tty bool, cmdArray []string, res *subsystems.ResourceConfig, volume string) {
	parent, wPipe := container.NewParentProcessPipe(tty, volume)

	if err := parent.Start(); err != nil {
		logrus.Error(err.Error())
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
		container.DelWorkSpace("/root/myoverlayfs", volume)
	}

}

func sendInitCmds(cmdArray []string, writePipe *os.File) {
	cmd := strings.Join(cmdArray, " ")
	logrus.Infof("Container init command: %s", cmd)
	_, _ = writePipe.Write([]byte(cmd))
	_ = writePipe.Close()
}

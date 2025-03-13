package main

import (
	"encoding/json"
	"mydocker/container"
	"mydocker/utils"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/sirupsen/logrus"
)

const (
	EnvStart = "mydocker_start"
)

func startContainer(containerID string) {
	infoDir := filepath.Join(container.InfoLoc, containerID)
	_, err := os.Stat(infoDir)
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Errorf("can not find container [%s]", containerID)
		} else {
			logrus.Errorf("can not check container [%s] status", containerID)
		}
	}

	infoFile := filepath.Join(infoDir, container.ConfigName)
	containerInfo, err := getInfoByContainerID(containerID)
	if err != nil {
		logrus.Errorf("read container [%s] info failed, err: %v", containerID, err)
	}

	// 创建子进程
	cmd := exec.Command("/proc/self/exe", "start")

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWNET | syscall.CLONE_NEWNS | syscall.CLONE_NEWPID |
			syscall.CLONE_NEWIPC,
	}

	logFile := filepath.Join(container.InfoLoc, containerID, container.GetLogFile(containerID))
	logfd, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logrus.Errorf("open log file failed, err: %v", err)
		return
	}
	cmd.Stdout = logfd
	cmd.Stderr = logfd
	cmd.Dir = utils.GetMerged(containerID)
	cmd.Env = append(os.Environ(), "mydocker_start=true")

	if err := cmd.Start(); err != nil {
		logrus.Errorf("start container [%s] failed, err; %v", containerID, err)
		return
	}

	// 修改容器状态
	containerInfo.Pid = strconv.Itoa(cmd.Process.Pid)
	containerInfo.Status = container.RUNNING
	file, err := os.OpenFile(infoFile, os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		logrus.Errorf("open container information file failed, err: %v", err)
	}
	defer file.Close()

	if err = json.NewEncoder(file).Encode(containerInfo); err != nil {
		logrus.Errorf("dump into json failed, err: %v", err)
		return
	}
}

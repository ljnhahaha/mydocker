package main

import (
	"encoding/json"
	"fmt"
	"mydocker/container"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	_ "mydocker/nsenter"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	EnvExecPid = "mydocker_pid"
	EnvExecCmd = "mydocker_cmd"
)

func ExecContainer(containerID string, cmdArr []string) {
	pid, err := getPidByContainerID(containerID)
	if err != nil {
		log.Error(err)
	}

	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmdStr := strings.Join(cmdArr, " ")
	log.Infof("container pid: %s, cmd: %s", pid, cmdStr)
	// 通过环境变量为cgo中constructor函数传递参数
	// os.Setenv 设置的环境变量只会影响当前进程和子进程
	_ = os.Setenv(EnvExecPid, pid)
	_ = os.Setenv(EnvExecCmd, cmdStr)

	// 将指定容器内的环境变量传递给新进程
	containerEnvs, err := getEnvByPID(pid)
	if err != nil {
		log.Errorf("get env failed, %v", err)
	} else {
		cmd.Env = append(os.Environ(), containerEnvs...)
	}

	if err = cmd.Run(); err != nil {
		log.Errorf("exec container %s err %v", containerID, err)
	}
}

func getPidByContainerID(containerID string) (string, error) {
	infoFilePath := filepath.Join(container.InfoLoc, containerID, container.ConfigName)
	content, err := os.ReadFile(infoFilePath)
	if err != nil {
		return "", errors.WithMessagef(err, "read config file %s failed", infoFilePath)
	}

	containerInfo := new(container.Info)
	if err := json.Unmarshal(content, containerInfo); err != nil {
		return "", errors.WithMessage(err, "json unmarshal failed")
	}

	return containerInfo.Pid, nil
}

func getEnvByPID(pid string) ([]string, error) {
	path := fmt.Sprintf("/proc/%s/environ", pid)
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	envs := strings.Split(string(content), "\u0000")
	return envs, nil
}

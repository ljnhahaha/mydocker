package main

import (
	"encoding/json"
	"fmt"
	"mydocker/container"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	log "github.com/sirupsen/logrus"
)

func stopContainer(containerID string) {
	// 查询容器信息
	containerInfo, err := getInfoByContainerID(containerID)
	if err != nil {
		log.Error(err)
		return
	}
	pidInt, err := strconv.Atoi(containerInfo.Pid)
	if err != nil {
		log.Errorf("convert string to int failed, %v", err)
		return
	}

	// 发送SIGTERM信号
	if err = syscall.Kill(pidInt, syscall.SIGTERM); err != nil {
		log.Errorf("kill process %d failed, %v", pidInt, err)
		return
	}

	// 修改容器信息: 1. 修改容器状态 2. 清空PID
	containerInfo.Status = container.STOP
	containerInfo.Pid = ""
	newContentBytes, err := json.Marshal(containerInfo)
	if err != nil {
		log.Errorf("json marshal failed, %v", err)
		return
	}

	// 更新容器信息文件
	infoFilePath := filepath.Join(container.InfoLoc, containerID, container.ConfigName)
	if err = os.WriteFile(infoFilePath, newContentBytes, 0622); err != nil {
		log.Errorf("write file %s failed, %v", infoFilePath, err)
	}
}

func getInfoByContainerID(containerID string) (*container.Info, error) {
	infoFilePath := filepath.Join(container.InfoLoc, containerID, container.ConfigName)
	content, err := os.ReadFile(infoFilePath)
	if err != nil {
		return nil, fmt.Errorf("read file %s failed, %v", infoFilePath, err)
	}
	info := new(container.Info)
	if err = json.Unmarshal(content, info); err != nil {
		return nil, fmt.Errorf("unmarshal json failed, %v", err)
	}
	return info, nil
}

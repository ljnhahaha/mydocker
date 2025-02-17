package main

import (
	"mydocker/container"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	log "github.com/sirupsen/logrus"
)

func removeContainer(containerID string, force bool) {
	containerInfo, err := getInfoByContainerID(containerID)
	if err != nil {
		log.Error(err)
		return
	}

	switch containerInfo.Status {
	case container.STOP:
		dirPath := filepath.Join(container.InfoLoc, containerID)
		if err = os.RemoveAll(dirPath); err != nil {
			log.Errorf("remove dir %s failed, %v", dirPath, err)
			return
		}
		container.DelWorkSpace(containerID, containerInfo.Volume)
	case container.RUNNING:
		if !force {
			log.Errorf("container {%s} is running, please stop it at first or use [-f]", containerID)
			return
		}
		pidInt, err := strconv.Atoi(containerInfo.Pid)
		if err != nil {
			log.Errorf("convert string to int failed, %v", err)
			return
		}

		if err = syscall.Kill(pidInt, syscall.SIGTERM); err != nil {
			log.Errorf("kill process %d failed, %v", pidInt, err)
			return
		}
		infoDirPath := filepath.Join(container.InfoLoc, containerID)
		if err = os.RemoveAll(infoDirPath); err != nil {
			log.Errorf("remove dir %s failed, %v", infoDirPath, err)
			return
		}
		container.DelWorkSpace(containerID, containerInfo.Volume)
	default:
		log.Errorf("couldn't remove container, invalid status: %s", containerInfo.Status)
		return
	}
}

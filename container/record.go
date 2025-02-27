// record container info
package container

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"mydocker/utils"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

func randStringBytes(n int) string {
	letterBytes := "1234567890"
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rng.Intn(len(letterBytes))]
	}

	return string(b)
}

func RecordContainerInfo(containerPID int, commandArray []string, containerName, containerID, volume, net, ip, image string,
	portMapping []string) (*Info, error) {
	if containerName == "" {
		containerName = containerID
	}
	command := strings.Join(commandArray, " ")

	containerInfo := &Info{
		Pid:         strconv.Itoa(containerPID),
		Id:          containerID,
		Name:        containerName,
		Command:     command,
		CreatedTime: time.Now().Format("2006-01-02 15:04:05"),
		Status:      RUNNING,
		Volume:      volume,
		NetworkName: net,
		IP:          ip,
		PortMapping: portMapping,
		Image:       image,
	}

	jsonBytes, err := json.Marshal(containerInfo)
	if err != nil {
		return containerInfo, errors.WithMessage(err, "container info marshal failed")
	}
	jsonStr := string(jsonBytes)

	// 容器信息路径: InfoLoc/{containerID}/
	dirPath := filepath.Join(InfoLoc, containerID)
	exists, _ := utils.PathExist(dirPath)
	if !exists {
		if err = os.MkdirAll(dirPath, 0622); err != nil {
			return containerInfo, errors.WithMessagef(err, "mkdir %s failed", dirPath)
		}
	}

	fileName := filepath.Join(dirPath, ConfigName)
	file, err := os.Create(fileName)
	if err != nil {
		return containerInfo, errors.WithMessagef(err, "create file %s failed", fileName)
	}
	defer file.Close()

	if _, err = file.WriteString(jsonStr); err != nil {
		return containerInfo, errors.WithMessagef(err, "write container info to file %s failed", fileName)
	}

	return containerInfo, nil
}

func GenerateContainerID() string {
	return randStringBytes(IDLength)
}

func DelContainerInfo(containerID string) error {
	dirPath := filepath.Join(InfoLoc, containerID)
	if err := os.RemoveAll(dirPath); err != nil {
		return errors.WithMessagef(err, "del container info at path: %s failed", dirPath)
	}
	return nil
}

func GetLogFile(containerID string) string {
	logFile := fmt.Sprintf(LogFile, containerID)
	return logFile
}

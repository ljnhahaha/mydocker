package container

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

func OutputContainerLog(containerID string) {
	logFilePath := filepath.Join(InfoLoc, containerID, GetLogFile(containerID))
	file, err := os.Open(logFilePath)
	if err != nil {
		logrus.Errorf("open log file %s failed", logFilePath)
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		logrus.Errorf("read log file %s failed", logFilePath)
		return
	}

	_, err = fmt.Fprint(os.Stdout, string(content))
	if err != nil {
		logrus.Error("output log info failed")
		return
	}
}

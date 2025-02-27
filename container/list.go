package container

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/sirupsen/logrus"
)

// 遍历读取InfoLoc下的文件并格式化输出
func ListContainers() {
	files, err := os.ReadDir(InfoLoc)
	if err != nil {
		logrus.Errorf("read file %s failed, err: %v", files, err)
	}

	infoList := make([]*Info, 0, len(files))
	for _, f := range files {
		containerInfo, err := getContainerInfo(f)
		if err != nil {
			logrus.Errorf("read container info err: %v", err)
			continue
		}
		infoList = append(infoList, containerInfo)
	}

	// 使用tabwriter进行格式化输出
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	_, err = fmt.Fprint(w, "ID\tNAME\tIMAGE\tPID\tIP\tSTATUS\tCOMMAND\tCREATED\n")
	if err != nil {
		logrus.Errorf("Fprint err: %v", err)
	}

	for _, info := range infoList {
		_, err = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			info.Id,
			info.Name,
			info.Image,
			info.Pid,
			info.IP,
			info.Status,
			info.Command,
			info.CreatedTime)
		if err != nil {
			logrus.Errorf("Fprint err: %v", err)
		}
	}

	if err = w.Flush(); err != nil {
		logrus.Errorf("flush err: %v", err)
	}
}

func getContainerInfo(f os.DirEntry) (*Info, error) {
	configFilePath := filepath.Join(InfoLoc, f.Name(), ConfigName)
	content, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}

	info := new(Info)
	if err = json.Unmarshal(content, info); err != nil {
		return nil, err
	}

	return info, nil
}

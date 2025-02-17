package container

import (
	"fmt"
	"mydocker/utils"
	"os/exec"

	"github.com/sirupsen/logrus"
)

func CommitContainer(containerID, imageName string) {
	mntPath := utils.GetMerged(containerID)
	imageTar := utils.GetImage(imageName)

	exists, err := utils.PathExist(imageTar)
	if err != nil {
		logrus.Errorf("cannot check if %s exists or not, %v", imageTar, err)
	}
	if exists {
		logrus.Errorf("file %s has already existed", imageTar)
	}

	fmt.Printf("commit container to: %s\n", imageTar)
	// -C 切换到指定目录，这样只会打包目录内的文件
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", mntPath, ".").CombinedOutput(); err != nil {
		logrus.Errorf("commit failed, err: %v", err)
	}
}

package container

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

func CommitContainer(imageName string) {
	mntPath := "/root/myoverlayfs/merged"
	imageTar := filepath.Join("/root", imageName+".tar")
	fmt.Printf("commit container to: %s\n", imageTar)

	// -C 切换到指定目录，这样只会打包目录内的文件
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", mntPath, ".").CombinedOutput(); err != nil {
		logrus.Errorf("commit failed, err: %v", err)
	}
}

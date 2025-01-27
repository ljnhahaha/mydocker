package container

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

// volumeParse 通过冒号分割解析volume目录，比如 -v /tmp:/tmp
func volumeParse(volume string) (string, string, error) {
	parts := strings.Split(volume, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid volume %s", volume)
	}

	srcPath, targetPath := parts[0], parts[1]
	if srcPath == "" || targetPath == "" {
		return "", "", fmt.Errorf("invalid volume %s", volume)
	}

	return srcPath, targetPath, nil
}

func mountVolume(mntPath, hostPath, containerPath string) {
	// 创建宿主机目录
	if err := os.Mkdir(hostPath, 0777); err != nil {
		logrus.Infof("make host dir fail, dir: %s, err: %v", hostPath, err)
	}

	// 拼接出容器fs中的路径对应宿主机路径
	containerPathInHost := filepath.Join(mntPath, containerPath)
	if err := os.Mkdir(containerPathInHost, 0777); err != nil {
		logrus.Infof("make container dir fail, dir: %s, err: %v", containerPathInHost, err)
	}

	// mount -o bind /hostPath /containerPath
	// 将 hostPath 挂载到 containerPath
	cmd := exec.Command("mount", "-o", "bind", hostPath, containerPathInHost)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		logrus.Errorf("mount volume fail, err: %v", err)
	}
}

func umountVolume(mntPath, containerPath string) {
	containerPathInHost := filepath.Join(mntPath, containerPath)
	cmd := exec.Command("umount", containerPathInHost)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		logrus.Errorf("umount volume fail, err: %v", err)
	}
}

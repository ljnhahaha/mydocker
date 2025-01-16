package container

import (
	"mydocker/utils"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// 创建挂载OverlayFS所需的文件并挂载OverlayFS
func NewWorkSpace(rootPath string) {
	createLower(rootPath)
	createDirs(rootPath)
	mountOverlayFS(rootPath)
}

// 将 busybox 挂载为 overlayfs 的 lower filesystem
func createLower(rootPath string) {
	busyboxPath := filepath.Join(rootPath, "busybox/")
	busyboxTarPath := filepath.Join(rootPath, "busybox.tar")

	exist, err := utils.PathExist(busyboxPath)
	if err != nil {
		logrus.Info(err)
	}

	// 如果不存在busybox路径，则将busybox.tar解压到对应路径
	if !exist {
		if err = os.Mkdir(busyboxPath, 0777); err != nil {
			logrus.Errorf("overlay lower mkdir fail, %v", err)
		}

		if _, err = exec.Command("tar", "-xvf", busyboxTarPath, "-C", busyboxPath).CombinedOutput(); err != nil {
			logrus.Errorf("untar %s fail, %v", busyboxTarPath, err)
		}
	}
}

// 创建挂载 overlayfs 中 upper filesystem & work filesystem & mergerd filesystem的文件夹
func createDirs(rootPath string) {
	upperPath := filepath.Join(rootPath, "upper/")
	if err := os.Mkdir(upperPath, 0777); err != nil {
		logrus.Errorf("overlay upper mkdir fail, %v", err)
	}

	workPath := filepath.Join(rootPath, "work/")
	if err := os.Mkdir(workPath, 0777); err != nil {
		logrus.Errorf("overlay work mkdir fail, %v", err)
	}

	mergedPath := filepath.Join(rootPath, "merged/")
	if err := os.Mkdir(mergedPath, 0777); err != nil {
		logrus.Errorf("overlay merged mkdir fail, %v", err)
	}
}

// 挂载OverlayFS
// mount -t overlay overlay -o lowerdir=lower1:lower2:lower3,upperdir=upper,workdir=work mergedir
func mountOverlayFS(rootPath string) {
	dirArgs := "lowerdir=" + rootPath + "/busybox" + ",upperdir=" + rootPath + "/upper" + ",workdir=" + rootPath + "/work"
	mergedPath := filepath.Join(rootPath, "merged")
	cmd := exec.Command("mount", "-t", "overlay", "overlay", "-o", dirArgs, mergedPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		logrus.Error(err)
	}

}

// umount OverlayFS 并且删除 upper, work, merged 文件夹
func DelWorkSpace(rootPath string) {
	umountOverlayFS(rootPath)
	delDirs(rootPath)
}

// umount overlayfs and delete merged dir
func umountOverlayFS(rootPath string) {
	mntPath := filepath.Join(rootPath, "merged")
	cmd := exec.Command("umount", mntPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Error(err)
	}
}

func delDirs(rootPath string) {
	upperPath := filepath.Join(rootPath, "upper/")
	if err := os.RemoveAll(upperPath); err != nil {
		logrus.Errorf("overlay upper remove fail, %v", err)
	}

	workPath := filepath.Join(rootPath, "work/")
	if err := os.RemoveAll(workPath); err != nil {
		logrus.Errorf("overlay work remove fail, %v", err)
	}

	mergedPath := filepath.Join(rootPath, "merged/")
	if err := os.RemoveAll(mergedPath); err != nil {
		logrus.Errorf("overlay merged remove fail, %v", err)
	}
}

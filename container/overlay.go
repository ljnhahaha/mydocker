package container

import (
	"mydocker/utils"
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"
)

// 创建挂载OverlayFS所需的文件并挂载OverlayFS
// 如果指定了volume还需要挂载volume
func NewWorkSpace(containerID, imageName, volume string) {
	rootPath := utils.GetRoot(containerID)
	if err := os.Mkdir(rootPath, 0777); err != nil {
		logrus.Errorf("mkdir %s failed, %v", rootPath, err)
		return
	}

	createLower(containerID, imageName)
	createDirs(containerID)
	mountOverlayFS(containerID)

	if volume != "" {
		mntPath := utils.GetMerged(containerID)
		hostPath, containerPath, err := volumeParse(volume)
		if err != nil {
			logrus.Errorf("parse volume path fail, err: %v", err)
			return
		}
		mountVolume(mntPath, hostPath, containerPath)
	}
}

// 将指定镜像挂载为 overlayfs 的 lower filesystem
func createLower(containerID, imageName string) {
	lowerPath := utils.GetLower(containerID)
	imageTarPath := utils.GetImage(imageName)

	exist, err := utils.PathExist(lowerPath)
	if err != nil {
		logrus.Info(err)
	}

	// 如果不存在busybox路径，则将busybox.tar解压到对应路径
	if !exist {
		if err = os.Mkdir(lowerPath, 0777); err != nil {
			logrus.Errorf("overlay lower mkdir fail, %v", err)
		}

		if _, err = exec.Command("tar", "-xvf", imageTarPath, "-C", lowerPath).CombinedOutput(); err != nil {
			logrus.Errorf("untar %s fail, %v", imageTarPath, err)
		}
	}
}

// 创建挂载 overlayfs 中 upper filesystem & work filesystem & mergerd filesystem的文件夹
func createDirs(containerID string) {
	upperPath := utils.GetUpper(containerID)
	if err := os.Mkdir(upperPath, 0777); err != nil {
		logrus.Errorf("overlay upper mkdir fail, %v", err)
	}

	workPath := utils.GetWork(containerID)
	if err := os.Mkdir(workPath, 0777); err != nil {
		logrus.Errorf("overlay work mkdir fail, %v", err)
	}

	mergedPath := utils.GetMerged(containerID)
	if err := os.Mkdir(mergedPath, 0777); err != nil {
		logrus.Errorf("overlay merged mkdir fail, %v", err)
	}
}

// 挂载OverlayFS
// mount -t overlay overlay -o lowerdir=lower1:lower2:lower3,upperdir=upper,workdir=work mergedir
func mountOverlayFS(containerID string) {
	dirArgs := utils.GetOverlayFSDir(containerID)
	mergedPath := utils.GetMerged(containerID)
	cmd := exec.Command("mount", "-t", "overlay", "overlay", "-o", dirArgs, mergedPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		logrus.Error(err)
	}

}

// 先umount volume，再umount OverlayFS 并且删除 upper, work, merged 文件夹
// 否则会导致 volume 中的文件也被删除
func DelWorkSpace(containerID, volume string) {
	if volume != "" {
		mntPath := utils.GetMerged(containerID)
		_, containerPath, err := volumeParse(volume)
		if err != nil {
			logrus.Errorf("parse volume path fail, err: %v", err)
			return
		}
		umountVolume(mntPath, containerPath)
	}
	umountOverlayFS(containerID)
	delDirs(containerID)
}

// umount overlayfs and delete merged dir
func umountOverlayFS(containerID string) {
	mntPath := utils.GetMerged(containerID)
	cmd := exec.Command("umount", mntPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Error(err)
	}
}

func delDirs(containerID string) {
	// upperPath := utils.GetUpper(containerID)
	// if err := os.RemoveAll(upperPath); err != nil {
	// 	logrus.Errorf("overlay upper remove fail, %v", err)
	// }

	// workPath := utils.GetWork(containerID)
	// if err := os.RemoveAll(workPath); err != nil {
	// 	logrus.Errorf("overlay work remove fail, %v", err)
	// }

	// mergedPath := utils.GetMerged(containerID)
	// if err := os.RemoveAll(mergedPath); err != nil {
	// 	logrus.Errorf("overlay merged remove fail, %v", err)
	// }
	rootPath := utils.GetRoot(containerID)
	if err := os.RemoveAll(rootPath); err != nil {
		logrus.Errorf("remove path %s failed, %v", rootPath, err)
	}
}

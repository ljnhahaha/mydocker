// 容器进程创建前的准备工作，包括
// 1. 初始化命令cmd
// 2. 创建Namespace进行视图隔离
// 3. 准备Overlayfs相关文件挂载

package container

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"mydocker/utils"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	RUNNING    = "running"
	STOP       = "stopped"
	EXIST      = "existed"
	InfoLoc    = "/var/lib/mydocker/containers/"
	ConfigName = "config.json"
	IDLength   = 10
	LogFile    = "%s-json.log"
)

type Info struct {
	Pid         string   `json:"pid"`         // 容器的init进程在宿主机上的PID
	Id          string   `json:"id"`          // 容器的ID
	Name        string   `json:"name"`        // 容器名
	Command     string   `json:"command"`     // 容器的启动命令
	CreatedTime string   `json:"createdtime"` // 容器的创建时间
	Image       string   `json:"image"`       // 容器镜像
	Status      string   `json:"status"`      // 容器状态
	Volume      string   `json:"volume"`      // 容器挂载的volume
	NetworkName string   `json:"network"`     // 容器所在网络名
	IP          string   `json:"ip"`          // 容器IP
	PortMapping []string `json:"portmapping"` // 容器端口映射
}

// Instantiate a child process initialization command
// func NewParentProcess(tty bool, command string) *exec.Cmd {
// 	args := []string{"init", command}

// 	// /proc/self/exe 表示当前正在运行的可执行文件的路径（符号链接到当前进程的可执行文件)
// 	cmd := exec.Command("/proc/self/exe", args...)

// 	// 利用 Namespace 进行资源隔离
// 	cmd.SysProcAttr = &syscall.SysProcAttr{
// 		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWNET | syscall.CLONE_NEWNS | syscall.CLONE_NEWPID |
// 			syscall.CLONE_NEWIPC,
// 	}

// 	if tty {
// 		cmd.Stdin = os.Stdin
// 		cmd.Stdout = os.Stdout
// 		cmd.Stderr = os.Stderr
// 	}

// 	return cmd
// }

// 创建子进程启动命令，通过Pipe，父进程向子进程传递参数
func NewParentProcessPipe(tty bool, volume, containerID, imageName string, envSlice []string) (*exec.Cmd, *os.File, error) {
	rPipe, wPipe, err := os.Pipe()

	if err != nil {
		logrus.Error(err)
	}

	// /proc/self/exe 表示当前正在运行的可执行文件的路径（符号链接到当前进程的可执行文件)
	cmd := exec.Command("/proc/self/exe", "init")

	// 利用 Namespace 进行资源隔离
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWNET | syscall.CLONE_NEWNS | syscall.CLONE_NEWPID |
			syscall.CLONE_NEWIPC,
	}

	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		// 将执行输出记录到指定的日志文件中
		dirPath := filepath.Join(InfoLoc, containerID)
		exists, _ := utils.PathExist(dirPath)
		if !exists {
			if err := os.MkdirAll(dirPath, 0622); err != nil {
				return nil, nil, errors.Wrapf(err, "mkdir %s failed", dirPath)
			}
		}

		stdLogFilePath := filepath.Join(dirPath, GetLogFile(containerID))
		stdLogFile, err := os.Create(stdLogFilePath)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "create log file %s failed", stdLogFilePath)
		}

		cmd.Stdout = stdLogFile
		cmd.Stderr = stdLogFile
	}

	// 通过ExtraFile将rPipe传递给子进程
	cmd.ExtraFiles = []*os.File{rPipe}

	// File Systems
	err = NewWorkSpace(containerID, imageName, volume)
	if err != nil {
		return nil, nil, errors.Wrap(err, "create work space failed")
	}
	// Specify work dir
	cmd.Dir = utils.GetMerged(containerID)

	cmd.Env = append(os.Environ(), envSlice...)

	return cmd, wPipe, nil
}

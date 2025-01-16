// 容器进程创建前的准备工作，包括
// 1. 初始化命令cmd
// 2. 创建Namespace进行视图隔离
// 3. 准备Overlayfs相关文件挂载

package container

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/sirupsen/logrus"
)

// Instantiate a child process initialization command
func NewParentProcess(tty bool, command string) *exec.Cmd {
	args := []string{"init", command}

	// /proc/self/exe 表示当前正在运行的可执行文件的路径（符号链接到当前进程的可执行文件)
	cmd := exec.Command("/proc/self/exe", args...)

	// 利用 Namespace 进行资源隔离
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWNET | syscall.CLONE_NEWNS | syscall.CLONE_NEWPID |
			syscall.CLONE_NEWIPC,
	}

	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	return cmd
}

// 通过Pipe，父进程向子进程传递参数
func NewParentProcessPipe(tty bool) (*exec.Cmd, *os.File) {
	rPipe, wPipe, err := os.Pipe()

	if err != nil {
		logrus.Error(err.Error())
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
	}

	cmd.ExtraFiles = []*os.File{rPipe}
	rootPath := "/root/myoverlayfs"
	NewWorkSpace(rootPath)
	cmd.Dir = "/root/myoverlayfs/merged"

	return cmd, wPipe
}

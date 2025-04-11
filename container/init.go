// 容器初始化时需要进行的操作，此时容器进程已经创建

package container

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func RunContainerInitProcess() error {
	setUpMount()

	cmdArray := readUserCommand()
	// 根据命令查找环境变量，找到可执行文件
	path, err := exec.LookPath(cmdArray[0])
	if err != nil {
		log.Error(err)
	}

	// 利用syscall.Exec()方法调用execve系统调用，覆盖当前进程，使容器中运行的
	// command 成为PID 1 (实际上运行的第一个command是 mydocker init ...)
	// 第一个参数为可执行二进制文件路径， 如 "/bin/ls"
	// 第二个参数为具体命令 []string, 如 ["ls", "./"]
	// 第三个参数为环境变量
	if err := syscall.Exec(path, cmdArray[0:], os.Environ()); err != nil {
		log.Errorf("RunContainerInitProcess exec: %s", err.Error())
	}

	return nil
}

func readUserCommand() []string {
	pipe := os.NewFile(uintptr(3), "pipe")
	// Pipe为空时，子进程阻塞
	msg, err := io.ReadAll(pipe)
	if err != nil {
		log.Errorf("error while reading from pipe %v", err)
	}

	msgStr := string(msg)
	return strings.Split(msgStr, " ")
}

// Mount "/proc", make process information visible.
func setUpMount() {
	wd, err := os.Getwd()
	if err != nil {
		log.Error(err)
		return
	}
	log.Infof("Current work directory is %s", wd)

	// systemd 加入linux之后, mount namespace 就变成 shared by default, 所以你必须显示
	// 声明你要这个新的mount namespace独立。
	// 如果不先做 private mount，会导致挂载事件外泄，后续执行 pivotRoot 会出现 invalid argument 错误
	syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")

	err = pivotRoot(wd)
	if err != nil {
		log.Error(err)
	}

	// mount时禁止以下行为
	// 重新挂载 procfs
	defaltMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	_ = syscall.Mount("proc", "/proc", "proc", uintptr(defaltMountFlags), "")

	// 由于前面 pivotRoot 切换了 rootfs，因此这里重新 mount 一下 /dev 目录
	// tmpfs 是基于 件系 使用 RAM、swap 分区来存储。
	// 不挂载 /dev，会导致容器内部无法访问和使用许多设备，这可能导致系统无法正常工作
	syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755")
}

// 为当前容器挂载新的rootfs
func pivotRoot(root string) error {
	// pivot_root 要求 new_root 是一个挂载点
	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return errors.Wrap(err, "mount rootfs to itself")
	}

	// 为旧rootfs创建一个临时挂载点 put_old
	pivotDir := filepath.Join(root, ".pivot_root")
	if err := os.Mkdir(pivotDir, 0777); err != nil {
		return err
	}

	// PivotRoot系统调用将rootfs换到新的路径，并且将原来的rootfs挂载到一个文件夹中
	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return errors.WithMessagef(err, "pivotRoot fail, newroot: %s, putold: %s", root, pivotDir)
	}

	if err := syscall.Chdir("/"); err != nil {
		return errors.WithMessage(err, "change to / fail")
	}

	// 将旧rootfs umount以便删除
	pivotDir = filepath.Join("/", ".pivot_root")
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return errors.WithMessage(err, "umount pivot dir fail")
	}

	return os.Remove(pivotDir)
}

func StartContainerInitProcess() error {
	setUpMount()

	startCmd := "top"
	// 根据命令查找环境变量，找到可执行文件
	path, err := exec.LookPath(startCmd)
	if err != nil {
		log.Error(err)
	}

	// 利用syscall.Exec()方法调用execve系统调用，覆盖当前进程，使容器中运行的
	// command 成为PID 1 (实际上运行的第一个command是 mydocker init ...)
	// 第一个参数为可执行二进制文件路径， 如 "/bin/ls"
	// 第二个参数为具体命令 []string, 如 ["ls", "./"]
	// 第三个参数为环境变量
	if err := syscall.Exec(path, []string{startCmd}, os.Environ()); err != nil {
		log.Errorf("RunContainerInitProcess exec: %s", err.Error())
	}

	return nil
}

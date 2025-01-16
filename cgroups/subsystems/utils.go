package subsystems

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Get the absolute path of a Cgroup
func getCgroupPath(subsystem Subsystem, cgroup string, autoCreate bool) (string, error) {
	mountPath, err := findSubsystemMountPath(subsystem.Name())
	if err != nil {
		logrus.Error(err)
	}
	cgroupPath := path.Join(mountPath, cgroup)

	if !autoCreate {
		return cgroupPath, nil
	}

	_, err = os.Stat(cgroupPath)
	if err != nil && os.IsNotExist(err) {
		err = os.Mkdir(cgroupPath, 0755)
		return cgroupPath, err
	}

	return cgroupPath, errors.Wrap(err, "Create cgroup")
}

const mountPointIndex = 4

// Find the hierarchy cgroup dir from /proc/self/mountinfo
func findSubsystemMountPath(subsystem string) (string, error) {
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	var txt string
	for scanner.Scan() {
		// It will be like this:
		// 136 133 0:55 / /sys/fs/cgroup/cpu rw,nosuid,nodev,noexec,relatime shared:41 - cgroup cgroup rw,cpu
		txt = scanner.Text()
		fileds := strings.Split(txt, " ")

		subs := strings.Split(fileds[len(fileds)-1], ",")

		for _, s := range subs {
			if s == subsystem {
				return fileds[mountPointIndex], nil
			}
		}
	}

	if err = scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("mount dir of %s not found", subsystem)
}

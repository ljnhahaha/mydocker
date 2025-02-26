package subsystemsv2

import (
	"mydocker/utils"
	"os"
	"path/filepath"
	"strconv"

	"github.com/pkg/errors"
)

const unifiedCgroupPath = "/sys/fs/cgroup"

// Get the absolute path of a Cgroup
func getCgroupPath(cgroup string, autoCreate bool) (string, error) {
	cgroupPath := filepath.Join(unifiedCgroupPath, cgroup)
	if !autoCreate {
		return cgroupPath, nil
	}

	_, err := os.Stat(cgroupPath)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(cgroupPath, 0755)
			return cgroupPath, err
		}
	}

	return cgroupPath, errors.Wrap(err, "unknown file stat")
}

func applyCgroup(pid int, cgroup string) error {
	cgroupPath := filepath.Join(unifiedCgroupPath, cgroup)
	if exist, err := utils.PathExist(cgroupPath); !exist {
		return errors.Wrapf(err, "cgroup [%s] not exists", cgroup)
	}

	if err := os.WriteFile(filepath.Join(cgroupPath, "cgroup.procs"), []byte(strconv.Itoa(pid)), 0644); err != nil {
		return errors.Wrap(err, "add process to cgroup falied")
	}
	return nil
}

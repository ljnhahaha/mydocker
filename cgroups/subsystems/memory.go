package subsystems

import (
	"os"
	"path"
	"strconv"

	"mydocker/utils"

	"github.com/pkg/errors"
)

type MemorySubsystem struct {
}

func (ms *MemorySubsystem) Name() string {
	return "memory"
}

func (ms *MemorySubsystem) Set(cgroup string, rcfg *ResourceConfig) error {
	if rcfg.MemoryLimit == "" {
		return nil
	}

	cgroupPath, err := getCgroupPath(ms, cgroup, true)
	if err != nil {
		return err
	}

	if err = os.WriteFile(path.Join(cgroupPath, "memory.limit_in_bytes"), []byte(rcfg.MemoryLimit), 0644); err != nil {
		return errors.Wrap(err, "set memory fail")
	}

	return nil
}

func (ms *MemorySubsystem) Apply(cgroup string, pid int, rcfg *ResourceConfig) error {
	if rcfg.MemoryLimit == "" {
		return nil
	}

	cgroupPath, err := getCgroupPath(ms, cgroup, false)
	if err != nil {
		return errors.Wrapf(err, "get cgroup %s", cgroupPath)
	}

	if exist, err := utils.PathExist(cgroupPath); !exist {
		return errors.Wrap(err, "memory cgroup does not exist")
	}

	if err = os.WriteFile(path.Join(cgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
		return errors.Wrap(err, "set process fail")
	}

	return nil
}

func (ms *MemorySubsystem) Remove(cgroup string) error {
	cgroupPath, err := getCgroupPath(ms, cgroup, false)
	if err != nil {
		return err
	}

	return os.RemoveAll(cgroupPath)
}

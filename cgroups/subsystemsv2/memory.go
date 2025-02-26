package subsystemsv2

import (
	"os"
	"path"

	"mydocker/cgroups/resource"

	"github.com/pkg/errors"
)

type MemorySubsystem struct {
}

func (ms *MemorySubsystem) Name() string {
	return "memory"
}

func (ms *MemorySubsystem) Set(cgroup string, rcfg *resource.ResourceConfig) error {
	if rcfg.MemoryLimit == "" {
		return nil
	}

	cgroupPath, err := getCgroupPath(cgroup, true)
	if err != nil {
		return err
	}

	if err = os.WriteFile(path.Join(cgroupPath, "memory.max"), []byte(rcfg.MemoryLimit), 0644); err != nil {
		return errors.Wrap(err, "set memory fail")
	}

	return nil
}

func (ms *MemorySubsystem) Apply(cgroup string, pid int, rcfg *resource.ResourceConfig) error {
	if rcfg.MemoryLimit == "" {
		return nil
	}
	return applyCgroup(pid, cgroup)
}

func (ms *MemorySubsystem) Remove(cgroup string) error {
	cgroupPath, err := getCgroupPath(cgroup, false)
	if err != nil {
		return err
	}

	return os.RemoveAll(cgroupPath)
}

package subsystems

import (
	"os"
	"path"
	"strconv"

	"github.com/pkg/errors"
)

type CpusetSubsystem struct {
}

func (css *CpusetSubsystem) Name() string {
	return "cpuset"
}

func (css *CpusetSubsystem) Set(cgroup string, rcfg *ResourceConfig) error {
	if rcfg.CpuSet == "" {
		return nil
	}

	cgroupPath, err := getCgroupPath(css, cgroup, true)
	if err != nil {
		return err
	}

	if err = os.WriteFile(path.Join(cgroupPath, "cpuset.cpus"), []byte(rcfg.CpuSet), 0644); err != nil {
		return errors.Wrap(err, "set cpuset fail")
	}

	return nil
}

func (css *CpusetSubsystem) Apply(cgroup string, pid int, rcfg *ResourceConfig) error {
	if rcfg.CpuSet == "" {
		return nil
	}

	cgroupPath, err := getCgroupPath(css, cgroup, false)
	if err != nil {
		return err
	}

	if err = isFileExist(cgroupPath); err != nil {
		return err
	}

	if err = os.WriteFile(path.Join(cgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
		return errors.Wrap(err, "set process fail")
	}

	return nil
}

func (css *CpusetSubsystem) Remove(cgroup string) error {
	cgroupPath, err := getCgroupPath(css, cgroup, false)
	if err != nil {
		return err
	}
	return os.RemoveAll(cgroupPath)
}

package subsystemsv2

import (
	"os"
	"path"

	"mydocker/cgroups/resource"

	"github.com/pkg/errors"
)

type CpusetSubsystem struct {
}

func (css *CpusetSubsystem) Name() string {
	return "cpuset"
}

func (css *CpusetSubsystem) Set(cgroup string, rcfg *resource.ResourceConfig) error {
	if rcfg.CpuSet == "" {
		return nil
	}

	cgroupPath, err := getCgroupPath(cgroup, true)
	if err != nil {
		return err
	}

	if err = os.WriteFile(path.Join(cgroupPath, "cpuset.cpus"), []byte(rcfg.CpuSet), 0644); err != nil {
		return errors.Wrap(err, "set cpuset fail")
	}

	return nil
}

func (css *CpusetSubsystem) Apply(cgroup string, pid int, rcfg *resource.ResourceConfig) error {
	if rcfg.CpuSet == "" {
		return nil
	}
	return applyCgroup(pid, cgroup)
}

func (css *CpusetSubsystem) Remove(cgroup string) error {
	cgroupPath, err := getCgroupPath(cgroup, false)
	if err != nil {
		return err
	}
	return os.RemoveAll(cgroupPath)
}

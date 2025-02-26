package subsystemsv2

import (
	"fmt"
	"os"
	"path"

	"github.com/pkg/errors"

	"mydocker/cgroups/resource"
)

const PeriodDefault = 100000

type CpuSubsystem struct {
}

func (cs *CpuSubsystem) Name() string {
	return "cpu"
}

func (cs *CpuSubsystem) Set(cgroup string, rcfg *resource.ResourceConfig) error {
	if rcfg.CpuCfsQuota == 0 {
		return nil
	}

	cgroupPath, err := getCgroupPath(cgroup, true)
	if err != nil {
		return err
	}

	// cgroupv2 中统一用cpu.max
	if rcfg.CpuCfsQuota != 0 {
		cpuQuota := PeriodDefault / 100 * rcfg.CpuCfsQuota
		configStr := fmt.Sprintf("%d %d", cpuQuota, PeriodDefault)
		if err = os.WriteFile(path.Join(cgroupPath, "cpu.max"), []byte(configStr), 0644); err != nil {
			return errors.Wrap(err, "set cpu quota fail")
		}
	}

	return nil
}

func (cs *CpuSubsystem) Apply(cgroup string, pid int, rcfg *resource.ResourceConfig) error {
	if rcfg.CpuCfsQuota == 0 {
		return nil
	}
	return applyCgroup(pid, cgroup)
}

func (cs *CpuSubsystem) Remove(cgroup string) error {
	cgroupPath, err := getCgroupPath(cgroup, false)
	if err != nil {
		return err
	}
	return os.RemoveAll(cgroupPath)
}

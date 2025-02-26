package subsystemsv1

import (
	"os"
	"path"
	"strconv"

	"github.com/pkg/errors"

	"mydocker/cgroups/resource"
	"mydocker/utils"
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

	cgroupPath, err := getCgroupPath(cs, cgroup, true)
	if err != nil {
		return err
	}

	if rcfg.CpuCfsQuota != 0 {
		if err = os.WriteFile(path.Join(cgroupPath, "cpu.cfs_period_us"), []byte(strconv.Itoa(PeriodDefault)), 0644); err != nil {
			return errors.Wrap(err, "set cpu period fail")
		}

		cpuQuota := PeriodDefault / 100 * rcfg.CpuCfsQuota
		if err = os.WriteFile(path.Join(cgroupPath, "cpu.cfs_quota_us"), []byte(strconv.Itoa(cpuQuota)), 0644); err != nil {
			return errors.Wrap(err, "set cpu quota fail")
		}
	}

	return nil
}

func (cs *CpuSubsystem) Apply(cgroup string, pid int, rcfg *resource.ResourceConfig) error {
	if rcfg.CpuCfsQuota == 0 {
		return nil
	}

	cgroupPath, err := getCgroupPath(cs, cgroup, false)
	if err != nil {
		return err
	}

	if exist, err := utils.PathExist(cgroupPath); !exist {
		return errors.Wrap(err, "cpu cgroup does not exist")
	}

	if err = os.WriteFile(path.Join(cgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
		return errors.Wrap(err, "set process fail")
	}

	return nil
}

func (cs *CpuSubsystem) Remove(cgroup string) error {
	cgroupPath, err := getCgroupPath(cs, cgroup, false)
	if err != nil {
		return err
	}
	return os.RemoveAll(cgroupPath)
}

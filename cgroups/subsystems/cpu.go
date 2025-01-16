package subsystems

import (
	"os"
	"path"
	"strconv"

	"github.com/pkg/errors"
)

const PeriodDefault = 100000

type CpuSubsystem struct {
}

func (cs *CpuSubsystem) Name() string {
	return "cpu"
}

func (cs *CpuSubsystem) Set(cgroup string, rcfg *ResourceConfig) error {
	if rcfg.CpuCfsQuota == 0 && rcfg.CpuShare == "" {
		return nil
	}

	cgroupPath, err := getCgroupPath(cs, cgroup, true)
	if err != nil {
		return err
	}

	if rcfg.CpuShare != "" {
		if err := os.WriteFile(path.Join(cgroupPath, "cpu.shares"), []byte(rcfg.CpuShare), 0644); err != nil {
			return errors.Wrap(err, "set cpushare fail")
		}
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

func (cs *CpuSubsystem) Apply(cgroup string, pid int, rcfg *ResourceConfig) error {
	if rcfg.CpuCfsQuota == 0 && rcfg.CpuShare == "" {
		return nil
	}

	cgroupPath, err := getCgroupPath(cs, cgroup, false)
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

func (cs *CpuSubsystem) Remove(cgroup string) error {
	cgroupPath, err := getCgroupPath(cs, cgroup, false)
	if err != nil {
		return err
	}
	return os.RemoveAll(cgroupPath)
}

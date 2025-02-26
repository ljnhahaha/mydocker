package cgroups

import (
	"mydocker/cgroups/resource"
	"mydocker/cgroups/subsystemsv2"

	"github.com/sirupsen/logrus"
)

// 对所有subsystem中的cgroups进行管理
type CgroupManagerV2 struct {
	// relative path
	Path       string
	Resource   *resource.ResourceConfig
	Subsystems []resource.Subsystem
}

func NewCgroupManagerV2(path string) *CgroupManagerV2 {
	return &CgroupManagerV2{
		Path:       path,
		Subsystems: subsystemsv2.SubsystemSet,
	}
}

func (m *CgroupManagerV2) Set(res *resource.ResourceConfig) error {
	for _, subs := range m.Subsystems {
		err := subs.Set(m.Path, res)
		if err != nil {
			logrus.Errorf("Set system: %s, err: %s", subs.Name(), err.Error())
			return nil
		}
	}

	m.Resource = res

	return nil
}

func (m *CgroupManagerV2) Apply(pid int, res *resource.ResourceConfig) error {
	for _, subs := range m.Subsystems {
		err := subs.Apply(m.Path, pid, res)
		if err != nil {
			logrus.Errorf("Add proc: %d to system: %s, err: %s", pid, subs.Name(), err.Error())
		}
	}

	return nil
}

func (m *CgroupManagerV2) Destroy() error {
	for _, subs := range m.Subsystems {
		err := subs.Remove(m.Path)
		if err != nil {
			logrus.Warn(err)
		}
	}

	return nil
}

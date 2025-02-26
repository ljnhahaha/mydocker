package cgroups

import (
	"mydocker/cgroups/resource"
	"mydocker/cgroups/subsystemsv1"

	"github.com/sirupsen/logrus"
)

// 对所有subsystem中的cgroups进行管理
type CgroupManagerV1 struct {
	// relative path
	Path       string
	Resource   *resource.ResourceConfig
	Subsystems []resource.Subsystem
}

func NewCgroupManagerV1(path string) *CgroupManagerV1 {
	return &CgroupManagerV1{
		Path:       path,
		Subsystems: subsystemsv1.SubsystemSet,
	}
}

func (m *CgroupManagerV1) Set(res *resource.ResourceConfig) error {
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

func (m *CgroupManagerV1) Apply(pid int, res *resource.ResourceConfig) error {
	for _, subs := range m.Subsystems {
		err := subs.Apply(m.Path, pid, res)
		if err != nil {
			logrus.Errorf("Add proc: %d to system: %s, err: %s", pid, subs.Name(), err.Error())
		}
	}

	return nil
}

func (m *CgroupManagerV1) Destroy() error {
	for _, subs := range m.Subsystems {
		err := subs.Remove(m.Path)
		if err != nil {
			logrus.Warn(err)
		}
	}

	return nil
}

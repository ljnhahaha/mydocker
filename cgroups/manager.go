package cgroups

import (
	"mydocker/cgroups/subsystems"

	"github.com/sirupsen/logrus"
)

// 对所有subsystem中的cgroups进行管理
type CgroupManager struct {
	// relative path
	Path     string
	Resource *subsystems.ResourceConfig
}

func NewCgroupManager(path string) *CgroupManager {
	return &CgroupManager{
		Path: path,
	}
}

func (m *CgroupManager) Set(res *subsystems.ResourceConfig) error {
	for _, subs := range subsystems.SubsystemSet {
		err := subs.Set(m.Path, res)
		if err != nil {
			logrus.Errorf("Set system: %s, err: %s", subs.Name(), err.Error())
			return nil
		}
	}

	m.Resource = res

	return nil
}

func (m *CgroupManager) Apply(pid int, res *subsystems.ResourceConfig) error {
	for _, subs := range subsystems.SubsystemSet {
		err := subs.Apply(m.Path, pid, res)
		if err != nil {
			logrus.Errorf("Add proc: %d to system: %s, err: %s", pid, subs.Name(), err.Error())
		}
	}

	return nil
}

func (m *CgroupManager) Destroy() error {
	for _, subs := range subsystems.SubsystemSet {
		err := subs.Remove(m.Path)
		if err != nil {
			logrus.Warn(err)
		}
	}

	return nil
}

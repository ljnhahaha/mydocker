package cgroups

import (
	"mydocker/cgroups/resource"
)

type CgroupManager interface {
	Set(res *resource.ResourceConfig) error
	Apply(pid int, res *resource.ResourceConfig) error
	Destroy() error
}

func NewCgroupManager(path string) CgroupManager {
	if isUnifiedCgroup() {
		return NewCgroupManagerV2(path)
	}
	return NewCgroupManagerV1(path)
}

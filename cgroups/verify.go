package cgroups

import (
	"os"
	"sync"

	"golang.org/x/sys/unix"
)

const unifiedMountPoint = "/sys/fs/cgroup"

var (
	once      sync.Once
	isUnified bool
)

func isUnifiedCgroup() bool {
	once.Do(func() {
		var st unix.Statfs_t
		err := unix.Statfs(unifiedMountPoint, &st)
		if err != nil && os.IsNotExist(err) {
			isUnified = false
			return
		}
		isUnified = (st.Type == unix.CGROUP2_SUPER_MAGIC)
	})

	return isUnified
}

package subsystems

// Resource control subsystems
type Subsystem interface {
	Name() string

	// Set resources in cgroups, `path` represents dir path of cgroup
	// in virtual file system
	Set(cgroup string, rcfg *ResourceConfig) error

	// Add process into cgroup
	Apply(cgroup string, pid int, rcfg *ResourceConfig) error

	// Remove cgruop from subsystem
	Remove(cgroup string) error
}

type ResourceConfig struct {
	MemoryLimit string
	CpuCfsQuota int
	CpuShare    string
	CpuSet      string
}

var SubsystemSet = []Subsystem{
	&CpuSubsystem{},
	&CpusetSubsystem{},
	&MemorySubsystem{},
}

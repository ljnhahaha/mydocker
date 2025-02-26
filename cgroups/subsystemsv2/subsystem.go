package subsystemsv2

import (
	"mydocker/cgroups/resource"
)

var SubsystemSet = []resource.Subsystem{
	&CpuSubsystem{},
	&CpusetSubsystem{},
	&MemorySubsystem{},
}

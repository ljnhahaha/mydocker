package network

import (
	"net"

	"github.com/vishvananda/netlink"
)

type Network struct {
	Name    string     // 网络名
	IPRange *net.IPNet // 网段
	Driver  string     // 网络驱动名
}

type Endpoint struct {
	ID          string           `json:"id"`
	Device      netlink.Veth     `json:"dev"`
	IPAddr      net.IP           `json:"ip"`
	MacAddr     net.HardwareAddr `json:"mac"`
	Network     *Network
	PortMapping []string
}

type Driver interface {
	Name() string
	Create(subnet string, name string) (*Network, error)
	Delete(*Network) error
	Connect(*Network, *Endpoint) error
	Disconnect(*Endpoint) error
}

type IPAMer interface {
	Allocate(subnet *net.IPNet) (net.IP, error)
	Release(subnet *net.IPNet, ipaddr *net.IP) error
}

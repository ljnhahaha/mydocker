package network

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

type BridgeNetworkDriver struct {
}

func (b *BridgeNetworkDriver) Name() string {
	return "bridge"
}

func (b *BridgeNetworkDriver) Create(subnet string, name string) (*Network, error) {
	ip, ipRange, _ := net.ParseCIDR(subnet)
	ipRange.IP = ip
	n := &Network{
		Name:    name,
		IPRange: ipRange,
		Driver:  b.Name(),
	}
	if err := b.initBridge(n); err != nil {
		return nil, err
	}
	return n, nil
}

// 删除 Bridge
func (b *BridgeNetworkDriver) Delete(network *Network) error {
	// 删除所有路由
	err := delIPRoute(network.Name, network.IPRange.IP.String())
	if err != nil {
		return errors.WithMessage(err, "delete routes failed")
	}

	// 删除 iptables 规则
	err = deleteIPTables(network.Name, network.IPRange)
	if err != nil {
		return errors.WithMessage(err, "delete iptables failed")
	}

	// 删除 Bridge
	err = b.delBridge(network)
	if err != nil {
		return errors.WithMessage(err, "delete bridge failed")
	}

	return nil
}

// 连接一个网络和 Endpoint
func (b *BridgeNetworkDriver) Connect(network *Network, endpoint *Endpoint) error {
	bridgeName := network.Name
	br, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return errors.Wrapf(err, "retrieve bridge %s failed", bridgeName)
	}

	la := netlink.NewLinkAttrs()
	la.Name = endpoint.ID[:5]
	// 设置Veth的master属性，veth的一端挂载到对应的Bridge
	la.MasterIndex = br.Attrs().Index
	endpoint.Device = netlink.Veth{
		LinkAttrs: la,
		PeerName:  "cif-" + endpoint.ID[:5], // veth 另一端的名字
	}

	if err = netlink.LinkAdd(&endpoint.Device); err != nil {
		return errors.Wrap(err, "add endpoint device failed")
	}

	if err = netlink.LinkSetUp(&endpoint.Device); err != nil {
		return errors.Wrap(err, "set endpoint device up failed")
	}
	return nil
}

// 关于Veth的删除：通过测试发现在容器进程结束后的一段时间内veth会自动被删除
func (b *BridgeNetworkDriver) Disconnect(endpoint *Endpoint) error {
	vethName := endpoint.ID[:5]
	veth, err := netlink.LinkByName(vethName)
	if err != nil {
		return errors.Wrapf(err, "retrieve veth [%s] failed", vethName)
	}
	// 从网桥解绑
	if err = netlink.LinkSetNoMaster(veth); err != nil {
		return err
	}

	// 删除 veth pair
	if err = netlink.LinkDel(veth); err != nil {
		return errors.Wrapf(err, "delete veth [%s] failed", vethName)
	}
	vPeerName := fmt.Sprintf("cif-%s", vethName)
	vethPeer, err := netlink.LinkByName(vPeerName)
	if err != nil {
		return errors.Wrapf(err, "retrieve veth [%s] failed", vPeerName)
	}

	if err = netlink.LinkDel(vethPeer); err != nil {
		return errors.Wrapf(err, "delete veth [%s] failed", vethName)
	}

	return nil
}

func (b *BridgeNetworkDriver) initBridge(n *Network) error {
	// create bridge
	bridgeName := n.Name
	if err := createBridgeInterface(bridgeName); err != nil {
		return err
	}

	// 设置bridge的地址和路由
	if err := setInterfaceIP(bridgeName, n.IPRange.String()); err != nil {
		return err
	}

	// 启动bridge
	if err := setInterfaceUP(bridgeName); err != nil {
		return err
	}

	// 设置 iptables SNAT 规则
	if err := setupIPTables(bridgeName, n.IPRange); err != nil {
		return errors.Wrap(err, "set iptables failed")
	}

	return nil
}

func (b *BridgeNetworkDriver) delBridge(network *Network) error {
	name := network.Name
	link, err := netlink.LinkByName(name)
	if err != nil {
		return errors.Wrapf(err, "retrieve network [%s] failed", name)
	}

	if err = netlink.LinkDel(link); err != nil {
		return errors.Wrapf(err, "delete network [%s] failed", name)
	}

	return nil
}

// 创建 bridge 设备
func createBridgeInterface(bridgeName string) error {
	// check device already existed?
	_, err := net.InterfaceByName(bridgeName)
	if err == nil || !isInterfaceNotExist(err) {
		return errors.Wrap(err, "interface existed")
	}

	// 创建设备属性
	la := netlink.NewLinkAttrs()
	la.Name = bridgeName
	// 创建网桥对象
	bridge := &netlink.Bridge{
		LinkAttrs: la,
	}
	// 添加网桥虚拟设备
	if err = netlink.LinkAdd(bridge); err != nil {
		return errors.Wrap(err, "create bridge failed")
	}
	return nil
}

func isInterfaceNotExist(err error) bool {
	return strings.Contains(err.Error(), "no such network interface")
}

// ip addr add xxx
func setInterfaceIP(name, rawIP string) error {
	retries := 2
	var iface netlink.Link
	var err error

	for i := 0; i < retries; i++ {
		iface, err = netlink.LinkByName(name)
		if err == nil {
			break
		}
		log.Debugf("error retrieving bridge netlink [%s], trying times [%d/%d]", name, i, retries)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return errors.Wrap(err, "retrieve bridge failed")
	}

	ipNet, err := netlink.ParseIPNet(rawIP)
	if err != nil {
		return errors.Wrap(err, "parse ip failed")
	}

	// 为一个interface添加ip，如果配置了网段信息，如 192.168.0.0/24
	// 则也会配置路由 192.168.0.0/24 到这个网络接口上
	addr := &netlink.Addr{IPNet: ipNet}
	return netlink.AddrAdd(iface, addr)
}

// ip link set xxx up
func setInterfaceUP(name string) error {
	iface, err := netlink.LinkByName(name)
	if err != nil {
		return errors.Wrap(err, "retrieve bridge failed")
	}

	if err := netlink.LinkSetUp(iface); err != nil {
		return errors.Wrap(err, "set bridge up failed")
	}
	return nil
}

// iptables -t nat -A POSTROUTING -s {subnet} ! -o {deviceName} -j MASQUERADE
func setupIPTables(bridgeName string, subnet *net.IPNet) error {
	return configIPTables(bridgeName, subnet, false)
}

func deleteIPTables(bridgeName string, subnet *net.IPNet) error {
	return configIPTables(bridgeName, subnet, true)
}

func configIPTables(name string, subnet *net.IPNet, isDelete bool) error {
	action := "-A"
	if isDelete {
		action = "-D"
	}

	_, subnet, _ = net.ParseCIDR(subnet.String())
	iptablesCmd := fmt.Sprintf("-t nat %s POSTROUTING -s %s ! -o %s -j MASQUERADE", action, subnet.String(), name)
	cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
	log.Infoln("SNAT cmd: ", cmd.String())
	// 执行并获得输出
	output, err := cmd.Output()
	if err != nil {
		log.Errorf("iptalbes output: %v", output)
	}
	return err
}

// ip address del xxx
func delIPRoute(name string, rawIP string) error {
	retries := 2
	var err error
	var iface netlink.Link
	for i := 0; i < retries; i++ {
		iface, err = netlink.LinkByName(name)
		if err == nil {
			break
		}
		log.Debugf("error retrieving bridge netlink [%s], trying times [%d/%d]", name, i, retries)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return errors.Wrap(err, "retrieve bridge failed")
	}

	routeList, err := netlink.RouteList(iface, netlink.FAMILY_V4)
	if err != nil {
		return errors.Wrapf(err, "list route of %s failed", name)
	}
	for _, route := range routeList {
		if route.Dst.String() == rawIP {
			err = netlink.RouteDel(&route)
			if err != nil {
				log.Errorf("delete route %v failed, %v", route, err)
			}
		}
	}
	return nil
}

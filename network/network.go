package network

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"mydocker/container"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/tabwriter"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

var (
	defaultNetworkPath = "/var/lib/mydocker/network/network/"
	drivers            = map[string]Driver{}
	ipAllocator        = IPAMer(IPAllocator)
)

func init() {
	var bridgeDriver = BridgeNetworkDriver{}
	drivers[bridgeDriver.Name()] = &bridgeDriver

	if _, err := os.Stat(defaultNetworkPath); err != nil {
		if !os.IsNotExist(err) {
			log.Error(err)
			return
		}

		if err = os.MkdirAll(defaultNetworkPath, 0644); err != nil {
			log.Error(err)
			return
		}
	}
}

func (net *Network) dump(dumpPath string) error {
	if _, err := os.Stat(dumpPath); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if err = os.MkdirAll(dumpPath, 0644); err != nil {
			return err
		}
	}
	dumpFilePath := filepath.Join(dumpPath, net.Name)
	fd, err := os.OpenFile(dumpFilePath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer fd.Close()

	jsonByte, err := json.Marshal(net)
	if err != nil {
		return err
	}
	_, err = fd.Write(jsonByte)
	if err != nil {
		return err
	}
	return nil
}

func (net *Network) load(dumpPath string) error {
	contentByte, err := os.ReadFile(dumpPath)
	if err != nil {
		return errors.Wrapf(err, "read %s failed", dumpPath)
	}
	if err = json.Unmarshal(contentByte, net); err != nil {
		return errors.Wrap(err, "unmarshal json failed")
	}
	return nil
}

func (net *Network) remove(dumpPath string) error {
	fullPath := filepath.Join(dumpPath, net.Name)
	if _, err := os.Stat(fullPath); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		return nil
	}

	return os.Remove(fullPath)
}

// 加载network文件夹下的所有网络
func loadNetworks() (map[string]*Network, error) {
	networks := map[string]*Network{}

	err := filepath.Walk(defaultNetworkPath, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		network := new(Network)
		_, networkName := filepath.Split(path)
		network.Name = networkName
		if err := network.load(path); err != nil {
			log.Error(err)
		}
		networks[networkName] = network
		return nil
	})

	return networks, err
}

func CreateNetwork(driver, subnet, name string) error {
	// 1. 解析子网
	_, cidr, _ := net.ParseCIDR(subnet)

	// 2. 分配一个IP
	ip, err := ipAllocator.Allocate(cidr)
	if err != nil {
		return err
	}
	cidr.IP = ip

	// 3. 创建网络设备
	n, err := drivers[driver].Create(cidr.String(), name)
	if err != nil {
		return err
	}

	// 4. 保存网络信息
	return n.dump(defaultNetworkPath)
}

func ListNetwork() {
	networks, err := loadNetworks()
	if err != nil {
		log.Error(err)
	}

	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprint(w, "NAME\tIP Range\tDriver\n")
	for _, n := range networks {
		fmt.Fprintf(w, "%s\t%v\t%s\n", n.Name, n.IPRange, n.Driver)
	}

	if err = w.Flush(); err != nil {
		log.Error(err)
	}
}

func DeleteNetwork(name string) error {
	networks, err := loadNetworks()
	if err != nil {
		return errors.WithMessage(err, "load networks failed")
	}

	n, ok := networks[name]
	if !ok {
		return fmt.Errorf("retrieve network %s failed", name)
	}

	err = ipAllocator.Release(n.IPRange, &n.IPRange.IP)
	if err != nil {
		return errors.Wrap(err, "release net failed")
	}

	err = drivers[n.Driver].Delete(n)
	if err != nil {
		return errors.Wrap(err, "delete net failed")
	}

	return n.remove(defaultNetworkPath)
}

// 将容器接入指定网络
func Connect(networkName string, info *container.Info) (net.IP, error) {
	networks, err := loadNetworks()
	if err != nil {
		return nil, errors.WithMessage(err, "load networks failed")
	}

	n, ok := networks[networkName]
	if !ok {
		return nil, fmt.Errorf("retrieve network %s failed", networkName)
	}

	ip, err := ipAllocator.Allocate(n.IPRange)
	if err != nil {
		return nil, errors.Wrap(err, "allocate ip failed")
	}

	ep := &Endpoint{
		ID:          fmt.Sprintf("%s-%s", info.Id, networkName),
		IPAddr:      ip,
		Network:     n,
		PortMapping: info.PortMapping,
	}

	// 将veth一端连接到的bridge
	if err = drivers[n.Driver].Connect(n, ep); err != nil {
		return ip, errors.Wrapf(err, "attach to network %s failed", n.Name)
	}

	// 将veth另一端连接到容器内部
	if err = configEndpointIpAddressAndRoute(ep, info); err != nil {
		return ip, err
	}

	return ip, addPortMapping(ep)
}

// 将容器从指定网络中移除
func Disconnect(info *container.Info) error {
	networkName := info.NetworkName
	networks, err := loadNetworks()
	if err != nil {
		return errors.WithMessage(err, "load networks failed")
	}
	n, ok := networks[networkName]
	if !ok {
		return fmt.Errorf("retrieve network %s failed", networkName)
	}

	ip := net.ParseIP(info.IP)
	ep := &Endpoint{
		ID:          fmt.Sprintf("%s-%s", info.Id, networkName),
		Network:     n,
		PortMapping: info.PortMapping,
		IPAddr:      ip,
	}

	if err = delPortMapping(ep); err != nil {
		return errors.WithMessage(err, "delete iptables failed")
	}

	drivers[n.Driver].Disconnect(ep)

	if err = ipAllocator.Release(n.IPRange, &ip); err != nil {
		return errors.Wrap(err, "release ip failed")
	}
	return nil
}

func configEndpointIpAddressAndRoute(ep *Endpoint, info *container.Info) error {
	peer, err := netlink.LinkByName(ep.Device.PeerName)
	if err != nil {
		return err
	}

	defer enterContainerNS(&peer, info)()

	interfaceIP := ep.Network.IPRange
	interfaceIP.IP = ep.IPAddr
	if err = setInterfaceIP(ep.Device.PeerName, interfaceIP.String()); err != nil {
		return errors.Wrap(err, "set ip failed")
	}
	if err = setInterfaceUP(ep.Device.PeerName); err != nil {
		return errors.Wrap(err, "set up link failed")
	}
	// 开启namespace中的本地循环Loopback, lo
	if err = setInterfaceUP("lo"); err != nil {
		return errors.Wrap(err, "set loopback up failed")
	}

	_, cidr, _ := net.ParseCIDR("0.0.0.0/0")
	// 配置默认路由，由veth发送给bridge
	defaultRoute := &netlink.Route{
		LinkIndex: peer.Attrs().Index,
		Gw:        ep.Network.IPRange.IP, // bridge ip
		Dst:       cidr,
	}
	if err = netlink.RouteAdd(defaultRoute); err != nil {
		return errors.Wrap(err, "add route failed")
	}
	return nil
}

func enterContainerNS(enLink *netlink.Link, info *container.Info) func() {
	// /proc/{pid}/ns/net 即进程pid的namespace
	f, err := os.OpenFile(fmt.Sprintf("/proc/%s/ns/net", info.Pid), os.O_RDONLY, 0)
	if err != nil {
		log.Error(err)
	}

	fdNS := f.Fd()

	// 锁定goroutine在当前线程，避免gouroutine的调度
	// 否则就不能保证一直在所需的namespace中了
	runtime.LockOSThread()

	// 将veth的另一端移动到容器的namespace中
	if err := netlink.LinkSetNsFd(*enLink, int(fdNS)); err != nil {
		log.Errorf("set link netns failed, %v", err)
	}

	// 获取当前net namespace
	oriNS, err := netns.Get()
	if err != nil {
		log.Errorf("get current netns failed, %v", err)
	}

	// 进程进入容器内的 net namespace，方便设置IP等
	if err = netns.Set(netns.NsHandle(fdNS)); err != nil {
		log.Errorf("enter netns failed, %v", err)
	}

	// 返回原来的 net namespace
	return func() {
		netns.Set(oriNS)
		oriNS.Close()
		runtime.UnlockOSThread()
		f.Close()
	}
}

func addPortMapping(ep *Endpoint) error {
	return configPortMapping(ep, false)
}

func delPortMapping(ep *Endpoint) error {
	return configPortMapping(ep, true)
}

func configPortMapping(ep *Endpoint, isDel bool) error {
	action := "-A"
	if isDel {
		action = "-D"
	}

	for _, pm := range ep.PortMapping {
		portMapping := strings.Split(pm, ":")
		if len(portMapping) < 2 {
			log.Errorf("invalid port mapping format %s", pm)
		}

		// iptables -t nat -A PREROUTING ! -i testbridge -p tcp -m tcp --dport 8080 -j DNAT --to-destination 10.0.0.4:80
		iptablesCmd := fmt.Sprintf("-t nat %s PREROUTING ! -i %s -p tcp -m tcp --dport %s -j DNAT --to-destination %s:%s",
			action, ep.Network.Name, portMapping[0], ep.IPAddr.String(), portMapping[1])
		cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
		log.Infoln("DNAT cmd: ", cmd.String())

		output, err := cmd.Output()
		if err != nil {
			log.Errorf("iptables output: %v", output)
			continue
		}
	}
	return nil
}

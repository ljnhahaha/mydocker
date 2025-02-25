// IP Address Manager
package network

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

const ipamDefaltAllocatorPath = "/var/lib/mydocker/network/ipam/subnet.json"

type IPAM struct {
	SubnetAllocatorPath string             // 分配文件的存放地址
	Subnets             *map[string]string // 网段与其对应的分配位图数组
}

var IPAllocator = &IPAM{
	SubnetAllocatorPath: ipamDefaltAllocatorPath,
}

// 加载网络分配信息
func (ipam *IPAM) load() error {
	if _, err := os.Stat(ipam.SubnetAllocatorPath); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		return nil
	}

	content, err := os.ReadFile(ipam.SubnetAllocatorPath)
	if err != nil {
		return errors.WithMessage(err, "read subnet file failed")
	}
	if err = json.Unmarshal(content, ipam.Subnets); err != nil {
		return errors.WithMessage(err, "unmarshal json failed")
	}
	return nil
}

// 存储网络分配信息
func (ipam *IPAM) dump() error {
	ipamDir, _ := filepath.Split(ipam.SubnetAllocatorPath)
	if _, err := os.Stat(ipamDir); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if err = os.MkdirAll(ipamDir, 0644); err != nil {
			return err
		}
	}

	subnetFile, err := os.OpenFile(ipam.SubnetAllocatorPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	jsonByte, err := json.Marshal(ipam.Subnets)
	if err != nil {
		return err
	}
	_, err = subnetFile.Write(jsonByte)
	if err != nil {
		return err
	}
	return nil
}

func (ipam *IPAM) Allocate(subnet *net.IPNet) (net.IP, error) {
	ipam.Subnets = &map[string]string{}
	err := ipam.load()
	if err != nil {
		return nil, errors.Wrap(err, "load subnet info failed")
	}

	// CIDR解析，IP, 子网, err
	_, subnet, _ = net.ParseCIDR(subnet.String())
	ones, size := subnet.Mask.Size()
	if _, exists := (*ipam.Subnets)[subnet.String()]; !exists {
		(*ipam.Subnets)[subnet.String()] = strings.Repeat("0", 1<<uint8(size-ones))
	}

	var ip net.IP
	for c := range (*ipam.Subnets)[subnet.String()] {
		if (*ipam.Subnets)[subnet.String()][c] == '0' {
			ipalloc := []byte((*ipam.Subnets)[subnet.String()])
			ipalloc[c] = '1'
			(*ipam.Subnets)[subnet.String()] = string(ipalloc)
			ip = subnet.IP

			for t := 4; t > 0; t-- {
				[]byte(ip)[4-t] += uint8(c >> ((t - 1) * 8))
			}
			ip[3] += 1
			break
		}
	}
	err = ipam.dump()
	if err != nil {
		return nil, errors.Wrap(err, "dump subnet info failed")
	}
	return ip, nil
}

func (ipam *IPAM) Release(subnet *net.IPNet, ipaddr *net.IP) error {
	ipam.Subnets = &map[string]string{}
	err := ipam.load()
	if err != nil {
		return errors.Wrap(err, "load subnet info failed")
	}

	_, subnet, _ = net.ParseCIDR(subnet.String())
	c := 0
	releaseIP := ipaddr.To4()
	releaseIP[3] -= 1
	for t := 4; t > 0; t-- {
		c += int((releaseIP[t-1] - subnet.IP[t-1]) << ((4 - t) * 8))
	}

	ipalloc := []byte((*ipam.Subnets)[subnet.String()])
	ipalloc[c] = '0'
	(*ipam.Subnets)[subnet.String()] = string(ipalloc)

	err = ipam.dump()
	if err != nil {
		return errors.Wrap(err, "dump subnet info failed")
	}

	return nil
}

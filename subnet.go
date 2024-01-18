package main

import (
	"log"
	"net"
)

type subNet struct {
	ip   net.IP
	mask net.IPMask
}

func initSubnet(ipWithMask string) (*subNet, error) {
	ip, ipNet, err := net.ParseCIDR(ipWithMask)
	if err != nil {
		return nil, err
	}
	return &subNet{ip, ipNet.Mask}, nil
}

func (sunnet *subNet) isIPInSubnet(ip net.IP) bool {
	return sunnet.ip.Equal(ip.Mask(sunnet.mask))
}

func (sunnet *subNet) dumpSubnet() {
	log.Println("  --subnet-- ")
	log.Println("  ip:   ", sunnet.ip.String())
	log.Println("  mask: ", sunnet.mask.String())
}

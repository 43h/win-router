package main

import (
	"net"
	"strings"
)

var ARPTable map[string](net.HardwareAddr) = map[string](net.HardwareAddr){}

func parseARP(line string) {
	lines := strings.Split(line, "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) == 3 {
			ip := fields[0]
			mac := fields[1]
			_, ok := ARPTable[ip]
			if !ok {
				ARPTable[net.IP(ip).To4().String()] = net.HardwareAddr(mac)
			}
		}
	}
}

func addArp(ip string, mac string) {
	_, ok := ARPTable[ip]
	if !ok {
		ARPTable[ip] = net.HardwareAddr(mac)
	}
}

func getArp(ip string) (net.HardwareAddr, bool) {
	v, ok := ARPTable[ip]
	if ok {
		return v, true
	}
	return nil, false
}

func getArpWithSearch(ip string, name string) (net.HardwareAddr, bool) {
	if v, ok := ARPTable[ip]; ok {
		return v, true
	} else {
		go arpRequest(ip)
	}
	return nil, false
}

func arpRequest(ip string) {
	mac, err := net.ParseMAC("f8:56:c3:12:1a:be")
	if err == nil {
		ARPTable[ip] = mac
	}
}

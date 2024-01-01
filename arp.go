package main

import (
	"bytes"
	"fmt"
	"net"
	"os/exec"
	"strings"
)

type ARPItem struct {
	ip  net.IP
	mac net.HardwareAddr
}

var ARPTable map[string](ARPItem) = map[string](ARPItem){}

func readARPTable() bool {
	cmd := exec.Command("arp", "-a")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Println("fail to read arp table, ", err)
		return false
	} else {
		parseARPTable(out.String())
	}
	return true
}

func parseARPTable(output string) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) == 3 {
			ip := fields[0]
			mac := fields[1]
			_, ok := ARPTable[ip]
			if !ok {
				item := ARPItem{net.IP(ip), net.HardwareAddr(mac)}
				ARPTable[ip] = item
			}
		}
	}
}

func getArpByIp(ip net.IP) net.HardwareAddr {
	v, ok := ARPTable[string(ip)]
	if ok {
		if string(ip) == string(v.ip) {
			return v.mac
		}
	}
	return nil
}

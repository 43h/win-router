package main

import (
	"bytes"
	"fmt"
	"net"
	"os/exec"
	"strings"
)

type Router struct {
	subnet net.IP
	mask   net.IP
	gw     net.IP
	ifIp   net.IP //接口
}

var routerTable map[string](Router) = map[string](Router){}

func readRouterTable() bool {
	cmd := exec.Command("route", "print", "-4")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Println("fail to read route, ", err)
		return false
	} else {
		parseRouteTable(out.String())
	}
	return true
}

func parseRouteTable(output string) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 5 && fields[0] != "Network" {
			destination := fields[0]
			mask := fields[1]
			gateway := fields[2]
			interfaceName := fields[len(fields)-1]

			fmt.Printf("Destination: %s, Mask: %s, Gateway: %s, Interface: %s\n", destination, mask, gateway, interfaceName)
		}
	}
}

func getRouter(dip net.IP) net.IP { //根据目的ip查询转发出口接口

	return nil
}

package main

import (
	"fmt"
	"github.com/google/gopacket/pcap"
	"net"
)

func parseArg(arg []string) bool {
	var argNum = len(arg)
	if argNum == 1 {
		return true
	} else if argNum == 2 {
		if arg[1] == "-h" || arg[1] == "--help" || arg[1] == "help" {
			showHelp()
		} else if arg[1] == "-n" {
			showNicByNet()
		} else if arg[1] == "-p" {
			showNicByPcap()
		} else if arg[1] == "-d" {
			dumpConf()
		} else if arg[1] == "-l" {
			logFileFlag = false
		} else if arg[1] == "-v" {
			showVersion()
			return false
		} else {
			fmt.Println("  unknown param\n exit")
			showHelp()
		}
		return false
	} else {
		fmt.Println("  unknown param")
		showHelp()
		return false
	}
}

func showHelp() {
	fmt.Println(" -v: show version")
	fmt.Println(" -h: help")
	fmt.Println(" -n: show NIC by net")
	fmt.Println(" -p: show NIC by pcap")
	fmt.Println(" -d: dump conf")
}

func showNicByPcap() {
	devices, err := pcap.FindAllDevs()
	if err != nil {
		panic(err)
	}

	// Print device information
	fmt.Println("Pcap Devices:  ")
	for _, device := range devices {
		fmt.Println("-----------------------")
		fmt.Println("  Name:        ", device.Name)
		fmt.Println("  Description: ", device.Description)
		fmt.Println("  Addresses:   ", device.Description)

		for _, address := range device.Addresses {
			fmt.Println("  IP address: ", address.IP)
			fmt.Println("  Netmask:    ", address.Netmask)
		}
	}
}

func showNicByNet() {
	ifs, err := net.Interfaces()
	if err != nil {
		panic(err)
	}

	fmt.Println("Net Devices")
	for _, f := range ifs {
		fmt.Println("-----------------------")
		fmt.Println("  Name:        ", f.Name)
		fmt.Println("  MAC:         ", f.HardwareAddr)
		fmt.Println("  Index:       ", f.Index)
		address, err := f.Addrs()
		if err == nil {
			for _, value := range address {
				fmt.Println("  IP address: ", value.String())
				fmt.Println("  Network:    ", value.Network())
			}
		}
	}
}

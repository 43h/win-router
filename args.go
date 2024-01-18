package main

import (
	"fmt"
	"github.com/google/gopacket/pcap"
	"log"
	"net"
)

func parseArg(arg []string) bool {
	var argNum = len(arg)
	if argNum == 1 {
		return true
	} else if argNum == 2 {
		if arg[1] == "-h" || arg[1] == "--help" {
			showHelpInfo()
		} else if arg[1] == "-n" {
			showNicInfo2()
		} else if arg[1] == "-nn" {
			showNicInfo()
		} else if arg[1] == "-d" {
			dumpConf()
		} else {
			fmt.Println("  unknown param\nexit")
			showHelpInfo()
		}
		return false
	} else {
		fmt.Println("  unknown param")
		showHelpInfo()
		return false
	}

	return true
}

func showHelpInfo() {
	fmt.Println("-h: help")
	fmt.Println("-n: show NIC")
	fmt.Println("-nn: show NIC")
	fmt.Println("-d: dump conf")
}

func showNicInfo() {
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

func showNicInfo1() {
	// Find all devices
	// 获取所有网卡
	devices, err := pcap.FindAllDevs()
	if err != nil {
		log.Fatal(err)
	}

	// Find exact device
	// 根据网卡名称从所有网卡中取到精确的网卡
	//var device pcap.Interface
	for _, d := range devices {
		fmt.Println(d.Name)
		////if d.Name == *deviceName {
		//	device = d
		//}
	}

	// 根据网卡的ipv4地址获取网卡的mac地址，用于后面判断数据包的方向
	//macAddr, err := findMacAddrByIp(findDeviceIpv4(device))
	//if err != nil {
	//	panic(err)
	//}

	//fmt.Printf("Chosen device's IPv4: %s\n", findDeviceIpv4(device))
	//fmt.Printf("Chosen device's MAC: %s\n", macAddr)
}

func showNicInfo2() {
	infs, err := net.Interfaces()
	if err != nil {
		panic(err)
	}

	fmt.Println("Net Devices")
	for _, f := range infs {
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

package main

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

/**
 * local pc<--->LAN|WAN<--->Intelnet
 */

type Lan struct {
	valid   bool
	ip      net.IP
	mac     net.HardwareAddr
	rip     net.IP
	rmac    net.HardwareAddr
	rflag   bool
	maskLen uint32
	name    string
	cfgName string
	que     chan gopacket.Packet
	handle  *pcap.Handle
}

type Wan struct {
	valid   bool
	ip      net.IP
	mac     net.HardwareAddr
	gwMac   net.HardwareAddr
	name    string
	cfgName string
	que     chan gopacket.Packet
	handle  *pcap.Handle
}

var lan Lan = Lan{rip: make(net.IP, 4), rmac: make(net.HardwareAddr, 6)}
var wan Wan = Wan{}

func parseNic(line string) {
	nic := line[0:3]
	name := line[5 : len(line)-1]
	if nic == "lan" {
		parseLanNic(name)
	} else if nic == "wan" {
		parseWanNic(name)
	} else if nic == "gw:" { //Hack: add wan mac manually
		wan.gwMac, _ = net.ParseMAC(line[3:])
	}
}

func parseLanNic(name string) {
	lan.cfgName = name
	infs, _ := net.Interfaces()
	for _, f := range infs {
		if name == f.Name {
			lan.mac = f.HardwareAddr

			address, err := f.Addrs()
			if err == nil {
				for _, value := range address {
					index := strings.Index(value.String(), ":") //check ipv6
					if index != -1 {                            //skip ipv6
						continue
					}

					index = strings.Index(value.String(), "/")
					if index != -1 {
						lan.ip = net.ParseIP(value.String()[0:index]).To4()
						len, _ := strconv.Atoi(value.String()[index+1:])
						lan.maskLen = uint32(len)
						lan.valid = true
					}
					break
				}
			}
		}
	}
}

func parseWanNic(name string) {
	wan.cfgName = name
	infs, _ := net.Interfaces()
	for _, f := range infs {
		if name == f.Name {
			wan.mac = f.HardwareAddr

			address, err := f.Addrs()
			if err == nil {
				for _, value := range address {
					index := strings.Index(value.String(), ":") //check ipv6
					if index != -1 {                            //skip ipv6
						continue
					}

					index = strings.Index(value.String(), "/") //127.0.0.1/28
					if index != -1 {
						wan.ip = net.ParseIP(value.String()[0:index]).To4()
						wan.valid = true
					}
					break
				}
			}
		}
	}
}

func loopNicName() bool {
	devices, err := pcap.FindAllDevs()
	if err != nil {
		fmt.Println(err)
		return false
	}

	// Find exact device
	// 根据网卡名称从所有网卡中取到精确的网卡
	//var device pcap.Interface
	for _, d := range devices {
		for _, value := range d.Addresses {
			if value.IP.String() == lan.ip.String() {
				lan.name = d.Name
			} else if value.IP.String() == wan.ip.String() {
				wan.name = d.Name
			}
		}
	}
	return true
}

func checkNic() bool {

	if loopNicName() == false {
		fmt.Println("fail to find device name")
		dumpNic()
		return false
	}
	dumpNic()
	if lan.valid == true {
		if wan.valid == true {

			return true
		}
	}

	return false
}

func dumpNic() {
	fmt.Println("--------------")
	fmt.Println("   lan:")
	fmt.Println("cfg-name:", lan.cfgName)
	fmt.Println(" name:   ", lan.name)
	fmt.Println(" valid:  ", lan.valid)
	fmt.Println(" ip:     ", lan.ip.String())
	fmt.Println(" mac:    ", lan.mac.String())
	fmt.Println(" maskLen:", lan.maskLen)

	fmt.Println("--------------")
	fmt.Println("   wan:")
	fmt.Println("cfg-name:", wan.cfgName)
	fmt.Println(" name:   ", wan.name)
	fmt.Println(" valid:  ", wan.valid)
	fmt.Println(" ip:     ", wan.ip.String())
	fmt.Println(" mac:    ", wan.mac.String())
	fmt.Println("gwmac:   ", wan.gwMac.String())
}

func getNicHandle(name string) (*pcap.Handle, bool) {
	inactive, err := pcap.NewInactiveHandle(name)
	if err != nil {
		log.Println("lan-recv: fail to open nic, ", err)
		return nil, false
	}
	defer inactive.CleanUp()

	err = inactive.SetImmediateMode(true)
	if err != nil {
		log.Println("lan-recv: fail to set mode, ", err)
		return nil, false
	}

	if err = inactive.SetTimeout(time.Second); err != nil {
		if err != nil {
			log.Println("lan-recv: fail to set timeout, ", err)
			return nil, false
		}
	}
	//if err = inactive.SetTimestampSource("foo"); err != nil {
	//	log.Fatal(err)
	//}

	// Finally, create the actual handle by calling Activate:
	handle, err := inactive.Activate() // after this, inactive is no longer valid
	if err != nil {
		log.Println("lan-recv: fail to active nic, ", err)
		return nil, false
	}

	handle.SetDirection(pcap.DirectionIn)
	return handle, true
}

func initNic() bool {
	var rst bool
	lan.handle, rst = getNicHandle(lan.name)
	if rst == false {
		return false
	}

	wan.handle, rst = getNicHandle(wan.name)
	if rst == false {
		return false
	}
	return true
}

func closeNic() {
	if lan.handle != nil {
		lan.handle.Close()
	}
	if wan.handle != nil {
		wan.handle.Close()
	}
}

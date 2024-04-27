package main

import (
	"log"
	"net"
	"strings"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

const (
	NICLAN = iota
	NICWAN
)

type NIC struct {
	valid    bool
	netName  string
	pcapName string
	nicType  int              //wan or lan
	ip       net.IP           //每个网卡暂时只支持一个IPv4，后续考虑支持多个IPv4
	mac      net.HardwareAddr //网口mac
	gwip     net.IP           //默认网关ip
	gwmac    net.HardwareAddr //默认网关mac
	handle   *pcap.Handle     //pcap handle
	que      chan gopacket.Packet
}

var lanNic NIC = NIC{valid: true, nicType: NICLAN}
var wanNic NIC = NIC{valid: true, nicType: NICWAN}

func addNic(nicType int, name string) bool {
	var nic *NIC = nil
	switch nicType {
	case CONFWANNAME:
		nic = &wanNic
	case CONFLANNAME:
		nic = &lanNic
	default:
		log.Println("  unknown param")
		return false
	}

	if !nic.initNic(name) {
		return false
	}
	if !nic.getNicPcapName() {
		return false
	}

	if !nic.openHandle() {
		return false
	}

	nic.que = make(chan gopacket.Packet, 10000)
	return true
}

func setNicGw(nicType int, value string) bool {
	if value[0] == '"' {
		value = value[1 : len(value)-1] //去掉前后引号
	}
	wanNic.gwip = net.ParseIP(value).To4()
	mac, rst := getMacByIp(value)
	if rst == true {
		wanNic.gwmac, _ = net.ParseMAC(mac)
	} else {
		log.Println("  fail to get mac of gateway, please 'ping' " + value + " first")
		return false
	}
	return true
}

func (nic *NIC) initNic(name string) bool {
	nname := name[1 : len(name)-1] //去掉前后引号
	infs, _ := net.Interfaces()
	for _, f := range infs {
		if nname == f.Name {
			nic.netName = f.Name
			nic.mac = f.HardwareAddr

			address, err := f.Addrs()
			if err == nil {
				for _, value := range address {
					index := strings.Index(value.String(), ":") //skip ipv6
					if index != -1 {
						continue
					}

					index = strings.Index(value.String(), "/")
					if index != -1 {
						nic.ip = net.ParseIP(value.String()[0:index]).To4()
						nic.mac = f.HardwareAddr
						nic.valid = true
					}
					return true
				}
			} else {
				log.Println("  fail to get ip address")
				return false
			}
		}
	}
	log.Println("  fail to get nic, ", nname)
	return false
}

func (nic *NIC) getNicPcapName() bool {
	devices, err := pcap.FindAllDevs()
	if err != nil {
		log.Println(err)
		return false
	}

	//根据IP匹配获取网卡对应的pcap名
	for _, d := range devices {
		for _, value := range d.Addresses {
			if value.IP.String() == nic.ip.String() {
				nic.pcapName = d.Name
				return true
			}
		}
	}
	return false
}

func (nic *NIC) openHandle() bool {
	inactive, err := pcap.NewInactiveHandle(nic.pcapName)
	if err != nil {
		log.Println("openHandle: fail to open nic, ", err)
		return false
	}
	defer inactive.CleanUp()

	err = inactive.SetImmediateMode(true)
	if err != nil {
		log.Println("openHandle: fail to set mode, ", err)
		return false
	}

	if err = inactive.SetTimeout(time.Second); err != nil {
		if err != nil {
			log.Println("openHandle: fail to set timeout, ", err)
			return false
		}
	}
	//if err = inactive.SetTimestampSource("foo"); err != nil {
	//	log.Fatal(err)
	//}

	// Finally, create the actual handle by calling Activate:
	handle, err := inactive.Activate() // after this, inactive is no longer valid
	if err != nil {
		log.Println("openHandle: fail to active nic, ", err)
		return false
	}

	//FIXME: 设置方向不生效
	// if err = handle.SetDirection(pcap.DirectionIn); err != nil {
	// 	log.Println("openHandle: fail to set direction, ", err)
	// }
	nic.handle = handle
	return true
}

func dumpNics() {
	lanNic.dumpNic()
	wanNic.dumpNic()
}

func (nic *NIC) dumpNic() {
	log.Println("-------------------")
	log.Println("netName: ", nic.netName)
	log.Println("pcapName: ", nic.pcapName)
	log.Println("valid: ", nic.valid)
	if nic.nicType == NICWAN {
		log.Println("type: wan")
	} else {
		log.Println("type: lan")
	}
	log.Println("ip: ", nic.ip.String())
	log.Println("mac: ", nic.mac.String())
	log.Println("gwip: ", nic.gwip.String())
	log.Println("gwmac: ", nic.gwmac.String())
}

package main

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"log"
	"net"
	"strings"
	"time"
)

/**
 * local pc<--->LAN|WAN<--->Intelnet
 */

type NIC struct {
	nicname  string
	pcapname string
	ip       net.IP
	mac      net.HardwareAddr
	rip      net.IP
	rmac     net.HardwareAddr
	que      chan gopacket.Packet
	handle   *pcap.Handle
	stat     statNic
}

var lan NIC = NIC{}
var wan NIC = NIC{}

func parseLine(line string) {
	splitLine := strings.Split(line, ":")
	if len(splitLine) < 2 {
		return
	}

	key := splitLine[0]
	value := splitLine[1]

	// 根据 key 和 value 进行相应的处理
	if key == "wan" {
		wan.initNic(value)
	} else if key == "wanrmac" {
		wan.rmac, _ = net.ParseMAC(value)
	} else if key == "lan" {
		lan.initNic(value)
	} else if key == "lanrip" {
		lan.rip = net.ParseIP(value).To4()
	} else if key == "lanrmac" {
		lan.rmac, _ = net.ParseMAC(value)
	}
}

func (nic *NIC) initNic(name string) {
	nic.nicname = name[1 : len(name)-1]
	infs, _ := net.Interfaces()
	for _, f := range infs {
		if nic.nicname == f.Name {
			nic.mac = f.HardwareAddr
			address, err := f.Addrs()
			if err == nil {
				for _, value := range address {
					index := strings.Index(value.String(), ":") //check ipv6
					if index != -1 {                            //skip ipv6
						continue
					}
					index = strings.Index(value.String(), "/")
					if index != -1 {
						nic.ip = net.ParseIP(value.String()[0:index]).To4()
						//len, _ := strconv.Atoi(value.String()[index+1:])
					}
					break
				}
			}
		}
	}
}

func initPcapName() bool {
	devices, err := pcap.FindAllDevs()
	if err != nil {
		log.Println(err)
		return false
	}

	for _, d := range devices {
		for _, value := range d.Addresses {
			if value.IP.String() == lan.ip.String() {
				lan.pcapname = d.Name
			} else if value.IP.String() == wan.ip.String() {
				wan.pcapname = d.Name
			}
		}
	}
	return true
}

func dumpNics() {
	log.Println("------lan-------")
	log.Println(" NicName: ", lan.nicname)
	log.Println("PcapName: ", lan.pcapname)
	log.Println("      ip: ", lan.ip.String())
	log.Println("     mac: ", lan.mac.String())
	log.Println("     rip: ", lan.rip.String())
	log.Println("    rmac: ", lan.rmac.String())

	log.Println("------wan-------")
	log.Println(" NicName: ", wan.nicname)
	log.Println("PcapName: ", wan.pcapname)
	log.Println("      ip: ", wan.ip.String())
	log.Println("     mac: ", wan.mac.String())
	log.Println("     rip: ", wan.rip.String())
	log.Println("    rmac: ", wan.rmac.String())
}

func (nic *NIC) initNicHandle() bool {
	inactive, err := pcap.NewInactiveHandle(nic.pcapname)
	if err != nil {
		log.Println("lan-recv: fail to open nic, ", err)
		return false
	}
	defer inactive.CleanUp()

	err = inactive.SetImmediateMode(true)
	if err != nil {
		log.Println("fail to set mode, ", err)
		return false
	}

	if err = inactive.SetTimeout(time.Second); err != nil {
		if err != nil {
			log.Println("fail to set timeout, ", err)
			return false
		}
	}
	//if err = inactive.SetTimestampSource("foo"); err != nil {
	//	log.Fatal(err)
	//}

	// Finally, create the actual handle by calling Activate:
	handle, err := inactive.Activate() // after this, inactive is no longer valid
	if err != nil {
		log.Println("lan-recv: fail to active nic, ", err)
		return false
	}

	handle.SetDirection(pcap.DirectionIn)
	nic.handle = handle
	return true
}

func initNics() bool {
	initPcapName()
	wan.initNicHandle()
	lan.initNicHandle()

	wan.que = make(chan gopacket.Packet, 10000)
	lan.que = make(chan gopacket.Packet, 10000)
	dumpNics()
	return true
}

func closeNics() {
	if lan.handle != nil {
		lan.handle.Close()
	}
	if wan.handle != nil {
		wan.handle.Close()
	}
}

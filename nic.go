package main

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	NICUNKNOWN = iota
	NICLAN
	NICWAN
)

type NIC struct {
	valid    bool
	ip       net.IP //每个网卡暂时只支持一个IPv4，后续考虑支持多个IPv4
	mac      net.HardwareAddr
	netName  string
	pcapName string
	que      chan gopacket.Packet
	handle   *pcap.Handle
	//stat     statNic
}

type LAN struct {
	nic    NIC
	subnet *subNet //检测子网
}

type WAN struct {
	nic      NIC
	portpool *PortPool //维护SNAT端口池
}

type NICPOOL struct { //暂时仅支持一个wan口，一个lan口
	lan LAN
	wan WAN
}

var nicpool NICPOOL = NICPOOL{}

func parseNicConf(conftype int32, value string) bool {
	switch conftype {
	case CONFWANNAME:
		nicpool.wan.nic.initNic(value)
	case CONFWANMAC:
		nicpool.wan.nic.mac, _ = net.ParseMAC(value)
	case CONFWANPORT:
		parts := strings.Split(value, "-")
		if len(parts) != 2 {
			return false
		}
		start, _ := strconv.Atoi(parts[0])
		end, _ := strconv.Atoi(parts[1])
		nicpool.wan.portpool = initPortPool(uint16(start), uint16(end))
	case CONFLANNAME:
		nicpool.lan.nic.initNic(value)
	case CONFLANIP:
		nicpool.lan.subnet, _ = initSubnet(value)
	default:
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
					index := strings.Index(value.String(), ":") //check ipv6
					if index != -1 {                            //skip ipv6
						continue
					}

					index = strings.Index(value.String(), "/")
					if index != -1 {
						nic.ip = net.ParseIP(value.String()[0:index]).To4()
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
	return false
}

func getNicPcapName() bool {
	devices, err := pcap.FindAllDevs()
	if err != nil {
		log.Println(err)
		return false
	}

	// Find exact device
	// 根据网卡名称从所有网卡中取到精确的网卡
	//var device pcap.Interface
	for _, d := range devices {
		for _, value := range d.Addresses {
			if value.IP.String() == nicpool.lan.nic.ip.String() {
				nicpool.lan.nic.pcapName = d.Name
			} else if value.IP.String() == nicpool.wan.nic.ip.String() {
				nicpool.wan.nic.pcapName = d.Name
			}
		}
	}
	return true
}

func (nic *NIC) initNicHandle() (*pcap.Handle, bool) {
	inactive, err := pcap.NewInactiveHandle(nic.pcapName)
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

func (nic *NIC) dumpNic() {
	log.Println("-------------------")
	log.Println("valid: ", nic, nic.valid)
	log.Println("ip: ", nic.ip.String())
	log.Println("mac: ", nic.mac.String())
	log.Println("netName: ", nic.netName)
	log.Println("pcapName: ", nic.pcapName)
}

func initNicPool() bool {
	getNicPcapName()
	nicpool.lan.nic.initNicHandle()
	nicpool.wan.nic.initNicHandle()

	nicpool.wan.nic.que = make(chan gopacket.Packet, 10000)
	nicpool.lan.nic.que = make(chan gopacket.Packet, 10000)

	return true
}

func destoryNicPool() {
	if nicpool.lan.nic.handle != nil {
		nicpool.lan.nic.handle.Close()
	}
	if nicpool.wan.nic.handle != nil {
		nicpool.wan.nic.handle.Close()
	}
}

func dumpNicPool() {
	log.Println("NIC POOL:")
	log.Println("  LAN:")
	nicpool.lan.nic.dumpNic()
	nicpool.lan.subnet.dumpSubnet()
	log.Println("  WAN:")
	nicpool.wan.nic.dumpNic()
	nicpool.wan.portpool.dumpPortPool()
}

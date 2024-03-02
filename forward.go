package main

import (
	"bytes"
	"encoding/binary"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"log"
	"net"
)

/**
 * local pc<--->LAN|WAN<--->Intelnet
 */

const (
	MACDSTSTART  = 0
	MACDSTEND    = 6
	MACSRCSTART  = 6
	MACSRCEND    = 12
	IPSRCSTART   = 26
	IPSRCEND     = 30
	IPDSTSTART   = 30
	IPDSTEND     = 34
	PORTSRCSTART = 34
	PORTSRCEND   = 36
	PORTDSTSTART = 36
	PORTDSTEND   = 38
)

var forwordtable map[uint16]string = map[uint16]string{}
var forwardiplist map[string]bool = map[string]bool{} //转发ip列表

func forward() {
	var lan, wan *NIC
	for _, nic := range nics {
		if nic.nicType == NICLAN {
			lan = &nic
		} else if nic.nicType == NICWAN {
			wan = &nic
		}
	}

	go rcvPkt(lan)
	go rcvPkt(wan)

	var pkt gopacket.Packet
	var fromlan bool
	for {
		select {
		case pkt = <-lan.que:
			fromlan = true
		case pkt = <-wan.que:
			fromlan = false
		}
		data := pkt.Data()
		//start to handle pkt
		if fromlan == true { //from lan, to wan
			_, ok := forwardiplist[string(data[IPDSTSTART:IPDSTEND])] //记录上行ip
			if !ok {
				forwardiplist[string(data[IPDSTSTART:IPDSTEND])] = true
			}
			//使用源端口标识流
			srcPort := binary.BigEndian.Uint16(data[PORTSRCSTART:PORTSRCEND])
			_, ok = forwordtable[srcPort]
			if !ok {
				forwordtable[srcPort] = string(data[IPSRCSTART:IPSRCEND])
			} else {
				forwordtable[srcPort] = string(data[IPSRCSTART:IPSRCEND])
			}
			//记录ip与mac
			addArp(string(pkt.Data()[IPSRCSTART:IPSRCEND]), string(data[MACSRCSTART:MACSRCEND]))
			handleLanPkt(pkt, wan)
		} else { //from wan, to lan
			log.Println("from wan")
			_, ok := forwardiplist[string(data[IPSRCSTART:IPSRCEND])] //检测下行源ip
			if !ok {
				log.Println("drop", data[26:30])
				continue
			}
			handleWanData(pkt, lan)
		}
	}
}

func rcvPkt(nic *NIC) {
	handle := nic.handle
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		if len(packet.Data()) < 34 { //mac header + ip header
			continue
		}

		if bytes.Equal(packet.Data()[MACDSTSTART:MACDSTEND], nic.mac) == false { //dst mac != self mac
			continue
		}

		if bytes.Equal(packet.Data()[12:14], []byte{0x08, 0x00}) == false { //skip !ipv4
			continue
		}

		if nic.nicType == NICLAN {
			if bytes.Equal(packet.Data()[IPDSTSTART:IPDSTEND], nic.ip) == true { // skip dst ip == self ip
				continue
			}
		}

		if bytes.Equal(packet.Data()[23:24], []byte{0x11}) == true { // skip udp do it later
			continue
		}

		nic.que <- packet
	}
}

func handleWanData(pkt gopacket.Packet, nic *NIC) bool {
	data := pkt.Data()
	data = data[:getdatalen(pkt)]
	log.Println("handleWanData", data)
	dstPort := binary.BigEndian.Uint16(pkt.Data()[36:38])
	v, ok := forwordtable[dstPort]
	if ok {
		//replace dst ip
		log.Println("get dst ip", v)
		copy(data[IPDSTSTART:IPDSTEND], v)
	}

	mac, ok := getArp(v)
	if ok {
		//replace dst mac
		copy(data[MACDSTSTART:MACDSTEND], mac)
		log.Println("get dst mac", mac)
	} else {
		log.Println("no dst mac")
	}
	//replace src mac
	copy(data[MACSRCSTART:MACSRCEND], nic.mac)

	//calculate ip checksum
	copy(data[24:26], []byte{0, 0})
	checksum := ipchecksum(pkt.Data()[14:34])
	copy(data[24:26], []byte{byte(checksum >> 8), byte(checksum & 0xff)}) //set ip checksum

	//update tcp checksum
	if bytes.Equal(data[23:24], []byte{0x06}) == true { //tcp
		checksum = tcpchecksum(data[34:], data[26:30], data[30:34])
		copy(data[50:52], []byte{byte(checksum >> 8), byte(checksum & 0xff)}) //set tcp checksum
	} else if bytes.Equal(data[23:24], []byte{0x11}) == true { //udp

	}
	log.Println("forward to lan", data)
	err := nic.handle.WritePacketData(data)
	if err == nil {
		log.Println("forward to lan success")
		return true
	} else {
		log.Println("forward to lan fail")
	}
	return false
}

func handleLanPkt(pkt gopacket.Packet, nic *NIC) bool {
	data := pkt.Data()
	data = data[:getdatalen(pkt)]

	mac, rst := getArpWithSearch(string(data[30:34]), nic.netName)
	if rst {
		copy(data[MACDSTSTART:MACDSTEND], mac)
	}
	//replace src mac
	copy(data[MACSRCSTART:MACSRCEND], nic.mac)
	//replace source ip
	copy(data[IPSRCSTART:IPSRCEND], nic.ip)

	//calculate ip checksum
	copy(data[24:26], []byte{0, 0}) //set ip checksum
	checksum := ipchecksum(data[14:34])
	copy(data[24:26], []byte{byte(checksum >> 8), byte(checksum & 0xff)}) //set ip checksum

	//update tcp checksum
	if bytes.Equal(data[23:24], []byte{0x06}) == true { //tcp

		checksum = tcpchecksum(data[34:], data[26:30], data[30:34])
		copy(data[50:52], []byte{byte(checksum >> 8), byte(checksum & 0xff)}) //set tcp checksum
	} else if bytes.Equal(data[23:24], []byte{0x11}) == true { //udp

	}
	err := nic.handle.WritePacketData(data)
	if err == nil {
		log.Println("forward to wan success")
		return true
	} else {
		log.Println("forward to wan fail")
	}
	return false
}

func ipchecksum(data []byte) uint16 {
	var (
		sum    uint32
		length int = len(data)
		index  int
	)

	// 以每两个字节（16位）为一组进行求和
	for length > 1 {
		sum += uint32(binary.BigEndian.Uint16(data[index : index+2]))
		index += 2
		length -= 2
	}

	// 如果字节数为奇数，将最后一个字节单独相加
	if length > 0 {
		sum += uint32(data[index])
	}

	sum += (sum >> 16)

	// 取反得到校验和
	return uint16(^sum)
}
func caltcpchecksum(data []byte) uint16 {
	var sum uint32
	length := len(data)

	for i := 0; i < length-1; i += 2 {
		sum += uint32(data[i])<<8 | uint32(data[i+1])
	}

	if length%2 == 1 {
		sum += uint32(data[length-1]) << 8
	}

	for sum > 0xffff {
		sum = (sum >> 16) + (sum & 0xffff)
	}

	return uint16(^sum)
}

type pseudoHeader struct {
	SourceAddress      [4]byte
	DestinationAddress [4]byte
	Zero               uint8
	Protocol           uint8
	TCPLength          uint16
}

func tcpchecksum(data []byte, srcIP, dstIP net.IP) uint16 {

	pHeader := pseudoHeader{}
	copy(pHeader.SourceAddress[:], srcIP.To4())
	copy(pHeader.DestinationAddress[:], dstIP.To4())
	pHeader.Protocol = 6                  // TCP protocol number
	pHeader.TCPLength = uint16(len(data)) // TCP header length
	pHeaderBytes := make([]byte, 12)
	binary.BigEndian.PutUint32(pHeaderBytes[0:], binary.BigEndian.Uint32(pHeader.SourceAddress[:]))
	binary.BigEndian.PutUint32(pHeaderBytes[4:], binary.BigEndian.Uint32(pHeader.DestinationAddress[:]))
	pHeaderBytes[8] = pHeader.Zero
	pHeaderBytes[9] = pHeader.Protocol
	binary.BigEndian.PutUint16(pHeaderBytes[10:], pHeader.TCPLength)
	data[16] = 0
	data[17] = 0
	check := append(pHeaderBytes, data...)
	return caltcpchecksum(check)
}

func getdatalen(packet gopacket.Packet) uint16 {
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	if ipLayer != nil {
		ip, _ := ipLayer.(*layers.IPv4)
		return ip.Length + 14
	}
	return 0
}

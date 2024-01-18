package main

import (
	"bytes"
	"fmt"
	"github.com/google/gopacket"
	"net"
)

/**
 * local pc<--->LAN|WAN<--->Intelnet
 */

type forwardTable struct {
	lrip   net.IP           //PC侧IP
	lrport uint16           //PC侧端口
	lrmac  net.HardwareAddr //pc侧mac

	wrip   net.IP           //Intelnet侧IP
	wrport uint16           //Intelnet侧端口
	wlport uint16           //port NAT
	wrmac  net.HardwareAddr //上行出口mac
}

type forwardKey struct { //上行四元组流匹配
	ip   [8]byte //源和目的IP
	port [4]byte //源和目的端口
}

var forwordtable map[forwardKey]forwardTable = map[forwardKey]forwardTable{}

func forward() {
	go recvLan()
	//go recvWan()
	//go sendLan()
	//go sendWan()
}

func recvLan() {
	lan := nicpool.lan
	wan := nicpool.wan
	lnic := nicpool.lan.nic

	handle := lan.nic.handle
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		if len(packet.Data()) < 34 { //mac header + ip header
			continue
		}

		if bytes.Equal(packet.Data()[0:6], lnic.mac) == false { //promiscuous mode,dst mac != self mac
			continue
		}

		if bytes.Equal(packet.Data()[12:14], []byte{0x08, 0x00}) == false { //skip !ipv4
			continue
		}

		if bytes.Equal(packet.Data()[30:34], lnic.ip) == true { // skip dst ip == self ip
			continue
		}

		if lan.subnet.isIPInSubnet(packet.Data()[30:34]) == false { //filter by subnet
			continue
		}
		fmt.Println("bingo")
		
		//if lan.rflag == false {
		//	lan.rip = make(net.IP, 4)
		//	copy(lan.rip, packet.Data()[26:30])
		//	lan.rmac = make(net.HardwareAddr, 6)
		//	copy(lan.rmac, packet.Data()[6:12])
		//	lan.rflag = true
		//}
		//_, ok := ipmap[string(packet.Data()[30:34])]
		//if !ok {
		//	ipmap[string(packet.Data()[30:34])] = true
		//}

		wan.nic.que <- packet
	}
}

//func sendLan() {
//	lan := nicpool.lan.nic
//	handle := lan.handle
//	for {
//		select {
//		case pkt := <-lan.que:
//			{
//				//replace dst mac
//				copy(pkt.Data()[0:6], lan.rmac)
//				//replace src mac
//				copy(pkt.Data()[6:12], lan.mac)
//
//				//replace dst ip
//				copy(pkt.Data()[30:34], lan.rip)
//
//				//calculate ip checksum
//				copy(pkt.Data()[24:26], []byte{0, 0})
//				checksum := checksum(pkt.Data()[14:34])
//				copy(pkt.Data()[24:26], []byte{byte(checksum >> 8), byte(checksum & 0xff)}) //set ip checksum
//
//				//update tcp checksum
//				if bytes.Equal(pkt.Data()[23:24], []byte{0x06}) == true { //tcp
//					checksum = tcpchecksum(pkt.Data()[34:], pkt.Data()[26:30], pkt.Data()[30:34])
//					copy(pkt.Data()[50:52], []byte{byte(checksum >> 8), byte(checksum & 0xff)}) //set tcp checksum
//				} else if bytes.Equal(pkt.Data()[23:24], []byte{0x11}) == true { //udp
//
//				}
//
//				err := handle.WritePacketData(pkt.Data())
//				if err != nil {
//					log.Println("lan-send: fail to send data, ", err)
//				}
//			}
//		}
//	}
//}
//
//func recvWan() {
//	lan := nicpool.lan.nic
//	wan := nicpool.wan.nic
//	handle := wan.handle
//	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
//	for packet := range packetSource.Packets() {
//		if len(packet.Data()) < 34 { //mac header + ip header
//			continue
//		}
//		if bytes.Equal(packet.Data()[0:6], wan.mac) == false { //promiscuous mode,dst mac != self mac
//			continue
//		}
//
//		if bytes.Equal(packet.Data()[12:14], []byte{0x08, 0x00}) == false { //skip !ipv4
//			continue
//		}
//
//		if bytes.Equal(packet.Data()[30:34], wan.ip) == false { // skip dst ip != self ip
//			continue
//		}
//		_, ok := ipmap[string(packet.Data()[26:30])]
//		if ok {
//
//			lan.que <- packet
//		}
//	}
//}
//
//func sendWan() {
//	wan := nicpool.wan.nic
//	handle := wan.handle
//	for {
//		select {
//		case pkt := <-wan.que:
//			{
//				//replace dst mac
//				copy(pkt.Data()[0:6], wan.mac)
//				//replace src mac
//				copy(pkt.Data()[6:12], wan.mac)
//				//replace source ip
//				copy(pkt.Data()[26:30], wan.ip)
//				copy(pkt.Data()[24:26], []byte{0, 0}) //set ip checksum
//				//calculate ip checksum
//				checksum := checksum(pkt.Data()[14:34])
//				copy(pkt.Data()[24:26], []byte{byte(checksum >> 8), byte(checksum & 0xff)}) //set ip checksum
//
//				//update tcp checksum
//				if bytes.Equal(pkt.Data()[23:24], []byte{0x06}) == true { //tcp
//					checksum = tcpchecksum(pkt.Data()[34:], pkt.Data()[26:30], pkt.Data()[30:34])
//					copy(pkt.Data()[50:52], []byte{byte(checksum >> 8), byte(checksum & 0xff)}) //set tcp checksum
//				} else if bytes.Equal(pkt.Data()[23:24], []byte{0x11}) == true { //udp
//
//				}
//
//				err := handle.WritePacketData(pkt.Data())
//				if err != nil {
//					log.Println("wan-send: fail to send data, ", err)
//				}
//			}
//		}
//	}
//}
//
//func checksum(data []byte) uint16 {
//	var (
//		sum    uint32
//		length int = len(data)
//		index  int
//	)
//
//	// 以每两个字节（16位）为一组进行求和
//	for length > 1 {
//		sum += uint32(binary.BigEndian.Uint16(data[index : index+2]))
//		index += 2
//		length -= 2
//	}
//
//	// 如果字节数为奇数，将最后一个字节单独相加
//	if length > 0 {
//		sum += uint32(data[index])
//	}
//
//	sum += (sum >> 16)
//
//	// 取反得到校验和
//	return uint16(^sum)
//}
//
//func caltcpchecksum(data []byte) uint16 {
//	var sum uint32
//	length := len(data)
//
//	for i := 0; i < length-1; i += 2 {
//		sum += uint32(data[i])<<8 | uint32(data[i+1])
//	}
//
//	if length%2 == 1 {
//		sum += uint32(data[length-1]) << 8
//	}
//
//	for sum > 0xffff {
//		sum = (sum >> 16) + (sum & 0xffff)
//	}
//
//	return uint16(^sum)
//}
//
//type pseudoHeader struct {
//	SourceAddress      [4]byte
//	DestinationAddress [4]byte
//	Zero               uint8
//	Protocol           uint8
//	TCPLength          uint16
//}
//
//func tcpchecksum(data []byte, srcIP, dstIP net.IP) uint16 {
//
//	pHeader := pseudoHeader{}
//	copy(pHeader.SourceAddress[:], srcIP.To4())
//	copy(pHeader.DestinationAddress[:], dstIP.To4())
//	pHeader.Protocol = 6                  // TCP protocol number
//	pHeader.TCPLength = uint16(len(data)) // TCP header length
//
//	pHeaderBytes := make([]byte, 12)
//	binary.BigEndian.PutUint32(pHeaderBytes[0:], binary.BigEndian.Uint32(pHeader.SourceAddress[:]))
//	binary.BigEndian.PutUint32(pHeaderBytes[4:], binary.BigEndian.Uint32(pHeader.DestinationAddress[:]))
//	pHeaderBytes[8] = pHeader.Zero
//	pHeaderBytes[9] = pHeader.Protocol
//	binary.BigEndian.PutUint16(pHeaderBytes[10:], pHeader.TCPLength)
//	data[16] = 0
//	data[17] = 0
//	check := append(pHeaderBytes, data...)
//	return caltcpchecksum(check)
//}

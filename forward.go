package main

import (
	"bytes"
	"encoding/binary"
	"github.com/google/gopacket"
	"log"
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
	go recvWan()
	go sendLan()
	go sendWan()
}

func recvLan() {
	lan := nicpool.lan.nic
	handle := lan.handle
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		if len(packet.Data()) < 34 { //mac header + ip header
			continue
		}
		if bytes.Equal(packet.Data()[0:6], lan.mac) == false { //promiscuous mode,dst mac != self mac
			continue
		}

		if bytes.Equal(packet.Data()[12:14], []byte{0x08, 0x00}) == false { //skip !ipv4
			continue
		}

		if bytes.Equal(packet.Data()[30:34], lan.ip) == true { // skip dst ip == self ip
			continue
		}

		if lan.rflag == false {
			lan.rip = make(net.IP, 4)
			copy(lan.rip, packet.Data()[26:30])
			lan.rmac = make(net.HardwareAddr, 6)
			copy(lan.rmac, packet.Data()[6:12])
			lan.rflag = true
		}
		_, ok := ipmap[string(packet.Data()[30:34])]
		if !ok {
			ipmap[string(packet.Data()[30:34])] = true
		}
		lan.stat.rx += uint32(len(packet.Data()))
		wan.que <- packet
	}
}

func sendLan() {
	lan := nicpool.lan.nic
	handle := lan.handle
	for {
		select {
		case pkt := <-lan.que:
			{
				//replace dst mac
				copy(pkt.Data()[0:6], lan.rmac)
				//replace src mac
				copy(pkt.Data()[6:12], lan.mac)

				//replace dst ip
				copy(pkt.Data()[30:34], lan.rip)

				//calculate ip checksum
				copy(pkt.Data()[24:26], []byte{0, 0})
				checksum := checksum(pkt.Data()[14:34])
				copy(pkt.Data()[24:26], []byte{byte(checksum >> 8), byte(checksum & 0xff)}) //set ip checksum

				//update tcp checksum
				if bytes.Equal(pkt.Data()[23:24], []byte{0x06}) == true { //tcp
					checksum = tcpchecksum(pkt.Data()[34:], pkt.Data()[26:30], pkt.Data()[30:34])
					copy(pkt.Data()[50:52], []byte{byte(checksum >> 8), byte(checksum & 0xff)}) //set tcp checksum
				} else if bytes.Equal(pkt.Data()[23:24], []byte{0x11}) == true { //udp

				}
				lan.stat.tx += uint32(len(pkt.Data()))
				err := handle.WritePacketData(pkt.Data())
				if err != nil {
					lan.stat.txRrr += 1
					log.Println("lan-send: fail to send data, ", err)
				}
			}
		}
	}
}

func recvWan() {
	wan := nicpool.wan.nic
	handle := wan.handle
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		if len(packet.Data()) < 34 { //mac header + ip header
			continue
		}
		if bytes.Equal(packet.Data()[0:6], wan.mac) == false { //promiscuous mode,dst mac != self mac
			continue
		}

		if bytes.Equal(packet.Data()[12:14], []byte{0x08, 0x00}) == false { //skip !ipv4
			continue
		}

		if bytes.Equal(packet.Data()[30:34], wan.ip) == false { // skip dst ip != self ip
			continue
		}
		_, ok := ipmap[string(packet.Data()[26:30])]
		if ok {
			wan.stat.rx += uint32(len(packet.Data()))
			lan.que <- packet
		}
	}
}

func sendWan() {
	wan := nicpool.wan.nic
	handle := wan.handle
	for {
		select {
		case pkt := <-wan.que:
			{
				//replace dst mac
				copy(pkt.Data()[0:6], wan.gwMac)
				//replace src mac
				copy(pkt.Data()[6:12], wan.mac)
				//replace source ip
				copy(pkt.Data()[26:30], wan.ip)
				copy(pkt.Data()[24:26], []byte{0, 0}) //set ip checksum
				//calculate ip checksum
				checksum := checksum(pkt.Data()[14:34])
				copy(pkt.Data()[24:26], []byte{byte(checksum >> 8), byte(checksum & 0xff)}) //set ip checksum

				//update tcp checksum
				if bytes.Equal(pkt.Data()[23:24], []byte{0x06}) == true { //tcp
					checksum = tcpchecksum(pkt.Data()[34:], pkt.Data()[26:30], pkt.Data()[30:34])
					copy(pkt.Data()[50:52], []byte{byte(checksum >> 8), byte(checksum & 0xff)}) //set tcp checksum
				} else if bytes.Equal(pkt.Data()[23:24], []byte{0x11}) == true { //udp

				}
				wan.stat.tx += uint32(len(pkt.Data()))
				err := handle.WritePacketData(pkt.Data())
				if err != nil {
					wan.stat.txRrr += 1
					log.Println("wan-send: fail to send data, ", err)
				}
			}
		}
	}
}

func checksum(data []byte) uint16 {
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

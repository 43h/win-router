package main

import (
	"bytes"
	"encoding/binary"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"log"
	"net"
)

var ipmap map[string]bool = map[string]bool{} //记录上行IP

func forward() {
	go recvLan()
	go recvWan()
	go sendLan()
	go sendWan()
}

func recvLan() {
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

		if bytes.Equal(packet.Data()[26:30], lan.rip) == false {
			continue
		}

		_, ok := ipmap[string(packet.Data()[30:34])]
		if !ok {
			ipmap[string(packet.Data()[30:34])] = true
		}

		lan.stat.rx += uint32(len(packet.Data()))
		lan.stat.rxall += 1
		wan.que <- packet
	}
}

func sendLan() {
	handle := lan.handle
	for {
		select {
		case pkt := <-lan.que:
			{
				data := pkt.Data()
				data = data[:getdatalen(pkt)]
				//replace dst mac
				copy(data[0:6], lan.rmac)
				//replace src mac
				copy(data[6:12], lan.mac)

				//replace dst ip
				copy(data[30:34], lan.rip)

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
				lan.stat.tx += uint32(len(data))
				err := handle.WritePacketData(data)
				if err != nil {
					lan.stat.txRrr += 1
					log.Println("lan-send: fail to send data, ", err)
				} else {
					lan.stat.txall += 1
				}
			}
		}
	}
}

func recvWan() {
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
			wan.stat.rxall += 1
			lan.que <- packet
		}
	}
}

func sendWan() {
	handle := wan.handle
	for {
		select {
		case pkt := <-wan.que:
			{
				data := pkt.Data()
				data = data[:getdatalen(pkt)]
				//replace dst mac
				copy(data[0:6], wan.rmac)
				//replace src mac
				copy(data[6:12], wan.mac)
				//replace source ip
				copy(data[26:30], wan.ip)

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
				wan.stat.tx += uint32(len(data))
				err := handle.WritePacketData(data)
				if err != nil {
					wan.stat.txRrr += 1
					log.Println("wan-send: fail to send data, ", err)
				} else {
					wan.stat.txall += 1
				}
			}
		}
	}
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

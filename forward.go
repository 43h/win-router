package main

import (
	"bytes"
	"encoding/binary"
	"net"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

/**
 * local pc<--->LAN|WAN<--->Intelnet
 */

// 普通IP报文各层偏移
const (
	MACDSTSTART = 0
	MACDSTEND   = 6
	MACSRCSTART = 6
	MACSRCEND   = 12
	ETHTYPESTART
	ETHTYPEEND   = 14
	IPPROTOSTAR  = 23
	IPPROTOEND   = 24
	IPSRCSTART   = 26
	IPSRCEND     = 30
	IPDSTSTART   = 30
	IPDSTEND     = 34
	PORTSRCSTART = 34
	PORTSRCEND   = 36
	PORTDSTSTART = 36
	PORTDSTEND   = 38
)

type threetuple struct {
	ip    [4]byte //上行使用目的IP/下行使用源IP
	port  [2]byte //上行源端口
	port2 [2]byte //上行目的端口
}

type forwardinfo struct {
	ip        [4]byte //上行源IP
	mac       [6]byte //上行源mac
	timestamp int64
}

var forwordtable map[threetuple]forwardinfo = map[threetuple]forwardinfo{}

func forward() {
	lan := &lanNic
	wan := &wanNic

	go rcvPkt(lan)
	go rcvPkt(wan)
	go handleTimeout()

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
			key := threetuple{([4]byte)(data[IPDSTSTART:IPDSTEND]), ([2]byte)(data[PORTSRCSTART:PORTSRCEND]), ([2]byte)(data[PORTDSTSTART:PORTDSTEND])}
			if bytes.Equal(data[IPPROTOSTAR:IPPROTOEND], []byte{0x01}) == true {
				key.port = [2]byte{0, 0}
				key.port2 = [2]byte{0, 0}
			}
			fwdinfo, exist := forwordtable[key]
			if exist { //记录存在
				sip := fwdinfo.ip[:]
				if bytes.Equal(data[IPSRCSTART:IPSRCEND], sip) == true { //记录已存在可以转发
					fwdinfo.timestamp = time.Now().Unix()
					handleUpstreamPkt(pkt)
				}
			} else { //添加新记录
				//检测源IP
				copy(fwdinfo.ip[:], data[IPSRCSTART:IPSRCEND])
				copy(fwdinfo.mac[:], data[MACSRCSTART:MACSRCEND])
				fwdinfo.timestamp = time.Now().Unix()
				forwordtable[key] = fwdinfo
				handleUpstreamPkt(pkt)
			}
		} else { //from wan, to lan
			key := threetuple{([4]byte)(data[IPSRCSTART:IPSRCEND]), ([2]byte)(data[PORTDSTSTART:PORTDSTEND]), ([2]byte)(data[PORTSRCSTART:PORTSRCEND])}
			if bytes.Equal(data[IPPROTOSTAR:IPPROTOEND], []byte{0x01}) == true {
				key.port = [2]byte{0, 0}
				key.port2 = [2]byte{0, 0}
			}
			fwdinfo, exists := forwordtable[key]
			if exists {
				handleDownstreamPkt(pkt, fwdinfo)
			}
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

		if bytes.Equal(packet.Data()[ETHTYPESTART:ETHTYPEEND], []byte{0x08, 0x00}) == false { //skip !ipv4
			continue
		}

		if nic.nicType == NICLAN {
			if bytes.Equal(packet.Data()[IPDSTSTART:IPDSTEND], nic.ip) == true { // skip dst ip == self ip
				continue
			}
		}

		if bytes.Equal(packet.Data()[IPPROTOSTAR:IPPROTOEND], []byte{0x06}) == true { //tcp
			nic.que <- packet
		} else if bytes.Equal(packet.Data()[IPPROTOSTAR:IPPROTOEND], []byte{0x01}) == true { //icmp
			nic.que <- packet
		} //udp to do
	}
}

func handleUpstreamPkt(pkt gopacket.Packet) bool {
	nic := &wanNic
	data := pkt.Data()
	data = data[:getdatalen(pkt)]

	copy(data[MACDSTSTART:MACDSTEND], nic.gwmac)
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
		return true
	}
	return false
}

func handleDownstreamPkt(pkt gopacket.Packet, fwdinfo forwardinfo) bool {
	nic := &lanNic
	data := pkt.Data()
	data = data[:getdatalen(pkt)]

	copy(data[MACDSTSTART:MACDSTEND], fwdinfo.mac[:])
	copy(data[MACSRCSTART:MACSRCEND], nic.mac)

	copy(data[IPDSTSTART:IPDSTEND], fwdinfo.ip[:])
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
	err := nic.handle.WritePacketData(data)
	if err == nil {

		return true
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

func handleTimeout() {
	for {
		for k, v := range forwordtable {
			if time.Now().Unix()-v.timestamp > 10 { //超过10秒无报文，则删除转发信息
				delete(forwordtable, k)
			}
		}
		time.Sleep(300 * time.Millisecond)
	}
}

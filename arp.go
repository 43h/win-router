package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
	"time"

	"github.com/mdlayher/arp"
)

type ARPItem struct {
	ip  net.IP
	mac net.HardwareAddr
}

var ARPTable map[string](ARPItem) = map[string](ARPItem){}

func readARPTable() bool {
	cmd := exec.Command("arp", "-a")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Println("fail to read arp table, ", err)
		return false
	} else {
		parseARPTable(out.String())
	}
	return true
}

func parseARPTable(output string) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) == 3 {
			ip := fields[0]
			mac := fields[1]
			_, ok := ARPTable[ip]
			if !ok {
				item := ARPItem{net.IP(ip), net.HardwareAddr(mac)}
				ARPTable[ip] = item
			}
		}
	}
}

func getArpByIp(ip net.IP) net.HardwareAddr {
	v, ok := ARPTable[string(ip)]
	if ok {
		if string(ip) == string(v.ip) {
			return v.mac
		}
	}
	return nil
}

func request() {
	// 设置网络接口
	ifi, err := net.InterfaceByName("eth0")
	if err != nil {
		log.Fatal(err)
	}

	// 创建ARP客户端
	client, err := arp.Dial(ifi)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// 设置目标IP地址
	ip := net.IPv4(192, 168, 1, 1)

	// 发送ARP请求
	if err := client.Request(ip); err != nil {
		log.Fatal(err)
	}

	// 等待ARP回复
	for {
		msg, _, err := client.Read()
		if err != nil {
			log.Fatal(err)
		}

		// 如果收到的是ARP回复，并且源IP地址与目标IP地址相同，则打印源硬件地址
		if msg.Operation == arp.OperationReply && msg.SenderIP.Equal(ip) {
			log.Println("ARP reply from", msg.SenderHardwareAddr)
			return
		}

		// 如果在5秒内没有收到ARP回复，则退出
		if time.Since(msg.Time) > 5*time.Second {
			log.Fatal("timeout")
		}
	}
}

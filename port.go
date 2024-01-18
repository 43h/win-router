package main

import (
	"fmt"
	"net"
)

type Port struct {
    uint16 port
	uint16 used
}

type PortPool struct {
uint16 start
uint16 end
uint16 size
uint16 used
uint16 next
Port *port
uint16
}

func initPort(uint16 start, uint16 end) *PortPool {
	p := &PortPool{}
	p.start = start
	p.end = end
	p.size = end - start + 1
	p.used = 0
	p.next = 0
	p.port = make([]Port, p.size)
	for i := 0; i < p.size; i++ {
		p.port[i].port = start + i
		p.port[i].used = 0
	}
	return p
}



// 分配端口
func (*PortPool)allocatePort() uint16{
    if p.used == p.size {
		return 0
	}

	for{
		if p.port[p.next].used == 0 {
			p.port[p.next].used = 1
			p.used++
	
			p.next++
			if p.next >= p.size {
				p.next = 0
			}
	
			return p.port[p.next].port
		
	}
	
return 0
}

// 检测端口是否被占用
func isPortUsed(port int) (bool) {
	return false
}

// 回收端口
func releasePort(port int){
	return nil
}
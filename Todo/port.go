package main

import (
	"log"
)

type Port struct {
	port uint16
	used uint16
}

type PortPool struct {
	start uint16
	end   uint16
	size  uint16
	used  uint16
	next  uint16
	port  []Port
}

func initPortPool(start, end uint16) *PortPool {
	p := &PortPool{}
	p.start = start
	p.end = end
	p.size = end - start + 1
	p.used = 0
	p.next = 0
	p.port = make([]Port, p.size)
	var i uint16
	for i = 0; i < p.size; i++ {
		p.port[i].port = start + i
		p.port[i].used = 0
	}
	return p
}

// 分配端口
func (p *PortPool) allocatePort() uint16 {
	if p.used == p.size {

		return 0
	}
	for {
		if p.port[p.next].used == 0 {
			p.port[p.next].used = 1
			p.used++
			return p.port[p.next].port
		}
		p.next = (p.next + 1) % p.size
	}
	return 0
}

func (p *PortPool) freePort(port uint16) {
	if port >= p.start && port <= p.end {
		p.port[port-p.start].used = 0
		p.used--
	}
}

func (p *PortPool) isPortUsed(port uint16) bool {
	if port >= p.start && port <= p.end {
		return p.port[port-p.start].used == 1
	}
	return false
}

func (p *PortPool) isFull() bool {
	return p.used == p.size
}
func (p *PortPool) uesd() uint16 {
	return p.used
}

func (p *PortPool) dumpPortPool() {
	log.Println("--port pool--")
	log.Println("start:", p.start)
	log.Println("end:  ", p.end)
	log.Println("size: ", p.size)
	log.Println("used: ", p.used)
	log.Println("next: ", p.next)
}

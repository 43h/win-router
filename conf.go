package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

const confFile = "conf.txt"

const (
	CONFWANNAME = iota
	CONFWANGW   //WAN口网关
	CONFLANNAME
)

func loadConf() bool {
	f, err := os.Open(confFile)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if parseLine(scanner.Text()) == false {
			return false
		}
	}
	return true
}

func parseLine(line string) bool {
	parts := strings.Split(line, ":")
	if len(parts) != 2 {
		return false
	}

	key := parts[0]
	value := parts[1]
	fmt.Println(key, value)
	if key == "wan-name" {
		addNic(CONFWANNAME, value)
	} else if key == "lan-name" {
		addNic(CONFLANNAME, value)
	} else if key == "wan-gwip" {
		if setNicGw(CONFWANGW, value) == false {
			return false
		}
	} else {
		log.Println("  unknown param ", key, value)
		return false
	}
	return true
}

func dumpConf() bool {
	f, err := os.Open(confFile)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

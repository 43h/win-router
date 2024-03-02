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
	CONFLANGW //LAN口网关
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
		parseLine(scanner.Text())
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
	if key == "wan" {
		addNic(CONFWANNAME, value)
	} else if key == "lan" {
		addNic(CONFLANNAME, value)
	} else if key == "wangw" {
		setNicGw(CONFWANGW, value)
	} else if key == "langw" {
		setNicGw(CONFLANGW, value)
	} else {
		log.Println("  unknown param")
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

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const confFile = "conf.txt"

const (
	CONFWANNAME = iota
	CONFWANMAC
	CONFWANPORT
	CONFLANNAME
	CONFLANIP
)

func loadConf() {
	f, err := os.Open(confFile)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		parseLine(scanner.Text())
	}
}

func parseLine(line string) bool {
	parts := strings.Split(line, ":")
	if len(parts) != 2 {
		return false
	}

	key := parts[0]
	value := parts[1]

	if key == "wan" {
		parseNicConf(CONFWANNAME, value)
	} else if key == "wmac" {
		parseNicConf(CONFWANMAC, value)
	} else if key == "wport" {
		parseNicConf(CONFWANPORT, value)
	} else if key == "lan" {
		parseNicConf(CONFLANNAME, value)
	} else if key == "lip" {
		parseNicConf(CONFLANIP, value)
	} else {
		fmt.Println("  unknown param")
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

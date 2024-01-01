package main

import (
	"bufio"
	"fmt"
	"os"
)

const confFile = "conf.txt"

func loadConf() {
	f, err := os.Open(confFile)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		parseNic(scanner.Text())
	}
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

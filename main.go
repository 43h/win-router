package main

import (
	"fmt"
	"os"
)

const version = "0.1.2-1to1-20240121\n"

func parseArg(arg []string) bool {
	var argNum = len(arg)
	if argNum == 1 {
		return true
	} else if argNum == 2 {
		if arg[1] == "-h" || arg[1] == "--help" {
			showHelpInfo()
		} else if arg[1] == "-n" {
			showNicInfo2()
		} else if arg[1] == "-nn" {
			showNicInfo()
		} else if arg[1] == "-d" {
			dumpConf()
		} else {
			fmt.Println("  unknow param\nexit")
			showHelpInfo()
		}
		return false
	} else {
		fmt.Println("  unknow param")
		showHelpInfo()
		return false
	}

	return true
}

func main() {
	fmt.Println("win-router version:", version)
	if parseArg(os.Args) == false {
		return
	}

	initLog()
	defer closeLog()

	loadConf()

	initNics()
	defer closeNics()

	forward()

	showStat()
}

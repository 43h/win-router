package main

import (
	"fmt"
	"os"
)

const version = "0.1.2-1to1-20240121\n"
func main() {
	fmt.Println("win-router version:", version)
	if parseArg(os.Args) == false {
		return
	}

	initLog()
	defer closeLog()

	loadConf()

	initNicPool()
	defer destoryNicPool()
	dumpNicPool()

	forward()
	for {
	}
}

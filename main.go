package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	if parseArg(os.Args) == false {
		fmt.Println("  parseArg failed\nexit")
		return
	}

	if initLog() == false {
		fmt.Println("  initLog failed\nexit")
		return
	}
	defer closeLog()

	if loadConf() == false {
		fmt.Println("  loadConf failed\nexit")
		return
	}

	dumpNics()

	forward()
	for {
		time.Sleep(100 * time.Second)
	}
}

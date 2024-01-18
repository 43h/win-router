package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println(showVersion())

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

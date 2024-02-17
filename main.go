package main

import (
	"os"
)

func main() {
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

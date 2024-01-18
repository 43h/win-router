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

	if parseConf() == false{
		return
	}

	initLog()
	defer closeLog()

	if checkNic() == false {
		return
	}

	if initNic() == false {
		return
	}
	defer closeNic()
	
	forward()
	showStat()
}
package main

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
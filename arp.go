package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func getMacByIp(ip string) (string, bool) {
	cmd := exec.Command("arp", "-a")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Println("fail to read arp table, ", err)
		return "", false
	}

	lines := strings.Split(out.String(), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) == 3 {
			if fields[0] == ip {
				return fields[1], true
			}
		}
	}
	return "", false
}

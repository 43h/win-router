package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

var logHandle *os.File

func initLog() bool {
	fileName := "log" + time.Now().Format("2006-01-02-15-04-05") + ".txt"
	logHandle, err := os.OpenFile(fileName, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("fail to open file, ", err)
		return false
	}

	log.SetOutput(logHandle)
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	return true
}

func closeLog() {
	logHandle.Close()
}

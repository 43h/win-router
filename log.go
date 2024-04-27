package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var logFileFlag = true
var logHandle *os.File

func initLog() bool {
	if logFileFlag == false {
		return true
	}

	delLog()
	fileName := time.Now().Format("2006-01-02-15-04-05") + ".log"
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
	if logHandle != nil {
		err := logHandle.Close()
		if err != nil {
			log.Println("fail to close file, ", err)
		}
	}
}

func delLog() {
	dir := "./"
	//clean log files
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".log") {
			err := os.Remove(path)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		log.Println("fail to del log, ", err)
	}
}

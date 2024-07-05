package main

import "fmt"

const productName = "win-router"
const version = "0.2.3"
const updatetime = "20240705"

func showVersion() {
	fmt.Println(productName + "\n" + version + "." + updatetime)
}

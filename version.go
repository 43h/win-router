package main

import "fmt"

const productName = "win-router"
const version = "0.2.2"
const updatetime = "20240427"

func showVersion() {
	fmt.Println(productName + "\n" + version + "." + updatetime)
}

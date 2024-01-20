package main

import (
	"fmt"
	"github.com/rivo/tview"
	"time"
)

type statNic struct {
	rx    uint32
	tx    uint32
	txRrr uint32
	rxall uint32
	txall uint32
}

// 用于打印输出
var lanStat statNic
var wanStat statNic

func doStat() {
	time.Sleep(time.Second)
	lanStat.rx = lan.stat.rx
	lanStat.tx = lan.stat.tx
	lanStat.txRrr += lan.stat.txRrr
	lanStat.rxall += lan.stat.rx
	lanStat.txall += lan.stat.tx

	wanStat.rx = wan.stat.rx
	wanStat.tx = wan.stat.tx
	wanStat.txRrr += wan.stat.txRrr
	wanStat.rxall += wan.stat.rx
	wanStat.txall += wan.stat.tx

	lan.stat = statNic{0, 0, 0, 0, 0}
	wan.stat = statNic{0, 0, 0, 0, 0}
}

const refreshInterval = 1000 * time.Millisecond

var (
	view *tview.Modal
	app  *tview.Application
)

func currentStatString() string {
	t := time.Now()

	str := fmt.Sprintf(t.Format("time 15:04:05"))
	str += "\n---lan---\n"
	str += fmt.Sprintf("    rx: %.3f KB/s\n", float32(lanStat.rx)/1000)
	str += fmt.Sprintf("    tx: %.3f KB/s\n", float32(lanStat.tx)/1000)
	str += fmt.Sprintf("all-rx: %d \n", lanStat.rxall)
	str += fmt.Sprintf("all-tx: %d \n", lanStat.txall)
	str += fmt.Sprintf(" TxErr: %d\n", lanStat.txRrr)

	str += fmt.Sprintf("\n---wan---\n")
	str += fmt.Sprintf("    rx: %.3f KB/s\n", float32(wanStat.rx)/1000)
	str += fmt.Sprintf("    tx: %.3f KB/s\n", float32(wanStat.tx)/1000)
	str += fmt.Sprintf(" TxErr: %d\n", wanStat.txRrr)
	str += fmt.Sprintf("all-rx: %d \n", wanStat.rxall)
	str += fmt.Sprintf("all-tx: %d \n", wanStat.txall)
	return str
}

func updateStat() {
	for {
		time.Sleep(refreshInterval)
		doStat()
		app.QueueUpdateDraw(func() {
			view.SetText(currentStatString())
		})
	}
}

func showStat() {
	app = tview.NewApplication()
	view = tview.NewModal().
		SetText(currentStatString()).
		AddButtons([]string{"Quit", "..."}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Quit" {
				app.Stop()
			}
		})

	go updateStat()
	if err := app.SetRoot(view, false).Run(); err != nil {
		panic(err)
	}
}

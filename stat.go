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
}

var lanStat statNic
var wanStat statNic

func doStat() {
	time.Sleep(time.Second)
	lanStat = lan.stat
	lan.stat = statNic{0, 0, lan.stat.txRrr}
	wanStat = wan.stat
	wan.stat = statNic{0, 0, wan.stat.txRrr}
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
	str += fmt.Sprintf(" TxErr: %d\n", lanStat.txRrr)

	str += fmt.Sprintf("\n---wan---\n")
	str += fmt.Sprintf("    rx: %.3f KB/s\n", float32(wanStat.rx)/1000)
	str += fmt.Sprintf("    tx: %.3f KB/s\n", float32(wanStat.tx)/1000)
	str += fmt.Sprintf(" TxErr: %d\n", wanStat.txRrr)

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
		AddButtons([]string{"Quit"}).
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

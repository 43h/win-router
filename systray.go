package main

import (
	"github.com/getlantern/systray"
	"github.com/getlantern/systray/example/icon"
)

func main() {
	// The systray.Run function takes a onReady function and an onExit function
	systray.Run(onReady, onExit)
}

func onReady() {
	// Set the icon in the system tray
	systray.SetIcon(icon.Data)

	// Set the tooltip to appear when you hover over the icon
	systray.SetTooltip("This is a tooltip")

	// Create a menu item
	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")

	// This goroutine keeps running in the background until mQuit is clicked, at which point we quit the app
	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()
}

func onExit() {
	// This is where you'd put any cleanup code for when your app is about to quit
}

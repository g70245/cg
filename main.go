package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const (
	CLASS = "Blue"
	ON    = "On"
	OFF   = "Off"
)

func main() {
	targetWindows, _ := FindWindows(CLASS)
	targetWindowStrs := targetWindows.Get()
	stopChan := make(chan bool, len(targetWindows))

	// base
	myApp := app.New()
	window := myApp.NewWindow("CG")
	window.Resize(fyne.NewSize(200, 300))

	// choose the party leader
	dropDownLabel := widget.NewLabel("Lead PID")
	dropDown := widget.NewSelect(targetWindowStrs, func(pid string) {
		GLOBAL_PARTY_LEAD_HWND = pid
	})
	dropDown.PlaceHolder = "Choose the party lead"
	dropDownForm := container.NewHBox(dropDownLabel, dropDown)

	// lever
	var lever *widget.Button
	lever = widget.NewButton(ON, func() {
		switch lever.Text {
		case ON:
			for _, window := range targetWindows {
				Worker(window, stopChan)
			}
			turn(OFF, lever)
		case OFF:
			for range targetWindows {
				stopChan <- true
			}
			turn(ON, lever)
		}
	})

	// cancel the chosen party leader
	clear := widget.NewButton("Clear", func() {
		dropDown.ClearSelected()
	})

	// container
	buttonGrid := container.NewGridWithColumns(2, lever, clear)
	mainControl := container.NewVBox(dropDownForm, buttonGrid)
	window.SetContent(mainControl)
	window.ShowAndRun()

	close(stopChan)
	fmt.Println("Exit")
}

func turn(text string, button *widget.Button) {
	button.SetText(text)
	button.Refresh()
}

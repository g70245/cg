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
	hWnds, _ := FindWindows(CLASS)
	workers := CreateWorkers(hWnds)
	handles := workers.GetHandles()
	stopChan := make(chan bool, len(workers))

	// base
	myApp := app.New()
	window := myApp.NewWindow("CG")
	window.Resize(fyne.NewSize(200, 300))

	// choose the party leader
	dropDownLabel := widget.NewLabel("Lead PID")
	dropDown := widget.NewSelect(handles, func(pid string) {
		GLOBAL_PARTY_LEAD_HWND = pid
	})
	dropDown.PlaceHolder = "Choose the party lead"
	dropDownForm := container.NewHBox(dropDownLabel, dropDown)

	// lever
	var lever *widget.Button
	lever = widget.NewButton(ON, func() {
		switch lever.Text {
		case ON:
			for _, w := range workers {
				w.Work(stopChan)
			}
			turn(OFF, lever)
		case OFF:
			for range workers {
				stopChan <- true
			}
			turn(ON, lever)
		}
	})

	// strategy
	// movementStrategies := make([]string, len(targetWindows))
	// for i := range movementStrategies {
	// 	movementStrategies[i] = MOVEMENT_STRATEGIES[0]
	// }

	// for i := range targetWindows {
	// 	label := widget.NewLabel(targetWindowStrs[i])
	// }

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

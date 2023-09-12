package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
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
	window.Resize(fyne.NewSize(340, 160))

	// choose the party leader
	leadSelectorLabel := widget.NewLabel("Lead Handle")
	leadSelectorLabel.TextStyle = fyne.TextStyle{Italic: true, Bold: true}
	leadSelector := widget.NewSelect(handles, func(handle string) {
		GLOBAL_PARTY_LEAD_HWND = handle
	})
	leadSelector.PlaceHolder = "Choose lead"
	leadSelectorContainer := container.New(layout.NewFormLayout(), leadSelectorLabel, leadSelector)

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

	// configuration for every worker
	for i := range workers {
		workers[i].movementMode = MovementMode(MOVEMENT_MODES[0])
	}

	configContainer := container.New(layout.NewFormLayout())
	for i := range workers {
		worker := &workers[i]
		movementModeSelectorLabel := widget.NewLabel("Handle " + worker.GetHandle())
		movementModeSelectorLabel.TextStyle = fyne.TextStyle{Italic: true, Bold: true}
		movementModeSelector := widget.NewSelect(MOVEMENT_MODES, func(movementMode string) {
			worker.movementMode = MovementMode(movementMode)
		})
		movementModeSelector.PlaceHolder = MOVEMENT_MODES[0]
		configContainer.Add(movementModeSelectorLabel)
		configContainer.Add(movementModeSelector)
	}

	// cancel the chosen party leader
	clear := widget.NewButton("Clear", func() {
		leadSelector.ClearSelected()
	})

	// container
	buttonGrid := container.NewGridWithColumns(2, lever, clear)
	mainControl := container.NewVBox(leadSelectorContainer, buttonGrid)
	separator := widget.NewSeparator()
	content := container.NewVBox(mainControl, separator, configContainer)
	window.SetContent(content)
	window.ShowAndRun()

	close(stopChan)
	fmt.Println("Exit")
}

func turn(text string, button *widget.Button) {
	button.SetText(text)
	button.Refresh()
}

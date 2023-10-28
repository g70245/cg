package main

import (
	. "cg/game"
	. "cg/system"

	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/exp/slices"
)

func productionContainer(games Games) (*fyne.Container, map[string]chan bool) {
	stopChans := make(map[string]chan bool)
	productions := make(map[string]*fyne.Container)

	gamesCheckGroup := widget.NewCheckGroup(games.GetAll(), nil)
	gamesCheckGroup.Horizontal = true

	productionsContainer := container.NewVBox()
	newProductionButton := widget.NewButtonWithIcon("New Production", theme.ContentAddIcon(), func() {
		gamesSelectorDialog := dialog.NewCustom("Select games", "Create", gamesCheckGroup, window)
		gamesSelectorDialog.Resize(fyne.NewSize(240, 166))

		gamesSelectorDialog.SetOnClosed(func() {
			for _, game := range games.GetAll() {
				if slices.Contains(gamesCheckGroup.Selected, game) {
					if _, ok := stopChans[game]; !ok {
						newProductionWidget, newStopChan := newProductionContainer(game, games, nil)
						stopChans[game] = newStopChan
						productions[game] = newProductionWidget
						productionsContainer.Add(newProductionWidget)
					}
				} else {
					stop(stopChans[game])
					productionsContainer.Remove(productions[game])
					delete(stopChans, game)
					delete(productions, game)
				}
			}

			window.SetContent(window.Content())
			window.Resize(fyne.NewSize(APP_WIDTH, APP_HEIGHT))
		})
		gamesSelectorDialog.Show()
	})

	main := container.NewVBox(newProductionButton, productionsContainer)
	return main, stopChans
}

func newProductionContainer(handle string, games Games, destroy func()) (productionWidget *fyne.Container, stopChan chan bool) {
	stopChan = make(chan bool, 1)
	worker := CreateProductionWorker(games.Peek(handle), logDir)

	var nicknameButton *widget.Button
	nicknameEntry := widget.NewEntry()
	nicknameEntry.SetPlaceHolder("Enter nickname")
	nicknameButton = widget.NewButtonWithIcon(handle, theme.AccountIcon(), func() {
		nicknameDialog := dialog.NewCustom("Enter nickname", "Ok", nicknameEntry, window)
		nicknameDialog.SetOnClosed(func() {
			nickname := ""
			if nicknameEntry.Text != "" {
				nickname = fmt.Sprintf("(%s)", nicknameEntry.Text)
			}
			nicknameButton.SetText(fmt.Sprintf("%s%s", handle, nickname))
		})
		nicknameDialog.Show()
	})
	nicknameButton.Alignment = widget.ButtonAlignLeading

	var lever *widget.Button
	lever = widget.NewButtonWithIcon("", theme.MediaPlayIcon(), func() {
		switch lever.Icon {
		case theme.MediaPlayIcon():
			worker.Work(stopChan)
			turn(theme.MediaStopIcon(), lever)
		case theme.MediaStopIcon():
			stopChan <- true
			StopBeeper()
			turn(theme.MediaPlayIcon(), lever)
		}
	})

	productionWidget = container.NewGridWithColumns(6, nicknameButton, lever)
	return
}

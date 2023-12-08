package main

import (
	. "cg/game"

	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/exp/slices"
)

type ProductionWorkers struct {
	containers map[string]*fyne.Container
	stopChans  map[string]chan bool
}

func (pw *ProductionWorkers) doesExist(game string) bool {
	_, ok := pw.stopChans[game]
	return ok
}

func (pw *ProductionWorkers) stop(game string) {
	pw.stopChans[game] <- true
	close(pw.stopChans[game])
}

func (pw *ProductionWorkers) stopAll() {
	for k := range pw.stopChans {
		pw.stopChans[k] <- true
	}
}

func productionContainer(games Games) (*fyne.Container, ProductionWorkers) {

	pw := ProductionWorkers{make(map[string]*fyne.Container), make(map[string]chan bool)}

	productionsContainer := container.NewVBox()
	newProductionButton := widget.NewButtonWithIcon("New Production", theme.ContentAddIcon(), func() {

		gamesCheckGroup := widget.NewCheckGroup(games.GetSortedKeys(), nil)
		gamesCheckGroup.Horizontal = true

		gamesSelectorDialog := dialog.NewCustom("Select games", "Create", gamesCheckGroup, window)
		gamesSelectorDialog.Resize(fyne.NewSize(240, 166))

		gamesSelectorDialog.SetOnClosed(func() {
			for _, game := range games.GetSortedKeys() {
				if slices.Contains(gamesCheckGroup.Selected, game) {
					if !pw.doesExist(game) {
						newContainer, newStopChan := newProductionContainer(game, games, nil)
						pw.containers[game] = newContainer
						pw.stopChans[game] = newStopChan
						productionsContainer.Add(newContainer)
					}
				} else {
					if pw.doesExist(game) {
						productionsContainer.Remove(pw.containers[game])

						pw.stop(game)
						delete(pw.stopChans, game)
						delete(pw.containers, game)
					}
				}
			}

			window.SetContent(window.Content())
			window.Resize(fyne.NewSize(APP_WIDTH, APP_HEIGHT))
		})
		gamesSelectorDialog.Show()

		notifyBeeperAndLogConfig("About Production")
	})

	main := container.NewVBox(newProductionButton, productionsContainer)
	return main, pw
}

func newProductionContainer(handle string, games Games, destroy func()) (*fyne.Container, chan bool) {
	stopChan := make(chan bool, 1)
	worker := CreateProductionWorker(games.Peek(handle), logDir, stopChan)

	var nicknameButton *widget.Button
	nicknameEntry := widget.NewEntry()
	nicknameEntry.SetPlaceHolder("Enter nickname")
	nicknameButton = widget.NewButtonWithIcon(handle, theme.AccountIcon(), func() {
		nicknameDialog := dialog.NewCustom("Enter nickname", "Ok", nicknameEntry, window)
		nicknameDialog.SetOnClosed(func() {
			nickname := ""
			if nicknameEntry.Text != "" {
				worker.Name = nicknameEntry.Text
				nickname = fmt.Sprintf("(%s)", nicknameEntry.Text)
			} else {
				worker.Name = NAME_NONE
			}
			nicknameButton.SetText(fmt.Sprintf("%s%s", handle, nickname))
		})
		nicknameDialog.Show()
	})
	nicknameButton.Alignment = widget.ButtonAlignLeading

	var isGatheringButton *widget.Button
	isGatheringButton = widget.NewButtonWithIcon("Gathering", theme.CheckButtonIcon(), func() {
		switch isGatheringButton.Icon {
		case theme.CheckButtonCheckedIcon():
			worker.GatheringMode = false
			turn(theme.CheckButtonIcon(), isGatheringButton)
		case theme.CheckButtonIcon():
			worker.GatheringMode = true
			turn(theme.CheckButtonCheckedIcon(), isGatheringButton)
		}
	})

	var switchButton *widget.Button
	switchButton = widget.NewButtonWithIcon("", theme.MediaPlayIcon(), func() {
		switch switchButton.Icon {
		case theme.MediaPlayIcon():
			worker.Work()
			turn(theme.MediaStopIcon(), switchButton)
		case theme.MediaStopIcon():
			worker.Stop()
			turn(theme.MediaPlayIcon(), switchButton)
		}
	})

	return container.NewGridWithColumns(6, nicknameButton, isGatheringButton, switchButton), stopChan
}

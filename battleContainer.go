package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/lxn/win"
	"golang.org/x/exp/maps"
)

func battleContainer(idleGames Games) (*fyne.Container, map[int]chan bool) {
	autoGroups := make(map[int]map[string]win.HWND)
	stopChans := make(map[int]chan bool)

	groupTabs := container.NewAppTabs()
	groupTabs.SetTabLocation(container.TabLocationBottom)
	groupTabs.Hide()

	newGroupButton := widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
		if len(autoGroups) == 3 {
			return
		}

		var newGroup map[string]win.HWND
		gamesChoosingDialog := dialog.NewCustom("Choose Games", "create", widget.NewCheckGroup(maps.Keys(idleGames), func(games []string) {
			newGroup = make(map[string]win.HWND)
			for _, game := range games {
				newGroup[game] = idleGames.peek(game)
			}
		}), window)

		gamesChoosingDialog.SetOnClosed(func() {
			if len(newGroup) == 0 {
				return
			}

			idleGames.remove(maps.Keys(newGroup))

			newGroupContainer, stopChan := newBatttleGroupContainer(newGroup)
			id := len(autoGroups)
			autoGroups[id] = newGroup
			stopChans[id] = stopChan

			groupTabs.Append(container.NewTabItem("Group "+fmt.Sprint(len(autoGroups)), newGroupContainer))
			groupTabs.Show()

		})
		gamesChoosingDialog.Show()
	})

	menu := container.NewVBox(widget.NewSeparator(), container.NewHBox(newGroupButton))

	main := container.NewBorder(menu, nil, nil, nil, groupTabs)
	return main, stopChans
}

const (
	ON  = "On"
	OFF = "Off"
)

func newBatttleGroupContainer(games map[string]win.HWND) (*fyne.Container, chan bool) {
	var leadHandle string
	workers := CreateBattleWorkers(maps.Values(games))
	stopChan := make(chan bool, len(workers))

	/* Main Widget */
	leadSelectorLabel := widget.NewLabel("Leader")
	leadSelectorLabel.TextStyle = fyne.TextStyle{Italic: true, Bold: true}
	leadSelector := widget.NewSelect(maps.Keys(games), func(handle string) {
		leadHandle = handle
	})
	leadSelector.PlaceHolder = "Choose a leader"
	leadSelectorContainer := container.New(layout.NewFormLayout(), leadSelectorLabel, leadSelector)

	var lever *widget.Button
	lever = widget.NewButtonWithIcon("", theme.MediaPlayIcon(), func() {
		switch lever.Icon {
		case theme.MediaPlayIcon():
			for _, w := range workers {
				w.Work(&leadHandle, stopChan)
			}
			turn(theme.MediaStopIcon(), lever)
		case theme.MediaStopIcon():
			for range workers {
				stopChan <- true
			}
			turn(theme.MediaPlayIcon(), lever)
		}
	})

	refresh := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() {
		leadSelector.ClearSelected()
	})

	delete := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
	})
	mainButtons := container.NewGridWithColumns(3, lever, refresh, delete)
	mainWidget := container.NewVBox(mainButtons, leadSelectorContainer)

	/* Configuration Widget */
	configContainer := container.New(layout.NewFormLayout())
	for i := range workers {
		worker := &workers[i]
		movementModeSelectorLabel := widget.NewLabel("Game " + worker.getHandle())
		movementModeSelectorLabel.TextStyle = fyne.TextStyle{Italic: true, Bold: true}
		movementModeSelector := widget.NewSelect(BATTLE_MOVEMENT_MODES, func(movementMode string) {
			worker.movementMode = BattleMovementMode(movementMode)
		})
		movementModeSelector.PlaceHolder = BATTLE_MOVEMENT_MODES[0]
		configContainer.Add(movementModeSelectorLabel)
		configContainer.Add(movementModeSelector)
	}

	autoBattleWidget := container.NewVBox(mainWidget, configContainer)
	return autoBattleWidget, stopChan
}

func turn(icon fyne.Resource, button *widget.Button) {
	button.SetIcon(icon)
	button.Refresh()
}

func stop(stopChan chan bool) {
	i := 0
	for i < cap(stopChan) {
		stopChan <- true
		i++
	}
}

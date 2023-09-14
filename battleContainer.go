package main

import (
	. "cg/game"

	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/lxn/win"
	. "github.com/lxn/win"
	"golang.org/x/exp/maps"
)

func battleContainer(idleGames Games) (*fyne.Container, map[int]chan bool) {
	id := 0
	autoGroups := make(map[int]map[string]HWND)
	stopChans := make(map[int]chan bool)

	groupTabs := container.NewAppTabs()
	groupTabs.SetTabLocation(container.TabLocationBottom)
	groupTabs.Hide()

	newGroupButton := widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
		if len(autoGroups) == 3 {
			return
		}

		var newGroup map[string]HWND
		gamesChoosingDialog := dialog.NewCustom("Choose Games", "create", widget.NewCheckGroup(maps.Keys(idleGames), func(games []string) {
			newGroup = make(map[string]HWND)
			for _, game := range games {
				newGroup[game] = idleGames.Peek(game)
			}
		}), window)

		gamesChoosingDialog.SetOnClosed(func() {
			if len(newGroup) == 0 {
				return
			}

			idleGames.Remove(maps.Keys(newGroup))
			var newTabItem *container.TabItem

			newGroupContainer, stopChan := newBatttleGroupContainer(newGroup, func(id int) func() {
				return func() {
					delete(autoGroups, id)
					delete(stopChans, id)
					idleGames.Add(newGroup)
					groupTabs.Remove(newTabItem)
					if len(autoGroups) == 0 {
						groupTabs.Hide()
					}
				}
			}(id))
			autoGroups[id] = newGroup
			stopChans[id] = stopChan

			newTabItem = container.NewTabItem("Group "+fmt.Sprint(id), newGroupContainer)
			groupTabs.Append(newTabItem)
			groupTabs.Show()
			id++
		})
		gamesChoosingDialog.Show()
	})

	menu := container.NewVBox(container.NewHBox(newGroupButton))

	main := container.NewBorder(menu, nil, nil, nil, groupTabs)
	return main, stopChans
}

const (
	ON  = "On"
	OFF = "Off"
)

func newBatttleGroupContainer(games map[string]win.HWND, destroy func()) (autoBattleWidget *fyne.Container, stopChan chan bool) {
	var leadHandle string
	workers := CreateBattleWorkers(maps.Values(games))
	stopChan = make(chan bool, len(workers))

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
		stop(stopChan)
		close(stopChan)
		destroy()
	})
	mainButtons := container.NewGridWithColumns(3, lever, refresh, delete)
	mainWidget := container.NewVBox(mainButtons, leadSelectorContainer)

	/* Configuration Widget */
	configContainer := container.New(layout.NewFormLayout())
	for i := range workers {
		worker := &workers[i]
		movementModeSelectorLabel := widget.NewLabel("Game " + worker.GetHandle())
		movementModeSelectorLabel.TextStyle = fyne.TextStyle{Italic: true, Bold: true}
		movementModeSelector := widget.NewSelect(BATTLE_MOVEMENT_MODES, func(movementMode string) {
			worker.SetMovementMode(BattleMovementMode(movementMode))
		})
		movementModeSelector.PlaceHolder = "Choose movenent mode"
		configContainer.Add(movementModeSelectorLabel)
		configContainer.Add(movementModeSelector)
	}

	autoBattleWidget = container.NewVBox(mainWidget, configContainer)
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

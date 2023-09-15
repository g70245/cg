package main

import (
	"cg/game"
	. "cg/game"
	"image/color"
	"math"
	"strings"

	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
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

	newGroupButton := widget.NewButtonWithIcon("New Group", theme.ContentAddIcon(), func() {
		if len(autoGroups) == 3 {
			return
		}

		var newGroup map[string]HWND
		gamesChoosingCheckGroup := widget.NewCheckGroup(maps.Keys(idleGames), func(games []string) {
			newGroup = make(map[string]HWND)
			for _, game := range games {
				newGroup[game] = idleGames.Peek(game)
			}
		})
		gamesChoosingCheckGroup.Horizontal = true
		gamesChoosingDialog := dialog.NewCustom("Choose games", "Create", gamesChoosingCheckGroup, window)

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
					window.SetContent(window.Content())
					window.Resize(fyne.NewSize(APP_WIDTH, APP_HEIGHT))
				}
			}(id))
			autoGroups[id] = newGroup
			stopChans[id] = stopChan

			newTabItem = container.NewTabItem("Group "+fmt.Sprint(id), newGroupContainer)
			groupTabs.Append(newTabItem)
			groupTabs.Show()

			window.SetContent(window.Content())
			window.Resize(fyne.NewSize(APP_WIDTH, APP_HEIGHT))
			id++
		})
		gamesChoosingDialog.Show()
	})

	menu := container.NewVBox(newGroupButton)

	main := container.NewBorder(menu, nil, nil, nil, groupTabs)
	return main, stopChans
}

const (
	ON  = "On"
	OFF = "Off"
)

func newBatttleGroupContainer(games map[string]HWND, destroy func()) (autoBattleWidget *fyne.Container, stopChan chan bool) {
	var leadHandle string
	workers := CreateBattleWorkers(maps.Values(games))
	stopChan = make(chan bool, len(workers))

	/* Main Widget */
	leadSelector := widget.NewSelect(maps.Keys(games), func(handle string) {
		leadHandle = handle
	})
	leadSelector.PlaceHolder = "Lead"

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
	lever.Importance = widget.DangerImportance

	refresh := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() {
		leadSelector.ClearSelected()
	})
	refresh.Importance = widget.HighImportance

	delete := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		stop(stopChan)
		close(stopChan)
		destroy()
	})
	delete.Importance = widget.WarningImportance
	mainButtons := container.NewGridWithColumns(4, leadSelector, refresh, delete, lever)
	mainWidget := container.NewVBox(mainButtons)

	/* Configuration Widget */
	configContainer := container.NewVBox()
	for i := range workers {
		workerMenuContainer := container.NewGridWithColumns(4)
		worker := &workers[i]
		movementModeSelectorLabel := widget.NewLabel("Game " + worker.GetHandle())
		movementModeSelectorLabel.TextStyle = fyne.TextStyle{Italic: true, Bold: true}
		movementModeSelector := widget.NewSelect(BATTLE_MOVEMENT_MODES, func(movementMode string) {
			worker.MovementState.Mode = BattleMovementMode(movementMode)
		})
		movementModeSelector.PlaceHolder = "Move Way"

		var statesViewer *fyne.Container

		humanStateSelector := widget.NewButtonWithIcon("Man Action", theme.ContentAddIcon(), func() {
			worker.ActionState.HumanStates = worker.ActionState.HumanStates[:0]
			clear(worker.ActionState.HumanSkillIds)
			statesViewer.Objects = generateTags(*worker)
			statesViewer.Refresh()

			var attackButton *widget.Button
			var defenceButton *widget.Button
			var escapeButton *widget.Button
			var catchButton *widget.Button
			var bombButton *widget.Button
			var potionButton *widget.Button
			var recallButton *widget.Button
			var moveButton *widget.Button
			var hangButton *widget.Button
			var skillSelector *widget.Select
			var stealSelector *widget.Select
			var rideSelector *widget.Select
			var healSelector *widget.Select

			attackButton = widget.NewButton(game.H_A_ATTACK, func() {
				worker.ActionState.HumanStates = append(worker.ActionState.HumanStates, H_A_ATTACK)
				statesViewer.Objects = generateTags(*worker)
				statesViewer.Refresh()

				catchButton.Disable()
				bombButton.Disable()
				recallButton.Disable()
				rideSelector.Disable()
			})
			attackButton.Importance = widget.WarningImportance
			defenceButton = widget.NewButton(game.H_A_Defence, func() {
				worker.ActionState.HumanStates = append(worker.ActionState.HumanStates, H_A_Defence)
				statesViewer.Objects = generateTags(*worker)
				statesViewer.Refresh()

				bombButton.Disable()
				rideSelector.Disable()
				recallButton.Disable()
			})
			defenceButton.Importance = widget.WarningImportance
			escapeButton = widget.NewButton(game.H_A_ESCAPE, func() {
				worker.ActionState.HumanStates = append(worker.ActionState.HumanStates, H_A_ESCAPE)
				statesViewer.Objects = generateTags(*worker)
				statesViewer.Refresh()

				attackButton.Disable()
				escapeButton.Disable()
				bombButton.Disable()
				stealSelector.Disable()
				rideSelector.Disable()
				recallButton.Disable()
				moveButton.Disable()
				hangButton.Disable()
				skillSelector.Disable()
			})
			escapeButton.Importance = widget.WarningImportance
			catchButton = widget.NewButton(game.H_O_Catch, func() {
				worker.ActionState.HumanStates = append(worker.ActionState.HumanStates, H_O_Catch)
				statesViewer.Objects = generateTags(*worker)
				statesViewer.Refresh()
			})
			catchButton.Importance = widget.SuccessImportance
			bombButton = widget.NewButton(game.H_A_BOMB, func() {
				worker.ActionState.HumanStates = append(worker.ActionState.HumanStates, H_A_BOMB)
				statesViewer.Objects = generateTags(*worker)
				statesViewer.Refresh()

				catchButton.Disable()
				escapeButton.Disable()
				bombButton.Disable()
				recallButton.Disable()
				moveButton.Disable()
				hangButton.Disable()
				stealSelector.Disable()
				healSelector.Disable()
			})
			bombButton.Importance = widget.HighImportance
			potionButton = widget.NewButton(game.H_O_Potion, func() {
				worker.ActionState.HumanStates = append(worker.ActionState.HumanStates, H_O_Potion)
				statesViewer.Objects = generateTags(*worker)
				statesViewer.Refresh()
			})
			potionButton.Importance = widget.HighImportance
			recallButton = widget.NewButton(game.H_O_PET, func() {
				worker.ActionState.HumanStates = append(worker.ActionState.HumanStates, H_O_PET)
				statesViewer.Objects = generateTags(*worker)
				statesViewer.Refresh()

				recallButton.Disable()
				bombButton.Disable()
				rideSelector.Disable()
			})
			recallButton.Importance = widget.HighImportance
			moveButton = widget.NewButton(game.H_A_MOVE, func() {
				worker.ActionState.HumanStates = append(worker.ActionState.HumanStates, H_A_MOVE)
				statesViewer.Objects = generateTags(*worker)
				statesViewer.Refresh()

				bombButton.Disable()
				rideSelector.Disable()
				recallButton.Disable()
			})
			moveButton.Importance = widget.WarningImportance
			hangButton = widget.NewButton(game.H_A_HANG, func() {
				worker.ActionState.HumanStates = append(worker.ActionState.HumanStates, H_A_HANG)
				statesViewer.Objects = generateTags(*worker)
				statesViewer.Refresh()

				attackButton.Disable()
				defenceButton.Disable()
				escapeButton.Disable()
				catchButton.Disable()
				bombButton.Disable()
				skillSelector.Disable()
				stealSelector.Disable()
				rideSelector.Disable()
				recallButton.Disable()
				moveButton.Disable()
				potionButton.Disable()
				healSelector.Disable()
				hangButton.Disable()
			})
			hangButton.Importance = widget.WarningImportance
			skillSelector = widget.NewSelect([]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}, func(s string) {
				worker.ActionState.HumanStates = append(worker.ActionState.HumanStates, H_A_SKILL)
				worker.ActionState.HumanSkillIds[len(worker.ActionState.HumanStates)-1] = s
				statesViewer.Objects = generateTags(*worker)
				statesViewer.Refresh()

				bombButton.Disabled()
				catchButton.Disable()
				skillSelector.Disable()
				stealSelector.Disable()
				rideSelector.Disable()
				healSelector.Disable()
			})
			skillSelector.PlaceHolder = H_A_SKILL
			stealSelector = widget.NewSelect([]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}, func(s string) {
				worker.ActionState.HumanStates = append(worker.ActionState.HumanStates, H_A_STEAL)
				worker.ActionState.HumanSkillIds[len(worker.ActionState.HumanStates)-1] = s
				statesViewer.Objects = generateTags(*worker)
				statesViewer.Refresh()

				bombButton.Disable()
				skillSelector.Disable()
				stealSelector.Disable()
				rideSelector.Disable()
				healSelector.Disable()
			})
			stealSelector.PlaceHolder = H_A_STEAL
			rideSelector = widget.NewSelect([]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}, func(s string) {
				worker.ActionState.HumanStates = append(worker.ActionState.HumanStates, H_O_RIDE)
				worker.ActionState.HumanSkillIds[len(worker.ActionState.HumanStates)-1] = s
				statesViewer.Objects = generateTags(*worker)
				statesViewer.Refresh()

				catchButton.Disable()
				escapeButton.Disable()
				bombButton.Disable()
				recallButton.Disable()
				moveButton.Disable()
				hangButton.Disable()
				skillSelector.Disable()
				stealSelector.Disable()
				rideSelector.Disable()
				healSelector.Disable()
			})
			rideSelector.PlaceHolder = H_O_RIDE
			healSelector = widget.NewSelect([]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}, func(s string) {
				worker.ActionState.HumanStates = append(worker.ActionState.HumanStates, H_O_S_HEAL)
				worker.ActionState.HumanSkillIds[len(worker.ActionState.HumanStates)-1] = s
				statesViewer.Objects = generateTags(*worker)
				statesViewer.Refresh()

				bombButton.Disable()
				stealSelector.Disable()
			})
			healSelector.PlaceHolder = H_O_S_HEAL

			actionStatesContainer := container.NewGridWithColumns(4,
				attackButton,
				defenceButton,
				escapeButton,
				moveButton,
				hangButton,
				bombButton,
				recallButton,
				potionButton,
				catchButton,
				skillSelector,
				stealSelector,
				healSelector,
				rideSelector,
			)

			actionStatesDialog := dialog.NewCustom("Select man actions with orders", "Leave", actionStatesContainer, window)
			actionStatesDialog.Show()
		})

		petStateSelector := widget.NewButtonWithIcon("Pet Action", theme.ContentAddIcon(), func() {
			worker.ActionState.PetStates = worker.ActionState.PetStates[:0]
			clear(worker.ActionState.PetSkillIds)
			statesViewer.Objects = generateTags(*worker)
			statesViewer.Refresh()

			var petAttackButton *widget.Button
			var petHangButton *widget.Button
			var petSkillSelector *widget.Select
			var petDefenceSelector *widget.Select
			var petRideSelector *widget.Select

			petAttackButton = widget.NewButton(game.P_ATTACK, func() {
				worker.ActionState.PetStates = append(worker.ActionState.PetStates, P_ATTACK)
				statesViewer.Objects = generateTags(*worker)
				statesViewer.Refresh()

				petAttackButton.Disable()
				petHangButton.Disable()
				petRideSelector.Disable()
			})
			petAttackButton.Importance = widget.WarningImportance
			petHangButton = widget.NewButton(game.P_HANG, func() {
				worker.ActionState.PetStates = append(worker.ActionState.PetStates, P_HANG)
				statesViewer.Objects = generateTags(*worker)
				statesViewer.Refresh()

				petAttackButton.Disable()
				petHangButton.Disable()
				petSkillSelector.Disable()
				petRideSelector.Disable()
			})
			petHangButton.Importance = widget.WarningImportance
			petSkillSelector = widget.NewSelect([]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}, func(s string) {
				worker.ActionState.PetStates = append(worker.ActionState.PetStates, P_SkILL)
				worker.ActionState.PetSkillIds[len(worker.ActionState.PetStates)-1] = s
				statesViewer.Objects = generateTags(*worker)
				statesViewer.Refresh()

				petHangButton.Disable()
				petRideSelector.Disable()
			})
			petSkillSelector.PlaceHolder = P_SkILL
			petDefenceSelector = widget.NewSelect([]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}, func(s string) {
				worker.ActionState.PetStates = append(worker.ActionState.PetStates, P_DEFENCE)
				worker.ActionState.PetSkillIds[len(worker.ActionState.PetStates)-1] = s
				statesViewer.Objects = generateTags(*worker)
				statesViewer.Refresh()

				petDefenceSelector.Disable()
				petHangButton.Disable()
				petRideSelector.Disable()
			})
			petDefenceSelector.PlaceHolder = P_DEFENCE
			petRideSelector = widget.NewSelect([]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}, func(s string) {
				worker.ActionState.PetStates = append(worker.ActionState.PetStates, P_RIDE)
				worker.ActionState.PetSkillIds[len(worker.ActionState.PetStates)-1] = s
				statesViewer.Objects = generateTags(*worker)
				statesViewer.Refresh()

				petHangButton.Disable()
				petRideSelector.Disable()
			})
			petRideSelector.PlaceHolder = P_RIDE

			actionStatesContainer := container.NewGridWithColumns(4,
				petAttackButton,
				petHangButton,
				petSkillSelector,
				petDefenceSelector,
				petRideSelector,
			)

			actionStatesDialog := dialog.NewCustom("Select pet actions with orders", "Leave", actionStatesContainer, window)
			actionStatesDialog.Show()

		})

		// statesViewer = container.NewGridWrap(fyne.NewSize(82, 20), generateTags(*worker)...)
		statesViewer = container.NewHBox(generateTags(*worker)...)

		humanStateSelector.Importance = widget.MediumImportance
		petStateSelector.Importance = widget.MediumImportance
		workerMenuContainer.Add(movementModeSelectorLabel)
		workerMenuContainer.Add(movementModeSelector)
		workerMenuContainer.Add(humanStateSelector)
		workerMenuContainer.Add(petStateSelector)
		workerContainer := container.NewVBox(workerMenuContainer, statesViewer)
		configContainer.Add(workerContainer)
	}

	autoBattleWidget = container.NewVBox(widget.NewSeparator(), mainWidget, widget.NewSeparator(), configContainer)
	return autoBattleWidget, stopChan
}

func generateTags(worker BattleWorker) (tagContaines []fyne.CanvasObject) {
	tagContaines = createTagContainer(worker.ActionState.HumanStates, worker.ActionState.HumanSkillIds, true)
	tagContaines = append(tagContaines, createTagContainer(worker.ActionState.PetStates, worker.ActionState.PetSkillIds, false)...)
	return
}

var (
	humanOnceTagColor       = color.RGBA{245, 79, 0, uint8(math.Round(0.8 * 255))}
	humanOnceOptionTagColor = color.RGBA{0, 28, 245, uint8(math.Round(0.8 * 255))}
	humanOptionTagColor     = color.RGBA{35, 128, 24, uint8(math.Round(0.8 * 255))}
	petTagColor             = color.RGBA{92, 14, 99, uint8(math.Round(0.8 * 255))}
)

func createTagContainer(tags []string, skillIds map[int]string, isHuman bool) (tagContainers []fyne.CanvasObject) {
	for i, tag := range tags {
		tagColor := petTagColor
		if isHuman {
			switch {
			case strings.Contains(tag, "**"):
				tagColor = humanOnceTagColor
			case strings.Contains(tag, "*"):
				tagColor = humanOnceOptionTagColor
			default:
				tagColor = humanOptionTagColor
			}
		}

		tagCanvas := canvas.NewRectangle(tagColor)
		tagCanvas.SetMinSize(fyne.NewSize(60, 22))
		tagContainer := container.NewStack(tagCanvas)
		if value, ok := skillIds[i]; ok {
			tag += ":" + value
		}
		tagTextCanvas := canvas.NewText("\t"+tag+"\t\t", color.White)
		tagTextCanvas.Alignment = fyne.TextAlignCenter
		tagTextCanvas.TextStyle = fyne.TextStyle{Bold: true, Italic: true, TabWidth: 1}
		tagContainer.Add(tagTextCanvas)
		tagContainers = append(tagContainers, tagContainer)
	}
	return
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

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

	var leadSelectorDialog *dialog.CustomDialog
	var leadSelectorButton *widget.Button
	leadSelector := widget.NewRadioGroup(maps.Keys(games), func(s string) {
		leadHandle = s
		if leadHandle != "" {
			leadSelectorButton.SetText("Lead " + leadHandle)
		} else {
			leadSelectorButton.SetText("Choose Lead")
		}
		leadSelectorDialog.Hide()
	})
	leadSelectorDialog = dialog.NewCustomWithoutButtons("Choose a lead with this group", leadSelector, window)
	leadSelectorButton = widget.NewButton("Choose Lead", func() {
		leadSelectorDialog.Show()
	})
	leadSelectorButton.Importance = widget.HighImportance

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
		leadHandle = ""
		leadSelectorButton.SetText("Choose Lead")
	})
	refresh.Importance = widget.HighImportance

	delete := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		stop(stopChan)
		close(stopChan)
		destroy()
	})
	delete.Importance = widget.WarningImportance
	mainButtons := container.NewGridWithColumns(4, leadSelectorButton, refresh, delete, lever)
	mainWidget := container.NewVBox(mainButtons)

	/* Configuration Widget */
	configContainer := container.NewVBox()
	for i := range workers {
		workerMenuContainer := container.NewGridWithColumns(4)
		worker := &workers[i]
		gameNameLabel := widget.NewLabel("Game " + worker.GetHandle())
		gameNameLabel.TextStyle = fyne.TextStyle{Italic: true, Bold: true}

		var movementModeDialog *dialog.CustomDialog
		movementModeSelector := widget.NewRadioGroup(BATTLE_MOVEMENT_MODES, func(s string) {
			worker.MovementState.Mode = BattleMovementMode(s)
			movementModeDialog.Hide()
		})
		movementModeSelector.Required = true
		movementModeDialog = dialog.NewCustomWithoutButtons("Choose a movement mode", movementModeSelector, window)
		movementModeButton := widget.NewButtonWithIcon("Choose Move Way", theme.MailReplyIcon(), func() {
			movementModeDialog.Show()
		})

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
			var skillButton *widget.Button
			var stealButton *widget.Button
			var rideButton *widget.Button
			var healButton *widget.Button

			var level string
			var levelSelectorDialog *dialog.CustomDialog
			levelSelector := widget.NewRadioGroup([]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}, func(s string) {
				level = s
				levelSelectorDialog.Hide()
			})
			levelSelector.Horizontal = true
			levelSelectorDialog = dialog.NewCustomWithoutButtons("Choose skill level", levelSelector, window)
			levelSelectorDialog.SetOnClosed(func() {
				worker.ActionState.HumanSkillLevels[len(worker.ActionState.HumanStates)-1] = level
				statesViewer.Objects = generateTags(*worker)
				statesViewer.Refresh()
			})

			var id string
			var idSelectorDialog *dialog.CustomDialog
			idSelector := widget.NewRadioGroup([]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}, func(s string) {
				id = s
				idSelectorDialog.Hide()
			})
			idSelector.Horizontal = true
			idSelectorDialog = dialog.NewCustomWithoutButtons("Choose skill index", idSelector, window)
			idSelectorDialog.SetOnClosed(func() {
				worker.ActionState.HumanSkillIds[len(worker.ActionState.HumanStates)-1] = id
				levelSelectorDialog.Show()
			})

			attackButton = widget.NewButton(game.H_A_ATTACK, func() {
				worker.ActionState.HumanStates = append(worker.ActionState.HumanStates, H_A_ATTACK)
				statesViewer.Objects = generateTags(*worker)
				statesViewer.Refresh()

				catchButton.Disable()
				bombButton.Disable()
				recallButton.Disable()
				rideButton.Disable()
			})
			attackButton.Importance = widget.WarningImportance
			defenceButton = widget.NewButton(game.H_A_Defence, func() {
				worker.ActionState.HumanStates = append(worker.ActionState.HumanStates, H_A_Defence)
				statesViewer.Objects = generateTags(*worker)
				statesViewer.Refresh()

				bombButton.Disable()
				rideButton.Disable()
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
				stealButton.Disable()
				rideButton.Disable()
				recallButton.Disable()
				moveButton.Disable()
				hangButton.Disable()
				skillButton.Disable()
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
				stealButton.Disable()
				healButton.Disable()
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
				rideButton.Disable()
			})
			recallButton.Importance = widget.HighImportance
			moveButton = widget.NewButton(game.H_A_MOVE, func() {
				worker.ActionState.HumanStates = append(worker.ActionState.HumanStates, H_A_MOVE)
				statesViewer.Objects = generateTags(*worker)
				statesViewer.Refresh()

				bombButton.Disable()
				rideButton.Disable()
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
				skillButton.Disable()
				stealButton.Disable()
				rideButton.Disable()
				recallButton.Disable()
				moveButton.Disable()
				potionButton.Disable()
				healButton.Disable()
				hangButton.Disable()
			})
			hangButton.Importance = widget.WarningImportance
			skillButton = widget.NewButton(H_A_SKILL, func() {
				worker.ActionState.HumanStates = append(worker.ActionState.HumanStates, H_A_SKILL)
				idSelectorDialog.Show()

				bombButton.Disabled()
				catchButton.Disable()
				skillButton.Disable()
				stealButton.Disable()
				rideButton.Disable()
				healButton.Disable()
			})
			skillButton.Importance = widget.HighImportance
			stealButton = widget.NewButton(H_A_STEAL, func() {
				worker.ActionState.HumanStates = append(worker.ActionState.HumanStates, H_A_STEAL)
				idSelectorDialog.Show()

				bombButton.Disable()
				skillButton.Disable()
				stealButton.Disable()
				rideButton.Disable()
				healButton.Disable()
			})
			stealButton.Importance = widget.HighImportance
			rideButton = widget.NewButton(H_O_RIDE, func() {
				worker.ActionState.HumanStates = append(worker.ActionState.HumanStates, H_O_RIDE)
				idSelectorDialog.Show()

				catchButton.Disable()
				escapeButton.Disable()
				bombButton.Disable()
				recallButton.Disable()
				moveButton.Disable()
				hangButton.Disable()
				skillButton.Disable()
				stealButton.Disable()
				rideButton.Disable()
				healButton.Disable()
			})
			rideButton.Importance = widget.HighImportance
			healButton = widget.NewButton(H_O_S_HEAL, func() {
				worker.ActionState.HumanStates = append(worker.ActionState.HumanStates, H_O_S_HEAL)
				idSelectorDialog.Show()

				bombButton.Disable()
				stealButton.Disable()
			})
			healButton.Importance = widget.HighImportance

			actionStatesContainer := container.NewGridWithColumns(4,
				attackButton,
				defenceButton,
				escapeButton,
				moveButton,
				hangButton,
				bombButton,
				recallButton,
				potionButton,
				skillButton,
				stealButton,
				rideButton,
				healButton,
				catchButton,
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
			var petSkillButton *widget.Button
			var petDefenceButton *widget.Button
			var petRideButton *widget.Button

			var id string
			var idSelectorDialog *dialog.CustomDialog
			idSelector := widget.NewRadioGroup([]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}, func(s string) {
				id = s
				idSelectorDialog.Hide()
			})
			idSelector.Horizontal = true
			idSelectorDialog = dialog.NewCustomWithoutButtons("Choose skill index", idSelector, window)
			idSelectorDialog.SetOnClosed(func() {
				worker.ActionState.PetSkillIds[len(worker.ActionState.PetStates)-1] = id
				statesViewer.Objects = generateTags(*worker)
				statesViewer.Refresh()
			})

			petAttackButton = widget.NewButton(game.P_ATTACK, func() {
				worker.ActionState.PetStates = append(worker.ActionState.PetStates, P_ATTACK)
				statesViewer.Objects = generateTags(*worker)
				statesViewer.Refresh()

				petAttackButton.Disable()
				petHangButton.Disable()
				petRideButton.Disable()
			})
			petAttackButton.Importance = widget.WarningImportance
			petHangButton = widget.NewButton(game.P_HANG, func() {
				worker.ActionState.PetStates = append(worker.ActionState.PetStates, P_HANG)
				statesViewer.Objects = generateTags(*worker)
				statesViewer.Refresh()

				petAttackButton.Disable()
				petHangButton.Disable()
				petSkillButton.Disable()
				petRideButton.Disable()
			})
			petHangButton.Importance = widget.HighImportance
			petHangButton.Importance = widget.WarningImportance
			petSkillButton = widget.NewButton(P_SkILL, func() {
				worker.ActionState.PetStates = append(worker.ActionState.PetStates, P_SkILL)
				idSelectorDialog.Show()

				petHangButton.Disable()
				petRideButton.Disable()
			})
			petSkillButton.Importance = widget.HighImportance
			petDefenceButton = widget.NewButton(P_DEFENCE, func() {
				worker.ActionState.PetStates = append(worker.ActionState.PetStates, P_DEFENCE)
				idSelectorDialog.Show()

				petDefenceButton.Disable()
				petHangButton.Disable()
				petRideButton.Disable()
			})
			petDefenceButton.Importance = widget.HighImportance
			petRideButton = widget.NewButton(P_RIDE, func() {
				worker.ActionState.PetStates = append(worker.ActionState.PetStates, P_RIDE)
				idSelectorDialog.Show()

				petHangButton.Disable()
				petRideButton.Disable()
			})
			petRideButton.Importance = widget.HighImportance

			actionStatesContainer := container.NewGridWithColumns(4,
				petAttackButton,
				petHangButton,
				petSkillButton,
				petDefenceButton,
				petRideButton,
			)

			actionStatesDialog := dialog.NewCustom("Select pet actions with orders", "Leave", actionStatesContainer, window)
			actionStatesDialog.Show()

		})

		// statesViewer = container.NewGridWrap(fyne.NewSize(82, 20), generateTags(*worker)...)
		statesViewer = container.NewHBox(generateTags(*worker)...)

		humanStateSelector.Importance = widget.MediumImportance
		petStateSelector.Importance = widget.MediumImportance
		workerMenuContainer.Add(gameNameLabel)
		workerMenuContainer.Add(movementModeButton)
		workerMenuContainer.Add(humanStateSelector)
		workerMenuContainer.Add(petStateSelector)
		workerContainer := container.NewVBox(workerMenuContainer, statesViewer)
		configContainer.Add(workerContainer)
	}

	autoBattleWidget = container.NewVBox(widget.NewSeparator(), mainWidget, widget.NewSeparator(), configContainer)
	return autoBattleWidget, stopChan
}

func generateTags(worker BattleWorker) (tagContaines []fyne.CanvasObject) {
	tagContaines = createTagContainer(worker.ActionState, true)
	tagContaines = append(tagContaines, createTagContainer(worker.ActionState, false)...)
	return
}

var (
	humanOnceTagColor       = color.RGBA{245, 79, 0, uint8(math.Round(0.8 * 255))}
	humanOnceOptionTagColor = color.RGBA{0, 28, 245, uint8(math.Round(0.8 * 255))}
	humanOptionTagColor     = color.RGBA{35, 128, 24, uint8(math.Round(0.8 * 255))}
	petTagColor             = color.RGBA{92, 14, 99, uint8(math.Round(0.8 * 255))}
)

func createTagContainer(actionState BattleActionState, isHuman bool) (tagContainers []fyne.CanvasObject) {
	var tags []string
	if isHuman {
		tags = actionState.HumanStates
	} else {
		tags = actionState.PetStates
	}

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
		if isHuman {
			if v, ok := actionState.HumanSkillIds[i]; ok {
				tag = fmt.Sprintf("%s:%s:%s", tag, v, actionState.HumanSkillLevels[i])
			}
		} else {
			if v, ok := actionState.PetSkillIds[i]; ok {
				tag = fmt.Sprintf("%s:%s", tag, v)
			}
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

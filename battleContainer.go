package main

import (
	. "cg/game"
	. "cg/system"
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
		// if len(autoGroups) == 3 {
		// 	return
		// }

		var newGroup map[string]HWND
		gamesChoosingCheckGroup := widget.NewCheckGroup(maps.Keys(idleGames), func(games []string) {
			newGroup = make(map[string]HWND)
			for _, game := range games {
				newGroup[game] = idleGames.Peek(game)
			}
		})
		gamesChoosingCheckGroup.Horizontal = true
		gamesChoosingDialog := dialog.NewCustom("Choose games", "Create", gamesChoosingCheckGroup, window)
		gamesChoosingDialog.Resize(fyne.NewSize(240, 166))

		gamesChoosingDialog.SetOnClosed(func() {
			if len(newGroup) == 0 {
				return
			}

			// idleGames.Remove(maps.Keys(newGroup))
			var newTabItem *container.TabItem

			newGroupContainer, stopChan := newBatttleGroupContainer(newGroup, func(id int) func() {
				return func() {
					delete(autoGroups, id)
					delete(stopChans, id)

					// idleGames.Add(newGroup)

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

func newBatttleGroupContainer(games map[string]HWND, destroy func()) (autoBattleWidget *fyne.Container, stopChan chan bool) {
	var leadHandle string
	workers := CreateBattleWorkers(maps.Values(games), logDir)
	stopChan = make(chan bool, len(workers))

	var leadSelectorDialog *dialog.CustomDialog
	var leadSelectorButton *widget.Button
	leadSelector := widget.NewRadioGroup(maps.Keys(games), func(s string) {
		leadHandle = s
		if leadHandle != "" {
			leadSelectorButton.SetText("Lead: " + leadHandle)
		} else {
			leadSelectorButton.SetText("Choose Lead")
		}
		leadSelectorDialog.Hide()
	})
	leadSelectorDialog = dialog.NewCustomWithoutButtons("Select a lead with this group", leadSelector, window)
	leadSelectorButton = widget.NewButton("Select Lead", func() {
		leadSelectorDialog.Show()
	})
	leadSelectorButton.Importance = widget.HighImportance

	var lever *widget.Button
	lever = widget.NewButtonWithIcon("", theme.MediaPlayIcon(), func() {
		switch lever.Icon {
		case theme.MediaPlayIcon():
			for i := range workers {
				workers[i].Work(&leadHandle, stopChan)
			}
			turn(theme.MediaStopIcon(), lever)
		case theme.MediaStopIcon():
			for i := range workers {
				workers[i].ActionState.Enabled = false
				stopChan <- true
			}
			StopBeeper()
			turn(theme.MediaPlayIcon(), lever)
		}
	})
	lever.Importance = widget.WarningImportance

	delete := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		stop(stopChan)
		close(stopChan)
		destroy()
	})
	delete.Importance = widget.DangerImportance

	var inventoryCheck *widget.Button
	inventoryCheck = widget.NewButtonWithIcon("Check Inventory", theme.CheckButtonIcon(), func() {
		switch inventoryCheck.Icon {
		case theme.CheckButtonCheckedIcon():
			for i := range workers {
				workers[i].StopInventoryChecker()
			}
			turn(theme.CheckButtonIcon(), inventoryCheck)
		case theme.CheckButtonIcon():
			for i := range workers {
				workers[i].StartInventoryChecker()
			}
			turn(theme.CheckButtonCheckedIcon(), inventoryCheck)
		}
	})
	inventoryCheck.Importance = widget.HighImportance

	mainButtons := container.NewGridWithColumns(4, leadSelectorButton, inventoryCheck, delete, lever)
	mainWidget := container.NewVBox(mainButtons)

	/* Configuration Widget */
	configContainer := container.NewVBox()
	for i := range workers {
		workerMenuContainer := container.NewGridWithColumns(4)
		worker := &workers[i]

		gameNameLabelContainer := container.NewStack()
		gameNameLabel := canvas.NewText(worker.GetHandle(), color.RGBA{11, 86, 107, 255})
		gameNameLabel.Alignment = fyne.TextAlignCenter
		gameNameLabel.TextStyle = fyne.TextStyle{TabWidth: 1, Bold: true}
		gameNameLabel.TextSize = 16
		gameNameLabelContainer.Add(gameNameLabel)

		var movementModeButton *widget.Button
		var movementModeDialog *dialog.CustomDialog
		movementModeSelector := widget.NewRadioGroup(BATTLE_MOVEMENT_MODES, func(s string) {
			if s != "" {
				worker.MovementState.Mode = BattleMovementMode(s)
				movementModeButton.SetText(s)
			} else {
				worker.MovementState.Mode = BattleMovementMode(NONE)
				movementModeButton.SetText("Move Way")
			}
			movementModeDialog.Hide()
		})
		movementModeDialog = dialog.NewCustomWithoutButtons("Choose a Move Way", movementModeSelector, window)
		movementModeButton = widget.NewButtonWithIcon("Move Way", theme.MailReplyIcon(), func() {
			movementModeDialog.Show()
		})

		var statesViewer *fyne.Container
		var selector = widget.NewRadioGroup(nil, nil)
		selector.Horizontal = true
		selector.Required = true
		enableChan := make(chan bool)

		onClosed := func() {
			enableChan <- true
		}

		/* Control Unit Dialogs */
		var hCUSuccessDialog *dialog.CustomDialog
		var hCUFailureDialog *dialog.CustomDialog
		var pCUSuccessDialog *dialog.CustomDialog
		var pCUFailureDialog *dialog.CustomDialog
		hCUSuccesOnChanged := func(s string) {
			if s != "" {
				worker.ActionState.AddHumanSuccessControlUnit(s)
				hCUSuccessDialog.Hide()
			}
		}
		hCUFailureOnChanged := func(s string) {
			if s != "" {
				worker.ActionState.AddHumanFailureControlUnit(s)
				hCUFailureDialog.Hide()
			}
		}
		pCUSuccessOnChanged := func(s string) {
			if s != "" {
				worker.ActionState.AddPetSuccessControlUnit(s)
				pCUSuccessDialog.Hide()
			}
		}
		pCUFailureOnChanged := func(s string) {
			if s != "" {
				worker.ActionState.AddPetFailureControlUnit(s)
				pCUFailureDialog.Hide()
			}
		}
		cuOnClosed := func() {
			statesViewer.Objects = generateTags(*worker)
			statesViewer.Refresh()
			enableChan <- true
		}
		hCUSuccessDialog = dialog.NewCustomWithoutButtons("Select next move after successful execution", selector, window)
		hCUSuccessDialog.SetOnClosed(cuOnClosed)
		hCUFailureDialog = dialog.NewCustomWithoutButtons("Select next move after failed execution", selector, window)
		hCUFailureDialog.SetOnClosed(cuOnClosed)
		pCUSuccessDialog = dialog.NewCustomWithoutButtons("Select next move after successful execution", selector, window)
		pCUSuccessDialog.SetOnClosed(cuOnClosed)
		pCUFailureDialog = dialog.NewCustomWithoutButtons("Select next move after failed execution", selector, window)
		pCUFailureDialog.SetOnClosed(cuOnClosed)

		hCUSuccessSelectorDialog := SelectorDialog{
			hCUSuccessDialog,
			selector,
			ControlUnitOptions,
			hCUSuccesOnChanged,
		}
		hCUFailureSelectorDialog := SelectorDialog{
			hCUFailureDialog,
			selector,
			ControlUnitOptions,
			hCUFailureOnChanged,
		}
		pCUSuccessSelectorDialog := SelectorDialog{
			pCUSuccessDialog,
			selector,
			ControlUnitOptions,
			pCUSuccessOnChanged,
		}
		pCUFailureSelectorDialog := SelectorDialog{
			pCUFailureDialog,
			selector,
			ControlUnitOptions,
			pCUFailureOnChanged,
		}

		/* Param Dialog */
		var paramsDialog *dialog.CustomDialog
		humanParamsOnChanged := func(s string) {
			if s != "" {
				worker.ActionState.AddHumanParam(s)
				paramsDialog.Hide()
			}
		}
		petParamsOnChanged := func(s string) {
			if s != "" {
				worker.ActionState.AddPetParam(s)
				paramsDialog.Hide()
			}
		}
		paramsSelector := widget.NewRadioGroup(nil, nil)
		paramsSelector.Horizontal = true
		paramsSelector.Required = true
		paramsDialog = dialog.NewCustomWithoutButtons("Choose param", selector, window)
		paramsDialog.SetOnClosed(onClosed)

		hHealingRatioSelectorDialog := SelectorDialog{
			paramsDialog,
			selector,
			HealingOptions,
			humanParamsOnChanged,
		}
		pHealingRatioSelectorDialog := SelectorDialog{
			paramsDialog,
			selector,
			HealingOptions,
			petParamsOnChanged,
		}
		bombsSelectorDialog := SelectorDialog{
			paramsDialog,
			selector,
			Bombs.GetOptions(),
			humanParamsOnChanged,
		}

		/* Id Dialog */
		var idDialog *dialog.CustomDialog
		hIdOnChanged := func(s string) {
			if s != "" {
				worker.ActionState.AddHumanSkillId(s)
				idDialog.Hide()
			}
		}
		pIdOnChanged := func(s string) {
			if s != "" {
				worker.ActionState.AddPetSkillId(s)
				idDialog.Hide()
			}
		}
		idDialog = dialog.NewCustomWithoutButtons("Choose skill id", selector, window)
		idDialog.SetOnClosed(onClosed)
		hIdSelectorDialog := SelectorDialog{
			idDialog,
			selector,
			IdOptions,
			hIdOnChanged,
		}
		pIdSelectorDialog := SelectorDialog{
			idDialog,
			selector,
			IdOptions,
			pIdOnChanged,
		}

		/* Level Dialog */
		var levelDialog *dialog.CustomDialog
		levelOnChanged := func(s string) {
			if s != "" {
				worker.ActionState.AddHumanSkillLevel(s)
				levelDialog.Hide()
			}
		}
		levelDialog = dialog.NewCustomWithoutButtons("Choose skill level", selector, window)
		levelDialog.SetOnClosed(onClosed)

		levelSelectorDialog := SelectorDialog{
			levelDialog,
			selector,
			LevelOptions,
			levelOnChanged,
		}

		humanStateSelector := widget.NewButtonWithIcon("Man Action", theme.ContentAddIcon(), func() {
			worker.ActionState.ClearHumanStates()
			statesViewer.Objects = generateTags(*worker)
			statesViewer.Refresh()

			var attackButton *widget.Button
			var defendButton *widget.Button
			var escapeButton *widget.Button
			var catchButton *widget.Button
			var bombButton *widget.Button
			var potionButton *widget.Button
			var recallButton *widget.Button
			var moveButton *widget.Button
			var hangButton *widget.Button
			var skillButton *widget.Button
			var stealButton *widget.Button
			var trainButton *widget.Button
			var rideButton *widget.Button
			var healSelfButton *widget.Button
			var healOneButton *widget.Button
			var healMultiButton *widget.Button

			attackButton = widget.NewButton(H_A_ATTACK, func() {
				worker.ActionState.AddHumanState(H_A_ATTACK)

				dialogs := []SelectorDialog{
					hCUSuccessSelectorDialog,
					hCUFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)

				trainButton.Disable()
				stealButton.Disable()
			})
			attackButton.Importance = widget.WarningImportance

			defendButton = widget.NewButton(H_A_DEFEND, func() {
				worker.ActionState.AddHumanState(H_A_DEFEND)

				dialogs := []SelectorDialog{
					hCUSuccessSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			defendButton.Importance = widget.WarningImportance

			escapeButton = widget.NewButton(H_A_ESCAPE, func() {
				worker.ActionState.AddHumanState(H_A_ESCAPE)

				dialogs := []SelectorDialog{
					hCUFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)

				bombButton.Disable()
				stealButton.Disable()
				rideButton.Disable()
				recallButton.Disable()
				trainButton.Disable()
			})
			escapeButton.Importance = widget.WarningImportance

			catchButton = widget.NewButton(H_S_CATCH, func() {
				worker.ActionState.AddHumanState(H_S_CATCH)

				dialogs := []SelectorDialog{
					hCUSuccessSelectorDialog,
					hCUFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)

				trainButton.Disable()
			})
			catchButton.Importance = widget.SuccessImportance

			bombButton = widget.NewButton(H_O_BOMB, func() {
				worker.ActionState.AddHumanState(H_O_BOMB)

				dialogs := []SelectorDialog{
					bombsSelectorDialog,
					hCUSuccessSelectorDialog,
					hCUFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)

				catchButton.Disable()
				escapeButton.Disable()
				recallButton.Disable()
				hangButton.Disable()
				stealButton.Disable()
				healOneButton.Disable()
				trainButton.Disable()
				healMultiButton.Disable()
				healSelfButton.Disable()
			})
			bombButton.Importance = widget.HighImportance

			potionButton = widget.NewButton(H_O_POTION, func() {
				worker.ActionState.AddHumanState(H_O_POTION)

				dialogs := []SelectorDialog{
					hHealingRatioSelectorDialog,
					hCUSuccessSelectorDialog,
					hCUFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			potionButton.Importance = widget.HighImportance

			recallButton = widget.NewButton(H_O_PET_RECALL, func() {
				worker.ActionState.AddHumanState(H_O_PET_RECALL)

				dialogs := []SelectorDialog{
					hCUSuccessSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)

				rideButton.Disable()
			})
			recallButton.Importance = widget.HighImportance

			moveButton = widget.NewButton(H_A_MOVE, func() {
				worker.ActionState.AddHumanState(H_A_MOVE)

				dialogs := []SelectorDialog{
					hCUSuccessSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			moveButton.Importance = widget.WarningImportance

			hangButton = widget.NewButton(H_S_HANG, func() {
				worker.ActionState.AddHumanState(H_S_HANG)
				statesViewer.Objects = generateTags(*worker)
				statesViewer.Refresh()

				catchButton.Disable()
				bombButton.Disable()
				stealButton.Disable()
				rideButton.Disable()
				recallButton.Disable()
				potionButton.Disable()
				healOneButton.Disable()
				hangButton.Disable()
				healMultiButton.Disable()
				healSelfButton.Disable()
				trainButton.Disable()
			})
			hangButton.Importance = widget.SuccessImportance

			skillButton = widget.NewButton(H_O_SKILL, func() {
				worker.ActionState.AddHumanState(H_O_SKILL)

				dialogs := []SelectorDialog{
					hIdSelectorDialog,
					levelSelectorDialog,
					hCUSuccessSelectorDialog,
					hCUFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)

				bombButton.Disabled()
				catchButton.Disable()
				stealButton.Disable()
				trainButton.Disable()
			})
			skillButton.Importance = widget.HighImportance

			stealButton = widget.NewButton(H_S_STEAL, func() {
				worker.ActionState.AddHumanState(H_S_STEAL)

				dialogs := []SelectorDialog{
					hIdSelectorDialog,
					levelSelectorDialog,
					hCUSuccessSelectorDialog,
					hCUFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)

				bombButton.Disable()
				healOneButton.Disable()
				healMultiButton.Disable()
				healSelfButton.Disable()
				trainButton.Disable()
				catchButton.Disable()
			})
			stealButton.Importance = widget.SuccessImportance

			trainButton = widget.NewButton(H_S_TRAIN_SKILL, func() {
				worker.ActionState.AddHumanState(H_S_TRAIN_SKILL)

				dialogs := []SelectorDialog{
					hIdSelectorDialog,
					levelSelectorDialog,
					hCUSuccessSelectorDialog,
					hCUFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)

				bombButton.Disable()
				stealButton.Disable()
				rideButton.Disable()
				healOneButton.Disable()
				healMultiButton.Disable()
				healSelfButton.Disable()
				catchButton.Disable()
				hangButton.Disable()
				recallButton.Disable()

			})
			trainButton.Importance = widget.SuccessImportance

			rideButton = widget.NewButton(H_O_RIDE, func() {
				worker.ActionState.AddHumanState(H_O_RIDE)

				dialogs := []SelectorDialog{
					hIdSelectorDialog,
					levelSelectorDialog,
					hCUSuccessSelectorDialog,
					hCUFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)

				catchButton.Disable()
				escapeButton.Disable()
				bombButton.Disable()
				recallButton.Disable()
				moveButton.Disable()
				hangButton.Disable()
				skillButton.Disable()
				stealButton.Disable()
				rideButton.Disable()
			})
			rideButton.Importance = widget.HighImportance

			healSelfButton = widget.NewButton(H_O_SE_HEAL, func() {
				worker.ActionState.AddHumanState(H_O_SE_HEAL)

				dialogs := []SelectorDialog{
					hIdSelectorDialog,
					levelSelectorDialog,
					hHealingRatioSelectorDialog,
					hCUSuccessSelectorDialog,
					hCUFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)

				trainButton.Disable()
				bombButton.Disable()
				stealButton.Disable()
				healOneButton.Disable()
				healMultiButton.Disable()
			})
			healSelfButton.Importance = widget.HighImportance

			healOneButton = widget.NewButton(H_O_O_HEAL, func() {
				worker.ActionState.AddHumanState(H_O_O_HEAL)

				dialogs := []SelectorDialog{
					hIdSelectorDialog,
					levelSelectorDialog,
					hHealingRatioSelectorDialog,
					hCUSuccessSelectorDialog,
					hCUFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)

				healSelfButton.Disable()
				bombButton.Disable()
				stealButton.Disable()
			})
			healOneButton.Importance = widget.HighImportance

			healMultiButton = widget.NewButton(H_O_M_HEAL, func() {
				worker.ActionState.AddHumanState(H_O_M_HEAL)

				dialogs := []SelectorDialog{
					hIdSelectorDialog,
					levelSelectorDialog,
					hHealingRatioSelectorDialog,
					hCUSuccessSelectorDialog,
					hCUFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)

				trainButton.Disable()
				healSelfButton.Disable()
				bombButton.Disable()
				stealButton.Disable()
			})
			healMultiButton.Importance = widget.HighImportance

			actionStatesContainer := container.NewGridWithColumns(4,
				attackButton,
				defendButton,
				escapeButton,
				moveButton,
				skillButton,
				bombButton,
				recallButton,
				rideButton,
				potionButton,
				healSelfButton,
				healOneButton,
				healMultiButton,
				hangButton,
				catchButton,
				stealButton,
				trainButton,
			)

			actionStatesDialog := dialog.NewCustom("Select man actions with order", "Leave", actionStatesContainer, window)
			actionStatesDialog.Show()
		})

		petStateSelector := widget.NewButtonWithIcon("Pet Action", theme.ContentAddIcon(), func() {
			worker.ActionState.ClearPetStates()
			statesViewer.Objects = generateTags(*worker)
			statesViewer.Refresh()

			var petAttackButton *widget.Button
			var petHangButton *widget.Button
			var petSkillButton *widget.Button
			var petDefenceButton *widget.Button
			var petEscapeButton *widget.Button
			var petHealSelfButton *widget.Button
			var petHealOneButton *widget.Button
			var petRideButton *widget.Button
			var petOffRideButton *widget.Button

			petAttackButton = widget.NewButton(P_ATTACK, func() {
				worker.ActionState.AddPetState(P_ATTACK)

				dialogs := []SelectorDialog{
					pCUSuccessSelectorDialog,
					pCUFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			petAttackButton.Importance = widget.WarningImportance

			petHangButton = widget.NewButton(P_HANG, func() {
				worker.ActionState.AddPetState(P_HANG)
				statesViewer.Objects = generateTags(*worker)
				statesViewer.Refresh()

				petAttackButton.Disable()
				petHangButton.Disable()
				petSkillButton.Disable()
				petRideButton.Disable()
				petDefenceButton.Disable()
				petHealSelfButton.Disable()
			})
			petHangButton.Importance = widget.SuccessImportance

			petSkillButton = widget.NewButton(P_SkILL, func() {
				worker.ActionState.AddPetState(P_SkILL)

				dialogs := []SelectorDialog{
					pIdSelectorDialog,
					pCUSuccessSelectorDialog,
					pCUFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			petSkillButton.Importance = widget.HighImportance

			petDefenceButton = widget.NewButton(P_DEFEND, func() {
				worker.ActionState.AddPetState(P_DEFEND)

				dialogs := []SelectorDialog{
					pIdSelectorDialog,
					pCUSuccessSelectorDialog,
					pCUFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			petDefenceButton.Importance = widget.HighImportance

			petHealSelfButton = widget.NewButton(P_SE_HEAL, func() {
				worker.ActionState.AddPetState(P_SE_HEAL)

				dialogs := []SelectorDialog{
					pIdSelectorDialog,
					pHealingRatioSelectorDialog,
					pCUSuccessSelectorDialog,
					pCUFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			petHealSelfButton.Importance = widget.HighImportance

			petHealOneButton = widget.NewButton(P_O_HEAL, func() {
				worker.ActionState.AddPetState(P_O_HEAL)

				dialogs := []SelectorDialog{
					pIdSelectorDialog,
					pHealingRatioSelectorDialog,
					pCUSuccessSelectorDialog,
					pCUFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			petHealOneButton.Importance = widget.HighImportance

			petRideButton = widget.NewButton(P_RIDE, func() {
				worker.ActionState.AddPetState(P_RIDE)

				dialogs := []SelectorDialog{
					pIdSelectorDialog,
					pCUSuccessSelectorDialog,
					pCUFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)

				petRideButton.Disable()
			})
			petRideButton.Importance = widget.HighImportance

			petOffRideButton = widget.NewButton(P_OFF_RIDE, func() {
				worker.ActionState.AddPetState(P_OFF_RIDE)

				dialogs := []SelectorDialog{
					pIdSelectorDialog,
					pCUSuccessSelectorDialog,
					pCUFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)

				petOffRideButton.Disable()
			})
			petOffRideButton.Importance = widget.HighImportance

			petEscapeButton = widget.NewButton(P_ESCAPE, func() {
				worker.ActionState.AddPetState(P_ESCAPE)

				dialogs := []SelectorDialog{
					pCUFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			petEscapeButton.Importance = widget.HighImportance

			actionStatesContainer := container.NewGridWithColumns(4,
				petAttackButton,
				petSkillButton,
				petDefenceButton,
				petEscapeButton,
				petRideButton,
				petOffRideButton,
				petHealSelfButton,
				petHealOneButton,
				petHangButton,
			)

			actionStatesDialog := dialog.NewCustom("Select pet actions with order", "Leave", actionStatesContainer, window)
			actionStatesDialog.Show()

		})
		humanStateSelector.Importance = widget.MediumImportance
		petStateSelector.Importance = widget.MediumImportance

		workerMenuContainer.Add(gameNameLabelContainer)
		workerMenuContainer.Add(movementModeButton)
		workerMenuContainer.Add(humanStateSelector)
		workerMenuContainer.Add(petStateSelector)

		statesViewer = container.NewAdaptiveGrid(5, generateTags(*worker)...)

		workerContainer := container.NewVBox(workerMenuContainer, statesViewer)
		configContainer.Add(workerContainer)
	}

	autoBattleWidget = container.NewVBox(widget.NewSeparator(), mainWidget, widget.NewSeparator(), configContainer)
	return autoBattleWidget, stopChan
}

func activateSelector(options []string, selector *widget.RadioGroup, dialog *dialog.CustomDialog, onChanged func(s string), dialogChan chan bool) {
	go func() {
		<-dialogChan
		selector.Options = options
		selector.OnChanged = onChanged
		dialog.Show()
	}()
}

type SelectorDialog struct {
	dialog    *dialog.CustomDialog
	selector  *widget.RadioGroup
	options   []string
	onChanged func(s string)
}

func activateDialogs(selectorDialogs []SelectorDialog, enableChan chan bool) {

	go func() {
		for i, sd := range selectorDialogs {
			<-enableChan
			sd.selector.Disable()
			sd.selector.Options = sd.options
			sd.selector.OnChanged = sd.onChanged
			sd.selector.Selected = ""
			sd.selector.Enable()
			sd.dialog.Show()

			if i == len(selectorDialogs)-1 {
				go func() {
					<-enableChan
				}()
			}
		}
	}()

	enableChan <- true
}

func generateTags(worker BattleWorker) (tagContaines []fyne.CanvasObject) {
	tagContaines = createTagContainer(worker.ActionState, true)
	tagContaines = append(tagContaines, createTagContainer(worker.ActionState, false)...)
	return
}

var (
	humanFinishingTagColor   = color.RGBA{245, 79, 0, uint8(math.Round(0.8 * 255))}
	humanConditionedTagColor = color.RGBA{0, 28, 245, uint8(math.Round(0.8 * 255))}
	humanSpecialTagColor     = color.RGBA{35, 128, 24, uint8(math.Round(0.8 * 255))}
	petTagColor              = color.RGBA{92, 14, 99, uint8(math.Round(0.8 * 255))}
)

func createTagContainer(actionState BattleActionState, isHuman bool) (tagContainers []fyne.CanvasObject) {
	var tags []string
	if isHuman {
		tags = actionState.GetHumanStates()
	} else {
		tags = actionState.GetPetStates()
	}

	for i, tag := range tags {
		tagColor := petTagColor
		if isHuman {
			switch {
			case strings.Contains(tag, "**"):
				tagColor = humanFinishingTagColor
			case strings.Contains(tag, "*"):
				tagColor = humanConditionedTagColor
			default:
				tagColor = humanSpecialTagColor
			}
		}

		tagCanvas := canvas.NewRectangle(tagColor)
		tagCanvas.SetMinSize(fyne.NewSize(60, 22))
		tagContainer := container.NewStack(tagCanvas)
		if isHuman {
			if v := actionState.GetHumanSkillIds()[i]; v != "" {
				tag = fmt.Sprintf("%s:%s:%s", tag, v, actionState.GetHumanSkillLevels()[i])
			}
			if param := actionState.GetHumanParams()[i]; param != "" {
				tag = fmt.Sprintf("%s:%s", tag, param)
			}

			id := 0
			successControlUnit := actionState.GetHumanSuccessControlUnits()[i]
			if len(successControlUnit) > 0 {
				id = 1
			}
			tag = fmt.Sprintf("%s:%s", tag, strings.ToLower(successControlUnit[:id]))

			id = 0
			failureControlUnit := actionState.GetHumanFailureControlUnits()[i]
			if len(failureControlUnit) > 0 {
				id = 1
			}
			tag = fmt.Sprintf("%s:%s", tag, strings.ToLower(failureControlUnit[:id]))

		} else {
			if v := actionState.GetPetSkillIds()[i]; v != "" {
				tag = fmt.Sprintf("%s:%s", tag, v)
			}
			if param := actionState.GetPetParams()[i]; param != "" {
				tag = fmt.Sprintf("%s:%s", tag, param)
			}

			id := 0
			successControlUnit := actionState.GetPetSuccessControlUnits()[i]
			if len(successControlUnit) > 0 {
				id = 1
			}
			tag = fmt.Sprintf("%s:%s", tag, strings.ToLower(successControlUnit[:id]))

			id = 0
			failureControlUnit := actionState.GetPetFailureControlUnits()[i]
			if len(failureControlUnit) > 0 {
				id = 1
			}
			tag = fmt.Sprintf("%s:%s", tag, strings.ToLower(failureControlUnit[:id]))
		}

		tagTextCanvas := canvas.NewText(tag, color.White)
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

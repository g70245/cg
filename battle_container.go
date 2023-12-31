package main

import (
	. "cg/game"
	. "cg/system"
	"encoding/json"
	"errors"
	"image/color"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type BattleGroups struct {
	stopChans map[int]chan bool
}

func (bgs *BattleGroups) stopAll() {
	for k := range bgs.stopChans {
		stop(bgs.stopChans[k])
	}
}

func battleContainer(games Games) (*fyne.Container, BattleGroups) {
	id := 0
	bgs := BattleGroups{make(map[int]chan bool)}

	groupTabs := container.NewAppTabs()
	groupTabs.SetTabLocation(container.TabLocationBottom)
	groupTabs.Hide()

	newGroupButton := widget.NewButtonWithIcon("New Group", theme.ContentAddIcon(), func() {
		groupNameEntry := widget.NewEntry()
		groupNameEntry.SetPlaceHolder("Enter group name")

		gamesCheckGroup := widget.NewCheckGroup(games.GetSortedKeys(), nil)
		gamesCheckGroup.Horizontal = true

		gamesSelectorDialog := dialog.NewCustom("Select games", "Create", container.NewVBox(groupNameEntry, gamesCheckGroup), window)
		gamesSelectorDialog.Resize(fyne.NewSize(240, 166))

		gamesSelectorDialog.SetOnClosed(func() {
			if len(gamesCheckGroup.Selected) == 0 {
				return
			}

			var newTabItem *container.TabItem
			newGroupContainer, stopChan := newBatttleGroupContainer(games.New(gamesCheckGroup.Selected), games, func(id int) func() {
				return func() {
					delete(bgs.stopChans, id)

					groupTabs.Remove(newTabItem)
					if len(bgs.stopChans) == 0 {
						groupTabs.Hide()
					}

					window.SetContent(window.Content())
					window.Resize(fyne.NewSize(APP_WIDTH, APP_HEIGHT))
				}
			}(id))
			bgs.stopChans[id] = stopChan

			var newGroupName string
			if groupNameEntry.Text != "" {
				newGroupName = groupNameEntry.Text
			} else {
				newGroupName = "Group " + fmt.Sprint(id)
			}
			newTabItem = container.NewTabItem(newGroupName, newGroupContainer)
			groupTabs.Append(newTabItem)
			groupTabs.Show()

			window.SetContent(window.Content())
			window.Resize(fyne.NewSize(APP_WIDTH, APP_HEIGHT))
			id++
		})
		gamesSelectorDialog.Show()
	})

	menu := container.NewVBox(newGroupButton)
	main := container.NewBorder(menu, nil, nil, nil, groupTabs)

	return main, bgs
}

func newBatttleGroupContainer(games Games, allGames Games, destroy func()) (autoBattleWidget *fyne.Container, sharedStopChan chan bool) {
	manaChecker := new(string)
	sharedInventoryStatus := new(bool)
	sharedStopChan = make(chan bool, len(games))
	workers := CreateBattleWorkers(games.GetHWNDs(), logDir, manaChecker, sharedInventoryStatus, sharedStopChan)

	var manaCheckerSelectorDialog *dialog.CustomDialog
	var manaCheckerSelectorButton *widget.Button
	manaCheckerSelector := widget.NewRadioGroup(games.GetSortedKeys(), func(s string) {
		if s != "" {
			*manaChecker = fmt.Sprint(allGames[s])
			manaCheckerSelectorButton.SetText("Mana Checker: " + s)
		} else {
			*manaChecker = NONE
			manaCheckerSelectorButton.SetText("Select Mana Checker")
		}
		manaCheckerSelectorDialog.Hide()
	})
	manaCheckerSelectorDialog = dialog.NewCustomWithoutButtons("Select a mana checker with this group", manaCheckerSelector, window)
	manaCheckerSelectorButton = widget.NewButton("Select Mana Checker", func() {
		manaCheckerSelectorDialog.Show()

		notifyBeeperConfig("About Mana Checker")
	})
	manaCheckerSelectorButton.Importance = widget.HighImportance

	var switchButton *widget.Button
	switchButton = widget.NewButtonWithIcon("", theme.MediaPlayIcon(), func() {
		switch switchButton.Icon {
		case theme.MediaPlayIcon():
			for i := range workers {
				workers[i].Work()
			}
			turn(theme.MediaStopIcon(), switchButton)
		case theme.MediaStopIcon():
			for i := range workers {
				workers[i].Stop()
			}
			turn(theme.MediaPlayIcon(), switchButton)
		}
	})
	switchButton.Importance = widget.WarningImportance

	deleteButton := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		deleteDialog := dialog.NewConfirm("Delete group", "Do you really want to delete this group?", func(isDeleting bool) {
			if isDeleting {
				for i := range workers {
					workers[i].Stop()
				}

				close(sharedStopChan)
				destroy()
			}
		}, window)
		deleteDialog.SetConfirmImportance(widget.DangerImportance)
		deleteDialog.Show()
	})
	deleteButton.Importance = widget.DangerImportance

	var teleportAndResourceCheckerButton *widget.Button
	teleportAndResourceCheckerButton = widget.NewButtonWithIcon("Check TP & RES", theme.CheckButtonIcon(), func() {
		switch teleportAndResourceCheckerButton.Icon {
		case theme.CheckButtonCheckedIcon():
			for i := range workers {
				workers[i].StopTeleportAndResourceChecker()
			}
			turn(theme.CheckButtonIcon(), teleportAndResourceCheckerButton)
		case theme.CheckButtonIcon():
			for i := range workers {
				workers[i].StartTeleportAndResourceChecker()
			}
			turn(theme.CheckButtonCheckedIcon(), teleportAndResourceCheckerButton)

			notifyBeeperAndLogConfig("About Teleport & Resource Checker")
		}
	})
	teleportAndResourceCheckerButton.Importance = widget.HighImportance

	var activitiesCheckerButton *widget.Button
	activitiesCheckerButton = widget.NewButtonWithIcon("Check Activities", theme.CheckButtonIcon(), func() {
		switch activitiesCheckerButton.Icon {
		case theme.CheckButtonCheckedIcon():
			for i := range workers {
				workers[i].ActivityCheckerEnabled = false
			}
			turn(theme.CheckButtonIcon(), activitiesCheckerButton)
		case theme.CheckButtonIcon():
			for i := range workers {
				workers[i].ActivityCheckerEnabled = true
			}
			turn(theme.CheckButtonCheckedIcon(), activitiesCheckerButton)

			notifyBeeperAndLogConfig("About Activities Checker")
		}
	})
	activitiesCheckerButton.Importance = widget.HighImportance

	var logCheckersButton *widget.Button
	logCheckersButton = widget.NewButtonWithIcon("Log Checkers", theme.MenuIcon(), func() {
		dialog.NewCustom("Log Checkers", "Leave", container.NewAdaptiveGrid(4, teleportAndResourceCheckerButton, activitiesCheckerButton), window).Show()
	})
	logCheckersButton.Importance = widget.HighImportance

	var inventoryCheckerButton *widget.Button
	inventoryCheckerButton = widget.NewButtonWithIcon("Check Inventory", theme.CheckButtonIcon(), func() {
		switch inventoryCheckerButton.Icon {
		case theme.CheckButtonCheckedIcon():
			for i := range workers {
				workers[i].StopInventoryChecker()
			}
			turn(theme.CheckButtonIcon(), inventoryCheckerButton)
		case theme.CheckButtonIcon():
			for i := range workers {
				workers[i].StartInventoryChecker()
			}
			turn(theme.CheckButtonCheckedIcon(), inventoryCheckerButton)

			notifyBeeperConfig("About Inventory Checker")
		}
	})
	inventoryCheckerButton.Importance = widget.HighImportance

	mainButtons := container.NewGridWithColumns(5, manaCheckerSelectorButton, logCheckersButton, inventoryCheckerButton, deleteButton, switchButton)
	mainWidget := container.NewVBox(mainButtons)

	/* Configuration Widgets */
	configContainer := container.NewVBox()
	for i := range workers {
		workerMenuContainer := container.NewGridWithColumns(6)
		worker := &workers[i]

		var nicknameButton *widget.Button
		nicknameEntry := widget.NewEntry()
		nicknameEntry.SetPlaceHolder("Enter nickname")
		nicknameButton = widget.NewButtonWithIcon(allGames.FindKey(worker.GetHandle()), theme.AccountIcon(), func() {
			nicknameDialog := dialog.NewCustom("Enter nickname", "Ok", nicknameEntry, window)
			nicknameDialog.SetOnClosed(func() {
				if _, ok := allGames[nicknameEntry.Text]; nicknameEntry.Text == "" || ok {
					return
				}

				allGames.RemoveValue(worker.GetHandle())
				allGames.Add(nicknameEntry.Text, worker.GetHandle())
				nicknameButton.SetText(nicknameEntry.Text)
			})
			nicknameDialog.Show()
		})

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
		movementModeDialog = dialog.NewCustomWithoutButtons("Select a move way", movementModeSelector, window)
		movementModeButton = widget.NewButtonWithIcon("Move Way", theme.MailReplyIcon(), func() {
			movementModeDialog.Show()
		})

		var statesViewer *fyne.Container
		var selector = widget.NewRadioGroup(nil, nil)
		selector.Horizontal = true
		selector.Required = true
		enableChan := make(chan bool)

		onClosed := func() {
			statesViewer.Objects = generateTags(*worker)
			statesViewer.Refresh()
			enableChan <- true
		}

		/* Control Unit Dialogs */
		var cuSuccessDialog *dialog.CustomDialog
		var cuFailureDialog *dialog.CustomDialog

		activateJumpDialog := func(totalStates int, callback func(cu string)) {
			jumpEntry := widget.NewEntry()
			jumpEntry.Validator = func(offsetStr string) error {
				if offset, err := strconv.Atoi(offsetStr); err != nil {
					return err
				} else if offset >= totalStates-1 || offset < 1 {
					return errors.New("not a valid offset")
				}
				return nil
			}

			jumpDialog := dialog.NewForm("Enter next action offset", "Ok", "Dismiss", []*widget.FormItem{widget.NewFormItem("Offset", jumpEntry)}, func(isValid bool) {
				if isValid {
					callback(C_U_JUMP + jumpEntry.Text)
				}
			}, window)
			jumpDialog.Show()
		}
		humanCUSuccesOnChanged := func(s string) {
			if s != "" {
				if s == C_U_JUMP {
					activateJumpDialog(len(worker.ActionState.HumanStates), func(cu string) {
						worker.ActionState.AddHumanSuccessControlUnit(cu)
						cuSuccessDialog.Hide()
					})
				} else {
					worker.ActionState.AddHumanSuccessControlUnit(s)
					cuSuccessDialog.Hide()
				}
			}
		}
		humanCUFailureOnChanged := func(s string) {
			if s != "" {
				if s == C_U_JUMP {
					activateJumpDialog(len(worker.ActionState.HumanStates), func(cu string) {
						worker.ActionState.AddHumanFailureControlUnit(cu)
						cuFailureDialog.Hide()
					})
				} else {
					worker.ActionState.AddHumanFailureControlUnit(s)
					cuFailureDialog.Hide()
				}
			}
		}
		petCUSuccessOnChanged := func(s string) {
			if s != "" {
				if s == C_U_JUMP {
					activateJumpDialog(len(worker.ActionState.PetStates), func(cu string) {
						worker.ActionState.AddPetSuccessControlUnit(cu)
						cuSuccessDialog.Hide()
					})
				} else {
					worker.ActionState.AddPetSuccessControlUnit(s)
					cuSuccessDialog.Hide()
				}
			}
		}
		petCUFailureOnChanged := func(s string) {
			if s != "" {
				if s == C_U_JUMP {
					activateJumpDialog(len(worker.ActionState.PetStates), func(cu string) {
						worker.ActionState.AddPetFailureControlUnit(cu)
						cuFailureDialog.Hide()
					})
				} else {
					worker.ActionState.AddPetFailureControlUnit(s)
					cuFailureDialog.Hide()
				}
			}
		}

		cuOnClosed := func() {
			statesViewer.Objects = generateTags(*worker)
			statesViewer.Refresh()
			enableChan <- true
		}
		cuSuccessDialog = dialog.NewCustomWithoutButtons("Select next action after successful execution", selector, window)
		cuSuccessDialog.SetOnClosed(cuOnClosed)
		cuFailureDialog = dialog.NewCustomWithoutButtons("Select next action after failed execution", selector, window)
		cuFailureDialog.SetOnClosed(cuOnClosed)

		humanCUSuccessSelectorDialog := SelectorDialog{
			cuSuccessDialog,
			selector,
			ControlUnitOptions,
			humanCUSuccesOnChanged,
		}
		humanCUFailureSelectorDialog := SelectorDialog{
			cuFailureDialog,
			selector,
			ControlUnitOptions,
			humanCUFailureOnChanged,
		}
		petCUSuccessSelectorDialog := SelectorDialog{
			cuSuccessDialog,
			selector,
			ControlUnitOptions,
			petCUSuccessOnChanged,
		}
		petCUFailureSelectorDialog := SelectorDialog{
			cuFailureDialog,
			selector,
			ControlUnitOptions,
			petCUFailureOnChanged,
		}

		/* Param Dialogs */
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
		paramsDialog = dialog.NewCustomWithoutButtons("Select param", selector, window)
		paramsDialog.SetOnClosed(onClosed)

		// healing
		humanHealingRatioSelectorDialog := SelectorDialog{
			paramsDialog,
			selector,
			HealingOptions,
			humanParamsOnChanged,
		}

		petHealingRatioSelectorDialog := SelectorDialog{
			paramsDialog,
			selector,
			HealingOptions,
			petParamsOnChanged,
		}

		// bomb
		bombSelectorDialog := SelectorDialog{
			paramsDialog,
			selector,
			Bombs.GetOptions(),
			humanParamsOnChanged,
		}

		// threshold
		humanThresholdSelectorDialog := SelectorDialog{
			paramsDialog,
			selector,
			ThresholdOptions,
			humanParamsOnChanged,
		}

		/* Id Dialogs */
		var idDialog *dialog.CustomDialog
		humanIdOnChanged := func(s string) {
			if s != "" {
				worker.ActionState.AddHumanSkillId(s)
				idDialog.Hide()
			}
		}
		petIdOnChanged := func(s string) {
			if s != "" {
				worker.ActionState.AddPetSkillId(s)
				idDialog.Hide()
			}
		}
		idDialog = dialog.NewCustomWithoutButtons("Select skill id", selector, window)
		idDialog.SetOnClosed(onClosed)
		humanIdSelectorDialog := SelectorDialog{
			idDialog,
			selector,
			IdOptions,
			humanIdOnChanged,
		}
		petIdSelectorDialog := SelectorDialog{
			idDialog,
			selector,
			IdOptions,
			petIdOnChanged,
		}

		/* Level Dialog */
		var levelDialog *dialog.CustomDialog
		levelOnChanged := func(s string) {
			if s != "" {
				worker.ActionState.AddHumanSkillLevel(s)
				levelDialog.Hide()
			}
		}
		levelDialog = dialog.NewCustomWithoutButtons("Select skill level", selector, window)
		levelDialog.SetOnClosed(onClosed)
		levelSelectorDialog := SelectorDialog{
			levelDialog,
			selector,
			LevelOptions,
			levelOnChanged,
		}

		humanStateSelector := widget.NewButtonWithIcon("Man Actions", theme.ContentAddIcon(), func() {
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
			var thresholdSkillButton *widget.Button
			var stealButton *widget.Button
			var trainButton *widget.Button
			var rideButton *widget.Button
			var healSelfButton *widget.Button
			var healOneButton *widget.Button
			var healTShapeButton *widget.Button
			var healMultiButton *widget.Button

			attackButton = widget.NewButton(H_F_ATTACK, func() {
				worker.ActionState.AddHumanState(H_F_ATTACK)

				dialogs := []SelectorDialog{
					humanCUSuccessSelectorDialog,
					humanCUFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			attackButton.Importance = widget.WarningImportance

			defendButton = widget.NewButton(H_F_DEFEND, func() {
				worker.ActionState.AddHumanState(H_F_DEFEND)

				dialogs := []SelectorDialog{
					humanCUSuccessSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			defendButton.Importance = widget.WarningImportance

			escapeButton = widget.NewButton(H_F_ESCAPE, func() {
				worker.ActionState.AddHumanState(H_F_ESCAPE)

				dialogs := []SelectorDialog{
					humanCUFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			escapeButton.Importance = widget.WarningImportance

			catchButton = widget.NewButton(H_S_CATCH, func() {
				worker.ActionState.AddHumanState(H_S_CATCH)
				statesViewer.Objects = generateTags(*worker)
				statesViewer.Refresh()

				dialogs := []SelectorDialog{
					humanHealingRatioSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)

				notifyLogConfig("About Catch")
			})
			catchButton.Importance = widget.SuccessImportance

			bombButton = widget.NewButton(H_C_BOMB, func() {
				worker.ActionState.AddHumanState(H_C_BOMB)

				dialogs := []SelectorDialog{
					bombSelectorDialog,
					humanCUSuccessSelectorDialog,
					humanCUFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			bombButton.Importance = widget.HighImportance

			potionButton = widget.NewButton(H_C_POTION, func() {
				worker.ActionState.AddHumanState(H_C_POTION)

				dialogs := []SelectorDialog{
					humanHealingRatioSelectorDialog,
					humanCUSuccessSelectorDialog,
					humanCUFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			potionButton.Importance = widget.HighImportance

			recallButton = widget.NewButton(H_C_PET_RECALL, func() {
				worker.ActionState.AddHumanState(H_C_PET_RECALL)

				dialogs := []SelectorDialog{
					humanCUSuccessSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			recallButton.Importance = widget.HighImportance

			moveButton = widget.NewButton(H_F_MOVE, func() {
				worker.ActionState.AddHumanState(H_F_MOVE)

				dialogs := []SelectorDialog{
					humanCUSuccessSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			moveButton.Importance = widget.WarningImportance

			hangButton = widget.NewButton(H_S_HANG, func() {
				worker.ActionState.AddHumanState(H_S_HANG)
				statesViewer.Objects = generateTags(*worker)
				statesViewer.Refresh()
			})
			hangButton.Importance = widget.SuccessImportance

			skillButton = widget.NewButton(H_C_SKILL, func() {
				worker.ActionState.AddHumanState(H_C_SKILL)

				dialogs := []SelectorDialog{
					humanIdSelectorDialog,
					levelSelectorDialog,
					humanCUSuccessSelectorDialog,
					humanCUFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			skillButton.Importance = widget.HighImportance

			thresholdSkillButton = widget.NewButton(H_C_T_SKILL, func() {
				worker.ActionState.AddHumanState(H_C_T_SKILL)

				dialogs := []SelectorDialog{
					humanIdSelectorDialog,
					levelSelectorDialog,
					humanThresholdSelectorDialog,
					humanCUSuccessSelectorDialog,
					humanCUFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			thresholdSkillButton.Importance = widget.HighImportance

			stealButton = widget.NewButton(H_S_STEAL, func() {
				worker.ActionState.AddHumanState(H_S_STEAL)

				dialogs := []SelectorDialog{
					humanIdSelectorDialog,
					levelSelectorDialog,
					humanCUSuccessSelectorDialog,
					humanCUFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			stealButton.Importance = widget.SuccessImportance

			trainButton = widget.NewButton(H_S_TRAIN_SKILL, func() {
				worker.ActionState.AddHumanState(H_S_TRAIN_SKILL)

				dialogs := []SelectorDialog{
					humanIdSelectorDialog,
					levelSelectorDialog,
					humanCUSuccessSelectorDialog,
					humanCUFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			trainButton.Importance = widget.SuccessImportance

			rideButton = widget.NewButton(H_C_RIDE, func() {
				worker.ActionState.AddHumanState(H_C_RIDE)

				dialogs := []SelectorDialog{
					humanIdSelectorDialog,
					levelSelectorDialog,
					humanCUSuccessSelectorDialog,
					humanCUFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			rideButton.Importance = widget.HighImportance

			healSelfButton = widget.NewButton(H_C_SE_HEAL, func() {
				worker.ActionState.AddHumanState(H_C_SE_HEAL)

				dialogs := []SelectorDialog{
					humanIdSelectorDialog,
					levelSelectorDialog,
					humanHealingRatioSelectorDialog,
					humanCUSuccessSelectorDialog,
					humanCUFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			healSelfButton.Importance = widget.HighImportance

			healOneButton = widget.NewButton(H_C_O_HEAL, func() {
				worker.ActionState.AddHumanState(H_C_O_HEAL)

				dialogs := []SelectorDialog{
					humanIdSelectorDialog,
					levelSelectorDialog,
					humanHealingRatioSelectorDialog,
					humanCUSuccessSelectorDialog,
					humanCUFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			healOneButton.Importance = widget.HighImportance

			healTShapeButton = widget.NewButton(H_C_T_HEAL, func() {
				worker.ActionState.AddHumanState(H_C_T_HEAL)

				dialogs := []SelectorDialog{
					humanIdSelectorDialog,
					levelSelectorDialog,
					humanHealingRatioSelectorDialog,
					humanCUSuccessSelectorDialog,
					humanCUFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			healTShapeButton.Importance = widget.HighImportance

			healMultiButton = widget.NewButton(H_C_M_HEAL, func() {
				worker.ActionState.AddHumanState(H_C_M_HEAL)

				dialogs := []SelectorDialog{
					humanIdSelectorDialog,
					levelSelectorDialog,
					humanHealingRatioSelectorDialog,
					humanCUSuccessSelectorDialog,
					humanCUFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			healMultiButton.Importance = widget.HighImportance

			actionStatesContainer := container.NewGridWithColumns(4,
				attackButton,
				defendButton,
				escapeButton,
				moveButton,
				skillButton,
				thresholdSkillButton,
				bombButton,
				recallButton,
				rideButton,
				potionButton,
				healSelfButton,
				healOneButton,
				healTShapeButton,
				healMultiButton,
				hangButton,
				catchButton,
				stealButton,
				trainButton,
			)

			actionStatesDialog := dialog.NewCustom("Select man actions with order", "Leave", actionStatesContainer, window)
			actionStatesDialog.Show()
		})

		petStateSelector := widget.NewButtonWithIcon("Pet Actions", theme.ContentAddIcon(), func() {
			worker.ActionState.ClearPetStates()
			statesViewer.Objects = generateTags(*worker)
			statesViewer.Refresh()

			var petAttackButton *widget.Button
			var petHangButton *widget.Button
			var petSkillButton *widget.Button
			var petDefendButton *widget.Button
			var petEscapeButton *widget.Button
			var petHealSelfButton *widget.Button
			var petHealOneButton *widget.Button
			var petRideButton *widget.Button
			var petOffRideButton *widget.Button
			var petCatchButton *widget.Button

			petAttackButton = widget.NewButton(P_F_ATTACK, func() {
				worker.ActionState.AddPetState(P_F_ATTACK)

				dialogs := []SelectorDialog{
					petCUSuccessSelectorDialog,
					petCUFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			petAttackButton.Importance = widget.WarningImportance

			petHangButton = widget.NewButton(P_S_HANG, func() {
				worker.ActionState.AddPetState(P_S_HANG)
				statesViewer.Objects = generateTags(*worker)
				statesViewer.Refresh()
			})
			petHangButton.Importance = widget.SuccessImportance

			petSkillButton = widget.NewButton(P_C_SkILL, func() {
				worker.ActionState.AddPetState(P_C_SkILL)

				dialogs := []SelectorDialog{
					petIdSelectorDialog,
					petCUSuccessSelectorDialog,
					petCUFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			petSkillButton.Importance = widget.HighImportance

			petDefendButton = widget.NewButton(P_C_DEFEND, func() {
				worker.ActionState.AddPetState(P_C_DEFEND)

				dialogs := []SelectorDialog{
					petIdSelectorDialog,
					petCUSuccessSelectorDialog,
					petCUFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			petDefendButton.Importance = widget.HighImportance

			petHealSelfButton = widget.NewButton(P_C_SE_HEAL, func() {
				worker.ActionState.AddPetState(P_C_SE_HEAL)

				dialogs := []SelectorDialog{
					petIdSelectorDialog,
					petHealingRatioSelectorDialog,
					petCUSuccessSelectorDialog,
					petCUFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			petHealSelfButton.Importance = widget.HighImportance

			petHealOneButton = widget.NewButton(P_C_O_HEAL, func() {
				worker.ActionState.AddPetState(P_C_O_HEAL)

				dialogs := []SelectorDialog{
					petIdSelectorDialog,
					petHealingRatioSelectorDialog,
					petCUSuccessSelectorDialog,
					petCUFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			petHealOneButton.Importance = widget.HighImportance

			petRideButton = widget.NewButton(P_C_RIDE, func() {
				worker.ActionState.AddPetState(P_C_RIDE)

				dialogs := []SelectorDialog{
					petIdSelectorDialog,
					petCUSuccessSelectorDialog,
					petCUFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			petRideButton.Importance = widget.HighImportance

			petOffRideButton = widget.NewButton(P_C_OFF_RIDE, func() {
				worker.ActionState.AddPetState(P_C_OFF_RIDE)

				dialogs := []SelectorDialog{
					petIdSelectorDialog,
					petCUSuccessSelectorDialog,
					petCUFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			petOffRideButton.Importance = widget.HighImportance

			petEscapeButton = widget.NewButton(P_F_ESCAPE, func() {
				worker.ActionState.AddPetState(P_F_ESCAPE)

				dialogs := []SelectorDialog{
					petCUFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			petEscapeButton.Importance = widget.WarningImportance

			petCatchButton = widget.NewButton(P_S_CATCH, func() {
				worker.ActionState.AddPetState(P_S_CATCH)
				statesViewer.Objects = generateTags(*worker)
				statesViewer.Refresh()

				notifyLogConfig("About Catch")

				dialogs := []SelectorDialog{
					petHealingRatioSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			petCatchButton.Importance = widget.SuccessImportance

			actionStatesContainer := container.NewGridWithColumns(4,
				petAttackButton,
				petEscapeButton,
				petDefendButton,
				petSkillButton,
				petRideButton,
				petOffRideButton,
				petHealSelfButton,
				petHealOneButton,
				petCatchButton,
				petHangButton,
			)

			actionStatesDialog := dialog.NewCustom("Select pet actions with order", "Leave", actionStatesContainer, window)
			actionStatesDialog.Show()

		})
		humanStateSelector.Importance = widget.MediumImportance
		petStateSelector.Importance = widget.MediumImportance

		loadSettingButton := widget.NewButtonWithIcon("Load", theme.FolderOpenIcon(), func() {
			var fileOpenDialog *dialog.FileDialog
			fileOpenDialog = dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
				if uc != nil {
					var actionState BattleActionState
					file, openErr := os.Open(uc.URI().Path())
					defer file.Close()

					if openErr == nil {
						if buffer, readErr := io.ReadAll(file); readErr == nil {
							if json.Unmarshal(buffer, &actionState) == nil {
								actionState.SetHWND(worker.ActionState.GetHWND())
								worker.ActionState = actionState
								worker.ActionState.LogDir = logDir
								worker.ActionState.ManaChecker = manaChecker
								statesViewer.Objects = generateTags(*worker)
								statesViewer.Refresh()
							}
						}
					}
				}
			}, window)

			listableURI, _ := storage.ListerForURI(storage.NewFileURI(DEFAULT_ROOT + `\actions`))
			fileOpenDialog.SetLocation(listableURI)
			fileOpenDialog.SetFilter(storage.NewExtensionFileFilter([]string{".ac"}))
			fileOpenDialog.Show()
		})
		loadSettingButton.Importance = widget.MediumImportance

		saveSettingButton := widget.NewButtonWithIcon("Save", theme.DownloadIcon(), func() {
			fileSaveDialog := dialog.NewFileSave(func(uc fyne.URIWriteCloser, err error) {
				if uc != nil {
					if setting, marshalErr := json.Marshal(worker.ActionState); marshalErr == nil {
						if writeErr := os.WriteFile(uc.URI().Path(), setting, 0644); writeErr != nil {
							log.Fatalf("Cannot write to file: %s\n", uc.URI().Path())
						}
					}
				}
			}, window)
			listableURI, _ := storage.ListerForURI(storage.NewFileURI(DEFAULT_ROOT + `\actions`))
			fileSaveDialog.SetFileName("default.ac")
			fileSaveDialog.SetLocation(listableURI)
			fileSaveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".ac"}))
			fileSaveDialog.Show()
		})
		saveSettingButton.Importance = widget.MediumImportance

		workerMenuContainer.Add(nicknameButton)
		workerMenuContainer.Add(movementModeButton)
		workerMenuContainer.Add(humanStateSelector)
		workerMenuContainer.Add(petStateSelector)
		workerMenuContainer.Add(loadSettingButton)
		workerMenuContainer.Add(saveSettingButton)

		statesViewer = container.NewAdaptiveGrid(6, generateTags(*worker)...)

		workerContainer := container.NewVBox(workerMenuContainer, statesViewer)
		configContainer.Add(workerContainer)
	}

	autoBattleWidget = container.NewVBox(widget.NewSeparator(), mainWidget, widget.NewSeparator(), configContainer)
	return autoBattleWidget, sharedStopChan
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
	tagContaines = createTagContainers(worker.ActionState, true)
	tagContaines = append(tagContaines, createTagContainers(worker.ActionState, false)...)
	return
}

var (
	humanFinishingTagColor   = color.RGBA{235, 206, 100, uint8(math.Round(1 * 255))}
	humanConditionalTagColor = color.RGBA{100, 206, 235, uint8(math.Round(1 * 255))}
	humanSpecialTagColor     = color.RGBA{206, 235, 100, uint8(math.Round(1 * 255))}
	petFinishingTagColor     = color.RGBA{245, 79, 0, uint8(math.Round(0.8 * 255))}
	petConditionalTagColor   = color.RGBA{0, 79, 245, uint8(math.Round(0.8 * 255))}
	petSpecialTagColor       = color.RGBA{79, 245, 0, uint8(math.Round(0.8 * 255))}
)

func createTagContainers(actionState BattleActionState, isHuman bool) (tagContainers []fyne.CanvasObject) {
	var tags []string
	if isHuman {
		tags = actionState.GetHumanStates()
	} else {
		tags = actionState.GetPetStates()
	}

	for i, tag := range tags {
		tagColor := petSpecialTagColor
		if isHuman {
			switch {
			case strings.Contains(tag, "**"):
				tagColor = humanFinishingTagColor
			case strings.Contains(tag, "*"):
				tagColor = humanConditionalTagColor
			default:
				tagColor = humanSpecialTagColor
			}
		} else {
			switch {
			case strings.Contains(tag, "**"):
				tagColor = petFinishingTagColor
			case strings.Contains(tag, "*"):
				tagColor = petConditionalTagColor
			default:
				tagColor = petSpecialTagColor
			}
		}

		tagCanvas := canvas.NewRectangle(tagColor)
		tagCanvas.SetMinSize(fyne.NewSize(60, 22))
		tagContainer := container.NewStack(tagCanvas)

		if isHuman {
			if v := actionState.GetHumanSkillIds()[i]; v != "" {
				tag = fmt.Sprintf("%s:%s:%s", tag, v, actionState.GetHumanSkillLevels()[i])
			}
		} else {
			if v := actionState.GetPetSkillIds()[i]; v != "" {
				tag = fmt.Sprintf("%s:%s", tag, v)
			}
		}

		var params []string
		var successControlUnits []string
		var failureControlUnits []string

		if isHuman {
			params = actionState.GetHumanParams()
			successControlUnits = actionState.GetHumanSuccessControlUnits()
			failureControlUnits = actionState.GetHumanFailureControlUnits()
		} else {
			params = actionState.GetPetParams()
			successControlUnits = actionState.GetPetSuccessControlUnits()
			failureControlUnits = actionState.GetPetFailureControlUnits()
		}

		if param := params[i]; param != "" {
			tag = fmt.Sprintf("%s:%s", tag, param)
		}

		if len(successControlUnits[i]) > 0 {
			cuFirstLetter := strings.ToLower(successControlUnits[i][:1])
			tag = fmt.Sprintf("%s:%s", tag, cuFirstLetter)
			if cuFirstLetter == "j" {
				tag = fmt.Sprintf("%s%s", tag, successControlUnits[i][4:])
			}
		} else {
			tag = tag + ":"
		}

		if len(failureControlUnits[i]) > 0 {
			cuFirstLetter := strings.ToLower(failureControlUnits[i][:1])
			tag = fmt.Sprintf("%s:%s", tag, cuFirstLetter)
			if cuFirstLetter == "j" {
				tag = fmt.Sprintf("%s%s", tag, failureControlUnits[i][4:])
			}
		} else {
			tag = tag + ":"
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

func notifyBeeperConfig(title string) {
	if !Beeper.IsReady() {
		go func() {
			time.Sleep(200 * time.Millisecond)
			dialog.NewInformation(title, "Remember to setup the alert music!!!", window).Show()
		}()
	}
}

func notifyLogConfig(title string) {
	if *logDir == "" {
		go func() {
			time.Sleep(200 * time.Millisecond)
			dialog.NewInformation(title, "Remember to setup the log directory!!!", window).Show()
		}()
	}
}

func notifyBeeperAndLogConfig(title string) {
	if !Beeper.IsReady() || *logDir == "" {
		go func() {
			time.Sleep(200 * time.Millisecond)
			dialog.NewInformation(title, "Remember to setup the alert music and log directory!!!", window).Show()
		}()
	}
}

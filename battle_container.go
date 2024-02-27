package main

import (
	. "cg/game"
	. "cg/utils"
	"encoding/json"
	"errors"
	"image/color"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
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
	manaChecker := NO_MANA_CHECKER
	sharedInventoryStatus := new(bool)
	sharedStopChan = make(chan bool, len(games))
	workers := CreateBattleWorkers(games, logDir, &manaChecker, sharedInventoryStatus, sharedStopChan, new(sync.WaitGroup))

	var manaCheckerSelectorDialog *dialog.CustomDialog
	var manaCheckerSelectorButton *widget.Button
	manaCheckerOptions := []string{NO_MANA_CHECKER}
	manaCheckerOptions = append(manaCheckerOptions, games.GetSortedKeys()...)
	manaCheckerSelector := widget.NewRadioGroup(manaCheckerOptions, func(s string) {
		if game, ok := allGames[s]; ok {
			manaChecker = fmt.Sprint(game)
			manaCheckerSelectorButton.SetText(fmt.Sprintf("Mana Checker: %s", s))
		} else {
			manaChecker = NO_MANA_CHECKER
			manaCheckerSelectorButton.SetText(fmt.Sprintf("Mana Checker: %s", manaChecker))
		}
		manaCheckerSelectorDialog.Hide()
	})
	manaCheckerSelectorDialog = dialog.NewCustomWithoutButtons("Select a mana checker with this group", manaCheckerSelector, window)
	manaCheckerSelectorButton = widget.NewButton(fmt.Sprintf("Mana Checker: %s", manaChecker), func() {
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
		movementModeSelector := widget.NewRadioGroup(BATTLE_MOVEMENT_MODES.GetOptions(), func(s string) {
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
		var controlUnitSuccessDialog *dialog.CustomDialog
		var controlUnitFailureDialog *dialog.CustomDialog

		activateJumpDialog := func(totalStates int, callback func(jumpId int)) {
			jumpEntry := widget.NewEntry()
			jumpEntry.Validator = func(jumpIdStr string) error {
				if jumpId, err := strconv.Atoi(jumpIdStr); err != nil {
					return err
				} else if jumpId >= totalStates-1 || jumpId < 1 {
					return errors.New("not a valid offset")
				}
				return nil
			}

			jumpDialog := dialog.NewForm("Enter next action id", "Ok", "Dismiss", []*widget.FormItem{widget.NewFormItem("Jump Id", jumpEntry)}, func(isValid bool) {
				if isValid {
					jumpId, _ := strconv.Atoi(jumpEntry.Text)
					callback(jumpId)
				}
			}, window)
			jumpDialog.Show()
		}
		humanControlUnitSuccesOnChanged := func(s string) {
			if s == "" {
				return
			}

			if cu := ControlUnit(s); cu == ControlUnitJump {
				activateJumpDialog(len(worker.ActionState.HumanStates), func(jumpId int) {
					worker.ActionState.AddHumanSuccessControlUnit(cu)
					worker.ActionState.AddHumanSuccessJumpId(jumpId)
					controlUnitSuccessDialog.Hide()
				})
			} else {
				worker.ActionState.AddHumanSuccessControlUnit(cu)
				controlUnitSuccessDialog.Hide()
			}
		}
		humanControlUnitFailureOnChanged := func(s string) {
			if s == "" {
				return
			}

			if cu := ControlUnit(s); cu == ControlUnitJump {
				activateJumpDialog(len(worker.ActionState.HumanStates), func(jumpId int) {
					worker.ActionState.AddHumanFailureControlUnit(cu)
					worker.ActionState.AddHumanFailureJumpId(jumpId)
					controlUnitFailureDialog.Hide()
				})
			} else {
				worker.ActionState.AddHumanFailureControlUnit(cu)
				controlUnitFailureDialog.Hide()
			}
		}
		petControlUnitSuccessOnChanged := func(s string) {
			if s == "" {
				return
			}

			if cu := ControlUnit(s); cu == ControlUnitJump {
				activateJumpDialog(len(worker.ActionState.PetStates), func(jumpId int) {
					worker.ActionState.AddPetSuccessControlUnit(cu)
					worker.ActionState.AddPetSuccessJumpId(jumpId)
					controlUnitSuccessDialog.Hide()
				})
			} else {
				worker.ActionState.AddPetSuccessControlUnit(cu)
				controlUnitSuccessDialog.Hide()
			}
		}
		petControlUnitFailureOnChanged := func(s string) {
			if s == "" {
				return
			}
			if cu := ControlUnit(s); cu == ControlUnitJump {
				activateJumpDialog(len(worker.ActionState.PetStates), func(jumpId int) {
					worker.ActionState.AddPetFailureControlUnit(cu)
					worker.ActionState.AddPetFailueJumpId(jumpId)
					controlUnitFailureDialog.Hide()
				})
			} else {
				worker.ActionState.AddPetFailureControlUnit(cu)
				controlUnitFailureDialog.Hide()
			}
		}

		controlUnitOnClosed := func() {
			statesViewer.Objects = generateTags(*worker)
			statesViewer.Refresh()
			enableChan <- true
		}
		controlUnitSuccessDialog = dialog.NewCustomWithoutButtons("Select next action after successful execution", selector, window)
		controlUnitSuccessDialog.SetOnClosed(controlUnitOnClosed)
		controlUnitFailureDialog = dialog.NewCustomWithoutButtons("Select next action after failed execution", selector, window)
		controlUnitFailureDialog.SetOnClosed(controlUnitOnClosed)

		humanControlUnitSuccessSelectorDialog := SelectorDialog{
			controlUnitSuccessDialog,
			selector,
			ControlUnits.GetOptions(),
			humanControlUnitSuccesOnChanged,
		}
		humanControlUnitFailureSelectorDialog := SelectorDialog{
			controlUnitFailureDialog,
			selector,
			ControlUnits.GetOptions(),
			humanControlUnitFailureOnChanged,
		}
		petControlUnitSuccessSelectorDialog := SelectorDialog{
			controlUnitSuccessDialog,
			selector,
			ControlUnits.GetOptions(),
			petControlUnitSuccessOnChanged,
		}
		petControlUnitFailureSelectorDialog := SelectorDialog{
			controlUnitFailureDialog,
			selector,
			ControlUnits.GetOptions(),
			petControlUnitFailureOnChanged,
		}

		/* Param Dialogs */
		var paramDialog *dialog.CustomDialog
		humanParamOnChanged := func(s string) {
			if s != "" {
				worker.ActionState.AddHumanParam(s)
				paramDialog.Hide()
			}
		}
		petParamOnChanged := func(s string) {
			if s != "" {
				worker.ActionState.AddPetParam(s)
				paramDialog.Hide()
			}
		}
		paramSelector := widget.NewRadioGroup(nil, nil)
		paramSelector.Horizontal = true
		paramSelector.Required = true
		paramDialog = dialog.NewCustomWithoutButtons("Select param", selector, window)
		paramDialog.SetOnClosed(onClosed)

		/* Healing Dialog */
		humanHealingRatioSelectorDialog := SelectorDialog{
			paramDialog,
			selector,
			Ratios.GetOptions(),
			humanParamOnChanged,
		}

		petHealingRatioSelectorDialog := SelectorDialog{
			paramDialog,
			selector,
			Ratios.GetOptions(),
			petParamOnChanged,
		}

		/* Bomb Dialog */
		bombSelectorDialog := SelectorDialog{
			paramDialog,
			selector,
			Bombs.GetOptions(),
			humanParamOnChanged,
		}

		/* Threshold Dialog */
		humanThresholdSelectorDialog := SelectorDialog{
			paramDialog,
			selector,
			Thresholds.GetOptions(),
			humanParamOnChanged,
		}

		/* Offset Dialogs */
		var offsetDialog *dialog.CustomDialog
		humanOffsetOnChanged := func(s string) {
			if s != "" {
				offset, _ := strconv.Atoi(s)
				worker.ActionState.AddHumanSkillOffset(Offset(offset))
				offsetDialog.Hide()
			}
		}
		petOffsetOnChanged := func(s string) {
			if s != "" {
				offset, _ := strconv.Atoi(s)
				worker.ActionState.AddPetSkillOffset(Offset(offset))
				offsetDialog.Hide()
			}
		}
		offsetDialog = dialog.NewCustomWithoutButtons("Select skill offset", selector, window)
		offsetDialog.SetOnClosed(onClosed)
		humanOffsetSelectorDialog := SelectorDialog{
			offsetDialog,
			selector,
			Offsets.GetOptions(),
			humanOffsetOnChanged,
		}
		petOffsetSelectorDialog := SelectorDialog{
			offsetDialog,
			selector,
			Offsets.GetOptions(),
			petOffsetOnChanged,
		}

		/* Level Dialog */
		var levelDialog *dialog.CustomDialog
		levelOnChanged := func(s string) {
			if s != "" {
				level, _ := strconv.Atoi(s)
				worker.ActionState.AddHumanSkillLevel(Level(level))
				levelDialog.Hide()
			}
		}
		levelDialog = dialog.NewCustomWithoutButtons("Select skill level", selector, window)
		levelDialog.SetOnClosed(onClosed)
		levelSelectorDialog := SelectorDialog{
			levelDialog,
			selector,
			Levels.GetOptions(),
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

			attackButton = widget.NewButton(HumanAttack.String(), func() {
				worker.ActionState.AddHumanState(HumanAttack)

				dialogs := []SelectorDialog{
					humanControlUnitSuccessSelectorDialog,
					humanControlUnitFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			attackButton.Importance = widget.WarningImportance

			defendButton = widget.NewButton(HumanDefend.String(), func() {
				worker.ActionState.AddHumanState(HumanDefend)

				dialogs := []SelectorDialog{
					humanControlUnitSuccessSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			defendButton.Importance = widget.WarningImportance

			escapeButton = widget.NewButton(HumanEscape.String(), func() {
				worker.ActionState.AddHumanState(HumanEscape)

				dialogs := []SelectorDialog{
					humanControlUnitFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			escapeButton.Importance = widget.WarningImportance

			catchButton = widget.NewButton(HumanCatch.String(), func() {
				worker.ActionState.AddHumanState(HumanCatch)
				statesViewer.Objects = generateTags(*worker)
				statesViewer.Refresh()

				dialogs := []SelectorDialog{
					humanHealingRatioSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)

				notifyLogConfig("About Catch")
			})
			catchButton.Importance = widget.SuccessImportance

			bombButton = widget.NewButton(HumanBomb.String(), func() {
				worker.ActionState.AddHumanState(HumanBomb)

				dialogs := []SelectorDialog{
					bombSelectorDialog,
					humanControlUnitSuccessSelectorDialog,
					humanControlUnitFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			bombButton.Importance = widget.HighImportance

			potionButton = widget.NewButton(HumanPotion.String(), func() {
				worker.ActionState.AddHumanState(HumanPotion)

				dialogs := []SelectorDialog{
					humanHealingRatioSelectorDialog,
					humanControlUnitSuccessSelectorDialog,
					humanControlUnitFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			potionButton.Importance = widget.HighImportance

			recallButton = widget.NewButton(HumanRecall.String(), func() {
				worker.ActionState.AddHumanState(HumanRecall)

				dialogs := []SelectorDialog{
					humanControlUnitSuccessSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			recallButton.Importance = widget.HighImportance

			moveButton = widget.NewButton(HumanMove.String(), func() {
				worker.ActionState.AddHumanState(HumanMove)

				dialogs := []SelectorDialog{
					humanControlUnitSuccessSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			moveButton.Importance = widget.WarningImportance

			hangButton = widget.NewButton(HumanHang.String(), func() {
				worker.ActionState.AddHumanState(HumanHang)
				statesViewer.Objects = generateTags(*worker)
				statesViewer.Refresh()
			})
			hangButton.Importance = widget.SuccessImportance

			skillButton = widget.NewButton(HumanSkill.String(), func() {
				worker.ActionState.AddHumanState(HumanSkill)

				dialogs := []SelectorDialog{
					humanOffsetSelectorDialog,
					levelSelectorDialog,
					humanControlUnitSuccessSelectorDialog,
					humanControlUnitFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			skillButton.Importance = widget.HighImportance

			thresholdSkillButton = widget.NewButton(HumanThresholdSkill.String(), func() {
				worker.ActionState.AddHumanState(HumanThresholdSkill)

				dialogs := []SelectorDialog{
					humanOffsetSelectorDialog,
					levelSelectorDialog,
					humanThresholdSelectorDialog,
					humanControlUnitSuccessSelectorDialog,
					humanControlUnitFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			thresholdSkillButton.Importance = widget.HighImportance

			stealButton = widget.NewButton(HumanSteal.String(), func() {
				worker.ActionState.AddHumanState(HumanSteal)

				dialogs := []SelectorDialog{
					humanOffsetSelectorDialog,
					levelSelectorDialog,
					humanControlUnitSuccessSelectorDialog,
					humanControlUnitFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			stealButton.Importance = widget.SuccessImportance

			trainButton = widget.NewButton(HumanTrainSkill.String(), func() {
				worker.ActionState.AddHumanState(HumanTrainSkill)

				dialogs := []SelectorDialog{
					humanOffsetSelectorDialog,
					levelSelectorDialog,
					humanControlUnitSuccessSelectorDialog,
					humanControlUnitFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			trainButton.Importance = widget.SuccessImportance

			rideButton = widget.NewButton(HumanRide.String(), func() {
				worker.ActionState.AddHumanState(HumanRide)

				dialogs := []SelectorDialog{
					humanOffsetSelectorDialog,
					levelSelectorDialog,
					humanControlUnitSuccessSelectorDialog,
					humanControlUnitFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			rideButton.Importance = widget.HighImportance

			healSelfButton = widget.NewButton(HumanHealSelf.String(), func() {
				worker.ActionState.AddHumanState(HumanHealSelf)

				dialogs := []SelectorDialog{
					humanOffsetSelectorDialog,
					levelSelectorDialog,
					humanHealingRatioSelectorDialog,
					humanControlUnitSuccessSelectorDialog,
					humanControlUnitFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			healSelfButton.Importance = widget.HighImportance

			healOneButton = widget.NewButton(HumanHealOne.String(), func() {
				worker.ActionState.AddHumanState(HumanHealOne)

				dialogs := []SelectorDialog{
					humanOffsetSelectorDialog,
					levelSelectorDialog,
					humanHealingRatioSelectorDialog,
					humanControlUnitSuccessSelectorDialog,
					humanControlUnitFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			healOneButton.Importance = widget.HighImportance

			healTShapeButton = widget.NewButton(HumanHealTShaped.String(), func() {
				worker.ActionState.AddHumanState(HumanHealTShaped)

				dialogs := []SelectorDialog{
					humanOffsetSelectorDialog,
					levelSelectorDialog,
					humanHealingRatioSelectorDialog,
					humanControlUnitSuccessSelectorDialog,
					humanControlUnitFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			healTShapeButton.Importance = widget.HighImportance

			healMultiButton = widget.NewButton(HumanHealMulti.String(), func() {
				worker.ActionState.AddHumanState(HumanHealMulti)

				dialogs := []SelectorDialog{
					humanOffsetSelectorDialog,
					levelSelectorDialog,
					humanHealingRatioSelectorDialog,
					humanControlUnitSuccessSelectorDialog,
					humanControlUnitFailureSelectorDialog,
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

			petAttackButton = widget.NewButton(PetAttack.String(), func() {
				worker.ActionState.AddPetState(PetAttack)

				dialogs := []SelectorDialog{
					petControlUnitSuccessSelectorDialog,
					petControlUnitFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			petAttackButton.Importance = widget.WarningImportance

			petHangButton = widget.NewButton(PetHang.String(), func() {
				worker.ActionState.AddPetState(PetHang)
				statesViewer.Objects = generateTags(*worker)
				statesViewer.Refresh()
			})
			petHangButton.Importance = widget.SuccessImportance

			petSkillButton = widget.NewButton(PetSkill.String(), func() {
				worker.ActionState.AddPetState(PetSkill)

				dialogs := []SelectorDialog{
					petOffsetSelectorDialog,
					petControlUnitSuccessSelectorDialog,
					petControlUnitFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			petSkillButton.Importance = widget.HighImportance

			petDefendButton = widget.NewButton(PetDefend.String(), func() {
				worker.ActionState.AddPetState(PetDefend)

				dialogs := []SelectorDialog{
					petOffsetSelectorDialog,
					petControlUnitSuccessSelectorDialog,
					petControlUnitFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			petDefendButton.Importance = widget.HighImportance

			petHealSelfButton = widget.NewButton(PetHealSelf.String(), func() {
				worker.ActionState.AddPetState(PetHealSelf)

				dialogs := []SelectorDialog{
					petOffsetSelectorDialog,
					petHealingRatioSelectorDialog,
					petControlUnitSuccessSelectorDialog,
					petControlUnitFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			petHealSelfButton.Importance = widget.HighImportance

			petHealOneButton = widget.NewButton(PetHealOne.String(), func() {
				worker.ActionState.AddPetState(PetHealOne)

				dialogs := []SelectorDialog{
					petOffsetSelectorDialog,
					petHealingRatioSelectorDialog,
					petControlUnitSuccessSelectorDialog,
					petControlUnitFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			petHealOneButton.Importance = widget.HighImportance

			petRideButton = widget.NewButton(PetRide.String(), func() {
				worker.ActionState.AddPetState(PetRide)

				dialogs := []SelectorDialog{
					petOffsetSelectorDialog,
					petControlUnitSuccessSelectorDialog,
					petControlUnitFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			petRideButton.Importance = widget.HighImportance

			petOffRideButton = widget.NewButton(PetOffRide.String(), func() {
				worker.ActionState.AddPetState(PetOffRide)

				dialogs := []SelectorDialog{
					petOffsetSelectorDialog,
					petControlUnitSuccessSelectorDialog,
					petControlUnitFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			petOffRideButton.Importance = widget.HighImportance

			petEscapeButton = widget.NewButton(PetEscape.String(), func() {
				worker.ActionState.AddPetState(PetEscape)

				dialogs := []SelectorDialog{
					petControlUnitFailureSelectorDialog,
				}
				activateDialogs(dialogs, enableChan)
			})
			petEscapeButton.Importance = widget.WarningImportance

			petCatchButton = widget.NewButton(PetCatch.String(), func() {
				worker.ActionState.AddPetState(PetCatch)
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
								worker.ActionState.ManaChecker = &manaChecker
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
	tagContaines = createTagContainers(worker.ActionState, Human)
	tagContaines = append(tagContaines, createTagContainers(worker.ActionState, Pet)...)
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

func createTagContainers(actionState BattleActionState, role Role) (tagContainers []fyne.CanvasObject) {
	var tag string
	var anyStates []any
	if role == Human {
		for _, hs := range actionState.GetHumanStates() {
			anyStates = append(anyStates, hs)
		}
	} else {
		for _, ps := range actionState.GetPetStates() {
			anyStates = append(anyStates, ps)
		}
	}

	for _, state := range anyStates {
		tagColor := petSpecialTagColor
		if role == Human {
			tag = state.(HumanState).Action.String()
			switch {
			case strings.Contains(state.(HumanState).Action.String(), "**"):
				tagColor = humanFinishingTagColor
			case strings.Contains(state.(HumanState).Action.String(), "*"):
				tagColor = humanConditionalTagColor
			default:
				tagColor = humanSpecialTagColor
			}
		} else {
			tag = state.(PetState).Action.String()
			switch {
			case strings.Contains(state.(PetState).Action.String(), "**"):
				tagColor = petFinishingTagColor
			case strings.Contains(state.(PetState).Action.String(), "*"):
				tagColor = petConditionalTagColor
			default:
				tagColor = petSpecialTagColor
			}
		}

		tagCanvas := canvas.NewRectangle(tagColor)
		tagCanvas.SetMinSize(fyne.NewSize(60, 22))
		tagContainer := container.NewStack(tagCanvas)

		if role == Human {
			if offset := state.(HumanState).Offset; offset != 0 {
				tag = fmt.Sprintf("%s:%s:%d", tag, offset, state.(HumanState).Level)
			}
		} else {
			if offset := state.(PetState).Offset; offset != 0 {
				tag = fmt.Sprintf("%s:%s", tag, offset)
			}
		}

		var param string
		var successControlUnit string
		var successJumpId int
		var failureControlUnit string
		var failureJumpId int

		if role == Human {
			param = state.(HumanState).Param
			successControlUnit = string(state.(HumanState).SuccessControlUnit)
			failureControlUnit = string(state.(HumanState).FailureControlUnit)
			successJumpId = state.(HumanState).SuccessJumpId
			failureJumpId = state.(HumanState).FailureJumpId
		} else {
			param = state.(PetState).Param
			successControlUnit = string(state.(PetState).SuccessControlUnit)
			failureControlUnit = string(state.(PetState).FailureControlUnit)
			successJumpId = state.(PetState).SuccessJumpId
			failureJumpId = state.(PetState).FailureJumpId
		}

		if param != "" {
			tag = fmt.Sprintf("%s:%s", tag, param)
		}

		if len(successControlUnit) > 0 {
			controlUnitFirstLetter := strings.ToLower(successControlUnit[:1])
			tag = fmt.Sprintf("%s:%s", tag, controlUnitFirstLetter)
			if controlUnitFirstLetter == "j" {
				tag = fmt.Sprintf("%s%s", tag, successJumpId)
			}
		} else {
			tag = tag + ":"
		}

		if len(failureControlUnit) > 0 {
			controlUnitFirstLetter := strings.ToLower(failureControlUnit)
			tag = fmt.Sprintf("%s:%s", tag, controlUnitFirstLetter)
			if controlUnitFirstLetter == "j" {
				tag = fmt.Sprintf("%s%s", tag, failureJumpId)
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

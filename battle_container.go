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

func newBattleContainer(games Games) (*fyne.Container, BattleGroups) {
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

	logCheckersButton := widget.NewButtonWithIcon("Log Checkers", theme.MenuIcon(), func() {
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

		var aliasButton *widget.Button
		aliasButton = widget.NewButtonWithIcon(allGames.FindKey(worker.GetHandle()), theme.AccountIcon(), func() {
			aliasEntry := widget.NewEntry()
			aliasEntry.SetPlaceHolder("Enter alias")

			aliasDialog := dialog.NewCustom("Enter alias", "Ok", aliasEntry, window)
			aliasDialog.SetOnClosed(func() {
				if _, ok := allGames[aliasEntry.Text]; aliasEntry.Text == "" || ok {
					return
				}

				allGames.RemoveValue(worker.GetHandle())
				allGames.Add(aliasEntry.Text, worker.GetHandle())
				aliasButton.SetText(aliasEntry.Text)
			})
			aliasDialog.Show()
		})

		var movementModeButton *widget.Button
		var movementModeDialog *dialog.CustomDialog
		movementModeSelector := widget.NewRadioGroup(BATTLE_MOVEMENT_MODES.GetOptions(), func(s string) {
			if s != "" {
				worker.MovementState.Mode = BattleMovementMode(s)
				if worker.MovementState.Mode != None {
					movementModeButton.SetText(s)
				} else {
					movementModeButton.SetText("Move Way")
				}
			}
			movementModeDialog.Hide()
		})
		movementModeDialog = dialog.NewCustomWithoutButtons("Select a move way", movementModeSelector, window)
		movementModeButton = widget.NewButtonWithIcon("Move Way", theme.MailReplyIcon(), func() {
			movementModeDialog.Show()
		})

		var selector = widget.NewRadioGroup(nil, nil)
		selector.Horizontal = true
		selector.Required = true
		selectorDialogEnableChan := make(chan bool)
		getNewSelectorDialog := SelectorDialoger(selector)

		var actionsViewer *fyne.Container
		onClosed := func() {
			actionsViewer.Objects = generateTags(*worker)
			actionsViewer.Refresh()
			selectorDialogEnableChan <- true
		}

		/* Control Unit Dialogs */
		var successControlUnitDialog *dialog.CustomDialog
		var failureControlUnitDialog *dialog.CustomDialog

		activateJumpDialog := func(totalActions int, callback func(jumpId int)) {
			jumpEntry := widget.NewEntry()
			jumpEntry.Validator = func(jumpIdStr string) error {
				if jumpId, err := strconv.Atoi(jumpIdStr); err != nil {
					return err
				} else if jumpId >= totalActions-1 || jumpId < 1 {
					return errors.New("not a valid offset")
				}
				return nil
			}

			jumpDialog := dialog.NewForm("Enter next action id", "Ok", "Dismiss", []*widget.FormItem{widget.NewFormItem("Action Id", jumpEntry)}, func(isValid bool) {
				if isValid {
					jumpId, _ := strconv.Atoi(jumpEntry.Text)
					callback(jumpId)
				}
			}, window)
			jumpDialog.Show()
		}
		successControlUnitOnChanged := func(r Role) func(s string) {
			return func(s string) {
				if s == "" {
					return
				}

				if cu := ControlUnit(s); cu == Jump {
					activateJumpDialog(len(worker.ActionState.HumanActions), func(jumpId int) {
						worker.ActionState.AddSuccessControlUnit(r, cu)
						worker.ActionState.AddSuccessJumpId(r, jumpId)
						successControlUnitDialog.Hide()
					})
				} else {
					worker.ActionState.AddSuccessControlUnit(r, cu)
					successControlUnitDialog.Hide()
				}
			}
		}
		failureControlUnitOnChanged := func(r Role) func(s string) {
			return func(s string) {
				if s == "" {
					return
				}

				if cu := ControlUnit(s); cu == Jump {
					activateJumpDialog(len(worker.ActionState.HumanActions), func(jumpId int) {
						worker.ActionState.AddFailureControlUnit(r, cu)
						worker.ActionState.AddFailureJumpId(r, jumpId)
						failureControlUnitDialog.Hide()
					})
				} else {
					worker.ActionState.AddFailureControlUnit(r, cu)
					failureControlUnitDialog.Hide()
				}
			}
		}
		controlUnitOnClosed := func() {
			actionsViewer.Objects = generateTags(*worker)
			actionsViewer.Refresh()
			selectorDialogEnableChan <- true
		}
		successControlUnitDialog = dialog.NewCustomWithoutButtons("Select next action after successful execution", selector, window)
		successControlUnitDialog.SetOnClosed(controlUnitOnClosed)
		failureControlUnitDialog = dialog.NewCustomWithoutButtons("Select next action after failed execution", selector, window)
		failureControlUnitDialog.SetOnClosed(controlUnitOnClosed)

		successControlUnitSelectorDialog := getNewSelectorDialog(successControlUnitDialog, ControlUnits.GetOptions(), successControlUnitOnChanged)
		failureControlUnitSelectorDialog := getNewSelectorDialog(failureControlUnitDialog, ControlUnits.GetOptions(), failureControlUnitOnChanged)

		/* Param Dialogs */
		var paramDialog *dialog.CustomDialog
		paramOnChanged := func(r Role) func(s string) {
			return func(s string) {
				if s != "" {
					worker.ActionState.AddParam(r, s)
					paramDialog.Hide()
				}
			}
		}
		paramDialog = dialog.NewCustomWithoutButtons("Select param", selector, window)
		paramDialog.SetOnClosed(onClosed)

		healingRatioSelectorDialog := getNewSelectorDialog(paramDialog, Ratios.GetOptions(), paramOnChanged)
		bombSelectorDialog := getNewSelectorDialog(paramDialog, Bombs.GetOptions(), paramOnChanged)

		/* Threshold Dialog */
		var thresholdDialog *dialog.CustomDialog
		thresholdOnChanged := func(r Role) func(s string) {
			return func(s string) {
				if s != "" {
					worker.ActionState.AddThreshold(r, Threshold(s))
					thresholdDialog.Hide()
				}
			}
		}
		thresholdDialog = dialog.NewCustomWithoutButtons("Select threshold", selector, window)
		thresholdDialog.SetOnClosed(onClosed)

		thresholdSelectorDialog := getNewSelectorDialog(thresholdDialog, Thresholds.GetOptions(), thresholdOnChanged)

		/* Offset Dialog */
		var offsetDialog *dialog.CustomDialog
		offsetOnChanged := func(r Role) func(s string) {
			return func(s string) {
				if s != "" {
					offset, _ := strconv.Atoi(s)
					worker.ActionState.AddSkillOffset(r, Offset(offset))
					offsetDialog.Hide()
				}
			}
		}
		offsetDialog = dialog.NewCustomWithoutButtons("Select skill offset", selector, window)
		offsetDialog.SetOnClosed(onClosed)

		offsetSelectorDialog := getNewSelectorDialog(offsetDialog, Offsets.GetOptions(), offsetOnChanged)

		/* Level Dialog */
		var levelDialog *dialog.CustomDialog
		levelOnChanged := func(r Role) func(s string) {
			return func(s string) {
				if s != "" {
					level, _ := strconv.Atoi(s)
					worker.ActionState.AddHumanSkillLevel(Offset(level))
					levelDialog.Hide()
				}
			}
		}
		levelDialog = dialog.NewCustomWithoutButtons("Select skill level", selector, window)
		levelDialog.SetOnClosed(onClosed)

		levelSelectorDialog := getNewSelectorDialog(levelDialog, Levels.GetOptions(), levelOnChanged)

		/* Human Actions */
		humanActionSelector := widget.NewButtonWithIcon("Man Actions", theme.ContentAddIcon(), func() {
			worker.ActionState.ClearHumanActions()
			actionsViewer.Objects = generateTags(*worker)
			actionsViewer.Refresh()

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
				worker.ActionState.AddHumanAction(HumanAttack)

				dialogs := []func(){
					successControlUnitSelectorDialog(Human),
					failureControlUnitSelectorDialog(Human),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			attackButton.Importance = widget.WarningImportance

			defendButton = widget.NewButton(HumanDefend.String(), func() {
				worker.ActionState.AddHumanAction(HumanDefend)

				dialogs := []func(){
					successControlUnitSelectorDialog(Human),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			defendButton.Importance = widget.WarningImportance

			escapeButton = widget.NewButton(HumanEscape.String(), func() {
				worker.ActionState.AddHumanAction(HumanEscape)

				dialogs := []func(){
					failureControlUnitSelectorDialog(Human),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			escapeButton.Importance = widget.WarningImportance

			catchButton = widget.NewButton(HumanCatch.String(), func() {
				worker.ActionState.AddHumanAction(HumanCatch)
				actionsViewer.Objects = generateTags(*worker)
				actionsViewer.Refresh()

				dialogs := []func(){
					healingRatioSelectorDialog(Human),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)

				notifyLogConfig("About Catch")
			})
			catchButton.Importance = widget.SuccessImportance

			bombButton = widget.NewButton(HumanBomb.String(), func() {
				worker.ActionState.AddHumanAction(HumanBomb)

				dialogs := []func(){
					bombSelectorDialog(Human),
					successControlUnitSelectorDialog(Human),
					failureControlUnitSelectorDialog(Human),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			bombButton.Importance = widget.HighImportance

			potionButton = widget.NewButton(HumanPotion.String(), func() {
				worker.ActionState.AddHumanAction(HumanPotion)

				dialogs := []func(){
					healingRatioSelectorDialog(Human),
					successControlUnitSelectorDialog(Human),
					failureControlUnitSelectorDialog(Human),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			potionButton.Importance = widget.HighImportance

			recallButton = widget.NewButton(HumanRecall.String(), func() {
				worker.ActionState.AddHumanAction(HumanRecall)

				dialogs := []func(){
					successControlUnitSelectorDialog(Human),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			recallButton.Importance = widget.HighImportance

			moveButton = widget.NewButton(HumanMove.String(), func() {
				worker.ActionState.AddHumanAction(HumanMove)

				dialogs := []func(){
					successControlUnitSelectorDialog(Human),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			moveButton.Importance = widget.WarningImportance

			hangButton = widget.NewButton(HumanHang.String(), func() {
				worker.ActionState.AddHumanAction(HumanHang)
				actionsViewer.Objects = generateTags(*worker)
				actionsViewer.Refresh()
			})
			hangButton.Importance = widget.SuccessImportance

			skillButton = widget.NewButton(HumanSkill.String(), func() {
				worker.ActionState.AddHumanAction(HumanSkill)

				dialogs := []func(){
					offsetSelectorDialog(Human),
					levelSelectorDialog(Human),
					successControlUnitSelectorDialog(Human),
					failureControlUnitSelectorDialog(Human),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			skillButton.Importance = widget.HighImportance

			thresholdSkillButton = widget.NewButton(HumanThresholdSkill.String(), func() {
				worker.ActionState.AddHumanAction(HumanThresholdSkill)

				dialogs := []func(){
					offsetSelectorDialog(Human),
					levelSelectorDialog(Human),
					thresholdSelectorDialog(Human),
					successControlUnitSelectorDialog(Human),
					failureControlUnitSelectorDialog(Human),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			thresholdSkillButton.Importance = widget.HighImportance

			stealButton = widget.NewButton(HumanSteal.String(), func() {
				worker.ActionState.AddHumanAction(HumanSteal)

				dialogs := []func(){
					offsetSelectorDialog(Human),
					levelSelectorDialog(Human),
					successControlUnitSelectorDialog(Human),
					failureControlUnitSelectorDialog(Human),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			stealButton.Importance = widget.SuccessImportance

			trainButton = widget.NewButton(HumanTrainSkill.String(), func() {
				worker.ActionState.AddHumanAction(HumanTrainSkill)

				dialogs := []func(){
					offsetSelectorDialog(Human),
					levelSelectorDialog(Human),
					successControlUnitSelectorDialog(Human),
					failureControlUnitSelectorDialog(Human),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			trainButton.Importance = widget.SuccessImportance

			rideButton = widget.NewButton(HumanRide.String(), func() {
				worker.ActionState.AddHumanAction(HumanRide)

				dialogs := []func(){
					offsetSelectorDialog(Human),
					levelSelectorDialog(Human),
					successControlUnitSelectorDialog(Human),
					failureControlUnitSelectorDialog(Human),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			rideButton.Importance = widget.HighImportance

			healSelfButton = widget.NewButton(HumanHealSelf.String(), func() {
				worker.ActionState.AddHumanAction(HumanHealSelf)

				dialogs := []func(){
					offsetSelectorDialog(Human),
					levelSelectorDialog(Human),
					healingRatioSelectorDialog(Human),
					successControlUnitSelectorDialog(Human),
					failureControlUnitSelectorDialog(Human),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			healSelfButton.Importance = widget.HighImportance

			healOneButton = widget.NewButton(HumanHealOne.String(), func() {
				worker.ActionState.AddHumanAction(HumanHealOne)

				dialogs := []func(){
					offsetSelectorDialog(Human),
					levelSelectorDialog(Human),
					healingRatioSelectorDialog(Human),
					successControlUnitSelectorDialog(Human),
					failureControlUnitSelectorDialog(Human),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			healOneButton.Importance = widget.HighImportance

			healTShapeButton = widget.NewButton(HumanHealTShaped.String(), func() {
				worker.ActionState.AddHumanAction(HumanHealTShaped)

				dialogs := []func(){
					offsetSelectorDialog(Human),
					levelSelectorDialog(Human),
					healingRatioSelectorDialog(Human),
					successControlUnitSelectorDialog(Human),
					failureControlUnitSelectorDialog(Human),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			healTShapeButton.Importance = widget.HighImportance

			healMultiButton = widget.NewButton(HumanHealMulti.String(), func() {
				worker.ActionState.AddHumanAction(HumanHealMulti)

				dialogs := []func(){
					offsetSelectorDialog(Human),
					levelSelectorDialog(Human),
					healingRatioSelectorDialog(Human),
					successControlUnitSelectorDialog(Human),
					failureControlUnitSelectorDialog(Human),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			healMultiButton.Importance = widget.HighImportance

			actionsContainer := container.NewGridWithColumns(4,
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

			actionsDialog := dialog.NewCustom("Select man actions with order", "Leave", actionsContainer, window)
			actionsDialog.Show()
		})

		/* Pet Actions */
		petActionSelector := widget.NewButtonWithIcon("Pet Actions", theme.ContentAddIcon(), func() {
			worker.ActionState.ClearPetActions()
			actionsViewer.Objects = generateTags(*worker)
			actionsViewer.Refresh()

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
				worker.ActionState.AddPetAction(PetAttack)

				dialogs := []func(){
					successControlUnitSelectorDialog(Pet),
					failureControlUnitSelectorDialog(Pet),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			petAttackButton.Importance = widget.WarningImportance

			petHangButton = widget.NewButton(PetHang.String(), func() {
				worker.ActionState.AddPetAction(PetHang)
				actionsViewer.Objects = generateTags(*worker)
				actionsViewer.Refresh()
			})
			petHangButton.Importance = widget.SuccessImportance

			petSkillButton = widget.NewButton(PetSkill.String(), func() {
				worker.ActionState.AddPetAction(PetSkill)

				dialogs := []func(){
					offsetSelectorDialog(Pet),
					successControlUnitSelectorDialog(Pet),
					failureControlUnitSelectorDialog(Pet),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			petSkillButton.Importance = widget.HighImportance

			petDefendButton = widget.NewButton(PetDefend.String(), func() {
				worker.ActionState.AddPetAction(PetDefend)

				dialogs := []func(){
					offsetSelectorDialog(Pet),
					successControlUnitSelectorDialog(Pet),
					failureControlUnitSelectorDialog(Pet),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			petDefendButton.Importance = widget.HighImportance

			petHealSelfButton = widget.NewButton(PetHealSelf.String(), func() {
				worker.ActionState.AddPetAction(PetHealSelf)

				dialogs := []func(){
					offsetSelectorDialog(Pet),
					healingRatioSelectorDialog(Pet),
					successControlUnitSelectorDialog(Pet),
					failureControlUnitSelectorDialog(Pet),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			petHealSelfButton.Importance = widget.HighImportance

			petHealOneButton = widget.NewButton(PetHealOne.String(), func() {
				worker.ActionState.AddPetAction(PetHealOne)

				dialogs := []func(){
					offsetSelectorDialog(Pet),
					healingRatioSelectorDialog(Pet),
					successControlUnitSelectorDialog(Pet),
					failureControlUnitSelectorDialog(Pet),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			petHealOneButton.Importance = widget.HighImportance

			petRideButton = widget.NewButton(PetRide.String(), func() {
				worker.ActionState.AddPetAction(PetRide)

				dialogs := []func(){
					offsetSelectorDialog(Pet),
					successControlUnitSelectorDialog(Pet),
					failureControlUnitSelectorDialog(Pet),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			petRideButton.Importance = widget.HighImportance

			petOffRideButton = widget.NewButton(PetOffRide.String(), func() {
				worker.ActionState.AddPetAction(PetOffRide)

				dialogs := []func(){
					offsetSelectorDialog(Pet),
					successControlUnitSelectorDialog(Pet),
					failureControlUnitSelectorDialog(Pet),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			petOffRideButton.Importance = widget.HighImportance

			petEscapeButton = widget.NewButton(PetEscape.String(), func() {
				worker.ActionState.AddPetAction(PetEscape)

				dialogs := []func(){
					failureControlUnitSelectorDialog(Pet),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			petEscapeButton.Importance = widget.WarningImportance

			petCatchButton = widget.NewButton(PetCatch.String(), func() {
				worker.ActionState.AddPetAction(PetCatch)
				actionsViewer.Objects = generateTags(*worker)
				actionsViewer.Refresh()

				notifyLogConfig("About Catch")

				dialogs := []func(){
					healingRatioSelectorDialog(Pet),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			petCatchButton.Importance = widget.SuccessImportance

			actionsContainer := container.NewGridWithColumns(4,
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

			actionsDialog := dialog.NewCustom("Select pet actions with order", "Leave", actionsContainer, window)
			actionsDialog.Show()

		})
		humanActionSelector.Importance = widget.MediumImportance
		petActionSelector.Importance = widget.MediumImportance

		loadSettingButton := widget.NewButtonWithIcon("Load", theme.FolderOpenIcon(), func() {
			fileOpenDialog := dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
				if uc != nil {
					var actionState BattleActionState
					file, openErr := os.Open(uc.URI().Path())

					if openErr == nil {
						defer file.Close()
						if buffer, readErr := io.ReadAll(file); readErr == nil {
							if json.Unmarshal(buffer, &actionState) == nil {
								actionState.SetHWND(worker.ActionState.GetHWND())
								worker.ActionState = actionState
								worker.ActionState.LogDir = logDir
								worker.ActionState.ManaChecker = &manaChecker
								actionsViewer.Objects = generateTags(*worker)
								actionsViewer.Refresh()
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

		workerMenuContainer.Add(aliasButton)
		workerMenuContainer.Add(movementModeButton)
		workerMenuContainer.Add(humanActionSelector)
		workerMenuContainer.Add(petActionSelector)
		workerMenuContainer.Add(loadSettingButton)
		workerMenuContainer.Add(saveSettingButton)

		actionsViewer = container.NewAdaptiveGrid(6, generateTags(*worker)...)

		workerContainer := container.NewVBox(workerMenuContainer, actionsViewer)
		configContainer.Add(workerContainer)
	}

	autoBattleWidget = container.NewVBox(widget.NewSeparator(), mainWidget, widget.NewSeparator(), configContainer)
	return autoBattleWidget, sharedStopChan
}

func SelectorDialoger(selector *widget.RadioGroup) func(dialog *dialog.CustomDialog, options []string, onChanged func(r Role) func(s string)) func(r Role) func() {
	return func(dialog *dialog.CustomDialog, options []string, onChanged func(r Role) func(s string)) func(r Role) func() {
		return func(r Role) func() {
			return func() {
				selector.Disable()
				selector.Options = options
				selector.OnChanged = onChanged(r)
				selector.Selected = ""
				selector.Enable()
				dialog.Show()
			}
		}
	}
}

func activateDialogs(selectorDialogs []func(), enableChan chan bool) {

	go func() {
		for i, selectorDialog := range selectorDialogs {
			<-enableChan
			selectorDialog()

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
	var anyActions []any
	if role == Human {
		for _, hs := range actionState.GetHumanActions() {
			anyActions = append(anyActions, hs)
		}
	} else {
		for _, ps := range actionState.GetPetActions() {
			anyActions = append(anyActions, ps)
		}
	}

	for _, action := range anyActions {
		var tagColor color.RGBA
		if role == Human {
			tag = action.(HumanAction).Action.String()
			switch {
			case strings.Contains(action.(HumanAction).Action.String(), "**"):
				tagColor = humanFinishingTagColor
			case strings.Contains(action.(HumanAction).Action.String(), "*"):
				tagColor = humanConditionalTagColor
			default:
				tagColor = humanSpecialTagColor
			}
		} else {
			tag = action.(PetAction).Action.String()
			switch {
			case strings.Contains(action.(PetAction).Action.String(), "**"):
				tagColor = petFinishingTagColor
			case strings.Contains(action.(PetAction).Action.String(), "*"):
				tagColor = petConditionalTagColor
			default:
				tagColor = petSpecialTagColor
			}
		}

		tagCanvas := canvas.NewRectangle(tagColor)
		tagCanvas.SetMinSize(fyne.NewSize(60, 22))
		tagContainer := container.NewStack(tagCanvas)

		if role == Human {
			if offset := action.(HumanAction).Offset; offset != 0 {
				tag = fmt.Sprintf("%s:%d:%d", tag, offset, action.(HumanAction).Level)
			}
		} else {
			if offset := action.(PetAction).Offset; offset != 0 {
				tag = fmt.Sprintf("%s:%d", tag, offset)
			}
		}

		var param string
		var threshold Threshold
		var successControlUnit string
		var successJumpId int
		var failureControlUnit string
		var failureJumpId int

		if role == Human {
			param = action.(HumanAction).Param
			threshold = action.(HumanAction).Threshold
			successControlUnit = string(action.(HumanAction).SuccessControlUnit)
			failureControlUnit = string(action.(HumanAction).FailureControlUnit)
			successJumpId = action.(HumanAction).SuccessJumpId
			failureJumpId = action.(HumanAction).FailureJumpId
		} else {
			param = action.(PetAction).Param
			threshold = action.(PetAction).Threshold
			successControlUnit = string(action.(PetAction).SuccessControlUnit)
			failureControlUnit = string(action.(PetAction).FailureControlUnit)
			successJumpId = action.(PetAction).SuccessJumpId
			failureJumpId = action.(PetAction).FailureJumpId
		}

		if param != "" {
			tag = fmt.Sprintf("%s:%s", tag, param)
		}

		if threshold != "" {
			tag = fmt.Sprintf("%s:%s", tag, threshold)
		}

		if len(successControlUnit) > 0 {
			controlUnitFirstLetter := strings.ToLower(successControlUnit[:1])
			tag = fmt.Sprintf("%s:%s", tag, controlUnitFirstLetter)
			if controlUnitFirstLetter == "j" {
				tag = fmt.Sprintf("%s%d", tag, successJumpId)
			}
		} else {
			tag = tag + ":"
		}

		if len(failureControlUnit) > 0 {
			controlUnitFirstLetter := strings.ToLower(failureControlUnit[:1])
			tag = fmt.Sprintf("%s:%s", tag, controlUnitFirstLetter)
			if controlUnitFirstLetter == "j" {
				tag = fmt.Sprintf("%s%d", tag, failureJumpId)
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

func (bgs *BattleGroups) stopAll() {
	for k := range bgs.stopChans {
		stop(bgs.stopChans[k])
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

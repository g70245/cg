package container

import (
	"cg/game"
	"cg/game/battle"
	"cg/game/enum/character"
	"cg/game/enum/controlunit"
	"cg/game/enum/enemyorder"
	"cg/game/enum/movement"
	"cg/game/enum/offset"
	"cg/game/enum/pet"
	"cg/game/enum/role"
	"cg/game/enum/threshold"
	"cg/game/items"
	"cg/utils"
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

func newBattleContainer(games game.Games) (*fyne.Container, BattleGroups) {
	id := 0
	battleGroups := BattleGroups{make(map[int]chan bool)}

	groupTabs := container.NewAppTabs()
	groupTabs.SetTabLocation(container.TabLocationTop)
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
					delete(battleGroups.stopChans, id)

					groupTabs.Remove(newTabItem)
					if len(battleGroups.stopChans) == 0 {
						groupTabs.Hide()
					}

					window.SetContent(window.Content())
					window.Resize(fyne.NewSize(r.width, r.height))
				}
			}(id))
			battleGroups.stopChans[id] = stopChan

			var newGroupName string
			if groupNameEntry.Text != "" {
				newGroupName = groupNameEntry.Text
			} else {
				newGroupName = "Group " + fmt.Sprint(id)
			}
			newTabItem = container.NewTabItem(newGroupName, newGroupContainer)
			groupTabs.Append(newTabItem)
			groupTabs.Select(newTabItem)
			groupTabs.Show()

			window.SetContent(window.Content())
			window.Resize(fyne.NewSize(r.width, r.height))
			id++
		})
		gamesSelectorDialog.Show()
	})

	newBattleContainer := container.NewBorder(nil, newGroupButton, nil, nil, groupTabs)

	return newBattleContainer, battleGroups
}

func newBatttleGroupContainer(games game.Games, allGames game.Games, destroy func()) (autoBattleWidget *fyne.Container, sharedStopChan chan bool) {
	manaChecker := battle.NO_MANA_CHECKER
	sharedStopChan = make(chan bool, len(games))
	workers := battle.CreateWorkers(games, r.gameDir, &manaChecker, new(bool), sharedStopChan, new(sync.WaitGroup))

	gameWidget, actionViewers := generateGameWidget(gameWidgeOptions{
		games:       games,
		allGames:    allGames,
		manaChecker: &manaChecker,
		workers:     workers,
	})
	menuWidget := generateMenuWidget(menuWidgetOptions{
		games:          games,
		allGames:       allGames,
		manaChecker:    &manaChecker,
		workers:        workers,
		sharedStopChan: sharedStopChan,
		actionViewers:  actionViewers,
		destroy:        destroy,
	})

	autoBattleWidget = container.NewVBox(widget.NewSeparator(), menuWidget, widget.NewSeparator(), gameWidget)
	return autoBattleWidget, sharedStopChan
}

type gameWidgeOptions struct {
	games       game.Games
	allGames    game.Games
	manaChecker *string
	workers     battle.Workers
}

func generateGameWidget(options gameWidgeOptions) (gameWidget *fyne.Container, actionViewers []*fyne.Container) {
	gameWidget = container.NewVBox()
	actionViewers = []*fyne.Container{}

	for i := range options.workers {
		workerMenuContainer := container.NewGridWithColumns(6)
		worker := &options.workers[i]

		var aliasButton *widget.Button
		aliasButton = widget.NewButtonWithIcon(options.allGames.FindKey(worker.GetHandle()), theme.AccountIcon(), func() {
			aliasEntry := widget.NewEntry()
			aliasEntry.SetPlaceHolder("Enter alias")

			aliasDialog := dialog.NewCustom("Enter alias", "Ok", aliasEntry, window)
			aliasDialog.SetOnClosed(func() {
				if _, ok := options.allGames[aliasEntry.Text]; aliasEntry.Text == "" || ok {
					return
				}

				options.allGames.RemoveValue(worker.GetHandle())
				options.allGames.Add(aliasEntry.Text, worker.GetHandle())
				aliasButton.SetText(aliasEntry.Text)
			})
			aliasDialog.Show()
		})

		var movementModeButton *widget.Button
		var movementModeDialog *dialog.CustomDialog
		movementModeSelector := widget.NewRadioGroup(battle.MovementModes.GetOptions(), func(s string) {
			if s != "" {
				worker.MovementState.Mode = movement.Mode(s)
				if worker.MovementState.Mode != movement.None {
					movementModeButton.SetText(s)
				} else {
					movementModeButton.SetText("Move Way")
				}
			}
			movementModeDialog.Hide()
		})
		movementModeSelector.Required = true
		movementModeDialog = dialog.NewCustomWithoutButtons("Select a move way", movementModeSelector, window)
		movementModeButton = widget.NewButtonWithIcon("Move Way", theme.MailReplyIcon(), func() {
			movementModeDialog.Show()
		})

		var selector = widget.NewRadioGroup(nil, nil)
		selector.Horizontal = true
		selector.Required = true
		selectorDialogEnableChan := make(chan bool)
		getNewSelectorDialog := selectorDialoger(selector)

		var actionViewer *fyne.Container
		onClosed := func() {
			actionViewer.Objects = generateTags(*worker)
			actionViewer.Refresh()
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
		successControlUnitOnChanged := func(r role.Role) func(s string) {
			return func(s string) {
				if s == "" {
					return
				}

				if cu := controlunit.ControlUnit(s); cu == controlunit.Jump {
					activateJumpDialog(len(worker.ActionState.CharacterActions), func(jumpId int) {
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
		failureControlUnitOnChanged := func(r role.Role) func(s string) {
			return func(s string) {
				if s == "" {
					return
				}

				if cu := controlunit.ControlUnit(s); cu == controlunit.Jump {
					activateJumpDialog(len(worker.ActionState.CharacterActions), func(jumpId int) {
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
			actionViewer.Objects = generateTags(*worker)
			actionViewer.Refresh()
			selectorDialogEnableChan <- true
		}
		successControlUnitDialog = dialog.NewCustomWithoutButtons("Select next action after successful execution", selector, window)
		successControlUnitDialog.SetOnClosed(controlUnitOnClosed)
		failureControlUnitDialog = dialog.NewCustomWithoutButtons("Select next action after failed execution", selector, window)
		failureControlUnitDialog.SetOnClosed(controlUnitOnClosed)

		successControlUnitSelectorDialog := getNewSelectorDialog(successControlUnitDialog, battle.ControlUnits.GetOptions(), successControlUnitOnChanged)
		failureControlUnitSelectorDialog := getNewSelectorDialog(failureControlUnitDialog, battle.ControlUnits.GetOptions(), failureControlUnitOnChanged)

		/* Param Dialogs */
		var paramDialog *dialog.CustomDialog
		paramOnChanged := func(r role.Role) func(s string) {
			return func(s string) {
				if s != "" {
					worker.ActionState.AddParam(r, s)
					paramDialog.Hide()
				}
			}
		}
		paramDialog = dialog.NewCustomWithoutButtons("Select param", selector, window)
		paramDialog.SetOnClosed(onClosed)

		healingRatioSelectorDialog := getNewSelectorDialog(paramDialog, battle.Ratios.GetOptions(), paramOnChanged)
		bombSelectorDialog := getNewSelectorDialog(paramDialog, items.Bombs.GetOptions(), paramOnChanged)

		/* Threshold Dialog */
		var thresholdDialog *dialog.CustomDialog
		thresholdOnChanged := func(r role.Role) func(s string) {
			return func(s string) {
				if s != "" {
					worker.ActionState.AddThreshold(r, threshold.Threshold(s))
					thresholdDialog.Hide()
				}
			}
		}
		thresholdDialog = dialog.NewCustomWithoutButtons("Select threshold", selector, window)
		thresholdDialog.SetOnClosed(onClosed)

		thresholdSelectorDialog := getNewSelectorDialog(thresholdDialog, battle.Thresholds.GetOptions(), thresholdOnChanged)

		/* Offset Dialog */
		var offsetDialog *dialog.CustomDialog
		offsetOnChanged := func(r role.Role) func(s string) {
			return func(s string) {
				if s != "" {
					o, _ := strconv.Atoi(s)
					worker.ActionState.AddSkillOffset(r, offset.Offset(o))
					offsetDialog.Hide()
				}
			}
		}
		offsetDialog = dialog.NewCustomWithoutButtons("Select skill offset", selector, window)
		offsetDialog.SetOnClosed(onClosed)

		offsetSelectorDialog := getNewSelectorDialog(offsetDialog, battle.Offsets.GetOptions(), offsetOnChanged)

		/* Level Dialog */
		var levelDialog *dialog.CustomDialog
		levelOnChanged := func(r role.Role) func(s string) {
			return func(s string) {
				if s != "" {
					l, _ := strconv.Atoi(s)
					worker.ActionState.AddCharacterSkillLevel(offset.Offset(l))
					levelDialog.Hide()
				}
			}
		}
		levelDialog = dialog.NewCustomWithoutButtons("Select skill level", selector, window)
		levelDialog.SetOnClosed(onClosed)

		levelSelectorDialog := getNewSelectorDialog(levelDialog, battle.Levels.GetOptions(), levelOnChanged)

		/* Character Actions */
		characterActionSelector := widget.NewButtonWithIcon("Character Actions", theme.ContentAddIcon(), func() {
			worker.ActionState.ClearCharacterActions()
			actionViewer.Objects = generateTags(*worker)
			actionViewer.Refresh()

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
			var bloodMagicButton *widget.Button

			attackButton = widget.NewButton(character.Attack.String(), func() {
				worker.ActionState.AddCharacterAction(character.Attack)

				dialogs := []func(){
					successControlUnitSelectorDialog(role.Character),
					failureControlUnitSelectorDialog(role.Character),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			attackButton.Importance = widget.WarningImportance

			defendButton = widget.NewButton(character.Defend.String(), func() {
				worker.ActionState.AddCharacterAction(character.Defend)

				dialogs := []func(){
					successControlUnitSelectorDialog(role.Character),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			defendButton.Importance = widget.WarningImportance

			escapeButton = widget.NewButton(character.Escape.String(), func() {
				worker.ActionState.AddCharacterAction(character.Escape)

				dialogs := []func(){
					failureControlUnitSelectorDialog(role.Character),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			escapeButton.Importance = widget.WarningImportance

			catchButton = widget.NewButton(character.Catch.String(), func() {
				worker.ActionState.AddCharacterAction(character.Catch)
				actionViewer.Objects = generateTags(*worker)
				actionViewer.Refresh()

				dialogs := []func(){
					healingRatioSelectorDialog(role.Character),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)

				notifyLogConfig("About Catch")
			})
			catchButton.Importance = widget.SuccessImportance

			bombButton = widget.NewButton(character.Bomb.String(), func() {
				worker.ActionState.AddCharacterAction(character.Bomb)

				dialogs := []func(){
					bombSelectorDialog(role.Character),
					successControlUnitSelectorDialog(role.Character),
					failureControlUnitSelectorDialog(role.Character),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			bombButton.Importance = widget.HighImportance

			potionButton = widget.NewButton(character.Potion.String(), func() {
				worker.ActionState.AddCharacterAction(character.Potion)

				dialogs := []func(){
					healingRatioSelectorDialog(role.Character),
					successControlUnitSelectorDialog(role.Character),
					failureControlUnitSelectorDialog(role.Character),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			potionButton.Importance = widget.HighImportance

			recallButton = widget.NewButton(character.Recall.String(), func() {
				worker.ActionState.AddCharacterAction(character.Recall)

				dialogs := []func(){
					successControlUnitSelectorDialog(role.Character),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			recallButton.Importance = widget.HighImportance

			moveButton = widget.NewButton(character.Move.String(), func() {
				worker.ActionState.AddCharacterAction(character.Move)

				dialogs := []func(){
					successControlUnitSelectorDialog(role.Character),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			moveButton.Importance = widget.WarningImportance

			hangButton = widget.NewButton(character.Hang.String(), func() {
				worker.ActionState.AddCharacterAction(character.Hang)
				actionViewer.Objects = generateTags(*worker)
				actionViewer.Refresh()
			})
			hangButton.Importance = widget.SuccessImportance

			skillButton = widget.NewButton(character.Skill.String(), func() {
				worker.ActionState.AddCharacterAction(character.Skill)

				dialogs := []func(){
					offsetSelectorDialog(role.Character),
					levelSelectorDialog(role.Character),
					successControlUnitSelectorDialog(role.Character),
					failureControlUnitSelectorDialog(role.Character),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			skillButton.Importance = widget.HighImportance

			thresholdSkillButton = widget.NewButton(character.ThresholdSkill.String(), func() {
				worker.ActionState.AddCharacterAction(character.ThresholdSkill)

				dialogs := []func(){
					offsetSelectorDialog(role.Character),
					levelSelectorDialog(role.Character),
					thresholdSelectorDialog(role.Character),
					successControlUnitSelectorDialog(role.Character),
					failureControlUnitSelectorDialog(role.Character),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			thresholdSkillButton.Importance = widget.HighImportance

			stealButton = widget.NewButton(character.Steal.String(), func() {
				worker.ActionState.AddCharacterAction(character.Steal)

				dialogs := []func(){
					offsetSelectorDialog(role.Character),
					levelSelectorDialog(role.Character),
					successControlUnitSelectorDialog(role.Character),
					failureControlUnitSelectorDialog(role.Character),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			stealButton.Importance = widget.SuccessImportance

			trainButton = widget.NewButton(character.TrainSkill.String(), func() {
				worker.ActionState.AddCharacterAction(character.TrainSkill)

				dialogs := []func(){
					offsetSelectorDialog(role.Character),
					levelSelectorDialog(role.Character),
					successControlUnitSelectorDialog(role.Character),
					failureControlUnitSelectorDialog(role.Character),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			trainButton.Importance = widget.SuccessImportance

			rideButton = widget.NewButton(character.Ride.String(), func() {
				worker.ActionState.AddCharacterAction(character.Ride)

				dialogs := []func(){
					offsetSelectorDialog(role.Character),
					levelSelectorDialog(role.Character),
					successControlUnitSelectorDialog(role.Character),
					failureControlUnitSelectorDialog(role.Character),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			rideButton.Importance = widget.HighImportance

			healSelfButton = widget.NewButton(character.HealSelf.String(), func() {
				worker.ActionState.AddCharacterAction(character.HealSelf)

				dialogs := []func(){
					offsetSelectorDialog(role.Character),
					levelSelectorDialog(role.Character),
					healingRatioSelectorDialog(role.Character),
					successControlUnitSelectorDialog(role.Character),
					failureControlUnitSelectorDialog(role.Character),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			healSelfButton.Importance = widget.HighImportance

			healOneButton = widget.NewButton(character.HealOne.String(), func() {
				worker.ActionState.AddCharacterAction(character.HealOne)

				dialogs := []func(){
					offsetSelectorDialog(role.Character),
					levelSelectorDialog(role.Character),
					healingRatioSelectorDialog(role.Character),
					successControlUnitSelectorDialog(role.Character),
					failureControlUnitSelectorDialog(role.Character),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			healOneButton.Importance = widget.HighImportance

			healTShapeButton = widget.NewButton(character.HealTShaped.String(), func() {
				worker.ActionState.AddCharacterAction(character.HealTShaped)

				dialogs := []func(){
					offsetSelectorDialog(role.Character),
					levelSelectorDialog(role.Character),
					healingRatioSelectorDialog(role.Character),
					successControlUnitSelectorDialog(role.Character),
					failureControlUnitSelectorDialog(role.Character),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			healTShapeButton.Importance = widget.HighImportance

			healMultiButton = widget.NewButton(character.HealMulti.String(), func() {
				worker.ActionState.AddCharacterAction(character.HealMulti)

				dialogs := []func(){
					offsetSelectorDialog(role.Character),
					levelSelectorDialog(role.Character),
					healingRatioSelectorDialog(role.Character),
					successControlUnitSelectorDialog(role.Character),
					failureControlUnitSelectorDialog(role.Character),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			healMultiButton.Importance = widget.HighImportance

			bloodMagicButton = widget.NewButton(character.BloodMagic.String(), func() {
				worker.ActionState.AddCharacterAction(character.BloodMagic)

				dialogs := []func(){
					offsetSelectorDialog(role.Character),
					levelSelectorDialog(role.Character),
					healingRatioSelectorDialog(role.Character),
					successControlUnitSelectorDialog(role.Character),
					failureControlUnitSelectorDialog(role.Character),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			bloodMagicButton.Importance = widget.HighImportance

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
				bloodMagicButton,
				healSelfButton,
				healOneButton,
				healTShapeButton,
				healMultiButton,
				hangButton,
				catchButton,
				trainButton,
			)

			actionsDialog := dialog.NewCustom("Select man actions with order", "Leave", actionsContainer, window)
			actionsDialog.Show()
		})

		/* Pet Actions */
		petActionSelector := widget.NewButtonWithIcon("Pet Actions", theme.ContentAddIcon(), func() {
			worker.ActionState.ClearPetActions()
			actionViewer.Objects = generateTags(*worker)
			actionViewer.Refresh()

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

			petAttackButton = widget.NewButton(pet.Attack.String(), func() {
				worker.ActionState.AddPetAction(pet.Attack)

				dialogs := []func(){
					successControlUnitSelectorDialog(role.Pet),
					failureControlUnitSelectorDialog(role.Pet),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			petAttackButton.Importance = widget.WarningImportance

			petHangButton = widget.NewButton(pet.Hang.String(), func() {
				worker.ActionState.AddPetAction(pet.Hang)
				actionViewer.Objects = generateTags(*worker)
				actionViewer.Refresh()
			})
			petHangButton.Importance = widget.SuccessImportance

			petSkillButton = widget.NewButton(pet.Skill.String(), func() {
				worker.ActionState.AddPetAction(pet.Skill)

				dialogs := []func(){
					offsetSelectorDialog(role.Pet),
					successControlUnitSelectorDialog(role.Pet),
					failureControlUnitSelectorDialog(role.Pet),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			petSkillButton.Importance = widget.HighImportance

			petDefendButton = widget.NewButton(pet.Defend.String(), func() {
				worker.ActionState.AddPetAction(pet.Defend)

				dialogs := []func(){
					offsetSelectorDialog(role.Pet),
					successControlUnitSelectorDialog(role.Pet),
					failureControlUnitSelectorDialog(role.Pet),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			petDefendButton.Importance = widget.HighImportance

			petHealSelfButton = widget.NewButton(pet.HealSelf.String(), func() {
				worker.ActionState.AddPetAction(pet.HealSelf)

				dialogs := []func(){
					offsetSelectorDialog(role.Pet),
					healingRatioSelectorDialog(role.Pet),
					successControlUnitSelectorDialog(role.Pet),
					failureControlUnitSelectorDialog(role.Pet),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			petHealSelfButton.Importance = widget.HighImportance

			petHealOneButton = widget.NewButton(pet.HealOne.String(), func() {
				worker.ActionState.AddPetAction(pet.HealOne)

				dialogs := []func(){
					offsetSelectorDialog(role.Pet),
					healingRatioSelectorDialog(role.Pet),
					successControlUnitSelectorDialog(role.Pet),
					failureControlUnitSelectorDialog(role.Pet),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			petHealOneButton.Importance = widget.HighImportance

			petRideButton = widget.NewButton(pet.Ride.String(), func() {
				worker.ActionState.AddPetAction(pet.Ride)

				dialogs := []func(){
					offsetSelectorDialog(role.Pet),
					successControlUnitSelectorDialog(role.Pet),
					failureControlUnitSelectorDialog(role.Pet),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			petRideButton.Importance = widget.HighImportance

			petOffRideButton = widget.NewButton(pet.OffRide.String(), func() {
				worker.ActionState.AddPetAction(pet.OffRide)

				dialogs := []func(){
					offsetSelectorDialog(role.Pet),
					successControlUnitSelectorDialog(role.Pet),
					failureControlUnitSelectorDialog(role.Pet),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			petOffRideButton.Importance = widget.HighImportance

			petEscapeButton = widget.NewButton(pet.Escape.String(), func() {
				worker.ActionState.AddPetAction(pet.Escape)

				dialogs := []func(){
					failureControlUnitSelectorDialog(role.Pet),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			petEscapeButton.Importance = widget.WarningImportance

			petCatchButton = widget.NewButton(pet.Catch.String(), func() {
				worker.ActionState.AddPetAction(pet.Catch)
				actionViewer.Objects = generateTags(*worker)
				actionViewer.Refresh()

				notifyLogConfig("About Catch")

				dialogs := []func(){
					healingRatioSelectorDialog(role.Pet),
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
		characterActionSelector.Importance = widget.MediumImportance
		petActionSelector.Importance = widget.MediumImportance

		loadSettingButton := widget.NewButtonWithIcon("Load", theme.FolderOpenIcon(), func() {
			fileOpenDialog := dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
				if uc != nil {
					var actionState battle.ActionState
					file, openErr := os.Open(uc.URI().Path())

					if openErr == nil {
						defer file.Close()
						if buffer, readErr := io.ReadAll(file); readErr == nil {
							if json.Unmarshal(buffer, &actionState) == nil {
								actionState.SetHWND(worker.ActionState.GetHWND())
								worker.ActionState = actionState
								worker.ActionState.GameDir = r.gameDir
								worker.ActionState.ManaChecker = options.manaChecker
								actionViewer.Objects = generateTags(*worker)
								actionViewer.Refresh()
							}
						}
					}
				}
			}, window)

			listableURI, _ := storage.ListerForURI(storage.NewFileURI(r.actionDir + `\actions`))
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
			listableURI, _ := storage.ListerForURI(storage.NewFileURI(r.actionDir + `\actions`))
			fileSaveDialog.SetFileName("default.ac")
			fileSaveDialog.SetLocation(listableURI)
			fileSaveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".ac"}))
			fileSaveDialog.Show()
		})
		saveSettingButton.Importance = widget.MediumImportance

		workerMenuContainer.Add(aliasButton)
		workerMenuContainer.Add(movementModeButton)
		workerMenuContainer.Add(characterActionSelector)
		workerMenuContainer.Add(petActionSelector)
		workerMenuContainer.Add(loadSettingButton)
		workerMenuContainer.Add(saveSettingButton)

		actionViewer = container.NewAdaptiveGrid(6, generateTags(*worker)...)
		actionViewers = append(actionViewers, actionViewer)

		workerContainer := container.NewVBox(workerMenuContainer, actionViewer)
		gameWidget.Add(workerContainer)
	}

	return
}

type menuWidgetOptions struct {
	games          game.Games
	allGames       game.Games
	manaChecker    *string
	workers        battle.Workers
	sharedStopChan chan bool
	actionViewers  []*fyne.Container
	destroy        func()
}

func generateMenuWidget(options menuWidgetOptions) (menuWidget *fyne.Container) {
	var manaCheckerSelectorDialog *dialog.CustomDialog
	var manaCheckerSelectorButton *widget.Button
	manaCheckerOptions := []string{battle.NO_MANA_CHECKER}
	manaCheckerOptions = append(manaCheckerOptions, options.games.GetSortedKeys()...)
	manaCheckerSelector := widget.NewRadioGroup(manaCheckerOptions, func(s string) {
		if hWnd, ok := options.allGames[s]; ok {
			*options.manaChecker = fmt.Sprint(hWnd)
			manaCheckerSelectorButton.SetText(fmt.Sprintf("Mana Checker: %s", s))
		} else {
			*options.manaChecker = battle.NO_MANA_CHECKER
			manaCheckerSelectorButton.SetText(fmt.Sprintf("Mana Checker: %s", *options.manaChecker))
		}
		manaCheckerSelectorDialog.Hide()
	})
	manaCheckerSelector.Required = true
	manaCheckerSelectorDialog = dialog.NewCustomWithoutButtons("Select a mana checker with this group", manaCheckerSelector, window)
	manaCheckerSelectorButton = widget.NewButton(fmt.Sprintf("Mana Checker: %s", *options.manaChecker), func() {
		manaCheckerSelectorDialog.Show()

		notifyBeeperConfig("About Mana Checker")
	})
	manaCheckerSelectorButton.Importance = widget.HighImportance

	loadSettingButton := widget.NewButtonWithIcon("Load", theme.FolderOpenIcon(), func() {
		fileOpenDialog := dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
			if uc != nil {
				var actionState battle.ActionState
				file, openErr := os.Open(uc.URI().Path())

				if openErr == nil {
					defer file.Close()
					if buffer, readErr := io.ReadAll(file); readErr == nil {
						if json.Unmarshal(buffer, &actionState) == nil {
							for i := range options.workers {
								actionState.SetHWND(options.workers[i].ActionState.GetHWND())
								options.workers[i].ActionState = actionState
								options.workers[i].ActionState.GameDir = r.gameDir
								options.workers[i].ActionState.ManaChecker = options.manaChecker
								options.actionViewers[i].Objects = generateTags(options.workers[i])
								options.actionViewers[i].Refresh()
							}
						}
					}
				}
			}
		}, window)

		listableURI, _ := storage.ListerForURI(storage.NewFileURI(r.actionDir + `\actions`))
		fileOpenDialog.SetLocation(listableURI)
		fileOpenDialog.SetFilter(storage.NewExtensionFileFilter([]string{".ac"}))
		fileOpenDialog.Show()
	})
	loadSettingButton.Importance = widget.HighImportance

	deleteButton := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		deleteDialog := dialog.NewConfirm("Delete group", "Do you really want to delete this group?", func(isDeleting bool) {
			if isDeleting {
				for i := range options.workers {
					options.workers[i].Stop()
				}

				close(options.sharedStopChan)
				options.destroy()
			}
		}, window)
		deleteDialog.SetConfirmImportance(widget.DangerImportance)
		deleteDialog.Show()
	})
	deleteButton.Importance = widget.DangerImportance

	var switchButton *widget.Button
	switchButton = widget.NewButtonWithIcon("", theme.MediaPlayIcon(), func() {
		switch switchButton.Icon {
		case theme.MediaPlayIcon():
			for i := range options.workers {
				options.workers[i].Work()
			}
			turn(theme.MediaStopIcon(), switchButton)
		case theme.MediaStopIcon():
			for i := range options.workers {
				options.workers[i].Stop()
			}
			turn(theme.MediaPlayIcon(), switchButton)
		}
	})
	switchButton.Importance = widget.WarningImportance

	var teleportAndResourceCheckerButton *widget.Button
	teleportAndResourceCheckerButton = widget.NewButtonWithIcon("Check TP & RES", theme.CheckButtonIcon(), func() {
		switch teleportAndResourceCheckerButton.Icon {
		case theme.CheckButtonCheckedIcon():
			for i := range options.workers {
				options.workers[i].StopTeleportAndResourceChecker()
			}
			turn(theme.CheckButtonIcon(), teleportAndResourceCheckerButton)
		case theme.CheckButtonIcon():
			for i := range options.workers {
				options.workers[i].StartTeleportAndResourceChecker()
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
			for i := range options.workers {
				options.workers[i].ActivityCheckerEnabled = false
			}
			turn(theme.CheckButtonIcon(), activitiesCheckerButton)
		case theme.CheckButtonIcon():
			for i := range options.workers {
				options.workers[i].ActivityCheckerEnabled = true
			}
			turn(theme.CheckButtonCheckedIcon(), activitiesCheckerButton)

			notifyBeeperAndLogConfig("About Activities Checker")
		}
	})
	activitiesCheckerButton.Importance = widget.HighImportance
	var inventoryCheckerButton *widget.Button
	inventoryCheckerButton = widget.NewButtonWithIcon("Check Inventory", theme.CheckButtonIcon(), func() {
		switch inventoryCheckerButton.Icon {
		case theme.CheckButtonCheckedIcon():
			for i := range options.workers {
				options.workers[i].StopInventoryChecker()
			}
			turn(theme.CheckButtonIcon(), inventoryCheckerButton)
		case theme.CheckButtonIcon():
			for i := range options.workers {
				options.workers[i].StartInventoryChecker()
			}
			turn(theme.CheckButtonCheckedIcon(), inventoryCheckerButton)

			notifyBeeperConfig("About Inventory Checker")
		}
	})
	inventoryCheckerButton.Importance = widget.HighImportance
	checkersButton := widget.NewButtonWithIcon("Checkers", theme.MenuIcon(), func() {
		dialog.NewCustom("Checkers", "Leave", container.NewAdaptiveGrid(4, teleportAndResourceCheckerButton, activitiesCheckerButton, inventoryCheckerButton), window).Show()
	})
	checkersButton.Importance = widget.HighImportance

	enemyOrderRadio := widget.NewRadioGroup(battle.EnemyOrder.GetOptions(), func(order string) {
		for i := range options.workers {
			options.workers[i].EnemyOrder = enemyorder.EnemyOrder(order)
		}
	})
	enemyOrderRadio.Selected = enemyorder.Default.String()
	enemyOrderRadio.Horizontal = true
	enemyOrderRadio.Required = true
	enemyOrderButton := widget.NewButtonWithIcon("Enemy Order", theme.MenuIcon(), func() {
		dialog.NewCustom("Enemy Order", "Leave", container.NewAdaptiveGrid(4, enemyOrderRadio), window).Show()
	})
	enemyOrderButton.Importance = widget.HighImportance

	menuWidget = container.NewGridWithColumns(6, manaCheckerSelectorButton, checkersButton, enemyOrderButton, loadSettingButton, deleteButton, switchButton)
	return
}

func selectorDialoger(selector *widget.RadioGroup) func(dialog *dialog.CustomDialog, options []string, onChanged func(r role.Role) func(s string)) func(r role.Role) func() {
	return func(dialog *dialog.CustomDialog, options []string, onChanged func(r role.Role) func(s string)) func(r role.Role) func() {
		return func(r role.Role) func() {
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

func generateTags(worker battle.Worker) (tagContaines []fyne.CanvasObject) {
	tagContaines = createTagContainers(worker.ActionState, role.Character)
	tagContaines = append(tagContaines, createTagContainers(worker.ActionState, role.Pet)...)
	return
}

var (
	characterFinishingTagColor   = color.RGBA{235, 206, 100, uint8(math.Round(1 * 255))}
	characterConditionalTagColor = color.RGBA{100, 206, 235, uint8(math.Round(1 * 255))}
	characterSpecialTagColor     = color.RGBA{206, 235, 100, uint8(math.Round(1 * 255))}
	petFinishingTagColor         = color.RGBA{245, 79, 0, uint8(math.Round(0.8 * 255))}
	petConditionalTagColor       = color.RGBA{0, 79, 245, uint8(math.Round(0.8 * 255))}
	petSpecialTagColor           = color.RGBA{79, 245, 0, uint8(math.Round(0.8 * 255))}
)

func createTagContainers(actionState battle.ActionState, r role.Role) (tagContainers []fyne.CanvasObject) {
	var tag string
	var anyActions []any
	if r == role.Character {
		for _, hs := range actionState.GetCharacterActions() {
			anyActions = append(anyActions, hs)
		}
	} else {
		for _, ps := range actionState.GetPetActions() {
			anyActions = append(anyActions, ps)
		}
	}

	for _, action := range anyActions {
		var tagColor color.RGBA
		if r == role.Character {
			tag = action.(battle.CharacterAction).Action.String()
			switch {
			case strings.Contains(action.(battle.CharacterAction).Action.String(), "**"):
				tagColor = characterFinishingTagColor
			case strings.Contains(action.(battle.CharacterAction).Action.String(), "*"):
				tagColor = characterConditionalTagColor
			default:
				tagColor = characterSpecialTagColor
			}
		} else {
			tag = action.(battle.PetAction).Action.String()
			switch {
			case strings.Contains(action.(battle.PetAction).Action.String(), "**"):
				tagColor = petFinishingTagColor
			case strings.Contains(action.(battle.PetAction).Action.String(), "*"):
				tagColor = petConditionalTagColor
			default:
				tagColor = petSpecialTagColor
			}
		}

		tagCanvas := canvas.NewRectangle(tagColor)
		tagCanvas.SetMinSize(fyne.NewSize(60, 22))
		tagContainer := container.NewStack(tagCanvas)

		if r == role.Character {
			if offset := action.(battle.CharacterAction).Offset; offset != 0 {
				tag = fmt.Sprintf("%s:%d:%d", tag, offset, action.(battle.CharacterAction).Level)
			}
		} else {
			if offset := action.(battle.PetAction).Offset; offset != 0 {
				tag = fmt.Sprintf("%s:%d", tag, offset)
			}
		}

		var param string
		var threshold threshold.Threshold
		var successControlUnit controlunit.ControlUnit
		var successJumpId int
		var failureControlUnit controlunit.ControlUnit
		var failureJumpId int

		if r == role.Character {
			param = action.(battle.CharacterAction).Param
			threshold = action.(battle.CharacterAction).Threshold
			successControlUnit = action.(battle.CharacterAction).SuccessControlUnit
			failureControlUnit = action.(battle.CharacterAction).FailureControlUnit
			successJumpId = action.(battle.CharacterAction).SuccessJumpId
			failureJumpId = action.(battle.CharacterAction).FailureJumpId
		} else {
			param = action.(battle.PetAction).Param
			threshold = action.(battle.PetAction).Threshold
			successControlUnit = action.(battle.PetAction).SuccessControlUnit
			failureControlUnit = action.(battle.PetAction).FailureControlUnit
			successJumpId = action.(battle.PetAction).SuccessJumpId
			failureJumpId = action.(battle.PetAction).FailureJumpId
		}

		if param != "" {
			tag = fmt.Sprintf("%s:%s", tag, param)
		}

		if threshold != "" {
			tag = fmt.Sprintf("%s:%s", tag, threshold)
		}

		if len(successControlUnit) > 0 {
			controlUnitFirstLetter := strings.ToLower(successControlUnit[:1].String())
			tag = fmt.Sprintf("%s:%s", tag, controlUnitFirstLetter)
			if controlUnitFirstLetter == "j" {
				tag = fmt.Sprintf("%s%d", tag, successJumpId)
			}
		} else {
			tag = tag + ":"
		}

		if len(failureControlUnit) > 0 {
			controlUnitFirstLetter := strings.ToLower(failureControlUnit[:1].String())
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
	if !utils.Beeper.IsReady() {
		go func() {
			time.Sleep(200 * time.Millisecond)
			dialog.NewInformation(title, "Remember to setup the alert music!!!", window).Show()
		}()
	}
}

func notifyLogConfig(title string) {
	if *r.gameDir == "" {
		go func() {
			time.Sleep(200 * time.Millisecond)
			dialog.NewInformation(title, "Remember to setup the log directory!!!", window).Show()
		}()
	}
}

func notifyBeeperAndLogConfig(title string) {
	if !utils.Beeper.IsReady() || *r.gameDir == "" {
		go func() {
			time.Sleep(200 * time.Millisecond)
			dialog.NewInformation(title, "Remember to setup the alert music and log directory!!!", window).Show()
		}()
	}
}

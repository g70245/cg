package containers

import (
	"cg/game"
	"cg/game/battle"
	"cg/game/battle/enums/controlunit"
	"cg/game/battle/enums/movement"
	"cg/game/battle/enums/role"
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
		movementModeSelector := widget.NewRadioGroup(battle.MOVEMENT_MODES.GetOptions(), func(s string) {
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
		failureControlUnitOnChanged := func(r role.Role) func(s string) {
			return func(s string) {
				if s == "" {
					return
				}

				if cu := controlunit.ControlUnit(s); cu == controlunit.Jump {
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
		bombSelectorDialog := getNewSelectorDialog(paramDialog, game.Bombs.GetOptions(), paramOnChanged)

		/* Threshold Dialog */
		var thresholdDialog *dialog.CustomDialog
		thresholdOnChanged := func(r role.Role) func(s string) {
			return func(s string) {
				if s != "" {
					worker.ActionState.AddThreshold(r, battle.Threshold(s))
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
					offset, _ := strconv.Atoi(s)
					worker.ActionState.AddSkillOffset(r, battle.Offset(offset))
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
					level, _ := strconv.Atoi(s)
					worker.ActionState.AddHumanSkillLevel(battle.Offset(level))
					levelDialog.Hide()
				}
			}
		}
		levelDialog = dialog.NewCustomWithoutButtons("Select skill level", selector, window)
		levelDialog.SetOnClosed(onClosed)

		levelSelectorDialog := getNewSelectorDialog(levelDialog, battle.Levels.GetOptions(), levelOnChanged)

		/* Human Actions */
		humanActionSelector := widget.NewButtonWithIcon("Character Actions", theme.ContentAddIcon(), func() {
			worker.ActionState.ClearHumanActions()
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

			attackButton = widget.NewButton(battle.HumanAttack.String(), func() {
				worker.ActionState.AddHumanAction(battle.HumanAttack)

				dialogs := []func(){
					successControlUnitSelectorDialog(role.Human),
					failureControlUnitSelectorDialog(role.Human),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			attackButton.Importance = widget.WarningImportance

			defendButton = widget.NewButton(battle.HumanDefend.String(), func() {
				worker.ActionState.AddHumanAction(battle.HumanDefend)

				dialogs := []func(){
					successControlUnitSelectorDialog(role.Human),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			defendButton.Importance = widget.WarningImportance

			escapeButton = widget.NewButton(battle.HumanEscape.String(), func() {
				worker.ActionState.AddHumanAction(battle.HumanEscape)

				dialogs := []func(){
					failureControlUnitSelectorDialog(role.Human),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			escapeButton.Importance = widget.WarningImportance

			catchButton = widget.NewButton(battle.HumanCatch.String(), func() {
				worker.ActionState.AddHumanAction(battle.HumanCatch)
				actionViewer.Objects = generateTags(*worker)
				actionViewer.Refresh()

				dialogs := []func(){
					healingRatioSelectorDialog(role.Human),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)

				notifyLogConfig("About Catch")
			})
			catchButton.Importance = widget.SuccessImportance

			bombButton = widget.NewButton(battle.HumanBomb.String(), func() {
				worker.ActionState.AddHumanAction(battle.HumanBomb)

				dialogs := []func(){
					bombSelectorDialog(role.Human),
					successControlUnitSelectorDialog(role.Human),
					failureControlUnitSelectorDialog(role.Human),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			bombButton.Importance = widget.HighImportance

			potionButton = widget.NewButton(battle.HumanPotion.String(), func() {
				worker.ActionState.AddHumanAction(battle.HumanPotion)

				dialogs := []func(){
					healingRatioSelectorDialog(role.Human),
					successControlUnitSelectorDialog(role.Human),
					failureControlUnitSelectorDialog(role.Human),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			potionButton.Importance = widget.HighImportance

			recallButton = widget.NewButton(battle.HumanRecall.String(), func() {
				worker.ActionState.AddHumanAction(battle.HumanRecall)

				dialogs := []func(){
					successControlUnitSelectorDialog(role.Human),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			recallButton.Importance = widget.HighImportance

			moveButton = widget.NewButton(battle.HumanMove.String(), func() {
				worker.ActionState.AddHumanAction(battle.HumanMove)

				dialogs := []func(){
					successControlUnitSelectorDialog(role.Human),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			moveButton.Importance = widget.WarningImportance

			hangButton = widget.NewButton(battle.HumanHang.String(), func() {
				worker.ActionState.AddHumanAction(battle.HumanHang)
				actionViewer.Objects = generateTags(*worker)
				actionViewer.Refresh()
			})
			hangButton.Importance = widget.SuccessImportance

			skillButton = widget.NewButton(battle.HumanSkill.String(), func() {
				worker.ActionState.AddHumanAction(battle.HumanSkill)

				dialogs := []func(){
					offsetSelectorDialog(role.Human),
					levelSelectorDialog(role.Human),
					successControlUnitSelectorDialog(role.Human),
					failureControlUnitSelectorDialog(role.Human),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			skillButton.Importance = widget.HighImportance

			thresholdSkillButton = widget.NewButton(battle.HumanThresholdSkill.String(), func() {
				worker.ActionState.AddHumanAction(battle.HumanThresholdSkill)

				dialogs := []func(){
					offsetSelectorDialog(role.Human),
					levelSelectorDialog(role.Human),
					thresholdSelectorDialog(role.Human),
					successControlUnitSelectorDialog(role.Human),
					failureControlUnitSelectorDialog(role.Human),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			thresholdSkillButton.Importance = widget.HighImportance

			stealButton = widget.NewButton(battle.HumanSteal.String(), func() {
				worker.ActionState.AddHumanAction(battle.HumanSteal)

				dialogs := []func(){
					offsetSelectorDialog(role.Human),
					levelSelectorDialog(role.Human),
					successControlUnitSelectorDialog(role.Human),
					failureControlUnitSelectorDialog(role.Human),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			stealButton.Importance = widget.SuccessImportance

			trainButton = widget.NewButton(battle.HumanTrainSkill.String(), func() {
				worker.ActionState.AddHumanAction(battle.HumanTrainSkill)

				dialogs := []func(){
					offsetSelectorDialog(role.Human),
					levelSelectorDialog(role.Human),
					successControlUnitSelectorDialog(role.Human),
					failureControlUnitSelectorDialog(role.Human),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			trainButton.Importance = widget.SuccessImportance

			rideButton = widget.NewButton(battle.HumanRide.String(), func() {
				worker.ActionState.AddHumanAction(battle.HumanRide)

				dialogs := []func(){
					offsetSelectorDialog(role.Human),
					levelSelectorDialog(role.Human),
					successControlUnitSelectorDialog(role.Human),
					failureControlUnitSelectorDialog(role.Human),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			rideButton.Importance = widget.HighImportance

			healSelfButton = widget.NewButton(battle.HumanHealSelf.String(), func() {
				worker.ActionState.AddHumanAction(battle.HumanHealSelf)

				dialogs := []func(){
					offsetSelectorDialog(role.Human),
					levelSelectorDialog(role.Human),
					healingRatioSelectorDialog(role.Human),
					successControlUnitSelectorDialog(role.Human),
					failureControlUnitSelectorDialog(role.Human),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			healSelfButton.Importance = widget.HighImportance

			healOneButton = widget.NewButton(battle.HumanHealOne.String(), func() {
				worker.ActionState.AddHumanAction(battle.HumanHealOne)

				dialogs := []func(){
					offsetSelectorDialog(role.Human),
					levelSelectorDialog(role.Human),
					healingRatioSelectorDialog(role.Human),
					successControlUnitSelectorDialog(role.Human),
					failureControlUnitSelectorDialog(role.Human),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			healOneButton.Importance = widget.HighImportance

			healTShapeButton = widget.NewButton(battle.HumanHealTShaped.String(), func() {
				worker.ActionState.AddHumanAction(battle.HumanHealTShaped)

				dialogs := []func(){
					offsetSelectorDialog(role.Human),
					levelSelectorDialog(role.Human),
					healingRatioSelectorDialog(role.Human),
					successControlUnitSelectorDialog(role.Human),
					failureControlUnitSelectorDialog(role.Human),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			healTShapeButton.Importance = widget.HighImportance

			healMultiButton = widget.NewButton(battle.HumanHealMulti.String(), func() {
				worker.ActionState.AddHumanAction(battle.HumanHealMulti)

				dialogs := []func(){
					offsetSelectorDialog(role.Human),
					levelSelectorDialog(role.Human),
					healingRatioSelectorDialog(role.Human),
					successControlUnitSelectorDialog(role.Human),
					failureControlUnitSelectorDialog(role.Human),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			healMultiButton.Importance = widget.HighImportance

			bloodMagicButton = widget.NewButton(battle.HumanBloodMagic.String(), func() {
				worker.ActionState.AddHumanAction(battle.HumanBloodMagic)

				dialogs := []func(){
					offsetSelectorDialog(role.Human),
					levelSelectorDialog(role.Human),
					healingRatioSelectorDialog(role.Human),
					successControlUnitSelectorDialog(role.Human),
					failureControlUnitSelectorDialog(role.Human),
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
				stealButton,
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

			petAttackButton = widget.NewButton(battle.PetAttack.String(), func() {
				worker.ActionState.AddPetAction(battle.PetAttack)

				dialogs := []func(){
					successControlUnitSelectorDialog(role.Pet),
					failureControlUnitSelectorDialog(role.Pet),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			petAttackButton.Importance = widget.WarningImportance

			petHangButton = widget.NewButton(battle.PetHang.String(), func() {
				worker.ActionState.AddPetAction(battle.PetHang)
				actionViewer.Objects = generateTags(*worker)
				actionViewer.Refresh()
			})
			petHangButton.Importance = widget.SuccessImportance

			petSkillButton = widget.NewButton(battle.PetSkill.String(), func() {
				worker.ActionState.AddPetAction(battle.PetSkill)

				dialogs := []func(){
					offsetSelectorDialog(role.Pet),
					successControlUnitSelectorDialog(role.Pet),
					failureControlUnitSelectorDialog(role.Pet),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			petSkillButton.Importance = widget.HighImportance

			petDefendButton = widget.NewButton(battle.PetDefend.String(), func() {
				worker.ActionState.AddPetAction(battle.PetDefend)

				dialogs := []func(){
					offsetSelectorDialog(role.Pet),
					successControlUnitSelectorDialog(role.Pet),
					failureControlUnitSelectorDialog(role.Pet),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			petDefendButton.Importance = widget.HighImportance

			petHealSelfButton = widget.NewButton(battle.PetHealSelf.String(), func() {
				worker.ActionState.AddPetAction(battle.PetHealSelf)

				dialogs := []func(){
					offsetSelectorDialog(role.Pet),
					healingRatioSelectorDialog(role.Pet),
					successControlUnitSelectorDialog(role.Pet),
					failureControlUnitSelectorDialog(role.Pet),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			petHealSelfButton.Importance = widget.HighImportance

			petHealOneButton = widget.NewButton(battle.PetHealOne.String(), func() {
				worker.ActionState.AddPetAction(battle.PetHealOne)

				dialogs := []func(){
					offsetSelectorDialog(role.Pet),
					healingRatioSelectorDialog(role.Pet),
					successControlUnitSelectorDialog(role.Pet),
					failureControlUnitSelectorDialog(role.Pet),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			petHealOneButton.Importance = widget.HighImportance

			petRideButton = widget.NewButton(battle.PetRide.String(), func() {
				worker.ActionState.AddPetAction(battle.PetRide)

				dialogs := []func(){
					offsetSelectorDialog(role.Pet),
					successControlUnitSelectorDialog(role.Pet),
					failureControlUnitSelectorDialog(role.Pet),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			petRideButton.Importance = widget.HighImportance

			petOffRideButton = widget.NewButton(battle.PetOffRide.String(), func() {
				worker.ActionState.AddPetAction(battle.PetOffRide)

				dialogs := []func(){
					offsetSelectorDialog(role.Pet),
					successControlUnitSelectorDialog(role.Pet),
					failureControlUnitSelectorDialog(role.Pet),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			petOffRideButton.Importance = widget.HighImportance

			petEscapeButton = widget.NewButton(battle.PetEscape.String(), func() {
				worker.ActionState.AddPetAction(battle.PetEscape)

				dialogs := []func(){
					failureControlUnitSelectorDialog(role.Pet),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			petEscapeButton.Importance = widget.WarningImportance

			petCatchButton = widget.NewButton(battle.PetCatch.String(), func() {
				worker.ActionState.AddPetAction(battle.PetCatch)
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
		humanActionSelector.Importance = widget.MediumImportance
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

			listableURI, _ := storage.ListerForURI(storage.NewFileURI(*r.gameDir + `\actions`))
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
			listableURI, _ := storage.ListerForURI(storage.NewFileURI(*r.gameDir + `\actions`))
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

		listableURI, _ := storage.ListerForURI(storage.NewFileURI(*r.gameDir + `\actions`))
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

	menuWidget = container.NewGridWithColumns(5, manaCheckerSelectorButton, checkersButton, loadSettingButton, deleteButton, switchButton)
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
	tagContaines = createTagContainers(worker.ActionState, role.Human)
	tagContaines = append(tagContaines, createTagContainers(worker.ActionState, role.Pet)...)
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

func createTagContainers(actionState battle.ActionState, r role.Role) (tagContainers []fyne.CanvasObject) {
	var tag string
	var anyActions []any
	if r == role.Human {
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
		if r == role.Human {
			tag = action.(battle.HumanAction).Action.String()
			switch {
			case strings.Contains(action.(battle.HumanAction).Action.String(), "**"):
				tagColor = humanFinishingTagColor
			case strings.Contains(action.(battle.HumanAction).Action.String(), "*"):
				tagColor = humanConditionalTagColor
			default:
				tagColor = humanSpecialTagColor
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

		if r == role.Human {
			if offset := action.(battle.HumanAction).Offset; offset != 0 {
				tag = fmt.Sprintf("%s:%d:%d", tag, offset, action.(battle.HumanAction).Level)
			}
		} else {
			if offset := action.(battle.PetAction).Offset; offset != 0 {
				tag = fmt.Sprintf("%s:%d", tag, offset)
			}
		}

		var param string
		var threshold battle.Threshold
		var successControlUnit string
		var successJumpId int
		var failureControlUnit string
		var failureJumpId int

		if r == role.Human {
			param = action.(battle.HumanAction).Param
			threshold = action.(battle.HumanAction).Threshold
			successControlUnit = string(action.(battle.HumanAction).SuccessControlUnit)
			failureControlUnit = string(action.(battle.HumanAction).FailureControlUnit)
			successJumpId = action.(battle.HumanAction).SuccessJumpId
			failureJumpId = action.(battle.HumanAction).FailureJumpId
		} else {
			param = action.(battle.PetAction).Param
			threshold = action.(battle.PetAction).Threshold
			successControlUnit = string(action.(battle.PetAction).SuccessControlUnit)
			failureControlUnit = string(action.(battle.PetAction).FailureControlUnit)
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

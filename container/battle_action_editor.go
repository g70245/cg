package container

import (
	"cg/game"
	"cg/game/battle"
	"cg/game/enum/character"
	"cg/game/enum/controlunit"
	"cg/game/enum/movement"
	"cg/game/enum/offset"
	"cg/game/enum/pet"
	"cg/game/enum/role"
	"cg/game/enum/threshold"
	"cg/game/items"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const separator = "    "

type gameWidgeOptions struct {
	games       game.Games
	allGames    game.Games
	manaChecker *battle.ManaChecker
	workers     battle.Workers
}

func generateGameWidget(options gameWidgeOptions) (gameWidget *fyne.Container, actionViewers []*fyne.Container) {
	gameWidget = container.NewVBox()
	actionViewers = []*fyne.Container{}

	for i := range options.workers {
		workerMenuContainer := container.NewGridWithColumns(7)
		worker := options.workers[i]

		var aliasButton *widget.Button
		aliasButton = widget.NewButtonWithIcon(options.allGames.FindKey(worker.GetHandle()), theme.AccountIcon(), func() {
			aliasEntry := widget.NewEntry()
			aliasEntry.SetPlaceHolder("Alias")

			aliasDialog := dialog.NewCustom("Set Alias", "Save", aliasEntry, window)
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
				mode := movement.Mode(s)
				worker.SetMovementMode(mode)
				if mode != movement.None {
					movementModeButton.SetText(s)
				} else {
					movementModeButton.SetText("Movement")
				}
			}
			movementModeDialog.Hide()
		})
		movementModeSelector.Required = true
		movementModeDialog = dialog.NewCustomWithoutButtons("Select Movement Pattern", movementModeSelector, window)
		movementModeButton = widget.NewButtonWithIcon("Movement", theme.MailReplyIcon(), func() {
			movementModeDialog.Show()
		})

		enemyOrderBindingStr := binding.NewString()
		enemyOrderCheckGroup := widget.NewCheckGroup(battle.EnemyPositions.GetOptions(), func(s []string) {
			enemyOrderBindingStr.Set(strings.Join(s, separator))
		})
		enemyOrderCheckGroup.Horizontal = true
		enemyOrderLabel := widget.NewLabelWithData(enemyOrderBindingStr)
		enemyOrderButton := widget.NewButtonWithIcon("Target Priority", theme.SearchIcon(), func() {
			order := worker.CustomEnemyOrder()
			enemyOrderCheckGroup.Selected = order
			enemyOrderBindingStr.Set(strings.Join(order, separator))

			d := dialog.NewCustom("Target Priority", "Apply", container.NewVBox(enemyOrderCheckGroup, enemyOrderLabel), window)
			d.SetOnClosed(func() {
				worker.SetCustomEnemyOrder(enemyOrderCheckGroup.Selected)
			})
			d.Show()
		})

		var selector = widget.NewRadioGroup(nil, nil)
		selector.Horizontal = true
		selector.Required = true
		selectorDialogEnableChan := make(chan bool)
		getNewSelectorDialog := selectorDialoger(selector)

		var actionViewer *fyne.Container
		updateActionState := func(update func(*battle.ActionState)) {
			worker.UpdateActionState(update)
		}
		refreshActionViewer := func() {
			actionViewer.Objects = generateTags(worker.ActionStateSnapshot())
			actionViewer.Refresh()
		}
		onClosed := func() {
			refreshActionViewer()
			selectorDialogEnableChan <- true
		}

		/* Control Unit Dialogs */
		var successControlUnitDialog *dialog.CustomDialog
		var failureControlUnitDialog *dialog.CustomDialog

		activateJumpDialog := func(totalActions int, callback func(jumpId int)) {
			jumpEntry := widget.NewEntry()
			jumpEntry.Validator = func(jumpIdStr string) error {
				return validateActionID(jumpIdStr, totalActions-2)
			}

			jumpDialog := dialog.NewForm("Jump to Action", "Jump", "Cancel", []*widget.FormItem{widget.NewFormItem("Action ID", jumpEntry)}, func(isValid bool) {
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
					actionState := worker.ActionStateSnapshot()
					activateJumpDialog(len(actionState.CharacterActions), func(jumpId int) {
						updateActionState(func(actionState *battle.ActionState) {
							actionState.AddSuccessControlUnit(r, cu)
							actionState.AddSuccessJumpId(r, jumpId)
						})
						successControlUnitDialog.Hide()
					})
				} else {
					updateActionState(func(actionState *battle.ActionState) {
						actionState.AddSuccessControlUnit(r, cu)
					})
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
					actionState := worker.ActionStateSnapshot()
					activateJumpDialog(len(actionState.CharacterActions), func(jumpId int) {
						updateActionState(func(actionState *battle.ActionState) {
							actionState.AddFailureControlUnit(r, cu)
							actionState.AddFailureJumpId(r, jumpId)
						})
						failureControlUnitDialog.Hide()
					})
				} else {
					updateActionState(func(actionState *battle.ActionState) {
						actionState.AddFailureControlUnit(r, cu)
					})
					failureControlUnitDialog.Hide()
				}
			}
		}
		controlUnitOnClosed := func() {
			refreshActionViewer()
			selectorDialogEnableChan <- true
		}
		successControlUnitDialog = dialog.NewCustomWithoutButtons("Select Next Action After Success", selector, window)
		successControlUnitDialog.SetOnClosed(controlUnitOnClosed)
		failureControlUnitDialog = dialog.NewCustomWithoutButtons("Select Next Action After Failure", selector, window)
		failureControlUnitDialog.SetOnClosed(controlUnitOnClosed)

		successControlUnitSelectorDialog := getNewSelectorDialog(successControlUnitDialog, battle.ControlUnits.GetOptions(), successControlUnitOnChanged)
		failureControlUnitSelectorDialog := getNewSelectorDialog(failureControlUnitDialog, battle.ControlUnits.GetOptions(), failureControlUnitOnChanged)

		/* Param Dialogs */
		var paramDialog *dialog.CustomDialog
		paramOnChanged := func(r role.Role) func(s string) {
			return func(s string) {
				if s != "" {
					updateActionState(func(actionState *battle.ActionState) {
						actionState.AddParam(r, s)
					})
					paramDialog.Hide()
				}
			}
		}
		paramDialog = dialog.NewCustomWithoutButtons("Select Parameter", selector, window)
		paramDialog.SetOnClosed(onClosed)

		healingRatioSelectorDialog := getNewSelectorDialog(paramDialog, battle.Ratios.GetOptions(), paramOnChanged)
		bombSelectorDialog := getNewSelectorDialog(paramDialog, items.Bombs.GetOptions(), paramOnChanged)

		/* Threshold Dialog */
		var thresholdDialog *dialog.CustomDialog
		thresholdOnChanged := func(r role.Role) func(s string) {
			return func(s string) {
				if s != "" {
					updateActionState(func(actionState *battle.ActionState) {
						actionState.AddThreshold(r, threshold.Threshold(s))
					})
					thresholdDialog.Hide()
				}
			}
		}
		thresholdDialog = dialog.NewCustomWithoutButtons("Select Enemy Count", selector, window)
		thresholdDialog.SetOnClosed(onClosed)

		thresholdSelectorDialog := getNewSelectorDialog(thresholdDialog, battle.Thresholds.GetOptions(), thresholdOnChanged)

		/* Offset Dialog */
		var offsetDialog *dialog.CustomDialog
		offsetOnChanged := func(r role.Role) func(s string) {
			return func(s string) {
				if s != "" {
					o, _ := strconv.Atoi(s)
					updateActionState(func(actionState *battle.ActionState) {
						actionState.AddSkillOffset(r, offset.Offset(o))
					})
					offsetDialog.Hide()
				}
			}
		}
		offsetDialog = dialog.NewCustomWithoutButtons("Select Skill Position", selector, window)
		offsetDialog.SetOnClosed(onClosed)

		offsetSelectorDialog := getNewSelectorDialog(offsetDialog, battle.Offsets.GetOptions(), offsetOnChanged)

		/* Level Dialog */
		var levelDialog *dialog.CustomDialog
		levelOnChanged := func(r role.Role) func(s string) {
			return func(s string) {
				if s != "" {
					l, _ := strconv.Atoi(s)
					updateActionState(func(actionState *battle.ActionState) {
						actionState.AddCharacterSkillLevel(offset.Offset(l))
					})
					levelDialog.Hide()
				}
			}
		}
		levelDialog = dialog.NewCustomWithoutButtons("Select Skill Level", selector, window)
		levelDialog.SetOnClosed(onClosed)

		levelSelectorDialog := getNewSelectorDialog(levelDialog, battle.Levels.GetOptions(), levelOnChanged)

		/* Character Actions */
		characterActionSelector := widget.NewButtonWithIcon("Character Actions", theme.ContentAddIcon(), func() {
			updateActionState(func(actionState *battle.ActionState) {
				actionState.ClearCharacterActions()
			})
			refreshActionViewer()

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
				updateActionState(func(actionState *battle.ActionState) { actionState.AddCharacterAction(character.Attack) })

				dialogs := []func(){
					successControlUnitSelectorDialog(role.Character),
					failureControlUnitSelectorDialog(role.Character),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			attackButton.Importance = widget.WarningImportance

			defendButton = widget.NewButton(character.Defend.String(), func() {
				updateActionState(func(actionState *battle.ActionState) { actionState.AddCharacterAction(character.Defend) })

				dialogs := []func(){
					successControlUnitSelectorDialog(role.Character),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			defendButton.Importance = widget.WarningImportance

			escapeButton = widget.NewButton(character.Escape.String(), func() {
				updateActionState(func(actionState *battle.ActionState) { actionState.AddCharacterAction(character.Escape) })

				dialogs := []func(){
					failureControlUnitSelectorDialog(role.Character),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			escapeButton.Importance = widget.WarningImportance

			catchButton = widget.NewButton(character.Catch.String(), func() {
				updateActionState(func(actionState *battle.ActionState) { actionState.AddCharacterAction(character.Catch) })
				refreshActionViewer()

				dialogs := []func(){
					healingRatioSelectorDialog(role.Character),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)

				notifyLogConfig("Catch Setup")
			})
			catchButton.Importance = widget.SuccessImportance

			bombButton = widget.NewButton(character.Bomb.String(), func() {
				updateActionState(func(actionState *battle.ActionState) { actionState.AddCharacterAction(character.Bomb) })

				dialogs := []func(){
					bombSelectorDialog(role.Character),
					successControlUnitSelectorDialog(role.Character),
					failureControlUnitSelectorDialog(role.Character),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			bombButton.Importance = widget.HighImportance

			potionButton = widget.NewButton(character.Potion.String(), func() {
				updateActionState(func(actionState *battle.ActionState) { actionState.AddCharacterAction(character.Potion) })

				dialogs := []func(){
					healingRatioSelectorDialog(role.Character),
					successControlUnitSelectorDialog(role.Character),
					failureControlUnitSelectorDialog(role.Character),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			potionButton.Importance = widget.HighImportance

			recallButton = widget.NewButton(character.Recall.String(), func() {
				updateActionState(func(actionState *battle.ActionState) { actionState.AddCharacterAction(character.Recall) })

				dialogs := []func(){
					successControlUnitSelectorDialog(role.Character),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			recallButton.Importance = widget.HighImportance

			moveButton = widget.NewButton(character.Move.String(), func() {
				updateActionState(func(actionState *battle.ActionState) { actionState.AddCharacterAction(character.Move) })

				dialogs := []func(){
					successControlUnitSelectorDialog(role.Character),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			moveButton.Importance = widget.WarningImportance

			hangButton = widget.NewButton(character.Hang.String(), func() {
				updateActionState(func(actionState *battle.ActionState) { actionState.AddCharacterAction(character.Hang) })
				refreshActionViewer()
			})
			hangButton.Importance = widget.SuccessImportance

			skillButton = widget.NewButton(character.Skill.String(), func() {
				updateActionState(func(actionState *battle.ActionState) { actionState.AddCharacterAction(character.Skill) })

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
				updateActionState(func(actionState *battle.ActionState) { actionState.AddCharacterAction(character.ThresholdSkill) })

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
				updateActionState(func(actionState *battle.ActionState) { actionState.AddCharacterAction(character.Steal) })

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
				updateActionState(func(actionState *battle.ActionState) { actionState.AddCharacterAction(character.TrainSkill) })

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
				updateActionState(func(actionState *battle.ActionState) { actionState.AddCharacterAction(character.Ride) })

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
				updateActionState(func(actionState *battle.ActionState) { actionState.AddCharacterAction(character.HealSelf) })

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
				updateActionState(func(actionState *battle.ActionState) { actionState.AddCharacterAction(character.HealOne) })

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
				updateActionState(func(actionState *battle.ActionState) { actionState.AddCharacterAction(character.HealTShaped) })

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
				updateActionState(func(actionState *battle.ActionState) { actionState.AddCharacterAction(character.HealMulti) })

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
				updateActionState(func(actionState *battle.ActionState) { actionState.AddCharacterAction(character.BloodMagic) })

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

			actionsDialog := dialog.NewCustom("Add Character Actions in Order", "Close", actionsContainer, window)
			actionsDialog.Show()
		})

		/* Pet Actions */
		petActionSelector := widget.NewButtonWithIcon("Pet Actions", theme.ContentAddIcon(), func() {
			updateActionState(func(actionState *battle.ActionState) {
				actionState.ClearPetActions()
			})
			refreshActionViewer()

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
				updateActionState(func(actionState *battle.ActionState) { actionState.AddPetAction(pet.Attack) })

				dialogs := []func(){
					successControlUnitSelectorDialog(role.Pet),
					failureControlUnitSelectorDialog(role.Pet),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			petAttackButton.Importance = widget.WarningImportance

			petHangButton = widget.NewButton(pet.Hang.String(), func() {
				updateActionState(func(actionState *battle.ActionState) { actionState.AddPetAction(pet.Hang) })
				refreshActionViewer()
			})
			petHangButton.Importance = widget.SuccessImportance

			petSkillButton = widget.NewButton(pet.Skill.String(), func() {
				updateActionState(func(actionState *battle.ActionState) { actionState.AddPetAction(pet.Skill) })

				dialogs := []func(){
					offsetSelectorDialog(role.Pet),
					successControlUnitSelectorDialog(role.Pet),
					failureControlUnitSelectorDialog(role.Pet),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			petSkillButton.Importance = widget.HighImportance

			petDefendButton = widget.NewButton(pet.Defend.String(), func() {
				updateActionState(func(actionState *battle.ActionState) { actionState.AddPetAction(pet.Defend) })

				dialogs := []func(){
					offsetSelectorDialog(role.Pet),
					successControlUnitSelectorDialog(role.Pet),
					failureControlUnitSelectorDialog(role.Pet),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			petDefendButton.Importance = widget.HighImportance

			petHealSelfButton = widget.NewButton(pet.HealSelf.String(), func() {
				updateActionState(func(actionState *battle.ActionState) { actionState.AddPetAction(pet.HealSelf) })

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
				updateActionState(func(actionState *battle.ActionState) { actionState.AddPetAction(pet.HealOne) })

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
				updateActionState(func(actionState *battle.ActionState) { actionState.AddPetAction(pet.Ride) })

				dialogs := []func(){
					offsetSelectorDialog(role.Pet),
					successControlUnitSelectorDialog(role.Pet),
					failureControlUnitSelectorDialog(role.Pet),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			petRideButton.Importance = widget.HighImportance

			petOffRideButton = widget.NewButton(pet.OffRide.String(), func() {
				updateActionState(func(actionState *battle.ActionState) { actionState.AddPetAction(pet.OffRide) })

				dialogs := []func(){
					offsetSelectorDialog(role.Pet),
					successControlUnitSelectorDialog(role.Pet),
					failureControlUnitSelectorDialog(role.Pet),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			petOffRideButton.Importance = widget.HighImportance

			petEscapeButton = widget.NewButton(pet.Escape.String(), func() {
				updateActionState(func(actionState *battle.ActionState) { actionState.AddPetAction(pet.Escape) })

				dialogs := []func(){
					failureControlUnitSelectorDialog(role.Pet),
				}
				activateDialogs(dialogs, selectorDialogEnableChan)
			})
			petEscapeButton.Importance = widget.WarningImportance

			petCatchButton = widget.NewButton(pet.Catch.String(), func() {
				updateActionState(func(actionState *battle.ActionState) { actionState.AddPetAction(pet.Catch) })
				refreshActionViewer()

				notifyLogConfig("Catch Setup")

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

			actionsDialog := dialog.NewCustom("Add Pet Actions in Order", "Close", actionsContainer, window)
			actionsDialog.Show()

		})
		characterActionSelector.Importance = widget.MediumImportance
		petActionSelector.Importance = widget.MediumImportance

		loadSettingButton := widget.NewButtonWithIcon("Load", theme.FolderOpenIcon(), func() {
			fileOpenDialog := dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
				if err != nil {
					showErrorMessage(actionConfigSelectionError)
					return
				}
				if uc == nil {
					return
				}

				actionState, err := loadActionConfiguration(uc)
				if err != nil {
					showErrorMessage(actionConfigLoadError)
					return
				}
				worker.ReplaceActionState(actionState)
				refreshActionViewer()
			}, window)

			listableURI, _ := storage.ListerForURI(storage.NewFileURI(r.actionDir + `\actions`))
			fileOpenDialog.SetLocation(listableURI)
			fileOpenDialog.SetFilter(storage.NewExtensionFileFilter([]string{".ac"}))
			fileOpenDialog.Show()
		})
		loadSettingButton.Importance = widget.MediumImportance

		saveSettingButton := widget.NewButtonWithIcon("Save", theme.DownloadIcon(), func() {
			fileSaveDialog := dialog.NewFileSave(func(uc fyne.URIWriteCloser, err error) {
				if err != nil {
					showErrorMessage(actionConfigDestinationError)
					return
				}
				if uc == nil {
					return
				}

				if err := saveActionConfiguration(uc, worker.ActionStateSnapshot()); err != nil {
					showErrorMessage(actionConfigSaveError)
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
		workerMenuContainer.Add(enemyOrderButton)
		workerMenuContainer.Add(characterActionSelector)
		workerMenuContainer.Add(petActionSelector)
		workerMenuContainer.Add(loadSettingButton)
		workerMenuContainer.Add(saveSettingButton)

		actionViewer = container.NewHBox(generateTags(worker.ActionStateSnapshot())...)
		actionViewers = append(actionViewers, actionViewer)

		actionViewerScroll := container.NewHScroll(actionViewer)
		workerContainer := container.NewVBox(workerMenuContainer, actionViewerScroll)
		gameWidget.Add(workerContainer)
	}

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

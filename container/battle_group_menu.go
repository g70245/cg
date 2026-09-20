package container

import (
	"cg/game/battle"
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type menuWidgetOptions struct {
	partyState       *battle.PartyState
	vitalsMonitor    *battle.VitalsMonitor
	customEnemyOrder []string
	workers          battle.Workers
	sharedStopChan   chan bool
	actionViewers    []*fyne.Container
	destroy          func()
	restoreFullView  func()
}

type battleGroupMenu struct {
	container     *fyne.Container
	fullObjects   []fyne.CanvasObject
	switchButton  *widget.Button
	restoreButton *widget.Button
}

func newBattleGroupMenu(fullObjects []fyne.CanvasObject, switchButton, restoreButton *widget.Button) *battleGroupMenu {
	return &battleGroupMenu{
		container:     container.NewGridWithColumns(5, fullObjects...),
		fullObjects:   fullObjects,
		switchButton:  switchButton,
		restoreButton: restoreButton,
	}
}

func (menu *battleGroupMenu) setCompact(compact bool) {
	if compact {
		menu.container.Layout = layout.NewBorderLayout(nil, nil, nil, menu.restoreButton)
		menu.container.Objects = []fyne.CanvasObject{menu.switchButton, menu.restoreButton}
	} else {
		menu.container.Layout = layout.NewGridLayoutWithColumns(5)
		menu.container.Objects = menu.fullObjects
	}
	menu.container.Refresh()
}

func generateMenuWidget(options menuWidgetOptions) *battleGroupMenu {
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
			for i := range options.workers {
				options.workers[i].ReplaceActionState(actionState)
				options.actionViewers[i].Objects = generateTags(options.workers[i].ActionStateSnapshot())
				options.actionViewers[i].Refresh()
			}
		}, window)

		listableURI, _ := storage.ListerForURI(storage.NewFileURI(r.actionDir + `\actions`))
		fileOpenDialog.SetLocation(listableURI)
		fileOpenDialog.SetFilter(storage.NewExtensionFileFilter([]string{".ac"}))
		fileOpenDialog.Show()
	})
	loadSettingButton.Importance = widget.HighImportance

	deleteButton := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		deleteDialog := dialog.NewConfirm("Delete Battle Group", "Delete this battle group? All running tasks in the group will stop.", func(isDeleting bool) {
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
			options.vitalsMonitor.Reset()
			started := false
			for i := range options.workers {
				if options.workers[i].Work() {
					started = true
				}
			}
			if started {
				turn(theme.MediaStopIcon(), switchButton)
			}
		case theme.MediaStopIcon():
			for i := range options.workers {
				options.workers[i].Stop()
			}
			turn(theme.MediaPlayIcon(), switchButton)
		}
	})
	switchButton.Importance = widget.WarningImportance
	restoreButton := widget.NewButtonWithIcon("", theme.ViewFullScreenIcon(), options.restoreFullView)

	var teleportAndResourceCheckerButton *widget.Button
	teleportAndResourceCheckerButton = widget.NewButtonWithIcon("Teleport / Lure", theme.CheckButtonIcon(), func() {
		switch teleportAndResourceCheckerButton.Icon {
		case theme.CheckButtonCheckedIcon():
			for i := range options.workers {
				options.workers[i].StopTeleportAndResourceChecker()
			}
			turn(theme.CheckButtonIcon(), teleportAndResourceCheckerButton)
		case theme.CheckButtonIcon():
			if !validateLogConfig("Teleport and Lure Monitoring") {
				return
			}
			for i := range options.workers {
				options.workers[i].StartTeleportAndResourceChecker()
			}
			turn(theme.CheckButtonCheckedIcon(), teleportAndResourceCheckerButton)

			notifyBeeperAndLogConfig("Teleport and Lure Monitoring")
		}
	})
	teleportAndResourceCheckerButton.Importance = widget.HighImportance
	var activitiesCheckerButton *widget.Button
	activitiesCheckerButton = widget.NewButtonWithIcon("Activities", theme.CheckButtonIcon(), func() {
		switch activitiesCheckerButton.Icon {
		case theme.CheckButtonCheckedIcon():
			for i := range options.workers {
				options.workers[i].SetActivityCheckerEnabled(false)
			}
			turn(theme.CheckButtonIcon(), activitiesCheckerButton)
		case theme.CheckButtonIcon():
			if !validateLogConfig("Activity Monitoring") {
				return
			}
			for i := range options.workers {
				options.workers[i].SetActivityCheckerEnabled(true)
			}
			turn(theme.CheckButtonCheckedIcon(), activitiesCheckerButton)

			notifyBeeperAndLogConfig("Activity Monitoring")
		}
	})
	activitiesCheckerButton.Importance = widget.HighImportance
	var levelOneCheckerButton *widget.Button
	levelOneCheckerButton = widget.NewButtonWithIcon("Level 1", theme.CheckButtonIcon(), func() {
		switch levelOneCheckerButton.Icon {
		case theme.CheckButtonCheckedIcon():
			for i := range options.workers {
				options.workers[i].SetLevelOneCheckerEnabled(false)
			}
			turn(theme.CheckButtonIcon(), levelOneCheckerButton)
		case theme.CheckButtonIcon():
			if !validateBeeperConfig("Level 1 Monitoring") {
				return
			}
			for i := range options.workers {
				options.workers[i].SetLevelOneCheckerEnabled(true)
			}
			turn(theme.CheckButtonCheckedIcon(), levelOneCheckerButton)
		}
	})
	levelOneCheckerButton.Importance = widget.HighImportance
	var flawlessPetCheckerButton *widget.Button
	flawlessPetCheckerButton = widget.NewButtonWithIcon("Flawless Pet", theme.CheckButtonIcon(), func() {
		switch flawlessPetCheckerButton.Icon {
		case theme.CheckButtonCheckedIcon():
			for i := range options.workers {
				options.workers[i].SetFlawlessPetCheckerEnabled(false)
			}
			turn(theme.CheckButtonIcon(), flawlessPetCheckerButton)
		case theme.CheckButtonIcon():
			if !validateBeeperConfig("Flawless Pet Monitoring") {
				return
			}
			for i := range options.workers {
				options.workers[i].SetFlawlessPetCheckerEnabled(true)
			}
			turn(theme.CheckButtonCheckedIcon(), flawlessPetCheckerButton)
		}
	})
	flawlessPetCheckerButton.Importance = widget.HighImportance
	var partyButton *widget.Button
	partyButton = widget.NewButtonWithIcon("Party", theme.CheckButtonIcon(), func() {
		enabled := partyButton.Icon == theme.CheckButtonIcon()
		options.partyState.SetEnabled(enabled)
		if enabled {
			turn(theme.CheckButtonCheckedIcon(), partyButton)
		} else {
			turn(theme.CheckButtonIcon(), partyButton)
		}
	})
	partyButton.Importance = widget.HighImportance
	mpCheckerButton := newRatioMonitorButton("MP", "MP Monitoring", battle.MPRatios.GetOptions(), options.vitalsMonitor.SetCharacterMana)
	petMPCheckerButton := newRatioMonitorButton("Pet MP", "Pet MP Monitoring", battle.MPRatios.GetOptions(), options.vitalsMonitor.SetPetMana)
	healthCheckerButton := newRatioMonitorButton("HP", "HP Monitoring", battle.Ratios.GetOptions(), options.vitalsMonitor.SetCharacterHealth)
	petHealthCheckerButton := newRatioMonitorButton("Pet HP", "Pet HP Monitoring", battle.Ratios.GetOptions(), options.vitalsMonitor.SetPetHealth)
	var inventoryCheckerButton *widget.Button
	inventoryCheckerButton = widget.NewButtonWithIcon("Inventory", theme.CheckButtonIcon(), func() {
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

			notifyBeeperConfig("Inventory Monitoring")
		}
	})
	inventoryCheckerButton.Importance = widget.HighImportance
	monitoringDialog := dialog.NewCustom("Monitoring", "Close", container.NewGridWithColumns(5,
		partyButton, teleportAndResourceCheckerButton, inventoryCheckerButton, levelOneCheckerButton, flawlessPetCheckerButton,
		mpCheckerButton, petMPCheckerButton, healthCheckerButton, petHealthCheckerButton, activitiesCheckerButton,
	), window)
	checkersButton := widget.NewButtonWithIcon("Monitoring", theme.MenuIcon(), func() {
		monitoringDialog.Show()
		monitoringDialog.Resize(monitoringDialog.MinSize())
	})
	checkersButton.Importance = widget.HighImportance

	enemyOrderBindingStr := binding.NewString()
	enemyOrderCheckGroup := widget.NewCheckGroup(battle.EnemyPositions.GetOptions(), func(s []string) {
		enemyOrderBindingStr.Set(strings.Join(s, separator))
	})
	enemyOrderCheckGroup.Horizontal = true
	enemyOrderLabel := widget.NewLabelWithData(enemyOrderBindingStr)
	enemyOrderButton := widget.NewButtonWithIcon("Target Priority", theme.SearchIcon(), func() {
		tempSelected := make([]string, len(enemyOrderCheckGroup.Selected))
		copy(tempSelected, enemyOrderCheckGroup.Selected)

		d := dialog.NewCustom("Target Priority", "Cancel", container.NewVBox(enemyOrderCheckGroup, enemyOrderLabel), window)

		applyButton := widget.NewButton("Apply", func() {
			for i := range options.workers {
				options.workers[i].SetCustomEnemyOrder(enemyOrderCheckGroup.Selected)
			}
			d.Hide()
		})

		leaveButton := widget.NewButton("Cancel", func() {
			enemyOrderCheckGroup.Selected = tempSelected
			enemyOrderBindingStr.Set(strings.Join(enemyOrderCheckGroup.Selected, separator))
			d.Hide()
		})

		d.SetButtons([]fyne.CanvasObject{applyButton, leaveButton})
		d.Show()
	})
	enemyOrderButton.Importance = widget.HighImportance

	fullObjects := []fyne.CanvasObject{checkersButton, enemyOrderButton, loadSettingButton, deleteButton, switchButton}
	return newBattleGroupMenu(fullObjects, switchButton, restoreButton)
}

func newRatioMonitorButton(label, title string, ratioOptions []string, set func(bool, float32)) *widget.Button {
	var ratio float32
	ratioSelector := widget.NewRadioGroup(ratioOptions, nil)
	ratioSelector.Horizontal = true
	ratioSelector.Required = true

	var button *widget.Button
	ratioDialog := dialog.NewCustomConfirm(title, "Apply", "Cancel", ratioSelector, func(apply bool) {
		if !apply || ratioSelector.Selected == "" {
			return
		}
		value, err := strconv.ParseFloat(ratioSelector.Selected, 32)
		if err != nil {
			return
		}
		ratio = float32(value)
		set(true, ratio)
		button.SetText(ratioMonitorButtonText(label, ratio))
		turn(theme.CheckButtonCheckedIcon(), button)
	}, window)

	button = widget.NewButtonWithIcon(label, theme.CheckButtonIcon(), func() {
		switch button.Icon {
		case theme.CheckButtonCheckedIcon():
			set(false, ratio)
			turn(theme.CheckButtonIcon(), button)
		case theme.CheckButtonIcon():
			if !validateBeeperConfig(title) {
				return
			}
			ratioDialog.Show()
			ratioDialog.Resize(ratioDialog.MinSize())
		}
	})
	button.Importance = widget.HighImportance
	return button
}

func ratioMonitorButtonText(label string, ratio float32) string {
	return fmt.Sprintf("%s: %.0f%%", label, ratio*100)
}

func turn(icon fyne.Resource, button *widget.Button) {
	button.SetIcon(icon)
	button.Refresh()
}

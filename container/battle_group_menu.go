package container

import (
	"cg/game"
	"cg/game/battle"
	"fmt"
	"sort"
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
	games            game.Games
	allGames         game.Games
	manaChecker      *battle.ManaChecker
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

func currentManaCheckerOptions(games, allGames game.Games) []string {
	aliases := make([]string, 0, len(games))
	for _, hWnd := range games.GetHWNDs() {
		if alias := allGames.FindKey(hWnd); alias != "" {
			aliases = append(aliases, alias)
		}
	}
	sort.Strings(aliases)

	return append([]string{battle.NO_MANA_CHECKER}, aliases...)
}

func currentManaCheckerAlias(manaChecker *battle.ManaChecker, allGames game.Games) string {
	selectedHandle := manaChecker.Get()
	if selectedHandle == battle.NO_MANA_CHECKER {
		return battle.NO_MANA_CHECKER
	}

	for alias, hWnd := range allGames {
		if fmt.Sprint(hWnd) == selectedHandle {
			return alias
		}
	}
	return battle.NO_MANA_CHECKER
}

func generateMenuWidget(options menuWidgetOptions) *battleGroupMenu {
	var manaCheckerSelectorDialog *dialog.CustomDialog
	var manaCheckerSelectorButton *widget.Button
	manaCheckerOptions := currentManaCheckerOptions(options.games, options.allGames)
	manaCheckerSelector := widget.NewRadioGroup(manaCheckerOptions, func(s string) {
		if hWnd, ok := options.allGames[s]; ok {
			options.manaChecker.Set(fmt.Sprint(hWnd))
			manaCheckerSelectorButton.SetText(fmt.Sprintf("Mana: %s", s))
		} else {
			options.manaChecker.Set(battle.NO_MANA_CHECKER)
			manaCheckerSelectorButton.SetText(fmt.Sprintf("Mana: %s", options.manaChecker.Get()))
		}
		manaCheckerSelectorDialog.Hide()
	})
	manaCheckerSelector.Required = true
	manaCheckerSelectorDialog = dialog.NewCustomWithoutButtons("Select Game for Mana Monitoring", manaCheckerSelector, window)
	manaCheckerSelectorButton = widget.NewButton(fmt.Sprintf("Mana: %s", options.manaChecker.Get()), func() {
		manaCheckerSelectorDialog.Show()

		notifyBeeperConfig("Mana Monitoring Setup")
	})
	manaCheckerSelectorButton.Importance = widget.HighImportance
	refreshManaCheckerSelector := func() {
		selectedAlias := currentManaCheckerAlias(options.manaChecker, options.allGames)
		manaCheckerSelector.Options = currentManaCheckerOptions(options.games, options.allGames)
		manaCheckerSelector.Selected = selectedAlias
		manaCheckerSelectorButton.SetText(fmt.Sprintf("Mana: %s", selectedAlias))
		manaCheckerSelector.Refresh()
	}

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
	teleportAndResourceCheckerButton = widget.NewButtonWithIcon("Teleport / Resources", theme.CheckButtonIcon(), func() {
		switch teleportAndResourceCheckerButton.Icon {
		case theme.CheckButtonCheckedIcon():
			for i := range options.workers {
				options.workers[i].StopTeleportAndResourceChecker()
			}
			turn(theme.CheckButtonIcon(), teleportAndResourceCheckerButton)
		case theme.CheckButtonIcon():
			if !validateLogConfig("Teleport and Resource Monitoring") {
				return
			}
			for i := range options.workers {
				options.workers[i].StartTeleportAndResourceChecker()
			}
			turn(theme.CheckButtonCheckedIcon(), teleportAndResourceCheckerButton)

			notifyBeeperAndLogConfig("Teleport and Resource Monitoring")
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
	monitoringDialog := dialog.NewCustom("Monitoring", "Close", container.NewGridWithColumns(4, manaCheckerSelectorButton, teleportAndResourceCheckerButton, activitiesCheckerButton, inventoryCheckerButton), window)
	checkersButton := widget.NewButtonWithIcon("Monitoring", theme.MenuIcon(), func() {
		refreshManaCheckerSelector()
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

func turn(icon fyne.Resource, button *widget.Button) {
	button.SetIcon(icon)
	button.Refresh()
}

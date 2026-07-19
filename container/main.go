package container

import (
	"cg/game"
	"cg/utils"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var (
	window fyne.Window
	r      robot
)

const compactWindowMinimumWidth float32 = 360

type robot struct {
	main                   *fyne.Container
	gameDirMu              sync.RWMutex
	gameDir                string
	actionDir              string
	width                  float32
	height                 float32
	games                  game.Games
	close                  func()
	battleCompactButton    *widget.Button
	onBattleCompactChanged func(bool, fyne.CanvasObject)
}

func App(title, gameDir string, width, height float32) {
	cg := app.New()
	window = cg.NewWindow(title)
	window.Resize(fyne.NewSize(width, height))

	var content *fyne.Container
	battleCompactButton := widget.NewButtonWithIcon("", theme.ViewRestoreIcon(), nil)
	r = robot{
		games:               game.NewGames(),
		gameDir:             gameDir,
		actionDir:           gameDir,
		width:               width,
		height:              height,
		battleCompactButton: battleCompactButton,
		onBattleCompactChanged: func(compact bool, battleContent fyne.CanvasObject) {
			if compact {
				compactContent := container.NewPadded(battleContent)
				window.SetContent(compactContent)
				compactSize := compactContent.MinSize()
				compactSize.Width = fyne.Max(compactSize.Width, compactWindowMinimumWidth)
				window.Resize(compactSize)
				return
			}

			window.SetContent(content)
			window.Resize(fyne.NewSize(width, height))
		},
	}
	r.generateRobotContainer()

	refreshButton := widget.NewButtonWithIcon("Refresh Games", theme.ViewRefreshIcon(), func() {
		refreshDialog := dialog.NewConfirm("Refresh Games", "Refresh the game list? All running tasks will stop.", func(isRefreshing bool) {
			if isRefreshing {
				content.Remove(r.main)
				r.close()

				r.refresh()
				content.Add(r.main)
				window.SetContent(window.Content())
				window.Resize(fyne.NewSize(width, height))
			}
		}, window)
		refreshDialog.SetConfirmImportance(widget.DangerImportance)
		refreshDialog.Show()
	})
	refreshButton.Importance = widget.DangerImportance

	listableURI, _ := storage.ListerForURI(storage.NewFileURI(gameDir))
	var alertDialogButton *widget.Button
	alertDialogButton = widget.NewButtonWithIcon("Alert Music", theme.FolderIcon(), func() {
		alertDialog := dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
			if err != nil {
				showErrorMessage(alertMusicSelectionError)
				return
			}
			if uc == nil {
				return
			}
			if err := utils.Beeper.Init(uc.URI().Path()); err != nil {
				alertDialogButton.SetIcon(theme.FolderIcon())
				showErrorMessage(alertMusicInitializationError)
				return
			}
			alertDialogButton.SetIcon(theme.MediaMusicIcon())
		}, window)
		alertDialog.SetLocation(listableURI)
		alertDialog.SetFilter(storage.NewExtensionFileFilter([]string{".mp3"}))
		alertDialog.Show()
	})
	alertDialogButton.Importance = widget.HighImportance

	var gameDirDialogButton *widget.Button
	gameDirDialogButton = widget.NewButtonWithIcon("Game Folder", theme.FolderIcon(), func() {
		gameDirDialog := dialog.NewFolderOpen(func(lu fyne.ListableURI, err error) {
			if err != nil {
				showErrorMessage(gameFolderSelectionError)
				return
			}
			if lu == nil {
				return
			}
			r.setGameDir(lu.Path())
			gameDirDialogButton.SetIcon(theme.FolderOpenIcon())
		}, window)
		gameDirDialog.SetLocation(listableURI)
		gameDirDialog.Show()
	})
	gameDirDialogButton.Importance = widget.HighImportance

	menuButtons := container.NewHBox(refreshButton, alertDialogButton, gameDirDialogButton, battleCompactButton)
	menu := container.NewBorder(nil, nil, nil, nil, menuButtons)
	content = container.NewBorder(menu, nil, nil, nil, r.main)

	/* shortcuts */
	muteShortcut := &desktop.CustomShortcut{KeyName: fyne.Key0, Modifier: fyne.KeyModifierControl}
	window.Canvas().AddShortcut(muteShortcut, func(shortcut fyne.Shortcut) {
		utils.Beeper.Stop()
	})

	window.SetContent(content)
	window.ShowAndRun()
}

func (r *robot) getGameDir() string {
	r.gameDirMu.RLock()
	defer r.gameDirMu.RUnlock()
	return r.gameDir
}

func (r *robot) setGameDir(gameDir string) {
	r.gameDirMu.Lock()
	r.gameDir = gameDir
	r.gameDirMu.Unlock()
}

func (r *robot) refresh() {
	r.games.AddGames(game.NewGames())
	r.generateRobotContainer()
}

func (r *robot) generateRobotContainer() {

	autoBattleContainer, autoBattleGroups := newBattleContainer(r.games, r.battleCompactButton, r.onBattleCompactChanged)
	productionContainer, productionWorkers := newProductionContainer(r.games)

	battleTab := container.NewTabItem("Battle", autoBattleContainer)
	productionTab := container.NewTabItem("Production", productionContainer)
	tabs := container.NewAppTabs(
		battleTab,
		productionTab,
	)
	tabs.OnSelected = func(selected *container.TabItem) {
		if selected == battleTab {
			r.battleCompactButton.Show()
		} else {
			r.battleCompactButton.Hide()
		}
		window.Content().Refresh()
	}

	tabs.SetTabLocation(container.TabLocationBottom)
	r.battleCompactButton.Show()
	main := container.NewStack(tabs)

	r.main = main
	r.close = func() {
		main.RemoveAll()
		autoBattleGroups.stopAll()
		productionWorkers.stopAll()
	}
}

package containers

import (
	"cg/game"
	"cg/utils"

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

type robot struct {
	main    *fyne.Container
	gameDir *string
	width   float32
	height  float32
	games   game.Games
	close   func()
}

func App(title, gameDir string, width, height float32) {
	cg := app.New()
	window = cg.NewWindow(title)
	window.Resize(fyne.NewSize(width, height))

	r = robot{
		games:   game.NewGames(),
		gameDir: new(string),
		width:   width,
		height:  height,
	}
	r.generateRobotContainer()

	var content *fyne.Container
	refreshButton := widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), func() {
		refreshDialog := dialog.NewConfirm("Refresh games", "Do you really want to refresh?", func(isRefreshing bool) {
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
			if uc != nil {
				utils.Beeper.Init(uc.URI().Path())
				alertDialogButton.SetIcon(theme.MediaMusicIcon())
			} else {
				utils.Beeper.Close()
				alertDialogButton.SetIcon(theme.FolderIcon())
			}
		}, window)
		alertDialog.SetLocation(listableURI)
		alertDialog.SetFilter(storage.NewExtensionFileFilter([]string{".mp3"}))
		alertDialog.Show()
	})
	alertDialogButton.Importance = widget.HighImportance

	var gameDirDialogButton *widget.Button
	gameDirDialogButton = widget.NewButtonWithIcon("Game Directory", theme.FolderIcon(), func() {
		gameDirDialog := dialog.NewFolderOpen(func(lu fyne.ListableURI, err error) {
			if lu != nil {
				*r.gameDir = lu.Path()
				gameDirDialogButton.SetIcon(theme.FolderOpenIcon())
			} else {
				*r.gameDir = ""
				gameDirDialogButton.SetIcon(theme.FolderIcon())
			}
		}, window)
		gameDirDialog.SetLocation(listableURI)
		gameDirDialog.Show()
	})
	gameDirDialogButton.Importance = widget.HighImportance

	menu := container.NewHBox(refreshButton, alertDialogButton, gameDirDialogButton)
	content = container.NewBorder(menu, nil, nil, nil, r.main)

	/* shortcuts */
	muteShortcut := &desktop.CustomShortcut{KeyName: fyne.Key0, Modifier: fyne.KeyModifierControl}
	window.Canvas().AddShortcut(muteShortcut, func(shortcut fyne.Shortcut) {
		utils.Beeper.Stop()
	})

	window.SetContent(content)
	window.ShowAndRun()
}

func (r *robot) refresh() {
	r.games.AddGames(game.NewGames())
	r.generateRobotContainer()
}

func (r *robot) generateRobotContainer() {

	autoBattleContainer, autoBattleGroups := newBattleContainer(r.games)
	productionContainer, productionWorkers := newProductionContainer(r.games)

	tabs := container.NewAppTabs(
		container.NewTabItem("Battle", autoBattleContainer),
		container.NewTabItem("Produce", productionContainer),
	)

	tabs.SetTabLocation(container.TabLocationBottom)
	main := container.NewStack(tabs)

	r.main = main
	r.close = func() {
		main.RemoveAll()
		autoBattleGroups.stopAll()
		productionWorkers.stopAll()
	}
}

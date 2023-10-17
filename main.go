package main

import (
	"cg/system"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/exp/maps"
)

const (
	APP_NAME     = "CG"
	APP_WIDTH    = 960
	APP_HEIGHT   = 420
	DEFAULT_ROOT = `D:\CG`
)

var (
	window fyne.Window
	logDir = new(string)
)

func main() {
	cg := app.New()
	window = cg.NewWindow(APP_NAME)
	window.Resize(fyne.NewSize(APP_WIDTH, APP_HEIGHT))

	var content *fyne.Container

	robot := generateRobotContainer()
	refreshButton := widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), func() {
		refreshDialog := dialog.NewConfirm("Refresh games", "Do you really want to refresh?", func(isRefreshing bool) {
			if isRefreshing {
				content.Remove(robot.main)
				robot.close()

				robot = generateRobotContainer()
				content.Add(robot.main)
				window.SetContent(window.Content())
				window.Resize(fyne.NewSize(APP_WIDTH, APP_HEIGHT))
			}
		}, window)
		refreshDialog.SetConfirmImportance(widget.DangerImportance)
		refreshDialog.Show()
	})
	refreshButton.Importance = widget.DangerImportance

	listableURI, _ := storage.ListerForURI(storage.NewFileURI(DEFAULT_ROOT))
	var alertDialogButton *widget.Button
	alertDialogButton = widget.NewButtonWithIcon("Alert Music", theme.FolderIcon(), func() {
		alertDialog := dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
			if uc != nil {
				system.CloseBeeper()
				system.CreateBeeper(uc.URI().Path())
				alertDialogButton.SetIcon(theme.MediaMusicIcon())
			} else {
				system.CloseBeeper()
				alertDialogButton.SetIcon(theme.FolderIcon())
			}
		}, window)
		alertDialog.SetLocation(listableURI)
		alertDialog.SetFilter(storage.NewExtensionFileFilter([]string{".mp3"}))
		alertDialog.Show()
	})
	alertDialogButton.Importance = widget.HighImportance

	var logDialogButton *widget.Button
	logDialogButton = widget.NewButtonWithIcon("Log Directory", theme.FolderIcon(), func() {
		logDirDialog := dialog.NewFolderOpen(func(lu fyne.ListableURI, err error) {
			if lu != nil {
				*logDir = lu.Path()
				logDialogButton.SetIcon(theme.FolderOpenIcon())
			} else {
				*logDir = ""
				logDialogButton.SetIcon(theme.FolderIcon())
			}
		}, window)
		logDirDialog.SetLocation(listableURI)
		logDirDialog.Show()
	})
	logDialogButton.Importance = widget.HighImportance

	menu := container.NewHBox(refreshButton, alertDialogButton, logDialogButton)
	content = container.NewBorder(menu, nil, nil, nil, robot.main)

	/* shortcuts */
	muteShortcut := &desktop.CustomShortcut{KeyName: fyne.Key0, Modifier: fyne.KeyModifierControl}
	window.Canvas().AddShortcut(muteShortcut, func(shortcut fyne.Shortcut) {
		system.StopBeeper()
	})

	window.SetContent(content)
	window.ShowAndRun()
}

type Robot struct {
	main  *fyne.Container
	close func()
}

func generateRobotContainer() Robot {
	idleGames := system.FindWindows()

	autoBattleWidget, autoBattleStopChans := battleContainer(idleGames)

	tabs := container.NewAppTabs(
		container.NewTabItem("Auto Battle", autoBattleWidget),
	)

	tabs.SetTabLocation(container.TabLocationTop)
	main := container.NewStack(tabs)
	robot := Robot{main, func() {
		main.RemoveAll()
		stopWorkers(maps.Values(autoBattleStopChans))
	}}
	return robot
}

func stopWorkers(stopChans []chan bool) {
	for _, stopChan := range stopChans {
		stop(stopChan)
	}
}

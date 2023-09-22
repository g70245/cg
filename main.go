package main

import (
	"cg/system"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/exp/maps"
)

const (
	TARGET_CLASS = "Blue"
	APP_NAME     = "CG"
	APP_WIDTH    = 960
	APP_HEIGHT   = 460
)

var (
	window fyne.Window
	logDir = new(string)
)

func test() {
	// hWnd := getHWND()
	// x, y, _ := game.Test(hWnd)
	// system.MoveMouse(hWnd, x, y)
	// CheckColor(
	// 	hWnd,
	// 	x,
	// 	y,
	// 	x+250,
	// 	y+200,
	// 	8948665,
	// )
	os.Exit(0)
}

func main() {
	// test()

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

	var alertDialogButton *widget.Button
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
	alertDialogButton = widget.NewButtonWithIcon("Alert Music", theme.FolderIcon(), func() {
		alertDialog.Show()
	})
	alertDialogButton.Importance = widget.HighImportance

	var logDialogButton *widget.Button
	logDirDialog := dialog.NewFolderOpen(func(lu fyne.ListableURI, err error) {
		if lu != nil {
			*logDir = lu.Path()
			logDialogButton.SetIcon(theme.FolderOpenIcon())
		} else {
			logDialogButton.SetIcon(theme.FolderIcon())
		}
	}, window)
	logDialogButton = widget.NewButtonWithIcon("Log Directory", theme.FolderIcon(), func() {
		logDirDialog.Show()
	})
	logDialogButton.Importance = widget.HighImportance

	menu := container.NewHBox(refreshButton, alertDialogButton, logDialogButton)
	content = container.NewBorder(menu, nil, nil, nil, robot.main)
	window.SetContent(content)
	window.ShowAndRun()
}

type Robot struct {
	main  *fyne.Container
	close func()
}

func generateRobotContainer() Robot {
	idleGames := system.FindWindows(TARGET_CLASS)

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

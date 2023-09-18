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
	APP_WIDTH    = 761
	APP_HEIGHT   = 428
)

var window fyne.Window
var logDir = new(string)

func test() {
	// hWnd := getHWND()
	os.Exit(0)
}

func main() {
	// test()

	cg := app.New()
	window = cg.NewWindow(APP_NAME)
	window.Resize(fyne.NewSize(APP_WIDTH, APP_HEIGHT))
	window.Content().Move(fyne.NewPos(600, 600))

	var content *fyne.Container

	robot := generateRobotContainer()
	refreshButton := widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), func() {
		content.Remove(robot.main)
		robot.close()

		robot = generateRobotContainer()

		content.Add(robot.main)
		window.SetContent(window.Content())
		window.Resize(fyne.NewSize(APP_WIDTH, APP_HEIGHT))
	})
	refreshButton.Importance = widget.DangerImportance

	var logDialogButton *widget.Button
	logDirDialog := dialog.NewFolderOpen(func(lu fyne.ListableURI, err error) {
		if lu != nil {
			*logDir = lu.Path()
			logDialogButton.SetText(*logDir)
			logDialogButton.SetIcon(theme.FolderOpenIcon())
		}
	}, window)
	logDialogButton = widget.NewButtonWithIcon("Choose Log Directory", theme.FolderIcon(), func() {
		logDirDialog.Show()
	})
	logDialogButton.Importance = widget.HighImportance

	// menu := container.NewGridWrap(fyne.NewSize(100, 30), refreshButton, logDialogButton)
	menu := container.NewHBox(refreshButton, logDialogButton)
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
		stopAll(maps.Values(autoBattleStopChans))
	}}
	return robot
}

func stopAll(stopChans []chan bool) {
	for _, stopChan := range stopChans {
		stop(stopChan)
	}
}

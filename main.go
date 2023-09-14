package main

import (
	"cg/system"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/exp/maps"
)

const (
	TARGET_CLASS = "Blue"
	APP_NAME     = "CG"
	APP_WIDTH    = 300
	APP_HEIGHT   = 260
	GAME_WIDTH   = 640
	GAME_HEIGHT  = 480
)

var window fyne.Window

func main() {
	// checkTargets := []CheckTarget{BATTLE_COMMAND_ATTACK}
	// PrintColorFromData(checkTargets)
	cg := app.New()
	window = cg.NewWindow(APP_NAME)
	window.Resize(fyne.NewSize(APP_WIDTH, APP_HEIGHT))

	var content *fyne.Container

	robot := generateRobotContainer()
	refreshButton := widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), func() {
		robot.close()
		robot = generateRobotContainer()
		content.Add(robot.main)
		content.Refresh()
	})
	menu := container.NewGridWrap(fyne.NewSize(100, 30), refreshButton)

	content = container.NewBorder(menu, nil, nil, nil, robot.main)
	window.SetContent(content)
	window.ShowAndRun()
	log.Println("Exit")
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

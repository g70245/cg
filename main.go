package main

import (
	"cg/system"

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
	APP_WIDTH    = 620
	APP_HEIGHT   = 380
)

var window fyne.Window

func main() {
	// y := 120
	// for y < 400 {
	// 	check := game.CheckTarget{}
	// 	check.Set(136, int32(y))
	// 	PrintColorFromData([]game.CheckTarget{check})
	// 	y += 16
	// }
	// os.Exit(0)
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

	menu := container.NewGridWrap(fyne.NewSize(100, 30), refreshButton)
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

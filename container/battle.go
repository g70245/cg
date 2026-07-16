package container

import (
	"cg/game"
	"cg/game/battle"
	"fmt"
	"sync"
	"sync/atomic"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type BattleGroups struct {
	stopChans map[int]chan bool
	views     map[int]*battleGroupView
	compact   bool
}

type battleGroupView struct {
	container   *fyne.Container
	menu        *battleGroupMenu
	fullObjects []fyne.CanvasObject
}

func newBattleGroupView(menu *battleGroupMenu, gameWidget *fyne.Container) *battleGroupView {
	fullObjects := []fyne.CanvasObject{widget.NewSeparator(), menu.container, widget.NewSeparator(), gameWidget}
	return &battleGroupView{
		container:   container.NewVBox(fullObjects...),
		menu:        menu,
		fullObjects: fullObjects,
	}
}

func (view *battleGroupView) setCompact(compact bool) {
	view.menu.setCompact(compact)
	if compact {
		view.container.Objects = []fyne.CanvasObject{view.menu.container}
	} else {
		view.container.Objects = view.fullObjects
	}
	view.container.Refresh()
}

func newBattleContainer(games game.Games, compactButton *widget.Button, onCompactChanged func(bool, fyne.CanvasObject)) (*fyne.Container, BattleGroups) {
	id := 0
	battleGroups := BattleGroups{
		stopChans: make(map[int]chan bool),
		views:     make(map[int]*battleGroupView),
	}

	groupTabs := container.NewAppTabs()
	groupTabs.SetTabLocation(container.TabLocationTop)
	groupTabs.Hide()

	var setCompact func(bool)
	newGroupButton := widget.NewButtonWithIcon("New Battle Group", theme.ContentAddIcon(), func() {
		groupNameEntry := widget.NewEntry()
		groupNameEntry.SetPlaceHolder("Group Name")

		gamesCheckGroup := widget.NewCheckGroup(games.GetSortedKeys(), nil)
		gamesCheckGroup.Horizontal = true

		gamesSelectorDialog := dialog.NewCustom("Select Games", "Create", container.NewVBox(groupNameEntry, gamesCheckGroup), window)
		gamesSelectorDialog.Resize(fyne.NewSize(240, 166))

		gamesSelectorDialog.SetOnClosed(func() {
			if len(gamesCheckGroup.Selected) == 0 {
				return
			}

			var newTabItem *container.TabItem
			newGroupView, stopChan := newBatttleGroupContainer(games.New(gamesCheckGroup.Selected), games, func(id int) func() {
				return func() {
					delete(battleGroups.stopChans, id)
					delete(battleGroups.views, id)

					groupTabs.Remove(newTabItem)
					if len(battleGroups.stopChans) == 0 {
						groupTabs.Hide()
						compactButton.Disable()
					}

					window.SetContent(window.Content())
					window.Resize(fyne.NewSize(r.width, r.height))
				}
			}(id), func() {
				setCompact(false)
			})
			battleGroups.stopChans[id] = stopChan
			battleGroups.views[id] = newGroupView
			newGroupView.setCompact(battleGroups.compact)
			compactButton.Enable()

			var newGroupName string
			if groupNameEntry.Text != "" {
				newGroupName = groupNameEntry.Text
			} else {
				newGroupName = "Group " + fmt.Sprint(id)
			}
			newTabItem = container.NewTabItem(newGroupName, newGroupView.container)
			groupTabs.Append(newTabItem)
			groupTabs.Select(newTabItem)
			groupTabs.Show()

			window.SetContent(window.Content())
			window.Resize(fyne.NewSize(r.width, r.height))
			id++
		})
		gamesSelectorDialog.Show()
	})

	newBattleContainer := container.NewBorder(nil, newGroupButton, nil, nil, groupTabs)
	setCompact = func(compact bool) {
		battleGroups.compact = compact
		for _, view := range battleGroups.views {
			view.setCompact(compact)
		}

		if compact {
			newGroupButton.Hide()
			compactButton.SetText("")
			compactButton.SetIcon(theme.ViewFullScreenIcon())
		} else {
			newGroupButton.Show()
			compactButton.SetText("")
			compactButton.SetIcon(theme.ViewRestoreIcon())
		}
		newBattleContainer.Refresh()

		if onCompactChanged != nil {
			onCompactChanged(compact, newBattleContainer)
		}
	}
	compactButton.SetText("")
	compactButton.SetIcon(theme.ViewRestoreIcon())
	compactButton.OnTapped = func() {
		setCompact(!battleGroups.compact)
	}
	compactButton.Disable()

	return newBattleContainer, battleGroups
}

func newBatttleGroupContainer(games game.Games, allGames game.Games, destroy, restoreFullView func()) (groupView *battleGroupView, sharedStopChan chan bool) {
	manaChecker := battle.NewManaChecker()
	sharedStopChan = make(chan bool, len(games))
	workers := battle.CreateWorkers(games, r.getGameDir, manaChecker, new(atomic.Bool), sharedStopChan, new(sync.WaitGroup))

	gameWidget, actionViewers := generateGameWidget(gameWidgeOptions{
		games:       games,
		allGames:    allGames,
		manaChecker: manaChecker,
		workers:     workers,
	})
	menu := generateMenuWidget(menuWidgetOptions{
		games:           games,
		allGames:        allGames,
		manaChecker:     manaChecker,
		workers:         workers,
		sharedStopChan:  sharedStopChan,
		actionViewers:   actionViewers,
		destroy:         destroy,
		restoreFullView: restoreFullView,
	})

	groupView = newBattleGroupView(menu, gameWidget)
	return groupView, sharedStopChan
}

func stop(stopChan chan bool) {
	i := 0
	for i < cap(stopChan) {
		stopChan <- true
		i++
	}
}

func (bgs *BattleGroups) stopAll() {
	for k := range bgs.stopChans {
		stop(bgs.stopChans[k])
	}
}

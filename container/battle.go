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
}

func newBattleContainer(games game.Games) (*fyne.Container, BattleGroups) {
	id := 0
	battleGroups := BattleGroups{make(map[int]chan bool)}

	groupTabs := container.NewAppTabs()
	groupTabs.SetTabLocation(container.TabLocationTop)
	groupTabs.Hide()

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
			newGroupContainer, stopChan := newBatttleGroupContainer(games.New(gamesCheckGroup.Selected), games, func(id int) func() {
				return func() {
					delete(battleGroups.stopChans, id)

					groupTabs.Remove(newTabItem)
					if len(battleGroups.stopChans) == 0 {
						groupTabs.Hide()
					}

					window.SetContent(window.Content())
					window.Resize(fyne.NewSize(r.width, r.height))
				}
			}(id))
			battleGroups.stopChans[id] = stopChan

			var newGroupName string
			if groupNameEntry.Text != "" {
				newGroupName = groupNameEntry.Text
			} else {
				newGroupName = "Group " + fmt.Sprint(id)
			}
			newTabItem = container.NewTabItem(newGroupName, newGroupContainer)
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

	return newBattleContainer, battleGroups
}

func newBatttleGroupContainer(games game.Games, allGames game.Games, destroy func()) (autoBattleWidget *fyne.Container, sharedStopChan chan bool) {
	manaChecker := battle.NewManaChecker()
	sharedStopChan = make(chan bool, len(games))
	workers := battle.CreateWorkers(games, r.getGameDir, manaChecker, new(atomic.Bool), sharedStopChan, new(sync.WaitGroup))

	gameWidget, actionViewers := generateGameWidget(gameWidgeOptions{
		games:       games,
		allGames:    allGames,
		manaChecker: manaChecker,
		workers:     workers,
	})
	menuWidget := generateMenuWidget(menuWidgetOptions{
		games:          games,
		allGames:       allGames,
		manaChecker:    manaChecker,
		workers:        workers,
		sharedStopChan: sharedStopChan,
		actionViewers:  actionViewers,
		destroy:        destroy,
	})

	autoBattleWidget = container.NewVBox(widget.NewSeparator(), menuWidget, widget.NewSeparator(), gameWidget)
	return autoBattleWidget, sharedStopChan
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

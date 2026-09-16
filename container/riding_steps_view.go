package container

import (
	"cg/game"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/g70245/win"
)

const ridingStepsUpdateInterval = time.Second

type ridingStepsTarget struct {
	alias string
	hWnd  win.HWND
}

type ridingStepsView struct {
	container fyne.CanvasObject
	status    binding.String
	readSteps func(win.HWND) (uint32, error)
	interval  time.Duration

	mu        sync.RWMutex
	targets   map[win.HWND]string
	visible   bool
	onVisible func(bool)
	wake      chan struct{}
	cancel    chan struct{}
	done      chan struct{}
	closeOnce sync.Once
}

func newRidingStepsView(games, allGames game.Games) *ridingStepsView {
	return newRidingStepsViewWith(games, allGames, game.ReadRidingSteps, ridingStepsUpdateInterval)
}

func newRidingStepsViewWith(games, allGames game.Games, readSteps func(win.HWND) (uint32, error), interval time.Duration) *ridingStepsView {
	status := binding.NewString()
	label := widget.NewLabelWithData(status)
	view := &ridingStepsView{
		container: container.NewHScroll(label),
		status:    status,
		readSteps: readSteps,
		interval:  interval,
		targets:   make(map[win.HWND]string, len(games)),
		wake:      make(chan struct{}, 1),
		cancel:    make(chan struct{}),
		done:      make(chan struct{}),
	}
	view.container.Hide()
	view.refreshAliases(games, allGames)
	go view.run()
	return view
}

func (view *ridingStepsView) refreshAliases(games, allGames game.Games) {
	targets := make(map[win.HWND]string, len(games))
	for groupAlias, hWnd := range games {
		alias := allGames.FindKey(hWnd)
		if alias == "" {
			alias = groupAlias
		}
		targets[hWnd] = alias
	}

	view.mu.Lock()
	view.targets = targets
	view.mu.Unlock()

	select {
	case view.wake <- struct{}{}:
	default:
	}
}

func (view *ridingStepsView) setVisibilityChanged(callback func(bool)) {
	view.mu.Lock()
	view.onVisible = callback
	visible := view.visible
	view.mu.Unlock()

	callback(visible)
}

func (view *ridingStepsView) run() {
	defer close(view.done)
	view.update()

	ticker := time.NewTicker(view.interval)
	defer ticker.Stop()
	for {
		select {
		case <-view.cancel:
			return
		case <-view.wake:
			view.update()
		case <-ticker.C:
			view.update()
		}
	}
}

func (view *ridingStepsView) update() {
	view.mu.RLock()
	targets := make([]ridingStepsTarget, 0, len(view.targets))
	for hWnd, alias := range view.targets {
		targets = append(targets, ridingStepsTarget{alias: alias, hWnd: hWnd})
	}
	view.mu.RUnlock()

	sort.Slice(targets, func(i, j int) bool {
		return targets[i].alias < targets[j].alias
	})

	values := make([]string, 0, len(targets))
	for _, target := range targets {
		steps, err := view.readSteps(target.hWnd)
		if err != nil {
			values = append(values, fmt.Sprintf("%s: ERR", target.alias))
			continue
		}
		if steps == 0 {
			continue
		}
		values = append(values, fmt.Sprintf("%s: %d", target.alias, steps))
	}

	if len(values) == 0 {
		_ = view.status.Set("")
		view.setVisible(false)
		return
	}

	_ = view.status.Set("R  " + strings.Join(values, " | "))
	view.setVisible(true)
}

func (view *ridingStepsView) isVisible() bool {
	view.mu.RLock()
	defer view.mu.RUnlock()
	return view.visible
}

func (view *ridingStepsView) setVisible(visible bool) {
	view.mu.Lock()
	if view.visible == visible {
		view.mu.Unlock()
		return
	}
	view.visible = visible
	callback := view.onVisible
	view.mu.Unlock()

	if visible {
		view.container.Show()
	} else {
		view.container.Hide()
	}
	if callback != nil {
		callback(visible)
	}
}

func (view *ridingStepsView) close() {
	view.closeOnce.Do(func() {
		close(view.cancel)
		<-view.done
	})
}

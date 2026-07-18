package container

import (
	"cg/game"
	"cg/game/navigation"
	"fmt"
	"image/color"
	"sort"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/g70245/win"
)

const (
	navigationOffOption      = "Navigation Off"
	navigationUpdateInterval = 500 * time.Millisecond
)

type navigationSnapshot struct {
	east, south int
	routes      []navigation.Route
}

type battleNavigationView struct {
	container         *fyne.Container
	controls          *fyne.Container
	selector          *widget.Select
	listFrame         *fyne.Container
	collapsedControls []fyne.CanvasObject
	expandedControls  []fyne.CanvasObject
	collapsedObjects  []fyne.CanvasObject
	expandedObjects   []fyne.CanvasObject

	position    binding.String
	status      binding.String
	statusLabel *widget.Label
	routeList   *widget.List
	scrollToTop func()
	routesMu    sync.RWMutex
	routes      []string

	games        game.Games
	allGames     game.Games
	gameDir      func() string
	readSnapshot func(win.HWND) (navigationSnapshot, error)
	interval     time.Duration
	onCollapsed  func()

	mu             sync.Mutex
	displayMu      sync.Mutex
	scrollRevision uint64
	aliases        map[string]win.HWND
	selected       string
	compact        bool
	revision       uint64
	cancel         chan struct{}
	wake           chan struct{}
	done           chan struct{}
	cache          navigation.FileCache
	resolver       navigation.PathResolver
}

func newBattleNavigationView(games, allGames game.Games, gameDir func() string, onCollapsed func()) *battleNavigationView {
	view := &battleNavigationView{
		games:       games,
		allGames:    allGames,
		gameDir:     gameDir,
		interval:    navigationUpdateInterval,
		selected:    navigationOffOption,
		aliases:     make(map[string]win.HWND),
		position:    binding.NewString(),
		status:      binding.NewString(),
		onCollapsed: onCollapsed,
	}
	view.readSnapshot = view.loadSnapshot
	view.refreshAliases()

	view.selector = widget.NewSelect(view.aliasOptions(), view.selectAlias)
	view.selector.PlaceHolder = navigationOffOption
	view.selector.Selected = navigationOffOption

	positionLabel := widget.NewLabelWithData(view.position)
	view.statusLabel = widget.NewLabelWithData(view.status)
	view.routeList = widget.NewList(
		func() int {
			view.routesMu.RLock()
			defer view.routesMu.RUnlock()
			return len(view.routes)
		},
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, object fyne.CanvasObject) {
			view.routesMu.RLock()
			defer view.routesMu.RUnlock()
			if id < len(view.routes) {
				object.(*widget.Label).SetText(view.routes[id])
			}
		},
	)
	view.scrollToTop = view.routeList.ScrollToTop
	listSize := canvas.NewRectangle(color.Transparent)
	listSize.SetMinSize(fyne.NewSize(280, 132))
	view.listFrame = container.NewMax(listSize, view.routeList)

	view.collapsedControls = []fyne.CanvasObject{view.selector}
	view.expandedControls = append(append([]fyne.CanvasObject(nil), view.collapsedControls...),
		positionLabel,
		view.statusLabel,
	)
	view.controls = container.NewVBox(view.collapsedControls...)
	view.collapsedObjects = []fyne.CanvasObject{view.controls}
	view.expandedObjects = []fyne.CanvasObject{view.controls, view.listFrame}
	view.container = container.NewVBox(view.collapsedObjects...)
	view.clearDisplay("Navigation is off.")
	return view
}

func (view *battleNavigationView) aliasOptions() []string {
	view.mu.Lock()
	defer view.mu.Unlock()
	return view.aliasOptionsLocked()
}

func (view *battleNavigationView) aliasOptionsLocked() []string {
	aliases := make([]string, 0, len(view.aliases))
	for alias := range view.aliases {
		aliases = append(aliases, alias)
	}
	sort.Strings(aliases)
	return append([]string{navigationOffOption}, aliases...)
}

func (view *battleNavigationView) refreshAliases() {
	aliases := make(map[string]win.HWND)
	for _, hWnd := range view.games.GetHWNDs() {
		if alias := view.allGames.FindKey(hWnd); alias != "" {
			aliases[alias] = hWnd
		}
	}

	view.mu.Lock()
	view.aliases = aliases
	view.mu.Unlock()
}

func (view *battleNavigationView) selectAlias(alias string) {
	view.refreshAliases()

	view.mu.Lock()
	if _, ok := view.aliases[alias]; !ok {
		alias = navigationOffOption
	}
	view.selected = alias
	view.revision++
	compact := view.compact
	running := view.cancel != nil
	wake := view.wake
	view.mu.Unlock()

	if alias == navigationOffOption {
		view.stopUpdater()
		view.selector.Selected = navigationOffOption
		view.selector.Refresh()
		view.setExpanded(false)
		view.clearDisplay("Navigation is off.")
		if compact && view.onCollapsed != nil {
			view.onCollapsed()
		}
		return
	}

	view.setExpanded(true)
	view.clearDisplay("Loading route data...")
	if compact {
		if running {
			select {
			case wake <- struct{}{}:
			default:
			}
		} else {
			view.startUpdater()
		}
	}
}

func (view *battleNavigationView) setCompact(compact bool) {
	view.stopUpdater()
	view.refreshAliases()

	view.mu.Lock()
	view.compact = compact
	selected := view.selected
	_, valid := view.aliases[selected]
	options := view.aliasOptionsLocked()
	if selected != navigationOffOption && !valid {
		view.selected = navigationOffOption
		selected = navigationOffOption
	}
	view.mu.Unlock()

	view.selector.Options = options
	view.selector.Selected = selected
	view.selector.Refresh()
	view.setExpanded(selected != navigationOffOption)
	if compact && selected != navigationOffOption {
		view.startUpdater()
	}
}

func (view *battleNavigationView) startUpdater() {
	view.mu.Lock()
	if !view.compact || view.selected == navigationOffOption || view.cancel != nil {
		view.mu.Unlock()
		return
	}
	cancel := make(chan struct{})
	wake := make(chan struct{}, 1)
	done := make(chan struct{})
	view.cancel = cancel
	view.wake = wake
	view.done = done
	view.mu.Unlock()

	go view.runUpdater(cancel, wake, done)
}

func (view *battleNavigationView) runUpdater(cancel, wake <-chan struct{}, done chan<- struct{}) {
	defer close(done)
	if !view.update(cancel) {
		return
	}

	ticker := time.NewTicker(view.interval)
	defer ticker.Stop()
	for {
		select {
		case <-cancel:
			return
		case <-wake:
			if !view.update(cancel) {
				return
			}
		case <-ticker.C:
			if !view.update(cancel) {
				return
			}
		}
	}
}

func (view *battleNavigationView) update(cancel <-chan struct{}) bool {
	view.mu.Lock()
	alias := view.selected
	hWnd, ok := view.aliases[alias]
	reader := view.readSnapshot
	revision := view.revision
	view.mu.Unlock()
	if !ok {
		view.resetMissingAlias(cancel)
		return false
	}

	snapshot, err := reader(hWnd)
	select {
	case <-cancel:
		return false
	default:
	}
	view.mu.Lock()
	current := view.revision == revision && view.selected == alias && view.aliases[alias] == hWnd
	view.mu.Unlock()
	if !current {
		return true
	}
	view.displayMu.Lock()
	defer view.displayMu.Unlock()
	select {
	case <-cancel:
		return false
	default:
	}
	view.mu.Lock()
	current = view.revision == revision && view.selected == alias && view.aliases[alias] == hWnd
	view.mu.Unlock()
	if !current {
		return true
	}
	if err != nil {
		view.setDisplay("", "Map data unavailable.", nil)
		return true
	}

	position := fmt.Sprintf("Position: (%d, %d)", snapshot.east, snapshot.south)
	if len(snapshot.routes) == 0 {
		view.setDisplay(position, "No routes found.", nil)
		return true
	}

	items := make([]string, len(snapshot.routes))
	for i, route := range snapshot.routes {
		items[i] = fmt.Sprintf("%s  (%d, %d) - %s", route.Direction, route.East, route.South, route.Type)
	}
	view.setDisplay(position, "", items)
	if view.scrollRevision != revision {
		view.scrollRevision = revision
		view.scrollToTop()
	}
	return true
}

func (view *battleNavigationView) resetMissingAlias(cancel <-chan struct{}) {
	select {
	case <-cancel:
		return
	default:
	}
	view.mu.Lock()
	view.selected = navigationOffOption
	compact := view.compact
	view.mu.Unlock()
	view.selector.Selected = navigationOffOption
	view.selector.Refresh()
	view.setExpanded(false)
	view.clearDisplay("Navigation is off.")
	if compact && view.onCollapsed != nil {
		view.onCollapsed()
	}
}

func (view *battleNavigationView) stopUpdater() {
	view.mu.Lock()
	cancel := view.cancel
	done := view.done
	view.cancel = nil
	view.wake = nil
	view.done = nil
	view.mu.Unlock()
	if cancel == nil {
		return
	}
	close(cancel)
	<-done
}

func (view *battleNavigationView) close() {
	view.stopUpdater()
}

func (view *battleNavigationView) setExpanded(expanded bool) {
	if expanded {
		view.controls.Objects = view.expandedControls
		view.controls.Refresh()
		view.container.Layout = layout.NewBorderLayout(view.controls, nil, nil, nil)
		view.container.Objects = view.expandedObjects
	} else {
		view.controls.Objects = view.collapsedControls
		view.controls.Refresh()
		view.container.Layout = layout.NewVBoxLayout()
		view.container.Objects = view.collapsedObjects
	}
	view.container.Refresh()
}

func (view *battleNavigationView) clearDisplay(status string) {
	view.displayMu.Lock()
	defer view.displayMu.Unlock()
	view.setDisplay("", status, nil)
}

func (view *battleNavigationView) setDisplay(position, status string, routes []string) {
	view.position.Set(position)
	view.status.Set(status)
	statusVisible := status != ""
	visibilityChanged := view.statusLabel.Visible() != statusVisible
	if !statusVisible {
		view.statusLabel.Hide()
	} else {
		view.statusLabel.Show()
	}
	if visibilityChanged {
		view.controls.Refresh()
		view.container.Refresh()
	}
	view.routesMu.Lock()
	view.routes = append(view.routes[:0], routes...)
	view.routesMu.Unlock()
	view.routeList.Refresh()
}

func (view *battleNavigationView) routeSnapshot() []string {
	view.routesMu.RLock()
	defer view.routesMu.RUnlock()
	return append([]string(nil), view.routes...)
}

func (view *battleNavigationView) loadSnapshot(hWnd win.HWND) (navigationSnapshot, error) {
	path, err := view.resolver.Resolve(view.gameDir(), game.GetMapCode(hWnd))
	if err != nil {
		return navigationSnapshot{}, err
	}
	stairs, err := view.cache.Load(path)
	if err != nil {
		return navigationSnapshot{}, err
	}
	position := game.GetCurrentGamePos(hWnd)
	east, south := int(position.X), int(position.Y)
	return navigationSnapshot{
		east:   east,
		south:  south,
		routes: navigation.BuildRoutes(stairs, east, south),
	}, nil
}

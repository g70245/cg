package container

import (
	"cg/game"
	"cg/game/battle"
	"cg/game/enum/movement"
	"cg/game/navigation"
	"cg/utils"
	"fmt"
	"image/color"
	"math"
	"sort"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
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

type navigationFloorIdentity struct {
	code uint32
	name string
}

type battleNavigationView struct {
	container         *fyne.Container
	controls          *fyne.Container
	selector          *widget.Select
	destination       *widget.Select
	navigationButton  *widget.Button
	listFrame         *fyne.Container
	collapsedControls []fyne.CanvasObject
	expandedControls  []fyne.CanvasObject
	collapsedObjects  []fyne.CanvasObject
	expandedObjects   []fyne.CanvasObject

	position              binding.String
	status                binding.String
	statusLabel           *widget.Label
	navigationStatus      binding.String
	navigationStatusLabel *widget.Label
	routeList             *widget.List
	scrollToTop           func()
	routesMu              sync.RWMutex
	routes                []string

	games               game.Games
	allGames            game.Games
	gameDir             func() string
	readSnapshot        func(win.HWND) (navigationSnapshot, error)
	newRunner           func(win.HWND) navigation.Runner
	closeWindows        func(win.HWND)
	beeperReady         func() bool
	playBeeper          func()
	stopBeeper          func()
	notifyBeeperMissing func()
	interval            time.Duration
	onCollapsed         func()

	mu               sync.Mutex
	mapLoadMu        sync.Mutex
	displayMu        sync.Mutex
	scrollRevision   uint64
	aliases          map[string]win.HWND
	selected         string
	compact          bool
	revision         uint64
	cancel           chan struct{}
	wake             chan struct{}
	done             chan struct{}
	cache            navigation.FileCache
	resolver         navigation.PathResolver
	mapIdentities    map[win.HWND]navigationFloorIdentity
	workers          map[win.HWND]*battle.Worker
	navigationStates map[win.HWND]*navigation.TraversalState
	navigationCancel chan struct{}
	navigationDone   chan struct{}
}

func newBattleNavigationView(games, allGames game.Games, gameDir func() string, onCollapsed func()) *battleNavigationView {
	view := &battleNavigationView{
		games:            games,
		allGames:         allGames,
		gameDir:          gameDir,
		interval:         navigationUpdateInterval,
		selected:         navigationOffOption,
		aliases:          make(map[string]win.HWND),
		position:         binding.NewString(),
		status:           binding.NewString(),
		navigationStatus: binding.NewString(),
		navigationStates: make(map[win.HWND]*navigation.TraversalState),
		mapIdentities:    make(map[win.HWND]navigationFloorIdentity),
		onCollapsed:      onCollapsed,
	}
	view.readSnapshot = view.loadSnapshot
	view.newRunner = view.navigationRunner
	view.closeWindows = game.CloseAllWindows
	view.beeperReady = utils.Beeper.IsReady
	view.playBeeper = utils.Beeper.Play
	view.stopBeeper = utils.Beeper.Stop
	view.notifyBeeperMissing = func() { notifyBeeperConfig("Navigation Setup") }
	view.refreshAliases()

	view.selector = widget.NewSelect(view.aliasOptions(), view.selectAlias)
	view.selector.PlaceHolder = navigationOffOption
	view.selector.Selected = navigationOffOption

	positionLabel := widget.NewLabelWithData(view.position)
	view.statusLabel = widget.NewLabelWithData(view.status)
	view.navigationStatusLabel = widget.NewLabelWithData(view.navigationStatus)
	view.navigationStatusLabel.Hide()
	positionRow := container.NewHBox(positionLabel, view.navigationStatusLabel)
	view.destination = widget.NewSelect([]string{string(navigation.StairUp), string(navigation.StairDown)}, func(selected string) {
		view.stopNavigation(true)
		view.clearNavigationExitAttempts()
	})
	view.destination.Selected = string(navigation.StairUp)
	view.navigationButton = widget.NewButtonWithIcon("", theme.MediaPlayIcon(), view.toggleNavigation)
	view.navigationButton.Importance = widget.WarningImportance
	navigationControls := container.NewBorder(nil, nil, nil, view.navigationButton, view.destination)
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
		navigationControls,
		positionRow,
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
	view.stopNavigation(true)
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
	if !compact {
		view.stopNavigation(true)
	}
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
	view.stopNavigation(true)
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
	done := view.stopNavigation(true)
	if done != nil {
		<-done
	}
	view.stopUpdater()
}

func (view *battleNavigationView) setWorkers(workers battle.Workers) {
	view.mu.Lock()
	view.workers = make(map[win.HWND]*battle.Worker, len(workers))
	for _, worker := range workers {
		view.workers[worker.GetHandle()] = worker
	}
	view.mu.Unlock()
}

func (view *battleNavigationView) toggleNavigation() {
	view.mu.Lock()
	running := view.navigationCancel != nil
	view.mu.Unlock()
	if running {
		view.stopNavigation(false)
		view.setNavigationStatus("Navigation stopped")
		return
	}
	view.startNavigation()
}

func (view *battleNavigationView) startNavigation() {
	status, _ := view.navigationStatus.Get()
	if status == navigation.StatusVerification && view.stopBeeper != nil {
		view.stopBeeper()
	}
	view.stopNavigation(true)
	view.mu.Lock()
	alias := view.selected
	hWnd, ok := view.aliases[alias]
	compact := view.compact
	worker := view.workers[hWnd]
	view.mu.Unlock()
	if !compact || !ok || alias == navigationOffOption {
		return
	}
	if worker != nil && worker.MovementMode() != movement.None {
		view.setNavigationStatus("Disable battle movement first.")
		return
	}
	if view.beeperReady == nil || !view.beeperReady() {
		view.setNavigationStatus("Set alert music.")
		if view.notifyBeeperMissing != nil {
			view.notifyBeeperMissing()
		}
		return
	}
	view.closeWindows(hWnd)
	targetType := navigation.StairType(view.destination.Selected)
	cancel := make(chan struct{})
	done := make(chan struct{})
	view.mu.Lock()
	view.navigationCancel = cancel
	view.navigationDone = done
	view.mu.Unlock()
	turn(theme.MediaStopIcon(), view.navigationButton)

	runner := view.newRunner(hWnd)
	go func() {
		defer close(done)
		err := runner.Run(cancel, targetType, func(status string) {
			view.mu.Lock()
			current := view.navigationDone == done
			view.mu.Unlock()
			if current {
				view.setNavigationStatus(status)
			}
		})
		view.mu.Lock()
		current := view.navigationDone == done
		if current {
			view.navigationCancel = nil
			view.navigationDone = nil
		}
		view.mu.Unlock()
		if !current {
			return
		}
		if err != nil {
			view.setNavigationStatus("Navigation unavailable.")
		}
		turn(theme.MediaPlayIcon(), view.navigationButton)
	}()
}

func (view *battleNavigationView) navigationRunner(hWnd win.HWND) navigation.Runner {
	state := view.navigationState(hWnd)
	return navigation.Runner{
		ReadSnapshot: func() (navigation.MovementSnapshot, error) {
			return view.loadMovementSnapshot(hWnd)
		},
		CanMove: view.allWindowsCanMove,
		Move: func(delta navigation.Point) {
			game.MoveMapOffset(hWnd, delta.East, delta.South)
		},
		State: state,
		VerificationTriggered: func() bool {
			return game.IsVerificationTriggered(view.gameDir())
		},
		OnVerification: view.playBeeper,
	}
}

func (view *battleNavigationView) navigationState(hWnd win.HWND) *navigation.TraversalState {
	view.mu.Lock()
	state := view.navigationStates[hWnd]
	if state == nil {
		state = &navigation.TraversalState{}
		view.navigationStates[hWnd] = state
	}
	view.mu.Unlock()
	return state
}

func (view *battleNavigationView) clearNavigationExitAttempts() {
	view.mu.Lock()
	states := make([]*navigation.TraversalState, 0, len(view.navigationStates))
	for _, state := range view.navigationStates {
		states = append(states, state)
	}
	view.mu.Unlock()
	for _, state := range states {
		state.ClearExitAttempts()
	}
}

func (view *battleNavigationView) stopNavigation(clearStatus bool) chan struct{} {
	view.mu.Lock()
	cancel := view.navigationCancel
	done := view.navigationDone
	view.navigationCancel = nil
	view.navigationDone = nil
	view.mu.Unlock()
	if cancel != nil {
		close(cancel)
	}
	if view.navigationButton != nil {
		turn(theme.MediaPlayIcon(), view.navigationButton)
	}
	if clearStatus {
		view.setNavigationStatus("")
	}
	return done
}

func (view *battleNavigationView) allWindowsCanMove() bool {
	for _, hWnd := range view.games.GetHWNDs() {
		if game.GetScene(hWnd) != game.NORMAL_SCENE {
			return false
		}
	}
	return true
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

func (view *battleNavigationView) setNavigationStatus(status string) {
	view.navigationStatus.Set(status)
	visible := status != ""
	visibilityChanged := view.navigationStatusLabel.Visible() != visible
	if visible {
		view.navigationStatusLabel.Show()
	} else {
		view.navigationStatusLabel.Hide()
	}
	if visibilityChanged {
		view.controls.Refresh()
		view.container.Refresh()
	}
}

func (view *battleNavigationView) routeSnapshot() []string {
	view.routesMu.RLock()
	defer view.routesMu.RUnlock()
	return append([]string(nil), view.routes...)
}

func (view *battleNavigationView) loadSnapshot(hWnd win.HWND) (navigationSnapshot, error) {
	_, _, path, data, position, err := view.loadMap(hWnd)
	if err != nil {
		return navigationSnapshot{}, err
	}
	if !navigation.IsMazePath(view.gameDir(), path) {
		view.navigationState(hWnd).Reset()
	}
	east, south := int(position.X), int(position.Y)
	return navigationSnapshot{
		east:   east,
		south:  south,
		routes: navigation.BuildRoutes(data.Stairs, east, south),
	}, nil
}

func (view *battleNavigationView) loadMovementSnapshot(hWnd win.HWND) (navigation.MovementSnapshot, error) {
	code, name, path, data, position, err := view.loadMap(hWnd)
	if err != nil {
		return navigation.MovementSnapshot{}, err
	}
	roundedEast := math.Round(position.X)
	roundedSouth := math.Round(position.Y)
	return navigation.MovementSnapshot{
		MapCode: code,
		MapName: name,
		Position: navigation.Point{
			East:  int(roundedEast),
			South: int(roundedSouth),
		},
		PositionSettled: math.Abs(position.X-roundedEast) < 0.1 && math.Abs(position.Y-roundedSouth) < 0.1,
		ExactEast:       position.X,
		ExactSouth:      position.Y,
		Map:             data,
		InMaze:          navigation.IsMazePath(view.gameDir(), path),
	}, nil
}

func (view *battleNavigationView) loadMap(hWnd win.HWND) (uint32, string, string, navigation.MapData, game.GamePos, error) {
	view.mapLoadMu.Lock()
	defer view.mapLoadMu.Unlock()

	code := game.GetMapCode(hWnd)
	name := game.GetMapName(hWnd)
	identity := navigationFloorIdentity{code: code, name: name}
	if view.mapIdentities == nil {
		view.mapIdentities = make(map[win.HWND]navigationFloorIdentity)
	}
	previous, knownIdentity := view.mapIdentities[hWnd]
	floorChanged := knownIdentity && previous != identity
	if floorChanged {
		view.cache.Invalidate()
	}
	path, err := view.resolver.Resolve(view.gameDir(), code)
	if err != nil {
		return 0, "", "", navigation.MapData{}, game.GamePos{}, err
	}
	data, err := view.cache.Load(view.gameDir(), path)
	if err != nil {
		return 0, "", "", navigation.MapData{}, game.GamePos{}, err
	}
	if game.GetMapCode(hWnd) != code || game.GetMapName(hWnd) != name {
		view.cache.Invalidate()
		return 0, "", "", navigation.MapData{}, game.GamePos{}, navigation.ErrMapUpdating
	}
	position := game.GetCurrentGamePos(hWnd)
	if floorChanged && navigation.IsMazePath(view.gameDir(), path) && !hasNearbyTransition(data.Stairs, position) {
		view.cache.Invalidate()
		return 0, "", "", navigation.MapData{}, game.GamePos{}, navigation.ErrMapUpdating
	}
	view.mapIdentities[hWnd] = identity
	return code, name, path, data, position, nil
}

func hasNearbyTransition(stairs []navigation.Stair, position game.GamePos) bool {
	east := int(math.Round(position.X))
	south := int(math.Round(position.Y))
	for _, stair := range stairs {
		if absInt(stair.East-east) <= 1 && absInt(stair.South-south) <= 1 {
			return true
		}
	}
	return false
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

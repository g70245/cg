package container

import (
	"cg/game"
	"cg/game/navigation"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	fynetest "fyne.io/fyne/v2/test"
	"github.com/g70245/win"
)

func TestNavigationIsOptInAndFollowsCompactLifecycle(t *testing.T) {
	testApp := fynetest.NewApp()
	defer testApp.Quit()

	games := game.Games{"selected": win.HWND(123)}
	allGames := game.Games{"Alpha": win.HWND(123)}
	var collapses atomic.Int32
	view := newBattleNavigationView(games, allGames, func() string { return "" }, func() { collapses.Add(1) })
	defer view.close()
	view.interval = 10 * time.Millisecond

	var reads atomic.Int32
	var scrolls atomic.Int32
	view.scrollToTop = func() { scrolls.Add(1) }
	view.readSnapshot = func(hWnd win.HWND) (navigationSnapshot, error) {
		reads.Add(1)
		return navigationSnapshot{
			east:  10,
			south: 20,
			routes: []navigation.Route{{
				Stair:     navigation.Stair{East: 12, South: 18, Type: navigation.StairUp},
				Direction: "↗",
			}}}, nil
	}

	if got, want := view.selector.Options, []string{navigationOffOption, "Alpha"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("selector options = %v, want %v", got, want)
	}
	if got, want := len(view.container.Objects), len(view.collapsedObjects); got != want {
		t.Fatalf("Off object count = %d, want %d", got, want)
	}
	if got := collapses.Load(); got != 0 {
		t.Fatalf("initial collapse callbacks = %d, want 0", got)
	}
	view.setCompact(true)
	time.Sleep(25 * time.Millisecond)
	if got := reads.Load(); got != 0 {
		t.Fatalf("default Off performed %d reads", got)
	}

	view.selectAlias("Alpha")
	if got, want := len(view.container.Objects), len(view.expandedObjects); got != want {
		t.Fatalf("selected object count = %d, want %d", got, want)
	}
	view.container.Resize(fyne.NewSize(320, 400))
	if got := view.listFrame.Size().Height; got <= 132 {
		t.Fatalf("resized route list height = %.1f, want greater than 132", got)
	}
	waitForNavigationTest(t, func() bool {
		return len(view.routeSnapshot()) == 1 && scrolls.Load() == 1
	})
	if position, _ := view.position.Get(); position != "Position: (10, 20)" {
		t.Fatalf("position = %q", position)
	}
	if routes := view.routeSnapshot(); !reflect.DeepEqual(routes, []string{"↗  (12, 18) - Up"}) {
		t.Fatalf("routes = %v", routes)
	}
	if status, _ := view.status.Get(); status != "" {
		t.Fatalf("route status = %q, want empty", status)
	}
	if view.statusLabel.Visible() {
		t.Fatal("route statistics row is still visible")
	}

	view.mu.Lock()
	firstDone := view.done
	view.mu.Unlock()
	readsBeforeSwitch := reads.Load()
	view.selectAlias("Alpha")
	select {
	case <-firstDone:
		t.Fatal("switching aliases stopped the shared updater")
	default:
	}
	view.mu.Lock()
	secondDone := view.done
	view.mu.Unlock()
	if firstDone != secondDone {
		t.Fatal("switching aliases started a duplicate updater")
	}
	waitForNavigationTest(t, func() bool { return reads.Load() > readsBeforeSwitch })
	waitForNavigationTest(t, func() bool { return scrolls.Load() == 2 })

	view.setCompact(false)
	stoppedAt := reads.Load()
	time.Sleep(25 * time.Millisecond)
	if got := reads.Load(); got != stoppedAt {
		t.Fatalf("leaving compact view performed %d additional reads", got-stoppedAt)
	}

	view.setCompact(true)
	waitForNavigationTest(t, func() bool { return reads.Load() > stoppedAt })
	view.selectAlias(navigationOffOption)
	if got, want := len(view.container.Objects), len(view.collapsedObjects); got != want {
		t.Fatalf("Off object count = %d, want %d", got, want)
	}
	if got := collapses.Load(); got != 1 {
		t.Fatalf("collapse callbacks = %d, want 1", got)
	}
	stoppedAt = reads.Load()
	time.Sleep(25 * time.Millisecond)
	if got := reads.Load(); got != stoppedAt {
		t.Fatalf("Off performed %d additional reads", got-stoppedAt)
	}
}

func TestNavigationMissingAliasResetsToOff(t *testing.T) {
	testApp := fynetest.NewApp()
	defer testApp.Quit()

	view := newBattleNavigationView(
		game.Games{"selected": win.HWND(123)},
		game.Games{"Alpha": win.HWND(123)},
		func() string { return "" },
		nil,
	)
	defer view.close()
	view.interval = 10 * time.Millisecond
	view.readSnapshot = func(win.HWND) (navigationSnapshot, error) {
		return navigationSnapshot{}, nil
	}
	view.setCompact(true)
	view.selectAlias("Alpha")

	view.mu.Lock()
	delete(view.aliases, "Alpha")
	view.mu.Unlock()
	waitForNavigationTest(t, func() bool {
		view.mu.Lock()
		defer view.mu.Unlock()
		return view.selected == navigationOffOption
	})
	if status, _ := view.status.Get(); status != "Navigation is off." {
		t.Fatalf("status = %q, want navigation off", status)
	}
}

func TestNavigationSwitchDoesNotWaitForPreviousRead(t *testing.T) {
	testApp := fynetest.NewApp()
	defer testApp.Quit()

	view := newBattleNavigationView(
		game.Games{"first": win.HWND(123), "second": win.HWND(456)},
		game.Games{"Alpha": win.HWND(123), "Beta": win.HWND(456)},
		func() string { return "" },
		nil,
	)
	defer view.close()
	firstReadStarted := make(chan struct{})
	releaseFirstRead := make(chan struct{})
	view.readSnapshot = func(hWnd win.HWND) (navigationSnapshot, error) {
		if hWnd == win.HWND(123) {
			select {
			case <-firstReadStarted:
			default:
				close(firstReadStarted)
			}
			<-releaseFirstRead
			return navigationSnapshot{east: 1, south: 1}, nil
		}
		return navigationSnapshot{east: 2, south: 2}, nil
	}
	view.setCompact(true)
	view.selectAlias("Alpha")
	select {
	case <-firstReadStarted:
	case <-time.After(time.Second):
		t.Fatal("first read did not start")
	}

	switched := make(chan struct{})
	go func() {
		view.selectAlias("Beta")
		close(switched)
	}()
	select {
	case <-switched:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("alias switch waited for the previous read")
	}
	close(releaseFirstRead)
	waitForNavigationTest(t, func() bool {
		position, _ := view.position.Get()
		return position == "Position: (2, 2)"
	})
}

func TestNavigationHidingStatusRemovesItsLayoutSpace(t *testing.T) {
	testApp := fynetest.NewApp()
	defer testApp.Quit()

	view := newBattleNavigationView(game.Games{}, game.Games{}, func() string { return "" }, nil)
	view.setExpanded(true)
	view.setDisplay("", "Map data unavailable.", nil)
	view.container.Resize(fyne.NewSize(320, 400))
	withStatus := view.listFrame.Position().Y

	view.setDisplay("Position: (10, 20)", "", []string{"↗  (12, 18) - Up"})
	withoutStatus := view.listFrame.Position().Y
	if withoutStatus >= withStatus {
		t.Fatalf("list Y after hiding status = %.1f, want less than %.1f", withoutStatus, withStatus)
	}
}

func waitForNavigationTest(t *testing.T, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("timed out waiting for navigation state")
}

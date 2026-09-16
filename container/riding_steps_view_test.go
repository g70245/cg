package container

import (
	"cg/game"
	"errors"
	"testing"
	"time"

	"github.com/g70245/win"
)

func TestRidingStepsViewDisplaysSortedNonzeroAliasesAndReadErrors(t *testing.T) {
	games := game.Games{
		"old-zero":  win.HWND(3),
		"old-beta":  win.HWND(2),
		"old-alpha": win.HWND(1),
	}
	allGames := game.Games{
		"Zero":  win.HWND(3),
		"Beta":  win.HWND(2),
		"Alpha": win.HWND(1),
	}
	view := newRidingStepsViewWith(games, allGames, func(hWnd win.HWND) (uint32, error) {
		switch hWnd {
		case win.HWND(1):
			return 439, nil
		case win.HWND(2):
			return 0, errors.New("read failed")
		default:
			return 0, nil
		}
	}, time.Hour)
	defer view.close()

	waitForRidingStepsText(t, view, "R  Alpha: 439 | Beta: ERR")
	if !view.container.Visible() {
		t.Fatal("riding steps view is hidden with nonzero steps")
	}
}

func TestRidingStepsViewHidesWhenAllStepsAreZero(t *testing.T) {
	games := game.Games{"Alpha": win.HWND(1)}
	view := newRidingStepsViewWith(games, games, func(win.HWND) (uint32, error) {
		return 0, nil
	}, time.Hour)
	defer view.close()

	waitForRidingStepsText(t, view, "")
	if view.container.Visible() {
		t.Fatal("riding steps view is visible when all steps are zero")
	}
}

func TestRidingStepsViewRefreshesAliases(t *testing.T) {
	games := game.Games{"Old": win.HWND(1)}
	allGames := game.Games{"Old": win.HWND(1)}
	view := newRidingStepsViewWith(games, allGames, func(win.HWND) (uint32, error) {
		return 480, nil
	}, time.Hour)
	defer view.close()

	waitForRidingStepsText(t, view, "R  Old: 480")
	allGames.RemoveValue(win.HWND(1))
	allGames.Add("New", win.HWND(1))
	view.refreshAliases(games, allGames)
	waitForRidingStepsText(t, view, "R  New: 480")
}

func TestRidingStepsViewCloseStopsUpdater(t *testing.T) {
	view := newRidingStepsViewWith(game.Games{}, game.Games{}, func(win.HWND) (uint32, error) {
		return 0, nil
	}, time.Hour)
	view.close()

	select {
	case <-view.done:
	default:
		t.Fatal("riding steps updater is still running after close")
	}
}

func waitForRidingStepsText(t *testing.T, view *ridingStepsView, want string) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if got, _ := view.status.Get(); got == want {
			return
		}
		time.Sleep(time.Millisecond)
	}
	got, _ := view.status.Get()
	t.Fatalf("riding steps text = %q, want %q", got, want)
}

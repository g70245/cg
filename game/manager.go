package game

import (
	. "cg/internal"
	"sort"

	"github.com/g70245/win"
	"golang.org/x/exp/maps"
)

type Games map[string]win.HWND

func NewGames() Games {
	return FindWindows()
}

func (gs Games) New(selected []string) Games {
	newGames := make(map[string]win.HWND)
	for _, game := range selected {
		newGames[game] = gs.Peek(game)
	}
	return newGames
}

func (gs Games) Take(game string) win.HWND {
	target := gs[game]
	delete(gs, game)
	return target
}

func (gs Games) Peek(game string) win.HWND {
	target := gs[game]
	return target
}

func (gs Games) Remove(games []string) {
	for _, game := range games {
		delete(gs, game)
	}
}

func (gs Games) RemoveValue(hWnd win.HWND) {
	for k, v := range gs {
		if v == hWnd {
			delete(gs, k)
			break
		}
	}
}

func (gs Games) FindKey(hWnd win.HWND) (key string) {
	for k, v := range gs {
		if v == hWnd {
			key = k
			break
		}
	}
	return
}

func (gs Games) Add(k string, v win.HWND) {
	gs[k] = v
}

func (gs Games) AddGames(games map[string]win.HWND) {
	for k, v := range games {
		gs[k] = v
	}
}

func (gs Games) GetSortedKeys() []string {
	keys := maps.Keys(gs)
	sort.Sort(sort.StringSlice(keys))
	return keys
}

func (gs Games) GetHWNDs() []win.HWND {
	hWnds := make([]win.HWND, 0)

	keys := maps.Keys(gs)
	sort.Sort(sort.StringSlice(keys))
	for _, key := range keys {
		hWnds = append(hWnds, gs.Peek(key))
	}

	return hWnds
}

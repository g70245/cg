package game

import (
	. "cg/system"
	"sort"

	. "github.com/g70245/win"
	"golang.org/x/exp/maps"
)

type Games map[string]HWND

func NewGames() Games {
	return FindWindows()
}

func (gs Games) New(selected []string) Games {
	newGames := make(map[string]HWND)
	for _, game := range selected {
		newGames[game] = gs.Peek(game)
	}
	return newGames
}

func (gs Games) Take(game string) HWND {
	target := gs[game]
	delete(gs, game)
	return target
}

func (gs Games) Peek(game string) HWND {
	target := gs[game]
	return target
}

func (gs Games) Remove(games []string) {
	for _, game := range games {
		delete(gs, game)
	}
}

func (gs Games) Add(games map[string]HWND) {
	for k, v := range games {
		gs[k] = v
	}
}

func (gs Games) GetSortedKeys() []string {
	keys := maps.Keys(gs)
	sort.Sort(sort.StringSlice(keys))
	return keys
}

func (gs Games) GetHWNDs() []HWND {
	hWnds := make([]HWND, 0)

	keys := maps.Keys(gs)
	sort.Sort(sort.StringSlice(keys))
	for _, key := range keys {
		hWnds = append(hWnds, gs.Peek(key))
	}

	return hWnds
}

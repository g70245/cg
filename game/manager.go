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

func (gs Games) Take(key string) win.HWND {
	target := gs[key]
	delete(gs, key)
	return target
}

func (gs Games) Peek(key string) win.HWND {
	target := gs[key]
	return target
}

func (gs Games) RemoveKeys(keys []string) {
	for _, game := range keys {
		delete(gs, game)
	}
}

func (gs Games) RemoveKey(key string) {
	delete(gs, key)
}

func (gs Games) RemoveValue(value win.HWND) {
	for k, v := range gs {
		if v == value {
			delete(gs, k)
			break
		}
	}
}

func (gs Games) FindKey(value win.HWND) (key string) {
	for k, v := range gs {
		if v == value {
			key = k
			break
		}
	}
	return
}

func (gs Games) Add(k string, v win.HWND) {
	gs[k] = v
}

func (gs Games) AddGames(addGames Games) {
	if gs == nil {
		gs = make(Games)
	}

	for _, v := range gs {
		if key := addGames.FindKey(v); key != "" {
			addGames.RemoveKey(key)
		}
	}

	for k, v := range addGames {
		gs[k] = v
	}
}

func (gs Games) GetSortedKeys() []string {
	keys := maps.Keys(gs)
	sort.Strings(keys)
	return keys
}

func (gs Games) GetHWNDs() []win.HWND {
	hWnds := make([]win.HWND, 0)

	keys := maps.Keys(gs)
	sort.Strings(keys)
	for _, key := range keys {
		hWnds = append(hWnds, gs.Peek(key))
	}

	return hWnds
}

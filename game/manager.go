package game

import (
	. "github.com/lxn/win"
)

type Games map[string]HWND

func newGames() Games {
	return Games(make(map[string]HWND))
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

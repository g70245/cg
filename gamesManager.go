package main

import (
	"github.com/lxn/win"
)

type Games map[string]win.HWND

func newGames() Games {
	return Games(make(map[string]win.HWND))
}

func (gs Games) take(game string) win.HWND {
	target := gs[game]
	delete(gs, game)
	return target
}

func (gs Games) peek(game string) win.HWND {
	target := gs[game]
	return target
}

func (gs Games) remove(games []string) {
	for _, game := range games {
		delete(gs, game)
	}
}

func (gs Games) add(games map[string]win.HWND) {
	for k, v := range games {
		gs[k] = v
	}
}

package main

import "github.com/lxn/win"

type Scene struct {
	x     int32
	y     int32
	color win.COLORREF
}

var (
	NORMAL_SCENE = Scene{92, 26, COLOR_GENERAL_SCENE}
	BATTLE_SCENE = Scene{}
)

func GetScene(hWnd win.HWND) Scene {
	if GetColor(hWnd, NORMAL_SCENE.x, NORMAL_SCENE.y) == NORMAL_SCENE.color {
		return NORMAL_SCENE
	}
	return BATTLE_SCENE
}

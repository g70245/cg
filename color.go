package main

import (
	"github.com/lxn/win"
)

const (
	COLOR_ANY = 0

	COLOR_SKILL        = 65536
	COLOR_SCENE_NORMAL = 15595514

	COLOR_MENU_BUTTON_NORMAL     = 15135992
	COLOR_MENU_BUTTON_T          = 15201528
	COLOR_MENU_BUTTON_POPOUT     = 10331818
	COLOR_MENU_BUTTON_R_POPOUT   = 10331817
	COLOR_BATTLE_COMMAND_ENABLE  = 7125907
	COLOR_BATTLE_COMMAND_DISABLE = 6991316

	COLOR_BATTLE_STAGE_PET = 15267320
)

func GetColor(hWnd win.HWND, x, y int32) (color win.COLORREF) {
	hdc := win.GetDC(hWnd)
	color = win.GetPixel(hdc, x, y)
	win.ReleaseDC(hWnd, hdc)
	return
}

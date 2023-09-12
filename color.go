package main

import (
	"github.com/lxn/win"
)

const (
	COLOR_SKILL         = 65536
	COLOR_GENERAL_SCENE = 15595514
)

func GetColor(hWnd win.HWND, x, y int32) (color win.COLORREF) {
	hdc := win.GetDC(hWnd)
	color = win.GetPixel(hdc, x, y)
	win.ReleaseDC(hWnd, hdc)
	return
}

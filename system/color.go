package system

import (
	"github.com/lxn/win"
)

func GetColor(hWnd win.HWND, x, y int32) (color win.COLORREF) {
	hdc := win.GetDC(hWnd)
	color = win.GetPixel(hdc, x, y)
	win.ReleaseDC(hWnd, hdc)
	return
}

package internal

import (
	"github.com/g70245/win"
)

func GetColor(hWnd win.HWND, x, y int32) (color win.COLORREF) {
	hdc := win.GetDC(hWnd)
	color = win.GetPixel(hdc, x, y)
	win.ReleaseDC(hWnd, hdc)
	return
}

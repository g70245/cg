package system

import (
	. "github.com/lxn/win"
)

func GetColor(hWnd HWND, x, y int32) (color COLORREF) {
	hdc := GetDC(hWnd)
	color = GetPixel(hdc, x, y)
	ReleaseDC(hWnd, hdc)
	return
}

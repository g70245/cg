package system

import (
	. "github.com/g70245/win"
)

func GetColor(hWnd HWND, x, y int32) (color COLORREF) {
	hdc := GetDC(hWnd)
	color = GetPixel(hdc, x, y)
	ReleaseDC(hWnd, hdc)
	return
}

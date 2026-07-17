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

func AreaContainsColor(hWnd win.HWND, originX, originY, destinationX, destinationY int32, expectedColor win.COLORREF) bool {
	hdc := win.GetDC(hWnd)
	defer win.ReleaseDC(hWnd, hdc)

	for x := originX; x <= destinationX; x++ {
		for y := originY; y <= destinationY; y++ {
			if win.GetPixel(hdc, x, y) == expectedColor {
				return true
			}
		}
	}

	return false
}

package system

import (
	"time"

	. "github.com/lxn/win"
)

const (
	GAME_WIDTH  = 640
	GAME_HEIGHT = 480
)

func MouseMsg(hWnd HWND, x, y int32, action uint32) {
	wparam := uintptr(0)
	lparam := uintptr(y<<16 | x)
	PostMessage(hWnd, action, wparam, lparam)
}

func LeftClick(hWnd HWND, x, y int32) {
	MouseMsg(hWnd, int32(x), int32(y), WM_MOUSEMOVE)
	MouseMsg(hWnd, int32(x), int32(y), WM_LBUTTONDOWN)
	MouseMsg(hWnd, int32(x), int32(y), WM_LBUTTONUP)
}

func MoveToMiddle(hWnd HWND) {
	MouseMsg(hWnd, int32(GAME_WIDTH/2), int32(GAME_HEIGHT/2), WM_MOUSEMOVE)
}

func KeyCombinationMsg(hWnd HWND, lkey, rkey uintptr) {
	lScanCode := MapVirtualKeyEx(uint32(lkey), 0)
	rScanCode := MapVirtualKeyEx(uint32(rkey), 0)
	llParam := (0x00000001 | (lScanCode << 16))
	rlParam := (0x00000001 | (rScanCode << 16))

	PostMessage(hWnd, WM_KEYDOWN, lkey, uintptr(llParam))
	PostMessage(hWnd, WM_KEYDOWN, rkey, uintptr(rlParam))

	time.Sleep(200 * time.Millisecond)

	llParam |= 0xC0000000
	rlParam |= 0xC0000000
	PostMessage(hWnd, WM_KEYUP, rkey, uintptr(rlParam))
	PostMessage(hWnd, WM_KEYUP, lkey, uintptr(llParam))
}

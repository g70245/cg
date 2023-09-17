package system

import (
	"time"

	. "github.com/lxn/win"
)

const (
	CLICK_DURATION = 60
)

func MouseMsg(hWnd HWND, x, y int32, action uint32) {
	wparam := uintptr(0)
	lparam := uintptr(y<<16 | x)
	PostMessage(hWnd, action, wparam, lparam)
}

func MoveToNowhere(hWnd HWND) {
	MouseMsg(hWnd, int32(-1), int32(-1), WM_MOUSEMOVE)
}

func LeftClick(hWnd HWND, x, y int32) {
	MouseMsg(hWnd, int32(x), int32(y), WM_MOUSEMOVE)
	time.Sleep(CLICK_DURATION * time.Millisecond)
	MouseMsg(hWnd, int32(x), int32(y), WM_LBUTTONDOWN)
	MouseMsg(hWnd, int32(x), int32(y), WM_LBUTTONUP)
}

func DoubleClick(hWnd HWND, x, y int32) {
	MouseMsg(hWnd, int32(x), int32(y), WM_MOUSEMOVE)
	time.Sleep(CLICK_DURATION * time.Millisecond)
	MouseMsg(hWnd, int32(x), int32(y), WM_LBUTTONDBLCLK)
}

func RightClick(hWnd HWND, x, y int32) {
	MouseMsg(hWnd, int32(x), int32(y), WM_MOUSEMOVE)
	time.Sleep(CLICK_DURATION * time.Millisecond)
	MouseMsg(hWnd, int32(x), int32(y), WM_RBUTTONDOWN)
	MouseMsg(hWnd, int32(x), int32(y), WM_RBUTTONUP)
}

func KeyCombinationMsg(hWnd HWND, lkey, rkey uintptr) {
	lScanCode := MapVirtualKeyEx(uint32(lkey), 0)
	rScanCode := MapVirtualKeyEx(uint32(rkey), 0)
	llParam := (0x00000001 | (lScanCode << 16))
	rlParam := (0x00000001 | (rScanCode << 16)) | (0 << 24) | (0 << 29) | (0 << 30) | (0 << 31)

	PostMessage(hWnd, WM_KEYDOWN, lkey, uintptr(llParam))
	PostMessage(hWnd, WM_KEYDOWN, rkey, uintptr(rlParam))

	time.Sleep(180 * time.Millisecond)

	llParam |= 0xC0000000
	rlParam |= 0xC0000000
	PostMessage(hWnd, WM_KEYUP, rkey, uintptr(rlParam))
	PostMessage(hWnd, WM_KEYUP, lkey, uintptr(llParam))
}

package system

import (
	"time"

	. "github.com/g70245/win"
)

const (
	CLICK_INTERVAL                   = 80
	DOUBLE_CLICK_INTERVAL            = 120
	KEY_INTERVAL                     = 80
	MOUSE_MOVE_NOWHERE_INTERVAL      = 40
	MOUSE_MOVE_INTERVAL              = 140
	BAD_COMPUTER_MOUSE_MOVE_INTERVAL = 180
)

func mouseMsg(hWnd HWND, x, y int32, action uint32) {
	wparam := uintptr(0)
	lparam := uintptr(y<<16 | x)
	PostMessage(hWnd, action, wparam, lparam)
}

func MoveCursorToNowhere(hWnd HWND) {
	MoveCursorWithDuration(hWnd, -1, -1, MOUSE_MOVE_NOWHERE_INTERVAL)
}

func MoveCursorWithDuration(hWnd HWND, x, y int32, d time.Duration) {
	mouseMsg(hWnd, x, y, WM_MOUSEMOVE)
	time.Sleep(d * time.Millisecond)
}

func MoveCursor(hWnd HWND, x, y int32) {
	MoveCursorWithDuration(hWnd, x, y, MOUSE_MOVE_INTERVAL)
}

func LeftClick(hWnd HWND, x, y int32) {
	MoveCursor(hWnd, x, y)
	mouseMsg(hWnd, x, y, WM_LBUTTONDOWN)
	mouseMsg(hWnd, x, y, WM_LBUTTONUP)
	time.Sleep(CLICK_INTERVAL * time.Millisecond)
}

func DoubleClickRepeatedly(hWnd HWND, x, y int32) {
	MoveCursor(hWnd, x, y)
	mouseMsg(hWnd, x, y, WM_LBUTTONDBLCLK)
	time.Sleep(DOUBLE_CLICK_INTERVAL * time.Millisecond)
	mouseMsg(hWnd, x, y, WM_LBUTTONDBLCLK)
	time.Sleep(DOUBLE_CLICK_INTERVAL * time.Millisecond)
}

func DoubleClick(hWnd HWND, x, y int32) {
	MoveCursor(hWnd, x, y)
	mouseMsg(hWnd, x, y, WM_LBUTTONDBLCLK)
	time.Sleep(DOUBLE_CLICK_INTERVAL * time.Millisecond)
}

func RightClick(hWnd HWND, x, y int32) {
	MoveCursor(hWnd, x, y)
	mouseMsg(hWnd, x, y, WM_RBUTTONDOWN)
	mouseMsg(hWnd, x, y, WM_RBUTTONUP)
	time.Sleep(CLICK_INTERVAL * time.Millisecond)
}

func KeyCombinationMsg(hWnd HWND, lkey, rkey uintptr) {
	lScanCode := MapVirtualKeyEx(uint32(lkey), 0)
	rScanCode := MapVirtualKeyEx(uint32(rkey), 0)
	llParam := (0x00000001 | (lScanCode << 16))
	rlParam := (0x00000001 | (rScanCode << 16)) | (0 << 24) | (0 << 29) | (0 << 30) | (0 << 31)

	PostMessage(hWnd, WM_KEYDOWN, lkey, uintptr(llParam))
	PostMessage(hWnd, WM_KEYDOWN, rkey, uintptr(rlParam))

	time.Sleep(KEY_INTERVAL * time.Millisecond)

	llParam |= 0xC0000000
	rlParam |= 0xC0000000
	PostMessage(hWnd, WM_KEYUP, rkey, uintptr(rlParam))
	PostMessage(hWnd, WM_KEYUP, lkey, uintptr(llParam))
	KeyMsg(hWnd, 0x08)
}

func KeyMsg(hWnd HWND, key uintptr) {
	PostMessage(hWnd, WM_KEYDOWN, key, 0)
	time.Sleep(KEY_INTERVAL * time.Millisecond)
	PostMessage(hWnd, WM_KEYUP, key, 0)
}

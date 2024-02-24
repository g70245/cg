package internal

import (
	"time"

	. "github.com/g70245/win"
)

const (
	DURATION_CLICK          = 80 * time.Millisecond
	DURATION_DOUBLE_CLICK   = 120 * time.Millisecond
	DURATION_CURSOR         = 140 * time.Millisecond
	DURATION_CURSOR_NOWHERE = 40 * time.Millisecond
	DURATION_KEY            = 80 * time.Millisecond
)

func MoveCursorToNowhere(hWnd HWND) {
	MoveCursorWithDuration(hWnd, -1, -1, DURATION_CURSOR_NOWHERE)
}

func MoveCursorWithDuration(hWnd HWND, x, y int32, d time.Duration) {
	PostCursorMsg(hWnd, x, y, WM_MOUSEMOVE)
	time.Sleep(d)
}

func LeftClick(hWnd HWND, x, y int32) {
	MoveCursorWithDuration(hWnd, x, y, DURATION_CURSOR)
	PostCursorMsg(hWnd, x, y, WM_LBUTTONDOWN)
	PostCursorMsg(hWnd, x, y, WM_LBUTTONUP)
	time.Sleep(DURATION_CLICK)
}

func DoubleClickRepeatedly(hWnd HWND, x, y int32) {
	MoveCursorWithDuration(hWnd, x, y, DURATION_CURSOR)
	PostCursorMsg(hWnd, x, y, WM_LBUTTONDBLCLK)
	time.Sleep(DURATION_DOUBLE_CLICK)
	PostCursorMsg(hWnd, x, y, WM_LBUTTONDBLCLK)
	time.Sleep(DURATION_DOUBLE_CLICK)
}

func DoubleClick(hWnd HWND, x, y int32) {
	MoveCursorWithDuration(hWnd, x, y, DURATION_CURSOR)
	PostCursorMsg(hWnd, x, y, WM_LBUTTONDBLCLK)
	time.Sleep(DURATION_DOUBLE_CLICK)
}

func RightClick(hWnd HWND, x, y int32) {
	MoveCursorWithDuration(hWnd, x, y, DURATION_CURSOR)
	PostCursorMsg(hWnd, x, y, WM_RBUTTONDOWN)
	PostCursorMsg(hWnd, x, y, WM_RBUTTONUP)
	time.Sleep(DURATION_CLICK)
}

func PostHotkeyMsg(hWnd HWND, lkey, rkey uintptr) {
	lScanCode := MapVirtualKeyEx(uint32(lkey), 0)
	rScanCode := MapVirtualKeyEx(uint32(rkey), 0)
	llParam := (0x00000001 | (lScanCode << 16))
	rlParam := (0x00000001 | (rScanCode << 16)) | (0 << 24) | (0 << 29) | (0 << 30) | (0 << 31)

	PostMessage(hWnd, WM_KEYDOWN, lkey, uintptr(llParam))
	PostMessage(hWnd, WM_KEYDOWN, rkey, uintptr(rlParam))

	time.Sleep(DURATION_KEY)

	llParam |= 0xC0000000
	rlParam |= 0xC0000000
	PostMessage(hWnd, WM_KEYUP, rkey, uintptr(rlParam))
	PostMessage(hWnd, WM_KEYUP, lkey, uintptr(llParam))
	PostKeyMsg(hWnd, 0x08)
}

func PostKeyMsg(hWnd HWND, key uintptr) {
	PostMessage(hWnd, WM_KEYDOWN, key, 0)
	time.Sleep(DURATION_KEY)
	PostMessage(hWnd, WM_KEYUP, key, 0)
}

func PostCursorMsg(hWnd HWND, x, y int32, action uint32) {
	wparam := uintptr(0)
	lparam := uintptr(y<<16 | x)
	PostMessage(hWnd, action, wparam, lparam)
}

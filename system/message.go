package system

import (
	"time"

	"github.com/lxn/win"
	. "github.com/lxn/win"
)

const (
	CLICK_INTERVAL                   = 60
	DOUBLE_CLICK_INTERVAL            = 120
	KEY_INTERVAL                     = 140
	MOUSE_MOVE_INTERVAL              = 160
	BAD_COMPUTER_MOUSE_MOVE_INTERVAL = 180
)

func MouseMsg(hWnd HWND, x, y int32, action uint32) {
	wparam := uintptr(0)
	lparam := uintptr(y<<16 | x)
	PostMessage(hWnd, action, wparam, lparam)
}

func MouseMsgWithIndicator(hWnd HWND, x, y int32, action uint32, wparam uintptr) {
	lparam := uintptr(y<<16 | x)
	PostMessage(hWnd, action, wparam, lparam)
}

func MoveToNowhere(hWnd HWND) {
	MouseMsg(hWnd, -1, -1, WM_MOUSEMOVE)
}

func MoveMouseWithDuration(hWnd HWND, x, y int32, d time.Duration) {
	MouseMsg(hWnd, x, y, win.WM_MOUSEMOVE)
	time.Sleep(d * time.Millisecond)
}

func MoveMouse(hWnd HWND, x, y int32) {
	MouseMsg(hWnd, x, y, win.WM_MOUSEMOVE)
	time.Sleep(MOUSE_MOVE_INTERVAL * time.Millisecond)
}

func LeftClick(hWnd HWND, x, y int32) {
	MoveMouse(hWnd, x, y)
	MouseMsg(hWnd, x, y, WM_LBUTTONDOWN)
	MouseMsg(hWnd, x, y, WM_LBUTTONUP)
	time.Sleep(CLICK_INTERVAL * time.Millisecond)
}

func DoubleClick(hWnd HWND, x, y int32) {
	MoveMouseWithDuration(hWnd, x, y, BAD_COMPUTER_MOUSE_MOVE_INTERVAL)
	MouseMsg(hWnd, x, y, WM_LBUTTONDBLCLK)
	time.Sleep(DOUBLE_CLICK_INTERVAL * time.Millisecond)
	MouseMsg(hWnd, x, y, WM_LBUTTONDBLCLK)
	time.Sleep(DOUBLE_CLICK_INTERVAL * time.Millisecond)
}

func RightClick(hWnd HWND, x, y int32) {
	MoveMouse(hWnd, x, y)
	MouseMsg(hWnd, x, y, WM_RBUTTONDOWN)
	MouseMsg(hWnd, x, y, WM_RBUTTONUP)
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
}

func KeyMsg(hWnd HWND, key uintptr) {
	PostMessage(hWnd, WM_KEYDOWN, key, 0)
	time.Sleep(KEY_INTERVAL * time.Millisecond)
	PostMessage(hWnd, WM_KEYDOWN, key, 0)
}

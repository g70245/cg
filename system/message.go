package system

import (
	"time"

	. "github.com/lxn/win"
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

func KeyCombinationMsg(hWnd HWND, lkey, rkey uintptr) {

	// MapVirtualKey(0x45,MAPVK_VK_TO_VSC));
	PostMessage(hWnd, WM_SYSKEYDOWN, lkey, uintptr(0))
	PostMessage(hWnd, WM_SYSKEYDOWN, VK_LMENU, uintptr(0))
	PostMessage(hWnd, WM_KEYDOWN, rkey, uintptr(0))
	time.Sleep(300 * time.Millisecond)
	PostMessage(hWnd, WM_KEYUP, rkey, uintptr(0))
	PostMessage(hWnd, WM_SYSKEYUP, VK_LMENU, uintptr(0))
	PostMessage(hWnd, WM_SYSKEYUP, lkey, uintptr(0))

}

package main

import (
	. "github.com/lxn/win"
)

func Act(handle HWND, x, y int32, action uint32) {
	wparam := uintptr(0)
	lparam := uintptr(y<<16 | x)
	PostMessage(handle, action, wparam, lparam)
}

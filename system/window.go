package system

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/g70245/win"
)

const (
	TARGET_CLASS = "Blue"
)

func FindWindows() map[string]win.HWND {
	handles := make(map[string]win.HWND)

	lpEnumFunc := syscall.NewCallback(func(h syscall.Handle, p uintptr) uintptr {

		maxCount := 260
		result := make([]uint16, maxCount)

		win.GetClassName(win.HWND(h), &result[0], maxCount)

		if syscall.UTF16ToString(result) == TARGET_CLASS {
			handles[fmt.Sprint(h)] = win.HWND(h)
		}
		return 1 // continue
	})

	win.EnumChildWindows(0, lpEnumFunc, 0)

	return handles
}

func FindChildWindows(parentHWnd win.HWND) {

	lpEnumFunc := func(h win.HWND, p uintptr) uintptr {
		// write matching algorithn here
		return 1 // continue
	}

	win.EnumChildWindows(parentHWnd, uintptr(unsafe.Pointer(&lpEnumFunc)), 0)
}

package internal

import (
	"fmt"
	"slices"
	"syscall"

	"github.com/g70245/win"
)

var (
	TARGET_CLASSES = []string{"Blue", "Sandbox:CG1:Blue"}
)

func FindWindows() map[string]win.HWND {
	handles := make(map[string]win.HWND)

	lpEnumFunc := syscall.NewCallback(func(h syscall.Handle, p uintptr) uintptr {

		maxCount := 260
		result := make([]uint16, maxCount)

		win.GetClassName(win.HWND(h), &result[0], maxCount)

		if slices.Contains(TARGET_CLASSES, syscall.UTF16ToString(result)) {
			handles[fmt.Sprint(h)] = win.HWND(h)
		}
		return 1 // continue
	})

	win.EnumChildWindows(0, lpEnumFunc, 0)

	return handles
}

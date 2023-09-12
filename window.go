package main

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/lxn/win"
	. "github.com/lxn/win"
)

type Window struct {
	hWnd HWND
}

func (w Window) Get() string {
	return fmt.Sprint(w.hWnd)
}

type Windows []Window

func (ws Windows) Get() []string {
	hWnds := make([]string, 0)
	for _, w := range ws {
		hWnds = append(hWnds, w.Get())
	}
	return hWnds
}

func FindWindows(class string) (Windows, error) {
	hWnds := make([]Window, 0)

	lpEnumFunc := syscall.NewCallback(func(h syscall.Handle, p uintptr) uintptr {

		maxCount := 260
		result := make([]uint16, maxCount)

		GetClassName(HWND(h), &result[0], maxCount)

		if syscall.UTF16ToString(result) == class {
			hWnds = append(hWnds, Window{HWND(h)})
		}
		return 1 // continue
	})

	EnumChildWindows(0, lpEnumFunc, 0)

	if len(hWnds) == 0 {
		return nil, fmt.Errorf("No window with class '%s' found", class)
	}
	return hWnds, nil
}

func FindChildWindows(parentHWnd HWND) {

	lpEnumFunc := func(h HWND, p uintptr) uintptr {
		// write matching algorithn here
		return 1 // continue
	}

	win.EnumChildWindows(parentHWnd, uintptr(unsafe.Pointer(&lpEnumFunc)), 0)
}

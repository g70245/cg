package main

import (
	"fmt"
	"os"
	"time"

	"github.com/lxn/win"
)

func PrintCursorPos(hWnd win.HWND) {
	for {
		var lpPoint win.POINT
		win.GetCursorPos(&lpPoint)
		fmt.Println(lpPoint)
		time.Sleep(360 * time.Millisecond)
	}
}

func PrintColorFromData(hWnd win.HWND, checkTargets []CheckTarget) {
	for _, target := range checkTargets {
		fmt.Print(target, " ")
		fmt.Println(GetColor(hWnd, target.x, target.y))
		Act(hWnd, int32(target.x), int32(target.y), win.WM_MOUSEMOVE)
		time.Sleep(360 * time.Millisecond)
	}
	os.Exit(0)
}

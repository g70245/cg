package main

import (
	sys "cg/system"

	"fmt"
	"os"
	"time"

	"github.com/lxn/win"
	"golang.org/x/exp/maps"
)

func PrintCursorPos(hWnd win.HWND) {
	for {
		var lpPoint win.POINT
		win.GetCursorPos(&lpPoint)
		fmt.Println(lpPoint)
		time.Sleep(360 * time.Millisecond)
	}
}

func PrintColorFromData(checkTargets []CheckTarget) {
	games := sys.FindWindows(TARGET_CLASS)
	for _, target := range checkTargets {
		fmt.Print(target, " ")
		fmt.Println(sys.GetColor(maps.Values(games)[0], target.x, target.y))
		sys.MouseMsg(maps.Values(games)[0], int32(target.x), int32(target.y), win.WM_MOUSEMOVE)
		time.Sleep(360 * time.Millisecond)
	}
	os.Exit(0)
}

func Test() {
	games := sys.FindWindows(TARGET_CLASS)
	i := 0
	for i < 5 {
		for _, h := range games {
			sys.KeyCombinationMsg(h, win.VK_SHIFT, win.VK_F12)
		}
		time.Sleep(200 * time.Millisecond)
		i++
	}
	os.Exit(0)
}

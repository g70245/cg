package main

import (
	. "cg/game"
	. "cg/system"
	"log"

	"os"
	"time"

	"github.com/lxn/win"
	"golang.org/x/exp/maps"
)

func PrintCursorPos(hWnd win.HWND) {
	for {
		var lpPoint win.POINT
		win.GetCursorPos(&lpPoint)
		log.Println(lpPoint)
		log.Println(GetColor(hWnd, lpPoint.X-30, lpPoint.Y))
		time.Sleep(800 * time.Millisecond)
	}
}

func PrintColorFromData(checkTargets []CheckTarget) {
	games := FindWindows(TARGET_CLASS)
	for _, target := range checkTargets {
		log.Print(target, " ")
		log.Println(GetColor(maps.Values(games)[0], target.GetX(), target.GetY()-25))
		MouseMsg(maps.Values(games)[0], int32(target.GetX()), int32(target.GetY()), win.WM_MOUSEMOVE)
		time.Sleep(360 * time.Millisecond)
	}
	os.Exit(0)
}

func KeyCombination() {
	games := FindWindows(TARGET_CLASS)
	i := 0
	for i < 5 {
		for _, h := range games {
			KeyCombinationMsg(h, win.VK_SHIFT, win.VK_F12)
		}
		time.Sleep(200 * time.Millisecond)
		i++
	}
	os.Exit(0)
}

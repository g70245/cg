package main

import (
	"cg/game"
	. "cg/game"
	"cg/system"
	. "cg/system"
	"fmt"
	"log"

	"os"
	"time"

	"github.com/lxn/win"
	"golang.org/x/exp/maps"
)

const (
	Offset = 26
)

var someTestData = []game.CheckTarget{
	game.PLAYER_L_1_H,
	game.PLAYER_L_2_H,
	game.PLAYER_L_3_H,
	game.PLAYER_L_4_H,
	game.PLAYER_L_5_H,
	game.PLAYER_L_1_P,
	game.PLAYER_L_2_P,
	game.PLAYER_L_3_P,
	game.PLAYER_L_4_P,
	game.PLAYER_L_5_P,
}

func PrintCursorPosColor(hWnd win.HWND) {
	for {
		var lpPoint win.POINT
		win.GetCursorPos(&lpPoint)
		fmt.Printf("(%d,%d) %d\n", lpPoint.X, lpPoint.Y-Offset, GetColor(hWnd, lpPoint.X-40, lpPoint.Y-Offset))
		time.Sleep(800 * time.Millisecond)
	}
}

func PrintColorFromData(hWnd win.HWND, checkTargets []CheckTarget) {
	for _, target := range checkTargets {
		log.Print(target, " ")
		log.Println(GetColor(hWnd, target.GetX(), target.GetY()))
		MouseMsg(hWnd, target.GetX(), target.GetX(), win.WM_MOUSEMOVE)
		time.Sleep(300 * time.Millisecond)
	}
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
}

func CheckColor(hWnd win.HWND, oX, oY, dX, dY int32, color win.COLORREF) {
	defer os.Exit(0)

	x := oX
	for x <= dX {
		y := oY
		for y <= dY {
			if GetColor(hWnd, x, y) == color {
				MouseMsg(hWnd, x, y, win.WM_MOUSEMOVE)
				fmt.Printf("Found at (%d, %d)\n", x, y)
				return
			}
			y += 1
		}
		x += 1
	}
}

func PrintColor(hWnd win.HWND, oX, oY, dX, dY int32) {
	defer os.Exit(0)

	x := oX
	for x <= dX {
		y := oY
		MouseMsg(hWnd, x, y, win.WM_MOUSEMOVE)
		time.Sleep(200 * time.Millisecond)
		for y <= dY {
			fmt.Printf("(%d, %d) %d\n", x, y, GetColor(hWnd, x, y))
			y += 1
		}
		x += 1
		MouseMsg(hWnd, x, y, win.WM_MOUSEMOVE)
		time.Sleep(200 * time.Millisecond)
	}
}

func getHWND() win.HWND {
	return maps.Values(system.FindWindows(TARGET_CLASS))[0]
}

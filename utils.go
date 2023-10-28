package main

import (
	. "cg/game"
	. "cg/system"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"

	"time"

	"github.com/g70245/win"
	"golang.org/x/exp/maps"
)

const (
	Offset = 26
)

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
	games := FindWindows()
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
	x := oX
	for x <= dX {
		y := oY
		// MouseMsg(hWnd, x, y, win.WM_MOUSEMOVE)
		// time.Sleep(200 * time.Millisecond)
		for y <= dY {
			fmt.Printf("(%d, %d) %d\n", x, y, GetColor(hWnd, x, y))
			y += 1
		}
		x += 1
		// MouseMsg(hWnd, x, y, win.WM_MOUSEMOVE)
		time.Sleep(200 * time.Millisecond)
	}
}

func getHWND() win.HWND {
	for _, h := range maps.Values(FindWindows()) {
		if fmt.Sprint(h) == "7933658" {
			return h
		}
	}
	return maps.Values(FindWindows())[0]
}

func getGoId() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}

func Test() {
	// hWnd := getHWND()
	// PrintCursorPosColor(hWnd)
	// CheckColor(hWnd, 584, 126, 588, 130, 11113016)
	// PrintColor(hWnd, 584, 126, 588, 130)
	// MoveMouse(hWnd, x, y)
	os.Exit(0)
}

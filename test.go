package main

import (
	. "cg/game"
	. "cg/internal"
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
	Y_OFFSET_GAME_CONTENT = 26
)

func printCursorPosColor(hWnd win.HWND) {
	for {
		var lpPoint win.POINT
		win.GetCursorPos(&lpPoint)
		log.Printf("(%d,%d) %d\n", lpPoint.X, lpPoint.Y-Y_OFFSET_GAME_CONTENT, GetColor(hWnd, lpPoint.X-40, lpPoint.Y-Y_OFFSET_GAME_CONTENT))
		time.Sleep(800 * time.Millisecond)
	}
}

func printColorsOfCheckTargets(hWnd win.HWND, checkTargets []CheckTarget) {
	for _, target := range checkTargets {
		log.Print(target, " ")
		log.Println(GetColor(hWnd, target.X, target.Y))
		// MoveCursorWithDuration(hWnd, target.X(), target.GetY(), 100*time.Millisecond)
		time.Sleep(300 * time.Millisecond)
	}
}

func pressHotkey() {
	games := FindWindows()
	i := 0
	for i < 5 {
		for _, h := range games {
			PostHotkeyMsg(h, win.VK_SHIFT, win.VK_F12)
		}
		time.Sleep(200 * time.Millisecond)
		i++
	}
}

func checkAreaColors(hWnd win.HWND, oX, oY, dX, dY int32, color win.COLORREF) {
	x := oX
	for x <= dX {
		y := oY
		for y <= dY {
			if GetColor(hWnd, x, y) == color {
				MoveCursorWithDuration(hWnd, x, y, 100*time.Millisecond)
				log.Printf("Found at (%d, %d)\n", x, y)
				return
			}
			y += 1
		}
		x += 1
	}
}

func printAreaColors(hWnd win.HWND, oX, oY, dX, dY int32) {
	x := oX
	for x <= dX {
		y := oY
		for y <= dY {
			log.Printf("(%d, %d) %d\n", x, y, GetColor(hWnd, x, y))
			y += 1
		}
		x += 1
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
	os.Exit(0)
}

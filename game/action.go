package game

import (
	. "cg/system"
	"time"

	. "github.com/lxn/win"
)

const (
	ACTION_INTERVAL = 170
)

func closeAllWindow(hWnd HWND) {
	KeyCombinationMsg(hWnd, VK_SHIFT, VK_F12)
	time.Sleep(ACTION_INTERVAL * time.Millisecond)
}

func openWindowByShortcut(hWnd HWND, key uintptr) {
	KeyCombinationMsg(hWnd, VK_SHIFT, VK_F12)
	time.Sleep(ACTION_INTERVAL * time.Millisecond)
	KeyCombinationMsg(hWnd, VK_CONTROL, key)
	time.Sleep(ACTION_INTERVAL * time.Millisecond)
	resetAllWindowsPosition(hWnd)
}

func resetAllWindowsPosition(hWnd HWND) {
	KeyCombinationMsg(hWnd, VK_CONTROL, VK_F12)
	time.Sleep(ACTION_INTERVAL * time.Millisecond)
}

func useHumanSkill(hWnd HWND, x, y int32, id, level int) {
	LeftClick(hWnd, x, y+int32((id-1)*16))
	time.Sleep(ACTION_INTERVAL * time.Millisecond)
	LeftClick(hWnd, x, y+int32((level-1)*16))
	time.Sleep(ACTION_INTERVAL * time.Millisecond)
}

func usePetSkill(hWnd HWND, x, y int32, id int) {
	LeftClick(hWnd, x, y+int32((id-1)*16))
	time.Sleep(ACTION_INTERVAL * time.Millisecond)
}

func clearChat(hWnd HWND) {
	KeyMsg(hWnd, VK_HOME)
	time.Sleep(ACTION_INTERVAL * time.Millisecond)
}

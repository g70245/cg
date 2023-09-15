package game

import (
	. "cg/system"
	"time"

	. "github.com/lxn/win"
)

func closeAll(hWnd HWND) {
	KeyCombinationMsg(hWnd, VK_SHIFT, VK_F12)
	time.Sleep(260 * time.Millisecond)
}

func openHumanWindow(hWnd HWND, key uintptr) {
	KeyCombinationMsg(hWnd, VK_SHIFT, VK_F12)
	time.Sleep(160 * time.Millisecond)
	KeyCombinationMsg(hWnd, VK_CONTROL, key)
	time.Sleep(160 * time.Millisecond)
	resetAllWindowPos(hWnd)
}

func resetAllWindowPos(hWnd HWND) {
	KeyCombinationMsg(hWnd, VK_CONTROL, VK_F12)
	time.Sleep(160 * time.Millisecond)
}

func useHumanSkill(hWnd HWND, x, y int32, id, level int) {
	LeftClick(hWnd, x, y+int32(id*16))
	time.Sleep(160 * time.Millisecond)
	LeftClick(hWnd, WINDOW_SKILL.x, y+int32(level*16))
	time.Sleep(160 * time.Millisecond)
}

func usePetSkill(hWnd HWND, x, y int32, id int) {
	LeftClick(hWnd, x, y+int32(id*16))
	time.Sleep(160 * time.Millisecond)
}

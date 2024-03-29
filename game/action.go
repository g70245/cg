package game

import (
	. "cg/internal"

	"time"

	"github.com/g70245/win"
)

const (
	DURATION_ACTION                    = 200 * time.Millisecond
	DURATION_ACTION_SKILL              = 320 * time.Millisecond
	DURATION_INVENTORY_CHECKER_WAITING = 400 * time.Millisecond
)

func closeAllWindows(hWnd win.HWND) {
	PostHotkeyMsg(hWnd, win.VK_SHIFT, win.VK_F12)
	time.Sleep(DURATION_ACTION)
}

func resetAllWindows(hWnd win.HWND) {
	PostHotkeyMsg(hWnd, win.VK_CONTROL, win.VK_F12)
	time.Sleep(DURATION_ACTION)
}

func openWindow(hWnd win.HWND, key uintptr) {
	RightClick(hWnd, GAME_WIDTH/2, 28)
	closeAllWindows(hWnd)
	PostHotkeyMsg(hWnd, win.VK_CONTROL, key)
	time.Sleep(DURATION_ACTION)
	resetAllWindows(hWnd)
}

func switchWindow(hWnd win.HWND, key uintptr) {
	PostHotkeyMsg(hWnd, win.VK_CONTROL, key)
	time.Sleep(DURATION_ACTION)
	resetAllWindows(hWnd)
}

func useHumanSkill(hWnd win.HWND, x, y int32, id, level int) {
	LeftClick(hWnd, x, y+int32((id-1)*16))
	time.Sleep(DURATION_ACTION)
	LeftClick(hWnd, x, y+int32((level-1)*16))
	time.Sleep(DURATION_ACTION_SKILL)
}

func usePetSkill(hWnd win.HWND, x, y int32, id int) {
	LeftClick(hWnd, x, y+int32((id-1)*16))
	time.Sleep(DURATION_ACTION)
}

func clearChat(hWnd win.HWND) {
	PostKeyMsg(hWnd, win.VK_HOME)
	time.Sleep(DURATION_ACTION)
}

func openInventory(hWnd win.HWND) {
	openWindow(hWnd, KEY_INVENTORY)
}

func useItem(hWnd win.HWND, x, y int32) {
	DoubleClickRepeatedly(hWnd, x, y)
}

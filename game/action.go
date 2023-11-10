package game

import (
	sys "cg/system"

	"time"

	. "github.com/g70245/win"
)

const (
	ACTION_INTERVAL                         = 200
	BATTLE_RESULT_DISAPPEARING_INTERVAL_SEC = 2
	INVENTORY_CHECKER_WAITING_INTERVAL      = 400
)

func closeAllWindows(hWnd HWND) {
	sys.KeyCombinationMsg(hWnd, VK_SHIFT, VK_F12)
	time.Sleep(ACTION_INTERVAL * time.Millisecond)
}

func openWindowByShortcut(hWnd HWND, key uintptr) {
	sys.RightClick(hWnd, GAME_WIDTH/2, 28)
	closeAllWindows(hWnd)
	sys.KeyCombinationMsg(hWnd, VK_CONTROL, key)
	time.Sleep(ACTION_INTERVAL * time.Millisecond)
	resetAllWindowsPosition(hWnd)
}

func leverWindowByShortcutWithoutClosingOtherWindows(hWnd HWND, key uintptr) {
	sys.KeyCombinationMsg(hWnd, VK_CONTROL, key)
	time.Sleep(ACTION_INTERVAL * time.Millisecond)
	resetAllWindowsPosition(hWnd)
}

func resetAllWindowsPosition(hWnd HWND) {
	sys.KeyCombinationMsg(hWnd, VK_CONTROL, VK_F12)
	time.Sleep(ACTION_INTERVAL * time.Millisecond)
}

func useHumanSkill(hWnd HWND, x, y int32, id, level int) {
	sys.LeftClick(hWnd, x, y+int32((id-1)*16))
	time.Sleep(ACTION_INTERVAL * time.Millisecond)
	sys.LeftClick(hWnd, x, y+int32((level-1)*16))
	time.Sleep(ACTION_INTERVAL * time.Millisecond)
}

func usePetSkill(hWnd HWND, x, y int32, id int) {
	sys.LeftClick(hWnd, x, y+int32((id-1)*16))
	time.Sleep(ACTION_INTERVAL * time.Millisecond)
}

func clearChat(hWnd HWND) {
	sys.KeyMsg(hWnd, VK_HOME)
	time.Sleep(ACTION_INTERVAL * time.Millisecond)
}

func checkInventory(hWnd HWND) bool {
	defer closeAllWindows(hWnd)
	defer time.Sleep(INVENTORY_CHECKER_WAITING_INTERVAL * time.Millisecond)

	time.Sleep(BATTLE_RESULT_DISAPPEARING_INTERVAL_SEC * time.Second)
	closeAllWindows(hWnd)
	sys.LeftClick(hWnd, GAME_WIDTH/2, GAME_HEIGHT/2)

	openWindowByShortcut(hWnd, 0x45)

	if px, py, ok := getNSItemWindowPos(hWnd); ok {
		return !isAnyItemSlotFree(hWnd, px, py)
	}
	return false
}

func checkActivityInventory(hWnd HWND) bool {
	defer closeAllWindows(hWnd)
	defer time.Sleep(INVENTORY_CHECKER_WAITING_INTERVAL * time.Millisecond)

	time.Sleep(BATTLE_RESULT_DISAPPEARING_INTERVAL_SEC * time.Second)
	closeAllWindows(hWnd)
	sys.LeftClick(hWnd, GAME_WIDTH/2, GAME_HEIGHT/2)

	openWindowByShortcut(hWnd, 0x45)

	if px, py, ok := getNSItemWindowPos(hWnd); ok {
		return !isMoreThanTwoItemSlotsFree(hWnd, px, py)
	}
	return false
}

func checkInventoryWithoutClosingAllWindows(hWnd HWND) bool {
	defer leverWindowByShortcutWithoutClosingOtherWindows(hWnd, 0x45)
	leverWindowByShortcutWithoutClosingOtherWindows(hWnd, 0x45)

	if px, py, ok := getNSItemWindowPos(hWnd); ok {
		return !isAnyItemSlotFree(hWnd, px, py)
	}
	return false
}

func getMapName(hWnd HWND) string {
	return sys.ReadMemoryString(hWnd, 0x95C870, 32)
}

type GamePos struct {
	x, y float64
}

func getCurrentGamePos(hWnd HWND) GamePos {
	fx := sys.ReadMemoryFloat32(hWnd, 0x95C88C, 32)
	fy := sys.ReadMemoryFloat32(hWnd, 0x95C890, 32)
	return GamePos{float64(fx / 64), float64(fy / 64)}
}

package game

import (
	"cg/system"

	"time"

	. "github.com/g70245/win"
)

const (
	ACTION_INTERVAL                         = 200
	INVENTORY_CHECKER_WAITING_INTERVAL      = 400
	BATTLE_RESULT_DISAPPEARING_INTERVAL_SEC = 2
)

func closeAllWindows(hWnd HWND) {
	system.KeyCombinationMsg(hWnd, VK_SHIFT, VK_F12)
	time.Sleep(ACTION_INTERVAL * time.Millisecond)
}

func openWindowByShortcut(hWnd HWND, key uintptr) {
	system.RightClick(hWnd, GAME_WIDTH/2, 28)
	closeAllWindows(hWnd)
	system.KeyCombinationMsg(hWnd, VK_CONTROL, key)
	time.Sleep(ACTION_INTERVAL * time.Millisecond)
	resetPosition(hWnd)
}

func switchWindowWithShortcut(hWnd HWND, key uintptr) {
	system.KeyCombinationMsg(hWnd, VK_CONTROL, key)
	time.Sleep(ACTION_INTERVAL * time.Millisecond)
	resetPosition(hWnd)
}

func resetPosition(hWnd HWND) {
	system.KeyCombinationMsg(hWnd, VK_CONTROL, VK_F12)
	time.Sleep(ACTION_INTERVAL * time.Millisecond)
}

func useHumanSkill(hWnd HWND, x, y int32, id, level int) {
	system.LeftClick(hWnd, x, y+int32((id-1)*16))
	time.Sleep(ACTION_INTERVAL * time.Millisecond)
	system.LeftClick(hWnd, x, y+int32((level-1)*16))
	time.Sleep(ACTION_INTERVAL * time.Millisecond)
}

func usePetSkill(hWnd HWND, x, y int32, id int) {
	system.LeftClick(hWnd, x, y+int32((id-1)*16))
	time.Sleep(ACTION_INTERVAL * time.Millisecond)
}

func clearChat(hWnd HWND) {
	system.KeyMsg(hWnd, VK_HOME)
	time.Sleep(ACTION_INTERVAL * time.Millisecond)
}

func openInventory(hWnd HWND) {
	openWindowByShortcut(hWnd, 0x45)
}

func isInventoryFull(hWnd HWND) bool {
	defer closeAllWindows(hWnd)
	defer time.Sleep(INVENTORY_CHECKER_WAITING_INTERVAL * time.Millisecond)

	time.Sleep(BATTLE_RESULT_DISAPPEARING_INTERVAL_SEC * time.Second)
	closeAllWindows(hWnd)
	system.LeftClick(hWnd, GAME_WIDTH/2, GAME_HEIGHT/2)

	openInventory(hWnd)

	if px, py, ok := getNSItemWindowPos(hWnd); ok {
		return !isAnyItemSlotFree(hWnd, px, py)
	}
	return false
}

func isInventoryFullForActivity(hWnd HWND) bool {
	defer closeAllWindows(hWnd)
	defer time.Sleep(INVENTORY_CHECKER_WAITING_INTERVAL * time.Millisecond)

	time.Sleep(BATTLE_RESULT_DISAPPEARING_INTERVAL_SEC * time.Second)
	closeAllWindows(hWnd)
	system.LeftClick(hWnd, GAME_WIDTH/2, GAME_HEIGHT/2)

	openInventory(hWnd)

	if px, py, ok := getNSItemWindowPos(hWnd); ok {
		return !isMoreThanTwoItemSlotsFree(hWnd, px, py)
	}
	return false
}

func isInventoryFullWithoutClosingAllWindows(hWnd HWND) bool {
	defer switchWindowWithShortcut(hWnd, 0x45)
	switchWindowWithShortcut(hWnd, 0x45)

	if px, py, ok := getNSItemWindowPos(hWnd); ok {
		return !isAnyItemSlotFree(hWnd, px, py)
	}
	return false
}

func getMapName(hWnd HWND) string {
	return system.ReadMemoryString(hWnd, 0x95C870, 32)
}

type GamePos struct {
	x, y float64
}

func getCurrentGamePos(hWnd HWND) GamePos {
	fx := system.ReadMemoryFloat32(hWnd, 0x95C88C, 32)
	fy := system.ReadMemoryFloat32(hWnd, 0x95C890, 32)
	return GamePos{float64(fx / 64), float64(fy / 64)}
}

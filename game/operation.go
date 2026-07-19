package game

import (
	"cg/internal"

	"time"

	"github.com/g70245/win"
)

const (
	DURATION_ACTION                    = 200 * time.Millisecond
	DURATION_ACTION_SKILL              = 320 * time.Millisecond
	DURATION_INVENTORY_CHECKER_WAITING = 400 * time.Millisecond
	durationNavigationCursor           = 50 * time.Millisecond
	durationNavigationClick            = 30 * time.Millisecond
)

func CloseAllWindows(hWnd win.HWND) {
	internal.PostHotkeyMsg(hWnd, win.VK_SHIFT, win.VK_F12)
	time.Sleep(DURATION_ACTION)
}

func ResetAllWindows(hWnd win.HWND) {
	internal.PostHotkeyMsg(hWnd, win.VK_CONTROL, win.VK_F12)
	time.Sleep(DURATION_ACTION)
}

func OpenWindow(hWnd win.HWND, key uintptr) {
	internal.RightClick(hWnd, GAME_WIDTH/2, 28)
	CloseAllWindows(hWnd)
	internal.PostHotkeyMsg(hWnd, win.VK_CONTROL, key)
	time.Sleep(DURATION_ACTION)
	ResetAllWindows(hWnd)
}

func SwitchWindow(hWnd win.HWND, key uintptr) {
	internal.PostHotkeyMsg(hWnd, win.VK_CONTROL, key)
	time.Sleep(DURATION_ACTION)
	ResetAllWindows(hWnd)
}

func UseCharacterSkill(hWnd win.HWND, x, y int32, id, level int) {
	internal.LeftClick(hWnd, x, y+int32((id-1)*16))
	time.Sleep(DURATION_ACTION)
	internal.LeftClick(hWnd, x, y+int32((level-1)*16))
	time.Sleep(DURATION_ACTION_SKILL)
}

func UsePetSkill(hWnd win.HWND, x, y int32, id int) {
	internal.LeftClick(hWnd, x, y+int32((id-1)*16))
	time.Sleep(DURATION_ACTION)
}

func ClearChat(hWnd win.HWND) {
	internal.PostKeyMsg(hWnd, win.VK_HOME)
	time.Sleep(DURATION_ACTION)
}

func OpenInventory(hWnd win.HWND) {
	OpenWindow(hWnd, KEY_INVENTORY)
}

func UseItem(hWnd win.HWND, x, y int32) {
	internal.DoubleClickRepeatedly(hWnd, x, y)
}

func MoveMapOffset(hWnd win.HWND, deltaEast, deltaSouth int) {
	x, y := mapStepTarget(deltaEast, deltaSouth)
	internal.LeftClickWithDuration(hWnd, x, y, durationNavigationCursor, durationNavigationClick)
}

func mapStepTarget(deltaEast, deltaSouth int) (int32, int32) {
	const (
		tileHalfWidth  = 32
		tileHalfHeight = 24
	)
	x := int32(GAME_WIDTH/2 + tileHalfWidth*(deltaEast+deltaSouth))
	y := int32(GAME_HEIGHT/2 + tileHalfHeight*(deltaSouth-deltaEast))
	return x, y
}

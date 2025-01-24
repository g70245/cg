package game

import (
	"cg/game/items"
	"cg/internal"

	"github.com/g70245/win"
)

type CheckTarget struct {
	X     int32
	Y     int32
	Color win.COLORREF
}

var (
	NOWHERE_SCENE = CheckTarget{}
	NORMAL_SCENE  = CheckTarget{92, 26, COLOR_SCENE_NORMAL}
	BATTLE_SCENE  = CheckTarget{108, 10, COLOR_SCENE_BATTLE}

	MENU_STATUS           = CheckTarget{60, 468, COLOR_MENU_BUTTON_NORMAL}
	MENU_SKILL            = CheckTarget{136, 468, COLOR_MENU_BUTTON_NORMAL}
	MENU_INVENTORY        = CheckTarget{212, 468, COLOR_MENU_BUTTON_NORMAL}
	MENU_PET              = CheckTarget{288, 468, COLOR_MENU_BUTTON_NORMAL}
	MENU_CONTACT          = CheckTarget{366., 468, COLOR_MENU_BUTTON_CONTACT}
	MENU_ESC              = CheckTarget{514, 468, COLOR_MENU_BUTTON_NORMAL}
	MENU_STATUS_POPOUT    = CheckTarget{60, 468, COLOR_MENU_BUTTON_POPOUT}
	MENU_SKILL_POPOUT     = CheckTarget{136, 468, COLOR_MENU_BUTTON_POPOUT}
	MENU_INVENTORY_POPOUT = CheckTarget{212, 468, COLOR_MENU_BUTTON_POPOUT}
	MENU_PET_POPOUT       = CheckTarget{288, 468, COLOR_MENU_BUTTON_PET_POPOUT}
	MENU_CONTACT_POPOUT   = CheckTarget{366., 468, COLOR_MENU_BUTTON_POPOUT}
	MENU_ESC_POPOUT       = CheckTarget{514, 468, COLOR_MENU_BUTTON_POPOUT}

	NORMAL_WINDOW_ITEM_MONEY_PIVOT = CheckTarget{354, 134, COLOR_NS_INVENTORY_PIVOT}
)

func IsBattleScene(hWnd win.HWND) bool {
	return GetScene(hWnd) == BATTLE_SCENE
}

func GetScene(hWnd win.HWND) CheckTarget {
	if internal.GetColor(hWnd, NORMAL_SCENE.X, NORMAL_SCENE.Y) == NORMAL_SCENE.Color {
		return NORMAL_SCENE
	} else if internal.GetColor(hWnd, BATTLE_SCENE.X, BATTLE_SCENE.Y) == BATTLE_SCENE.Color {
		return BATTLE_SCENE
	}
	return NOWHERE_SCENE
}

func GetItemWindowPos(hWnd win.HWND) (int32, int32, bool) {
	internal.MoveCursorToNowhere(hWnd)
	x := NORMAL_WINDOW_ITEM_MONEY_PIVOT.X
	for x <= NORMAL_WINDOW_ITEM_MONEY_PIVOT.X+34 {
		y := NORMAL_WINDOW_ITEM_MONEY_PIVOT.Y
		for y <= NORMAL_WINDOW_ITEM_MONEY_PIVOT.Y+54 {
			if internal.GetColor(hWnd, x, y) == NORMAL_WINDOW_ITEM_MONEY_PIVOT.Color {
				return x, y + 20, true
			}
			y += 1
		}
		x += 1
	}
	return 0, 0, false
}

func IsAnyInventorySlotFree(hWnd win.HWND, px, py int32) bool {
	internal.MoveCursorToNowhere(hWnd)
	x := px
	y := py
	var i, j int32

	for i = 0; i < 5; i++ {
		for j = 0; j < 4; j++ {
			if IsInventorySlotFree(hWnd, x+i*ITEM_COL_LEN, y+j*ITEM_COL_LEN) {
				return true
			}
		}
	}

	return false
}

func AreMoreThanTwoInventorySlotsFree(hWnd win.HWND, px, py int32) bool {
	internal.MoveCursorToNowhere(hWnd)
	x := px
	y := py

	counter := 0
	var i, j int32

	for i = 0; i < 5; i++ {
		for j = 0; j < 4; j++ {
			if IsInventorySlotFree(hWnd, x+i*ITEM_COL_LEN, y+j*ITEM_COL_LEN) {
				counter++
			}
		}
	}

	return counter > 2
}

func IsInventorySlotFree(hWnd win.HWND, px, py int32) bool {
	x := px
	for x < px+30 {
		y := py
		for y < py+30 {
			if internal.GetColor(hWnd, x, y) != COLOR_NS_INVENTORY_SLOT_EMPTY {
				return false
			}
			y += 5
		}
		x += 5
	}
	return true
}

func GetItemPos(hWnd win.HWND, px, py int32, color win.COLORREF, granularity int32) (int32, int32, bool) {
	internal.MoveCursorToNowhere(hWnd)

	x := px
	y := py
	var i, j int32

	for j = 0; j < 4; j++ {
		for i = 0; i < 5; i++ {
			if tx, ty, found := searchSlotForColor(hWnd, x+i*ITEM_COL_LEN, y+j*ITEM_COL_LEN, color, granularity); found {
				return tx, ty, found
			}
		}
	}

	return 0, 0, false
}

func searchSlotForColor(hWnd win.HWND, px, py int32, color win.COLORREF, granularity int32) (int32, int32, bool) {
	x := px
	for x < px+30 {
		y := py
		for y < py+30 {
			currentColor := internal.GetColor(hWnd, x, y)
			if currentColor == color {
				return x, y, true
			} else if currentColor == items.COLOR_ITEM_CAN_NOT_BE_USED {
				return 0, 0, false
			}
			y += granularity
		}
		x += granularity
	}
	return 0, 0, false
}

func IsInventoryFullWithoutClosingAllWindows(hWnd win.HWND) bool {
	defer SwitchWindow(hWnd, 0x45)
	SwitchWindow(hWnd, 0x45)

	if px, py, ok := GetItemWindowPos(hWnd); ok {
		return !IsAnyInventorySlotFree(hWnd, px, py)
	}
	return false
}

func GetMapName(hWnd win.HWND) string {
	return internal.ReadMemoryString(hWnd, MEMORY_MAP_NAME, 32)
}

type GamePos struct {
	X, Y float64
}

func GetCurrentGamePos(hWnd win.HWND) GamePos {
	fx := internal.ReadMemoryFloat32(hWnd, MEMORY_MAP_POS_X, 32)
	fy := internal.ReadMemoryFloat32(hWnd, MEMORY_MAP_POS_Y, 32)
	return GamePos{float64(fx / 64), float64(fy / 64)}
}

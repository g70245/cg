package game

import (
	. "cg/internal"

	"time"

	"github.com/g70245/win"
)

type CheckTarget struct {
	x     int32
	y     int32
	color win.COLORREF
}

func (t *CheckTarget) GetX() int32 {
	return t.x
}

func (t *CheckTarget) GetY() int32 {
	return t.y
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

func getScene(hWnd win.HWND) CheckTarget {
	if GetColor(hWnd, NORMAL_SCENE.x, NORMAL_SCENE.y) == NORMAL_SCENE.color {
		return NORMAL_SCENE
	} else if GetColor(hWnd, BATTLE_SCENE.x, BATTLE_SCENE.y) == BATTLE_SCENE.color {
		return BATTLE_SCENE
	}
	return NOWHERE_SCENE
}

func getItemWindowPos(hWnd win.HWND) (int32, int32, bool) {
	MoveCursorToNowhere(hWnd)
	x := NORMAL_WINDOW_ITEM_MONEY_PIVOT.x
	for x <= NORMAL_WINDOW_ITEM_MONEY_PIVOT.x+34 {
		y := NORMAL_WINDOW_ITEM_MONEY_PIVOT.y
		for y <= NORMAL_WINDOW_ITEM_MONEY_PIVOT.y+54 {
			if GetColor(hWnd, x, y) == NORMAL_WINDOW_ITEM_MONEY_PIVOT.color {
				return x, y + 20, true
			}
			y += 1
		}
		x += 1
	}
	return 0, 0, false
}

func isAnyInventorySlotFree(hWnd win.HWND, px, py int32) bool {
	MoveCursorToNowhere(hWnd)
	x := px
	y := py
	var i, j int32

	for i = 0; i < 5; i++ {
		for j = 0; j < 4; j++ {
			if isInventorySlotFree(hWnd, x+i*ITEM_COL_LEN, y+j*ITEM_COL_LEN) {
				return true
			}
		}
	}

	return false
}

func areMoreThanTwoInventorySlotsFree(hWnd win.HWND, px, py int32) bool {
	MoveCursorToNowhere(hWnd)
	x := px
	y := py

	counter := 0
	var i, j int32

	for i = 0; i < 5; i++ {
		for j = 0; j < 4; j++ {
			if isInventorySlotFree(hWnd, x+i*ITEM_COL_LEN, y+j*ITEM_COL_LEN) {
				counter++
			}
		}
	}

	return counter > 2
}

func isInventorySlotFree(hWnd win.HWND, px, py int32) bool {
	x := px
	for x < px+30 {
		y := py
		for y < py+30 {
			if GetColor(hWnd, x, y) != COLOR_NS_INVENTORY_SLOT_EMPTY {
				return false
			}
			y += 5
		}
		x += 5
	}
	return true
}

func getItemPos(hWnd win.HWND, px, py int32, color win.COLORREF, granularity int32) (int32, int32, bool) {
	MoveCursorToNowhere(hWnd)

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

// deprecated
// type pos struct {
// 	x, y  int32
// 	found bool
// }

// func getItemPosWithThreads(hWnd win.HWND, px, py int32, color win.COLORREF, granularity int32) (int32, int32, bool) {
// 	MoveCursorToNowhere(hWnd)

// 	x := px
// 	y := py

// 	var wg sync.WaitGroup
// 	wg.Add(4)

// 	var i, j int32
// 	target := pos{}

// 	for j = 0; j < 4; j++ {
// 		go func(j int32, wg *sync.WaitGroup) {
// 			defer wg.Done()

// 			for i = 0; i < 5; i++ {
// 				if tx, ty, found := searchSlotForColor(hWnd, x+i*ITEM_COL_LEN, y+j*ITEM_COL_LEN, color, granularity); found {
// 					target = pos{tx, ty, found}
// 					return
// 				}
// 			}
// 		}(j, &wg)
// 	}

// 	wg.Wait()
// 	return target.x, target.y, target.found
// }

func searchSlotForColor(hWnd win.HWND, px, py int32, color win.COLORREF, granularity int32) (int32, int32, bool) {
	x := px
	for x < px+30 {
		y := py
		for y < py+30 {
			currentColor := GetColor(hWnd, x, y)
			if currentColor == color {
				return x, y, true
			} else if currentColor == COLOR_ITEM_CAN_NOT_BE_USED {
				return 0, 0, false
			}
			y += granularity
		}
		x += granularity
	}
	return 0, 0, false
}

func isInventoryFull(hWnd win.HWND) bool {
	defer closeAllWindows(hWnd)
	defer time.Sleep(DURATION_INVENTORY_CHECKER_WAITING)

	time.Sleep(DURATION_BATTLE_RESULT_DISAPPEARING)
	closeAllWindows(hWnd)

	LeftClick(hWnd, GAME_WIDTH/2, GAME_HEIGHT/2)

	openInventory(hWnd)

	if px, py, ok := getItemWindowPos(hWnd); ok {
		return !isAnyInventorySlotFree(hWnd, px, py)
	}
	return false
}

func isInventoryFullWithoutClosingAllWindows(hWnd win.HWND) bool {
	defer switchWindow(hWnd, 0x45)
	switchWindow(hWnd, 0x45)

	if px, py, ok := getItemWindowPos(hWnd); ok {
		return !isAnyInventorySlotFree(hWnd, px, py)
	}
	return false
}

func getMapName(hWnd win.HWND) string {
	return ReadMemoryString(hWnd, MEMORY_MAP_NAME, 32)
}

type GamePos struct {
	x, y float64
}

func getCurrentGamePos(hWnd win.HWND) GamePos {
	fx := ReadMemoryFloat32(hWnd, MEMORY_MAP_POS_X, 32)
	fy := ReadMemoryFloat32(hWnd, MEMORY_MAP_POS_Y, 32)
	return GamePos{float64(fx / 64), float64(fy / 64)}
}

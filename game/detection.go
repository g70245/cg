package game

import (
	. "cg/internal"

	"strconv"
	"strings"
	"sync"
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

const (
	COLOR_ANY = 0

	COLOR_SCENE_NORMAL = 15595514
	COLOR_SCENE_BATTLE = 15595514

	COLOR_MENU_BUTTON_NORMAL     = 15135992
	COLOR_MENU_BUTTON_POPOUT     = 10331818
	COLOR_MENU_BUTTON_CONTACT    = 15201528
	COLOR_MENU_BUTTON_PET_POPOUT = 10331817
	COLOR_MENU_HIDDEN            = 7568253

	COLOR_BATTLE_COMMAND_ENABLE  = 7125907
	COLOR_BATTLE_COMMAND_DISABLE = 6991316

	COLOR_BATTLE_STAGE_HUMAN = 15398392
	COLOR_BATTLE_STAGE_PET   = 8599608

	COLOR_BATTLE_BLOOD_UPPER      = 9211135
	COLOR_BATTLE_BLOOD_LOWER      = 255
	COLOR_BATTLE_MANA_UPPER       = 16758653
	COLOR_BATTLE_MANA_LOWER       = 16740864
	COLOR_BATTLE_BLOOD_MANA_EMPTY = 65536

	COLOR_BATTLE_RECALL_BUTTON = 7694643
	COLOR_BATTLE_NAME          = 37083

	COLOR_WINDOW_SKILL_UNSELECTED   = 4411988
	COLOR_WINDOW_SKILL_BOTTOM_SPACE = 11575428

	COLOR_NS_INVENTORY_SLOT_EMPTY = 15793151
	COLOR_PR_INVENTORY_SLOT_EMPTY = 15202301

	COLOR_NS_INVENTORY_PIVOT = 15967
	COLOR_BS_INVENTORY_PIVOT = 15967
	COLOR_PR_INVENTORY_PIVOT = 11113016

	COLOR_PR_PRODUCE_BUTTON = 7683891
	COLOR_PR_NOT_PRODUCING  = 11390937

	COLOR_ITEM_CAN_NOT_BE_USED = 255
	COLOR_ITEM_BOMB_7B         = 10936306
	COLOR_ITEM_BOMB_8B         = 14614527 // 8388607, 4194303
	COLOR_ITEM_BOMB_9A         = 30719    // 5626258
	COLOR_ITEM_POTION          = 16448250 // 8948665
)

var (
	NOWHERE_SCENE = CheckTarget{}
	NORMAL_SCENE  = CheckTarget{92, 26, COLOR_SCENE_NORMAL}
	BATTLE_SCENE  = CheckTarget{108, 10, COLOR_SCENE_BATTLE}

	MON_POS_T_1 = CheckTarget{26, 238, COLOR_ANY}
	MON_POS_T_2 = CheckTarget{92, 204, COLOR_ANY}
	MON_POS_T_3 = CheckTarget{162, 171, COLOR_ANY}
	MON_POS_T_4 = CheckTarget{225, 133, COLOR_ANY}
	MON_POS_T_5 = CheckTarget{282, 94, COLOR_ANY}
	MON_POS_B_1 = CheckTarget{93, 289, COLOR_ANY}
	MON_POS_B_2 = CheckTarget{156, 254, COLOR_ANY}
	MON_POS_B_3 = CheckTarget{216, 218, COLOR_ANY}
	MON_POS_B_4 = CheckTarget{284, 187, COLOR_ANY}
	MON_POS_B_5 = CheckTarget{343, 148, COLOR_ANY}

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

	BATTLE_COMMAND_ATTACK  = CheckTarget{386, 28, COLOR_ANY}
	BATTLE_COMMAND_DEFENCE = CheckTarget{386, 54, COLOR_ANY}
	BATTLE_COMMAND_SKILL   = CheckTarget{450, 28, COLOR_ANY}
	BATTLE_COMMAND_ITEM    = CheckTarget{450, 54, COLOR_ANY}
	BATTLE_COMMAND_PET     = CheckTarget{524, 28, COLOR_ANY}
	BATTLE_COMMAND_MOVE    = CheckTarget{524, 54, COLOR_ANY}
	BATTLE_COMMAND_ESCAPE  = CheckTarget{594, 54, COLOR_ANY}

	BATTLE_COMMAND_PET_SKILL_RIDING = CheckTarget{524, 54, COLOR_ANY}
	BATTLE_COMMAND_PET_SKILL_ESCAPE = CheckTarget{594, 54, COLOR_ANY}

	BATTLE_STAGE_HUMAN = CheckTarget{594, 28, COLOR_BATTLE_STAGE_HUMAN}
	BATTLE_STAGE_PET   = CheckTarget{594, 28, COLOR_BATTLE_STAGE_PET}

	BATTLE_WINDOW_SKILL_FIRST       = CheckTarget{154, 132, COLOR_WINDOW_SKILL_UNSELECTED}
	BATTLE_WINDOW_PET_RECALL_BUTTON = CheckTarget{384, 280, COLOR_ANY}

	BATTLE_WINDOW_ITEM_MONEY_PIVOT = CheckTarget{196, 114, COLOR_BS_INVENTORY_PIVOT}
	NORMAL_WINDOW_ITEM_MONEY_PIVOT = CheckTarget{354, 134, COLOR_NS_INVENTORY_PIVOT}
	PRODUCTION_WINDOW_ITEM_PIVOT   = CheckTarget{560, 100, COLOR_PR_INVENTORY_PIVOT}

	PLAYER_L_1_H = CheckTarget{329, 431, COLOR_ANY}
	PLAYER_L_2_H = CheckTarget{394, 396, COLOR_ANY}
	PLAYER_L_3_H = CheckTarget{460, 361, COLOR_ANY}
	PLAYER_L_4_H = CheckTarget{524, 326, COLOR_ANY}
	PLAYER_L_5_H = CheckTarget{589, 291, COLOR_ANY}
	PLAYER_L_1_P = CheckTarget{269, 386, COLOR_ANY}
	PLAYER_L_2_P = CheckTarget{333, 350, COLOR_ANY}
	PLAYER_L_3_P = CheckTarget{397, 314, COLOR_ANY}
	PLAYER_L_4_P = CheckTarget{460, 277, COLOR_ANY}
	PLAYER_L_5_P = CheckTarget{524, 241, COLOR_ANY}
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

	if counter > 2 {
		return true
	}
	return false
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

type pos struct {
	x, y  int32
	found bool
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
func getItemPosWithThreads(hWnd win.HWND, px, py int32, color win.COLORREF, granularity int32) (int32, int32, bool) {
	MoveCursorToNowhere(hWnd)

	x := px
	y := py

	var wg sync.WaitGroup
	wg.Add(4)

	var i, j int32
	target := pos{}

	for j = 0; j < 4; j++ {
		go func(j int32, wg *sync.WaitGroup) {
			defer wg.Done()

			for i = 0; i < 5; i++ {
				if tx, ty, found := searchSlotForColor(hWnd, x+i*ITEM_COL_LEN, y+j*ITEM_COL_LEN, color, granularity); found {
					target = pos{tx, ty, found}
					return
				}
			}
		}(j, &wg)
	}

	wg.Wait()
	return target.x, target.y, target.found
}

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

var (
	LOG_TELEPORTING        = []string{"被不可思", "你感覺到一股"}
	LOG_OUT_OF_RESOURCE    = []string{"道具已經用完了"}
	LOG_ACTIVITY           = []string{"發現野生一級", "南瓜之王", "虎王"}
	LOG_PRODUCTION_FAILURE = []string{}
)

const (
	DURATION_LOG_ACTIVITY        = 5 * time.Second
	DURATION_LOG_TELEPORTING     = 30 * time.Second
	DURATION_LOG_OUT_OF_RESOURCE = 30 * time.Second
)

func doesEncounterActivityMonsters(dir string) bool {
	if dir == "" {
		return false
	}

	return checkWord(dir, 5, DURATION_LOG_ACTIVITY, LOG_ACTIVITY)
}

func isTeleported(dir string) bool {
	if dir == "" {
		return false
	}
	return checkWord(dir, 5, DURATION_LOG_TELEPORTING, LOG_TELEPORTING)
}

func isOutOfResource(dir string) bool {
	if dir == "" {
		return false
	}
	return checkWord(dir, 5, 30*time.Second, LOG_OUT_OF_RESOURCE)
}

func checkWord(dir string, lineCount int, before time.Duration, words []string) bool {
	lines := GetLastLinesOfLog(dir, lineCount)
	now := time.Now()
	for i := range lines {
		h, hErr := strconv.Atoi(lines[i][1:3])
		m, mErrr := strconv.Atoi(lines[i][4:6])
		s, sErr := strconv.Atoi(lines[i][7:9])
		if hErr != nil || mErrr != nil || sErr != nil {
			continue
		}

		logTime := time.Date(now.Year(), now.Month(), now.Day(), h, m, s, 0, time.Local)
		for j := range words {
			if !logTime.Before(now.Add(-before)) && strings.Contains(lines[i], words[j]) {
				return true
			}
		}
	}
	return false
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

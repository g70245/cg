package game

import (
	sys "cg/system"
	"strings"

	. "github.com/lxn/win"
)

const (
	GAME_WIDTH  = 640
	GAME_HEIGHT = 480
)

type CheckTarget struct {
	x     int32
	y     int32
	color COLORREF
}

func (c *CheckTarget) GetX() int32 {
	return c.x
}

func (c *CheckTarget) GetY() int32 {
	return c.y
}

func (c *CheckTarget) Set(x, y int32) {
	c.x = x
	c.y = y
}

const (
	COLOR_ANY = 0

	COLOR_SCENE_NORMAL = 15595514
	COLOR_SCENE_BATTLE = 15595514

	COLOR_MENU_BUTTON_NORMAL     = 15135992
	COLOR_MENU_BUTTON_T          = 15201528
	COLOR_MENU_BUTTON_POPOUT     = 10331818
	COLOR_MENU_BUTTON_R_POPOUT   = 10331817
	COLOR_BATTLE_COMMAND_ENABLE  = 7125907
	COLOR_BATTLE_COMMAND_DISABLE = 6991316

	COLOR_BATTLE_STAGE_HUMAN = 15398392
	COLOR_BATTLE_STAGE_PET   = 8599608

	// COLOR_WINDOW_SKILL_TOP        = 65536
	COLOR_WINDOW_SKILL_UNSELECTED = 4411988
	COLOR_HUMAN_OUT_OF_MANA       = 11575428

	COLOR_BATTLE_BLOOD_UPPER   = 921135
	COLOR_BATTLE_BLOOD_LOWER   = 255
	COLOR_BATTLE_MANA_UPPER    = 16758653
	COLOR_BATTLE_MANA_LOWER    = 16740864
	COLOR_BATTLE_NO_BLOOD_MANA = 65536

	COLOR_BATTLE_ITEM_CAN_NOT_BE_USED = 255
	COLOR_WINDOW_ITEM_PIVOT           = 16777215

	COLOR_ITEM_BOMB_9A = 8388607
)

var (
	NOWHERE_SCENE = CheckTarget{}
	NORMAL_SCENE  = CheckTarget{92, 26, COLOR_SCENE_NORMAL}
	BATTLE_SCENE  = CheckTarget{108, 10, COLOR_SCENE_BATTLE}

	MON_POS_T_1 = CheckTarget{28, 256, COLOR_ANY}
	MON_POS_T_2 = CheckTarget{90, 218, COLOR_ANY}
	MON_POS_T_3 = CheckTarget{156, 184, COLOR_ANY}
	MON_POS_T_4 = CheckTarget{216, 146, COLOR_ANY}
	MON_POS_T_5 = CheckTarget{282, 112, COLOR_ANY}
	MON_POS_B_1 = CheckTarget{92, 310, COLOR_ANY}
	MON_POS_B_2 = CheckTarget{152, 268, COLOR_ANY}
	MON_POS_B_3 = CheckTarget{214, 226, COLOR_ANY}
	MON_POS_B_4 = CheckTarget{278, 192, COLOR_ANY}
	MON_POS_B_5 = CheckTarget{346, 162, COLOR_ANY}

	MENU_Q          = CheckTarget{60, 468, COLOR_MENU_BUTTON_NORMAL}
	MENU_W          = CheckTarget{136, 468, COLOR_MENU_BUTTON_NORMAL}
	MENU_E          = CheckTarget{212, 468, COLOR_MENU_BUTTON_NORMAL}
	MENU_R          = CheckTarget{288, 468, COLOR_MENU_BUTTON_NORMAL}
	MENU_T          = CheckTarget{366., 468, COLOR_MENU_BUTTON_T}
	MENU_Y          = CheckTarget{442, 468, COLOR_MENU_BUTTON_NORMAL}
	MENU_ESC        = CheckTarget{514, 468, COLOR_MENU_BUTTON_NORMAL}
	MENU_Q_POPOUT   = CheckTarget{60, 468, COLOR_MENU_BUTTON_POPOUT}
	MENU_W_POPOUT   = CheckTarget{136, 468, COLOR_MENU_BUTTON_POPOUT}
	MENU_E_POPOUT   = CheckTarget{212, 468, COLOR_MENU_BUTTON_POPOUT}
	MENU_R_POPOUT   = CheckTarget{288, 468, COLOR_MENU_BUTTON_R_POPOUT}
	MENU_T_POPOUT   = CheckTarget{366., 468, COLOR_MENU_BUTTON_POPOUT}
	MENU_Y_POPOUT   = CheckTarget{442, 468, COLOR_MENU_BUTTON_POPOUT}
	MENU_ESC_POPOUT = CheckTarget{514, 468, COLOR_MENU_BUTTON_POPOUT}

	BATTLE_COMMAND_ATTACK  = CheckTarget{386, 28, COLOR_ANY}
	BATTLE_COMMAND_DEFENCE = CheckTarget{386, 54, COLOR_ANY}
	BATTLE_COMMAND_SKILL   = CheckTarget{450, 28, COLOR_ANY}
	BATTLE_COMMAND_PET     = CheckTarget{524, 28, COLOR_ANY}
	BATTLE_COMMAND_MOVE    = CheckTarget{524, 54, COLOR_ANY}
	BATTLE_COMMAND_ESCAPE  = CheckTarget{594, 54, COLOR_ANY}

	BATTLE_STAGE_HUMAN = CheckTarget{594, 28, COLOR_BATTLE_STAGE_HUMAN}
	BATTLE_STAGE_PET   = CheckTarget{594, 28, COLOR_BATTLE_STAGE_PET}

	WINDOW_SKILL_FIRST      = CheckTarget{140, 120, COLOR_WINDOW_SKILL_UNSELECTED}
	WINDOW_ITEM_MONEY_CLUMN = CheckTarget{140, 120, COLOR_WINDOW_ITEM_PIVOT}

	PLAYER_L_1_H = CheckTarget{329, 457 - 26, COLOR_ANY}
	PLAYER_L_2_H = CheckTarget{394, 422 - 26, COLOR_ANY}
	PLAYER_L_3_H = CheckTarget{460, 387 - 26, COLOR_ANY}
	PLAYER_L_4_H = CheckTarget{524, 352 - 26, COLOR_ANY}
	PLAYER_L_5_H = CheckTarget{589, 317 - 26, COLOR_ANY}
	PLAYER_L_1_P = CheckTarget{269, 412 - 26, COLOR_ANY}
	PLAYER_L_2_P = CheckTarget{333, 376 - 26, COLOR_ANY}
	PLAYER_L_3_P = CheckTarget{397, 340 - 26, COLOR_ANY}
	PLAYER_L_4_P = CheckTarget{460, 303 - 26, COLOR_ANY}
	PLAYER_L_5_P = CheckTarget{524, 267 - 26, COLOR_ANY}
)

func getScene(hWnd HWND) CheckTarget {
	if sys.GetColor(hWnd, NORMAL_SCENE.x, NORMAL_SCENE.y) == NORMAL_SCENE.color {
		return NORMAL_SCENE
	} else if sys.GetColor(hWnd, BATTLE_SCENE.x, BATTLE_SCENE.y) == BATTLE_SCENE.color {
		return BATTLE_SCENE
	}
	return NOWHERE_SCENE
}

var popOutMenuCheckList = []CheckTarget{MENU_Q_POPOUT, MENU_W_POPOUT, MENU_E_POPOUT, MENU_R_POPOUT, MENU_T_POPOUT, MENU_Y_POPOUT}

func getPopOutMenus(hWnd HWND) (popOutMenus []CheckTarget) {
	for _, target := range popOutMenuCheckList {
		if sys.GetColor(hWnd, target.x, target.y) == target.color {
			popOutMenus = append(popOutMenus, target)
		}
	}
	return
}

func isBattleCommandEnable(hWnd HWND, checkTarget CheckTarget) bool {
	return sys.GetColor(hWnd, checkTarget.x, checkTarget.y) == COLOR_BATTLE_COMMAND_ENABLE
}

func isHumanStageStable(hWnd HWND) bool {
	return sys.GetColor(hWnd, BATTLE_STAGE_HUMAN.x, BATTLE_STAGE_HUMAN.y) == BATTLE_STAGE_HUMAN.color
}

func isPetStageStable(hWnd HWND) bool {
	return sys.GetColor(hWnd, BATTLE_STAGE_PET.x, BATTLE_STAGE_PET.y) == BATTLE_STAGE_PET.color
}

func isPetSkillWindowOpend(hWnd HWND) bool {
	return sys.GetColor(hWnd, BATTLE_COMMAND_ESCAPE.x, BATTLE_COMMAND_ESCAPE.y) == COLOR_BATTLE_COMMAND_ENABLE
}

func didHumanAttack(hWnd HWND) bool {
	return sys.GetColor(hWnd, BATTLE_STAGE_HUMAN.x, BATTLE_STAGE_HUMAN.y) != BATTLE_STAGE_HUMAN.color
}

func didPetAttack(hWnd HWND) bool {
	return sys.GetColor(hWnd, BATTLE_STAGE_PET.x, BATTLE_STAGE_PET.y) != BATTLE_STAGE_PET.color
}

func getSkillWindowPos(hWnd HWND) (int32, int32, bool) {
	x := WINDOW_SKILL_FIRST.x
	for x <= 180 {
		y := WINDOW_SKILL_FIRST.y
		for y <= 232 {
			if sys.GetColor(hWnd, x, y) == WINDOW_SKILL_FIRST.color {
				return x, y, true
			}
			y += 4
		}
		x += 4
	}
	return 0, 0, false
}

func getItemWindowPos(hWnd HWND) (int32, int32, bool) {
	x := WINDOW_ITEM_MONEY_CLUMN.x
	for x <= 160 {
		y := WINDOW_ITEM_MONEY_CLUMN.y
		for y <= 166 {
			if sys.GetColor(hWnd, x, y) == WINDOW_ITEM_MONEY_CLUMN.color {
				return x, y, true
			}
			y += 2
		}
		x += 2
	}
	return 0, 0, false
}

func getItemPos(hWnd HWND, px, py int32, color COLORREF) (int32, int32, bool) {
	x := px - 10
	for x <= px+45*5 {
		y := py + 10
		for y <= py+45*4 {
			if sys.GetColor(hWnd, x, y) == color {
				return x, y, true
			}
			y += 4
		}
		x += 4
	}
	return 0, 0, false
}

func isHumanOutOfMana(hWnd HWND, x, y int32) bool {
	sys.LeftClick(hWnd, GAME_WIDTH/2, 28)
	if sys.GetColor(hWnd, x, y) == COLOR_HUMAN_OUT_OF_MANA {
		return true
	}
	return false
}

func isPetOutOfMana(hWnd HWND) bool {
	if sys.GetColor(hWnd, BATTLE_COMMAND_ESCAPE.x, BATTLE_COMMAND_ESCAPE.y) == COLOR_BATTLE_COMMAND_ENABLE {
		return true
	}
	return false
}

var stopWords = []string{"被不可思"}

func isTransmittedToOtherMap(dir string) bool {
	return strings.Contains(sys.GetLastLineOfLog(dir), stopWords[0])
}

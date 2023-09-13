package main

import (
	sys "cg/system"

	"github.com/lxn/win"
)

type CheckTarget struct {
	x     int32
	y     int32
	color win.COLORREF
}

const (
	COLOR_ANY = 0

	COLOR_SKILL        = 65536
	COLOR_SCENE_NORMAL = 15595514

	COLOR_MENU_BUTTON_NORMAL     = 15135992
	COLOR_MENU_BUTTON_T          = 15201528
	COLOR_MENU_BUTTON_POPOUT     = 10331818
	COLOR_MENU_BUTTON_R_POPOUT   = 10331817
	COLOR_BATTLE_COMMAND_ENABLE  = 7125907
	COLOR_BATTLE_COMMAND_DISABLE = 6991316

	COLOR_BATTLE_STAGE_PET = 15267320
)

var (
	NORMAL_SCENE = CheckTarget{92, 26, COLOR_SCENE_NORMAL}
	BATTLE_SCENE = CheckTarget{}

	MON_POS_T_1 = CheckTarget{28, 260, COLOR_ANY}
	MON_POS_T_2 = CheckTarget{94, 224, COLOR_ANY}
	MON_POS_T_3 = CheckTarget{160, 188, COLOR_ANY}
	MON_POS_T_4 = CheckTarget{220, 150, COLOR_ANY}
	MON_POS_T_5 = CheckTarget{282, 114, COLOR_ANY}
	MON_POS_B_1 = CheckTarget{100, 312, COLOR_ANY}
	MON_POS_B_2 = CheckTarget{160, 268, COLOR_ANY}
	MON_POS_B_3 = CheckTarget{214, 226, COLOR_ANY}
	MON_POS_B_4 = CheckTarget{290, 200, COLOR_ANY}
	MON_POS_B_5 = CheckTarget{348, 166, COLOR_ANY}

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

	BATTLE_COMMAND_ATTACK = CheckTarget{386, 28, COLOR_ANY}
	BATTLE_COMMAND_SKILL  = CheckTarget{450, 54, COLOR_ANY}
	BATTLE_COMMAND_PET    = CheckTarget{594, 54, COLOR_ANY}
	BATTLE_COMMAND_ESCAPE = CheckTarget{594, 54, COLOR_ANY}

	BATTLE_STAGE_PET = CheckTarget{594, 54, COLOR_BATTLE_STAGE_PET}
)

func GetScene(hWnd win.HWND) CheckTarget {
	if sys.GetColor(hWnd, NORMAL_SCENE.x, NORMAL_SCENE.y) == NORMAL_SCENE.color {
		return NORMAL_SCENE
	}
	return BATTLE_SCENE
}

var popOutMenuCheckList = []CheckTarget{MENU_Q_POPOUT, MENU_W_POPOUT, MENU_E_POPOUT, MENU_R_POPOUT, MENU_T_POPOUT, MENU_Y_POPOUT}

func GetPopOutMenus(hWnd win.HWND) (popOutMenus []CheckTarget) {
	for _, target := range popOutMenuCheckList {
		if sys.GetColor(hWnd, target.x, target.y) == target.color {
			popOutMenus = append(popOutMenus, target)
		}
	}
	return
}

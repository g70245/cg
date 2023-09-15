package game

import (
	sys "cg/system"
	"fmt"

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

	COLOR_WINDOW_SKILL_TOP        = 65536
	COLOR_WINDOW_SKILL_UNSELECTED = 4411988
)

var (
	NOWHERE_SCENE = CheckTarget{}
	NORMAL_SCENE  = CheckTarget{92, 26, COLOR_SCENE_NORMAL}
	BATTLE_SCENE  = CheckTarget{108, 10, COLOR_SCENE_BATTLE}

	MON_POS_T_1 = CheckTarget{28, 256, COLOR_ANY}
	MON_POS_T_2 = CheckTarget{90, 218, COLOR_ANY}
	MON_POS_T_3 = CheckTarget{156, 184, COLOR_ANY}
	MON_POS_T_4 = CheckTarget{216, 144, COLOR_ANY}
	MON_POS_T_5 = CheckTarget{282, 112, COLOR_ANY}
	MON_POS_B_1 = CheckTarget{100, 308, COLOR_ANY}
	MON_POS_B_2 = CheckTarget{160, 268, COLOR_ANY}
	MON_POS_B_3 = CheckTarget{214, 226, COLOR_ANY}
	MON_POS_B_4 = CheckTarget{284, 190, COLOR_ANY}
	MON_POS_B_5 = CheckTarget{348, 160, COLOR_ANY}

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
	BATTLE_COMMAND_PET    = CheckTarget{524, 28, COLOR_ANY}
	BATTLE_COMMAND_ESCAPE = CheckTarget{594, 54, COLOR_ANY}

	BATTLE_STAGE_HUMAN = CheckTarget{594, 28, COLOR_BATTLE_STAGE_HUMAN}
	BATTLE_STAGE_PET   = CheckTarget{594, 28, COLOR_BATTLE_STAGE_PET}

	WINDOW_SKILL = CheckTarget{136, 126, COLOR_WINDOW_SKILL_TOP}
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
	x, y := WINDOW_SKILL.x, WINDOW_SKILL.y
	for x <= 146 {
		for y <= 188 {
			if sys.GetColor(hWnd, x, y) == WINDOW_SKILL.color {
				if sys.GetColor(hWnd, x, y+16) == COLOR_WINDOW_SKILL_UNSELECTED {
					fmt.Println(x, y)
					return x, y, true
				}
			}
			y += 5
		}
		x += 5
	}
	return 0, 0, false
}

func isOutOfMana(hWnd HWND, x, y int32) bool {
	if sys.GetColor(hWnd, x, y) == WINDOW_SKILL.color {
		return true
	}
	return false
}

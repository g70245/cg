package game

import (
	sys "cg/system"
	"strings"

	. "github.com/lxn/win"
)

const (
	GAME_WIDTH   = 640
	GAME_HEIGHT  = 480
	ITEM_COL_LEN = 50
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

	COLOR_MENU_BUTTON_NORMAL   = 15135992
	COLOR_MENU_BUTTON_T        = 15201528
	COLOR_MENU_BUTTON_POPOUT   = 10331818
	COLOR_MENU_BUTTON_R_POPOUT = 10331817

	COLOR_BATTLE_COMMAND_ENABLE  = 7125907
	COLOR_BATTLE_COMMAND_DISABLE = 6991316

	COLOR_BATTLE_STAGE_HUMAN = 15398392
	COLOR_BATTLE_STAGE_PET   = 8599608

	COLOR_BATTLE_BLOOD_UPPER   = 9211135
	COLOR_BATTLE_BLOOD_LOWER   = 255
	COLOR_BATTLE_MANA_UPPER    = 16758653
	COLOR_BATTLE_MANA_LOWER    = 16740864
	COLOR_BATTLE_NO_BLOOD_MANA = 65536
	COLOR_BATTLE_RECALL_BUTTON = 7694643
	COLOR_BATTLE_SELF_TITLE    = 37083

	COLOR_WINDOW_SKILL_UNSELECTED        = 4411988
	COLOR_WINDOW_SKILL_HUMAN_OUT_OF_MANA = 11575428

	COLOR_WINDOW_ITEM_CAN_NOT_BE_USED = 255
	COLOR_WINDOW_ITEM_EMPTY           = 15793151
	COLOR_NS_ITEM_PIVOT               = 15967
	COLOR_BS_ITEM_PIVOT               = 15967 //6190476 //16777215

	COLOR_ITEM_BOMB_8B = 8388607
	COLOR_ITEM_BOMB_9A = 13974896
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

	BATTLE_COMMAND_PET_SKILL_RIDING = CheckTarget{524, 54, COLOR_ANY}
	BATTLE_COMMAND_PET_SKILL_ESCAPE = CheckTarget{594, 54, COLOR_ANY}

	BATTLE_STAGE_HUMAN = CheckTarget{594, 28, COLOR_BATTLE_STAGE_HUMAN}
	BATTLE_STAGE_PET   = CheckTarget{594, 28, COLOR_BATTLE_STAGE_PET}

	BATTLE_WINDOW_SKILL_FIRST       = CheckTarget{156, 140, COLOR_WINDOW_SKILL_UNSELECTED}
	BATTLE_WINDOW_ITEM_MONEY_CLUMN  = CheckTarget{196, 114, COLOR_BS_ITEM_PIVOT}
	BATTLE_WINDOW_PET_RECALL_BUTTON = CheckTarget{384, 280, COLOR_ANY}

	NORMAL_WINDOW_ITEM_MONEY_CLUMN = CheckTarget{348, 144, COLOR_NS_ITEM_PIVOT}

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

func isPetSkillWindowOpendWhileRiding(hWnd HWND) bool {
	return isPetSkillWindowOpendWhileRiding(hWnd)
}

func HumanTargetingChecker(hWnd HWND) bool {
	return sys.GetColor(hWnd, BATTLE_STAGE_HUMAN.x, BATTLE_STAGE_HUMAN.y) != BATTLE_STAGE_HUMAN.color
}

func PetTargetingChecker(hWnd HWND) bool {
	return sys.GetColor(hWnd, BATTLE_STAGE_PET.x, BATTLE_STAGE_PET.y) != BATTLE_STAGE_PET.color
}

func getSkillWindowPos(hWnd HWND) (int32, int32, bool) {
	x := BATTLE_WINDOW_SKILL_FIRST.x
	for x <= 164 {
		y := BATTLE_WINDOW_SKILL_FIRST.y
		for y <= 232 {
			if sys.GetColor(hWnd, x, y) == BATTLE_WINDOW_SKILL_FIRST.color {
				return x, y, true
			}
			y += 2
		}
		x += 2
	}
	return 0, 0, false
}

func getBSItemWindowPos(hWnd HWND) (int32, int32, bool) {
	x := BATTLE_WINDOW_ITEM_MONEY_CLUMN.x
	for x <= BATTLE_WINDOW_ITEM_MONEY_CLUMN.x+50 {
		y := BATTLE_WINDOW_ITEM_MONEY_CLUMN.y
		for y <= BATTLE_WINDOW_ITEM_MONEY_CLUMN.y+50 {
			if sys.GetColor(hWnd, x, y) == BATTLE_WINDOW_ITEM_MONEY_CLUMN.color {
				return x - 78, y + 20, true
			}
			y += 2
		}
		x += 2
	}
	return 0, 0, false
}

func getNSItemWindowPos(hWnd HWND) (int32, int32, bool) {
	x := NORMAL_WINDOW_ITEM_MONEY_CLUMN.x
	for x <= NORMAL_WINDOW_ITEM_MONEY_CLUMN.x+54 {
		y := NORMAL_WINDOW_ITEM_MONEY_CLUMN.y
		for y <= NORMAL_WINDOW_ITEM_MONEY_CLUMN.y+44 {
			if sys.GetColor(hWnd, x, y) == NORMAL_WINDOW_ITEM_MONEY_CLUMN.color {
				return x, y + 20, true
			}
			y += 2
		}
		x += 2
	}
	return 0, 0, false
}

func isAnyItemSlotFree(hWnd HWND, px, py int32) bool {

	x := px
	y := py
	var i, j int32

	for i = 0; i < 5; i++ {
		for j = 0; j < 4; j++ {
			if isSlotEmpty(hWnd, x+i*ITEM_COL_LEN, y+j*ITEM_COL_LEN) {
				return true
			}
		}
	}

	return false
}

func isSlotEmpty(hWnd HWND, px, py int32) bool {
	x := px
	for x < px+30 {
		y := py
		for y < py+30 {
			if sys.GetColor(hWnd, x, y) != COLOR_WINDOW_ITEM_EMPTY {
				return false
			}
			y += 5
		}
		x += 5
	}
	return true
}

func getItemPos(hWnd HWND, px, py int32, color COLORREF) (int32, int32, bool) {
	x := px
	y := py
	var i, j int32

	for i = 0; i < 5; i++ {
		for j = 0; j < 4; j++ {
			if tx, ty, found := searchSlotForColor(hWnd, x+i*ITEM_COL_LEN, y+j*ITEM_COL_LEN, color); found {
				return tx, ty, found
			}
		}
	}

	return 0, 0, false
}

func searchSlotForColor(hWnd HWND, px, py int32, color COLORREF) (int32, int32, bool) {
	x := px
	for x < px+30 {
		y := py
		for y < py+30 {
			if sys.GetColor(hWnd, x, y) == color {
				return x, y, true
			}
			y += 5
		}
		x += 5
	}
	return 0, 0, false
}

func isHumanOutOfMana(hWnd HWND, x, y int32) bool {
	if sys.GetColor(hWnd, x, y+16*10) == COLOR_WINDOW_SKILL_HUMAN_OUT_OF_MANA {
		return true
	}
	return false
}

func isPetOutOfMana(hWnd HWND) bool {
	if sys.GetColor(hWnd, BATTLE_COMMAND_PET_SKILL_ESCAPE.x, BATTLE_COMMAND_PET_SKILL_ESCAPE.y) == COLOR_BATTLE_COMMAND_ENABLE {
		return true
	}
	return false
}

func isRidingOutOfMana(hWnd HWND) bool {
	if sys.GetColor(hWnd, BATTLE_COMMAND_PET_SKILL_RIDING.x, BATTLE_COMMAND_PET_SKILL_RIDING.y) == COLOR_BATTLE_COMMAND_ENABLE {
		return true
	}
	return false
}

func isOnRide(hWnd HWND) bool {
	return sys.GetColor(hWnd, BATTLE_COMMAND_PET_SKILL_ESCAPE.x, BATTLE_COMMAND_PET_SKILL_ESCAPE.y) == COLOR_BATTLE_COMMAND_DISABLE &&
		(sys.GetColor(hWnd, BATTLE_COMMAND_PET_SKILL_RIDING.x, BATTLE_COMMAND_PET_SKILL_RIDING.y) == COLOR_BATTLE_COMMAND_DISABLE ||
			sys.GetColor(hWnd, BATTLE_COMMAND_PET_SKILL_RIDING.x, BATTLE_COMMAND_PET_SKILL_RIDING.y) == COLOR_BATTLE_COMMAND_ENABLE)
}

var stopWords = []string{"被不可思"}

func isTPedToOtherMap(dir string) bool {
	return strings.Contains(sys.GetLastLineOfLog(dir), stopWords[0])
}

func canRecall(hWnd HWND) bool {
	return sys.GetColor(hWnd, BATTLE_WINDOW_PET_RECALL_BUTTON.x, BATTLE_WINDOW_PET_RECALL_BUTTON.y) == COLOR_BATTLE_RECALL_BUTTON
}

var allTargets = []CheckTarget{
	PLAYER_L_3_H,
	PLAYER_L_2_H,
	PLAYER_L_4_H,
	PLAYER_L_1_H,
	PLAYER_L_5_H,
	PLAYER_L_1_P,
	PLAYER_L_2_P,
	PLAYER_L_3_P,
	PLAYER_L_4_P,
	PLAYER_L_5_P,
}

func isLifeBelow(hWnd HWND, ratio float32, checkTarget *CheckTarget) bool {
	healthPoint := int32(ratio*30) + checkTarget.x
	return sys.GetColor(hWnd, healthPoint, checkTarget.y) != COLOR_BATTLE_BLOOD_UPPER &&
		sys.GetColor(hWnd, checkTarget.x, checkTarget.y) == COLOR_BATTLE_BLOOD_UPPER
}

func searchOneLifeBelow(hWnd HWND, ratio float32) (*CheckTarget, bool) {
	for i := range allTargets {
		if isLifeBelow(hWnd, ratio, &allTargets[i]) {
			return &allTargets[i], true
		}
	}
	return nil, false
}

func countLifeBelow(hWnd HWND, ratio float32) (count int) {
	for i := range allTargets {
		if isLifeBelow(hWnd, ratio, &allTargets[i]) {
			count++
		}
	}
	return
}

var allPlayers = []CheckTarget{
	PLAYER_L_1_H,
	PLAYER_L_2_H,
	PLAYER_L_3_H,
	PLAYER_L_4_H,
	PLAYER_L_5_H,
}

var allPets = []CheckTarget{
	PLAYER_L_1_P,
	PLAYER_L_2_P,
	PLAYER_L_3_P,
	PLAYER_L_4_P,
	PLAYER_L_5_P,
}

func getSelfTarget(hWnd HWND, isHuman bool) (*CheckTarget, bool) {
	targets := allPlayers
	if !isHuman {
		targets = allPets
	}

	for i := range targets {
		x := targets[i].x + 8
		for x <= targets[i].x+30 {
			y := targets[i].y - 10
			for y >= targets[i].y-26 {
				if sys.GetColor(hWnd, x, y) == COLOR_BATTLE_SELF_TITLE {
					return &targets[i], true
				}
				y--
			}
			x++
		}
	}
	return nil, false
}

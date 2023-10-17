package game

import (
	sys "cg/system"

	"strconv"
	"strings"
	"sync"
	"time"

	. "github.com/g70245/win"
)

const (
	GAME_WIDTH   = 640
	GAME_HEIGHT  = 480
	ITEM_COL_LEN = 50

	DETECT_MON_TARGET_MOUSE_MOV_INTERVAL = 80
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
	COLOR_MENU_HIDDEN          = 7568253

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

	COLOR_WINDOW_SKILL_UNSELECTED   = 4411988
	COLOR_WINDOW_SKILL_BOTTOM_SPACE = 11575428

	COLOR_WINDOW_ITEM_EMPTY = 15793151
	COLOR_NS_ITEM_PIVOT     = 15967
	COLOR_BS_ITEM_PIVOT     = 15967

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
	BATTLE_COMMAND_ITEM    = CheckTarget{450, 54, COLOR_ANY}
	BATTLE_COMMAND_PET     = CheckTarget{524, 28, COLOR_ANY}
	BATTLE_COMMAND_MOVE    = CheckTarget{524, 54, COLOR_ANY}
	BATTLE_COMMAND_ESCAPE  = CheckTarget{594, 54, COLOR_ANY}

	BATTLE_COMMAND_PET_SKILL_RIDING = CheckTarget{524, 54, COLOR_ANY}
	BATTLE_COMMAND_PET_SKILL_ESCAPE = CheckTarget{594, 54, COLOR_ANY}

	BATTLE_STAGE_HUMAN = CheckTarget{594, 28, COLOR_BATTLE_STAGE_HUMAN}
	BATTLE_STAGE_PET   = CheckTarget{594, 28, COLOR_BATTLE_STAGE_PET}

	BATTLE_WINDOW_SKILL_FIRST       = CheckTarget{154, 132, COLOR_WINDOW_SKILL_UNSELECTED}
	BATTLE_WINDOW_ITEM_MONEY_CLUMN  = CheckTarget{196, 114, COLOR_BS_ITEM_PIVOT}
	BATTLE_WINDOW_PET_RECALL_BUTTON = CheckTarget{384, 280, COLOR_ANY}

	NORMAL_WINDOW_ITEM_MONEY_CLUMN = CheckTarget{348, 144, COLOR_NS_ITEM_PIVOT}

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

func getScene(hWnd HWND) CheckTarget {
	if sys.GetColor(hWnd, NORMAL_SCENE.x, NORMAL_SCENE.y) == NORMAL_SCENE.color {
		return NORMAL_SCENE
	} else if sys.GetColor(hWnd, BATTLE_SCENE.x, BATTLE_SCENE.y) == BATTLE_SCENE.color {
		return BATTLE_SCENE
	}
	return NOWHERE_SCENE
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

func isPetSkillWindowOpened(hWnd HWND) bool {
	return sys.GetColor(hWnd, BATTLE_COMMAND_ESCAPE.x, BATTLE_COMMAND_ESCAPE.y) == COLOR_BATTLE_COMMAND_ENABLE
}

func isPetSkillWindowOpenedWhileRiding(hWnd HWND) bool {
	return isPetSkillWindowOpenedWhileRiding(hWnd)
}

func isHumanActionSuccessful(hWnd HWND) bool {
	sys.MoveToNowhere(hWnd)
	return sys.GetColor(hWnd, BATTLE_STAGE_HUMAN.x, BATTLE_STAGE_HUMAN.y) != BATTLE_STAGE_HUMAN.color
}

func isPetActionSuccessful(hWnd HWND) bool {
	sys.MoveToNowhere(hWnd)
	return sys.GetColor(hWnd, BATTLE_STAGE_PET.x, BATTLE_STAGE_PET.y) != BATTLE_STAGE_PET.color
}

func getSkillWindowPos(hWnd HWND) (int32, int32, bool) {
	sys.MoveToNowhere(hWnd)
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
	sys.MoveToNowhere(hWnd)
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
	sys.MoveToNowhere(hWnd)
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
	sys.MoveToNowhere(hWnd)
	x := px
	y := py
	var i, j int32

	for i = 0; i < 5; i++ {
		for j = 0; j < 4; j++ {
			if isSlotFree(hWnd, x+i*ITEM_COL_LEN, y+j*ITEM_COL_LEN) {
				return true
			}
		}
	}

	return false
}

func isSlotFree(hWnd HWND, px, py int32) bool {
	x := px
	for x < px+32 {
		y := py
		for y < py+32 {
			if sys.GetColor(hWnd, x, y) != COLOR_WINDOW_ITEM_EMPTY {
				return false
			}
			y += 4
		}
		x += 4
	}
	return true
}

func isItemWindowStuck(hWnd HWND) bool {
	return sys.GetColor(hWnd, BATTLE_COMMAND_ITEM.x, BATTLE_COMMAND_ITEM.y) == COLOR_BATTLE_COMMAND_ENABLE
}

func isBSItemWindowStillOpened(hWnd HWND, x, y int32) bool {
	return sys.GetColor(hWnd, x+78, y-20) == BATTLE_WINDOW_ITEM_MONEY_CLUMN.color
}

type pos struct {
	x, y  int32
	found bool
}

func getItemPos(hWnd HWND, px, py int32, color COLORREF, granularity int32) (int32, int32, bool) {
	sys.MoveToNowhere(hWnd)

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

func getItemPosByThreads(hWnd HWND, px, py int32, color COLORREF, granularity int32) (int32, int32, bool) {
	sys.MoveToNowhere(hWnd)

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

func searchSlotForColor(hWnd HWND, px, py int32, color COLORREF, granularity int32) (int32, int32, bool) {
	x := px
	for x < px+30 {
		y := py
		for y < py+30 {
			// fmt.Println(x, y, sys.GetColor(hWnd, x, y))
			currentColor := sys.GetColor(hWnd, x, y)
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

func doesHumanMissSkillButton(hWnd HWND, x, y int32) bool {
	return sys.GetColor(hWnd, x, y+16*10) == COLOR_WINDOW_SKILL_BOTTOM_SPACE
}

func doesPetMissSkillButton(hWnd HWND) bool {
	return sys.GetColor(hWnd, BATTLE_COMMAND_PET_SKILL_ESCAPE.x, BATTLE_COMMAND_PET_SKILL_ESCAPE.y) == COLOR_BATTLE_COMMAND_ENABLE
}

func doesOnRidingMissSkillButtton(hWnd HWND) bool {
	return sys.GetColor(hWnd, BATTLE_COMMAND_PET_SKILL_RIDING.x, BATTLE_COMMAND_PET_SKILL_RIDING.y) == COLOR_BATTLE_COMMAND_ENABLE
}

func isOnRide(hWnd HWND) bool {
	sys.MoveToNowhere(hWnd)

	return sys.GetColor(hWnd, BATTLE_COMMAND_PET_SKILL_ESCAPE.x, BATTLE_COMMAND_PET_SKILL_ESCAPE.y) == COLOR_BATTLE_COMMAND_DISABLE &&
		(sys.GetColor(hWnd, BATTLE_COMMAND_PET_SKILL_RIDING.x, BATTLE_COMMAND_PET_SKILL_RIDING.y) == COLOR_BATTLE_COMMAND_DISABLE ||
			sys.GetColor(hWnd, BATTLE_COMMAND_PET_SKILL_RIDING.x, BATTLE_COMMAND_PET_SKILL_RIDING.y) == COLOR_BATTLE_COMMAND_ENABLE)
}

var teleportWords = []string{"被不可思", "你感覺到一股"}

func isTeleported(dir string) bool {
	if dir != "" {
		for _, stopWord := range teleportWords {
			if strings.Contains(sys.GetLastLineOfLog(dir), stopWord) {
				return true
			}
		}
	}
	return false
}

func isOutOfResource(dir string) bool {
	if dir != "" {
		if strings.Contains(sys.GetLastLineOfLog(dir), "道具已經用完了") {
			return true
		}
	}
	return false
}

func doesEncounterAnyBaby(dir string) bool {
	lines := sys.GetLinesOfLog(dir, 5)
	now := time.Now()
	for i := range lines {
		h, hErr := strconv.Atoi(lines[i][1:3])
		m, mErrr := strconv.Atoi(lines[i][4:6])
		s, sErr := strconv.Atoi(lines[i][7:9])
		if hErr != nil || mErrr != nil || sErr != nil {
			continue
		}

		logTime := time.Date(now.Year(), now.Month(), now.Day(), h, m, s, 0, time.Local)
		if !logTime.Before(now.Add(-1*time.Minute)) && strings.Contains(lines[i], "發現野生一級") {
			return true
		}
	}
	return false
}

func canRecall(hWnd HWND) bool {
	sys.MoveToNowhere(hWnd)

	return sys.GetColor(hWnd, BATTLE_WINDOW_PET_RECALL_BUTTON.x, BATTLE_WINDOW_PET_RECALL_BUTTON.y) == COLOR_BATTLE_RECALL_BUTTON
}

var allTargets = []CheckTarget{
	PLAYER_L_1_P,
	PLAYER_L_2_P,
	PLAYER_L_3_P,
	PLAYER_L_4_P,
	PLAYER_L_5_P,
	PLAYER_L_1_H,
	PLAYER_L_2_H,
	PLAYER_L_3_H,
	PLAYER_L_4_H,
	PLAYER_L_5_H,
}

func isAnyPlayerOutOfMana(hWnd HWND) bool {
	for _, h := range allHumans {
		oy := h.y + 3
		manaPoint := h.x + 2
		if sys.GetColor(hWnd, manaPoint, oy) != COLOR_BATTLE_MANA_UPPER &&
			sys.GetColor(hWnd, h.x, h.y) == COLOR_BATTLE_BLOOD_UPPER {
			return true
		}
	}
	return false
}

func isLifeBelow(hWnd HWND, ratio float32, checkTarget *CheckTarget) bool {
	healthPoint := int32(ratio*30) + checkTarget.x
	return sys.GetColor(hWnd, healthPoint, checkTarget.y) != COLOR_BATTLE_BLOOD_UPPER &&
		sys.GetColor(hWnd, checkTarget.x, checkTarget.y) == COLOR_BATTLE_BLOOD_UPPER
}

func searchLifeBelow(hWnd HWND, ratio float32) (*CheckTarget, bool) {
	sys.MoveToNowhere(hWnd)

	for i := range allTargets {
		if isLifeBelow(hWnd, ratio, &allTargets[i]) {
			return &allTargets[i], true
		}
	}
	return nil, false
}

func countLifeBelow(hWnd HWND, ratio float32) (count int) {
	sys.MoveToNowhere(hWnd)

	for i := range allTargets {
		if isLifeBelow(hWnd, ratio, &allTargets[i]) {
			count++
		}
	}
	return
}

func searchTShapeLifeBelow(hWnd HWND, ratio float32) (*CheckTarget, bool) {
	detectedTargets := make([][]int, 2)
	for i := range detectedTargets {
		detectedTargets[i] = make([]int, 5)
	}
	counter := 0
	for i := range allTargets {
		if isLifeBelow(hWnd, ratio, &allTargets[i]) {
			counter++
			weight := 1 + i/5
			detectedTargets[i/5][i%5] += weight
			detectedTargets[(i/5)^1][i%5] += weight
			if i%5 > 0 {
				detectedTargets[i/5][i%5-1] += weight
			}
			if i%5 < 4 {
				detectedTargets[i/5][i%5+1] += weight
			}
		}
	}
	if counter >= 2 {
		max := 0
		maxId := 0
		for i := range allTargets {
			if max <= detectedTargets[i/5][i%5] && doesPlayerTargetExist(hWnd, allTargets[i]) {
				max = detectedTargets[i/5][i%5]
				maxId = i
			}
		}

		if max >= 2+maxId/5 {
			return &allTargets[maxId], true
		}
	}
	return nil, false
}

var allHumans = []CheckTarget{
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
	sys.MoveToNowhere(hWnd)

	targets := allHumans
	if !isHuman {
		targets = allPets
	}

	for i := range targets {
		if !doesPlayerTargetExist(hWnd, targets[i]) {
			continue
		}

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

func doesPlayerTargetExist(hWnd HWND, target CheckTarget) bool {
	return sys.GetColor(hWnd, target.x, target.y) == COLOR_BATTLE_BLOOD_UPPER
}

var allMonsterTargets = []CheckTarget{
	MON_POS_T_1,
	MON_POS_T_2,
	MON_POS_T_3,
	MON_POS_T_4,
	MON_POS_T_5,
	MON_POS_B_1,
	MON_POS_B_2,
	MON_POS_B_3,
	MON_POS_B_4,
	MON_POS_B_5,
}

func getEnemyTargets(hWnd HWND, checkTargets []CheckTarget) []CheckTarget {
	targets := []CheckTarget{}

	for i := range checkTargets {
		sys.MoveMouseWithInterval(hWnd, checkTargets[i].x, checkTargets[i].y, DETECT_MON_TARGET_MOUSE_MOV_INTERVAL)
		if sys.GetColor(hWnd, MENU_T.x, MENU_T.y) == COLOR_MENU_HIDDEN {
			targets = append(targets, checkTargets[i])
		}
	}
	return targets
}

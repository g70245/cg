package game

import (
	. "cg/internal"
	"math/rand"
	"slices"

	"time"

	"github.com/g70245/win"
)

const (
	DURATION_MONSTER_DETECING_CURSOR_MOV = 80 * time.Millisecond
	DURATION_DID_MISS_SKILL_WAITING      = 80 * time.Millisecond
	DURATION_BATTLE_RESULT_DISAPPEARING  = 2 * time.Second
)

var (
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
)

func (b *BattleActionState) isBattleCommandEnable(checkTarget CheckTarget) bool {
	return GetColor(b.hWnd, checkTarget.x, checkTarget.y) == COLOR_BATTLE_COMMAND_ENABLE
}

func (b *BattleActionState) isHumanStageStable() bool {
	return GetColor(b.hWnd, BATTLE_STAGE_HUMAN.x, BATTLE_STAGE_HUMAN.y) == BATTLE_STAGE_HUMAN.color
}

func (b *BattleActionState) isPetStageStable() bool {
	return GetColor(b.hWnd, BATTLE_STAGE_PET.x, BATTLE_STAGE_PET.y) == BATTLE_STAGE_PET.color
}

func (b *BattleActionState) isPetSkillWindowOpened() bool {
	return GetColor(b.hWnd, BATTLE_COMMAND_ESCAPE.x, BATTLE_COMMAND_ESCAPE.y) == COLOR_BATTLE_COMMAND_ENABLE
}

func (b *BattleActionState) isHumanActionSuccessful() bool {
	MoveCursorToNowhere(b.hWnd)
	return GetColor(b.hWnd, BATTLE_STAGE_HUMAN.x, BATTLE_STAGE_HUMAN.y) != BATTLE_STAGE_HUMAN.color
}

func (b *BattleActionState) isPetActionSuccessful() bool {
	MoveCursorToNowhere(b.hWnd)
	return GetColor(b.hWnd, BATTLE_STAGE_PET.x, BATTLE_STAGE_PET.y) != BATTLE_STAGE_PET.color
}

func (b *BattleActionState) getSkillWindowPos() (int32, int32, bool) {
	MoveCursorToNowhere(b.hWnd)
	x := BATTLE_WINDOW_SKILL_FIRST.x
	for x <= 164 {
		y := BATTLE_WINDOW_SKILL_FIRST.y
		for y <= 232 {
			if GetColor(b.hWnd, x, y) == BATTLE_WINDOW_SKILL_FIRST.color {
				return x, y, true
			}
			y += 2
		}
		x += 2
	}
	return 0, 0, false
}

func (b *BattleActionState) getItemWindowPos() (int32, int32, bool) {
	MoveCursorToNowhere(b.hWnd)
	x := BATTLE_WINDOW_ITEM_MONEY_PIVOT.x
	for x <= BATTLE_WINDOW_ITEM_MONEY_PIVOT.x+50 {
		y := BATTLE_WINDOW_ITEM_MONEY_PIVOT.y
		for y <= BATTLE_WINDOW_ITEM_MONEY_PIVOT.y+50 {
			if GetColor(b.hWnd, x, y) == BATTLE_WINDOW_ITEM_MONEY_PIVOT.color {
				return x - 78, y + 20, true
			}
			y += 2
		}
		x += 2
	}
	return 0, 0, false
}

func (b *BattleActionState) isInventoryWindowStuck() bool {
	return GetColor(b.hWnd, BATTLE_COMMAND_ITEM.x, BATTLE_COMMAND_ITEM.y) == COLOR_BATTLE_COMMAND_ENABLE
}

func (b *BattleActionState) isItemWindowStillOpened(x, y int32) bool {
	return GetColor(b.hWnd, x+78, y-20) == BATTLE_WINDOW_ITEM_MONEY_PIVOT.color
}

func (b *BattleActionState) didHumanMissSkill(x, y int32) bool {
	MoveCursorToNowhere(b.hWnd)
	time.Sleep(DURATION_DID_MISS_SKILL_WAITING)
	return GetColor(b.hWnd, x, y+16*10) == COLOR_WINDOW_SKILL_BOTTOM_SPACE
}

func (b *BattleActionState) didPetMissSkill() bool {
	return GetColor(b.hWnd, BATTLE_COMMAND_PET_SKILL_ESCAPE.x, BATTLE_COMMAND_PET_SKILL_ESCAPE.y) == COLOR_BATTLE_COMMAND_ENABLE
}

func (b *BattleActionState) didOnRideMissSkill() bool {
	return GetColor(b.hWnd, BATTLE_COMMAND_PET_SKILL_RIDING.x, BATTLE_COMMAND_PET_SKILL_RIDING.y) == COLOR_BATTLE_COMMAND_ENABLE
}

func (b *BattleActionState) isOnRide() bool {
	MoveCursorToNowhere(b.hWnd)

	return GetColor(b.hWnd, BATTLE_COMMAND_PET_SKILL_ESCAPE.x, BATTLE_COMMAND_PET_SKILL_ESCAPE.y) == COLOR_BATTLE_COMMAND_DISABLE &&
		(GetColor(b.hWnd, BATTLE_COMMAND_PET_SKILL_RIDING.x, BATTLE_COMMAND_PET_SKILL_RIDING.y) == COLOR_BATTLE_COMMAND_DISABLE ||
			GetColor(b.hWnd, BATTLE_COMMAND_PET_SKILL_RIDING.x, BATTLE_COMMAND_PET_SKILL_RIDING.y) == COLOR_BATTLE_COMMAND_ENABLE)
}

func (b *BattleActionState) canRecall() bool {
	MoveCursorToNowhere(b.hWnd)

	return GetColor(b.hWnd, BATTLE_WINDOW_PET_RECALL_BUTTON.x, BATTLE_WINDOW_PET_RECALL_BUTTON.y) == COLOR_BATTLE_RECALL_BUTTON
}

func (p *ProductionWorker) getItemWindowPos() (int32, int32, bool) {
	MoveCursorToNowhere(p.hWnd)
	x := PRODUCTION_WINDOW_ITEM_PIVOT.x
	for x <= PRODUCTION_WINDOW_ITEM_PIVOT.x+54 {
		y := PRODUCTION_WINDOW_ITEM_PIVOT.y
		for y <= PRODUCTION_WINDOW_ITEM_PIVOT.y+34 {
			if GetColor(p.hWnd, x, y) == PRODUCTION_WINDOW_ITEM_PIVOT.color {
				return x - 4*50 - 30, y + 28, true
			}
			y += 1
		}
		x += 1
	}
	return 0, 0, false
}

var allPlayers = []CheckTarget{
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

func (b *BattleActionState) isAnyPlayerOutOfMana() bool {
	for _, h := range allHumans {
		oy := h.y + 3
		manaPoint := h.x + 2
		if GetColor(b.hWnd, manaPoint, oy) != COLOR_BATTLE_MANA_UPPER &&
			GetColor(b.hWnd, h.x, h.y) == COLOR_BATTLE_BLOOD_UPPER {
			return true
		}
	}
	return false
}

func (b *BattleActionState) isHealthLowerThan(ratio float32, checkTarget *CheckTarget) bool {
	healthPoint := int32(ratio*30) + checkTarget.x
	oy := checkTarget.y + 3
	return GetColor(b.hWnd, healthPoint, checkTarget.y) != COLOR_BATTLE_BLOOD_UPPER &&
		GetColor(b.hWnd, checkTarget.x, oy) == COLOR_BATTLE_MANA_UPPER
}

func (b *BattleActionState) searchHealthLowerThan(ratio float32) (*CheckTarget, bool) {
	MoveCursorToNowhere(b.hWnd)

	copiedAllTargets := slices.Clone(allPlayers)
	rand.New(rand.NewSource(time.Now().UnixNano()))
	rand.Shuffle(len(copiedAllTargets), func(i, j int) { copiedAllTargets[i], copiedAllTargets[j] = copiedAllTargets[j], copiedAllTargets[i] })

	for i := range copiedAllTargets {
		if b.isHealthLowerThan(ratio, &copiedAllTargets[i]) {
			return &copiedAllTargets[i], true
		}
	}
	return nil, false
}

func (b *BattleActionState) countHealthLowerThan(ratio float32) (count int) {
	MoveCursorToNowhere(b.hWnd)

	for i := range allPlayers {
		if b.isHealthLowerThan(ratio, &allPlayers[i]) {
			count++
		}
	}
	return
}

func (b *BattleActionState) searchTShapeHealthLowerThan(ratio float32) (*CheckTarget, bool) {
	detectedTargets := make([][]int, 2)
	for i := range detectedTargets {
		detectedTargets[i] = make([]int, 5)
	}
	counter := 0
	for i := range allPlayers {
		if b.isHealthLowerThan(ratio, &allPlayers[i]) {
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
		for i := range allPlayers {
			if max <= detectedTargets[i/5][i%5] && b.doesPlayerTargetExist(allPlayers[i]) {
				max = detectedTargets[i/5][i%5]
				maxId = i
			}
		}

		if max >= 2+maxId/5 {
			return &allPlayers[maxId], true
		}
	}
	return nil, false
}

func (b *BattleActionState) getSelfTarget(isFirst bool) (*CheckTarget, bool) {
	MoveCursorToNowhere(b.hWnd)

	var targetRows [][]CheckTarget
	if isFirst {
		targetRows = append(targetRows, allHumans, allPets)
	} else {
		targetRows = append(targetRows, allPets, allHumans)
	}

	nameColors := []win.COLORREF{COLOR_BATTLE_SELF_NAME_1, COLOR_BATTLE_SELF_NAME_2, COLOR_BATTLE_SELF_NAME_3, COLOR_BATTLE_SELF_NAME_4, COLOR_BATTLE_SELF_NAME_5}

	for i := range targetRows {
		for j := range targetRows[i] {
			if !b.doesPlayerTargetExist(targetRows[i][j]) {
				continue
			}
			x := targetRows[i][j].x + 8
			for x <= targetRows[i][j].x+30 {
				y := targetRows[i][j].y - 10
				for y >= targetRows[i][j].y-26 {
					if slices.Contains(nameColors, GetColor(b.hWnd, x, y)) {
						return &targetRows[i][j], true
					}
					y--
				}
				x++
			}
		}
	}
	return nil, false
}

func (b *BattleActionState) doesPlayerTargetExist(target CheckTarget) bool {
	return GetColor(b.hWnd, target.x, target.y) == COLOR_BATTLE_BLOOD_UPPER
}

var allMonsters = []CheckTarget{
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

func (b *BattleActionState) getEnemies(checkTargets []CheckTarget) []CheckTarget {
	targets := []CheckTarget{}

	for i := range checkTargets {
		MoveCursorWithDuration(b.hWnd, checkTargets[i].x, checkTargets[i].y, DURATION_MONSTER_DETECING_CURSOR_MOV)
		if GetColor(b.hWnd, MENU_CONTACT.x, MENU_CONTACT.y) == COLOR_MENU_HIDDEN {
			targets = append(targets, checkTargets[i])
		}
	}
	return targets
}

func (bw *BattleWorker) isInventoryFullForActivity() bool {
	defer closeAllWindows(bw.hWnd)
	defer time.Sleep(DURATION_INVENTORY_CHECKER_WAITING)

	time.Sleep(DURATION_BATTLE_RESULT_DISAPPEARING)
	closeAllWindows(bw.hWnd)
	LeftClick(bw.hWnd, GAME_WIDTH/2, GAME_HEIGHT/2)

	openInventory(bw.hWnd)

	if px, py, ok := getItemWindowPos(bw.hWnd); ok {
		return !areMoreThanTwoInventorySlotsFree(bw.hWnd, px, py)
	}
	return false
}

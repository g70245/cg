package battle

import (
	"cg/game"
	"cg/internal"
	"math/rand"
	"slices"

	"time"

	"github.com/g70245/win"
)

const (
	DURATION_MONSTER_DETECING_CURSOR_MOV = 80 * time.Millisecond
	DURATION_DID_MISS_SKILL_WAITING      = 80 * time.Millisecond
	DRUATION_IS_ON_RIDE                  = 80 * time.Millisecond
	DURATION_BATTLE_RESULT_DISAPPEARING  = 2 * time.Second
)

var (
	MON_POS_T_1 = game.CheckTarget{26, 238, game.COLOR_ANY}
	MON_POS_T_2 = game.CheckTarget{92, 204, game.COLOR_ANY}
	MON_POS_T_3 = game.CheckTarget{162, 171, game.COLOR_ANY}
	MON_POS_T_4 = game.CheckTarget{225, 133, game.COLOR_ANY}
	MON_POS_T_5 = game.CheckTarget{282, 94, game.COLOR_ANY}
	MON_POS_B_1 = game.CheckTarget{93, 289, game.COLOR_ANY}
	MON_POS_B_2 = game.CheckTarget{156, 254, game.COLOR_ANY}
	MON_POS_B_3 = game.CheckTarget{216, 218, game.COLOR_ANY}
	MON_POS_B_4 = game.CheckTarget{284, 187, game.COLOR_ANY}
	MON_POS_B_5 = game.CheckTarget{343, 148, game.COLOR_ANY}

	PLAYER_L_1_H = game.CheckTarget{329, 431, game.COLOR_ANY}
	PLAYER_L_2_H = game.CheckTarget{394, 396, game.COLOR_ANY}
	PLAYER_L_3_H = game.CheckTarget{460, 361, game.COLOR_ANY}
	PLAYER_L_4_H = game.CheckTarget{524, 326, game.COLOR_ANY}
	PLAYER_L_5_H = game.CheckTarget{589, 291, game.COLOR_ANY}
	PLAYER_L_1_P = game.CheckTarget{269, 386, game.COLOR_ANY}
	PLAYER_L_2_P = game.CheckTarget{333, 350, game.COLOR_ANY}
	PLAYER_L_3_P = game.CheckTarget{397, 314, game.COLOR_ANY}
	PLAYER_L_4_P = game.CheckTarget{460, 277, game.COLOR_ANY}
	PLAYER_L_5_P = game.CheckTarget{524, 241, game.COLOR_ANY}

	BATTLE_COMMAND_ATTACK  = game.CheckTarget{386, 28, game.COLOR_ANY}
	BATTLE_COMMAND_DEFENCE = game.CheckTarget{386, 54, game.COLOR_ANY}
	BATTLE_COMMAND_SKILL   = game.CheckTarget{450, 28, game.COLOR_ANY}
	BATTLE_COMMAND_ITEM    = game.CheckTarget{450, 54, game.COLOR_ANY}
	BATTLE_COMMAND_PET     = game.CheckTarget{524, 28, game.COLOR_ANY}
	BATTLE_COMMAND_MOVE    = game.CheckTarget{524, 54, game.COLOR_ANY}
	BATTLE_COMMAND_ESCAPE  = game.CheckTarget{594, 54, game.COLOR_ANY}

	BATTLE_COMMAND_PET_SKILL_RIDING = game.CheckTarget{524, 54, game.COLOR_ANY}
	BATTLE_COMMAND_PET_SKILL_ESCAPE = game.CheckTarget{594, 54, game.COLOR_ANY}

	BATTLE_STAGE_HUMAN = game.CheckTarget{594, 28, COLOR_BATTLE_STAGE_HUMAN}
	BATTLE_STAGE_PET   = game.CheckTarget{594, 28, COLOR_BATTLE_STAGE_PET}

	BATTLE_WINDOW_SKILL_FIRST       = game.CheckTarget{154, 132, game.COLOR_WINDOW_SKILL_UNSELECTED}
	BATTLE_WINDOW_PET_RECALL_BUTTON = game.CheckTarget{384, 280, game.COLOR_ANY}

	BATTLE_WINDOW_ITEM_MONEY_PIVOT = game.CheckTarget{196, 114, COLOR_BS_INVENTORY_PIVOT}
)

func (b *BattleActionState) isBattleCommandEnable(checkTarget game.CheckTarget) bool {
	internal.MoveCursorToNowhere(b.hWnd)
	return internal.GetColor(b.hWnd, checkTarget.X, checkTarget.Y) == COLOR_BATTLE_COMMAND_ENABLE
}

func (b *BattleActionState) isHumanStageStable() bool {
	return internal.GetColor(b.hWnd, BATTLE_STAGE_HUMAN.X, BATTLE_STAGE_HUMAN.Y) == BATTLE_STAGE_HUMAN.Color
}

func (b *BattleActionState) isPetStageStable() bool {
	return internal.GetColor(b.hWnd, BATTLE_STAGE_PET.X, BATTLE_STAGE_PET.Y) == BATTLE_STAGE_PET.Color
}

func (b *BattleActionState) isPetSkillWindowOpened() bool {
	return internal.GetColor(b.hWnd, BATTLE_COMMAND_ESCAPE.X, BATTLE_COMMAND_ESCAPE.Y) == COLOR_BATTLE_COMMAND_ENABLE
}

func (b *BattleActionState) isHumanActionSuccessful() bool {
	internal.MoveCursorToNowhere(b.hWnd)
	return internal.GetColor(b.hWnd, BATTLE_STAGE_HUMAN.X, BATTLE_STAGE_HUMAN.Y) != BATTLE_STAGE_HUMAN.Color
}

func (b *BattleActionState) isPetActionSuccessful() bool {
	internal.MoveCursorToNowhere(b.hWnd)
	return internal.GetColor(b.hWnd, BATTLE_STAGE_PET.X, BATTLE_STAGE_PET.Y) != BATTLE_STAGE_PET.Color
}

func (b *BattleActionState) getSkillWindowPos() (int32, int32, bool) {
	internal.MoveCursorToNowhere(b.hWnd)
	x := BATTLE_WINDOW_SKILL_FIRST.X
	for x <= 164 {
		y := BATTLE_WINDOW_SKILL_FIRST.Y
		for y <= 232 {
			if internal.GetColor(b.hWnd, x, y) == BATTLE_WINDOW_SKILL_FIRST.Color {
				return x, y, true
			}
			y += 2
		}
		x += 2
	}
	return 0, 0, false
}

func (b *BattleActionState) getItemWindowPos() (int32, int32, bool) {
	internal.MoveCursorToNowhere(b.hWnd)
	x := BATTLE_WINDOW_ITEM_MONEY_PIVOT.X
	for x <= BATTLE_WINDOW_ITEM_MONEY_PIVOT.X+50 {
		y := BATTLE_WINDOW_ITEM_MONEY_PIVOT.Y
		for y <= BATTLE_WINDOW_ITEM_MONEY_PIVOT.Y+50 {
			if internal.GetColor(b.hWnd, x, y) == BATTLE_WINDOW_ITEM_MONEY_PIVOT.Color {
				return x - 78, y + 20, true
			}
			y += 2
		}
		x += 2
	}
	return 0, 0, false
}

func (b *BattleActionState) isInventoryWindowStuck() bool {
	return internal.GetColor(b.hWnd, BATTLE_COMMAND_ITEM.X, BATTLE_COMMAND_ITEM.Y) == COLOR_BATTLE_COMMAND_ENABLE
}

func (b *BattleActionState) isItemWindowStillOpened(x, y int32) bool {
	return internal.GetColor(b.hWnd, x+78, y-20) == BATTLE_WINDOW_ITEM_MONEY_PIVOT.Color
}

func (b *BattleActionState) didHumanMissSkill(x, y int32) bool {
	internal.MoveCursorToNowhere(b.hWnd)
	time.Sleep(DURATION_DID_MISS_SKILL_WAITING)
	return internal.GetColor(b.hWnd, x, y+16*10) == game.COLOR_WINDOW_SKILL_BOTTOM_SPACE
}

func (b *BattleActionState) didPetMissSkill() bool {
	return internal.GetColor(b.hWnd, BATTLE_COMMAND_PET_SKILL_ESCAPE.X, BATTLE_COMMAND_PET_SKILL_ESCAPE.Y) == COLOR_BATTLE_COMMAND_ENABLE
}

func (b *BattleActionState) didOnRideMissSkill() bool {
	return internal.GetColor(b.hWnd, BATTLE_COMMAND_PET_SKILL_RIDING.X, BATTLE_COMMAND_PET_SKILL_RIDING.Y) == COLOR_BATTLE_COMMAND_ENABLE
}

func (b *BattleActionState) isOnRide() bool {
	internal.MoveCursorToNowhere(b.hWnd)
	time.Sleep(DRUATION_IS_ON_RIDE)

	return internal.GetColor(b.hWnd, BATTLE_COMMAND_PET_SKILL_ESCAPE.X, BATTLE_COMMAND_PET_SKILL_ESCAPE.Y) == COLOR_BATTLE_COMMAND_DISABLE &&
		(internal.GetColor(b.hWnd, BATTLE_COMMAND_PET_SKILL_RIDING.X, BATTLE_COMMAND_PET_SKILL_RIDING.Y) == COLOR_BATTLE_COMMAND_DISABLE ||
			internal.GetColor(b.hWnd, BATTLE_COMMAND_PET_SKILL_RIDING.X, BATTLE_COMMAND_PET_SKILL_RIDING.Y) == COLOR_BATTLE_COMMAND_ENABLE)
}

func (b *BattleActionState) canRecall() bool {
	internal.MoveCursorToNowhere(b.hWnd)

	return internal.GetColor(b.hWnd, BATTLE_WINDOW_PET_RECALL_BUTTON.X, BATTLE_WINDOW_PET_RECALL_BUTTON.Y) == COLOR_BATTLE_RECALL_BUTTON
}

var allPlayers = []game.CheckTarget{
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

var allHumans = []game.CheckTarget{
	PLAYER_L_1_H,
	PLAYER_L_2_H,
	PLAYER_L_3_H,
	PLAYER_L_4_H,
	PLAYER_L_5_H,
}

var allPets = []game.CheckTarget{
	PLAYER_L_1_P,
	PLAYER_L_2_P,
	PLAYER_L_3_P,
	PLAYER_L_4_P,
	PLAYER_L_5_P,
}

func (b *BattleActionState) isAnyPlayerOutOfMana() bool {
	internal.MoveCursorToNowhere(b.hWnd)

	for _, h := range allHumans {
		oy := h.Y + 3
		manaPoint := h.X + 2
		if internal.GetColor(b.hWnd, manaPoint, oy) != COLOR_BATTLE_MANA_UPPER &&
			internal.GetColor(b.hWnd, h.X, h.Y) == COLOR_BATTLE_BLOOD_UPPER {
			return true
		}
	}
	return false
}

func (b *BattleActionState) isHealthLowerThan(ratio float32, checkTarget *game.CheckTarget) bool {

	healthPoint := int32(ratio*30) + checkTarget.X
	oy := checkTarget.Y + 3
	return internal.GetColor(b.hWnd, healthPoint, checkTarget.Y) != COLOR_BATTLE_BLOOD_UPPER &&
		internal.GetColor(b.hWnd, checkTarget.X, oy) == COLOR_BATTLE_MANA_UPPER
}

func (b *BattleActionState) searchHealthLowerThan(ratio float32) (*game.CheckTarget, bool) {
	internal.MoveCursorToNowhere(b.hWnd)

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
	internal.MoveCursorToNowhere(b.hWnd)

	for i := range allPlayers {
		if b.isHealthLowerThan(ratio, &allPlayers[i]) {
			count++
		}
	}
	return
}

func (b *BattleActionState) searchTShapeHealthLowerThan(ratio float32) (*game.CheckTarget, bool) {
	internal.MoveCursorToNowhere(b.hWnd)

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

func (b *BattleActionState) getSelfTarget(isFirst bool) (*game.CheckTarget, bool) {
	internal.MoveCursorToNowhere(b.hWnd)

	var targetRows [][]game.CheckTarget
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
			x := targetRows[i][j].X + 8
			for x <= targetRows[i][j].X+30 {
				y := targetRows[i][j].Y - 10
				for y >= targetRows[i][j].Y-26 {
					if slices.Contains(nameColors, internal.GetColor(b.hWnd, x, y)) {
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

func (b *BattleActionState) doesPlayerTargetExist(target game.CheckTarget) bool {
	return internal.GetColor(b.hWnd, target.X, target.Y) == COLOR_BATTLE_BLOOD_UPPER
}

var allMonsters = []game.CheckTarget{
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

func (b *BattleActionState) getEnemies(checkTargets []game.CheckTarget) []game.CheckTarget {
	targets := []game.CheckTarget{}

	for i := range checkTargets {
		internal.MoveCursorWithDuration(b.hWnd, checkTargets[i].X, checkTargets[i].Y, DURATION_MONSTER_DETECING_CURSOR_MOV)
		if internal.GetColor(b.hWnd, game.MENU_CONTACT.X, game.MENU_CONTACT.Y) == game.COLOR_MENU_HIDDEN {
			targets = append(targets, checkTargets[i])
		}
	}
	return targets
}

func isInventoryFullForActivity(hWnd win.HWND) bool {
	defer game.CloseAllWindows(hWnd)
	defer time.Sleep(game.DURATION_INVENTORY_CHECKER_WAITING)

	time.Sleep(DURATION_BATTLE_RESULT_DISAPPEARING)
	game.CloseAllWindows(hWnd)
	internal.LeftClick(hWnd, game.GAME_WIDTH/2, game.GAME_HEIGHT/2)

	game.OpenInventory(hWnd)

	if px, py, ok := game.GetItemWindowPos(hWnd); ok {
		return !game.AreMoreThanTwoInventorySlotsFree(hWnd, px, py)
	}
	return false
}

func isInventoryFull(hWnd win.HWND) bool {
	defer game.CloseAllWindows(hWnd)
	defer time.Sleep(game.DURATION_INVENTORY_CHECKER_WAITING)

	time.Sleep(DURATION_BATTLE_RESULT_DISAPPEARING)
	game.CloseAllWindows(hWnd)

	internal.LeftClick(hWnd, game.GAME_WIDTH/2, game.GAME_HEIGHT/2)

	game.OpenInventory(hWnd)

	if px, py, ok := game.GetItemWindowPos(hWnd); ok {
		return !game.IsAnyInventorySlotFree(hWnd, px, py)
	}
	return false
}

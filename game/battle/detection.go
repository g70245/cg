package battle

import (
	"cg/game"
	"cg/internal"
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
	BATTLE_COMMAND_ATTACK  = game.CheckTarget{X: 386, Y: 28, Color: game.COLOR_ANY}
	BATTLE_COMMAND_DEFENCE = game.CheckTarget{X: 386, Y: 54, Color: game.COLOR_ANY}
	BATTLE_COMMAND_SKILL   = game.CheckTarget{X: 450, Y: 28, Color: game.COLOR_ANY}
	BATTLE_COMMAND_ITEM    = game.CheckTarget{X: 450, Y: 54, Color: game.COLOR_ANY}
	BATTLE_COMMAND_PET     = game.CheckTarget{X: 524, Y: 28, Color: game.COLOR_ANY}
	BATTLE_COMMAND_MOVE    = game.CheckTarget{X: 524, Y: 54, Color: game.COLOR_ANY}
	BATTLE_COMMAND_ESCAPE  = game.CheckTarget{X: 594, Y: 54, Color: game.COLOR_ANY}

	BATTLE_COMMAND_PET_SKILL_RIDING = game.CheckTarget{X: 524, Y: 54, Color: game.COLOR_ANY}
	BATTLE_COMMAND_PET_SKILL_ESCAPE = game.CheckTarget{X: 594, Y: 54, Color: game.COLOR_ANY}

	BATTLE_STAGE_CHARACTER = game.CheckTarget{X: 594, Y: 28, Color: COLOR_BATTLE_STAGE_CHARACTER}
	BATTLE_STAGE_PET       = game.CheckTarget{X: 594, Y: 28, Color: COLOR_BATTLE_STAGE_PET}

	BATTLE_WINDOW_SKILL_FIRST       = game.CheckTarget{X: 154, Y: 132, Color: game.COLOR_WINDOW_SKILL_UNSELECTED}
	BATTLE_WINDOW_PET_RECALL_BUTTON = game.CheckTarget{X: 384, Y: 280, Color: game.COLOR_ANY}

	BATTLE_INVENTORY_MONEY_PIVOT = game.CheckTarget{X: 196, Y: 114, Color: COLOR_BS_INVENTORY_PIVOT}
)

func (s *ActionState) isBattleCommandEnable(checkTarget game.CheckTarget) bool {
	internal.MoveCursorToNowhere(s.hWnd)
	return internal.GetColor(s.hWnd, checkTarget.X, checkTarget.Y) == COLOR_BATTLE_COMMAND_ENABLE
}

func (s *ActionState) isCharacterStageStable() bool {
	return internal.GetColor(s.hWnd, BATTLE_STAGE_CHARACTER.X, BATTLE_STAGE_CHARACTER.Y) == BATTLE_STAGE_CHARACTER.Color
}

func (s *ActionState) isPetStageStable() bool {
	return internal.GetColor(s.hWnd, BATTLE_STAGE_PET.X, BATTLE_STAGE_PET.Y) == BATTLE_STAGE_PET.Color
}

func (s *ActionState) isPetSkillWindowOpened() bool {
	return internal.GetColor(s.hWnd, BATTLE_COMMAND_ESCAPE.X, BATTLE_COMMAND_ESCAPE.Y) == COLOR_BATTLE_COMMAND_ENABLE
}

func (s *ActionState) isCharacterActionSuccessful() bool {
	internal.MoveCursorToNowhere(s.hWnd)
	return internal.GetColor(s.hWnd, BATTLE_STAGE_CHARACTER.X, BATTLE_STAGE_CHARACTER.Y) != BATTLE_STAGE_CHARACTER.Color
}

func (s *ActionState) isPetActionSuccessful() bool {
	internal.MoveCursorToNowhere(s.hWnd)
	return internal.GetColor(s.hWnd, BATTLE_STAGE_PET.X, BATTLE_STAGE_PET.Y) != BATTLE_STAGE_PET.Color
}

func (s *ActionState) getSkillWindowPos() (int32, int32, bool) {
	internal.MoveCursorToNowhere(s.hWnd)
	x := BATTLE_WINDOW_SKILL_FIRST.X
	for x <= 164 {
		y := BATTLE_WINDOW_SKILL_FIRST.Y
		for y <= 232 {
			if internal.GetColor(s.hWnd, x, y) == BATTLE_WINDOW_SKILL_FIRST.Color {
				return x, y, true
			}
			y += 2
		}
		x += 2
	}
	return 0, 0, false
}

func (s *ActionState) getInventoryPos() (int32, int32, bool) {
	internal.MoveCursorToNowhere(s.hWnd)
	x := BATTLE_INVENTORY_MONEY_PIVOT.X
	for x <= BATTLE_INVENTORY_MONEY_PIVOT.X+50 {
		y := BATTLE_INVENTORY_MONEY_PIVOT.Y
		for y <= BATTLE_INVENTORY_MONEY_PIVOT.Y+50 {
			if internal.GetColor(s.hWnd, x, y) == BATTLE_INVENTORY_MONEY_PIVOT.Color {
				return x - 78, y + 20, true
			}
			y += 2
		}
		x += 2
	}
	return 0, 0, false
}

func (s *ActionState) isInventoryStuck() bool {
	return internal.GetColor(s.hWnd, BATTLE_COMMAND_ITEM.X, BATTLE_COMMAND_ITEM.Y) == COLOR_BATTLE_COMMAND_ENABLE
}

func (s *ActionState) isInventoryStillOpened(x, y int32) bool {
	return internal.GetColor(s.hWnd, x+78, y-20) == BATTLE_INVENTORY_MONEY_PIVOT.Color
}

func (s *ActionState) didCharacterMissSkill(x, y int32) bool {
	internal.MoveCursorToNowhere(s.hWnd)
	time.Sleep(DURATION_DID_MISS_SKILL_WAITING)
	return internal.GetColor(s.hWnd, x, y+16*10) == game.COLOR_WINDOW_SKILL_BOTTOM_SPACE
}

func (s *ActionState) didPetMissSkill() bool {
	return internal.GetColor(s.hWnd, BATTLE_COMMAND_PET_SKILL_ESCAPE.X, BATTLE_COMMAND_PET_SKILL_ESCAPE.Y) == COLOR_BATTLE_COMMAND_ENABLE
}

func (s *ActionState) didOnRideMissSkill() bool {
	return internal.GetColor(s.hWnd, BATTLE_COMMAND_PET_SKILL_RIDING.X, BATTLE_COMMAND_PET_SKILL_RIDING.Y) == COLOR_BATTLE_COMMAND_ENABLE
}

func (s *ActionState) isOnRide() bool {
	internal.MoveCursorToNowhere(s.hWnd)
	time.Sleep(DRUATION_IS_ON_RIDE)

	return internal.GetColor(s.hWnd, BATTLE_COMMAND_PET_SKILL_ESCAPE.X, BATTLE_COMMAND_PET_SKILL_ESCAPE.Y) == COLOR_BATTLE_COMMAND_DISABLE &&
		(internal.GetColor(s.hWnd, BATTLE_COMMAND_PET_SKILL_RIDING.X, BATTLE_COMMAND_PET_SKILL_RIDING.Y) == COLOR_BATTLE_COMMAND_DISABLE ||
			internal.GetColor(s.hWnd, BATTLE_COMMAND_PET_SKILL_RIDING.X, BATTLE_COMMAND_PET_SKILL_RIDING.Y) == COLOR_BATTLE_COMMAND_ENABLE)
}

func (s *ActionState) canRecall() bool {
	internal.MoveCursorToNowhere(s.hWnd)

	return internal.GetColor(s.hWnd, BATTLE_WINDOW_PET_RECALL_BUTTON.X, BATTLE_WINDOW_PET_RECALL_BUTTON.Y) == COLOR_BATTLE_RECALL_BUTTON
}

func (s *ActionState) isAnyCharacterOutOfMana() bool {
	internal.MoveCursorToNowhere(s.hWnd)

	for _, h := range allCharacters {
		oy := h.Y + 3
		manaPoint := h.X + 2
		if internal.GetColor(s.hWnd, h.X-1, oy) == COLOR_BATTLE_STATUS_PIVOT &&
			internal.GetColor(s.hWnd, manaPoint, oy) != COLOR_BATTLE_MANA_UPPER {
			return true
		}
	}
	return false
}

func (s *ActionState) isHealthLowerThan(ratio float32, checkTarget *game.CheckTarget) bool {

	healthPoint := int32(ratio*30) + checkTarget.X + 1
	oy := checkTarget.Y + 3
	return internal.GetColor(s.hWnd, checkTarget.X-1, oy) == COLOR_BATTLE_STATUS_PIVOT &&
		internal.GetColor(s.hWnd, healthPoint, checkTarget.Y) != COLOR_BATTLE_BLOOD_UPPER
}

func (s *ActionState) searchHealthLowerThan(ratio float32) (*game.CheckTarget, bool) {
	internal.MoveCursorToNowhere(s.hWnd)

	for i := range allPlayers {
		if s.isHealthLowerThan(ratio, &allPlayers[i]) {
			return &allPlayers[i], true
		}
	}
	return nil, false
}

func (s *ActionState) countHealthLowerThan(ratio float32) (count int) {
	internal.MoveCursorToNowhere(s.hWnd)

	for i := range allPlayers {
		if s.isHealthLowerThan(ratio, &allPlayers[i]) {
			count++
		}
	}
	return
}

func (s *ActionState) searchWeightedTShapedHealthLowerThan(ratio float32) (*game.CheckTarget, bool) {
	internal.MoveCursorToNowhere(s.hWnd)

	detectedTargets := make([][]int, 2)
	for i := range detectedTargets {
		detectedTargets[i] = make([]int, 5)
	}
	counter := 0
	for i := range allPlayers {
		if s.isHealthLowerThan(ratio, &allPlayers[i]) {
			counter++
			weight := 1 + (i/5 ^ 1)
			detectedTargets[i/5][i%5] += weight
			detectedTargets[i/5^1][i%5] += weight
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
		for i := len(allPlayers) - 1; i >= 0; i-- {
			if max <= detectedTargets[i/5][i%5] && s.doesPlayerExist(allPlayers[i]) {
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

func (s *ActionState) getSelfTarget(isFirst bool) (*game.CheckTarget, bool) {
	internal.MoveCursorToNowhere(s.hWnd)

	var targetRows [][]game.CheckTarget
	if isFirst {
		targetRows = append(targetRows, allCharacters, allPets)
	} else {
		targetRows = append(targetRows, allPets, allCharacters)
	}

	nameColors := []win.COLORREF{COLOR_BATTLE_SELF_NAME_1, COLOR_BATTLE_SELF_NAME_2, COLOR_BATTLE_SELF_NAME_3, COLOR_BATTLE_SELF_NAME_4, COLOR_BATTLE_SELF_NAME_5}

	for i := range targetRows {
		for j := range targetRows[i] {
			if !s.doesPlayerExist(targetRows[i][j]) {
				continue
			}
			x := targetRows[i][j].X + 8
			for x <= targetRows[i][j].X+30 {
				y := targetRows[i][j].Y - 10
				for y >= targetRows[i][j].Y-26 {
					if slices.Contains(nameColors, internal.GetColor(s.hWnd, x, y)) {
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

func (s *ActionState) doesPlayerExist(target game.CheckTarget) bool {
	return internal.GetColor(s.hWnd, target.X, target.Y) == COLOR_BATTLE_BLOOD_UPPER
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

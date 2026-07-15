package battle

import (
	"cg/game"
	"cg/game/enum/enemy"
	"cg/internal"
)

var (
	MON_POS_T_1 = game.CheckTarget{X: 26, Y: 238, Color: game.COLOR_ANY}
	MON_POS_T_2 = game.CheckTarget{X: 92, Y: 204, Color: game.COLOR_ANY}
	MON_POS_T_3 = game.CheckTarget{X: 162, Y: 171, Color: game.COLOR_ANY}
	MON_POS_T_4 = game.CheckTarget{X: 225, Y: 133, Color: game.COLOR_ANY}
	MON_POS_T_5 = game.CheckTarget{X: 282, Y: 94, Color: game.COLOR_ANY}
	MON_POS_B_1 = game.CheckTarget{X: 93, Y: 289, Color: game.COLOR_ANY}
	MON_POS_B_2 = game.CheckTarget{X: 156, Y: 254, Color: game.COLOR_ANY}
	MON_POS_B_3 = game.CheckTarget{X: 216, Y: 218, Color: game.COLOR_ANY}
	MON_POS_B_4 = game.CheckTarget{X: 284, Y: 187, Color: game.COLOR_ANY}
	MON_POS_B_5 = game.CheckTarget{X: 343, Y: 148, Color: game.COLOR_ANY}

	PLAYER_L_1_C = game.CheckTarget{X: 329, Y: 431, Color: game.COLOR_ANY}
	PLAYER_L_2_C = game.CheckTarget{X: 394, Y: 396, Color: game.COLOR_ANY}
	PLAYER_L_3_C = game.CheckTarget{X: 459, Y: 361, Color: game.COLOR_ANY}
	PLAYER_L_4_C = game.CheckTarget{X: 524, Y: 326, Color: game.COLOR_ANY}
	PLAYER_L_5_C = game.CheckTarget{X: 589, Y: 291, Color: game.COLOR_ANY}
	PLAYER_L_1_P = game.CheckTarget{X: 269, Y: 386, Color: game.COLOR_ANY}
	PLAYER_L_2_P = game.CheckTarget{X: 333, Y: 350, Color: game.COLOR_ANY}
	PLAYER_L_3_P = game.CheckTarget{X: 397, Y: 314, Color: game.COLOR_ANY}
	PLAYER_L_4_P = game.CheckTarget{X: 460, Y: 277, Color: game.COLOR_ANY}
	PLAYER_L_5_P = game.CheckTarget{X: 524, Y: 241, Color: game.COLOR_ANY}

	allPlayers    = []game.CheckTarget{PLAYER_L_1_C, PLAYER_L_2_C, PLAYER_L_3_C, PLAYER_L_4_C, PLAYER_L_5_C, PLAYER_L_1_P, PLAYER_L_2_P, PLAYER_L_3_P, PLAYER_L_4_P, PLAYER_L_5_P}
	allCharacters = []game.CheckTarget{PLAYER_L_1_C, PLAYER_L_2_C, PLAYER_L_3_C, PLAYER_L_4_C, PLAYER_L_5_C}
	allPets       = []game.CheckTarget{PLAYER_L_1_P, PLAYER_L_2_P, PLAYER_L_3_P, PLAYER_L_4_P, PLAYER_L_5_P}

	F4_W         = []game.CheckTarget{MON_POS_T_5, MON_POS_B_5, MON_POS_T_4, MON_POS_B_4, MON_POS_B_3, MON_POS_T_3, MON_POS_T_2, MON_POS_B_2, MON_POS_T_1, MON_POS_B_1}
	F4_P         = []game.CheckTarget{MON_POS_B_3, MON_POS_T_5, MON_POS_B_5, MON_POS_T_4, MON_POS_B_4, MON_POS_T_3, MON_POS_T_2, MON_POS_B_2, MON_POS_T_1, MON_POS_B_1}
	AllEnemies   = []game.CheckTarget{MON_POS_T_1, MON_POS_T_2, MON_POS_T_3, MON_POS_T_4, MON_POS_T_5, MON_POS_B_1, MON_POS_B_2, MON_POS_B_3, MON_POS_B_4, MON_POS_B_5}
	EnemyEnumMap = map[enemy.Position]game.CheckTarget{enemy.T1: MON_POS_T_1, enemy.T2: MON_POS_T_2, enemy.T3: MON_POS_T_3, enemy.T4: MON_POS_T_4, enemy.T5: MON_POS_T_5, enemy.B1: MON_POS_B_1, enemy.B2: MON_POS_B_2, enemy.B3: MON_POS_B_3, enemy.B4: MON_POS_B_4, enemy.B5: MON_POS_B_5}
)

func (s *ActionState) getEnemies(checkTargets []game.CheckTarget) []game.CheckTarget {
	targets := []game.CheckTarget{}

	for i := range checkTargets {
		internal.MoveCursorWithDuration(s.hWnd, checkTargets[i].X, checkTargets[i].Y, DURATION_MONSTER_DETECING_CURSOR_MOV)
		if internal.GetColor(s.hWnd, game.MENU_CONTACT.X, game.MENU_CONTACT.Y) == game.COLOR_MENU_HIDDEN {
			targets = append(targets, checkTargets[i])
		}
	}
	return targets
}

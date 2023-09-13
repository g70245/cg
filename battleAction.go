package main

import (
	"cg/system"
	"fmt"
	"time"

	"github.com/lxn/win"
)

var humanAttackOrder = []CheckTarget{MON_POS_B_3, MON_POS_T_3, MON_POS_B_2, MON_POS_B_4, MON_POS_T_2, MON_POS_T_4, MON_POS_B_1, MON_POS_B_5, MON_POS_T_1, MON_POS_T_5}
var petAttackOrder = []CheckTarget{MON_POS_T_3, MON_POS_B_3, MON_POS_T_2, MON_POS_T_4, MON_POS_B_2, MON_POS_B_4, MON_POS_T_1, MON_POS_T_5, MON_POS_B_1, MON_POS_B_5}

type HumanAttackMode string
type PetAttackMode string

const (
	NORMAL_ATTACK = "NORMAL_ATTACK"
	SKILL_ATTACK  = "SKILL_ATTACK"
	BOMB_ATTACK   = "BOMB_ATTACK"
)

type BattleState string

var battleStatesForSelector = []string{HUMAN_ATTACK_NOT_YET, CATCH_ATTACK, CATCH_ESCAPE}

const (
	CATCH_ESCAPE         = "CATCH_ESCAPE"
	CATCH_ATTACK         = "CATCH_ATTACK"
	POTION_HEAL          = "POTION_HEAL"
	MAGIC_HEAL           = "MAGIC_HEAL"
	HUMAN_ATTACK_NOT_YET = "HUMAN_ATTACK_NOT_YET"
	HUMAN_ATTACKED       = "HUMAN_ATTACKED"
	PET_ATTACKED         = "PET_ATTACKED"
)

type BattleActionState struct {
	hWnd            win.HWND
	state           BattleState
	humanMode       HumanAttackMode
	petMode         PetAttackMode
	humanSkillIndex int
	petSkillIndex   int
}

func (b *BattleActionState) Attack() {
	fmt.Printf("Handle %s's attack action begins\n", fmt.Sprint(b.hWnd))

	closeAll(b.hWnd)

	for GetScene(b.hWnd) == BATTLE_SCENE {
		switch b.state {
		case HUMAN_ATTACK_NOT_YET:
			if isHumanStageStable(b.hWnd) {
				switch b.humanMode {
				case NORMAL_ATTACK:
					if !IsBattleCommandEnable(b.hWnd, BATTLE_COMMAND_ATTACK) {
						system.LeftClick(b.hWnd, BATTLE_COMMAND_ATTACK.x, BATTLE_COMMAND_ATTACK.y)
						system.MoveToMiddle(b.hWnd)
						time.Sleep(200 * time.Millisecond)
					}
				case SKILL_ATTACK:
				}
				attack(b.hWnd, humanAttackOrder, DidHumanAttack)
				b.state = HUMAN_ATTACKED
				fmt.Printf("Handle %s human attacked\n", fmt.Sprint(b.hWnd))
				time.Sleep(100 * time.Millisecond)
			}
		case HUMAN_ATTACKED:
			if isPetStageStable(b.hWnd) {
				switch b.petMode {
				case NORMAL_ATTACK:
				case SKILL_ATTACK:
				}
				attack(b.hWnd, petAttackOrder, DidPetAttack)
				b.state = PET_ATTACKED
				fmt.Printf("Handle %s pet attacked\n", fmt.Sprint(b.hWnd))
				time.Sleep(1800 * time.Millisecond)
			}
		case PET_ATTACKED:
			if isHumanStageStable(b.hWnd) {
				b.state = HUMAN_ATTACK_NOT_YET
			} else {
				time.Sleep(400 * time.Millisecond)
			}
		}
	}

	fmt.Printf("Handle %s's attack action ended\n", fmt.Sprint(b.hWnd))
}

func attack(hWnd win.HWND, attackedTargets []CheckTarget, stageCheck func(hwnd win.HWND) bool) {
	for _, target := range attackedTargets {
		if stageCheck(hWnd) {
			break
		}
		system.LeftClick(hWnd, target.x, target.y)
		time.Sleep(100 * time.Millisecond)
	}
}

func closeAll(hwnd win.HWND) {
	system.KeyCombinationMsg(hwnd, win.VK_SHIFT, win.VK_F12)
}

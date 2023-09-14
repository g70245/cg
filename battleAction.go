package main

import (
	"cg/system"
	"fmt"
	"log"
	"time"

	"github.com/lxn/win"
)

const (
	ATTACK_INTERVAL = 180
	TURN_INTERVAL   = 2000
)

var humanBattleStatesForSelector = []string{H_A_ATTACK}
var petBattleStatesForSelector = []string{P_ATTACK}

type HumanState string
type PetState string

const (
	H_O_HEAL    = "Heal"
	H_O_Catch   = "Catch"
	H_O_Potion  = "Potion"
	H_A_ATTACK  = "Attack"
	H_A_Defence = "Defence"
	H_A_SKILL   = "Skill"
	H_A_BOMB    = "Bomb"
	H_A_ESCAPE  = "Escape"

	P_ATTACK = "Attack"
	P_Skill  = "Skill"

	R_ATTACK = "RAttack"
	R_SKILL  = "RSkill"
)

type BattleActionState struct {
	hWnd             win.HWND
	humanStates      []HumanState
	petStates        []PetState
	nextHumanStateID int
	nextPetStateID   int
	humanSkillID     map[int]int
	petSkillID       map[int]int
}

func (b *BattleActionState) Attack() {
	log.Printf("Handle %s's attack action begins\n", fmt.Sprint(b.hWnd))

	b.closeAll()

	for GetScene(b.hWnd) == BATTLE_SCENE {
		system.MoveOutOfFrame(b.hWnd)
		b.humanStateMachiine()
		system.MoveOutOfFrame(b.hWnd)
		b.petStateMachiine()
		time.Sleep(TURN_INTERVAL * time.Millisecond)
	}

	log.Printf("Handle %s's attack action ended\n", fmt.Sprint(b.hWnd))
}

func (b *BattleActionState) humanStateMachiine() {
	defer func() { b.nextHumanStateID = 0 }()

	for b.nextHumanStateID < len(b.humanStates) && GetScene(b.hWnd) == BATTLE_SCENE {
		if !isHumanStageStable(b.hWnd) {
			time.Sleep(200 * time.Millisecond)
			continue
		}

		switch b.humanStates[b.nextHumanStateID] {
		case H_A_ATTACK:
			b.enableBattleCommandAttack()
			if b.attack(humanAttackOrder, DidHumanAttack) {
				log.Printf("Handle %s human attacked\n", fmt.Sprint(b.hWnd))
				time.Sleep(100 * time.Millisecond)
				return
			}
		default:
			b.nextHumanStateID++
		}
	}
}

func (b *BattleActionState) petStateMachiine() {
	defer func() { b.nextPetStateID = 0 }()

	for b.nextPetStateID < len(b.petStates) {
		if !isPetStageStable(b.hWnd) && GetScene(b.hWnd) == BATTLE_SCENE {
			time.Sleep(200 * time.Millisecond)
			continue
		}

		switch b.petStates[b.nextPetStateID] {
		case P_ATTACK:
			if b.attack(petAttackOrder, DidPetAttack) {
				log.Printf("Handle %s pet attacked\n", fmt.Sprint(b.hWnd))
				time.Sleep(100 * time.Millisecond)
				return
			}
		default:
			b.nextPetStateID++
		}
	}
}

func (b *BattleActionState) enableBattleCommandAttack() {
	if !IsBattleCommandEnable(b.hWnd, BATTLE_COMMAND_ATTACK) {
		system.LeftClick(b.hWnd, BATTLE_COMMAND_ATTACK.x, BATTLE_COMMAND_ATTACK.y)
		time.Sleep(200 * time.Millisecond)
	}
}

var humanAttackOrder = []CheckTarget{MON_POS_B_3, MON_POS_T_3, MON_POS_B_2, MON_POS_B_4, MON_POS_T_2, MON_POS_T_4, MON_POS_B_1, MON_POS_B_5, MON_POS_T_1, MON_POS_T_5}
var petAttackOrder = []CheckTarget{MON_POS_T_5, MON_POS_T_1, MON_POS_B_5, MON_POS_B_1, MON_POS_T_4, MON_POS_T_2, MON_POS_B_4, MON_POS_B_2, MON_POS_T_3, MON_POS_B_3}

func (b *BattleActionState) attack(attackedTargets []CheckTarget, stageCheck func(hwnd win.HWND) bool) bool {
	for _, target := range attackedTargets {
		system.LeftClick(b.hWnd, target.x, target.y)
		time.Sleep(ATTACK_INTERVAL * time.Millisecond)
		if stageCheck(b.hWnd) {
			return true
		}
	}
	return false
}

func (b *BattleActionState) closeAll() {
	system.CloseAll(b.hWnd)
}

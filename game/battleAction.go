package game

import (
	sys "cg/system"
	"fmt"
	"log"
	"strconv"
	"time"

	. "github.com/lxn/win"
)

const (
	ATTACK_INTERVAL = 180
	TURN_INTERVAL   = 200
)

var humanBattleStatesForSelector = []string{H_A_ATTACK}
var petBattleStatesForSelector = []string{P_ATTACK}

const (
	H_A_ATTACK  = "**Attack"
	H_A_Defence = "**Defence"
	H_A_ESCAPE  = "**Escape"
	H_A_MOVE    = "**Move"
	H_A_HANG    = "**Hang"

	H_A_BOMB   = "*Bomb"
	H_O_PET    = "*Recall Pet"
	H_O_Catch  = "Catch"
	H_O_Potion = "*Potion"

	H_A_SKILL  = "*Skill"
	H_O_S_HEAL = "*Single Heal"
	H_A_STEAL  = "*Steal"
	H_O_RIDE   = "*Ride"

	P_ATTACK  = "Pet Attack"
	P_HANG    = "Pet Hang"
	P_SkILL   = "Pet Skill"
	P_DEFENCE = "Pet Defence"
	P_RIDE    = "Pet Ride"
)

type BattleActionState struct {
	hWnd             HWND
	HumanStates      []string
	PetStates        []string
	nextHumanStateId int
	nextPetStateId   int
	HumanSkillIds    map[int]string
	HumanSkillLevels map[int]string
	PetSkillIds      map[int]string

	Started bool
}

func (b *BattleActionState) Attack() {
	log.Printf("Handle %s's attack action begins\n", fmt.Sprint(b.hWnd))
	b.HandleStartedBySelf()
	defer b.HandleStartedBySelf()
	b.closeAll()

	for getScene(b.hWnd) == BATTLE_SCENE && b.Started {
		sys.MoveOutOfFrame(b.hWnd)
		b.humanStateMachiine()
		sys.MoveOutOfFrame(b.hWnd)
		b.petStateMachiine()
		time.Sleep(TURN_INTERVAL * time.Millisecond)
	}

	log.Printf("Handle %s's attack action ended\n", fmt.Sprint(b.hWnd))
}

func (b *BattleActionState) HandleStartedBySelf() {
	b.Started = !b.Started
}

func (b *BattleActionState) humanStateMachiine() {
	defer func() { b.nextHumanStateId = 0 }()

	for b.nextHumanStateId < len(b.HumanStates) && getScene(b.hWnd) == BATTLE_SCENE && !isPetStageStable(b.hWnd) {
		if !isHumanStageStable(b.hWnd) {
			if !b.Started {
				return
			}
			time.Sleep(200 * time.Millisecond)
			continue
		}

		switch b.HumanStates[b.nextHumanStateId] {
		case H_A_ATTACK:
			b.enableBattleCommandAttack()
			if b.attackTargets(humanAttackOrder, didHumanAttack) {
				log.Printf("Handle %s human attacked\n", fmt.Sprint(b.hWnd))
				return
			}
		case H_A_SKILL:
			openHumanWindow(b.hWnd, 0x57)
			if x, y, ok := getSkillWindowPos(b.hWnd); ok {
				id, _ := strconv.Atoi(b.HumanSkillIds[b.nextHumanStateId])
				level, _ := strconv.Atoi(b.HumanSkillLevels[b.nextHumanStateId])
				useHumanSkill(b.hWnd, x, y, id, level)
				if !b.handleHumanOutOfMana(x, y) {
					if b.attackTargets(humanAttackOrder, didHumanAttack) {
						log.Printf("Handle %s human used a skill\n", fmt.Sprint(b.hWnd))
						return
					}
				}
			}
		default:
		}

		b.nextHumanStateId++
	}
}

func (b *BattleActionState) petStateMachiine() {
	defer func() { b.nextPetStateId = 0 }()

	for b.nextPetStateId < len(b.PetStates) && !isHumanStageStable(b.hWnd) {
		if !isPetStageStable(b.hWnd) && getScene(b.hWnd) == BATTLE_SCENE {
			if !b.Started {
				return
			}
			time.Sleep(200 * time.Millisecond)
			continue
		}

		switch b.PetStates[b.nextPetStateId] {
		case P_ATTACK:
			b.openPetSkillWindow()
			if x, y, ok := getSkillWindowPos(b.hWnd); ok {
				usePetSkill(b.hWnd, x, y, 1)
				if b.attackTargets(petAttackOrder, didPetAttack) {
					log.Printf("Handle %s pet attacked\n", fmt.Sprint(b.hWnd))
					time.Sleep(100 * time.Millisecond)
					return
				}
			}

		case P_SkILL:
			b.openPetSkillWindow()
			if x, y, ok := getSkillWindowPos(b.hWnd); ok {
				id, _ := strconv.Atoi(b.PetSkillIds[b.nextPetStateId])
				usePetSkill(b.hWnd, x, y, id)
				if isOutOfMana(b.hWnd, x, y) {
					log.Printf("Handle %s Pet is out of mana\n", fmt.Sprint(b.hWnd))
				} else if b.attackTargets(petAttackOrder, didPetAttack) {
					log.Printf("Handle %s pet used a skill\n", fmt.Sprint(b.hWnd))
					time.Sleep(100 * time.Millisecond)
					return
				}
			}
		default:
		}

		b.nextPetStateId++
	}
}

func (b *BattleActionState) enableBattleCommandAttack() {
	if !isBattleCommandEnable(b.hWnd, BATTLE_COMMAND_ATTACK) {
		sys.LeftClick(b.hWnd, BATTLE_COMMAND_ATTACK.x, BATTLE_COMMAND_ATTACK.y)
		time.Sleep(160 * time.Millisecond)
	}
}

var humanAttackOrder = []CheckTarget{MON_POS_B_3, MON_POS_T_3, MON_POS_B_2, MON_POS_B_4, MON_POS_T_2, MON_POS_T_4, MON_POS_B_1, MON_POS_B_5, MON_POS_T_1, MON_POS_T_5}
var petAttackOrder = []CheckTarget{MON_POS_T_5, MON_POS_T_1, MON_POS_B_5, MON_POS_B_1, MON_POS_T_4, MON_POS_T_2, MON_POS_B_4, MON_POS_B_2, MON_POS_T_3, MON_POS_B_3}

func (b *BattleActionState) attackTargets(attackedTargets []CheckTarget, stageCheck func(hwnd HWND) bool) bool {
	for _, target := range attackedTargets {
		sys.LeftClick(b.hWnd, target.x, target.y)
		time.Sleep(ATTACK_INTERVAL * time.Millisecond)
		if stageCheck(b.hWnd) {
			return true
		}
	}
	return false
}

func (b *BattleActionState) closeAll() {
	closeAll(b.hWnd)
}

func (b *BattleActionState) handleHumanOutOfMana(x, y int32) bool {
	if isOutOfMana(b.hWnd, x, y) {
		log.Printf("Handle %s human is out of mana\n", fmt.Sprint(b.hWnd))
		sys.LeftClick(b.hWnd, MENU_W.x, MENU_W.y)
		b.closeAll()
		return true
	}
	return false
}

func (b *BattleActionState) openPetSkillWindow() {
	if !isPetSkillWindowOpend(b.hWnd) {
		sys.RightClick(b.hWnd, GAME_WIDTH/2, 28)
	}
	resetAllWindowPos(b.hWnd)
}

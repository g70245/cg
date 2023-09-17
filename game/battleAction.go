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
	TURN_INTERVAL         = 2000
	WAITING_LOOP_INTERVAL = 200
	ATTACK_INTERVAL       = 160

	BATTLE_ACTION_INTERVAL = 140
)

var humanBattleStatesForSelector = []string{H_A_ATTACK}
var petBattleStatesForSelector = []string{P_ATTACK}

const (
	H_A_ATTACK  = "**Attack"
	H_A_DEFENCE = "**Defence"
	H_A_ESCAPE  = "**Escape"
	H_A_MOVE    = "**Move"

	H_O_BOMB    = "*Bomb"
	H_O_PET     = "*Recall Pet"
	H_O_POTION  = "*Potion"
	H_O_SKILL   = "*Skill"
	H_O_SE_HEAL = "*Self Heal"
	H_O_S_HEAL  = "*Single Heal"
	H_O_M_HEAL  = "*Multi Heal"
	H_O_RIDE    = "*Ride"

	H_S_HANG  = "Hang"
	H_S_STEAL = "Steal"
	H_S_Catch = "Catch"
	H_S_TRAIN = "Train Skill"

	P_ATTACK  = "Pet Attack"
	P_HANG    = "Pet Hang"
	P_SkILL   = "Pet Skill"
	P_DEFENCE = "Pet Defence"
	P_HEAL    = "Pet Heal"
	P_RIDE    = "Pet Ride"
)

type BattleActionState struct {
	hWnd             HWND
	humanStates      []string
	petStates        []string
	nextHumanStateId int
	nextPetStateId   int
	humanSkillIds    []string
	humanSkillLevels []string
	petSkillIds      []string

	CanBattle bool
	isOnRide  bool
}

func CreateNewBattleActionState(hWnd HWND) BattleActionState {
	return BattleActionState{
		hWnd:             hWnd,
		humanStates:      []string{H_A_ATTACK},
		humanSkillIds:    []string{""},
		humanSkillLevels: []string{""},
		petStates:        []string{P_ATTACK},
		petSkillIds:      []string{""},
	}
}

func (b *BattleActionState) AddHumanState(newState string) {
	b.humanStates = append(b.humanStates, newState)
	b.humanSkillIds = append(b.humanSkillIds, "")
	b.humanSkillLevels = append(b.humanSkillLevels, "")
}

func (b *BattleActionState) AddPetState(newState string) {
	b.petStates = append(b.petStates, newState)
	b.petSkillIds = append(b.petSkillIds, "")
}

func (b *BattleActionState) AddHumanSkillId(newSkillId string) {
	b.humanSkillIds[len(b.humanSkillIds)-1] = newSkillId
}

func (b *BattleActionState) AddHumanLevel(newLevel string) {
	b.humanSkillLevels[len(b.humanSkillLevels)-1] = newLevel
}

func (b *BattleActionState) AddHPetSkillId(newSkillId string) {
	b.petSkillIds[len(b.petSkillIds)-1] = newSkillId
}

func (b *BattleActionState) ClearHumanStates() {
	b.humanStates = b.humanStates[:0]
	b.humanSkillIds = b.humanSkillIds[:0]
	b.humanSkillLevels = b.humanSkillLevels[:0]
}

func (b *BattleActionState) ClearPetStates() {
	b.petStates = b.petStates[:0]
	b.petSkillIds = b.petSkillIds[:0]
}

func (b *BattleActionState) GetHumanStates() []string {
	return b.humanStates
}

func (b *BattleActionState) GetHumanSkillIds() []string {
	return b.humanSkillIds
}

func (b *BattleActionState) GetHumanSkillLevels() []string {
	return b.humanSkillLevels
}

func (b *BattleActionState) GetPetStates() []string {
	return b.petStates
}

func (b *BattleActionState) GetPetSkillIds() []string {
	return b.petSkillIds
}

func (b *BattleActionState) Attack() {
	log.Printf("Handle %s's attack action begins\n", fmt.Sprint(b.hWnd))
	closeAllWindow(b.hWnd)

	for getScene(b.hWnd) == BATTLE_SCENE && b.CanBattle {
		b.humanStateMachiine()
		b.petStateMachiine()
		time.Sleep(TURN_INTERVAL * time.Millisecond)
	}

	b.nextHumanStateId = 0
	b.nextPetStateId = 0
	log.Printf("Handle %s's attack action ended\n", fmt.Sprint(b.hWnd))
}

func (b *BattleActionState) humanStateMachiine() {

	b.isOnRide = false

	for b.nextHumanStateId < len(b.humanStates) && getScene(b.hWnd) == BATTLE_SCENE && !isPetStageStable(b.hWnd) {
		if !isHumanStageStable(b.hWnd) {
			if !b.CanBattle {
				return
			}
			time.Sleep(WAITING_LOOP_INTERVAL * time.Millisecond)
			continue
		}

		switch b.humanStates[b.nextHumanStateId] {
		case H_A_ATTACK:
			b.enableBattleCommandAttack()
			if b.attackTargets(humanAttackOrder, didHumanAttack) {
				log.Printf("[%s] Handle %s attacked\n", H_A_ATTACK, fmt.Sprint(b.hWnd))
			}
		case H_O_SKILL:
			openHumanWindow(b.hWnd, 0x57)
			if x, y, ok := getSkillWindowPos(b.hWnd); ok {
				id, _ := strconv.Atoi(b.humanSkillIds[b.nextHumanStateId])
				level, _ := strconv.Atoi(b.humanSkillLevels[b.nextHumanStateId])
				useHumanSkill(b.hWnd, x, y, id, level)
				if isHumanOutOfMana(b.hWnd, x, y+16*10) {
					log.Printf("[%s] Handle %s is out of mana\n", H_O_SKILL, fmt.Sprint(b.hWnd))
				} else if b.attackTargets(humanAttackOrder, didHumanAttack) {
					log.Printf("[%s] Handle %s used a skill\n", H_O_SKILL, fmt.Sprint(b.hWnd))
				}
			}
		case H_A_DEFENCE:
			b.defend()
			log.Printf("[%s] Handle %s defended\n", H_A_DEFENCE, fmt.Sprint(b.hWnd))
		case H_A_MOVE:
			b.move()
			log.Printf("[%s] Handle %s moved\n", H_A_MOVE, fmt.Sprint(b.hWnd))
		case H_A_ESCAPE:
			b.escape()
			log.Printf("[%s] Handle %s escaped\n", H_A_ESCAPE, fmt.Sprint(b.hWnd))
		case H_S_HANG:
			log.Printf("[%s] Handle %s is idle\n", H_S_HANG, fmt.Sprint(b.hWnd))
			time.Sleep(30 * time.Second)
		case H_O_BOMB:
			openHumanWindow(b.hWnd, 0x45)
			if px, py, isPivotFound := getItemWindowPos(b.hWnd); isPivotFound {
				if x, y, ok := getItemPos(b.hWnd, px, py, COLOR_ITEM_BOMB_9A); ok {
					sys.DoubleClick(HWND(b.hWnd), x, y)
					if b.attackTargets(humanAttackOrder, didHumanAttack) {
						log.Printf("[%s] Handle %s throwed a bomb\n", H_O_BOMB, fmt.Sprint(b.hWnd))
					} else {
						log.Printf("[%s] Handle %s need bombs\n", H_O_BOMB, fmt.Sprint(b.hWnd))
					}
				}
			}
		case H_O_RIDE:
			if !b.isOnRide {
				openHumanWindow(b.hWnd, 0x57)
				if x, y, ok := getSkillWindowPos(b.hWnd); ok {
					id, _ := strconv.Atoi(b.humanSkillIds[b.nextHumanStateId])
					level, _ := strconv.Atoi(b.humanSkillLevels[b.nextHumanStateId])
					useHumanSkill(b.hWnd, x, y, id, level)
					if isHumanOutOfMana(b.hWnd, x, y+16*10) {
						log.Printf("[%s] Handle %s is out of mana\n", H_O_RIDE, fmt.Sprint(b.hWnd))
					} else {
						b.isOnRide = true
					}
				}
			}
		default:
		}

		b.nextHumanStateId++
	}

	if len(b.humanStates) == 0 {
		b.nextHumanStateId = 0
	} else {
		b.nextHumanStateId %= len(b.humanStates)
	}
}

func (b *BattleActionState) petStateMachiine() {
	for b.nextPetStateId < len(b.petStates) && !isHumanStageStable(b.hWnd) {
		if !isPetStageStable(b.hWnd) && getScene(b.hWnd) == BATTLE_SCENE {
			if !b.CanBattle {
				return
			}
			time.Sleep(WAITING_LOOP_INTERVAL * time.Millisecond)
			continue
		}

		switch b.petStates[b.nextPetStateId] {
		case P_ATTACK:
			b.openPetSkillWindow()
			if x, y, ok := getSkillWindowPos(b.hWnd); ok {
				usePetSkill(b.hWnd, x, y, 1)
				if b.attackTargets(petAttackOrder, didPetAttack) {
					log.Printf("[%s] Handle %s attacked\n", P_ATTACK, fmt.Sprint(b.hWnd))
				}
			}

		case P_SkILL:
			b.openPetSkillWindow()
			if x, y, ok := getSkillWindowPos(b.hWnd); ok {
				id, _ := strconv.Atoi(b.petSkillIds[b.nextPetStateId])
				usePetSkill(b.hWnd, x, y, id)
				if isPetOutOfMana(b.hWnd) {
					log.Printf("[%s] Handle %s is out of mana\n", fmt.Sprint(b.hWnd))
				} else if b.attackTargets(petAttackOrder, didPetAttack) {
					log.Printf("[%s] Handle %s used a skill\n", P_SkILL, fmt.Sprint(b.hWnd))
				}
			}
		case P_HANG:
			log.Printf("[%s] Handle %s pet is idle\n", P_HANG, fmt.Sprint(b.hWnd))
			time.Sleep(30 * time.Second)
		case P_DEFENCE:
			b.openPetSkillWindow()
			if x, y, ok := getSkillWindowPos(b.hWnd); ok {
				id, _ := strconv.Atoi(b.petSkillIds[b.nextPetStateId])
				usePetSkill(b.hWnd, x, y, id)
				if isPetOutOfMana(b.hWnd) {
					log.Printf("[%s] Handle %s is out of mana\n", P_DEFENCE, fmt.Sprint(b.hWnd))
				} else {
					log.Printf("[%s] Handle %s defended\n", H_A_DEFENCE, fmt.Sprint(b.hWnd))
				}
			}
		case P_HEAL:
			// TODO: check health status here
			b.openPetSkillWindow()
			if x, y, ok := getSkillWindowPos(b.hWnd); ok {
				id, _ := strconv.Atoi(b.petSkillIds[b.nextPetStateId])
				usePetSkill(b.hWnd, x, y, id)
				if isPetOutOfMana(b.hWnd) {
					log.Printf("[%s] Handle %s is out of mana\n", P_HEAL, fmt.Sprint(b.hWnd))
				}
			}
		default:
		}

		b.nextPetStateId++
	}

	if len(b.petStates) == 0 {
		b.nextPetStateId = 0
	} else {
		b.nextPetStateId %= len(b.petStates)
	}
}

func (b *BattleActionState) enableBattleCommandAttack() {
	if !isBattleCommandEnable(b.hWnd, BATTLE_COMMAND_ATTACK) {
		sys.LeftClick(b.hWnd, BATTLE_COMMAND_ATTACK.x, BATTLE_COMMAND_ATTACK.y)
		time.Sleep(BATTLE_ACTION_INTERVAL * time.Millisecond)
	}
}

func (b *BattleActionState) useItem(x, y int32) {
	sys.DoubleClick(b.hWnd, x, y)
	time.Sleep(BATTLE_ACTION_INTERVAL * time.Millisecond)
}

var humanAttackOrder = []CheckTarget{MON_POS_B_3, MON_POS_T_3, MON_POS_B_2, MON_POS_B_4, MON_POS_T_2, MON_POS_T_4, MON_POS_B_1, MON_POS_B_5, MON_POS_T_1, MON_POS_T_5}
var petAttackOrder = []CheckTarget{MON_POS_T_3, MON_POS_B_3, MON_POS_T_2, MON_POS_T_4, MON_POS_B_2, MON_POS_B_4, MON_POS_T_1, MON_POS_T_5, MON_POS_B_1, MON_POS_B_5}

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

func (b *BattleActionState) defend() {
	sys.LeftClick(b.hWnd, BATTLE_COMMAND_DEFENCE.x, BATTLE_COMMAND_DEFENCE.y)
	time.Sleep(BATTLE_ACTION_INTERVAL * time.Millisecond)
}

func (b *BattleActionState) escape() {
	sys.LeftClick(b.hWnd, BATTLE_COMMAND_ESCAPE.x, BATTLE_COMMAND_ESCAPE.y)
	time.Sleep(BATTLE_ACTION_INTERVAL * time.Millisecond)
}

func (b *BattleActionState) move() {
	sys.LeftClick(b.hWnd, BATTLE_COMMAND_MOVE.x, BATTLE_COMMAND_MOVE.y)
	time.Sleep(BATTLE_ACTION_INTERVAL * time.Millisecond)
}

func (b *BattleActionState) openPetSkillWindow() {
	if !isPetSkillWindowOpend(b.hWnd) {
		sys.RightClick(b.hWnd, GAME_WIDTH/2, 28)
	}
	resetAllWindowPos(b.hWnd)
}

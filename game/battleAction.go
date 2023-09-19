package game

import (
	sys "cg/system"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	. "github.com/lxn/win"
)

const (
	TURN_INTERVAL         = 300
	WAITING_LOOP_INTERVAL = 200
	ATTACK_INTERVAL       = 160

	BATTLE_ACTION_INTERVAL = 140
)

var humanBattleStatesForSelector = []string{H_A_ATTACK}
var petBattleStatesForSelector = []string{P_ATTACK}

const (
	H_A_ATTACK = "**Attack"
	H_A_DEFEND = "**Defend"
	H_A_ESCAPE = "**Escape"
	H_A_MOVE   = "**Move"

	H_O_BOMB       = "*Bomb"
	H_O_PET_RECALL = "*Recall Pet"
	H_O_POTION     = "*Potion"
	H_O_SKILL      = "*Skill"
	H_O_SE_HEAL    = "*Heal Self"
	H_O_O_HEAL     = "*Heal One"
	H_O_M_HEAL     = "*Heal Multi"
	H_O_RIDE       = "*Ride"

	H_S_HANG        = "Hang"
	H_S_STEAL       = "Steal"
	H_S_CATCH       = "Catch"
	H_S_TRAIN_SKILL = "Train Skill"

	P_ATTACK  = "Pet Attack"
	P_HANG    = "Pet Hang"
	P_SkILL   = "Pet Skill"
	P_DEFEND  = "Pet Defend"
	P_SE_HEAL = "Pet Heal Self"
	P_O_HEAL  = "Pet Heal One"
	P_RIDE    = "Pet Ride"
)

type BattleActionState struct {
	hWnd             HWND
	humanStates      []string
	petStates        []string
	nextHumanStateId int
	humanSkillIds    []string
	humanSkillLevels []string
	humanParams      []string
	nextPetStateId   int
	petSkillIds      []string
	petParams        []string

	Enabled         bool
	IsOutOfMana     bool
	IsRideTriggered bool
	isTrainingSkill bool
	isHumanHanging  bool
	isPetHanging    bool
}

func CreateNewBattleActionState(hWnd HWND) BattleActionState {
	return BattleActionState{
		hWnd:             hWnd,
		humanStates:      []string{H_A_ATTACK},
		humanSkillIds:    []string{""},
		humanSkillLevels: []string{""},
		humanParams:      []string{""},
		petStates:        []string{P_ATTACK},
		petSkillIds:      []string{""},
		petParams:        []string{""},
	}
}

func (b *BattleActionState) AddHumanState(newState string) {
	b.humanStates = append(b.humanStates, newState)
	b.humanSkillIds = append(b.humanSkillIds, "")
	b.humanSkillLevels = append(b.humanSkillLevels, "")
	b.humanParams = append(b.humanParams, "")
}

func (b *BattleActionState) AddPetState(newState string) {
	b.petStates = append(b.petStates, newState)
	b.petSkillIds = append(b.petSkillIds, "")
	b.petParams = append(b.petParams, "")
}

func (b *BattleActionState) AddHumanSkillId(newSkillId string) {
	b.humanSkillIds[len(b.humanSkillIds)-1] = newSkillId
}

func (b *BattleActionState) AddHumanSkillLevel(newLevel string) {
	b.humanSkillLevels[len(b.humanSkillLevels)-1] = newLevel
}

func (b *BattleActionState) AddHumanParams(param string) {
	b.humanParams[len(b.humanParams)-1] = param
}

func (b *BattleActionState) AddPetSkillId(newSkillId string) {
	b.petSkillIds[len(b.petSkillIds)-1] = newSkillId
}

func (b *BattleActionState) AddPetParams(param string) {
	b.petParams[len(b.petParams)-1] = param
}

func (b *BattleActionState) ClearHumanStates() {
	b.humanStates = b.humanStates[:0]
	b.humanSkillIds = b.humanSkillIds[:0]
	b.humanSkillLevels = b.humanSkillLevels[:0]
	b.humanParams = b.humanParams[:0]
}

func (b *BattleActionState) ClearPetStates() {
	b.petStates = b.petStates[:0]
	b.petSkillIds = b.petSkillIds[:0]
	b.petParams = b.petParams[:0]
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

func (b *BattleActionState) GetHumanParams() []string {
	return b.humanParams
}

func (b *BattleActionState) GetPetStates() []string {
	return b.petStates
}

func (b *BattleActionState) GetPetSkillIds() []string {
	return b.petSkillIds
}

func (b *BattleActionState) GetPetParams() []string {
	return b.petParams
}

func (b *BattleActionState) Act() {
	log.Printf("# Handle %s's battle begins\n", fmt.Sprint(b.hWnd))

	if b.isTrainingSkill && b.IsRideTriggered {
		b.IsRideTriggered = false
	}

	for getScene(b.hWnd) == BATTLE_SCENE && b.Enabled {
		b.executeHumanStateMachine()
		b.executePetStateMachiine()
		time.Sleep(TURN_INTERVAL * time.Millisecond)
	}

	b.nextHumanStateId = 0
	b.nextPetStateId = 0
	log.Printf("# Handle %s's battle ended\n", fmt.Sprint(b.hWnd))
}

func (b *BattleActionState) executeHumanStateMachine() {

	for b.nextHumanStateId < len(b.humanStates) && getScene(b.hWnd) == BATTLE_SCENE && !isPetStageStable(b.hWnd) {
		if !isHumanStageStable(b.hWnd) {
			if !b.Enabled {
				return
			}
			time.Sleep(WAITING_LOOP_INTERVAL * time.Millisecond)
			continue
		}

		// for re-ride
		if b.IsRideTriggered && !b.isTrainingSkill {
			b.IsRideTriggered = false
		}
		if b.isPetHanging {
			b.isPetHanging = false
			b.nextPetStateId++
		}

		switch b.humanStates[b.nextHumanStateId] {
		case H_A_ATTACK:
			b.enableBattleCommandAttack()
			if b.attack(humanAttackOrder, HumanTargetingChecker) {
				b.logH("attacked")
			} else {
				b.logH("missed a hit")
			}
		case H_O_SKILL:
			openWindowByShortcut(b.hWnd, 0x57)
			if x, y, ok := getSkillWindowPos(b.hWnd); ok {
				id, _ := strconv.Atoi(b.humanSkillIds[b.nextHumanStateId])
				level, _ := strconv.Atoi(b.humanSkillLevels[b.nextHumanStateId])
				useHumanSkill(b.hWnd, x, y, id, level)
				if isHumanOutOfMana(b.hWnd, x, y) {
					b.logH("is out of mana")
					b.IsOutOfMana = true
				} else if b.attack(humanAttackOrder, HumanTargetingChecker) {
					b.logH("used a skill")
				} else {
					b.logH("missed a hit")
				}
			} else {
				b.logH("cannot find the position of window")
			}
		case H_A_DEFEND:
			b.defend()
			b.logH("defended")
		case H_A_MOVE:
			b.move()
			b.logH("moved")
		case H_A_ESCAPE:
			b.escape()
			b.logH("escaped")
		case H_S_HANG:
			b.logH("is hanging")
			b.isHumanHanging = true
			return
		case H_O_BOMB:
			openWindowByShortcut(b.hWnd, 0x45)
			if px, py, isPivotFound := getItemWindowPos(b.hWnd); isPivotFound {
				if x, y, ok := getItemPos(b.hWnd, px, py, COLOR_ITEM_BOMB_9A); ok {
					sys.DoubleClick(HWND(b.hWnd), x, y)
					if b.attack(humanAttackOrder, HumanTargetingChecker) {
						b.logH("throwed a bomb")
					} else {
						b.logH("need bombs or missed a hit")
					}
				}
			} else {
				b.logH("cannot find the position of window")
			}
		case H_O_RIDE:
			openWindowByShortcut(b.hWnd, 0x57)
			if x, y, ok := getSkillWindowPos(b.hWnd); ok {
				id, _ := strconv.Atoi(b.humanSkillIds[b.nextHumanStateId])
				level, _ := strconv.Atoi(b.humanSkillLevels[b.nextHumanStateId])
				useHumanSkill(b.hWnd, x, y, id, level)
				if isHumanOutOfMana(b.hWnd, x, y) {
					b.logH("is out of mana")
					b.IsOutOfMana = true
				} else {
					b.logH("is tring to get on a pet")
					for i, v := range b.petStates {
						if v == P_RIDE {
							b.nextPetStateId = i
							break
						}
					}
				}
			} else {
				b.logH("cannot find the position of window")
			}
		case H_O_SE_HEAL:
			if self, ok := getSelfTarget(b.hWnd, true); ok {
				if isLifeBelow(b.hWnd, 0.6, self) {
					openWindowByShortcut(b.hWnd, 0x57)
					if x, y, ok := getSkillWindowPos(b.hWnd); ok {
						id, _ := strconv.Atoi(b.humanSkillIds[b.nextHumanStateId])
						level, _ := strconv.Atoi(b.humanSkillLevels[b.nextHumanStateId])
						useHumanSkill(b.hWnd, x, y, id, level)
						if isHumanOutOfMana(b.hWnd, x, y) {
							b.logH("is out of mana")
							b.IsOutOfMana = true
						} else {
							b.logH("healed self")
						}
					} else {
						b.logH("cannot find the position of window")
					}
				} else {
					b.logH("is a healthy boy")
				}
			} else {
				b.logH("cannot find self")
			}
		case H_O_O_HEAL:
			closeAllWindow(b.hWnd)
			clearChat(b.hWnd)
			if target, ok := searchOneLifeBelow(b.hWnd, 0.6); ok {
				openWindowByShortcut(b.hWnd, 0x57)
				if x, y, ok := getSkillWindowPos(b.hWnd); ok {
					id, _ := strconv.Atoi(b.humanSkillIds[b.nextHumanStateId])
					level, _ := strconv.Atoi(b.humanSkillLevels[b.nextHumanStateId])
					useHumanSkill(b.hWnd, x, y, id, level)
					if b.aim(target, HumanTargetingChecker) {
						b.logH("healed an ally")
					} else {
						b.logH("cannot target")
					}
				} else {
					b.logH("cannot find the position of window")
				}
			} else {
				b.logH("found no one needed to be taken care of")
			}
		case H_O_M_HEAL:
			closeAllWindow(b.hWnd)
			clearChat(b.hWnd)
			count := countLifeBelow(b.hWnd, 0.7)
			if count >= 6 {
				openWindowByShortcut(b.hWnd, 0x57)
				if x, y, ok := getSkillWindowPos(b.hWnd); ok {
					id, _ := strconv.Atoi(b.humanSkillIds[b.nextHumanStateId])
					level, _ := strconv.Atoi(b.humanSkillLevels[b.nextHumanStateId])
					useHumanSkill(b.hWnd, x, y, id, level)
					if b.aim(&PLAYER_L_3_H, HumanTargetingChecker) {
						b.logH("healed allies")
					} else {
						b.logH("cannot target")
					}
				} else {
					b.logH("cannot find the position of window")
				}
			} else {
				b.logH("found all good")
			}
		case H_O_PET_RECALL:
			openWindowByShortcut(b.hWnd, 0x52)
			if canRecall(b.hWnd) {
				b.recall()
				b.logH("recalled")
			} else {
				b.logH("already recalled")
			}
		case H_S_TRAIN_SKILL:
			openWindowByShortcut(b.hWnd, 0x57)
			if x, y, ok := getSkillWindowPos(b.hWnd); ok {
				id, _ := strconv.Atoi(b.humanSkillIds[b.nextHumanStateId])
				level, _ := strconv.Atoi(b.humanSkillLevels[b.nextHumanStateId])
				useHumanSkill(b.hWnd, x, y, id, level)
				if isHumanOutOfMana(b.hWnd, x, y) {
					b.logH("is out of mana")
					b.IsOutOfMana = true
				} else if b.aim(&PLAYER_L_3_P, HumanTargetingChecker) {
					b.logH("is training")
					b.nextHumanStateId = 0
					return
				}
			} else {
				b.logH("cannot find the position of window")
			}
		default:
		}

		b.nextHumanStateId++
	}

	if len(b.humanStates) != 0 {
		b.nextHumanStateId %= len(b.humanStates)
	} else {
		b.nextHumanStateId = 0
	}
}

func (b *BattleActionState) executePetStateMachiine() {

	for _, state := range b.humanStates {
		if state == H_S_TRAIN_SKILL {
			b.isTrainingSkill = true
		}
	}

	for b.nextPetStateId < len(b.petStates) && getScene(b.hWnd) == BATTLE_SCENE && !isHumanStageStable(b.hWnd) {
		if !isPetStageStable(b.hWnd) && getScene(b.hWnd) == BATTLE_SCENE {
			if !b.Enabled {
				return
			}
			time.Sleep(WAITING_LOOP_INTERVAL * time.Millisecond)
			continue
		}

		if b.isHumanHanging {
			b.isHumanHanging = false
			b.nextHumanStateId++
		}

		switch b.petStates[b.nextPetStateId] {
		case P_ATTACK:
			b.openPetSkillWindow()
			if x, y, ok := getSkillWindowPos(b.hWnd); ok {
				usePetSkill(b.hWnd, x, y, 1)
				if b.attack(petAttackOrder, PetTargetingChecker) {
					b.logP("attacked")
				} else {
					b.logP("missed a hit")
				}
			} else {
				b.logP("cannot find the position of window")
			}

		case P_SkILL:
			b.openPetSkillWindow()
			if x, y, ok := getSkillWindowPos(b.hWnd); ok {
				id, _ := strconv.Atoi(b.petSkillIds[b.nextPetStateId])
				usePetSkill(b.hWnd, x, y, id)
				if isPetOutOfMana(b.hWnd) || isRidingOutOfMana(b.hWnd) {
					b.logP("is out of mana")
					b.IsOutOfMana = true
				} else if b.attack(petAttackOrder, PetTargetingChecker) {
					b.logP("used a skill")
				} else {
					b.logP("missed a hit")
				}
			} else {
				b.logP("cannot find the position of window")
			}
		case P_HANG:
			b.logP("is hanging")
			b.isPetHanging = true
			return
		case P_DEFEND:
			b.openPetSkillWindow()
			if x, y, ok := getSkillWindowPos(b.hWnd); ok {
				id, _ := strconv.Atoi(b.petSkillIds[b.nextPetStateId])
				usePetSkill(b.hWnd, x, y, id)
				if isPetOutOfMana(b.hWnd) || isRidingOutOfMana(b.hWnd) {
					b.logP("is out of mana")
					b.IsOutOfMana = true
				} else {
					b.logP("defended")
				}
			} else {
				b.logP("cannot find the position of window")
			}
		case P_SE_HEAL:
			if self, ok := getSelfTarget(b.hWnd, true); ok {
				if isLifeBelow(b.hWnd, 0.6, self) {
					b.openPetSkillWindow()
					if x, y, ok := getSkillWindowPos(b.hWnd); ok {
						id, _ := strconv.Atoi(b.petSkillIds[b.nextPetStateId])
						usePetSkill(b.hWnd, x, y, id)
						if isPetOutOfMana(b.hWnd) || isRidingOutOfMana(b.hWnd) {
							b.logP("is out of mana")
							b.IsOutOfMana = true
						} else {
							b.logP("healed self")
						}
					} else {
						b.logP("cannot find the position of window")
					}
				} else {
					b.logP("is a healthy boy")
				}
			} else {
				b.logP("cannot find self")
			}
		case P_RIDE:
			if b.IsRideTriggered && b.isTrainingSkill {
				b.logP("is already off ride")
			} else if b.IsRideTriggered && !b.isTrainingSkill {
				b.logP("is already on ride")
			} else {
				b.openPetSkillWindow()
				if x, y, ok := getSkillWindowPos(b.hWnd); ok {
					id, _ := strconv.Atoi(b.petSkillIds[b.nextPetStateId])
					usePetSkill(b.hWnd, x, y, id)
					b.logP("succeeds on/off ride?")
					b.IsRideTriggered = true
				} else {
					b.logP("cannot find the position of window")
				}
			}
		case P_O_HEAL:
			closeAllWindow(b.hWnd)
			clearChat(b.hWnd)
			if target, ok := searchOneLifeBelow(b.hWnd, 0.6); ok {
				b.openPetSkillWindow()
				if x, y, ok := getSkillWindowPos(b.hWnd); ok {
					id, _ := strconv.Atoi(b.petSkillIds[b.nextPetStateId])
					usePetSkill(b.hWnd, x, y, id)
					if b.aim(target, PetTargetingChecker) {
						b.logP("healed an ally")
					} else {
						b.logP("cannot target")
					}
				} else {
					b.logP("cannot find the position of window")
				}

			} else {
				b.logP("found no one needed to be taken care of")
			}
		default:
		}

		b.nextPetStateId++
	}

	if len(b.petStates) != 0 {
		b.nextPetStateId %= len(b.petStates)
	} else {
		b.nextPetStateId = 0
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

func (b *BattleActionState) attack(attackedTargets []CheckTarget, stateChecker func(hwnd HWND) bool) bool {
	for _, target := range attackedTargets {
		sys.LeftClick(b.hWnd, target.x, target.y)
		time.Sleep(ATTACK_INTERVAL * time.Millisecond)
		if stateChecker(b.hWnd) {
			return true
		}
	}
	return false
}

func (b *BattleActionState) aim(target *CheckTarget, stateChecker func(hwnd HWND) bool) bool {
	sys.LeftClick(b.hWnd, target.x+15, target.y-22)
	time.Sleep(ATTACK_INTERVAL * time.Millisecond)
	return stateChecker(b.hWnd)
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

func (b *BattleActionState) recall() {
	sys.LeftClick(b.hWnd, BATTLE_WINDOW_PET_RECALL_BUTTON.x, BATTLE_WINDOW_PET_RECALL_BUTTON.y)
	time.Sleep(BATTLE_ACTION_INTERVAL * time.Millisecond)
}

func (b *BattleActionState) openPetSkillWindow() {
	closeAllWindow(b.hWnd)
	sys.RightClick(b.hWnd, GAME_WIDTH/2, 28)
	resetAllWindowsPosition(b.hWnd)
}

func (b *BattleActionState) logH(message string) {
	log.Printf("[%s]%-13s %s",
		fmt.Sprint(b.hWnd),
		fmt.Sprintf("[%s]", strings.Trim(b.humanStates[b.nextHumanStateId], "*")),
		fmt.Sprintf("Human %s", message),
	)
}

func (b *BattleActionState) logP(message string) {
	log.Printf("[%s]%-13s %s",
		fmt.Sprint(b.hWnd),
		fmt.Sprintf("[%s]", strings.Trim(b.petStates[b.nextPetStateId], "*")),
		fmt.Sprintf("Pet %s", message),
	)
}

package game

import (
	. "cg/internal"
	. "cg/utils"

	"fmt"
	"log"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"time"

	. "github.com/g70245/win"
)

const (
	DURATION_BATTLE_ACTION_LOOP_WAITING = 200 * time.Millisecond
	DURATION_BATTLE_ACTION_ATTACK       = 100 * time.Millisecond
	DURATION_BATTLE_ACTION_GENERAL      = 160 * time.Millisecond
)

type HumanState struct {
	Action             humanAction
	Offset             Offset
	Level              Offset
	Threshold          Threshold
	Param              string
	SuccessControlUnit ControlUnit
	FailureControlUnit ControlUnit
	SuccessJumpId      int
	FailureJumpId      int
}

type PetState struct {
	Action             petAction
	Offset             Offset
	Threshold          Threshold
	Param              string
	SuccessControlUnit ControlUnit
	FailureControlUnit ControlUnit
	SuccessJumpId      int
	FailureJumpId      int
}

type BattleActionState struct {
	hWnd HWND `json:"-"`

	HumanStates         []HumanState
	PetStates           []PetState
	currentHumanStateId int         `json:"-"`
	currentPetStateId   int         `json:"-"`
	currentControlUnit  ControlUnit `json:"-"`
	currentJumpId       int         `json:"-"`

	Enabled                bool `json:"-"`
	ActivityCheckerEnabled bool `json:"-"`
	isOutOfHealth          bool `json:"-"`
	isOutOfMana            bool `json:"-"`
	isHumanHanging         bool `json:"-"`
	isPetHanging           bool `json:"-"`

	ManaChecker *string `json:"-"`
	LogDir      *string `json:"-"`

	enemies              []CheckTarget `json:"-"`
	enemyDetectorCounter int           `json:"-"`
}

func (b *BattleActionState) Act() {
	log.Printf("# Handle %s's battle begins\n", fmt.Sprint(b.hWnd))

	for getScene(b.hWnd) == BATTLE_SCENE && b.Enabled {
		b.executeActivity()
		b.detectEnemies()
		b.checkHumanMana()
		b.executeHumanStateMachine()
		b.executePetStateMachiine()
		time.Sleep(DURATION_BATTLE_ACTION_LOOP_WAITING)
	}

	b.reset()
	log.Printf("@ Handle %s's battle ended\n", fmt.Sprint(b.hWnd))
}

func (b *BattleActionState) executeActivity() {
	if !b.ActivityCheckerEnabled {
		return
	}

	for getScene(b.hWnd) == BATTLE_SCENE && b.Enabled && (!b.isHumanStageStable() && !b.isPetStageStable()) {
		time.Sleep(DURATION_BATTLE_ACTION_LOOP_WAITING)
	}

	if doesEncounterActivityMonsters := doesEncounterActivityMonsters(*b.LogDir); doesEncounterActivityMonsters {
		b.logH("encounters the activity monster")
		Beeper.Play()

		for getScene(b.hWnd) == BATTLE_SCENE && b.Enabled {
			time.Sleep(DURATION_BATTLE_ACTION_LOOP_WAITING)
		}
	}
}

func (b *BattleActionState) executeHumanStateMachine() {

	for b.currentHumanStateId < len(b.HumanStates) && getScene(b.hWnd) == BATTLE_SCENE && b.isHumanStageStable() && b.Enabled {

		b.endPetHanging()

		switch b.HumanStates[b.currentHumanStateId].Action {
		case HumanAttack:
			b.enableBattleCommandAttack()
			if b.attack(b.isHumanActionSuccessful) {
				b.logH("attacked")
				b.setSuccessState(Human)
			} else {
				b.logH("missed a hit")
				b.setFailureState(Human)
			}
		case HumanSkill:
			openWindow(b.hWnd, KEY_SKILL)
			if x, y, ok := b.getSkillWindowPos(); ok {
				id := int(b.HumanStates[b.currentHumanStateId].Offset)
				level := int(b.HumanStates[b.currentHumanStateId].Level)
				useHumanSkill(b.hWnd, x, y, id, level)
				if b.didHumanMissSkill(x, y) {
					b.logH("missed the skill button or is out of mana")
				} else if b.isHumanActionSuccessful() {
					b.logH("used a skill")
					b.setSuccessState(Human)
				} else if b.attack(b.isHumanActionSuccessful) {
					b.logH("used a skill")
					b.setSuccessState(Human)
				} else {
					b.logH("missed a hit")
					b.setFailureState(Human)
				}
			} else {
				b.logH("cannot find the position of window")
				b.setFailureState(Human)
			}

		case HumanThresholdSkill:
			threshold, _ := strconv.Atoi(strings.Split(string(b.HumanStates[b.currentHumanStateId].Threshold), " ")[0])
			if len(b.enemies) < threshold {
				b.logH("performs next action due to too few enemies")
				break
			}

			openWindow(b.hWnd, KEY_SKILL)
			if x, y, ok := b.getSkillWindowPos(); ok {
				id := int(b.HumanStates[b.currentHumanStateId].Offset)
				level := int(b.HumanStates[b.currentHumanStateId].Level)
				useHumanSkill(b.hWnd, x, y, id, level)
				if b.didHumanMissSkill(x, y) {
					b.logH("missed the skill button or is out of mana")
				} else if b.isHumanActionSuccessful() {
					b.logH("used a skill")
					b.setSuccessState(Human)
				} else if b.attack(b.isHumanActionSuccessful) {
					b.logH("used a skill")
					b.setSuccessState(Human)
				} else {
					b.logH("missed a hit")
					b.setFailureState(Human)
				}
			} else {
				b.logH("cannot find the position of window")
				b.setFailureState(Human)
			}
		case HumanDefend:
			b.defend()
			b.logH("defended")
			b.setSuccessState(Human)
		case HumanMove:
			b.move()
			b.logH("moved")
			b.setSuccessState(Human)
		case HumanEscape:
			b.escape()
			b.logH("escaped")
			b.setFailureState(Human)
		case HumanHang:
			b.logH("is hanging")
			b.isHumanHanging = true
			b.currentControlUnit = Repeat
		case HumanBomb:
			var bomb Item
			for i := range Bombs {
				if Bombs[i].name == b.HumanStates[b.currentHumanStateId].Param {
					bomb = Bombs[i]
				}
			}

			if bomb == I_B_9A && len(b.enemies) < 5 {
				b.logH("performs next action due to too few enemies")
				break
			}

			openWindow(b.hWnd, KEY_INVENTORY)
			if px, py, isPivotFound := b.getItemWindowPos(); isPivotFound {
				if x, y, ok := getItemPos(b.hWnd, px, py, bomb.color, 2); ok {
					useItem(b.hWnd, x, y)
					if b.isItemWindowStillOpened(px, py) {
						b.logH("failed at double clicking")
						b.setFailureState(Human)
					} else if b.attack(b.isHumanActionSuccessful) {
						b.logH("throwed a bomb")
						b.setSuccessState(Human)
					} else {
						b.logH("missed a hit")
						b.setFailureState(Human)
					}

				} else {
					b.logH("cannot find a bomb")
					b.setFailureState(Human)
				}
			} else {
				b.logH("cannot find the position of window")
				b.setFailureState(Human)
			}
		case HumanPotion:
			closeAllWindows(b.hWnd)
			clearChat(b.hWnd)
			ratio, _ := strconv.ParseFloat(b.HumanStates[b.currentHumanStateId].Param, 32)
			if target, ok := b.searchHealthLowerThan(float32(ratio)); ok {
				openWindow(b.hWnd, KEY_INVENTORY)
				if px, py, isPivotFound := b.getItemWindowPos(); isPivotFound {
					if x, y, ok := getItemPos(b.hWnd, px, py, COLOR_ITEM_POTION, 3); ok {
						useItem(b.hWnd, x, y)
						if b.isItemWindowStillOpened(px, py) {
							b.logH("failed at double clicking")
							b.setFailureState(Human)
						} else if b.aim(target, b.isHumanActionSuccessful) {
							b.logH("healed an ally")
							b.setSuccessState(Human)
						} else {
							b.logH("cannot target")
							b.setFailureState(Human)
						}
					} else {
						b.logH("cannot find a potion")
						b.setFailureState(Human)
					}
				} else {
					b.logH("cannot find the position of window")
					b.setFailureState(Human)
				}
			} else {
				b.logH("found all good")
			}
		case HumanRide:
			openWindow(b.hWnd, KEY_SKILL)
			if x, y, ok := b.getSkillWindowPos(); ok {
				id := int(b.HumanStates[b.currentHumanStateId].Offset)
				level := int(b.HumanStates[b.currentHumanStateId].Level)
				useHumanSkill(b.hWnd, x, y, id, level)
				if b.didHumanMissSkill(x, y) {
					b.logH("missed the skill button or is out of mana")
				} else {
					b.logH("is tring to get on a pet")
					for i, v := range b.PetStates {
						if v.Action == PetRide {
							b.currentPetStateId = i
							break
						}
					}
					b.setSuccessState(Human)
				}
			} else {
				b.logH("cannot find the position of window")
				b.setFailureState(Human)
			}
		case HumanHealSelf:
			closeAllWindows(b.hWnd)
			clearChat(b.hWnd)
			if self, ok := b.getSelfTarget(true); ok {
				ratio, _ := strconv.ParseFloat(b.HumanStates[b.currentHumanStateId].Param, 32)

				if !b.isHealthLowerThan(float32(ratio), self) {
					b.logH("is a healthy boy")
					break
				}

				openWindow(b.hWnd, KEY_SKILL)
				if x, y, ok := b.getSkillWindowPos(); ok {
					id := int(b.HumanStates[b.currentHumanStateId].Offset)
					level := int(b.HumanStates[b.currentHumanStateId].Level)
					useHumanSkill(b.hWnd, x, y, id, level)
					if b.didHumanMissSkill(x, y) {
						b.logH("missed the skill button or is out of mana")
					} else {
						b.logH("healed self")
						b.setSuccessState(Human)
					}
				} else {
					b.logH("cannot find the position of window")
					b.setFailureState(Human)
				}

			} else {
				b.logH("cannot find self")
				b.setFailureState(Human)
			}
		case HumanHealOne:
			closeAllWindows(b.hWnd)
			clearChat(b.hWnd)
			ratio, _ := strconv.ParseFloat(b.HumanStates[b.currentHumanStateId].Param, 32)
			if target, ok := b.searchHealthLowerThan(float32(ratio)); ok {
				openWindow(b.hWnd, KEY_SKILL)
				if x, y, ok := b.getSkillWindowPos(); ok {
					id := int(b.HumanStates[b.currentHumanStateId].Offset)
					level := int(b.HumanStates[b.currentHumanStateId].Level)
					useHumanSkill(b.hWnd, x, y, id, level)
					if b.didHumanMissSkill(x, y) {
						b.logH("missed the skill button or is out of mana")
					} else if b.aim(target, b.isHumanActionSuccessful) {
						b.logH("healed an ally")
						b.setSuccessState(Human)
					} else {
						b.logH("cannot target")
						b.setFailureState(Human)
					}
				} else {
					b.logH("cannot find the position of window")
					b.setFailureState(Human)
				}
			} else {
				b.logH("found all good")
			}
		case HumanHealTShaped:
			closeAllWindows(b.hWnd)
			clearChat(b.hWnd)
			ratio, _ := strconv.ParseFloat(b.HumanStates[b.currentHumanStateId].Param, 32)
			if target, ok := b.searchTShapeHealthLowerThan(float32(ratio)); ok {
				openWindow(b.hWnd, KEY_SKILL)
				if x, y, ok := b.getSkillWindowPos(); ok {
					id := int(b.HumanStates[b.currentHumanStateId].Offset)
					level := int(b.HumanStates[b.currentHumanStateId].Level)
					useHumanSkill(b.hWnd, x, y, id, level)
					if b.didHumanMissSkill(x, y) {
						b.logH("missed the skill button or is out of mana")
					} else if b.aim(target, b.isHumanActionSuccessful) {
						b.logH("healed allies")
						b.setSuccessState(Human)
					} else {
						b.logH("cannot target")
						b.setFailureState(Human)
					}
				} else {
					b.logH("cannot find the position of window")
					b.setFailureState(Human)
				}
			} else {
				b.logH("found all good")
			}
		case HumanHealMulti:
			closeAllWindows(b.hWnd)
			clearChat(b.hWnd)
			ratio, _ := strconv.ParseFloat(b.HumanStates[b.currentHumanStateId].Param, 32)
			count := b.countHealthLowerThan(float32(ratio))
			if count >= 4 {
				openWindow(b.hWnd, KEY_SKILL)
				if x, y, ok := b.getSkillWindowPos(); ok {
					id := int(b.HumanStates[b.currentHumanStateId].Offset)
					level := int(b.HumanStates[b.currentHumanStateId].Level)
					useHumanSkill(b.hWnd, x, y, id, level)
					if b.didHumanMissSkill(x, y) {
						b.logH("missed the skill button or is out of mana")
					} else if b.aim(&PLAYER_L_3_H, b.isHumanActionSuccessful) {
						b.logH("healed allies")
						b.setSuccessState(Human)
					} else {
						b.logH("cannot target")
						b.setFailureState(Human)
					}
				} else {
					b.logH("cannot find the position of window")
					b.setFailureState(Human)
				}
			} else {
				b.logH("found all good")
			}
		case HumanRecall:
			openWindow(b.hWnd, KEY_PET)
			if b.canRecall() {
				b.recall()
				b.logH("recalled")
				b.currentControlUnit = Repeat
			} else {
				b.logH("already recalled")
				b.setSuccessState(Human)
			}
		case HumanTrainSkill:
			openWindow(b.hWnd, KEY_SKILL)
			if x, y, ok := b.getSkillWindowPos(); ok {
				id := int(b.HumanStates[b.currentHumanStateId].Offset)
				level := int(b.HumanStates[b.currentHumanStateId].Level)
				useHumanSkill(b.hWnd, x, y, id, level)
				if b.didHumanMissSkill(x, y) {
					b.logH("missed the skill button or is out of mana")
				} else if b.isHumanActionSuccessful() {
					b.logH("is training")
					b.setSuccessState(Human)
				} else if b.aim(&PLAYER_L_3_P, b.isHumanActionSuccessful) {
					b.logH("is training")
					b.setSuccessState(Human)
				}
			} else {
				b.logH("cannot find the position of window")
				b.setFailureState(Human)
			}
		case HumanCatch:
			closeAllWindows(b.hWnd)
			clearChat(b.hWnd)
			if self, ok := b.getSelfTarget(true); ok {
				ratio, _ := strconv.ParseFloat(b.HumanStates[b.currentHumanStateId].Param, 32)
				if b.isOutOfHealth = b.isHealthLowerThan(float32(ratio), self); b.isOutOfHealth {
					b.logH("is out of health")
				}
			}
		}

		b.updateCurrentStateId(Human)
	}

	if len(b.HumanStates) != 0 {
		b.currentHumanStateId %= len(b.HumanStates)
	} else {
		b.currentHumanStateId = 0
	}
}

func (b *BattleActionState) executePetStateMachiine() {

	for b.currentPetStateId < len(b.PetStates) && getScene(b.hWnd) == BATTLE_SCENE && b.isPetStageStable() && b.Enabled {

		b.endHumanHanging()

		switch b.PetStates[b.currentPetStateId].Action {
		case PetAttack:
			b.openPetSkillWindow()
			if x, y, ok := b.getSkillWindowPos(); ok {
				usePetSkill(b.hWnd, x, y, 1)
				if b.attack(b.isPetActionSuccessful) {
					b.logP("attacked")
					b.setSuccessState(Pet)
				} else {
					b.logP("missed a hit")
					b.setFailureState(Pet)
				}
			} else {
				b.logP("cannot find the position of window")
				b.setFailureState(Pet)
			}

		case PetSkill:
			b.openPetSkillWindow()
			if x, y, ok := b.getSkillWindowPos(); ok {
				id := int(b.PetStates[b.currentPetStateId].Offset)
				usePetSkill(b.hWnd, x, y, id)
				if b.didPetMissSkill() || b.didOnRideMissSkill() {
					b.logP("missed the skill button or is out of mana")
				} else if b.attack(b.isPetActionSuccessful) {
					b.logP("used a skill")
					b.setSuccessState(Pet)
				} else {
					b.logP("missed a hit")
					b.setFailureState(Pet)
				}
			} else {
				b.logP("cannot find the position of window")
				b.setFailureState(Pet)
			}
		case PetHang:
			b.logP("is hanging")
			b.isPetHanging = true
			b.currentControlUnit = Repeat
		case PetDefend:
			b.openPetSkillWindow()
			if x, y, ok := b.getSkillWindowPos(); ok {
				id := int(b.PetStates[b.currentPetStateId].Offset)
				usePetSkill(b.hWnd, x, y, id)
				if b.didPetMissSkill() || b.didOnRideMissSkill() {
					b.logP("missed the skill button or is out of mana")
				} else {
					b.logP("defended")
					b.setSuccessState(Pet)
				}
			} else {
				b.logP("cannot find the position of window")
				b.setFailureState(Pet)
			}
		case PetHealSelf:
			closeAllWindows(b.hWnd)
			clearChat(b.hWnd)
			if self, ok := b.getSelfTarget(false || b.isOnRide()); ok {
				ratio, _ := strconv.ParseFloat(b.PetStates[b.currentPetStateId].Param, 32)
				if !b.isHealthLowerThan(float32(ratio), self) {
					b.logP("is a healthy boy")
					break
				}

				b.openPetSkillWindow()
				if x, y, ok := b.getSkillWindowPos(); ok {
					id := int(b.PetStates[b.currentPetStateId].Offset)
					usePetSkill(b.hWnd, x, y, id)
					if b.didPetMissSkill() || b.didOnRideMissSkill() {
						b.logP("missed the skill button or is out of mana")
					} else {
						b.logP("healed self")
						b.setSuccessState(Pet)
					}
				} else {
					b.logP("cannot find the position of window")
					b.setFailureState(Pet)
				}
			} else {
				b.logP("cannot find self")
				b.setFailureState(Pet)
			}
		case PetRide:
			if b.isOnRide() {
				b.logP("is on ride")
				b.setSuccessState(Pet)
				break
			}

			b.openPetSkillWindow()
			if x, y, ok := b.getSkillWindowPos(); ok {
				id := int(b.PetStates[b.currentPetStateId].Offset)
				usePetSkill(b.hWnd, x, y, id)
				b.logP("tries to get on ride")
				b.currentControlUnit = Repeat
			} else {
				b.logP("cannot find the position of window")
				b.setFailureState(Pet)
			}
		case PetOffRide:
			if !b.isOnRide() {
				b.logP("is off ride")
				b.setSuccessState(Pet)
				break
			}

			b.openPetSkillWindow()
			if x, y, ok := b.getSkillWindowPos(); ok {
				id := int(b.PetStates[b.currentPetStateId].Offset)
				usePetSkill(b.hWnd, x, y, id)
				b.logP("tries to get off ride")
				b.currentControlUnit = Repeat
			} else {
				b.logP("cannot find the position of window")
				b.setFailureState(Pet)
			}
		case PetHealOne:
			closeAllWindows(b.hWnd)
			clearChat(b.hWnd)
			ratio, _ := strconv.ParseFloat(b.PetStates[b.currentPetStateId].Param, 32)
			if target, ok := b.searchHealthLowerThan(float32(ratio)); ok {
				b.openPetSkillWindow()
				if x, y, ok := b.getSkillWindowPos(); ok {
					id := int(b.PetStates[b.currentPetStateId].Offset)
					usePetSkill(b.hWnd, x, y, id)
					if b.aim(target, b.isPetActionSuccessful) {
						b.logP("healed an ally")
						b.setSuccessState(Pet)
					} else {
						b.logP("cannot target")
						b.setFailureState(Pet)
					}
				} else {
					b.logP("cannot find the position of window")
					b.setFailureState(Pet)
				}

			} else {
				b.logH("found all good")
			}
		case PetEscape:
			if !b.isOnRide() {
				b.logP("cannot escape while off ride")
				break
			}

			b.escape()
			b.logP("escaped")
			b.setFailureState(Pet)
		case PetCatch:
			closeAllWindows(b.hWnd)
			clearChat(b.hWnd)
			if self, ok := b.getSelfTarget(true); ok {
				ratio, _ := strconv.ParseFloat(b.PetStates[b.currentPetStateId].Param, 32)
				b.isOutOfHealth = b.isHealthLowerThan(float32(ratio), self)
				if b.isOutOfHealth {
					b.logP("is out of health")
				}
			}
		}

		b.updateCurrentStateId(Pet)
	}

	if len(b.PetStates) != 0 {
		b.currentPetStateId %= len(b.PetStates)
	} else {
		b.currentPetStateId = 0
	}
}

func (b *BattleActionState) reset() {
	b.currentHumanStateId = 0
	b.currentPetStateId = 0
	b.currentControlUnit = ""
	b.currentJumpId = 0
	b.enemyDetectorCounter = 0
}

func (b *BattleActionState) setSuccessState(r Role) {
	switch r {
	case Human:
		b.currentControlUnit = b.HumanStates[b.currentHumanStateId].SuccessControlUnit
		b.currentJumpId = b.HumanStates[b.currentHumanStateId].SuccessJumpId
	case Pet:
		b.currentControlUnit = b.PetStates[b.currentPetStateId].SuccessControlUnit
		b.currentJumpId = b.PetStates[b.currentPetStateId].SuccessJumpId
	}
}

func (b *BattleActionState) setFailureState(r Role) {
	switch r {
	case Human:
		b.currentControlUnit = b.HumanStates[b.currentHumanStateId].FailureControlUnit
		b.currentJumpId = b.HumanStates[b.currentHumanStateId].FailureJumpId
	case Pet:
		b.currentControlUnit = b.PetStates[b.currentPetStateId].FailureControlUnit
		b.currentJumpId = b.PetStates[b.currentPetStateId].FailureJumpId
	}
}

func (b *BattleActionState) updateCurrentStateId(r Role) {
	switch r {
	case Human:
		switch b.currentControlUnit {
		case StartOver:
			b.currentHumanStateId = 0
		case Continue:
			b.currentHumanStateId++
		case Repeat:
			return
		case Jump:
			b.currentHumanStateId = b.currentJumpId
		default:
			b.currentHumanStateId++
		}
	case Pet:
		switch b.currentControlUnit {
		case StartOver:
			b.currentPetStateId = 0
		case Continue:
			b.currentPetStateId++
		case Repeat:
			return
		case Jump:
			b.currentPetStateId = b.currentJumpId
		default:
			b.currentPetStateId++
		}
	}

	b.currentControlUnit = ""
	b.currentJumpId = 0
}

func (b BattleActionState) isManaChecker() bool {
	return *b.ManaChecker == fmt.Sprint(b.hWnd)
}

func (b *BattleActionState) enableBattleCommandAttack() {
	if !b.isBattleCommandEnable(BATTLE_COMMAND_ATTACK) {
		LeftClick(b.hWnd, BATTLE_COMMAND_ATTACK.x, BATTLE_COMMAND_ATTACK.y)
		time.Sleep(DURATION_BATTLE_ACTION_GENERAL)
	}
}

func (b *BattleActionState) attack(stateChecker func() bool) bool {
	targets := make([]CheckTarget, len(b.enemies))
	copy(targets, b.enemies)
	rand.New(rand.NewSource(time.Now().UnixNano()))
	rand.Shuffle(len(targets), func(i, j int) { targets[i], targets[j] = targets[j], targets[i] })
	for _, target := range targets {
		LeftClick(b.hWnd, target.x, target.y)
		time.Sleep(DURATION_BATTLE_ACTION_ATTACK)
		if stateChecker() {
			return true
		}
	}
	return false
}

func (b *BattleActionState) aim(target *CheckTarget, stateChecker func() bool) bool {
	LeftClick(b.hWnd, target.x+15, target.y-22)
	time.Sleep(DURATION_BATTLE_ACTION_ATTACK)
	return stateChecker()
}

func (b *BattleActionState) defend() {
	LeftClick(b.hWnd, BATTLE_COMMAND_DEFENCE.x, BATTLE_COMMAND_DEFENCE.y)
	time.Sleep(DURATION_BATTLE_ACTION_GENERAL)
}

func (b *BattleActionState) escape() {
	LeftClick(b.hWnd, BATTLE_COMMAND_ESCAPE.x, BATTLE_COMMAND_ESCAPE.y)
	time.Sleep(DURATION_BATTLE_ACTION_GENERAL)
}

func (b *BattleActionState) move() {
	LeftClick(b.hWnd, BATTLE_COMMAND_MOVE.x, BATTLE_COMMAND_MOVE.y)
	time.Sleep(DURATION_BATTLE_ACTION_GENERAL)
}

func (b *BattleActionState) recall() {
	LeftClick(b.hWnd, BATTLE_WINDOW_PET_RECALL_BUTTON.x, BATTLE_WINDOW_PET_RECALL_BUTTON.y)
	time.Sleep(DURATION_BATTLE_ACTION_GENERAL)
}

func (b *BattleActionState) openPetSkillWindow() {
	closeAllWindows(b.hWnd)
	RightClick(b.hWnd, GAME_WIDTH/2, 28)
	time.Sleep(DURATION_BATTLE_ACTION_GENERAL)
	resetAllWindows(b.hWnd)
}

func (b *BattleActionState) openSkillWindowWithMouse() {
	LeftClick(b.hWnd, BATTLE_COMMAND_SKILL.x, BATTLE_COMMAND_SKILL.y)
	time.Sleep(DURATION_BATTLE_ACTION_GENERAL)
}

func (b *BattleActionState) logH(message string) {
	header := fmt.Sprintf("[%s][%s]", fmt.Sprint(b.hWnd), strings.Trim(b.HumanStates[b.currentHumanStateId].Action.String(), "*"))
	log.Printf("%-26s %s",
		header,
		message,
	)
}

func (b *BattleActionState) logP(message string) {
	header := fmt.Sprintf("[%s][%s]", fmt.Sprint(b.hWnd), strings.Trim(b.PetStates[b.currentPetStateId].Action.String(), "*"))
	log.Printf("%-26s %s",
		header,
		message,
	)
}

func CreateNewBattleActionState(hWnd HWND, logDir, manaChecker *string) BattleActionState {
	return BattleActionState{
		hWnd: hWnd,
		HumanStates: []HumanState{
			{
				Action:             HumanAttack,
				SuccessControlUnit: Continue,
				FailureControlUnit: Continue,
			},
		},
		PetStates: []PetState{
			{
				Action:             PetAttack,
				SuccessControlUnit: Continue,
				FailureControlUnit: Continue,
			},
		},
		LogDir:      logDir,
		ManaChecker: manaChecker,
	}
}

func (b *BattleActionState) GetHWND() HWND {
	return b.hWnd
}

func (b *BattleActionState) SetHWND(hWnd HWND) {
	b.hWnd = hWnd
}

func (b *BattleActionState) AddHumanState(action humanAction) {
	newHumanState := HumanState{
		Action: action,
	}
	b.HumanStates = append(b.HumanStates, newHumanState)
}

func (b *BattleActionState) AddHumanSkillLevel(level Offset) {
	b.HumanStates[len(b.HumanStates)-1].Level = level
}

func (b *BattleActionState) AddPetState(action petAction) {
	newPetState := PetState{
		Action: action,
	}
	b.PetStates = append(b.PetStates, newPetState)
}

func (b *BattleActionState) AddSkillOffset(r Role, offset Offset) {
	switch r {
	case Human:
		b.HumanStates[len(b.HumanStates)-1].Offset = offset
	case Pet:
		b.PetStates[len(b.PetStates)-1].Offset = offset
	}
}

func (b *BattleActionState) AddThreshold(r Role, threshold Threshold) {
	b.HumanStates[len(b.HumanStates)-1].Threshold = threshold
}

func (b *BattleActionState) AddParam(r Role, param string) {
	switch r {
	case Human:
		b.HumanStates[len(b.HumanStates)-1].Param = param
	case Pet:
		b.PetStates[len(b.PetStates)-1].Param = param
	}
}

func (b *BattleActionState) AddSuccessControlUnit(r Role, cu ControlUnit) {
	switch r {
	case Human:
		b.HumanStates[len(b.HumanStates)-1].SuccessControlUnit = cu
	case Pet:
		b.PetStates[len(b.PetStates)-1].SuccessControlUnit = cu
	}
}
func (b *BattleActionState) AddSuccessJumpId(r Role, jumpId int) {
	switch r {
	case Human:
		b.HumanStates[len(b.HumanStates)-1].SuccessJumpId = jumpId
	case Pet:
		b.PetStates[len(b.PetStates)-1].SuccessJumpId = jumpId
	}
}

func (b *BattleActionState) AddFailureControlUnit(r Role, cu ControlUnit) {
	switch r {
	case Human:
		b.HumanStates[len(b.HumanStates)-1].FailureControlUnit = cu
	case Pet:
		b.PetStates[len(b.PetStates)-1].FailureControlUnit = cu
	}
}

func (b *BattleActionState) AddFailureJumpId(r Role, jumpId int) {
	switch r {
	case Human:
		b.HumanStates[len(b.HumanStates)-1].FailureJumpId = jumpId
	case Pet:
		b.PetStates[len(b.PetStates)-1].FailureJumpId = jumpId
	}
}

func (b *BattleActionState) ClearHumanStates() {
	b.HumanStates = b.HumanStates[:0]
}

func (b *BattleActionState) ClearPetStates() {
	b.PetStates = b.PetStates[:0]
}

func (b *BattleActionState) GetHumanStates() []HumanState {
	return b.HumanStates
}

func (b *BattleActionState) GetPetStates() []PetState {
	return b.PetStates
}

func (b *BattleActionState) detectEnemies() {

	for getScene(b.hWnd) == BATTLE_SCENE && b.Enabled && !b.isHumanStageStable() && !b.isPetStageStable() {
		time.Sleep(DURATION_BATTLE_ACTION_LOOP_WAITING)
	}

	// initial state: enemyDetectorCounter=0
	if getScene(b.hWnd) != BATTLE_SCENE || !b.Enabled || (b.enemyDetectorCounter != 0 && len(b.enemies) == 1) {
		return
	}

	if b.isInventoryWindowStuck() {
		b.openSkillWindowWithMouse()
	}
	closeAllWindows(b.hWnd)

	if b.enemyDetectorCounter == 0 || b.enemyDetectorCounter >= 3 {
		b.enemies = allMonsters
	}

	newEnemies := b.getEnemies(b.enemies)
	if reflect.DeepEqual(b.enemies, newEnemies) {
		b.enemyDetectorCounter++
	} else {
		b.enemyDetectorCounter = 1
	}
	b.enemies = newEnemies

	log.Printf("# Handle %s detected %d enemies\n", fmt.Sprint(b.hWnd), len(b.enemies))
}

func (b *BattleActionState) endHumanHanging() {
	if b.isHumanHanging {
		b.isHumanHanging = false
		cu := b.HumanStates[b.currentHumanStateId].SuccessControlUnit
		jumpId := b.HumanStates[b.currentHumanStateId].SuccessJumpId

		switch cu {
		case StartOver:
			b.currentHumanStateId = 0
		case Continue:
			b.currentHumanStateId++
		case Jump:
			b.currentHumanStateId = jumpId
		}
	}
}
func (b *BattleActionState) endPetHanging() {
	if b.isPetHanging {
		b.isPetHanging = false
		cu := b.PetStates[b.currentPetStateId].SuccessControlUnit
		jumpId := b.PetStates[b.currentPetStateId].SuccessJumpId

		switch cu {
		case StartOver:
			b.currentPetStateId = 0
		case Continue:
			b.currentPetStateId++
		case Jump:
			b.currentPetStateId = jumpId
		}
	}
}

func (b *BattleActionState) checkHumanMana() {
	if getScene(b.hWnd) != BATTLE_SCENE || !b.Enabled {
		return
	}

	if b.isManaChecker() {
		closeAllWindows(b.hWnd)
		clearChat(b.hWnd)
		if b.isOutOfMana = b.isAnyPlayerOutOfMana(); b.isOutOfMana {
			b.logH("someone is out of mana")
		}
	}
}

func TestAction(hWnd HWND) (x int32, y int32, successful bool) {
	closeAllWindows(hWnd)
	clearChat(hWnd)
	openWindow(hWnd, KEY_INVENTORY)
	return getItemPos(hWnd, x, y, 16448250, 3)
}

package game

import (
	"cg/internal"
	"cg/utils"

	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/g70245/win"
)

const (
	DURATION_BATTLE_ACTION_WAITING_LOOP = 100 * time.Millisecond
	DURATION_BATTLE_ACTION_ATTACK       = 100 * time.Millisecond
	DURATION_BATTLE_ACTION_GENERAL      = 160 * time.Millisecond

	TRAINING_COUNTER_THRESHOLD = 8
)

type HumanAction struct {
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

type PetAction struct {
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
	hWnd win.HWND `json:"-"`

	HumanActions         []HumanAction
	PetActions           []PetAction
	currentHumanActionId int         `json:"-"`
	currentPetActionId   int         `json:"-"`
	currentControlUnit   ControlUnit `json:"-"`
	currentJumpId        int         `json:"-"`

	Enabled                bool `json:"-"`
	ActivityCheckerEnabled bool `json:"-"`
	isOutOfHealth          bool `json:"-"`
	isOutOfMana            bool `json:"-"`
	isHumanHanging         bool `json:"-"`
	isPetHanging           bool `json:"-"`

	ManaChecker *string `json:"-"`
	GameDir     *string `json:"-"`

	enemies         []CheckTarget `json:"-"`
	trainingCounter int           `json:"-"`
}

func (b *BattleActionState) Act() {
	log.Printf("# Handle %s's battle begins\n", fmt.Sprint(b.hWnd))

	for getScene(b.hWnd) == BATTLE_SCENE && b.Enabled {
		b.wait()
		b.executeActivity()
		b.detectEnemies()
		b.checkHumanMana()
		b.executeHumanStateMachine()
		b.wait()
		b.executePetStateMachiine()
		b.wait()
		b.executePetStateMachiine()
	}

	b.reset()
	log.Printf("@ Handle %s's battle ended\n", fmt.Sprint(b.hWnd))
}

func (b *BattleActionState) executeActivity() {

	if !b.ActivityCheckerEnabled {
		return
	}

	if DoesEncounterActivityMonsters(*b.GameDir) {
		b.logH("encounters the activity monster")
		utils.Beeper.Play()

		for getScene(b.hWnd) == BATTLE_SCENE && b.Enabled {
			time.Sleep(DURATION_BATTLE_ACTION_WAITING_LOOP)
		}
	}
}

func (b *BattleActionState) executeHumanStateMachine() {

	for b.currentHumanActionId < len(b.HumanActions) && getScene(b.hWnd) == BATTLE_SCENE && b.isHumanStageStable() && b.Enabled {

		b.resetCurrentControlUnit()
		b.endPetHanging()

		switch b.HumanActions[b.currentHumanActionId].Action {
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
				offset := int(b.HumanActions[b.currentHumanActionId].Offset)
				level := int(b.HumanActions[b.currentHumanActionId].Level)
				useHumanSkill(b.hWnd, x, y, offset, level)
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
			threshold, _ := strconv.Atoi(strings.Split(string(b.HumanActions[b.currentHumanActionId].Threshold), " ")[0])
			if len(b.enemies) < threshold {
				b.logH("performs next action due to too few enemies")
				break
			}

			openWindow(b.hWnd, KEY_SKILL)
			if x, y, ok := b.getSkillWindowPos(); ok {
				offset := int(b.HumanActions[b.currentHumanActionId].Offset)
				level := int(b.HumanActions[b.currentHumanActionId].Level)
				useHumanSkill(b.hWnd, x, y, offset, level)
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
				if Bombs[i].name == b.HumanActions[b.currentHumanActionId].Param {
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
			ratio, _ := strconv.ParseFloat(b.HumanActions[b.currentHumanActionId].Param, 32)
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
				offset := int(b.HumanActions[b.currentHumanActionId].Offset)
				level := int(b.HumanActions[b.currentHumanActionId].Level)
				useHumanSkill(b.hWnd, x, y, offset, level)
				if b.didHumanMissSkill(x, y) {
					b.logH("missed the skill button or is out of mana")
				} else {
					b.logH("is tring to get on a pet")
					for i, v := range b.PetActions {
						if v.Action == PetRide {
							b.currentPetActionId = i
							break
						}
					}
					b.setSuccessState(Human)
				}
			} else {
				b.logH("cannot find the position of window")
				b.setFailureState(Human)
			}
		case HumanBloodMagic:
			closeAllWindows(b.hWnd)
			clearChat(b.hWnd)
			if self, ok := b.getSelfTarget(true); ok {
				ratio, _ := strconv.ParseFloat(b.HumanActions[b.currentHumanActionId].Param, 32)

				if !b.isHealthLowerThan(float32(ratio), self) {
					b.logH("is a healthy boy")
					break
				}

				openWindow(b.hWnd, KEY_SKILL)
				if x, y, ok := b.getSkillWindowPos(); ok {
					offset := int(b.HumanActions[b.currentHumanActionId].Offset)
					level := int(b.HumanActions[b.currentHumanActionId].Level)
					useHumanSkill(b.hWnd, x, y, offset, level)
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

			} else {
				b.logH("cannot find self")
			}
		case HumanHealSelf:
			closeAllWindows(b.hWnd)
			clearChat(b.hWnd)
			if self, ok := b.getSelfTarget(true); ok {
				ratio, _ := strconv.ParseFloat(b.HumanActions[b.currentHumanActionId].Param, 32)

				if !b.isHealthLowerThan(float32(ratio), self) {
					b.logH("is a healthy boy")
					break
				}

				openWindow(b.hWnd, KEY_SKILL)
				if x, y, ok := b.getSkillWindowPos(); ok {
					offset := int(b.HumanActions[b.currentHumanActionId].Offset)
					level := int(b.HumanActions[b.currentHumanActionId].Level)
					useHumanSkill(b.hWnd, x, y, offset, level)
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
			}
		case HumanHealOne:
			closeAllWindows(b.hWnd)
			clearChat(b.hWnd)
			ratio, _ := strconv.ParseFloat(b.HumanActions[b.currentHumanActionId].Param, 32)
			if target, ok := b.searchHealthLowerThan(float32(ratio)); ok {
				openWindow(b.hWnd, KEY_SKILL)
				if x, y, ok := b.getSkillWindowPos(); ok {
					offset := int(b.HumanActions[b.currentHumanActionId].Offset)
					level := int(b.HumanActions[b.currentHumanActionId].Level)
					useHumanSkill(b.hWnd, x, y, offset, level)
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
			ratio, _ := strconv.ParseFloat(b.HumanActions[b.currentHumanActionId].Param, 32)
			if target, ok := b.searchTShapeHealthLowerThan(float32(ratio)); ok {
				openWindow(b.hWnd, KEY_SKILL)
				if x, y, ok := b.getSkillWindowPos(); ok {
					offset := int(b.HumanActions[b.currentHumanActionId].Offset)
					level := int(b.HumanActions[b.currentHumanActionId].Level)
					useHumanSkill(b.hWnd, x, y, offset, level)
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
			ratio, _ := strconv.ParseFloat(b.HumanActions[b.currentHumanActionId].Param, 32)
			count := b.countHealthLowerThan(float32(ratio))
			if count >= 4 {
				openWindow(b.hWnd, KEY_SKILL)
				if x, y, ok := b.getSkillWindowPos(); ok {
					offset := int(b.HumanActions[b.currentHumanActionId].Offset)
					level := int(b.HumanActions[b.currentHumanActionId].Level)
					useHumanSkill(b.hWnd, x, y, offset, level)
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
			if self, ok := b.getSelfTarget(false); ok {
				openWindow(b.hWnd, KEY_SKILL)
				if x, y, ok := b.getSkillWindowPos(); ok {
					offset := int(b.HumanActions[b.currentHumanActionId].Offset)
					level := int(b.HumanActions[b.currentHumanActionId].Level)
					useHumanSkill(b.hWnd, x, y, offset, level)
					if b.didHumanMissSkill(x, y) {
						b.logH("missed the skill button or is out of mana")
					} else if b.isHumanActionSuccessful() {
						b.logH("is training")
						b.setSuccessState(Human)
					} else if b.aim(self, b.isHumanActionSuccessful) {
						b.logH("is training")
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
		case HumanCatch:
			closeAllWindows(b.hWnd)
			clearChat(b.hWnd)
			if self, ok := b.getSelfTarget(true); ok {
				ratio, _ := strconv.ParseFloat(b.HumanActions[b.currentHumanActionId].Param, 32)
				if b.isHealthLowerThan(float32(ratio), self) {
					b.isOutOfHealth = true
					b.logH("is out of health")
				}
			}
		}

		b.updateCurrentActionId(Human)
	}

	if len(b.HumanActions) != 0 {
		b.currentHumanActionId %= len(b.HumanActions)
	} else {
		b.currentHumanActionId = 0
	}
}

func (b *BattleActionState) executePetStateMachiine() {

	for b.currentPetActionId < len(b.PetActions) && getScene(b.hWnd) == BATTLE_SCENE && b.isPetStageStable() && b.Enabled {

		b.resetCurrentControlUnit()
		b.endHumanHanging()

		switch b.PetActions[b.currentPetActionId].Action {
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
		case PetEscape:
			if !b.isOnRide() {
				b.logP("cannot escape while off ride")
				break
			}

			b.escape()
			b.logP("escaped")
			b.setFailureState(Pet)
		case PetDefend:
			b.openPetSkillWindow()
			if x, y, ok := b.getSkillWindowPos(); ok {
				offset := int(b.PetActions[b.currentPetActionId].Offset)
				usePetSkill(b.hWnd, x, y, offset)
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
		case PetSkill:
			b.openPetSkillWindow()
			if x, y, ok := b.getSkillWindowPos(); ok {
				offset := int(b.PetActions[b.currentPetActionId].Offset)
				usePetSkill(b.hWnd, x, y, offset)
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
		case PetHealSelf:
			closeAllWindows(b.hWnd)
			clearChat(b.hWnd)
			if self, ok := b.getSelfTarget(b.isOnRide()); ok {
				ratio, _ := strconv.ParseFloat(b.PetActions[b.currentPetActionId].Param, 32)
				if !b.isHealthLowerThan(float32(ratio), self) {
					b.logP("is a healthy boy")
					break
				}

				b.openPetSkillWindow()
				if x, y, ok := b.getSkillWindowPos(); ok {
					offset := int(b.PetActions[b.currentPetActionId].Offset)
					usePetSkill(b.hWnd, x, y, offset)
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
			}
		case PetHealOne:
			closeAllWindows(b.hWnd)
			clearChat(b.hWnd)
			ratio, _ := strconv.ParseFloat(b.PetActions[b.currentPetActionId].Param, 32)
			if target, ok := b.searchHealthLowerThan(float32(ratio)); ok {
				b.openPetSkillWindow()
				if x, y, ok := b.getSkillWindowPos(); ok {
					offset := int(b.PetActions[b.currentPetActionId].Offset)
					usePetSkill(b.hWnd, x, y, offset)
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
		case PetRide:
			if b.isOnRide() {
				b.logP("is on ride")
				b.setSuccessState(Pet)
				break
			}

			b.openPetSkillWindow()
			if x, y, ok := b.getSkillWindowPos(); ok {
				offset := int(b.PetActions[b.currentPetActionId].Offset)
				usePetSkill(b.hWnd, x, y, offset)
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
				offset := int(b.PetActions[b.currentPetActionId].Offset)
				usePetSkill(b.hWnd, x, y, offset)
				b.logP("tries to get off ride")
				b.currentControlUnit = Repeat
			} else {
				b.logP("cannot find the position of window")
				b.setFailureState(Pet)
			}
		case PetCatch:
			closeAllWindows(b.hWnd)
			clearChat(b.hWnd)
			if self, ok := b.getSelfTarget(b.isOnRide()); ok {
				ratio, _ := strconv.ParseFloat(b.PetActions[b.currentPetActionId].Param, 32)
				if b.isHealthLowerThan(float32(ratio), self) {
					b.isOutOfHealth = true
					b.logP("is out of health")
				}
			}
		case PetHang:
			b.logP("is hanging")
			b.isPetHanging = true
			b.currentControlUnit = Repeat
		}

		b.updateCurrentActionId(Pet)
	}

	if len(b.PetActions) != 0 {
		b.currentPetActionId %= len(b.PetActions)
	} else {
		b.currentPetActionId = 0
	}
}

func (b *BattleActionState) reset() {
	b.currentHumanActionId = 0
	b.currentPetActionId = 0
	b.trainingCounter = 0
}

func (b *BattleActionState) resetCurrentControlUnit() {
	b.currentControlUnit = Undefined
	b.currentJumpId = 0
}

func (b *BattleActionState) setSuccessState(r Role) {
	switch r {
	case Human:
		b.currentControlUnit = b.HumanActions[b.currentHumanActionId].SuccessControlUnit
		b.currentJumpId = b.HumanActions[b.currentHumanActionId].SuccessJumpId
	case Pet:
		b.currentControlUnit = b.PetActions[b.currentPetActionId].SuccessControlUnit
		b.currentJumpId = b.PetActions[b.currentPetActionId].SuccessJumpId
	}
}

func (b *BattleActionState) setFailureState(r Role) {
	switch r {
	case Human:
		b.currentControlUnit = b.HumanActions[b.currentHumanActionId].FailureControlUnit
		b.currentJumpId = b.HumanActions[b.currentHumanActionId].FailureJumpId
	case Pet:
		b.currentControlUnit = b.PetActions[b.currentPetActionId].FailureControlUnit
		b.currentJumpId = b.PetActions[b.currentPetActionId].FailureJumpId
	}
}

func (b *BattleActionState) updateCurrentActionId(r Role) {
	switch r {
	case Human:
		switch b.currentControlUnit {
		case StartOver:
			b.currentHumanActionId = 0
		case Continue:
			b.currentHumanActionId++
		case Repeat:
			return
		case Jump:
			b.currentHumanActionId = b.currentJumpId
		default:
			b.currentHumanActionId++
		}
	case Pet:
		switch b.currentControlUnit {
		case StartOver:
			b.currentPetActionId = 0
		case Continue:
			b.currentPetActionId++
		case Repeat:
			return
		case Jump:
			b.currentPetActionId = b.currentJumpId
		default:
			b.currentPetActionId++
		}
	}

	b.currentControlUnit = Undefined
	b.currentJumpId = 0
}

func (b BattleActionState) isManaChecker() bool {
	return *b.ManaChecker == fmt.Sprint(b.hWnd)
}

func (b *BattleActionState) enableBattleCommandAttack() {
	if !b.isBattleCommandEnable(BATTLE_COMMAND_ATTACK) {
		internal.LeftClick(b.hWnd, BATTLE_COMMAND_ATTACK.x, BATTLE_COMMAND_ATTACK.y)
		time.Sleep(DURATION_BATTLE_ACTION_GENERAL)
	}
}

func (b *BattleActionState) attack(stateChecker func() bool) bool {
	targets := make([]CheckTarget, len(b.enemies))
	copy(targets, b.enemies)
	rand.New(rand.NewSource(time.Now().UnixNano()))
	rand.Shuffle(len(targets), func(i, j int) { targets[i], targets[j] = targets[j], targets[i] })
	for _, target := range targets {
		internal.LeftClick(b.hWnd, target.x, target.y)
		time.Sleep(DURATION_BATTLE_ACTION_ATTACK)
		if stateChecker() {
			return true
		}
	}
	return false
}

func (b *BattleActionState) aim(target *CheckTarget, stateChecker func() bool) bool {
	internal.LeftClick(b.hWnd, target.x+15, target.y-22)
	time.Sleep(DURATION_BATTLE_ACTION_ATTACK)
	return stateChecker()
}

func (b *BattleActionState) defend() {
	internal.LeftClick(b.hWnd, BATTLE_COMMAND_DEFENCE.x, BATTLE_COMMAND_DEFENCE.y)
	time.Sleep(DURATION_BATTLE_ACTION_GENERAL)
}

func (b *BattleActionState) escape() {
	internal.LeftClick(b.hWnd, BATTLE_COMMAND_ESCAPE.x, BATTLE_COMMAND_ESCAPE.y)
	time.Sleep(DURATION_BATTLE_ACTION_GENERAL)
}

func (b *BattleActionState) move() {
	internal.LeftClick(b.hWnd, BATTLE_COMMAND_MOVE.x, BATTLE_COMMAND_MOVE.y)
	time.Sleep(DURATION_BATTLE_ACTION_GENERAL)
}

func (b *BattleActionState) recall() {
	internal.LeftClick(b.hWnd, BATTLE_WINDOW_PET_RECALL_BUTTON.x, BATTLE_WINDOW_PET_RECALL_BUTTON.y)
	time.Sleep(DURATION_BATTLE_ACTION_GENERAL)
}

func (b *BattleActionState) openPetSkillWindow() {
	if b.isPetSkillWindowOpened() {
		resetAllWindows(b.hWnd)
	} else {
		closeAllWindows(b.hWnd)
		internal.RightClick(b.hWnd, GAME_WIDTH/2, 28)
		time.Sleep(DURATION_BATTLE_ACTION_GENERAL)
		resetAllWindows(b.hWnd)
	}
}

func (b *BattleActionState) openSkillWindowWithMouse() {
	internal.LeftClick(b.hWnd, BATTLE_COMMAND_SKILL.x, BATTLE_COMMAND_SKILL.y)
	time.Sleep(DURATION_BATTLE_ACTION_GENERAL)
}

func (b *BattleActionState) logH(message string) {
	header := fmt.Sprintf("[%s][%d][%s]", fmt.Sprint(b.hWnd), b.currentHumanActionId, strings.Trim(b.HumanActions[b.currentHumanActionId].Action.String(), "*"))
	log.Printf("%-26s %s",
		header,
		message,
	)
}

func (b *BattleActionState) logP(message string) {
	header := fmt.Sprintf("[%s][%d][%s]", fmt.Sprint(b.hWnd), b.currentPetActionId, strings.Trim(b.PetActions[b.currentPetActionId].Action.String(), "*"))
	log.Printf("%-26s %s",
		header,
		message,
	)
}

func CreateNewBattleActionState(hWnd win.HWND, gameDir, manaChecker *string) BattleActionState {
	return BattleActionState{
		hWnd: hWnd,
		HumanActions: []HumanAction{
			{
				Action:             HumanAttack,
				SuccessControlUnit: Continue,
				FailureControlUnit: Continue,
			},
		},
		PetActions: []PetAction{
			{
				Action:             PetAttack,
				SuccessControlUnit: Continue,
				FailureControlUnit: Continue,
			},
		},
		GameDir:     gameDir,
		ManaChecker: manaChecker,
	}
}

func (b *BattleActionState) GetHWND() win.HWND {
	return b.hWnd
}

func (b *BattleActionState) SetHWND(hWnd win.HWND) {
	b.hWnd = hWnd
}

func (b *BattleActionState) AddHumanAction(action humanAction) {
	newHumanState := HumanAction{
		Action: action,
	}
	b.HumanActions = append(b.HumanActions, newHumanState)
}

func (b *BattleActionState) AddHumanSkillLevel(level Offset) {
	b.HumanActions[len(b.HumanActions)-1].Level = level
}

func (b *BattleActionState) AddPetAction(action petAction) {
	newPetState := PetAction{
		Action: action,
	}
	b.PetActions = append(b.PetActions, newPetState)
}

func (b *BattleActionState) AddSkillOffset(r Role, offset Offset) {
	switch r {
	case Human:
		b.HumanActions[len(b.HumanActions)-1].Offset = offset
	case Pet:
		b.PetActions[len(b.PetActions)-1].Offset = offset
	}
}

func (b *BattleActionState) AddThreshold(r Role, threshold Threshold) {
	b.HumanActions[len(b.HumanActions)-1].Threshold = threshold
}

func (b *BattleActionState) AddParam(r Role, param string) {
	switch r {
	case Human:
		b.HumanActions[len(b.HumanActions)-1].Param = param
	case Pet:
		b.PetActions[len(b.PetActions)-1].Param = param
	}
}

func (b *BattleActionState) AddSuccessControlUnit(r Role, cu ControlUnit) {
	switch r {
	case Human:
		b.HumanActions[len(b.HumanActions)-1].SuccessControlUnit = cu
	case Pet:
		b.PetActions[len(b.PetActions)-1].SuccessControlUnit = cu
	}
}
func (b *BattleActionState) AddSuccessJumpId(r Role, jumpId int) {
	switch r {
	case Human:
		b.HumanActions[len(b.HumanActions)-1].SuccessJumpId = jumpId
	case Pet:
		b.PetActions[len(b.PetActions)-1].SuccessJumpId = jumpId
	}
}

func (b *BattleActionState) AddFailureControlUnit(r Role, cu ControlUnit) {
	switch r {
	case Human:
		b.HumanActions[len(b.HumanActions)-1].FailureControlUnit = cu
	case Pet:
		b.PetActions[len(b.PetActions)-1].FailureControlUnit = cu
	}
}

func (b *BattleActionState) AddFailureJumpId(r Role, jumpId int) {
	switch r {
	case Human:
		b.HumanActions[len(b.HumanActions)-1].FailureJumpId = jumpId
	case Pet:
		b.PetActions[len(b.PetActions)-1].FailureJumpId = jumpId
	}
}

func (b *BattleActionState) ClearHumanActions() {
	b.HumanActions = b.HumanActions[:0]
}

func (b *BattleActionState) ClearPetActions() {
	b.PetActions = b.PetActions[:0]
}

func (b *BattleActionState) GetHumanActions() []HumanAction {
	return b.HumanActions
}

func (b *BattleActionState) GetPetActions() []PetAction {
	return b.PetActions
}

func (b *BattleActionState) detectEnemies() {

	if getScene(b.hWnd) != BATTLE_SCENE || !b.Enabled {
		return
	}

	if b.isInventoryWindowStuck() {
		b.openSkillWindowWithMouse()
	}
	closeAllWindows(b.hWnd)

	if b.trainingCounter < TRAINING_COUNTER_THRESHOLD {
		b.enemies = b.getEnemies(allMonsters)
	}

	if len(b.enemies) == 1 && b.trainingCounter < TRAINING_COUNTER_THRESHOLD {
		b.trainingCounter++
	}

	log.Printf("# Handle %s detected %d enemies\n", fmt.Sprint(b.hWnd), len(b.enemies))
}

func (b *BattleActionState) endHumanHanging() {
	if b.isHumanHanging {
		b.isHumanHanging = false
		cu := b.HumanActions[b.currentHumanActionId].SuccessControlUnit
		jumpId := b.HumanActions[b.currentHumanActionId].SuccessJumpId

		switch cu {
		case StartOver:
			b.currentHumanActionId = 0
		case Continue:
			b.currentHumanActionId++
		case Jump:
			b.currentHumanActionId = jumpId
		}
	}
}
func (b *BattleActionState) endPetHanging() {
	if b.isPetHanging {
		b.isPetHanging = false
		cu := b.PetActions[b.currentPetActionId].SuccessControlUnit
		jumpId := b.PetActions[b.currentPetActionId].SuccessJumpId

		switch cu {
		case StartOver:
			b.currentPetActionId = 0
		case Continue:
			b.currentPetActionId++
		case Jump:
			b.currentPetActionId = jumpId
		}
	}
}

func (b *BattleActionState) checkHumanMana() {

	if getScene(b.hWnd) != BATTLE_SCENE || !b.Enabled {
		return
	}

	if b.isManaChecker() {
		b.logH("checks human mana")

		closeAllWindows(b.hWnd)
		clearChat(b.hWnd)
		if b.isOutOfMana = b.isAnyPlayerOutOfMana(); b.isOutOfMana {
			b.logH("someone is out of mana")
		}
	}
}

func (b *BattleActionState) wait() {

	for getScene(b.hWnd) == BATTLE_SCENE && b.Enabled && !b.isHumanStageStable() && !b.isPetStageStable() {
		time.Sleep(DURATION_BATTLE_ACTION_WAITING_LOOP)
	}
}

func TestAction(hWnd win.HWND) (x int32, y int32, successful bool) {
	closeAllWindows(hWnd)
	clearChat(hWnd)
	openWindow(hWnd, KEY_INVENTORY)
	return getItemPos(hWnd, x, y, 16448250, 3)
}

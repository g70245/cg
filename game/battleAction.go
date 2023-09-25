package game

import (
	sys "cg/system"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	. "github.com/lxn/win"
	"golang.org/x/exp/slices"
)

const (
	TURN_INTERVAL          = 300
	WAITING_LOOP_INTERVAL  = 200
	ATTACK_INTERVAL        = 100
	BATTLE_ACTION_INTERVAL = 160
)

const (
	H_F_ATTACK = "**Attack"
	H_F_DEFEND = "**Defend"
	H_F_ESCAPE = "**Escape"
	H_F_MOVE   = "**Move"

	H_C_BOMB       = "*Bomb"
	H_C_PET_RECALL = "*Recall Pet"
	H_C_POTION     = "*Potion"
	H_C_SKILL      = "*Skill"
	H_C_SE_HEAL    = "*Heal Self"
	H_C_O_HEAL     = "*Heal One"
	H_C_M_HEAL     = "*Heal Multi"
	H_C_RIDE       = "*Ride"

	H_S_HANG        = "Hang"
	H_S_STEAL       = "Steal"
	H_S_CATCH       = "Catch"
	H_S_TRAIN_SKILL = "Train Skill"

	P_F_ATTACK   = "**Pet Attack"
	P_F_ESCAPE   = "**Pet Escape"
	P_C_SkILL    = "*Pet Skill"
	P_C_DEFEND   = "*Pet Defend"
	P_C_SE_HEAL  = "*Pet Heal Self"
	P_C_O_HEAL   = "*Pet Heal One"
	P_C_RIDE     = "*Pet Ride"
	P_C_OFF_RIDE = "*Pet Off Ride"
	P_S_CATCH    = "*Pet Catch"
	P_S_HANG     = "Pet Hang"
)

const (
	C_U_START_OVER = "Start Over"
	C_U_CONTINUE   = "Continue"
	C_U_REPEAT     = "Repeat"
	C_U_JUMP       = "Jump"
)

var ControlUnitOptions = []string{C_U_START_OVER, C_U_CONTINUE, C_U_REPEAT, C_U_JUMP}
var HealingOptions = []string{"0.1", "0.2", "0.3", "0.4", "0.5", "0.6", "0.7", "0.8", "0.9"}
var IdOptions = []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}
var LevelOptions = []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}

type BattleActionState struct {
	hWnd                     HWND `json:"-"`
	HumanStates              []string
	nextHumanStateId         int `json:"-"`
	HumanSkillIds            []string
	HumanSkillLevels         []string
	HumanParams              []string
	HumanSuccessControlUnits []string
	HumanFailureControlUnits []string

	PetStates              []string
	nextPetStateId         int `json:"-"`
	PetSkillIds            []string
	PetParams              []string
	PetSuccessControlUnits []string
	PetFailureControlUnits []string

	Enabled            bool `json:"-"`
	isOutOfMana        bool `json:"-"`
	isHumanHanging     bool `json:"-"`
	isPetHanging       bool `json:"-"`
	isEncounteringBaBy bool `json:"-"`

	logDir *string `json:"-"`

	enemies []CheckTarget
}

func (b *BattleActionState) Act() {
	log.Printf("# Handle %s's battle begins\n", fmt.Sprint(b.hWnd))

	for getScene(b.hWnd) == BATTLE_SCENE && b.Enabled {
		b.detectEnemies()
		fmt.Println(b.enemies)
		b.executeHumanStateMachine()
		b.executePetStateMachiine()
		time.Sleep(TURN_INTERVAL * time.Millisecond)
	}

	b.nextHumanStateId = 0
	b.nextPetStateId = 0
	log.Printf("@ Handle %s's battle ended\n", fmt.Sprint(b.hWnd))
}

func (b *BattleActionState) executeHumanStateMachine() {

	for b.nextHumanStateId < len(b.HumanStates) && getScene(b.hWnd) == BATTLE_SCENE && !isPetStageStable(b.hWnd) {
		if !isHumanStageStable(b.hWnd) {
			if !b.Enabled {
				return
			}
			time.Sleep(WAITING_LOOP_INTERVAL * time.Millisecond)
			continue
		}

		if b.isPetHanging {
			b.isPetHanging = false
			cu := b.PetSuccessControlUnits[b.nextPetStateId]

			var offset int
			if strings.Contains(cu, C_U_JUMP) {
				i, _ := strconv.Atoi(cu[4:])
				offset = i
				cu = cu[:4]
			}

			switch cu {
			case C_U_START_OVER:
				b.nextPetStateId = 0
			case C_U_CONTINUE:
				b.nextPetStateId++
			case C_U_JUMP:
				b.nextPetStateId = offset
			}
		}

		var cu string

		switch b.HumanStates[b.nextHumanStateId] {
		case H_F_ATTACK:
			if b.isEncounteringBaBy {
				break
			}

			b.enableBattleCommandAttack()
			if b.attack(HumanTargetingChecker) {
				b.logH("attacked")
				cu = b.HumanSuccessControlUnits[b.nextHumanStateId]
			} else {
				b.logH("missed a hit")
				cu = b.HumanFailureControlUnits[b.nextHumanStateId]
			}
		case H_C_SKILL:
			if b.isEncounteringBaBy {
				break
			}

			openWindowByShortcut(b.hWnd, 0x57)
			if x, y, ok := getSkillWindowPos(b.hWnd); ok {
				id, _ := strconv.Atoi(b.HumanSkillIds[b.nextHumanStateId])
				level, _ := strconv.Atoi(b.HumanSkillLevels[b.nextHumanStateId])
				useHumanSkill(b.hWnd, x, y, id, level)
				if isHumanOutOfMana(b.hWnd, x, y) {
					b.logH("is out of mana")
					b.isOutOfMana = true
				} else if b.attack(HumanTargetingChecker) {
					b.logH("used a skill")
					cu = b.HumanSuccessControlUnits[b.nextHumanStateId]
				} else {
					b.logH("missed a hit")
					cu = b.HumanFailureControlUnits[b.nextHumanStateId]
				}
			} else {
				b.logH("cannot find the position of window")
				cu = b.HumanFailureControlUnits[b.nextHumanStateId]
			}
		case H_F_DEFEND:
			b.defend()
			b.logH("defended")
			cu = b.HumanSuccessControlUnits[b.nextHumanStateId]
		case H_F_MOVE:
			b.move()
			b.logH("moved")
			cu = b.HumanSuccessControlUnits[b.nextHumanStateId]
		case H_F_ESCAPE:
			if b.isEncounteringBaBy {
				break
			}

			b.escape()
			b.logH("escaped")
			cu = b.HumanFailureControlUnits[b.nextHumanStateId]
		case H_S_HANG:
			if slices.Contains(b.HumanStates, H_S_CATCH) && !b.isEncounteringBaBy {
				break
			}

			b.logH("is hanging")
			b.isHumanHanging = true
			return
		case H_C_BOMB:
			if b.isEncounteringBaBy {
				break
			}

			openWindowByShortcut(b.hWnd, 0x45)
			if px, py, isPivotFound := getBSItemWindowPos(b.hWnd); isPivotFound {
				var bomb Item
				for i := range Bombs {
					if Bombs[i].name == b.HumanParams[b.nextHumanStateId] {
						bomb = Bombs[i]
					}
				}
				if x, y, ok := getItemPos(b.hWnd, px, py, bomb.color, 3); ok {
					sys.DoubleClick(HWND(b.hWnd), x, y)
					if isBSItemWindowStillOpened(b.hWnd, px, py) {
						b.logH("failed at double clicking")
						cu = b.HumanFailureControlUnits[b.nextHumanStateId]
					} else if b.attack(HumanTargetingChecker) {
						b.logH("throwed a bomb")
						cu = b.HumanSuccessControlUnits[b.nextHumanStateId]
					} else {
						b.logH("missed a hit")
						cu = b.HumanFailureControlUnits[b.nextHumanStateId]
					}

				} else {
					b.logH("cannot find a bomb")
					cu = b.HumanFailureControlUnits[b.nextHumanStateId]
				}
			} else {
				b.logH("cannot find the position of window")
				cu = b.HumanFailureControlUnits[b.nextHumanStateId]
			}
		case H_C_POTION:
			closeAllWindow(b.hWnd)
			clearChat(b.hWnd)
			ratio, _ := strconv.ParseFloat(b.HumanParams[b.nextHumanStateId], 32)
			if target, ok := searchLifeBelow(b.hWnd, float32(ratio)); ok {
				openWindowByShortcut(b.hWnd, 0x45)
				if px, py, isPivotFound := getBSItemWindowPos(b.hWnd); isPivotFound {
					if x, y, ok := getItemPos(b.hWnd, px, py, COLOR_ITEM_POTION, 3); ok {
						sys.DoubleClick(HWND(b.hWnd), x, y)
						if isBSItemWindowStillOpened(b.hWnd, px, py) {
							b.logH("failed at double clicking")
							cu = b.HumanFailureControlUnits[b.nextHumanStateId]
						} else if b.aim(target, HumanTargetingChecker) {
							b.logH("healed an ally")
							cu = b.HumanSuccessControlUnits[b.nextHumanStateId]
						} else {
							b.logH("cannot target")
							cu = b.HumanFailureControlUnits[b.nextHumanStateId]
						}
					} else {
						b.logH("cannot find a potion")
						cu = b.HumanFailureControlUnits[b.nextHumanStateId]
					}
				} else {
					b.logH("cannot find the position of window")
					cu = b.HumanFailureControlUnits[b.nextHumanStateId]
				}
			} else {
				b.logH("found no one needed to be taken care of")
				cu = b.HumanSuccessControlUnits[b.nextHumanStateId]
			}
		case H_C_RIDE:
			if b.isEncounteringBaBy {
				break
			}

			openWindowByShortcut(b.hWnd, 0x57)
			if x, y, ok := getSkillWindowPos(b.hWnd); ok {
				id, _ := strconv.Atoi(b.HumanSkillIds[b.nextHumanStateId])
				level, _ := strconv.Atoi(b.HumanSkillLevels[b.nextHumanStateId])
				useHumanSkill(b.hWnd, x, y, id, level)
				if isHumanOutOfMana(b.hWnd, x, y) {
					b.logH("is out of mana")
					b.isOutOfMana = true
				} else {
					b.logH("is tring to get on a pet")
					for i, v := range b.PetStates {
						if v == P_C_RIDE {
							b.nextPetStateId = i
							break
						}
					}
					cu = b.HumanSuccessControlUnits[b.nextHumanStateId]
				}
			} else {
				b.logH("cannot find the position of window")
				cu = b.HumanFailureControlUnits[b.nextHumanStateId]
			}
		case H_C_SE_HEAL:
			closeAllWindow(b.hWnd)
			clearChat(b.hWnd)
			if self, ok := getSelfTarget(b.hWnd, true); ok {
				ratio, _ := strconv.ParseFloat(b.HumanParams[b.nextHumanStateId], 32)

				if !isLifeBelow(b.hWnd, float32(ratio), self) {
					b.logH("is a healthy boy")
					cu = b.HumanSuccessControlUnits[b.nextHumanStateId]
					break
				}

				openWindowByShortcut(b.hWnd, 0x57)
				if x, y, ok := getSkillWindowPos(b.hWnd); ok {
					id, _ := strconv.Atoi(b.HumanSkillIds[b.nextHumanStateId])
					level, _ := strconv.Atoi(b.HumanSkillLevels[b.nextHumanStateId])
					useHumanSkill(b.hWnd, x, y, id, level)
					if isHumanOutOfMana(b.hWnd, x, y) {
						b.logH("is out of mana")
						b.isOutOfMana = true
					} else {
						b.logH("healed self")
						cu = b.HumanSuccessControlUnits[b.nextHumanStateId]
					}
				} else {
					b.logH("cannot find the position of window")
					cu = b.HumanFailureControlUnits[b.nextHumanStateId]
				}

			} else {
				b.logH("cannot find self")
				cu = b.HumanFailureControlUnits[b.nextHumanStateId]
			}
		case H_C_O_HEAL:
			closeAllWindow(b.hWnd)
			clearChat(b.hWnd)
			ratio, _ := strconv.ParseFloat(b.HumanParams[b.nextHumanStateId], 32)
			if target, ok := searchLifeBelow(b.hWnd, float32(ratio)); ok {
				openWindowByShortcut(b.hWnd, 0x57)
				if x, y, ok := getSkillWindowPos(b.hWnd); ok {
					id, _ := strconv.Atoi(b.HumanSkillIds[b.nextHumanStateId])
					level, _ := strconv.Atoi(b.HumanSkillLevels[b.nextHumanStateId])
					useHumanSkill(b.hWnd, x, y, id, level)
					if isHumanOutOfMana(b.hWnd, x, y) {
						b.logH("is out of mana")
						b.isOutOfMana = true
					} else if b.aim(target, HumanTargetingChecker) {
						b.logH("healed an ally")
						cu = b.HumanSuccessControlUnits[b.nextHumanStateId]
					} else {
						b.logH("cannot target")
						cu = b.HumanFailureControlUnits[b.nextHumanStateId]
					}
				} else {
					b.logH("cannot find the position of window")
					cu = b.HumanFailureControlUnits[b.nextHumanStateId]
				}
			} else {
				b.logH("found no one needed to be taken care of")
				cu = b.HumanSuccessControlUnits[b.nextHumanStateId]
			}
		case H_C_M_HEAL:
			closeAllWindow(b.hWnd)
			clearChat(b.hWnd)
			ratio, _ := strconv.ParseFloat(b.HumanParams[b.nextHumanStateId], 32)
			count := countLifeBelow(b.hWnd, float32(ratio))
			if count >= 5 {
				openWindowByShortcut(b.hWnd, 0x57)
				if x, y, ok := getSkillWindowPos(b.hWnd); ok {
					id, _ := strconv.Atoi(b.HumanSkillIds[b.nextHumanStateId])
					level, _ := strconv.Atoi(b.HumanSkillLevels[b.nextHumanStateId])
					useHumanSkill(b.hWnd, x, y, id, level)
					if isHumanOutOfMana(b.hWnd, x, y) {
						b.logH("is out of mana")
						b.isOutOfMana = true
					} else if b.aim(&PLAYER_L_3_H, HumanTargetingChecker) {
						b.logH("healed allies")
						cu = b.HumanSuccessControlUnits[b.nextHumanStateId]
					} else {
						b.logH("cannot target")
						cu = b.HumanFailureControlUnits[b.nextHumanStateId]
					}
				} else {
					b.logH("cannot find the position of window")
					cu = b.HumanFailureControlUnits[b.nextHumanStateId]
				}
			} else {
				b.logH("found all good")
				cu = b.HumanSuccessControlUnits[b.nextHumanStateId]
			}
		case H_C_PET_RECALL:
			if slices.Contains(b.HumanStates, H_S_CATCH) && !b.isEncounteringBaBy {
				break
			}

			openWindowByShortcut(b.hWnd, 0x52)
			if canRecall(b.hWnd) {
				b.recall()
				b.logH("recalled")
				return
			} else {
				b.logH("already recalled")
				cu = b.HumanSuccessControlUnits[b.nextHumanStateId]
			}
		case H_S_TRAIN_SKILL:
			openWindowByShortcut(b.hWnd, 0x57)
			if x, y, ok := getSkillWindowPos(b.hWnd); ok {
				id, _ := strconv.Atoi(b.HumanSkillIds[b.nextHumanStateId])
				level, _ := strconv.Atoi(b.HumanSkillLevels[b.nextHumanStateId])
				useHumanSkill(b.hWnd, x, y, id, level)
				if isHumanOutOfMana(b.hWnd, x, y) {
					b.logH("is out of mana")
					b.isOutOfMana = true
				} else if b.aim(&PLAYER_L_3_P, HumanTargetingChecker) {
					b.logH("is training")
					cu = b.HumanSuccessControlUnits[b.nextHumanStateId]
				}
			} else {
				b.logH("cannot find the position of window")
				cu = b.HumanFailureControlUnits[b.nextHumanStateId]
			}
		case H_S_CATCH:
			if *b.logDir == "" {
				b.logH("please set up the log directory")
				break
			}

			if !DoesEncounterBaby(*b.logDir) {
				break
			}

			b.logH("encounters a baby")
			b.isEncounteringBaBy = true
			sys.PlayBeeper()
		}

		var offset int
		if strings.Contains(cu, C_U_JUMP) {
			i, _ := strconv.Atoi(cu[4:])
			offset = i
			cu = cu[:4]
		}
		switch cu {
		case C_U_START_OVER:
			b.nextHumanStateId = 0
		case C_U_CONTINUE:
			b.nextHumanStateId++
		case C_U_REPEAT:
			return
		case C_U_JUMP:
			b.nextHumanStateId = offset
		default: // out of mana
			b.nextHumanStateId++
		}
	}

	if len(b.HumanStates) != 0 {
		b.nextHumanStateId %= len(b.HumanStates)
	} else {
		b.nextHumanStateId = 0
	}
}

func (b *BattleActionState) executePetStateMachiine() {

	for b.nextPetStateId < len(b.PetStates) && getScene(b.hWnd) == BATTLE_SCENE && !isHumanStageStable(b.hWnd) {
		if !isPetStageStable(b.hWnd) && getScene(b.hWnd) == BATTLE_SCENE {
			if !b.Enabled {
				return
			}
			time.Sleep(WAITING_LOOP_INTERVAL * time.Millisecond)
			continue
		}

		if b.isHumanHanging {
			b.isHumanHanging = false
			cu := b.HumanSuccessControlUnits[b.nextHumanStateId]

			var offset int
			if strings.Contains(cu, C_U_JUMP) {
				i, _ := strconv.Atoi(cu[4:])
				offset = i
				cu = cu[:4]
			}
			switch cu {
			case C_U_START_OVER:
				b.nextHumanStateId = 0
			case C_U_CONTINUE:
				b.nextHumanStateId++
			case C_U_JUMP:
				b.nextHumanStateId = offset
			}
		}

		var cu string

		switch b.PetStates[b.nextPetStateId] {
		case P_F_ATTACK:
			if b.isEncounteringBaBy {
				break
			}

			b.openPetSkillWindow()
			if x, y, ok := getSkillWindowPos(b.hWnd); ok {
				usePetSkill(b.hWnd, x, y, 1)
				if b.attack(PetTargetingChecker) {
					b.logP("attacked")
					cu = b.PetSuccessControlUnits[b.nextPetStateId]
				} else {
					b.logP("missed a hit")
					cu = b.PetFailureControlUnits[b.nextPetStateId]
				}
			} else {
				b.logP("cannot find the position of window")
				cu = b.PetFailureControlUnits[b.nextPetStateId]
			}

		case P_C_SkILL:
			if b.isEncounteringBaBy {
				break
			}

			b.openPetSkillWindow()
			if x, y, ok := getSkillWindowPos(b.hWnd); ok {
				id, _ := strconv.Atoi(b.PetSkillIds[b.nextPetStateId])
				usePetSkill(b.hWnd, x, y, id)
				if isPetOutOfMana(b.hWnd) || isOnRidingOutOfMana(b.hWnd) {
					b.logP("is out of mana")
					b.isOutOfMana = true
				} else if b.attack(PetTargetingChecker) {
					b.logP("used a skill")
					cu = b.PetSuccessControlUnits[b.nextPetStateId]
				} else {
					b.logP("missed a hit")
					cu = b.PetFailureControlUnits[b.nextPetStateId]
				}
			} else {
				b.logP("cannot find the position of window")
				cu = b.PetFailureControlUnits[b.nextPetStateId]
			}
		case P_S_HANG:
			if slices.Contains(b.HumanStates, H_S_CATCH) && !b.isEncounteringBaBy {
				break
			}

			b.logP("is hanging")
			b.isPetHanging = true
			return
		case P_C_DEFEND:
			b.openPetSkillWindow()
			if x, y, ok := getSkillWindowPos(b.hWnd); ok {
				id, _ := strconv.Atoi(b.PetSkillIds[b.nextPetStateId])
				usePetSkill(b.hWnd, x, y, id)
				if isPetOutOfMana(b.hWnd) || isOnRidingOutOfMana(b.hWnd) {
					b.logP("is out of mana")
					b.isOutOfMana = true
				} else {
					b.logP("defended")
					cu = b.PetSuccessControlUnits[b.nextPetStateId]
				}
			} else {
				b.logP("cannot find the position of window")
				cu = b.PetFailureControlUnits[b.nextPetStateId]
			}
		case P_C_SE_HEAL:
			closeAllWindow(b.hWnd)
			clearChat(b.hWnd)
			if self, ok := getSelfTarget(b.hWnd, false || isOnRide(b.hWnd)); ok {
				ratio, _ := strconv.ParseFloat(b.PetParams[b.nextPetStateId], 32)
				if !isLifeBelow(b.hWnd, float32(ratio), self) {
					b.logP("is a healthy boy")
					cu = b.PetSuccessControlUnits[b.nextPetStateId]
					break
				}

				b.openPetSkillWindow()
				if x, y, ok := getSkillWindowPos(b.hWnd); ok {
					id, _ := strconv.Atoi(b.PetSkillIds[b.nextPetStateId])
					usePetSkill(b.hWnd, x, y, id)
					if isPetOutOfMana(b.hWnd) || isOnRidingOutOfMana(b.hWnd) {
						b.logP("is out of mana")
						b.isOutOfMana = true
					} else {
						b.logP("healed self")
						cu = b.PetSuccessControlUnits[b.nextPetStateId]
					}
				} else {
					b.logP("cannot find the position of window")
					cu = b.PetFailureControlUnits[b.nextPetStateId]
				}
			} else {
				b.logP("cannot find self")
				cu = b.PetFailureControlUnits[b.nextPetStateId]
			}
		case P_C_RIDE:
			if isOnRide(b.hWnd) {
				b.logP("is on ride")
				cu = b.PetSuccessControlUnits[b.nextPetStateId]
				break
			}

			b.openPetSkillWindow()
			if x, y, ok := getSkillWindowPos(b.hWnd); ok {
				id, _ := strconv.Atoi(b.PetSkillIds[b.nextPetStateId])
				usePetSkill(b.hWnd, x, y, id)
				b.logP("tries to get on ride")
				return
			} else {
				b.logP("cannot find the position of window")
				cu = b.PetFailureControlUnits[b.nextPetStateId]
			}
		case P_C_OFF_RIDE:
			if slices.Contains(b.HumanStates, H_S_CATCH) && !b.isEncounteringBaBy {
				break
			}

			if !isOnRide(b.hWnd) {
				b.logP("is off ride")
				cu = b.PetSuccessControlUnits[b.nextPetStateId]
				break
			}

			b.openPetSkillWindow()
			if x, y, ok := getSkillWindowPos(b.hWnd); ok {
				id, _ := strconv.Atoi(b.PetSkillIds[b.nextPetStateId])
				usePetSkill(b.hWnd, x, y, id)
				b.logP("tries to get off ride")
				return
			} else {
				b.logP("cannot find the position of window")
				cu = b.PetFailureControlUnits[b.nextPetStateId]
			}
		case P_C_O_HEAL:
			closeAllWindow(b.hWnd)
			clearChat(b.hWnd)
			ratio, _ := strconv.ParseFloat(b.PetParams[b.nextPetStateId], 32)
			if target, ok := searchLifeBelow(b.hWnd, float32(ratio)); ok {
				b.openPetSkillWindow()
				if x, y, ok := getSkillWindowPos(b.hWnd); ok {
					id, _ := strconv.Atoi(b.PetSkillIds[b.nextPetStateId])
					usePetSkill(b.hWnd, x, y, id)
					if b.aim(target, PetTargetingChecker) {
						b.logP("healed an ally")
						cu = b.PetSuccessControlUnits[b.nextPetStateId]
					} else {
						b.logP("cannot target")
						cu = b.PetFailureControlUnits[b.nextPetStateId]
					}
				} else {
					b.logP("cannot find the position of window")
					cu = b.PetFailureControlUnits[b.nextPetStateId]
				}

			} else {
				b.logP("found no one needed to be taken care of")
				cu = b.PetSuccessControlUnits[b.nextPetStateId]
			}
		case P_F_ESCAPE:
			if b.isEncounteringBaBy {
				break
			}

			if !isOnRide(b.hWnd) {
				b.logP("cannot escape while off ride")
				break
			}

			b.escape()
			b.logP("escaped")
			cu = b.PetFailureControlUnits[b.nextPetStateId]
		case P_S_CATCH:
			if *b.logDir == "" {
				b.logH("please set up the log directory")
				break
			}

			if !DoesEncounterBaby(*b.logDir) {
				break
			}

			b.logH("encounters a baby")
			b.isEncounteringBaBy = true
			sys.PlayBeeper()
		}

		var offset int
		if strings.Contains(cu, C_U_JUMP) {
			i, _ := strconv.Atoi(cu[4:])
			offset = i
			cu = cu[:4]
		}
		switch cu {
		case C_U_START_OVER:
			b.nextPetStateId = 0
		case C_U_CONTINUE:
			b.nextPetStateId++
		case C_U_REPEAT:
			return
		case C_U_JUMP:
			b.nextPetStateId = offset
		default: // out of mana
			b.nextPetStateId++
		}
	}

	if len(b.PetStates) != 0 {
		b.nextPetStateId %= len(b.PetStates)
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

func (b *BattleActionState) attack(stateChecker func(hwnd HWND) bool) bool {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	rand.Shuffle(len(b.enemies), func(i, j int) { b.enemies[i], b.enemies[j] = b.enemies[j], b.enemies[i] })
	for _, target := range b.enemies {
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
	time.Sleep(BATTLE_ACTION_INTERVAL * time.Millisecond)
	resetAllWindowsPosition(b.hWnd)
}

func (b *BattleActionState) logH(message string) {
	header := fmt.Sprintf("[%s][%s]", fmt.Sprint(b.hWnd), strings.Trim(b.HumanStates[b.nextHumanStateId], "*"))
	log.Printf("%-26s %s",
		header,
		message,
	)
}

func (b *BattleActionState) logP(message string) {
	header := fmt.Sprintf("[%s][%s]", fmt.Sprint(b.hWnd), strings.Trim(b.PetStates[b.nextPetStateId], "*"))
	log.Printf("%-26s %s",
		header,
		message,
	)
}

func CreateNewBattleActionState(hWnd HWND, logDir *string) BattleActionState {
	return BattleActionState{
		hWnd:                     hWnd,
		HumanStates:              []string{H_F_ATTACK},
		HumanSkillIds:            []string{""},
		HumanSkillLevels:         []string{""},
		HumanParams:              []string{""},
		HumanSuccessControlUnits: []string{C_U_CONTINUE},
		HumanFailureControlUnits: []string{C_U_CONTINUE},
		PetStates:                []string{P_F_ATTACK},
		PetSkillIds:              []string{""},
		PetParams:                []string{""},
		PetSuccessControlUnits:   []string{C_U_CONTINUE},
		PetFailureControlUnits:   []string{C_U_CONTINUE},
		logDir:                   logDir,
	}
}

func (b *BattleActionState) GetHWND() HWND {
	return b.hWnd
}

func (b *BattleActionState) SetHWND(hWnd HWND) {
	b.hWnd = hWnd
}

func (b *BattleActionState) AddHumanState(newState string) {
	b.HumanStates = append(b.HumanStates, newState)
	b.HumanSkillIds = append(b.HumanSkillIds, "")
	b.HumanSkillLevels = append(b.HumanSkillLevels, "")
	b.HumanParams = append(b.HumanParams, "")
	b.HumanSuccessControlUnits = append(b.HumanSuccessControlUnits, "")
	b.HumanFailureControlUnits = append(b.HumanFailureControlUnits, "")
}

func (b *BattleActionState) AddPetState(newState string) {
	b.PetStates = append(b.PetStates, newState)
	b.PetSkillIds = append(b.PetSkillIds, "")
	b.PetParams = append(b.PetParams, "")
	b.PetSuccessControlUnits = append(b.PetSuccessControlUnits, "")
	b.PetFailureControlUnits = append(b.PetFailureControlUnits, "")
}

func (b *BattleActionState) AddHumanSkillId(newSkillId string) {
	b.HumanSkillIds[len(b.HumanSkillIds)-1] = newSkillId
}

func (b *BattleActionState) AddHumanSkillLevel(newLevel string) {
	b.HumanSkillLevels[len(b.HumanSkillLevels)-1] = newLevel
}

func (b *BattleActionState) AddHumanParam(param string) {
	b.HumanParams[len(b.HumanParams)-1] = param
}

func (b *BattleActionState) AddHumanSuccessControlUnit(param string) {
	b.HumanSuccessControlUnits[len(b.HumanSuccessControlUnits)-1] = param
}

func (b *BattleActionState) AddHumanFailureControlUnit(param string) {
	b.HumanFailureControlUnits[len(b.HumanFailureControlUnits)-1] = param
}

func (b *BattleActionState) AddPetSkillId(newSkillId string) {
	b.PetSkillIds[len(b.PetSkillIds)-1] = newSkillId
}

func (b *BattleActionState) AddPetParam(param string) {
	b.PetParams[len(b.PetParams)-1] = param
}

func (b *BattleActionState) AddPetSuccessControlUnit(param string) {
	b.PetSuccessControlUnits[len(b.PetSuccessControlUnits)-1] = param
}

func (b *BattleActionState) AddPetFailureControlUnit(param string) {
	b.PetFailureControlUnits[len(b.PetFailureControlUnits)-1] = param
}

func (b *BattleActionState) ClearHumanStates() {
	b.HumanStates = b.HumanStates[:0]
	b.HumanSkillIds = b.HumanSkillIds[:0]
	b.HumanSkillLevels = b.HumanSkillLevels[:0]
	b.HumanParams = b.HumanParams[:0]
	b.HumanSuccessControlUnits = b.HumanSuccessControlUnits[:0]
	b.HumanFailureControlUnits = b.HumanFailureControlUnits[:0]
}

func (b *BattleActionState) ClearPetStates() {
	b.PetStates = b.PetStates[:0]
	b.PetSkillIds = b.PetSkillIds[:0]
	b.PetParams = b.PetParams[:0]
	b.PetSuccessControlUnits = b.PetSuccessControlUnits[:0]
	b.PetFailureControlUnits = b.PetFailureControlUnits[:0]
}

func (b *BattleActionState) GetHumanStates() []string {
	return b.HumanStates
}

func (b *BattleActionState) GetHumanSkillIds() []string {
	return b.HumanSkillIds
}

func (b *BattleActionState) GetHumanSkillLevels() []string {
	return b.HumanSkillLevels
}

func (b *BattleActionState) GetHumanParams() []string {
	return b.HumanParams
}

func (b *BattleActionState) GetHumanSuccessControlUnits() []string {
	return b.HumanSuccessControlUnits
}
func (b *BattleActionState) GetHumanFailureControlUnits() []string {
	return b.HumanFailureControlUnits
}

func (b *BattleActionState) GetPetStates() []string {
	return b.PetStates
}

func (b *BattleActionState) GetPetSkillIds() []string {
	return b.PetSkillIds
}

func (b *BattleActionState) GetPetParams() []string {
	return b.PetParams
}

func (b *BattleActionState) GetPetSuccessControlUnits() []string {
	return b.PetSuccessControlUnits
}

func (b *BattleActionState) GetPetFailureControlUnits() []string {
	return b.PetFailureControlUnits
}

func (b *BattleActionState) detectEnemies() {
	b.enemies = getEnemyTargets(b.hWnd)
}

func Test(hWnd HWND) (x int32, y int32, successful bool) {
	closeAllWindow(hWnd)
	clearChat(hWnd)
	openWindowByShortcut(hWnd, 0x45)
	x, y, successful = getBSItemWindowPos(hWnd)
	return getItemPos(hWnd, x, y, 16448250, 3)
	// return searchSlotForColor(hWnd, x, y, 30719, 1)
}

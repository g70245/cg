package battle

import (
	"cg/game"
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

type ActionState struct {
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

	enemies         []game.CheckTarget `json:"-"`
	trainingCounter int                `json:"-"`
}

func (s *ActionState) Act() {
	log.Printf("# Handle %s's battle begins\n", fmt.Sprint(s.hWnd))

	for game.IsBattleScene(s.hWnd) && s.Enabled {
		s.wait()
		s.executeActivity()
		s.detectEnemies()
		s.checkCharacterMana()
		s.executeHumanStateMachine()
		s.wait()
		s.executeHumanStateMachine()
		s.wait()
		s.executePetStateMachiine()
		s.wait()
		s.executePetStateMachiine()
	}

	s.reset()
	log.Printf("@ Handle %s's battle ended\n", fmt.Sprint(s.hWnd))
}

func (s *ActionState) executeActivity() {

	if !s.ActivityCheckerEnabled {
		return
	}

	if game.DoesEncounterActivityMonsters(*s.GameDir) {
		s.logH("encounters the activity monster")
		utils.Beeper.Play()

		for game.IsBattleScene(s.hWnd) && s.Enabled {
			time.Sleep(DURATION_BATTLE_ACTION_WAITING_LOOP)
		}
	}
}

func (s *ActionState) executeHumanStateMachine() {

	for s.currentHumanActionId < len(s.HumanActions) && game.IsBattleScene(s.hWnd) && s.isHumanStageStable() && s.Enabled {

		s.resetCurrentControlUnit()
		s.endPetHanging()

		switch s.HumanActions[s.currentHumanActionId].Action {
		case HumanAttack:
			s.enableBattleCommandAttack()
			if s.attack(s.isHumanActionSuccessful) {
				s.logH("attacked")
				s.setSuccessState(Human)
			} else {
				s.logH("missed a hit")
				s.setFailureState(Human)
			}
		case HumanSkill:
			game.OpenWindow(s.hWnd, game.KEY_SKILL)
			if x, y, ok := s.getSkillWindowPos(); ok {
				offset := int(s.HumanActions[s.currentHumanActionId].Offset)
				level := int(s.HumanActions[s.currentHumanActionId].Level)
				game.UseHumanSkill(s.hWnd, x, y, offset, level)
				if s.didHumanMissSkill(x, y) {
					s.logH("missed the skill button or is out of mana")
				} else if s.isHumanActionSuccessful() {
					s.logH("used a skill")
					s.setSuccessState(Human)
				} else if s.attack(s.isHumanActionSuccessful) {
					s.logH("used a skill")
					s.setSuccessState(Human)
				} else {
					s.logH("missed a hit")
					s.setFailureState(Human)
				}
			} else {
				s.logH("cannot find the position of window")
				s.setFailureState(Human)
			}

		case HumanThresholdSkill:
			threshold, _ := strconv.Atoi(strings.Split(string(s.HumanActions[s.currentHumanActionId].Threshold), " ")[0])
			if len(s.enemies) < threshold {
				s.logH("performs next action due to too few enemies")
				break
			}

			game.OpenWindow(s.hWnd, game.KEY_SKILL)
			if x, y, ok := s.getSkillWindowPos(); ok {
				offset := int(s.HumanActions[s.currentHumanActionId].Offset)
				level := int(s.HumanActions[s.currentHumanActionId].Level)
				game.UseHumanSkill(s.hWnd, x, y, offset, level)
				if s.didHumanMissSkill(x, y) {
					s.logH("missed the skill button or is out of mana")
				} else if s.isHumanActionSuccessful() {
					s.logH("used a skill")
					s.setSuccessState(Human)
				} else if s.attack(s.isHumanActionSuccessful) {
					s.logH("used a skill")
					s.setSuccessState(Human)
				} else {
					s.logH("missed a hit")
					s.setFailureState(Human)
				}
			} else {
				s.logH("cannot find the position of window")
				s.setFailureState(Human)
			}
		case HumanDefend:
			s.defend()
			s.logH("defended")
			s.setSuccessState(Human)
		case HumanMove:
			s.move()
			s.logH("moved")
			s.setSuccessState(Human)
		case HumanEscape:
			s.escape()
			s.logH("escaped")
			s.setFailureState(Human)
		case HumanHang:
			s.logH("is hanging")
			s.isHumanHanging = true
			s.currentControlUnit = Repeat
		case HumanBomb:
			var bomb Item
			for i := range Bombs {
				if Bombs[i].name == s.HumanActions[s.currentHumanActionId].Param {
					bomb = Bombs[i]
				}
			}

			if bomb == I_B_9A && len(s.enemies) < 5 {
				s.logH("performs next action due to too few enemies")
				break
			}

			game.OpenWindow(s.hWnd, game.KEY_INVENTORY)
			if px, py, isPivotFound := s.getInventoryPos(); isPivotFound {
				if x, y, ok := game.GetItemPos(s.hWnd, px, py, bomb.color, 2); ok {
					game.UseItem(s.hWnd, x, y)
					if s.isInventoryStillOpened(px, py) {
						s.logH("failed at double clicking")
						s.setFailureState(Human)
					} else if s.attack(s.isHumanActionSuccessful) {
						s.logH("throwed a bomb")
						s.setSuccessState(Human)
					} else {
						s.logH("missed a hit")
						s.setFailureState(Human)
					}

				} else {
					s.logH("cannot find a bomb")
					s.setFailureState(Human)
				}
			} else {
				s.logH("cannot find the position of window")
				s.setFailureState(Human)
			}
		case HumanPotion:
			game.CloseAllWindows(s.hWnd)
			game.ClearChat(s.hWnd)
			ratio, _ := strconv.ParseFloat(s.HumanActions[s.currentHumanActionId].Param, 32)
			if target, ok := s.searchHealthLowerThan(float32(ratio)); ok {
				game.OpenWindow(s.hWnd, game.KEY_INVENTORY)
				if px, py, isPivotFound := s.getInventoryPos(); isPivotFound {
					if x, y, ok := game.GetItemPos(s.hWnd, px, py, game.COLOR_ITEM_POTION, 3); ok {
						game.UseItem(s.hWnd, x, y)
						if s.isInventoryStillOpened(px, py) {
							s.logH("failed at double clicking")
							s.setFailureState(Human)
						} else if s.aim(target, s.isHumanActionSuccessful) {
							s.logH("healed an ally")
							s.setSuccessState(Human)
						} else {
							s.logH("cannot target")
							s.setFailureState(Human)
						}
					} else {
						s.logH("cannot find a potion")
					}
				} else {
					s.logH("cannot find the position of window")
					s.setFailureState(Human)
				}
			} else {
				s.logH("found all good")
			}
		case HumanRide:
			game.OpenWindow(s.hWnd, game.KEY_SKILL)
			if x, y, ok := s.getSkillWindowPos(); ok {
				offset := int(s.HumanActions[s.currentHumanActionId].Offset)
				level := int(s.HumanActions[s.currentHumanActionId].Level)
				game.UseHumanSkill(s.hWnd, x, y, offset, level)
				if s.didHumanMissSkill(x, y) {
					s.logH("missed the skill button or is out of mana")
				} else {
					s.logH("is tring to get on a pet")
					for i, v := range s.PetActions {
						if v.Action == PetRide {
							s.currentPetActionId = i
							break
						}
					}
					s.setSuccessState(Human)
				}
			} else {
				s.logH("cannot find the position of window")
				s.setFailureState(Human)
			}
		case HumanBloodMagic:
			game.CloseAllWindows(s.hWnd)
			game.ClearChat(s.hWnd)
			if self, ok := s.getSelfTarget(true); ok {
				ratio, _ := strconv.ParseFloat(s.HumanActions[s.currentHumanActionId].Param, 32)

				if !s.isHealthLowerThan(float32(ratio), self) {
					s.logH("is a healthy boy")
					break
				}

				game.OpenWindow(s.hWnd, game.KEY_SKILL)
				if x, y, ok := s.getSkillWindowPos(); ok {
					offset := int(s.HumanActions[s.currentHumanActionId].Offset)
					level := int(s.HumanActions[s.currentHumanActionId].Level)
					game.UseHumanSkill(s.hWnd, x, y, offset, level)
					if s.didHumanMissSkill(x, y) {
						s.logH("missed the skill button or is out of mana")
					} else if s.isHumanActionSuccessful() {
						s.logH("used a skill")
						s.setSuccessState(Human)
					} else if s.attack(s.isHumanActionSuccessful) {
						s.logH("used a skill")
						s.setSuccessState(Human)
					} else {
						s.logH("missed a hit")
						s.setFailureState(Human)
					}
				} else {
					s.logH("cannot find the position of window")
					s.setFailureState(Human)
				}

			} else {
				s.logH("cannot find self")
			}
		case HumanHealSelf:
			game.CloseAllWindows(s.hWnd)
			game.ClearChat(s.hWnd)
			if self, ok := s.getSelfTarget(true); ok {
				ratio, _ := strconv.ParseFloat(s.HumanActions[s.currentHumanActionId].Param, 32)

				if !s.isHealthLowerThan(float32(ratio), self) {
					s.logH("is a healthy boy")
					break
				}

				game.OpenWindow(s.hWnd, game.KEY_SKILL)
				if x, y, ok := s.getSkillWindowPos(); ok {
					offset := int(s.HumanActions[s.currentHumanActionId].Offset)
					level := int(s.HumanActions[s.currentHumanActionId].Level)
					game.UseHumanSkill(s.hWnd, x, y, offset, level)
					if s.didHumanMissSkill(x, y) {
						s.logH("missed the skill button or is out of mana")
					} else {
						s.logH("healed self")
						s.setSuccessState(Human)
					}
				} else {
					s.logH("cannot find the position of window")
					s.setFailureState(Human)
				}

			} else {
				s.logH("cannot find self")
			}
		case HumanHealOne:
			game.CloseAllWindows(s.hWnd)
			game.ClearChat(s.hWnd)
			ratio, _ := strconv.ParseFloat(s.HumanActions[s.currentHumanActionId].Param, 32)
			if target, ok := s.searchHealthLowerThan(float32(ratio)); ok {
				game.OpenWindow(s.hWnd, game.KEY_SKILL)
				if x, y, ok := s.getSkillWindowPos(); ok {
					offset := int(s.HumanActions[s.currentHumanActionId].Offset)
					level := int(s.HumanActions[s.currentHumanActionId].Level)
					game.UseHumanSkill(s.hWnd, x, y, offset, level)
					if s.didHumanMissSkill(x, y) {
						s.logH("missed the skill button or is out of mana")
					} else if s.aim(target, s.isHumanActionSuccessful) {
						s.logH("healed an ally")
						s.setSuccessState(Human)
					} else {
						s.logH("cannot target")
						s.setFailureState(Human)
					}
				} else {
					s.logH("cannot find the position of window")
					s.setFailureState(Human)
				}
			} else {
				s.logH("found all good")
			}
		case HumanHealTShaped:
			game.CloseAllWindows(s.hWnd)
			game.ClearChat(s.hWnd)
			ratio, _ := strconv.ParseFloat(s.HumanActions[s.currentHumanActionId].Param, 32)
			if target, ok := s.searchWeightedTShapedHealthLowerThan(float32(ratio)); ok {
				game.OpenWindow(s.hWnd, game.KEY_SKILL)
				if x, y, ok := s.getSkillWindowPos(); ok {
					offset := int(s.HumanActions[s.currentHumanActionId].Offset)
					level := int(s.HumanActions[s.currentHumanActionId].Level)
					game.UseHumanSkill(s.hWnd, x, y, offset, level)
					if s.didHumanMissSkill(x, y) {
						s.logH("missed the skill button or is out of mana")
					} else if s.aim(target, s.isHumanActionSuccessful) {
						s.logH("healed allies")
						s.setSuccessState(Human)
					} else {
						s.logH("cannot target")
						s.setFailureState(Human)
					}
				} else {
					s.logH("cannot find the position of window")
					s.setFailureState(Human)
				}
			} else {
				s.logH("found all good")
			}
		case HumanHealMulti:
			game.CloseAllWindows(s.hWnd)
			game.ClearChat(s.hWnd)
			ratio, _ := strconv.ParseFloat(s.HumanActions[s.currentHumanActionId].Param, 32)
			count := s.countHealthLowerThan(float32(ratio))
			if count >= 4 {
				game.OpenWindow(s.hWnd, game.KEY_SKILL)
				if x, y, ok := s.getSkillWindowPos(); ok {
					offset := int(s.HumanActions[s.currentHumanActionId].Offset)
					level := int(s.HumanActions[s.currentHumanActionId].Level)
					game.UseHumanSkill(s.hWnd, x, y, offset, level)
					if s.didHumanMissSkill(x, y) {
						s.logH("missed the skill button or is out of mana")
					} else if s.aim(&PLAYER_L_3_H, s.isHumanActionSuccessful) {
						s.logH("healed allies")
						s.setSuccessState(Human)
					} else {
						s.logH("cannot target")
						s.setFailureState(Human)
					}
				} else {
					s.logH("cannot find the position of window")
					s.setFailureState(Human)
				}
			} else {
				s.logH("found all good")
			}
		case HumanRecall:
			game.OpenWindow(s.hWnd, game.KEY_PET)
			if s.canRecall() {
				s.recall()
				s.logH("recalled")
				s.currentControlUnit = Repeat
			} else {
				s.logH("already recalled")
				s.setSuccessState(Human)
			}
		case HumanTrainSkill:
			if self, ok := s.getSelfTarget(false); ok {
				game.OpenWindow(s.hWnd, game.KEY_SKILL)
				if x, y, ok := s.getSkillWindowPos(); ok {
					offset := int(s.HumanActions[s.currentHumanActionId].Offset)
					level := int(s.HumanActions[s.currentHumanActionId].Level)
					game.UseHumanSkill(s.hWnd, x, y, offset, level)
					if s.didHumanMissSkill(x, y) {
						s.logH("missed the skill button or is out of mana")
					} else if s.isHumanActionSuccessful() {
						s.logH("is training")
						s.setSuccessState(Human)
					} else if s.aim(self, s.isHumanActionSuccessful) {
						s.logH("is training")
						s.setSuccessState(Human)
					}
				} else {
					s.logH("cannot find the position of window")
					s.setFailureState(Human)
				}
			} else {
				s.logH("cannot find self")
				s.setFailureState(Human)
			}
		case HumanCatch:
			game.CloseAllWindows(s.hWnd)
			game.ClearChat(s.hWnd)
			if self, ok := s.getSelfTarget(true); ok {
				ratio, _ := strconv.ParseFloat(s.HumanActions[s.currentHumanActionId].Param, 32)
				if s.isHealthLowerThan(float32(ratio), self) {
					s.isOutOfHealth = true
					s.logH("is out of health")
				}
			}
		}

		s.updateCurrentActionId(Human)
	}

	if len(s.HumanActions) != 0 {
		s.currentHumanActionId %= len(s.HumanActions)
	} else {
		s.currentHumanActionId = 0
	}
}

func (s *ActionState) executePetStateMachiine() {

	for s.currentPetActionId < len(s.PetActions) && game.IsBattleScene(s.hWnd) && s.isPetStageStable() && s.Enabled {

		s.resetCurrentControlUnit()
		s.endHumanHanging()

		switch s.PetActions[s.currentPetActionId].Action {
		case PetAttack:
			s.openPetSkillWindow()
			if x, y, ok := s.getSkillWindowPos(); ok {
				game.UsePetSkill(s.hWnd, x, y, 1)
				if s.attack(s.isPetActionSuccessful) {
					s.logP("attacked")
					s.setSuccessState(Pet)
				} else {
					s.logP("missed a hit")
					s.setFailureState(Pet)
				}
			} else {
				s.logP("cannot find the position of window")
				s.setFailureState(Pet)
			}
		case PetEscape:
			if !s.isOnRide() {
				s.logP("cannot escape while off ride")
				break
			}

			s.escape()
			s.logP("escaped")
			s.setFailureState(Pet)
		case PetDefend:
			s.openPetSkillWindow()
			if x, y, ok := s.getSkillWindowPos(); ok {
				offset := int(s.PetActions[s.currentPetActionId].Offset)
				game.UsePetSkill(s.hWnd, x, y, offset)
				if s.didPetMissSkill() || s.didOnRideMissSkill() {
					s.logP("missed the skill button or is out of mana")
				} else {
					s.logP("defended")
					s.setSuccessState(Pet)
				}
			} else {
				s.logP("cannot find the position of window")
				s.setFailureState(Pet)
			}
		case PetSkill:
			s.openPetSkillWindow()
			if x, y, ok := s.getSkillWindowPos(); ok {
				offset := int(s.PetActions[s.currentPetActionId].Offset)
				game.UsePetSkill(s.hWnd, x, y, offset)
				if s.didPetMissSkill() || s.didOnRideMissSkill() {
					s.logP("missed the skill button or is out of mana")
				} else if s.attack(s.isPetActionSuccessful) {
					s.logP("used a skill")
					s.setSuccessState(Pet)
				} else {
					s.logP("missed a hit")
					s.setFailureState(Pet)
				}
			} else {
				s.logP("cannot find the position of window")
				s.setFailureState(Pet)
			}
		case PetHealSelf:
			game.CloseAllWindows(s.hWnd)
			game.ClearChat(s.hWnd)
			if self, ok := s.getSelfTarget(s.isOnRide()); ok {
				ratio, _ := strconv.ParseFloat(s.PetActions[s.currentPetActionId].Param, 32)
				if !s.isHealthLowerThan(float32(ratio), self) {
					s.logP("is a healthy boy")
					break
				}

				s.openPetSkillWindow()
				if x, y, ok := s.getSkillWindowPos(); ok {
					offset := int(s.PetActions[s.currentPetActionId].Offset)
					game.UsePetSkill(s.hWnd, x, y, offset)
					if s.didPetMissSkill() || s.didOnRideMissSkill() {
						s.logP("missed the skill button or is out of mana")
					} else {
						s.logP("healed self")
						s.setSuccessState(Pet)
					}
				} else {
					s.logP("cannot find the position of window")
					s.setFailureState(Pet)
				}
			} else {
				s.logP("cannot find self")
			}
		case PetHealOne:
			game.CloseAllWindows(s.hWnd)
			game.ClearChat(s.hWnd)
			ratio, _ := strconv.ParseFloat(s.PetActions[s.currentPetActionId].Param, 32)
			if target, ok := s.searchHealthLowerThan(float32(ratio)); ok {
				s.openPetSkillWindow()
				if x, y, ok := s.getSkillWindowPos(); ok {
					offset := int(s.PetActions[s.currentPetActionId].Offset)
					game.UsePetSkill(s.hWnd, x, y, offset)
					if s.aim(target, s.isPetActionSuccessful) {
						s.logP("healed an ally")
						s.setSuccessState(Pet)
					} else {
						s.logP("cannot target")
						s.setFailureState(Pet)
					}
				} else {
					s.logP("cannot find the position of window")
					s.setFailureState(Pet)
				}

			} else {
				s.logH("found all good")
			}
		case PetRide:
			if s.isOnRide() {
				s.logP("is on ride")
				s.setSuccessState(Pet)
				break
			}

			s.openPetSkillWindow()
			if x, y, ok := s.getSkillWindowPos(); ok {
				offset := int(s.PetActions[s.currentPetActionId].Offset)
				game.UsePetSkill(s.hWnd, x, y, offset)
				s.logP("tries to get on ride")
				s.currentControlUnit = Repeat
			} else {
				s.logP("cannot find the position of window")
				s.setFailureState(Pet)
			}
		case PetOffRide:
			if !s.isOnRide() {
				s.logP("is off ride")
				s.setSuccessState(Pet)
				break
			}

			s.openPetSkillWindow()
			if x, y, ok := s.getSkillWindowPos(); ok {
				offset := int(s.PetActions[s.currentPetActionId].Offset)
				game.UsePetSkill(s.hWnd, x, y, offset)
				s.logP("tries to get off ride")
				s.currentControlUnit = Repeat
			} else {
				s.logP("cannot find the position of window")
				s.setFailureState(Pet)
			}
		case PetCatch:
			game.CloseAllWindows(s.hWnd)
			game.ClearChat(s.hWnd)
			if self, ok := s.getSelfTarget(s.isOnRide()); ok {
				ratio, _ := strconv.ParseFloat(s.PetActions[s.currentPetActionId].Param, 32)
				if s.isHealthLowerThan(float32(ratio), self) {
					s.isOutOfHealth = true
					s.logP("is out of health")
				}
			}
		case PetHang:
			s.logP("is hanging")
			s.isPetHanging = true
			s.currentControlUnit = Repeat
		}

		s.updateCurrentActionId(Pet)
	}

	if len(s.PetActions) != 0 {
		s.currentPetActionId %= len(s.PetActions)
	} else {
		s.currentPetActionId = 0
	}
}

func (s *ActionState) reset() {
	s.currentHumanActionId = 0
	s.currentPetActionId = 0
	s.trainingCounter = 0
}

func (s *ActionState) resetCurrentControlUnit() {
	s.currentControlUnit = Undefined
	s.currentJumpId = 0
}

func (s *ActionState) setSuccessState(r Role) {
	switch r {
	case Human:
		s.currentControlUnit = s.HumanActions[s.currentHumanActionId].SuccessControlUnit
		s.currentJumpId = s.HumanActions[s.currentHumanActionId].SuccessJumpId
	case Pet:
		s.currentControlUnit = s.PetActions[s.currentPetActionId].SuccessControlUnit
		s.currentJumpId = s.PetActions[s.currentPetActionId].SuccessJumpId
	}
}

func (s *ActionState) setFailureState(r Role) {
	switch r {
	case Human:
		s.currentControlUnit = s.HumanActions[s.currentHumanActionId].FailureControlUnit
		s.currentJumpId = s.HumanActions[s.currentHumanActionId].FailureJumpId
	case Pet:
		s.currentControlUnit = s.PetActions[s.currentPetActionId].FailureControlUnit
		s.currentJumpId = s.PetActions[s.currentPetActionId].FailureJumpId
	}
}

func (s *ActionState) updateCurrentActionId(r Role) {
	switch r {
	case Human:
		switch s.currentControlUnit {
		case StartOver:
			s.currentHumanActionId = 0
		case Continue:
			s.currentHumanActionId++
		case Repeat:
			return
		case Jump:
			s.currentHumanActionId = s.currentJumpId
		default:
			s.currentHumanActionId++
		}
	case Pet:
		switch s.currentControlUnit {
		case StartOver:
			s.currentPetActionId = 0
		case Continue:
			s.currentPetActionId++
		case Repeat:
			return
		case Jump:
			s.currentPetActionId = s.currentJumpId
		default:
			s.currentPetActionId++
		}
	}

	s.currentControlUnit = Undefined
	s.currentJumpId = 0
}

func (s ActionState) isManaChecker() bool {
	return *s.ManaChecker == fmt.Sprint(s.hWnd)
}

func (s *ActionState) enableBattleCommandAttack() {
	if !s.isBattleCommandEnable(BATTLE_COMMAND_ATTACK) {
		internal.LeftClick(s.hWnd, BATTLE_COMMAND_ATTACK.X, BATTLE_COMMAND_ATTACK.Y)
		time.Sleep(DURATION_BATTLE_ACTION_GENERAL)
	}
}

func (s *ActionState) attack(stateChecker func() bool) bool {
	targets := make([]game.CheckTarget, len(s.enemies))
	copy(targets, s.enemies)
	rand.New(rand.NewSource(time.Now().UnixNano()))
	rand.Shuffle(len(targets), func(i, j int) { targets[i], targets[j] = targets[j], targets[i] })
	for _, target := range targets {
		internal.LeftClick(s.hWnd, target.X, target.Y)
		time.Sleep(DURATION_BATTLE_ACTION_ATTACK)
		if stateChecker() {
			return true
		}
	}
	return false
}

func (s *ActionState) aim(target *game.CheckTarget, stateChecker func() bool) bool {
	internal.LeftClick(s.hWnd, target.X+15, target.Y-22)
	time.Sleep(DURATION_BATTLE_ACTION_ATTACK)
	return stateChecker()
}

func (s *ActionState) defend() {
	internal.LeftClick(s.hWnd, BATTLE_COMMAND_DEFENCE.X, BATTLE_COMMAND_DEFENCE.Y)
	time.Sleep(DURATION_BATTLE_ACTION_GENERAL)
}

func (s *ActionState) escape() {
	internal.LeftClick(s.hWnd, BATTLE_COMMAND_ESCAPE.X, BATTLE_COMMAND_ESCAPE.Y)
	time.Sleep(DURATION_BATTLE_ACTION_GENERAL)
}

func (s *ActionState) move() {
	internal.LeftClick(s.hWnd, BATTLE_COMMAND_MOVE.X, BATTLE_COMMAND_MOVE.Y)
	time.Sleep(DURATION_BATTLE_ACTION_GENERAL)
}

func (s *ActionState) recall() {
	internal.LeftClick(s.hWnd, BATTLE_WINDOW_PET_RECALL_BUTTON.X, BATTLE_WINDOW_PET_RECALL_BUTTON.Y)
	time.Sleep(DURATION_BATTLE_ACTION_GENERAL)
}

func (s *ActionState) openPetSkillWindow() {
	if s.isPetSkillWindowOpened() {
		game.ResetAllWindows(s.hWnd)
	} else {
		game.CloseAllWindows(s.hWnd)
		internal.RightClick(s.hWnd, game.GAME_WIDTH/2, 28)
		time.Sleep(DURATION_BATTLE_ACTION_GENERAL)
		game.ResetAllWindows(s.hWnd)
	}
}

func (s *ActionState) openSkillWindowWithMouse() {
	internal.LeftClick(s.hWnd, BATTLE_COMMAND_SKILL.X, BATTLE_COMMAND_SKILL.Y)
	time.Sleep(DURATION_BATTLE_ACTION_GENERAL)
}

func (s *ActionState) logH(message string) {
	header := fmt.Sprintf("[%s][%d][%s]", fmt.Sprint(s.hWnd), s.currentHumanActionId, strings.Trim(s.HumanActions[s.currentHumanActionId].Action.String(), "*"))
	log.Printf("%-26s %s",
		header,
		message,
	)
}

func (s *ActionState) logP(message string) {
	header := fmt.Sprintf("[%s][%d][%s]", fmt.Sprint(s.hWnd), s.currentPetActionId, strings.Trim(s.PetActions[s.currentPetActionId].Action.String(), "*"))
	log.Printf("%-26s %s",
		header,
		message,
	)
}

func CreateNewBattleActionState(hWnd win.HWND, gameDir, manaChecker *string) ActionState {
	return ActionState{
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

func (s *ActionState) GetHWND() win.HWND {
	return s.hWnd
}

func (s *ActionState) SetHWND(hWnd win.HWND) {
	s.hWnd = hWnd
}

func (s *ActionState) AddHumanAction(action humanAction) {
	newHumanState := HumanAction{
		Action: action,
	}
	s.HumanActions = append(s.HumanActions, newHumanState)
}

func (s *ActionState) AddHumanSkillLevel(level Offset) {
	s.HumanActions[len(s.HumanActions)-1].Level = level
}

func (s *ActionState) AddPetAction(action petAction) {
	newPetState := PetAction{
		Action: action,
	}
	s.PetActions = append(s.PetActions, newPetState)
}

func (s *ActionState) AddSkillOffset(r Role, offset Offset) {
	switch r {
	case Human:
		s.HumanActions[len(s.HumanActions)-1].Offset = offset
	case Pet:
		s.PetActions[len(s.PetActions)-1].Offset = offset
	}
}

func (s *ActionState) AddThreshold(r Role, threshold Threshold) {
	s.HumanActions[len(s.HumanActions)-1].Threshold = threshold
}

func (s *ActionState) AddParam(r Role, param string) {
	switch r {
	case Human:
		s.HumanActions[len(s.HumanActions)-1].Param = param
	case Pet:
		s.PetActions[len(s.PetActions)-1].Param = param
	}
}

func (s *ActionState) AddSuccessControlUnit(r Role, cu ControlUnit) {
	switch r {
	case Human:
		s.HumanActions[len(s.HumanActions)-1].SuccessControlUnit = cu
	case Pet:
		s.PetActions[len(s.PetActions)-1].SuccessControlUnit = cu
	}
}
func (s *ActionState) AddSuccessJumpId(r Role, jumpId int) {
	switch r {
	case Human:
		s.HumanActions[len(s.HumanActions)-1].SuccessJumpId = jumpId
	case Pet:
		s.PetActions[len(s.PetActions)-1].SuccessJumpId = jumpId
	}
}

func (s *ActionState) AddFailureControlUnit(r Role, cu ControlUnit) {
	switch r {
	case Human:
		s.HumanActions[len(s.HumanActions)-1].FailureControlUnit = cu
	case Pet:
		s.PetActions[len(s.PetActions)-1].FailureControlUnit = cu
	}
}

func (s *ActionState) AddFailureJumpId(r Role, jumpId int) {
	switch r {
	case Human:
		s.HumanActions[len(s.HumanActions)-1].FailureJumpId = jumpId
	case Pet:
		s.PetActions[len(s.PetActions)-1].FailureJumpId = jumpId
	}
}

func (s *ActionState) ClearHumanActions() {
	s.HumanActions = s.HumanActions[:0]
}

func (s *ActionState) ClearPetActions() {
	s.PetActions = s.PetActions[:0]
}

func (s *ActionState) GetHumanActions() []HumanAction {
	return s.HumanActions
}

func (s *ActionState) GetPetActions() []PetAction {
	return s.PetActions
}

func (s *ActionState) detectEnemies() {

	if !game.IsBattleScene(s.hWnd) || !s.Enabled {
		return
	}

	if s.isInventoryStuck() {
		s.openSkillWindowWithMouse()
	}
	game.CloseAllWindows(s.hWnd)

	if s.trainingCounter < TRAINING_COUNTER_THRESHOLD {
		s.enemies = s.getEnemies(allMonsters)
	}

	if len(s.enemies) == 1 && s.trainingCounter < TRAINING_COUNTER_THRESHOLD {
		s.trainingCounter++
	}

	log.Printf("# Handle %s detected %d enemies\n", fmt.Sprint(s.hWnd), len(s.enemies))
}

func (s *ActionState) endHumanHanging() {
	if s.isHumanHanging {
		s.isHumanHanging = false
		cu := s.HumanActions[s.currentHumanActionId].SuccessControlUnit
		jumpId := s.HumanActions[s.currentHumanActionId].SuccessJumpId

		switch cu {
		case StartOver:
			s.currentHumanActionId = 0
		case Continue:
			s.currentHumanActionId++
		case Jump:
			s.currentHumanActionId = jumpId
		}
	}
}
func (s *ActionState) endPetHanging() {
	if s.isPetHanging {
		s.isPetHanging = false
		cu := s.PetActions[s.currentPetActionId].SuccessControlUnit
		jumpId := s.PetActions[s.currentPetActionId].SuccessJumpId

		switch cu {
		case StartOver:
			s.currentPetActionId = 0
		case Continue:
			s.currentPetActionId++
		case Jump:
			s.currentPetActionId = jumpId
		}
	}
}

func (s *ActionState) checkCharacterMana() {

	if !game.IsBattleScene(s.hWnd) || !s.Enabled {
		return
	}

	if s.isManaChecker() {
		s.logH("checks characters' mana")

		game.CloseAllWindows(s.hWnd)
		game.ClearChat(s.hWnd)
		if s.isOutOfMana = s.isAnyCharacterOutOfMana(); s.isOutOfMana {
			s.logH("someone is out of mana")
		}
	}
}

func (s *ActionState) wait() {

	for game.IsBattleScene(s.hWnd) && s.Enabled && !s.isHumanStageStable() && !s.isPetStageStable() {
		time.Sleep(DURATION_BATTLE_ACTION_WAITING_LOOP)
	}
}

func TestAction(hWnd win.HWND) (x int32, y int32, successful bool) {
	game.CloseAllWindows(hWnd)
	game.ClearChat(hWnd)
	game.OpenWindow(hWnd, game.KEY_INVENTORY)
	return game.GetItemPos(hWnd, x, y, 16448250, 3)
}

package battle

import (
	"cg/game"
	"cg/game/enum"
	"cg/game/enum/character"
	"cg/game/enum/controlunit"
	"cg/game/enum/enemy"
	"cg/game/enum/offset"
	"cg/game/enum/pet"
	"cg/game/enum/ratio"
	"cg/game/enum/role"
	"cg/game/enum/threshold"
	"cg/game/items"
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

var (
	ControlUnits   = enum.GenericEnum[controlunit.ControlUnit]{List: []controlunit.ControlUnit{controlunit.StartOver, controlunit.Continue, controlunit.Repeat, controlunit.Jump}}
	Offsets        = enum.GenericEnum[offset.Offset]{List: []offset.Offset{offset.One, offset.Two, offset.Three, offset.Four, offset.Five, offset.Six, offset.Seven, offset.Eight, offset.Nine, offset.Ten}}
	Levels         = enum.GenericEnum[offset.Offset]{List: []offset.Offset{offset.One, offset.Two, offset.Three, offset.Four, offset.Five, offset.Six, offset.Seven, offset.Eight, offset.Nine, offset.Ten}}
	Ratios         = enum.GenericEnum[ratio.Ratio]{List: []ratio.Ratio{ratio.OneTenth, ratio.TwoTenth, ratio.ThreeTenth, ratio.FourTenth, ratio.FiveTenth, ratio.SixTenth, ratio.SevenTenth, ratio.EightTenth, ratio.NineTenth, ratio.One}}
	Thresholds     = enum.GenericEnum[threshold.Threshold]{List: []threshold.Threshold{threshold.OneFoe, threshold.TwoFoes, threshold.ThreeFoes, threshold.FourFoes, threshold.FiveFoes, threshold.SixFoes, threshold.SevenFoes, threshold.EightFoes, threshold.NineFoes, threshold.TenFoes}}
	EnemyPositions = enum.GenericEnum[enemy.Position]{List: []enemy.Position{enemy.T1, enemy.T2, enemy.T3, enemy.T4, enemy.T5, enemy.B1, enemy.B2, enemy.B3, enemy.B4, enemy.B5}}
)

const (
	DURATION_BATTLE_ACTION_WAITING_LOOP = 100 * time.Millisecond
	DURATION_BATTLE_ACTION_ATTACK       = 100 * time.Millisecond
	DURATION_BATTLE_ACTION_GENERAL      = 160 * time.Millisecond
)

type CharacterAction struct {
	Action             character.Action
	Offset             offset.Offset
	Level              offset.Offset
	Threshold          threshold.Threshold
	Param              string
	SuccessControlUnit controlunit.ControlUnit
	FailureControlUnit controlunit.ControlUnit
	SuccessJumpId      int
	FailureJumpId      int
}

type PetAction struct {
	Action             pet.Action
	Offset             offset.Offset
	Threshold          threshold.Threshold
	Param              string
	SuccessControlUnit controlunit.ControlUnit
	FailureControlUnit controlunit.ControlUnit
	SuccessJumpId      int
	FailureJumpId      int
}

type ActionState struct {
	hWnd win.HWND `json:"-"`

	CharacterActions         []CharacterAction
	PetActions               []PetAction
	currentCharacterActionId int                     `json:"-"`
	currentPetActionId       int                     `json:"-"`
	currentCU                controlunit.ControlUnit `json:"-"`
	currentJumpId            int                     `json:"-"`

	Enabled                bool               `json:"-"`
	ActivityCheckerEnabled bool               `json:"-"`
	EnemyOrder             enemy.Position     `json:"-"`
	CustomEnemies          []game.CheckTarget `json:"-"`

	isOutOfHealth      bool `json:"-"`
	isOutOfMana        bool `json:"-"`
	isCharacterHanging bool `json:"-"`
	isPetHanging       bool `json:"-"`

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
		s.executeCharacterStateMachine()
		s.wait()
		s.executeCharacterStateMachine()
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

func (s *ActionState) executeCharacterStateMachine() {

	for s.currentCharacterActionId < len(s.CharacterActions) && game.IsBattleScene(s.hWnd) && s.isCharacterStageStable() && s.Enabled {

		s.resetCurrentControlUnit()
		s.endPetHanging()

		switch s.CharacterActions[s.currentCharacterActionId].Action {
		case character.Attack:
			s.enableBattleCommandAttack()
			if s.attack(s.isCharacterActionSuccessful) {
				s.logH("attacked")
				s.setSuccessState(role.Character)
			} else {
				s.logH("missed a hit")
				s.setFailureState(role.Character)
			}
		case character.Skill:
			game.OpenWindow(s.hWnd, game.KEY_SKILL)
			if x, y, ok := s.getSkillWindowPos(); ok {
				offset := int(s.CharacterActions[s.currentCharacterActionId].Offset)
				level := int(s.CharacterActions[s.currentCharacterActionId].Level)
				game.UseCharacterSkill(s.hWnd, x, y, offset, level)
				if s.didCharacterMissSkill(x, y) {
					s.logH("missed the skill button or is out of mana")
				} else if s.isCharacterActionSuccessful() {
					s.logH("used a skill")
					s.setSuccessState(role.Character)
				} else if s.attack(s.isCharacterActionSuccessful) {
					s.logH("used a skill")
					s.setSuccessState(role.Character)
				} else {
					s.logH("missed a hit")
					s.setFailureState(role.Character)
				}
			} else {
				s.logH("cannot find the position of window")
				s.setFailureState(role.Character)
			}

		case character.ThresholdSkill:
			threshold, _ := strconv.Atoi(strings.Split(string(s.CharacterActions[s.currentCharacterActionId].Threshold), " ")[0])
			if len(s.enemies) < threshold {
				s.logH("performs next action due to too few enemies")
				break
			}

			game.OpenWindow(s.hWnd, game.KEY_SKILL)
			if x, y, ok := s.getSkillWindowPos(); ok {
				offset := int(s.CharacterActions[s.currentCharacterActionId].Offset)
				level := int(s.CharacterActions[s.currentCharacterActionId].Level)
				game.UseCharacterSkill(s.hWnd, x, y, offset, level)
				if s.didCharacterMissSkill(x, y) {
					s.logH("missed the skill button or is out of mana")
				} else if s.isCharacterActionSuccessful() {
					s.logH("used a skill")
					s.setSuccessState(role.Character)
				} else if s.attack(s.isCharacterActionSuccessful) {
					s.logH("used a skill")
					s.setSuccessState(role.Character)
				} else {
					s.logH("missed a hit")
					s.setFailureState(role.Character)
				}
			} else {
				s.logH("cannot find the position of window")
				s.setFailureState(role.Character)
			}
		case character.Defend:
			s.defend()
			s.logH("defended")
			s.setSuccessState(role.Character)
		case character.Move:
			s.move()
			s.logH("moved")
			s.setSuccessState(role.Character)
		case character.Escape:
			s.escape()
			s.logH("escaped")
			s.setFailureState(role.Character)
		case character.Hang:
			s.logH("is hanging")
			s.isCharacterHanging = true
			s.currentCU = controlunit.Repeat
		case character.Bomb:
			var bomb items.Item
			for i := range items.Bombs.List {
				if items.Bombs.List[i].Name == s.CharacterActions[s.currentCharacterActionId].Param {
					bomb = items.Bombs.List[i]
				}
			}

			if bomb == items.I_B_9A && len(s.enemies) < 4 {
				s.logH("performs next action due to too few enemies")
				break
			}

			game.OpenWindow(s.hWnd, game.KEY_INVENTORY)
			if px, py, isPivotFound := s.getInventoryPos(); isPivotFound {
				if x, y, ok := game.GetItemPos(s.hWnd, px, py, bomb.Color, 2); ok {
					game.UseItem(s.hWnd, x, y)
					if s.isInventoryStillOpened(px, py) {
						s.logH("failed at double clicking")
						s.setFailureState(role.Character)
					} else if s.attack(s.isCharacterActionSuccessful) {
						s.logH("throwed a bomb")
						s.setSuccessState(role.Character)
					} else {
						s.logH("missed a hit")
						s.setFailureState(role.Character)
					}

				} else {
					s.logH("cannot find a bomb")
					s.setFailureState(role.Character)
				}
			} else {
				s.logH("cannot find the position of window")
				s.setFailureState(role.Character)
			}
		case character.Potion:
			game.CloseAllWindows(s.hWnd)
			game.ClearChat(s.hWnd)
			ratio, _ := strconv.ParseFloat(s.CharacterActions[s.currentCharacterActionId].Param, 32)
			if target, ok := s.searchHealthLowerThan(float32(ratio)); ok {
				game.OpenWindow(s.hWnd, game.KEY_INVENTORY)
				if px, py, isPivotFound := s.getInventoryPos(); isPivotFound {
					if x, y, ok := game.GetItemPos(s.hWnd, px, py, items.COLOR_ITEM_POTION, 3); ok {
						game.UseItem(s.hWnd, x, y)
						if s.isInventoryStillOpened(px, py) {
							s.logH("failed at double clicking")
							s.setFailureState(role.Character)
						} else if s.aim(target, s.isCharacterActionSuccessful) {
							s.logH("healed an ally")
							s.setSuccessState(role.Character)
						} else {
							s.logH("cannot target")
							s.setFailureState(role.Character)
						}
					} else {
						s.logH("cannot find a potion")
					}
				} else {
					s.logH("cannot find the position of window")
					s.setFailureState(role.Character)
				}
			} else {
				s.logH("found all good")
			}
		case character.Ride:
			game.OpenWindow(s.hWnd, game.KEY_SKILL)
			if x, y, ok := s.getSkillWindowPos(); ok {
				offset := int(s.CharacterActions[s.currentCharacterActionId].Offset)
				level := int(s.CharacterActions[s.currentCharacterActionId].Level)
				game.UseCharacterSkill(s.hWnd, x, y, offset, level)
				if s.didCharacterMissSkill(x, y) {
					s.logH("missed the skill button or is out of mana")
				} else {
					s.logH("is tring to get on a pet")
					for i, v := range s.PetActions {
						if v.Action == pet.Ride {
							s.currentPetActionId = i
							break
						}
					}
					s.setSuccessState(role.Character)
				}
			} else {
				s.logH("cannot find the position of window")
				s.setFailureState(role.Character)
			}
		case character.BloodMagic:
			game.CloseAllWindows(s.hWnd)
			game.ClearChat(s.hWnd)
			if self, ok := s.getSelfTarget(true); ok {
				ratio, _ := strconv.ParseFloat(s.CharacterActions[s.currentCharacterActionId].Param, 32)

				if !s.isHealthLowerThan(float32(ratio), self) {
					s.logH("is a healthy boy")
					break
				}

				game.OpenWindow(s.hWnd, game.KEY_SKILL)
				if x, y, ok := s.getSkillWindowPos(); ok {
					offset := int(s.CharacterActions[s.currentCharacterActionId].Offset)
					level := int(s.CharacterActions[s.currentCharacterActionId].Level)
					game.UseCharacterSkill(s.hWnd, x, y, offset, level)
					if s.didCharacterMissSkill(x, y) {
						s.logH("missed the skill button or is out of mana")
					} else if s.isCharacterActionSuccessful() {
						s.logH("used a skill")
						s.setSuccessState(role.Character)
					} else if s.attack(s.isCharacterActionSuccessful) {
						s.logH("used a skill")
						s.setSuccessState(role.Character)
					} else {
						s.logH("missed a hit")
						s.setFailureState(role.Character)
					}
				} else {
					s.logH("cannot find the position of window")
					s.setFailureState(role.Character)
				}

			} else {
				s.logH("cannot find self")
			}
		case character.HealSelf:
			game.CloseAllWindows(s.hWnd)
			game.ClearChat(s.hWnd)
			if self, ok := s.getSelfTarget(true); ok {
				ratio, _ := strconv.ParseFloat(s.CharacterActions[s.currentCharacterActionId].Param, 32)

				if !s.isHealthLowerThan(float32(ratio), self) {
					s.logH("is a healthy boy")
					break
				}

				game.OpenWindow(s.hWnd, game.KEY_SKILL)
				if x, y, ok := s.getSkillWindowPos(); ok {
					offset := int(s.CharacterActions[s.currentCharacterActionId].Offset)
					level := int(s.CharacterActions[s.currentCharacterActionId].Level)
					game.UseCharacterSkill(s.hWnd, x, y, offset, level)
					if s.didCharacterMissSkill(x, y) {
						s.logH("missed the skill button or is out of mana")
					} else {
						s.logH("healed self")
						s.setSuccessState(role.Character)
					}
				} else {
					s.logH("cannot find the position of window")
					s.setFailureState(role.Character)
				}

			} else {
				s.logH("cannot find self")
			}
		case character.HealOne:
			game.CloseAllWindows(s.hWnd)
			game.ClearChat(s.hWnd)
			ratio, _ := strconv.ParseFloat(s.CharacterActions[s.currentCharacterActionId].Param, 32)
			if target, ok := s.searchHealthLowerThan(float32(ratio)); ok {
				game.OpenWindow(s.hWnd, game.KEY_SKILL)
				if x, y, ok := s.getSkillWindowPos(); ok {
					offset := int(s.CharacterActions[s.currentCharacterActionId].Offset)
					level := int(s.CharacterActions[s.currentCharacterActionId].Level)
					game.UseCharacterSkill(s.hWnd, x, y, offset, level)
					if s.didCharacterMissSkill(x, y) {
						s.logH("missed the skill button or is out of mana")
					} else if s.aim(target, s.isCharacterActionSuccessful) {
						s.logH("healed an ally")
						s.setSuccessState(role.Character)
					} else {
						s.logH("cannot target")
						s.setFailureState(role.Character)
					}
				} else {
					s.logH("cannot find the position of window")
					s.setFailureState(role.Character)
				}
			} else {
				s.logH("found all good")
			}
		case character.HealTShaped:
			game.CloseAllWindows(s.hWnd)
			game.ClearChat(s.hWnd)
			ratio, _ := strconv.ParseFloat(s.CharacterActions[s.currentCharacterActionId].Param, 32)
			if target, ok := s.searchWeightedTShapedHealthLowerThan(float32(ratio)); ok {
				game.OpenWindow(s.hWnd, game.KEY_SKILL)
				if x, y, ok := s.getSkillWindowPos(); ok {
					offset := int(s.CharacterActions[s.currentCharacterActionId].Offset)
					level := int(s.CharacterActions[s.currentCharacterActionId].Level)
					game.UseCharacterSkill(s.hWnd, x, y, offset, level)
					if s.didCharacterMissSkill(x, y) {
						s.logH("missed the skill button or is out of mana")
					} else if s.aim(target, s.isCharacterActionSuccessful) {
						s.logH("healed allies")
						s.setSuccessState(role.Character)
					} else {
						s.logH("cannot target")
						s.setFailureState(role.Character)
					}
				} else {
					s.logH("cannot find the position of window")
					s.setFailureState(role.Character)
				}
			} else {
				s.logH("found all good")
			}
		case character.HealMulti:
			game.CloseAllWindows(s.hWnd)
			game.ClearChat(s.hWnd)
			ratio, _ := strconv.ParseFloat(s.CharacterActions[s.currentCharacterActionId].Param, 32)
			count := s.countHealthLowerThan(float32(ratio))
			if count >= 4 {
				game.OpenWindow(s.hWnd, game.KEY_SKILL)
				if x, y, ok := s.getSkillWindowPos(); ok {
					offset := int(s.CharacterActions[s.currentCharacterActionId].Offset)
					level := int(s.CharacterActions[s.currentCharacterActionId].Level)
					game.UseCharacterSkill(s.hWnd, x, y, offset, level)
					if s.didCharacterMissSkill(x, y) {
						s.logH("missed the skill button or is out of mana")
					} else if s.aim(&PLAYER_L_3_C, s.isCharacterActionSuccessful) {
						s.logH("healed allies")
						s.setSuccessState(role.Character)
					} else {
						s.logH("cannot target")
						s.setFailureState(role.Character)
					}
				} else {
					s.logH("cannot find the position of window")
					s.setFailureState(role.Character)
				}
			} else {
				s.logH("found all good")
			}
		case character.Recall:
			game.OpenWindow(s.hWnd, game.KEY_PET)
			if s.canRecall() {
				s.recall()
				s.logH("recalled")
				s.currentCU = controlunit.Repeat
			} else {
				s.logH("already recalled")
				s.setSuccessState(role.Character)
			}
		case character.TrainSkill:
			game.OpenWindow(s.hWnd, game.KEY_SKILL)
			if x, y, ok := s.getSkillWindowPos(); ok {
				offset := int(s.CharacterActions[s.currentCharacterActionId].Offset)
				level := int(s.CharacterActions[s.currentCharacterActionId].Level)
				game.UseCharacterSkill(s.hWnd, x, y, offset, level)
				if s.didCharacterMissSkill(x, y) {
					s.logH("missed the skill button or is out of mana")
				} else if s.isCharacterActionSuccessful() {
					s.logH("is training")
					s.setSuccessState(role.Character)
				} else if s.aim(&PLAYER_L_3_C, s.isCharacterActionSuccessful) {
					s.logH("is training")
					s.setSuccessState(role.Character)
				}
			} else {
				s.logH("cannot find the position of window")
				s.setFailureState(role.Character)
			}
		case character.Catch:
			game.CloseAllWindows(s.hWnd)
			game.ClearChat(s.hWnd)
			if self, ok := s.getSelfTarget(true); ok {
				ratio, _ := strconv.ParseFloat(s.CharacterActions[s.currentCharacterActionId].Param, 32)
				if s.isHealthLowerThan(float32(ratio), self) {
					s.isOutOfHealth = true
					s.logH("is out of health")
				}
			}
		}

		s.updateCurrentActionId(role.Character)
	}

	if len(s.CharacterActions) != 0 {
		s.currentCharacterActionId %= len(s.CharacterActions)
	} else {
		s.currentCharacterActionId = 0
	}
}

func (s *ActionState) executePetStateMachiine() {

	for s.currentPetActionId < len(s.PetActions) && game.IsBattleScene(s.hWnd) && s.isPetStageStable() && s.Enabled {

		s.resetCurrentControlUnit()
		s.endCharacterHanging()

		switch s.PetActions[s.currentPetActionId].Action {
		case pet.Attack:
			s.openPetSkillWindow()
			if x, y, ok := s.getSkillWindowPos(); ok {
				game.UsePetSkill(s.hWnd, x, y, 1)
				if s.attack(s.isPetActionSuccessful) {
					s.logP("attacked")
					s.setSuccessState(role.Pet)
				} else {
					s.logP("missed a hit")
					s.setFailureState(role.Pet)
				}
			} else {
				s.logP("cannot find the position of window")
				s.setFailureState(role.Pet)
			}
		case pet.Escape:
			if !s.isOnRide() {
				s.logP("cannot escape while off ride")
				break
			}

			s.escape()
			s.logP("escaped")
			s.setFailureState(role.Pet)
		case pet.Defend:
			s.openPetSkillWindow()
			if x, y, ok := s.getSkillWindowPos(); ok {
				offset := int(s.PetActions[s.currentPetActionId].Offset)
				game.UsePetSkill(s.hWnd, x, y, offset)
				if s.didPetMissSkill() || s.didOnRideMissSkill() {
					s.logP("missed the skill button or is out of mana")
				} else {
					s.logP("defended")
					s.setSuccessState(role.Pet)
				}
			} else {
				s.logP("cannot find the position of window")
				s.setFailureState(role.Pet)
			}
		case pet.Skill:
			s.openPetSkillWindow()
			if x, y, ok := s.getSkillWindowPos(); ok {
				offset := int(s.PetActions[s.currentPetActionId].Offset)
				game.UsePetSkill(s.hWnd, x, y, offset)
				if s.didPetMissSkill() || s.didOnRideMissSkill() {
					s.logP("missed the skill button or is out of mana")
				} else if s.attack(s.isPetActionSuccessful) {
					s.logP("used a skill")
					s.setSuccessState(role.Pet)
				} else {
					s.logP("missed a hit")
					s.setFailureState(role.Pet)
				}
			} else {
				s.logP("cannot find the position of window")
				s.setFailureState(role.Pet)
			}
		case pet.HealSelf:
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
						s.setSuccessState(role.Pet)
					}
				} else {
					s.logP("cannot find the position of window")
					s.setFailureState(role.Pet)
				}
			} else {
				s.logP("cannot find self")
			}
		case pet.HealOne:
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
						s.setSuccessState(role.Pet)
					} else {
						s.logP("cannot target")
						s.setFailureState(role.Pet)
					}
				} else {
					s.logP("cannot find the position of window")
					s.setFailureState(role.Pet)
				}

			} else {
				s.logH("found all good")
			}
		case pet.Ride:
			if s.isOnRide() {
				s.logP("is on ride")
				s.setSuccessState(role.Pet)
				break
			}

			s.openPetSkillWindow()
			if x, y, ok := s.getSkillWindowPos(); ok {
				offset := int(s.PetActions[s.currentPetActionId].Offset)
				game.UsePetSkill(s.hWnd, x, y, offset)
				s.logP("tries to get on ride")
				s.currentCU = controlunit.Repeat
			} else {
				s.logP("cannot find the position of window")
				s.setFailureState(role.Pet)
			}
		case pet.OffRide:
			if !s.isOnRide() {
				s.logP("is off ride")
				s.setSuccessState(role.Pet)
				break
			}

			s.openPetSkillWindow()
			if x, y, ok := s.getSkillWindowPos(); ok {
				offset := int(s.PetActions[s.currentPetActionId].Offset)
				game.UsePetSkill(s.hWnd, x, y, offset)
				s.logP("tries to get off ride")
				s.currentCU = controlunit.Repeat
			} else {
				s.logP("cannot find the position of window")
				s.setFailureState(role.Pet)
			}
		case pet.Catch:
			game.CloseAllWindows(s.hWnd)
			game.ClearChat(s.hWnd)
			if self, ok := s.getSelfTarget(s.isOnRide()); ok {
				ratio, _ := strconv.ParseFloat(s.PetActions[s.currentPetActionId].Param, 32)
				if s.isHealthLowerThan(float32(ratio), self) {
					s.isOutOfHealth = true
					s.logP("is out of health")
				}
			}
		case pet.Hang:
			s.logP("is hanging")
			s.isPetHanging = true
			s.currentCU = controlunit.Repeat
		}

		s.updateCurrentActionId(role.Pet)
	}

	if len(s.PetActions) != 0 {
		s.currentPetActionId %= len(s.PetActions)
	} else {
		s.currentPetActionId = 0
	}
}

func (s *ActionState) reset() {
	s.currentCharacterActionId = 0
	s.currentPetActionId = 0
	s.trainingCounter = 0
}

func (s *ActionState) resetCurrentControlUnit() {
	s.currentCU = controlunit.Undefined
	s.currentJumpId = 0
}

func (s *ActionState) setSuccessState(r role.Role) {
	switch r {
	case role.Character:
		s.currentCU = s.CharacterActions[s.currentCharacterActionId].SuccessControlUnit
		s.currentJumpId = s.CharacterActions[s.currentCharacterActionId].SuccessJumpId
	case role.Pet:
		s.currentCU = s.PetActions[s.currentPetActionId].SuccessControlUnit
		s.currentJumpId = s.PetActions[s.currentPetActionId].SuccessJumpId
	}
}

func (s *ActionState) setFailureState(r role.Role) {
	switch r {
	case role.Character:
		s.currentCU = s.CharacterActions[s.currentCharacterActionId].FailureControlUnit
		s.currentJumpId = s.CharacterActions[s.currentCharacterActionId].FailureJumpId
	case role.Pet:
		s.currentCU = s.PetActions[s.currentPetActionId].FailureControlUnit
		s.currentJumpId = s.PetActions[s.currentPetActionId].FailureJumpId
	}
}

func (s *ActionState) updateCurrentActionId(r role.Role) {
	switch r {
	case role.Character:
		switch s.currentCU {
		case controlunit.StartOver:
			s.currentCharacterActionId = 0
		case controlunit.Continue:
			s.currentCharacterActionId++
		case controlunit.Repeat:
			return
		case controlunit.Jump:
			s.currentCharacterActionId = s.currentJumpId
		default:
			s.currentCharacterActionId++
		}
	case role.Pet:
		switch s.currentCU {
		case controlunit.StartOver:
			s.currentPetActionId = 0
		case controlunit.Continue:
			s.currentPetActionId++
		case controlunit.Repeat:
			return
		case controlunit.Jump:
			s.currentPetActionId = s.currentJumpId
		default:
			s.currentPetActionId++
		}
	}

	s.currentCU = controlunit.Undefined
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
	var targets []game.CheckTarget
	if len(s.CustomEnemies) != 0 {
		targets = s.enemies
	} else {
		targets = make([]game.CheckTarget, len(s.enemies))
		copy(targets, s.enemies)
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		r.Shuffle(len(targets), func(i, j int) { targets[i], targets[j] = targets[j], targets[i] })
	}

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
	header := fmt.Sprintf("[%s][%d][%s]", fmt.Sprint(s.hWnd), s.currentCharacterActionId, strings.Trim(s.CharacterActions[s.currentCharacterActionId].Action.String(), "*"))
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
		CharacterActions: []CharacterAction{
			{
				Action:             character.Attack,
				SuccessControlUnit: controlunit.Continue,
				FailureControlUnit: controlunit.Continue,
			},
		},
		PetActions: []PetAction{
			{
				Action:             pet.Attack,
				SuccessControlUnit: controlunit.Continue,
				FailureControlUnit: controlunit.Continue,
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

func (s *ActionState) AddCharacterAction(action character.Action) {
	newCharacterState := CharacterAction{
		Action: action,
	}
	s.CharacterActions = append(s.CharacterActions, newCharacterState)
}

func (s *ActionState) AddCharacterSkillLevel(level offset.Offset) {
	s.CharacterActions[len(s.CharacterActions)-1].Level = level
}

func (s *ActionState) AddPetAction(action pet.Action) {
	newPetState := PetAction{
		Action: action,
	}
	s.PetActions = append(s.PetActions, newPetState)
}

func (s *ActionState) AddSkillOffset(r role.Role, offset offset.Offset) {
	switch r {
	case role.Character:
		s.CharacterActions[len(s.CharacterActions)-1].Offset = offset
	case role.Pet:
		s.PetActions[len(s.PetActions)-1].Offset = offset
	}
}

func (s *ActionState) AddThreshold(r role.Role, threshold threshold.Threshold) {
	s.CharacterActions[len(s.CharacterActions)-1].Threshold = threshold
}

func (s *ActionState) AddParam(r role.Role, param string) {
	switch r {
	case role.Character:
		s.CharacterActions[len(s.CharacterActions)-1].Param = param
	case role.Pet:
		s.PetActions[len(s.PetActions)-1].Param = param
	}
}

func (s *ActionState) AddSuccessControlUnit(r role.Role, cu controlunit.ControlUnit) {
	switch r {
	case role.Character:
		s.CharacterActions[len(s.CharacterActions)-1].SuccessControlUnit = cu
	case role.Pet:
		s.PetActions[len(s.PetActions)-1].SuccessControlUnit = cu
	}
}
func (s *ActionState) AddSuccessJumpId(r role.Role, jumpId int) {
	switch r {
	case role.Character:
		s.CharacterActions[len(s.CharacterActions)-1].SuccessJumpId = jumpId
	case role.Pet:
		s.PetActions[len(s.PetActions)-1].SuccessJumpId = jumpId
	}
}

func (s *ActionState) AddFailureControlUnit(r role.Role, cu controlunit.ControlUnit) {
	switch r {
	case role.Character:
		s.CharacterActions[len(s.CharacterActions)-1].FailureControlUnit = cu
	case role.Pet:
		s.PetActions[len(s.PetActions)-1].FailureControlUnit = cu
	}
}

func (s *ActionState) AddFailureJumpId(r role.Role, jumpId int) {
	switch r {
	case role.Character:
		s.CharacterActions[len(s.CharacterActions)-1].FailureJumpId = jumpId
	case role.Pet:
		s.PetActions[len(s.PetActions)-1].FailureJumpId = jumpId
	}
}

func (s *ActionState) ClearCharacterActions() {
	s.CharacterActions = s.CharacterActions[:0]
}

func (s *ActionState) ClearPetActions() {
	s.PetActions = s.PetActions[:0]
}

func (s *ActionState) GetCharacterActions() []CharacterAction {
	return s.CharacterActions
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

	if len(s.CustomEnemies) != 0 {
		s.enemies = s.getEnemies(s.CustomEnemies)
	} else {
		s.enemies = s.getEnemies(AllEnemies)
	}

	log.Printf("# Handle %s detected %d enemies\n", fmt.Sprint(s.hWnd), len(s.enemies))
}

func (s *ActionState) endCharacterHanging() {
	if s.isCharacterHanging {
		s.isCharacterHanging = false
		cu := s.CharacterActions[s.currentCharacterActionId].SuccessControlUnit
		jumpId := s.CharacterActions[s.currentCharacterActionId].SuccessJumpId

		switch cu {
		case controlunit.StartOver:
			s.currentCharacterActionId = 0
		case controlunit.Continue:
			s.currentCharacterActionId++
		case controlunit.Jump:
			s.currentCharacterActionId = jumpId
		}
	}
}
func (s *ActionState) endPetHanging() {
	if s.isPetHanging {
		s.isPetHanging = false
		cu := s.PetActions[s.currentPetActionId].SuccessControlUnit
		jumpId := s.PetActions[s.currentPetActionId].SuccessJumpId

		switch cu {
		case controlunit.StartOver:
			s.currentPetActionId = 0
		case controlunit.Continue:
			s.currentPetActionId++
		case controlunit.Jump:
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

	for game.IsBattleScene(s.hWnd) && s.Enabled && !s.isCharacterStageStable() && !s.isPetStageStable() {
		time.Sleep(DURATION_BATTLE_ACTION_WAITING_LOOP)
	}
}

func Debug(hWnd win.HWND) (x int32, y int32, successful bool) {
	game.CloseAllWindows(hWnd)
	game.ClearChat(hWnd)
	game.OpenWindow(hWnd, game.KEY_INVENTORY)
	return game.GetItemPos(hWnd, x, y, 16448250, 3)
}

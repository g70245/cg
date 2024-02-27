package game

import "fmt"

var (
	ControlUnits controlUnits = []ControlUnit{ControlUnitStartOver, ControlUnitContinue, ControlUnitRepeat, ControlUnitJump}
	Offsets      offsets      = []Offset{OffsetOne, OffsetTwo, OffsetThree, OffsetFour, OffsetFive, OffsetSix, OffsetSeven, OffsetEight, OffsetNine, OffsetTen}
	Levels       levels       = []Level{LevelOne, LevelTwo, LevelThree, LevelFour, LevelFive, LevelSix, LevelSeven, LevelEight, LevelNine, LevelTen}
	Ratios       ratios       = []ratio{RatioOneTenth, RatioTwoTenth, RatioThreeTenth, RatioFourTenth, RatioFiveTenth, RatioSixTenth, RatioSevenTenth, RatioEightTenth, RatioNineTenth}
	Thresholds   thresholds   = []Threshold{ThresholdOne, ThresholdTwo, ThresholdThree, ThresholdFour, ThresholdFive, ThresholdSix, ThresholdSeven, ThresholdEight, ThresholdNine, ThresholdTen}
)

type Role int

const (
	Human Role = iota
	Pet
)

type ControlUnit string
type controlUnits []ControlUnit

const (
	ControlUnitStartOver ControlUnit = "Start Over"
	ControlUnitContinue  ControlUnit = "Continue"
	ControlUnitRepeat    ControlUnit = "Repeat"
	ControlUnitJump      ControlUnit = "Jump"
)

type humanAction int

const (
	HumanAttack humanAction = iota
	HumanDefend
	HumanEscape
	HumanMove
	HumanBomb
	HumanRecall
	HumanPotion
	HumanSkill
	HumanThresholdSkill
	HumanHealSelf
	HumanHealOne
	HumanHealTShaped
	HumanHealMulti
	HumanRide
	HumanHang
	HumanSteal
	HumanCatch
	HumanTrainSkill
)

type petAction int

const (
	PetAttack petAction = iota
	PetEscape
	PetSkill
	PetDefend
	PetHealSelf
	PetHealOne
	PetRide
	PetOffRide
	PetCatch
	PetHang
)

type Offset int
type offsets []Offset

const (
	OffsetOne   Offset = 1
	OffsetTwo   Offset = 2
	OffsetThree Offset = 3
	OffsetFour  Offset = 4
	OffsetFive  Offset = 5
	OffsetSix   Offset = 6
	OffsetSeven Offset = 7
	OffsetEight Offset = 8
	OffsetNine  Offset = 9
	OffsetTen   Offset = 10
)

type Level int
type levels []Level

const (
	LevelOne   Level = 1
	LevelTwo   Level = 2
	LevelThree Level = 3
	LevelFour  Level = 4
	LevelFive  Level = 5
	LevelSix   Level = 6
	LevelSeven Level = 7
	LevelEight Level = 8
	LevelNine  Level = 9
	LevelTen   Level = 10
)

type ratio float32
type ratios []ratio

const (
	RatioOneTenth   ratio = 0.1
	RatioTwoTenth   ratio = 0.2
	RatioThreeTenth ratio = 0.3
	RatioFourTenth  ratio = 0.4
	RatioFiveTenth  ratio = 0.5
	RatioSixTenth   ratio = 0.6
	RatioSevenTenth ratio = 0.7
	RatioEightTenth ratio = 0.8
	RatioNineTenth  ratio = 0.9
)

type Threshold string
type thresholds []Threshold

const (
	ThresholdOne   Threshold = "1 foe"
	ThresholdTwo   Threshold = "2 foes"
	ThresholdThree Threshold = "3 foes"
	ThresholdFour  Threshold = "4 foes"
	ThresholdFive  Threshold = "5 foes"
	ThresholdSix   Threshold = "6 foes"
	ThresholdSeven Threshold = "7 foes"
	ThresholdEight Threshold = "8 foes"
	ThresholdNine  Threshold = "9 foes"
	ThresholdTen   Threshold = "10 foes"
)

func (ha humanAction) String() string {
	switch ha {
	case HumanAttack:
		return "**Attack"
	case HumanDefend:
		return "**Defend"
	case HumanEscape:
		return "**Escape"
	case HumanMove:
		return "**Move"
	case HumanBomb:
		return "*Bomb"
	case HumanRecall:
		return "*Recall Pet"
	case HumanPotion:
		return "*Potion"
	case HumanSkill:
		return "*Skill"
	case HumanThresholdSkill:
		return "*Threshold Skill"
	case HumanHealSelf:
		return "*Heal Self"
	case HumanHealOne:
		return "*Heal One"
	case HumanHealTShaped:
		return "*Heal T-Shaped"
	case HumanHealMulti:
		return "*Heal Multi"
	case HumanRide:
		return "*Ride"
	case HumanHang:
		return "Hang"
	case HumanSteal:
		return "Steal"
	case HumanCatch:
		return "Catch"
	case HumanTrainSkill:
		return "Train Skill"
	default:
		return "Unknown"
	}
}

func (pa petAction) String() string {
	switch pa {
	case PetAttack:
		return "**Pet Attack"
	case PetEscape:
		return "**Pet Escape"
	case PetSkill:
		return "*Pet Skill"
	case PetDefend:
		return "*Pet Defend"
	case PetHealSelf:
		return "*Pet Heal Self"
	case PetHealOne:
		return "*Pet Heal One"
	case PetRide:
		return "*Pet Ride"
	case PetOffRide:
		return "*Pet Off Ride"
	case PetCatch:
		return "Pet Catch"
	case PetHang:
		return "Pet Hang"
	default:
		return "Unknown"
	}
}

func (c controlUnits) GetOptions() []string {
	var options []string
	for i := range c {
		options = append(options, fmt.Sprint(c[i]))
	}
	return options
}

func (ids offsets) GetOptions() []string {
	var options []string
	for i := range ids {
		options = append(options, fmt.Sprint(ids[i]))
	}
	return options
}

func (ls levels) GetOptions() []string {
	var options []string
	for i := range ls {
		options = append(options, fmt.Sprint(ls[i]))
	}
	return options
}

func (rs ratios) GetOptions() []string {
	var options []string
	for i := range rs {
		options = append(options, fmt.Sprint(rs[i]))
	}
	return options
}

func (ts thresholds) GetOptions() []string {
	var options []string
	for i := range ts {
		options = append(options, fmt.Sprint(ts[i]))
	}
	return options
}

package battle

import "fmt"

var (
	ControlUnits controlUnits = []ControlUnit{StartOver, Continue, Repeat, Jump}
	Offsets      offsets      = []Offset{One, Two, Three, Four, Five, Six, Seven, Eight, Nine, Ten}
	Levels       offsets      = []Offset{One, Two, Three, Four, Five, Six, Seven, Eight, Nine, Ten}
	Ratios       ratios       = []ratio{OneTenth, TwoTenth, ThreeTenth, FourTenth, FiveTenth, SixTenth, SevenTenth, EightTenth, NineTenth}
	Thresholds   thresholds   = []Threshold{OneFoe, TwoFoes, ThreeFoes, FourFoes, FiveFoes, SixFoes, SevenFoes, EightFoes, NineFoes, TenFoes}
)

type Role int

const (
	Human Role = iota
	Pet
)

type ControlUnit string
type controlUnits []ControlUnit

const (
	Undefined ControlUnit = "Undefined"
	StartOver ControlUnit = "Start Over"
	Continue  ControlUnit = "Continue"
	Repeat    ControlUnit = "Repeat"
	Jump      ControlUnit = "Jump"
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
	HumanBloodMagic
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
	PetDefend
	PetProtect
	PetSkill
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
	One   Offset = 1
	Two   Offset = 2
	Three Offset = 3
	Four  Offset = 4
	Five  Offset = 5
	Six   Offset = 6
	Seven Offset = 7
	Eight Offset = 8
	Nine  Offset = 9
	Ten   Offset = 10
)

type ratio float32
type ratios []ratio

const (
	OneTenth   ratio = 0.1
	TwoTenth   ratio = 0.2
	ThreeTenth ratio = 0.3
	FourTenth  ratio = 0.4
	FiveTenth  ratio = 0.5
	SixTenth   ratio = 0.6
	SevenTenth ratio = 0.7
	EightTenth ratio = 0.8
	NineTenth  ratio = 0.9
)

type Threshold string
type thresholds []Threshold

const (
	OneFoe    Threshold = "1 foe"
	TwoFoes   Threshold = "2 foes"
	ThreeFoes Threshold = "3 foes"
	FourFoes  Threshold = "4 foes"
	FiveFoes  Threshold = "5 foes"
	SixFoes   Threshold = "6 foes"
	SevenFoes Threshold = "7 foes"
	EightFoes Threshold = "8 foes"
	NineFoes  Threshold = "9 foes"
	TenFoes   Threshold = "10 foes"
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
	case HumanBloodMagic:
		return "*Blood Magic"
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

func (o offsets) GetOptions() []string {
	var options []string
	for i := range o {
		options = append(options, fmt.Sprint(o[i]))
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

package game

import "fmt"

type ControlUnit string
type ControlUnits []ControlUnit
type HumanAction int
type PetAction int

const (
	CU_START_OVER ControlUnit = "Start Over"
	CU_CONTINUE   ControlUnit = "Continue"
	CU_REPEAT     ControlUnit = "Repeat"
	CU_JUMP       ControlUnit = "Jump"
)

const (
	H_Attack HumanAction = iota
	H_Defend
	H_ESCAPE
	H_MOVE
	H_BOMB
	H_RECALL
	H_POTION
	H_SKILL
	H_THRS_SKILL
	H_SELF_HEAL
	H_ONE_HEAL
	H_T_SHAPED_HEAL
	H_MULTI_HEAL
	H_RIDE
	H_HANG
	H_STEAL
	H_CATCH
	H_TRAIN_SKILL
)

const (
	P_ATTACK PetAction = iota
	P_ESCAPE
	P_SkILL
	P_DEFEND
	P_SELF_HEAL
	P_ONE_HEAL
	P_RIDE
	P_OFF_RIDE
	P_CATCH
	P_HANG
)

func (c ControlUnits) GetOptions() []string {
	var options []string
	for i := range c {
		options = append(options, fmt.Sprint(c[i]))
	}
	return options
}

func (ha HumanAction) String() string {
	switch ha {
	case H_Attack:
		return "**Attack"
	case H_Defend:
		return "**Defend"
	case H_ESCAPE:
		return "**Escape"
	case H_MOVE:
		return "**Move"
	case H_BOMB:
		return "*Bomb"
	case H_RECALL:
		return "*Recall Pet"
	case H_POTION:
		return "*Potion"
	case H_SKILL:
		return "*Skill"
	case H_THRS_SKILL:
		return "*Threshold Skill"
	case H_SELF_HEAL:
		return "*Heal Self"
	case H_ONE_HEAL:
		return "*Heal One"
	case H_T_SHAPED_HEAL:
		return "*Heal T-Shaped"
	case H_MULTI_HEAL:
		return "*Heal Multi"
	case H_RIDE:
		return "*Ride"
	case H_HANG:
		return "Hang"
	case H_STEAL:
		return "Steal"
	case H_CATCH:
		return "Catch"
	case H_TRAIN_SKILL:
		return "Train Skill"
	default:
		return "Unknown"
	}
}

func (pa PetAction) String() string {
	switch pa {
	case P_ATTACK:
		return "**Pet Attack"
	case P_ESCAPE:
		return "**Pet Escape"
	case P_SkILL:
		return "*Pet Skill"
	case P_DEFEND:
		return "*Pet Defend"
	case P_SELF_HEAL:
		return "*Pet Heal Self"
	case P_ONE_HEAL:
		return "*Pet Heal One"
	case P_RIDE:
		return "*Pet Ride"
	case P_OFF_RIDE:
		return "*Pet Off Ride"
	case P_CATCH:
		return "Pet Catch"
	case P_HANG:
		return "Pet Hang"
	default:
		return "Unknown"
	}
}

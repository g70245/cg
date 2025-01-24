package human

type Action int

const (
	Attack Action = iota
	Defend
	Escape
	Move
	Bomb
	Recall
	Potion
	Skill
	ThresholdSkill
	BloodMagic
	HealSelf
	HealOne
	HealTShaped
	HealMulti
	Ride
	Hang
	Steal
	Catch
	TrainSkill
)

func (a Action) String() string {
	switch a {
	case Attack:
		return "**Attack"
	case Defend:
		return "**Defend"
	case Escape:
		return "**Escape"
	case Move:
		return "**Move"
	case Bomb:
		return "*Bomb"
	case Recall:
		return "*Recall Pet"
	case Potion:
		return "*Potion"
	case Skill:
		return "*Skill"
	case ThresholdSkill:
		return "*Threshold Skill"
	case BloodMagic:
		return "*Blood Magic"
	case HealSelf:
		return "*Heal Self"
	case HealOne:
		return "*Heal One"
	case HealTShaped:
		return "*Heal T-Shaped"
	case HealMulti:
		return "*Heal Multi"
	case Ride:
		return "*Ride"
	case Hang:
		return "Hang"
	case Steal:
		return "Steal"
	case Catch:
		return "Catch"
	case TrainSkill:
		return "Train Skill"
	default:
		return "Unknown"
	}
}

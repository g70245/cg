package pet

type Action int

const (
	Attack Action = iota
	Escape
	Defend
	Protect
	Skill
	HealSelf
	HealOne
	Ride
	OffRide
	Catch
	Hang
)

func (a Action) String() string {
	switch a {
	case Attack:
		return "**Pet Attack"
	case Escape:
		return "**Pet Escape"
	case Skill:
		return "*Pet Skill"
	case Defend:
		return "*Pet Defend"
	case HealSelf:
		return "*Pet Heal Self"
	case HealOne:
		return "*Pet Heal Ally"
	case Ride:
		return "*Pet Ride"
	case OffRide:
		return "*Pet Dismount"
	case Catch:
		return "Pet Catch"
	case Hang:
		return "Pet Wait"
	default:
		return "Unknown"
	}
}

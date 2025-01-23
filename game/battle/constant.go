package battle

import (
	"cg/game/battle/enums"
	"cg/game/battle/enums/controlunit"
	"cg/game/battle/enums/offset"
	"cg/game/battle/enums/ratio"
	"cg/game/battle/enums/threshold"
)

var (
	ControlUnits = enums.GenericEnum[controlunit.ControlUnit]{List: []controlunit.ControlUnit{controlunit.StartOver, controlunit.Continue, controlunit.Repeat, controlunit.Jump}}
	Offsets      = enums.GenericEnum[offset.Offset]{List: []offset.Offset{offset.One, offset.Two, offset.Three, offset.Four, offset.Five, offset.Six, offset.Seven, offset.Eight, offset.Nine, offset.Ten}}
	Levels       = enums.GenericEnum[offset.Offset]{List: []offset.Offset{offset.One, offset.Two, offset.Three, offset.Four, offset.Five, offset.Six, offset.Seven, offset.Eight, offset.Nine, offset.Ten}}
	Ratios       = enums.GenericEnum[ratio.Ratio]{List: []ratio.Ratio{ratio.OneTenth, ratio.TwoTenth, ratio.ThreeTenth, ratio.FourTenth, ratio.FiveTenth, ratio.SixTenth, ratio.SevenTenth, ratio.EightTenth, ratio.NineTenth}}
	Thresholds   = enums.GenericEnum[threshold.Threshold]{List: []threshold.Threshold{threshold.OneFoe, threshold.TwoFoes, threshold.ThreeFoes, threshold.FourFoes, threshold.FiveFoes, threshold.SixFoes, threshold.SevenFoes, threshold.EightFoes, threshold.NineFoes, threshold.TenFoes}}
)

const (
	COLOR_BATTLE_COMMAND_ENABLE  = 7125907
	COLOR_BATTLE_COMMAND_DISABLE = 6991316

	COLOR_BATTLE_STAGE_HUMAN = 15398392
	COLOR_BATTLE_STAGE_PET   = 8599608

	COLOR_BATTLE_BLOOD_UPPER      = 9211135
	COLOR_BATTLE_BLOOD_LOWER      = 255
	COLOR_BATTLE_MANA_UPPER       = 16758653
	COLOR_BATTLE_MANA_LOWER       = 16740864
	COLOR_BATTLE_BLOOD_MANA_EMPTY = 65536

	COLOR_BATTLE_RECALL_BUTTON = 7694643
	COLOR_BATTLE_SELF_NAME_1   = 37083
	COLOR_BATTLE_SELF_NAME_2   = 37086
	COLOR_BATTLE_SELF_NAME_3   = 37087
	COLOR_BATTLE_SELF_NAME_4   = 37050
	COLOR_BATTLE_SELF_NAME_5   = 37008

	COLOR_BS_INVENTORY_PIVOT = 15967
)

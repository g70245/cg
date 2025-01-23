package battle

import (
	"cg/game/battle/enums/controlunit"
	"cg/game/battle/enums/offset"
	"cg/game/battle/enums/ratio"
	"cg/game/battle/enums/threshold"
	"fmt"
)

var (
	ControlUnits controlUnits = []controlunit.ControlUnit{controlunit.StartOver, controlunit.Continue, controlunit.Repeat, controlunit.Jump}
	Offsets      offsets      = []offset.Offset{offset.One, offset.Two, offset.Three, offset.Four, offset.Five, offset.Six, offset.Seven, offset.Eight, offset.Nine, offset.Ten}
	Levels       offsets      = []offset.Offset{offset.One, offset.Two, offset.Three, offset.Four, offset.Five, offset.Six, offset.Seven, offset.Eight, offset.Nine, offset.Ten}
	Ratios       ratios       = []ratio.Ratio{ratio.OneTenth, ratio.TwoTenth, ratio.ThreeTenth, ratio.FourTenth, ratio.FiveTenth, ratio.SixTenth, ratio.SevenTenth, ratio.EightTenth, ratio.NineTenth}
	Thresholds   thresholds   = []threshold.Threshold{threshold.OneFoe, threshold.TwoFoes, threshold.ThreeFoes, threshold.FourFoes, threshold.FiveFoes, threshold.SixFoes, threshold.SevenFoes, threshold.EightFoes, threshold.NineFoes, threshold.TenFoes}
)

type controlUnits []controlunit.ControlUnit
type offsets []offset.Offset
type ratios []ratio.Ratio
type thresholds []threshold.Threshold

func (cus controlUnits) GetOptions() []string {
	var options []string
	for i := range cus {
		options = append(options, fmt.Sprint(cus[i]))
	}
	return options
}

func (os offsets) GetOptions() []string {
	var options []string
	for i := range os {
		options = append(options, fmt.Sprint(os[i]))
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

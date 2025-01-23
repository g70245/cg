package battle

import (
	"cg/game/battle/enums/controlunit"
	"fmt"
)

var (
	ControlUnits controlUnits = []controlunit.ControlUnit{controlunit.StartOver, controlunit.Continue, controlunit.Repeat, controlunit.Jump}
	Offsets      offsets      = []Offset{One, Two, Three, Four, Five, Six, Seven, Eight, Nine, Ten}
	Levels       offsets      = []Offset{One, Two, Three, Four, Five, Six, Seven, Eight, Nine, Ten}
	Ratios       ratios       = []ratio{OneTenth, TwoTenth, ThreeTenth, FourTenth, FiveTenth, SixTenth, SevenTenth, EightTenth, NineTenth}
	Thresholds   thresholds   = []Threshold{OneFoe, TwoFoes, ThreeFoes, FourFoes, FiveFoes, SixFoes, SevenFoes, EightFoes, NineFoes, TenFoes}
)

type controlUnits []controlunit.ControlUnit

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

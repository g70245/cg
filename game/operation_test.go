package game

import "testing"

func TestMapStepTargetUsesIsometricMapAxes(t *testing.T) {
	tests := []struct {
		name                  string
		deltaEast, deltaSouth int
		wantX, wantY          int32
	}{
		{name: "east", deltaEast: 1, wantX: 352, wantY: 216},
		{name: "west", deltaEast: -1, wantX: 288, wantY: 264},
		{name: "south", deltaSouth: 1, wantX: 352, wantY: 264},
		{name: "north", deltaSouth: -1, wantX: 288, wantY: 216},
		{name: "four east", deltaEast: 4, wantX: 448, wantY: 144},
		{name: "eight east", deltaEast: 8, wantX: 576, wantY: 48},
		{name: "eight south", deltaSouth: 8, wantX: 576, wantY: 432},
		{name: "four southeast", deltaEast: 4, deltaSouth: 4, wantX: 576, wantY: 240},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			x, y := mapStepTarget(test.deltaEast, test.deltaSouth)
			if x != test.wantX || y != test.wantY {
				t.Fatalf("mapStepTarget() = (%d, %d), want (%d, %d)", x, y, test.wantX, test.wantY)
			}
		})
	}
}

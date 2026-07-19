package navigation

import (
	"reflect"
	"testing"
)

func TestFindPathRoutesAroundBlockedCells(t *testing.T) {
	data := walkableMap(4, 3)
	data.Walkable[1] = false
	data.Walkable[5] = false
	data.Stairs = []Stair{{East: 3, South: 0, Type: StairUp}}

	path, target, ok := FindPath(data, Point{}, StairUp, nil)
	if !ok {
		t.Fatal("FindPath() found no path")
	}
	want := []Point{
		{East: 0, South: 0},
		{East: 0, South: 1},
		{East: 1, South: 2},
		{East: 2, South: 1},
		{East: 3, South: 0},
	}
	if !reflect.DeepEqual(path, want) {
		t.Fatalf("FindPath() = %v, want %v", path, want)
	}
	if target.Type != StairUp {
		t.Fatalf("target = %#v, want Up", target)
	}
}

func TestFindPathChoosesNearestReachableTarget(t *testing.T) {
	data := walkableMap(5, 3)
	data.Stairs = []Stair{
		{East: 4, South: 0, Type: StairDown},
		{East: 1, South: 2, Type: StairDown},
	}
	path, target, ok := FindPath(data, Point{}, StairDown, nil)
	if !ok {
		t.Fatal("FindPath() found no path")
	}
	if target.East != 1 || target.South != 2 || len(path) != 3 {
		t.Fatalf("target/path = %#v %v, want nearest stair at (1,2)", target, path)
	}
}

func TestFindPathDoesNotFallBackToPassage(t *testing.T) {
	data := walkableMap(2, 1)
	data.Stairs = []Stair{{East: 1, Type: StairPassage}}
	if path, target, ok := FindPath(data, Point{}, StairUp, nil); ok {
		t.Fatalf("FindPath() = %v, %#v, true; want Passage ignored", path, target)
	}
}

func TestFindPathDoesNotCrossOtherTransitions(t *testing.T) {
	data := walkableMap(3, 1)
	data.Stairs = []Stair{
		{East: 1, Type: StairPassage},
		{East: 2, Type: StairUp},
	}
	blocked := blockedTransitions(data, StairUp)
	if path, _, ok := FindPath(data, Point{}, StairUp, blocked); ok {
		t.Fatalf("FindPath() crossed Passage: %v", path)
	}
}

func TestFindPathHonorsTemporaryBlocks(t *testing.T) {
	data := walkableMap(3, 1)
	data.Stairs = []Stair{{East: 2, Type: StairUp}}
	if _, _, ok := FindPath(data, Point{}, StairUp, map[Point]bool{{East: 1}: true}); ok {
		t.Fatal("FindPath() crossed a temporary block")
	}
}

func TestFindPathCanLeaveKnownStartingCellMarkedBlocked(t *testing.T) {
	data := walkableMap(3, 1)
	data.Walkable[0] = false
	data.Stairs = []Stair{{East: 2, Type: StairUp}}
	path, _, ok := FindPath(data, Point{}, StairUp, nil)
	want := []Point{{}, {East: 1}, {East: 2}}
	if !ok || !reflect.DeepEqual(path, want) {
		t.Fatalf("FindPath() = %v, %v; want route from live starting cell %v", path, ok, want)
	}
}

func TestFindPathDoesNotSelectBlockedStartingTarget(t *testing.T) {
	data := walkableMap(2, 1)
	data.Stairs = []Stair{
		{Type: StairPassage},
		{East: 1, Type: StairPassage},
	}
	path, target, ok := FindPath(data, Point{}, StairPassage, map[Point]bool{{}: true})
	if !ok || target.East != 1 || !reflect.DeepEqual(path, []Point{{}, {East: 1}}) {
		t.Fatalf("FindPath() = %v, %#v, %v; want unblocked Passage at east 1", path, target, ok)
	}
}

func TestCanStepAllowsDiagonalWhenEitherSideIsOpen(t *testing.T) {
	tests := []struct {
		name      string
		eastOpen  bool
		southOpen bool
		want      bool
	}{
		{name: "east side", eastOpen: true, want: true},
		{name: "south side", southOpen: true, want: true},
		{name: "both sides", eastOpen: true, southOpen: true, want: true},
		{name: "closed corner", want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data := MapData{Width: 2, Height: 2, Known: []bool{true, true, true, true}, Walkable: []bool{true, test.eastOpen, test.southOpen, true}}
			if got := canStep(data, Point{}, Point{East: 1, South: 1}, nil); got != test.want {
				t.Fatalf("canStep() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestFindPathUsesLegalDiagonal(t *testing.T) {
	data := MapData{Width: 2, Height: 2, Known: []bool{true, true, true, true}, Walkable: []bool{true, true, false, true}}
	data.Stairs = []Stair{{East: 1, South: 1, Type: StairUp}}
	path, _, ok := FindPath(data, Point{}, StairUp, nil)
	want := []Point{{}, {East: 1, South: 1}}
	if !ok || !reflect.DeepEqual(path, want) {
		t.Fatalf("FindPath() = %v, %v; want diagonal %v", path, ok, want)
	}
}

func TestFindExplorationPathContinuesCurrentBranch(t *testing.T) {
	data := walkableMap(5, 3)
	for index := range data.Known {
		data.Known[index] = true
	}
	data.Known[4] = false
	data.Walkable[4] = false
	data.Known[11] = false
	data.Walkable[11] = false

	path, ok := FindExplorationPath(data, Point{East: 2, South: 1}, nil, nil, Point{East: 1})
	if !ok {
		t.Fatal("FindExplorationPath() found no frontier")
	}
	wantTarget := Point{East: 3, South: 1}
	if got := path[len(path)-1]; got != wantTarget {
		t.Fatalf("frontier = %v, want forward branch %v; path %v", got, wantTarget, path)
	}
	if len(path) < 2 || path[1].East <= path[0].East {
		t.Fatalf("path %v did not continue east", path)
	}
}

func TestFindExplorationPathSkipsCompletedAndTransitions(t *testing.T) {
	data := walkableMap(3, 1)
	data.Known[2] = false
	data.Walkable[2] = false
	data.Stairs = []Stair{{East: 1, Type: StairPassage}}
	blocked := blockedTransitions(data, StairUp)
	if path, ok := FindExplorationPath(data, Point{}, blocked, nil, Point{}); ok {
		t.Fatalf("FindExplorationPath() crossed Passage: %v", path)
	}

	data.Stairs = nil
	completed := map[Point]bool{{East: 1}: true}
	if path, ok := FindExplorationPath(data, Point{}, nil, completed, Point{}); ok {
		t.Fatalf("FindExplorationPath() returned completed frontier: %v", path)
	}
}

func walkableMap(width, height int) MapData {
	data := MapData{Width: width, Height: height, Known: make([]bool, width*height), Walkable: make([]bool, width*height)}
	for index := range data.Known {
		data.Known[index] = true
	}
	for index := range data.Walkable {
		data.Walkable[index] = true
	}
	return data
}

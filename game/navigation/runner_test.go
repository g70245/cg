package navigation

import (
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestRunnerMovesToExitAndStopsAfterLeavingMaze(t *testing.T) {
	data := walkableMap(2, 1)
	data.Stairs = []Stair{{East: 1, Type: StairUp}}
	var mu sync.Mutex
	snapshot := MovementSnapshot{MapCode: 699, Position: Point{}, PositionSettled: true, Map: data, InMaze: true}
	var moves []Point
	var statuses []string
	runner := Runner{
		PollInterval: time.Millisecond,
		StepTimeout:  20 * time.Millisecond,
		ReadSnapshot: func() (MovementSnapshot, error) {
			mu.Lock()
			defer mu.Unlock()
			return snapshot, nil
		},
		CanMove: func() bool { return true },
		Move: func(delta Point) {
			mu.Lock()
			moves = append(moves, delta)
			snapshot.Position = Point{East: 1}
			snapshot.MapCode = 1401
			snapshot.InMaze = false
			mu.Unlock()
		},
	}
	if err := runner.Run(make(chan struct{}), StairUp, func(status string) {
		mu.Lock()
		statuses = append(statuses, status)
		mu.Unlock()
	}); err != nil {
		t.Fatal(err)
	}

	mu.Lock()
	defer mu.Unlock()
	if !reflect.DeepEqual(moves, []Point{{East: 1}}) {
		t.Fatalf("moves = %v, want one east step", moves)
	}
	if statuses[len(statuses)-1] != StatusCompleted {
		t.Fatalf("last status = %q, want %q", statuses[len(statuses)-1], StatusCompleted)
	}
}

func TestTraversalStatePreservesEntryPassageUntilFloorChanges(t *testing.T) {
	state := &TraversalState{}
	entry := Stair{East: 1, Type: StairPassage}
	newExit := Stair{East: 4, Type: StairPassage}

	first := state.BeginFloor(699, Point{}, []Stair{entry})
	second := state.BeginFloor(699, Point{}, []Stair{entry, newExit})
	if !first[Point{East: 1}] || !second[Point{East: 1}] || second[Point{East: 4}] {
		t.Fatalf("same-floor entry passages changed: first %v second %v", first, second)
	}

	third := state.BeginFloor(700, Point{East: 3}, []Stair{newExit})
	if third[Point{East: 1}] || !third[Point{East: 4}] {
		t.Fatalf("new-floor entry passages = %v", third)
	}

	state.Reset()
	reset := state.BeginFloor(700, Point{}, []Stair{entry})
	if !reset[Point{East: 1}] || reset[Point{East: 4}] {
		t.Fatalf("reset entry passages = %v", reset)
	}
}

func TestTraversalStateDoesNotRememberDistantPassageAsFloorEntry(t *testing.T) {
	state := &TraversalState{}
	passage := Stair{East: 5, South: 5, Type: StairPassage}
	entry := state.BeginFloor(699, Point{}, []Stair{passage})
	if entry[Point{East: 5, South: 5}] {
		t.Fatalf("distant Passage was remembered as floor entry: %v", entry)
	}
}

func TestRunnerUsesNewPassageButNotRememberedEntry(t *testing.T) {
	data := walkableMap(3, 1)
	data.Stairs = []Stair{
		{Type: StairPassage},
		{East: 2, Type: StairPassage},
	}
	state := &TraversalState{}
	state.BeginFloor(699, Point{}, data.Stairs[:1])
	snapshot := MovementSnapshot{MapCode: 699, PositionSettled: true, Map: data, InMaze: true}
	var moves []Point
	runner := Runner{
		State:        state,
		PollInterval: time.Millisecond,
		ReadSnapshot: func() (MovementSnapshot, error) { return snapshot, nil },
		CanMove:      func() bool { return true },
		Move: func(delta Point) {
			moves = append(moves, delta)
			snapshot.Position = Point{East: 2}
			snapshot.MapCode = 1401
			snapshot.InMaze = false
		},
	}
	if err := runner.Run(make(chan struct{}), StairUp, nil); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(moves, []Point{{East: 2}}) {
		t.Fatalf("moves = %v, want new Passage at east 2", moves)
	}
}

func TestRunnerUsesDistantPassageAlreadyVisibleOnFloorEntry(t *testing.T) {
	data := walkableMap(3, 1)
	data.Stairs = []Stair{{East: 2, Type: StairPassage}}
	snapshot := MovementSnapshot{MapCode: 699, PositionSettled: true, Map: data, InMaze: true}
	var moves []Point
	runner := Runner{
		State:        &TraversalState{},
		PollInterval: time.Millisecond,
		ReadSnapshot: func() (MovementSnapshot, error) { return snapshot, nil },
		CanMove:      func() bool { return true },
		Move: func(delta Point) {
			moves = append(moves, delta)
			snapshot.Position = Point{East: 2}
			snapshot.MapCode = 1401
			snapshot.InMaze = false
		},
	}
	if err := runner.Run(make(chan struct{}), StairUp, nil); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(moves, []Point{{East: 2}}) {
		t.Fatalf("moves = %v, want distant Passage exit", moves)
	}
}

func TestRunnerPrefersSelectedStairOverNewPassage(t *testing.T) {
	data := walkableMap(3, 1)
	data.Stairs = []Stair{
		{Type: StairPassage},
		{East: 1, Type: StairDown},
		{East: 2, Type: StairPassage},
	}
	state := &TraversalState{}
	state.BeginFloor(699, Point{}, data.Stairs[:1])
	snapshot := MovementSnapshot{MapCode: 699, PositionSettled: true, Map: data, InMaze: true}
	var moves []Point
	runner := Runner{
		State:        state,
		PollInterval: time.Millisecond,
		ReadSnapshot: func() (MovementSnapshot, error) { return snapshot, nil },
		CanMove:      func() bool { return true },
		Move: func(delta Point) {
			moves = append(moves, delta)
			snapshot.Position = Point{East: 1}
			snapshot.MapCode = 1401
			snapshot.InMaze = false
		},
	}
	if err := runner.Run(make(chan struct{}), StairDown, nil); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(moves, []Point{{East: 1}}) {
		t.Fatalf("moves = %v, want selected Down stair", moves)
	}
}

func TestRunnerDoesNotMoveWhenStartedOutsideMaze(t *testing.T) {
	moved := false
	var statuses []string
	runner := Runner{
		PollInterval: time.Millisecond,
		ReadSnapshot: func() (MovementSnapshot, error) {
			return MovementSnapshot{MapCode: 1401, PositionSettled: true, Map: walkableMap(2, 1)}, nil
		},
		CanMove: func() bool { return true },
		Move:    func(Point) { moved = true },
	}
	if err := runner.Run(make(chan struct{}), StairUp, func(status string) {
		statuses = append(statuses, status)
	}); err != nil {
		t.Fatal(err)
	}
	if moved {
		t.Fatal("runner moved outside the maze")
	}
	if statuses[len(statuses)-1] != StatusNotInMaze {
		t.Fatalf("last status = %q, want %q", statuses[len(statuses)-1], StatusNotInMaze)
	}
}

func TestRunnerWaitsUntilEveryWindowCanMove(t *testing.T) {
	data := walkableMap(2, 1)
	data.Stairs = []Stair{{East: 1, Type: StairDown}}
	cancel := make(chan struct{})
	paused := make(chan struct{}, 1)
	moved := make(chan struct{}, 1)
	canMove := false
	var mu sync.Mutex
	runner := Runner{
		PollInterval: time.Millisecond,
		ReadSnapshot: func() (MovementSnapshot, error) {
			return MovementSnapshot{MapCode: 699, PositionSettled: true, Map: data, InMaze: true}, nil
		},
		CanMove: func() bool {
			mu.Lock()
			defer mu.Unlock()
			return canMove
		},
		Move: func(Point) { moved <- struct{}{} },
	}
	done := make(chan error, 1)
	go func() {
		done <- runner.Run(cancel, StairDown, func(status string) {
			if status == StatusPaused {
				select {
				case paused <- struct{}{}:
				default:
				}
			}
		})
	}()
	select {
	case <-paused:
	case <-time.After(time.Second):
		t.Fatal("runner did not pause")
	}
	select {
	case <-moved:
		t.Fatal("runner moved while CanMove was false")
	default:
	}
	mu.Lock()
	canMove = true
	mu.Unlock()
	select {
	case <-moved:
	case <-time.After(time.Second):
		t.Fatal("runner did not resume")
	}
	close(cancel)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

func TestRunnerDoesNotClickBetweenGridCells(t *testing.T) {
	data := walkableMap(2, 1)
	data.Stairs = []Stair{{East: 1, Type: StairUp}}
	cancel := make(chan struct{})
	moved := make(chan struct{}, 1)
	runner := Runner{
		PollInterval: time.Millisecond,
		ReadSnapshot: func() (MovementSnapshot, error) {
			return MovementSnapshot{MapCode: 699, Map: data, InMaze: true}, nil
		},
		CanMove: func() bool { return true },
		Move:    func(Point) { moved <- struct{}{} },
	}
	done := make(chan error, 1)
	go func() { done <- runner.Run(cancel, StairUp, nil) }()
	time.Sleep(10 * time.Millisecond)
	select {
	case <-moved:
		t.Fatal("runner clicked while the position was between grid cells")
	default:
	}
	close(cancel)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

func TestRunnerExploresUntilRequestedStairIsRevealed(t *testing.T) {
	data := MapData{
		Width:    3,
		Height:   1,
		Known:    []bool{true, true, false},
		Walkable: []bool{true, true, false},
		Stairs:   []Stair{{Type: StairPassage}},
	}
	var mu sync.Mutex
	snapshot := MovementSnapshot{MapCode: 699, PositionSettled: true, Map: data, InMaze: true}
	var moves []Point
	var statuses []string
	runner := Runner{
		PollInterval: time.Millisecond,
		StepTimeout:  20 * time.Millisecond,
		ReadSnapshot: func() (MovementSnapshot, error) {
			mu.Lock()
			defer mu.Unlock()
			return snapshot, nil
		},
		CanMove: func() bool { return true },
		Move: func(delta Point) {
			mu.Lock()
			defer mu.Unlock()
			moves = append(moves, delta)
			switch len(moves) {
			case 1:
				snapshot.Position = Point{East: 1}
				snapshot.Map.Known[2] = true
				snapshot.Map.Walkable[2] = true
				snapshot.Map.Stairs = append(snapshot.Map.Stairs, Stair{East: 2, Type: StairUp})
			case 2:
				snapshot.Position = Point{East: 2}
				snapshot.MapCode = 1401
				snapshot.InMaze = false
			default:
				t.Fatalf("unexpected extra move: %v", moves)
			}
		},
	}
	if err := runner.Run(make(chan struct{}), StairUp, func(status string) {
		mu.Lock()
		statuses = append(statuses, status)
		mu.Unlock()
	}); err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	defer mu.Unlock()
	if !reflect.DeepEqual(moves, []Point{{East: 1}, {East: 1}}) {
		t.Fatalf("moves = %v, want exploration then Up stair", moves)
	}
	if !containsStatus(statuses, StatusExploring) {
		t.Fatalf("statuses = %v, want %q", statuses, StatusExploring)
	}
	if statuses[len(statuses)-1] != StatusCompleted {
		t.Fatalf("last status = %q, want %q", statuses[len(statuses)-1], StatusCompleted)
	}
}

func TestBuildSegmentCombinesOnlyMatchingDirections(t *testing.T) {
	data := walkableMap(6, 4)
	straight := []Point{{}, {East: 1}, {East: 2}, {East: 3}, {East: 4}, {East: 5}}
	target, direction, segment := buildSegment(data, straight, 4)
	if target != (Point{East: 4}) || direction != (Point{East: 1}) || len(segment) != 5 {
		t.Fatalf("straight segment = target %v direction %v path %v", target, direction, segment)
	}

	turning := []Point{{}, {East: 1, South: 1}, {East: 2, South: 2}, {East: 3, South: 2}}
	target, direction, segment = buildSegment(data, turning, 4)
	if target != (Point{East: 2, South: 2}) || direction != (Point{East: 1, South: 1}) || len(segment) != 3 {
		t.Fatalf("turning segment = target %v direction %v path %v", target, direction, segment)
	}

	diagonal := []Point{{}, {East: 1, South: 1}, {East: 2, South: 2}, {East: 3, South: 3}, {East: 4, South: 4}, {East: 5, South: 5}}
	target, direction, segment = buildSegment(data, diagonal, 8)
	if target != (Point{East: 4, South: 4}) || direction != (Point{East: 1, South: 1}) || len(segment) != 5 {
		t.Fatalf("diagonal segment = target %v direction %v path %v, want four safe steps", target, direction, segment)
	}
}

func TestBuildSegmentKeepsValidatedStraightPathNearObjectsAndWalls(t *testing.T) {
	path := []Point{{East: 1, South: 1}, {East: 2, South: 1}, {East: 3, South: 1}, {East: 4, South: 1}}

	objectMap := walkableMap(6, 3)
	objectMap.Objects = make([]uint16, objectMap.Width*objectMap.Height)
	objectMap.Objects[2*objectMap.Width+2] = 15080
	target, _, segment := buildSegment(objectMap, path, 4)
	if target != path[3] || len(segment) != 4 {
		t.Fatalf("passable-object segment = target %v path %v, want full segment", target, segment)
	}

	wallMap := walkableMap(6, 3)
	wallMap.Walkable[2*wallMap.Width+2] = false
	target, _, segment = buildSegment(wallMap, path, 4)
	if target != path[3] || len(segment) != 4 {
		t.Fatalf("wall-adjacent segment = target %v path %v, want validated full segment", target, segment)
	}
}

func TestRunnerStepsOffSelectedStairThenReenters(t *testing.T) {
	data := walkableMap(2, 1)
	data.Stairs = []Stair{{Type: StairDown}}
	snapshot := MovementSnapshot{MapCode: 699, PositionSettled: true, Map: data, InMaze: true}
	var moves []Point
	runner := Runner{
		PollInterval: time.Millisecond,
		ReadSnapshot: func() (MovementSnapshot, error) { return snapshot, nil },
		CanMove:      func() bool { return true },
		Move: func(delta Point) {
			moves = append(moves, delta)
			switch len(moves) {
			case 1:
				snapshot.Position = Point{East: 1}
			case 2:
				snapshot.Position = Point{}
				snapshot.MapCode = 1401
				snapshot.InMaze = false
			default:
				t.Fatalf("unexpected extra move: %v", moves)
			}
		},
	}
	if err := runner.Run(make(chan struct{}), StairDown, nil); err != nil {
		t.Fatal(err)
	}
	want := []Point{{East: 1}, {East: -1}}
	if !reflect.DeepEqual(moves, want) {
		t.Fatalf("moves = %v, want step off and reenter %v", moves, want)
	}
}

func TestRunnerAutomaticallyRetriesAfterTransientBlock(t *testing.T) {
	data := walkableMap(2, 1)
	data.Stairs = []Stair{{East: 1, Type: StairUp}}
	snapshot := MovementSnapshot{MapCode: 699, PositionSettled: true, Map: data, InMaze: true}
	var moves []Point
	runner := Runner{
		PollInterval: time.Millisecond,
		StepTimeout:  2 * time.Millisecond,
		RetryLimit:   1,
		ReadSnapshot: func() (MovementSnapshot, error) { return snapshot, nil },
		CanMove:      func() bool { return true },
		Move: func(delta Point) {
			moves = append(moves, delta)
			if len(moves) == 2 {
				snapshot.Position = Point{East: 1}
				snapshot.MapCode = 1401
				snapshot.InMaze = false
			}
		},
	}
	if err := runner.Run(make(chan struct{}), StairUp, nil); err != nil {
		t.Fatal(err)
	}
	want := []Point{{East: 1}, {East: 1}}
	if !reflect.DeepEqual(moves, want) {
		t.Fatalf("moves = %v, want one automatic clean retry %v", moves, want)
	}
}

func TestRunnerAvoidsCurrentMonsterWhenAnotherRouteExists(t *testing.T) {
	data := walkableMap(3, 2)
	data.Monsters = make([]bool, data.Width*data.Height)
	data.Monsters[1] = true
	data.Stairs = []Stair{{East: 2, Type: StairUp}}
	snapshot := MovementSnapshot{MapCode: 699, PositionSettled: true, Map: data, InMaze: true}
	var moves []Point
	runner := Runner{
		PollInterval: time.Millisecond,
		ReadSnapshot: func() (MovementSnapshot, error) { return snapshot, nil },
		CanMove:      func() bool { return true },
		Move: func(delta Point) {
			moves = append(moves, delta)
			snapshot.MapCode = 1401
			snapshot.InMaze = false
		},
	}
	if err := runner.Run(make(chan struct{}), StairUp, nil); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(moves, []Point{{East: 1, South: 1}}) {
		t.Fatalf("moves = %v, want route around current monster", moves)
	}
}

func TestRunnerFallsBackThroughMonsterInOnlyCorridor(t *testing.T) {
	data := walkableMap(3, 1)
	data.Monsters = []bool{false, true, false}
	data.Stairs = []Stair{{East: 2, Type: StairUp}}
	snapshot := MovementSnapshot{MapCode: 699, PositionSettled: true, Map: data, InMaze: true}
	var moves []Point
	runner := Runner{
		PollInterval: time.Millisecond,
		ReadSnapshot: func() (MovementSnapshot, error) { return snapshot, nil },
		CanMove:      func() bool { return true },
		Move: func(delta Point) {
			moves = append(moves, delta)
			snapshot.MapCode = 1401
			snapshot.InMaze = false
		},
	}
	if err := runner.Run(make(chan struct{}), StairUp, nil); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(moves, []Point{{East: 2}}) {
		t.Fatalf("moves = %v, want fallback through only corridor", moves)
	}
}

func TestRunnerSendsOneEightCellStraightWaypoint(t *testing.T) {
	data := walkableMap(9, 1)
	data.Stairs = []Stair{{East: 8, Type: StairUp}}
	snapshot := MovementSnapshot{MapCode: 699, PositionSettled: true, Map: data, InMaze: true}
	var moves []Point
	runner := Runner{
		PollInterval: time.Millisecond,
		ReadSnapshot: func() (MovementSnapshot, error) { return snapshot, nil },
		CanMove:      func() bool { return true },
		Move: func(delta Point) {
			moves = append(moves, delta)
			snapshot.Position = Point{East: 8}
			snapshot.MapCode = 1401
			snapshot.InMaze = false
		},
	}
	if err := runner.Run(make(chan struct{}), StairUp, nil); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(moves, []Point{{East: 8}}) {
		t.Fatalf("moves = %v, want one eight-cell waypoint", moves)
	}
}

func TestRunnerQuicklyReplansAroundStalledSegment(t *testing.T) {
	data := walkableMap(3, 2)
	data.Stairs = []Stair{{East: 2, Type: StairUp}}
	snapshot := MovementSnapshot{MapCode: 699, PositionSettled: true, Map: data, InMaze: true}
	var moves []Point
	runner := Runner{
		PollInterval: time.Millisecond,
		StepTimeout:  3 * time.Millisecond,
		RetryLimit:   2,
		ReadSnapshot: func() (MovementSnapshot, error) { return snapshot, nil },
		CanMove:      func() bool { return true },
		Move: func(delta Point) {
			moves = append(moves, delta)
			switch len(moves) {
			case 1, 2:
				// The direct east segment is blocked and makes no progress.
			case 3:
				snapshot.Position = Point{East: 1, South: 1}
			case 4:
				snapshot.Position = Point{East: 2}
				snapshot.MapCode = 1401
				snapshot.InMaze = false
			default:
				t.Fatalf("unexpected extra move: %v", moves)
			}
		},
	}
	if err := runner.Run(make(chan struct{}), StairUp, nil); err != nil {
		t.Fatal(err)
	}
	want := []Point{{East: 2}, {East: 2}, {East: 1, South: 1}, {East: 1, South: -1}}
	if !reflect.DeepEqual(moves, want) {
		t.Fatalf("moves = %v, want %v", moves, want)
	}
}

func containsStatus(statuses []string, want string) bool {
	for _, status := range statuses {
		if status == want {
			return true
		}
	}
	return false
}

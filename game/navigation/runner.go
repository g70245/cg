package navigation

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

const (
	StatusLoading       = "Loading route data..."
	StatusMoving        = "Moving"
	StatusExploring     = "Exploring"
	StatusPaused        = "Paused for battle"
	StatusReplanning    = "Replanning"
	StatusCompleted     = "Maze exit reached"
	StatusUnavailable   = "Map data unavailable."
	StatusBlocked       = "Movement blocked"
	StatusNoExploration = "No unexplored route"
	StatusNotInMaze     = "Not in a maze"
	defaultPollInterval = 100 * time.Millisecond
	defaultStepTimeout  = 300 * time.Millisecond
	defaultRetryLimit   = 1
	defaultSegmentSteps = 8
)

var errRunnerConfiguration = errors.New("navigation runner is not configured")

type MovementSnapshot struct {
	MapCode         uint32
	Position        Point
	PositionSettled bool
	ExactEast       float64
	ExactSouth      float64
	Map             MapData
	InMaze          bool
}

type Runner struct {
	ReadSnapshot func() (MovementSnapshot, error)
	CanMove      func() bool
	Move         func(Point)
	State        *TraversalState
	PollInterval time.Duration
	StepTimeout  time.Duration
	RetryLimit   int
	SegmentSteps int
}

type TraversalState struct {
	mu            sync.Mutex
	initialized   bool
	mapCode       uint32
	entryPassages map[Point]bool
}

func (state *TraversalState) BeginFloor(mapCode uint32, position Point, stairs []Stair) map[Point]bool {
	state.mu.Lock()
	defer state.mu.Unlock()
	if !state.initialized || state.mapCode != mapCode {
		state.initialized = true
		state.mapCode = mapCode
		state.entryPassages = nearbyPassagePoints(stairs, position)
	}
	return clonePointSet(state.entryPassages)
}

func (state *TraversalState) Reset() {
	state.mu.Lock()
	state.initialized = false
	state.mapCode = 0
	state.entryPassages = nil
	state.mu.Unlock()
}

type pendingStep struct {
	target       Point
	path         []Point
	lastSettled  Point
	lastPoint    Point
	lastEast     float64
	lastSouth    float64
	lastProgress time.Time
}

func (runner Runner) Run(cancel <-chan struct{}, targetType StairType, report func(string)) error {
	if runner.ReadSnapshot == nil || runner.CanMove == nil || runner.Move == nil {
		return errRunnerConfiguration
	}
	if targetType != StairUp && targetType != StairDown {
		return fmt.Errorf("unsupported destination %q", targetType)
	}
	if report == nil {
		report = func(string) {}
	}
	interval := runner.PollInterval
	if interval <= 0 {
		interval = defaultPollInterval
	}
	stepTimeout := runner.StepTimeout
	if stepTimeout <= 0 {
		stepTimeout = defaultStepTimeout
	}
	retryLimit := runner.RetryLimit
	if retryLimit <= 0 {
		retryLimit = defaultRetryLimit
	}
	segmentSteps := runner.SegmentSteps
	if segmentSteps <= 0 {
		segmentSteps = defaultSegmentSteps
	}

	report(StatusLoading)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	blocked := make(map[Point]bool)
	failures := make(map[Point]int)
	completedFrontiers := make(map[Point]bool)
	heading := Point{}
	var currentCode uint32
	var initialized bool
	var entryPassages map[Point]bool
	var pending *pendingStep
	arrivedTarget := false
	recoveryUsed := false
	lastStatus := ""
	reportStatus := func(status string) {
		if status != lastStatus {
			lastStatus = status
			report(status)
		}
	}

	for {
		select {
		case <-cancel:
			return nil
		case <-ticker.C:
		}

		snapshot, err := runner.ReadSnapshot()
		if err != nil {
			reportStatus(StatusUnavailable)
			continue
		}
		if !snapshot.InMaze {
			if runner.State != nil {
				runner.State.Reset()
			}
			if initialized && snapshot.MapCode != currentCode {
				reportStatus(StatusCompleted)
				return nil
			}
			reportStatus(StatusNotInMaze)
			return nil
		}
		if !initialized || snapshot.MapCode != currentCode {
			currentCode = snapshot.MapCode
			initialized = true
			pending = nil
			blocked = make(map[Point]bool)
			failures = make(map[Point]int)
			completedFrontiers = make(map[Point]bool)
			heading = Point{}
			arrivedTarget = false
			recoveryUsed = false
			if runner.State != nil {
				entryPassages = runner.State.BeginFloor(currentCode, snapshot.Position, snapshot.Map.Stairs)
			} else {
				entryPassages = nearbyPassagePoints(snapshot.Map.Stairs, snapshot.Position)
			}
			reportStatus(StatusReplanning)
		}

		if !runner.CanMove() {
			if pending != nil {
				pending.lastProgress = time.Now()
			}
			reportStatus(StatusPaused)
			continue
		}

		if pending != nil {
			now := time.Now()
			progressed := snapshot.Position != pending.lastPoint ||
				mathAbs(snapshot.ExactEast-pending.lastEast) > 0.01 ||
				mathAbs(snapshot.ExactSouth-pending.lastSouth) > 0.01
			if progressed {
				recoveryUsed = false
				pending.lastPoint = snapshot.Position
				pending.lastEast = snapshot.ExactEast
				pending.lastSouth = snapshot.ExactSouth
				pending.lastProgress = now
			}
			if snapshot.PositionSettled {
				if snapshot.Position == pending.target {
					delete(failures, pending.target)
					arrivedTarget = stairPointHasType(snapshot.Map.Stairs, pending.target, targetType)
					pending = nil
					reportStatus(StatusReplanning)
				} else if pointOnPath(pending.path, snapshot.Position) {
					pending.lastSettled = snapshot.Position
				} else {
					pending = nil
					reportStatus(StatusReplanning)
				}
			}
			if pending != nil {
				if progressed || now.Sub(pending.lastProgress) < stepTimeout {
					continue
				}
				next, ok := nextPathPoint(pending.path, pending.lastSettled)
				if !ok {
					next = pending.target
				}
				failures[next]++
				if failures[next] >= retryLimit {
					blocked[next] = true
				}
				pending = nil
				reportStatus(StatusReplanning)
			}
		}
		if !snapshot.PositionSettled {
			continue
		}

		monsters := monsterPoints(snapshot.Map)
		pathBlocks := mergeBlocks(blocked, blockedTransitions(snapshot.Map, targetType))
		path, _, targetFound := FindPath(snapshot.Map, snapshot.Position, targetType, mergeBlocks(pathBlocks, monsters))
		if !targetFound && len(monsters) > 0 {
			path, _, targetFound = FindPath(snapshot.Map, snapshot.Position, targetType, pathBlocks)
		}
		passageFound := false
		if !targetFound {
			passageBlocks := mergeBlocks(blocked, entryPassages, blockedTransitions(snapshot.Map, StairPassage))
			path, _, passageFound = FindPath(snapshot.Map, snapshot.Position, StairPassage, mergeBlocks(passageBlocks, monsters))
			if !passageFound && len(monsters) > 0 {
				path, _, passageFound = FindPath(snapshot.Map, snapshot.Position, StairPassage, passageBlocks)
			}
		}
		exploring := false
		if !targetFound && !passageFound {
			path, exploring = FindExplorationPath(snapshot.Map, snapshot.Position, mergeBlocks(pathBlocks, monsters), completedFrontiers, heading)
			if !exploring && len(monsters) > 0 {
				path, exploring = FindExplorationPath(snapshot.Map, snapshot.Position, pathBlocks, completedFrontiers, heading)
			}
			if !exploring {
				if !recoveryUsed && (len(blocked) > 0 || len(completedFrontiers) > 0) {
					blocked = make(map[Point]bool)
					failures = make(map[Point]int)
					completedFrontiers = make(map[Point]bool)
					heading = Point{}
					recoveryUsed = true
					reportStatus(StatusReplanning)
					continue
				}
				if len(blocked) > 0 {
					reportStatus(StatusBlocked)
				} else {
					reportStatus(StatusNoExploration)
				}
				continue
			}
		}
		if len(path) < 2 {
			if exploring {
				completedFrontiers[snapshot.Position] = true
				reportStatus(StatusExploring)
			} else if targetFound && !arrivedTarget {
				stepOff, ok := targetStepOffPoint(snapshot.Map, snapshot.Position, pathBlocks)
				if ok {
					path = []Point{snapshot.Position, stepOff}
				} else {
					reportStatus(StatusBlocked)
					continue
				}
			} else {
				reportStatus(StatusMoving)
				continue
			}
		}

		waypoint, direction, segment := buildSegment(snapshot.Map, path, segmentSteps)
		delta := Point{East: waypoint.East - snapshot.Position.East, South: waypoint.South - snapshot.Position.South}
		select {
		case <-cancel:
			return nil
		default:
		}
		runner.Move(delta)
		heading = direction
		pending = &pendingStep{
			target:       waypoint,
			path:         segment,
			lastSettled:  snapshot.Position,
			lastPoint:    snapshot.Position,
			lastEast:     snapshot.ExactEast,
			lastSouth:    snapshot.ExactSouth,
			lastProgress: time.Now(),
		}
		if exploring {
			reportStatus(StatusExploring)
		} else {
			reportStatus(StatusMoving)
		}
	}
}

func buildSegment(_ MapData, path []Point, maxSteps int) (Point, Point, []Point) {
	direction := Point{East: path[1].East - path[0].East, South: path[1].South - path[0].South}
	if direction.East != 0 && direction.South != 0 && maxSteps > 4 {
		maxSteps = 4
	}
	end := 1
	for end+1 < len(path) && end < maxSteps {
		nextDirection := Point{East: path[end+1].East - path[end].East, South: path[end+1].South - path[end].South}
		if nextDirection != direction {
			break
		}
		end++
	}
	return path[end], direction, append([]Point(nil), path[:end+1]...)
}

func pointOnPath(path []Point, point Point) bool {
	for _, candidate := range path {
		if candidate == point {
			return true
		}
	}
	return false
}

func nextPathPoint(path []Point, point Point) (Point, bool) {
	for index := 0; index+1 < len(path); index++ {
		if path[index] == point {
			return path[index+1], true
		}
	}
	return Point{}, false
}

func mathAbs(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}

func mergeBlocks(blockSets ...map[Point]bool) map[Point]bool {
	merged := make(map[Point]bool)
	for _, blockSet := range blockSets {
		for point, isBlocked := range blockSet {
			if isBlocked {
				merged[point] = true
			}
		}
	}
	return merged
}

func nearbyPassagePoints(stairs []Stair, position Point) map[Point]bool {
	points := make(map[Point]bool)
	for _, stair := range stairs {
		if stair.Type != StairPassage || absInt(stair.East-position.East) > 1 || absInt(stair.South-position.South) > 1 {
			continue
		}
		points[Point{East: stair.East, South: stair.South}] = true
	}
	return points
}

func monsterPoints(data MapData) map[Point]bool {
	points := make(map[Point]bool)
	for index, occupied := range data.Monsters {
		if occupied {
			points[pointAt(data, index)] = true
		}
	}
	return points
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func clonePointSet(points map[Point]bool) map[Point]bool {
	cloned := make(map[Point]bool, len(points))
	for point, included := range points {
		if included {
			cloned[point] = true
		}
	}
	return cloned
}

func stairPointHasType(stairs []Stair, point Point, stairType StairType) bool {
	for _, stair := range stairs {
		if stair.Type == stairType && stair.East == point.East && stair.South == point.South {
			return true
		}
	}
	return false
}

func targetStepOffPoint(data MapData, current Point, blocked map[Point]bool) (Point, bool) {
	stairs := make(map[Point]bool, len(data.Stairs))
	for _, stair := range data.Stairs {
		stairs[Point{East: stair.East, South: stair.South}] = true
	}
	for _, direction := range orderedDirections(Point{}) {
		candidate := Point{East: current.East + direction.East, South: current.South + direction.South}
		if stairs[candidate] || !canStep(data, current, candidate, blocked) {
			continue
		}
		return candidate, true
	}
	return Point{}, false
}

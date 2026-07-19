package navigation

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

const (
	StatusLoading               = "Loading route data..."
	StatusMoving                = "Moving"
	StatusExploring             = "Exploring"
	StatusPaused                = "Paused for battle"
	StatusReplanning            = "Replanning"
	StatusCompleted             = "Maze exit reached"
	StatusUnavailable           = "Map data unavailable."
	StatusBlocked               = "Movement blocked"
	StatusNoExploration         = "No unexplored route"
	StatusNotInMaze             = "Not in a maze"
	StatusVerification          = "Verification required"
	defaultPollInterval         = 50 * time.Millisecond
	defaultVerificationInterval = 500 * time.Millisecond
	defaultStepTimeout          = 300 * time.Millisecond
	defaultRetryLimit           = 1
	defaultSegmentSteps         = 6
)

var (
	ErrMapUpdating         = errors.New("navigation map data is updating")
	errRunnerConfiguration = errors.New("navigation runner is not configured")
)

type MovementSnapshot struct {
	MapCode         uint32
	MapName         string
	Position        Point
	PositionSettled bool
	ExactEast       float64
	ExactSouth      float64
	Map             MapData
	InMaze          bool
}

type Runner struct {
	ReadSnapshot          func() (MovementSnapshot, error)
	CanMove               func() bool
	Move                  func(Point)
	State                 *TraversalState
	PollInterval          time.Duration
	StepTimeout           time.Duration
	RetryLimit            int
	SegmentSteps          int
	VerificationTriggered func() bool
	OnVerification        func()
	VerificationInterval  time.Duration
}

type TraversalState struct {
	mu               sync.Mutex
	initialized      bool
	mapCode          uint32
	mapName          string
	entryTransitions map[Point]bool
	entryIntent      StairType
	directionalEntry bool
	attemptedExits   map[floorIdentity]map[Point]bool
}

type floorIdentity struct {
	mapCode uint32
	mapName string
}

func (state *TraversalState) BeginFloor(mapCode uint32, mapName string, position Point, stairs []Stair, targetType StairType, includeAllTypes bool) map[Point]bool {
	state.mu.Lock()
	defer state.mu.Unlock()
	if !state.initialized || state.mapCode != mapCode || state.mapName != mapName {
		state.initialized = true
		state.mapCode = mapCode
		state.mapName = mapName
		state.entryTransitions = nearbyEntryTransitionPoints(stairs, position, includeAllTypes)
		state.entryIntent = targetType
		state.directionalEntry = includeAllTypes
	}
	if state.directionalEntry && state.entryIntent != targetType {
		return map[Point]bool{}
	}
	return clonePointSet(state.entryTransitions)
}

func (state *TraversalState) Reset() {
	state.mu.Lock()
	state.initialized = false
	state.mapCode = 0
	state.mapName = ""
	state.entryTransitions = nil
	state.entryIntent = ""
	state.directionalEntry = false
	state.attemptedExits = nil
	state.mu.Unlock()
}

func (state *TraversalState) AttemptedExits(mapCode uint32, mapName string) map[Point]bool {
	state.mu.Lock()
	defer state.mu.Unlock()
	return clonePointSet(state.attemptedExits[floorIdentity{mapCode: mapCode, mapName: mapName}])
}

func (state *TraversalState) MarkExitAttempt(mapCode uint32, mapName string, point Point) {
	state.mu.Lock()
	defer state.mu.Unlock()
	if state.attemptedExits == nil {
		state.attemptedExits = make(map[floorIdentity]map[Point]bool)
	}
	identity := floorIdentity{mapCode: mapCode, mapName: mapName}
	if state.attemptedExits[identity] == nil {
		state.attemptedExits[identity] = make(map[Point]bool)
	}
	state.attemptedExits[identity][point] = true
}

func (state *TraversalState) ClearExitAttempts() {
	state.mu.Lock()
	state.attemptedExits = nil
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
	probe        *exitProbe
}

type exitProbe struct {
	mapCode uint32
	mapName string
	point   Point
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
	verificationInterval := runner.VerificationInterval
	if verificationInterval <= 0 {
		verificationInterval = defaultVerificationInterval
	}

	report(StatusLoading)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	blocked := make(map[Point]bool)
	failures := make(map[Point]int)
	completedFrontiers := make(map[Point]bool)
	localExitAttempts := make(map[floorIdentity]map[Point]bool)
	heading := Point{}
	var currentCode uint32
	var currentName string
	var initialized bool
	var entryTransitions map[Point]bool
	var selectedEntryOverride map[Point]bool
	var pending *pendingStep
	var probeInFlight *exitProbe
	var forcedProbe *exitProbe
	arrivedTarget := false
	recoveryUsed := false
	lastStatus := ""
	var lastVerificationCheck time.Time
	reportStatus := func(status string) {
		if status != lastStatus {
			lastStatus = status
			report(status)
		}
	}
	attemptedExits := func(mapCode uint32, mapName string) map[Point]bool {
		if runner.State != nil {
			return runner.State.AttemptedExits(mapCode, mapName)
		}
		return clonePointSet(localExitAttempts[floorIdentity{mapCode: mapCode, mapName: mapName}])
	}
	markExitAttempt := func(probe *exitProbe) {
		if runner.State != nil {
			runner.State.MarkExitAttempt(probe.mapCode, probe.mapName, probe.point)
			return
		}
		identity := floorIdentity{mapCode: probe.mapCode, mapName: probe.mapName}
		if localExitAttempts[identity] == nil {
			localExitAttempts[identity] = make(map[Point]bool)
		}
		localExitAttempts[identity][probe.point] = true
	}

	for {
		select {
		case <-cancel:
			return nil
		case <-ticker.C:
		}
		now := time.Now()
		if runner.VerificationTriggered != nil && (lastVerificationCheck.IsZero() || now.Sub(lastVerificationCheck) >= verificationInterval) {
			lastVerificationCheck = now
			if runner.VerificationTriggered() {
				reportStatus(StatusVerification)
				if runner.OnVerification != nil {
					runner.OnVerification()
				}
				return nil
			}
		}

		snapshot, err := runner.ReadSnapshot()
		if err != nil {
			if errors.Is(err, ErrMapUpdating) {
				reportStatus(StatusLoading)
			} else {
				reportStatus(StatusUnavailable)
			}
			continue
		}
		if !snapshot.InMaze {
			if runner.State != nil {
				runner.State.Reset()
			}
			if initialized && (snapshot.MapCode != currentCode || snapshot.MapName != currentName) {
				reportStatus(StatusCompleted)
				return nil
			}
			reportStatus(StatusNotInMaze)
			return nil
		}
		if !initialized || snapshot.MapCode != currentCode || snapshot.MapName != currentName {
			firstSnapshot := !initialized
			floorChanged := initialized && (snapshot.MapCode != currentCode || snapshot.MapName != currentName)
			arrivedFromProbe := floorChanged && probeInFlight != nil
			if arrivedFromProbe {
				markExitAttempt(probeInFlight)
				probeInFlight = nil
			}
			forcedProbe = nil
			currentCode = snapshot.MapCode
			currentName = snapshot.MapName
			initialized = true
			pending = nil
			blocked = make(map[Point]bool)
			failures = make(map[Point]int)
			completedFrontiers = make(map[Point]bool)
			heading = Point{}
			arrivedTarget = false
			recoveryUsed = false
			if runner.State != nil {
				entryTransitions = runner.State.BeginFloor(currentCode, currentName, snapshot.Position, snapshot.Map.Stairs, targetType, floorChanged)
			} else {
				entryTransitions = nearbyEntryTransitionPoints(snapshot.Map.Stairs, snapshot.Position, floorChanged)
			}
			selectedEntryOverride = nil
			if (firstSnapshot || arrivedFromProbe) && entryTransitions[snapshot.Position] && stairPointHasType(snapshot.Map.Stairs, snapshot.Position, targetType) {
				selectedEntryOverride = map[Point]bool{snapshot.Position: true}
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
					if pending.probe != nil {
						probeInFlight = nil
					}
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
				if pending.probe != nil {
					probeInFlight = nil
				}
				pending = nil
				reportStatus(StatusReplanning)
			}
		}
		if !snapshot.PositionSettled {
			continue
		}
		if probeInFlight != nil {
			if snapshot.Position == probeInFlight.point {
				reportStatus(StatusMoving)
				continue
			}
			probeInFlight = nil
		}

		monsters := monsterPoints(snapshot.Map)
		pathBlocks := mergeBlocks(blocked, entryTransitions, blockedTransitions(snapshot.Map, targetType))
		for point := range selectedEntryOverride {
			delete(pathBlocks, point)
		}
		routeBlocks := pathBlocks
		path, _, targetFound := FindPath(snapshot.Map, snapshot.Position, targetType, mergeBlocks(pathBlocks, monsters))
		if !targetFound && len(monsters) > 0 {
			path, _, targetFound = FindPath(snapshot.Map, snapshot.Position, targetType, pathBlocks)
		}
		passageFound := false
		if !targetFound {
			passageBlocks := mergeBlocks(blocked, entryTransitions, blockedTransitions(snapshot.Map, StairPassage))
			path, _, passageFound = FindPath(snapshot.Map, snapshot.Position, StairPassage, mergeBlocks(passageBlocks, monsters))
			if !passageFound && len(monsters) > 0 {
				path, _, passageFound = FindPath(snapshot.Map, snapshot.Position, StairPassage, passageBlocks)
			}
		}
		probeFound := false
		probeTarget := Stair{}
		if !targetFound && !passageFound {
			probeType := oppositeStairType(targetType)
			if stairTypeCount(snapshot.Map.Stairs, probeType) >= 2 {
				probeBlocks := mergeBlocks(blocked, entryTransitions, attemptedExits(currentCode, currentName), blockedTransitions(snapshot.Map, probeType))
				if forcedProbe != nil && forcedProbe.mapCode == currentCode && forcedProbe.mapName == currentName {
					for _, stair := range snapshot.Map.Stairs {
						point := Point{East: stair.East, South: stair.South}
						if stair.Type == probeType && point != forcedProbe.point {
							probeBlocks[point] = true
						}
					}
				}
				path, probeTarget, probeFound = FindPath(snapshot.Map, snapshot.Position, probeType, mergeBlocks(probeBlocks, monsters))
				if !probeFound && len(monsters) > 0 {
					path, probeTarget, probeFound = FindPath(snapshot.Map, snapshot.Position, probeType, probeBlocks)
				}
				if probeFound {
					routeBlocks = probeBlocks
					if forcedProbe == nil {
						forcedProbe = &exitProbe{
							mapCode: currentCode,
							mapName: currentName,
							point:   Point{East: probeTarget.East, South: probeTarget.South},
						}
					}
				}
			}
		}
		exploring := false
		if !targetFound && !passageFound && !probeFound {
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
			} else if (targetFound && !arrivedTarget) || probeFound {
				stepOff, ok := targetStepOffPoint(snapshot.Map, snapshot.Position, routeBlocks)
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
		var probe *exitProbe
		if probeFound && waypoint == (Point{East: probeTarget.East, South: probeTarget.South}) {
			probe = &exitProbe{mapCode: currentCode, mapName: currentName, point: waypoint}
			probeInFlight = probe
		}
		pending = &pendingStep{
			target:       waypoint,
			path:         segment,
			lastSettled:  snapshot.Position,
			lastPoint:    snapshot.Position,
			lastEast:     snapshot.ExactEast,
			lastSouth:    snapshot.ExactSouth,
			lastProgress: time.Now(),
			probe:        probe,
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

func nearbyEntryTransitionPoints(stairs []Stair, position Point, includeAllTypes bool) map[Point]bool {
	points := make(map[Point]bool)
	for _, stair := range stairs {
		rememberType := includeAllTypes || stair.Type == StairPassage
		if !rememberType || absInt(stair.East-position.East) > 1 || absInt(stair.South-position.South) > 1 {
			continue
		}
		points[Point{East: stair.East, South: stair.South}] = true
	}
	return points
}

func oppositeStairType(stairType StairType) StairType {
	if stairType == StairUp {
		return StairDown
	}
	return StairUp
}

func stairTypeCount(stairs []Stair, stairType StairType) int {
	count := 0
	for _, stair := range stairs {
		if stair.Type == stairType {
			count++
		}
	}
	return count
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

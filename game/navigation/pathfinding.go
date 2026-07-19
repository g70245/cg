package navigation

func FindPath(data MapData, start Point, targetType StairType, blocked map[Point]bool) ([]Point, Stair, bool) {
	targets := stairTargets(data.Stairs, targetType)
	if len(targets) == 0 || !data.IsKnown(start) {
		return nil, Stair{}, false
	}

	cellCount := data.Width * data.Height
	parents := make([]int, cellCount)
	for index := range parents {
		parents[index] = -2
	}
	startIndex := pointIndex(data, start)
	parents[startIndex] = -1
	queue := make([]Point, 1, cellCount)
	queue[0] = start

	directions := orderedDirections(Point{})
	for head := 0; head < len(queue); head++ {
		current := queue[head]
		if stair, ok := targets[current]; ok && !blocked[current] {
			return rebuildPath(data, parents, current), stair, true
		}

		for _, direction := range directions {
			next := Point{East: current.East + direction.East, South: current.South + direction.South}
			if !canStep(data, current, next, blocked) {
				continue
			}
			nextIndex := pointIndex(data, next)
			if parents[nextIndex] != -2 {
				continue
			}
			parents[nextIndex] = pointIndex(data, current)
			queue = append(queue, next)
		}
	}

	return nil, Stair{}, false
}

func FindExplorationPath(data MapData, start Point, blocked, completed map[Point]bool, preferred Point) ([]Point, bool) {
	if !data.IsKnown(start) {
		return nil, false
	}
	cellCount := data.Width * data.Height
	parents := make([]int, cellCount)
	firstDirections := make([]Point, cellCount)
	for index := range parents {
		parents[index] = -2
	}
	startIndex := pointIndex(data, start)
	parents[startIndex] = -1
	queue := make([]Point, 1, cellCount)
	queue[0] = start

	directions := orderedDirections(preferred)
	var bestPath []Point
	bestRank := int(^uint(0) >> 1)
	for head := 0; head < len(queue); head++ {
		current := queue[head]
		currentIndex := pointIndex(data, current)
		if !completed[current] && !blocked[current] && isFrontier(data, current) {
			path := rebuildPath(data, parents, current)
			rank := explorationRank(firstDirections[currentIndex], preferred)
			if betterExplorationCandidate(path, rank, bestPath, bestRank) {
				bestPath = path
				bestRank = rank
			}
		}

		for _, direction := range directions {
			next := Point{East: current.East + direction.East, South: current.South + direction.South}
			if !canStep(data, current, next, blocked) {
				continue
			}
			nextIndex := pointIndex(data, next)
			if parents[nextIndex] != -2 {
				continue
			}
			parents[nextIndex] = currentIndex
			if current == start {
				firstDirections[nextIndex] = direction
			} else {
				firstDirections[nextIndex] = firstDirections[currentIndex]
			}
			queue = append(queue, next)
		}
	}
	return bestPath, len(bestPath) > 0
}

func blockedTransitions(data MapData, targetType StairType) map[Point]bool {
	blocked := make(map[Point]bool)
	for _, stair := range data.Stairs {
		if stair.Type != targetType {
			blocked[Point{East: stair.East, South: stair.South}] = true
		}
	}
	return blocked
}

func isFrontier(data MapData, point Point) bool {
	for _, direction := range orderedDirections(Point{}) {
		next := Point{East: point.East + direction.East, South: point.South + direction.South}
		if next.East < 0 || next.South < 0 || next.East >= data.Width || next.South >= data.Height {
			continue
		}
		if !data.IsKnown(next) {
			return true
		}
	}
	return false
}

func orderedDirections(preferred Point) []Point {
	all := []Point{
		{East: 1},
		{East: 1, South: 1},
		{South: 1},
		{East: -1, South: 1},
		{East: -1},
		{East: -1, South: -1},
		{South: -1},
		{East: 1, South: -1},
	}
	if preferred == (Point{}) {
		return all
	}
	ordered := []Point{preferred}
	for _, direction := range all {
		if direction != preferred && direction != reverse(preferred) {
			ordered = append(ordered, direction)
		}
	}
	return append(ordered, reverse(preferred))
}

func canStep(data MapData, from, to Point, blocked map[Point]bool) bool {
	deltaEast := to.East - from.East
	deltaSouth := to.South - from.South
	if deltaEast < -1 || deltaEast > 1 || deltaSouth < -1 || deltaSouth > 1 || deltaEast == 0 && deltaSouth == 0 {
		return false
	}
	if !data.IsWalkable(to) || blocked[to] {
		return false
	}
	if deltaEast == 0 || deltaSouth == 0 {
		return true
	}
	eastSide := Point{East: from.East + deltaEast, South: from.South}
	southSide := Point{East: from.East, South: from.South + deltaSouth}
	return data.IsWalkable(eastSide) || data.IsWalkable(southSide)
}

func explorationRank(first, preferred Point) int {
	if preferred == (Point{}) || first == (Point{}) {
		return 1
	}
	if first == preferred {
		return 0
	}
	if first == reverse(preferred) {
		return 2
	}
	return 1
}

func betterExplorationCandidate(path []Point, rank int, best []Point, bestRank int) bool {
	if len(best) == 0 || rank != bestRank {
		return len(best) == 0 || rank < bestRank
	}
	if len(path) != len(best) {
		return len(path) < len(best)
	}
	target := path[len(path)-1]
	bestTarget := best[len(best)-1]
	if target.East != bestTarget.East {
		return target.East < bestTarget.East
	}
	return target.South < bestTarget.South
}

func reverse(direction Point) Point {
	return Point{East: -direction.East, South: -direction.South}
}

func stairTargets(stairs []Stair, targetType StairType) map[Point]Stair {
	targets := make(map[Point]Stair)
	for _, stair := range stairs {
		if stair.Type == targetType {
			targets[Point{East: stair.East, South: stair.South}] = stair
		}
	}
	return targets
}

func pointIndex(data MapData, point Point) int {
	return point.South*data.Width + point.East
}

func pointAt(data MapData, index int) Point {
	return Point{East: index % data.Width, South: index / data.Width}
}

func rebuildPath(data MapData, parents []int, target Point) []Point {
	reversed := make([]Point, 0)
	for index := pointIndex(data, target); index >= 0; index = parents[index] {
		reversed = append(reversed, pointAt(data, index))
	}
	path := make([]Point, len(reversed))
	for index := range reversed {
		path[len(reversed)-1-index] = reversed[index]
	}
	return path
}

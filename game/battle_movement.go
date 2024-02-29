package game

import (
	. "cg/internal"
	"fmt"

	"log"
	"math"
	"math/rand"

	"github.com/g70245/win"
)

const (
	RADIUS     float64 = 120
	BIAS_ANGLE float64 = 30
)

type BattleMovementMode string
type BattleMovementModes []BattleMovementMode

func (m BattleMovementModes) GetOptions() []string {
	var options []string

	for _, mode := range m {
		options = append(options, fmt.Sprint(mode))
	}

	return options
}

const (
	None                     BattleMovementMode = "None"
	DIAGONAL                 BattleMovementMode = "Diagonal"
	BIASED_DIAGONAL          BattleMovementMode = "B. Diagonal"
	REVERSED_DIAGONAL        BattleMovementMode = "Reversed Diagonal"
	BIASED_REVERSED_DIAGONAL BattleMovementMode = "B. Reversed Diagonal"
	HYBRID_DIAGONAL          BattleMovementMode = "Hybrid Diagonal"
)

var BATTLE_MOVEMENT_MODES BattleMovementModes = []BattleMovementMode{None, DIAGONAL, REVERSED_DIAGONAL, BIASED_DIAGONAL, BIASED_REVERSED_DIAGONAL, HYBRID_DIAGONAL}

type BattleMovementState struct {
	hWnd   win.HWND
	origin GamePos
	Mode   BattleMovementMode
}

func (state *BattleMovementState) Move() {

	var x, y int32
	switch state.Mode {
	case DIAGONAL:
		x, y = diagonal(RADIUS, state, false, false)
	case REVERSED_DIAGONAL:
		x, y = diagonal(RADIUS, state, true, false)
	case BIASED_DIAGONAL:
		x, y = diagonal(RADIUS, state, false, true)
	case BIASED_REVERSED_DIAGONAL:
		x, y = diagonal(RADIUS, state, true, true)
	case HYBRID_DIAGONAL:
		x, y = diagonal(RADIUS, state, rand.Intn(2) != 0, true)
	default:
		x, y = none()
	}

	log.Printf("Handle %d moves to (%d, %d)\n", state.hWnd, x, y)

	LeftClick(state.hWnd, x, y)
}

func diagonal(radius float64, state *BattleMovementState, isReverse bool, isBiasable bool) (x, y int32) {
	xOrigin := int32(GAME_WIDTH / 2)
	yOrigin := int32(GAME_HEIGHT / 2)

	var randomBiasAngle float64
	if isBiasable {
		randomBiasAngle = rand.Float64() * BIAS_ANGLE * state.nextBiasMultiplier()
	}

	directionAngle := state.nextDirectionAngleMultilpier(isReverse) * math.Pi
	rotationAngle := ((45.0+randomBiasAngle)/180.0)*math.Pi + directionAngle

	if isReverse {
		x = xOrigin + int32(radius*math.Cos(rotationAngle))
		y = yOrigin + int32(radius*math.Sin(rotationAngle))
	} else {
		x = xOrigin + int32(radius*math.Cos(rotationAngle))
		y = yOrigin + int32(radius*math.Sin(rotationAngle))*-1
	}

	return
}

func none() (int32, int32) {
	return GAME_WIDTH / 2, GAME_HEIGHT / 2
}

func (state *BattleMovementState) nextDirectionAngleMultilpier(isReverse bool) float64 {
	current := getCurrentGamePos(state.hWnd)

	check := diagonalCondition
	if isReverse {
		check = reversedDiagonalCondition
	}

	if check(state.origin, current) {
		return 0
	} else {
		return 1
	}
}

func (state *BattleMovementState) nextBiasMultiplier() float64 {
	current := getCurrentGamePos(state.hWnd)

	switch {
	case current.x > state.origin.x && current.y <= state.origin.y:
		return 1
	case current.x > state.origin.x && current.y > state.origin.y:
		return -1
	case current.x <= state.origin.x && current.y <= state.origin.y:
		return -1
	case current.x <= state.origin.x && current.y > state.origin.y:
		return 1
	default:
		return []float64{1, -1}[rand.Intn(2)]
	}
}

func diagonalCondition(origin GamePos, current GamePos) bool {
	return current.x <= origin.x
}

func reversedDiagonalCondition(origin GamePos, current GamePos) bool {
	return current.y <= origin.y
}

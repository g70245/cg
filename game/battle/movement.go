package battle

import (
	"cg/game"
	"cg/internal"
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

type MovementMode string
type MovementModes []MovementMode

func (m MovementModes) GetOptions() []string {
	var options []string

	for _, mode := range m {
		options = append(options, fmt.Sprint(mode))
	}

	return options
}

const (
	None                     MovementMode = "None"
	DIAGONAL                 MovementMode = "Diagonal"
	BIASED_DIAGONAL          MovementMode = "B. Diagonal"
	REVERSED_DIAGONAL        MovementMode = "Reversed Diagonal"
	BIASED_REVERSED_DIAGONAL MovementMode = "B. Reversed Diagonal"
	HYBRID_DIAGONAL          MovementMode = "Hybrid Diagonal"
)

var MOVEMENT_MODES MovementModes = []MovementMode{None, DIAGONAL, REVERSED_DIAGONAL, BIASED_DIAGONAL, BIASED_REVERSED_DIAGONAL, HYBRID_DIAGONAL}

type MovementState struct {
	hWnd   win.HWND
	origin game.GamePos
	Mode   MovementMode
}

func (state *MovementState) Move() {

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

	internal.LeftClick(state.hWnd, x, y)
}

func diagonal(radius float64, state *MovementState, isReverse bool, isBiasable bool) (x, y int32) {
	xOrigin := int32(game.GAME_WIDTH / 2)
	yOrigin := int32(game.GAME_HEIGHT / 2)

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
	return game.GAME_WIDTH / 2, game.GAME_HEIGHT / 2
}

func (state *MovementState) nextDirectionAngleMultilpier(isReverse bool) float64 {
	current := game.GetCurrentGamePos(state.hWnd)

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

func (state *MovementState) nextBiasMultiplier() float64 {
	current := game.GetCurrentGamePos(state.hWnd)

	switch {
	case current.X > state.origin.X && current.Y <= state.origin.Y:
		return 1
	case current.X > state.origin.X && current.Y > state.origin.Y:
		return -1
	case current.X <= state.origin.X && current.Y <= state.origin.Y:
		return -1
	case current.X <= state.origin.X && current.Y > state.origin.Y:
		return 1
	default:
		return []float64{1, -1}[rand.Intn(2)]
	}
}

func diagonalCondition(origin game.GamePos, current game.GamePos) bool {
	return current.X <= origin.X
}

func reversedDiagonalCondition(origin game.GamePos, current game.GamePos) bool {
	return current.Y <= origin.Y
}

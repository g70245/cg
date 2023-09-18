package game

import (
	sys "cg/system"
	"log"

	"math"
	"math/rand"

	. "github.com/lxn/win"
)

var (
	RADIUS     float64 = 120
	BIAS_ANGLE float64 = 30
)

type BattleMovementMode string

const (
	NONE                          = "None"
	DIAGONAL_MODE                 = "Diagonal"
	BIASED_DIAGONAL_MODE          = "B. Diagonal"
	REVERSED_DIAGONAL_MODE        = "Reversed Diagonal"
	BIASED_REVERSED_DIAGONAL_MODE = "B. Reversed Diagonal"
	HYBRID_DIAGONAL_MODE          = "Hybrid Diagonal"
)

var BATTLE_MOVEMENT_MODES = []string{DIAGONAL_MODE, REVERSED_DIAGONAL_MODE, BIASED_DIAGONAL_MODE, BIASED_REVERSED_DIAGONAL_MODE, HYBRID_DIAGONAL_MODE}

type BattleMovementState struct {
	hWnd             HWND
	Mode             BattleMovementMode
	currentDirection int
}

func (state *BattleMovementState) nextDirection() int {
	state.currentDirection ^= 1
	return state.currentDirection
}

func (state *BattleMovementState) Move() {

	var x, y int
	switch state.Mode {
	case DIAGONAL_MODE:
		x, y = diagonal(state, false, false)
	case REVERSED_DIAGONAL_MODE:
		x, y = diagonal(state, true, false)
	case BIASED_DIAGONAL_MODE:
		x, y = diagonal(state, false, true)
	case BIASED_REVERSED_DIAGONAL_MODE:
		x, y = diagonal(state, true, true)
	case HYBRID_DIAGONAL_MODE:
		x, y = diagonal(state, rand.Intn(2) != 0, true)
	default:
		x, y = none()
	}

	log.Printf("Handle %d moves to (%d, %d)\n", state.hWnd, x, y)

	sys.LeftClick(state.hWnd, int32(x), int32(y))
}

func diagonal(state *BattleMovementState, isReverse bool, isBias bool) (x, y int) {
	xOrigin := GAME_WIDTH / 2
	yOrigin := GAME_HEIGHT / 2

	var randomBiasAngle float64
	if isBias {
		randomBiasAngle = rand.Float64()*BIAS_ANGLE*2 - BIAS_ANGLE
	}

	direction := float64(state.nextDirection()) * math.Pi
	rotationAngle := ((45.0+randomBiasAngle)/180.0)*math.Pi + direction

	if isReverse {
		x = xOrigin + int(RADIUS*math.Cos(rotationAngle))
		y = yOrigin + int(RADIUS*math.Sin(rotationAngle))
	} else {
		x = xOrigin + int(RADIUS*math.Cos(rotationAngle))
		y = yOrigin + int(RADIUS*math.Sin(rotationAngle))*-1
	}

	return
}

func none() (int, int) {
	return GAME_WIDTH / 2, GAME_HEIGHT / 2
}

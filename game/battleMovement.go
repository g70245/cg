package game

import (
	sys "cg/system"
	"log"

	"math"
	"math/rand"

	. "github.com/lxn/win"
)

var (
	RADIUS    = 100
	DIRECTION = []int{-1, 1}
)

type BattleMovementMode string

const (
	NONE               = "None"
	CIRCLE_MODE        = "Circle"
	DIAGONAL_MODE      = "Diagonal"
	BACK_DIAGONAL_MODE = "Back Diagonal"
)

var BATTLE_MOVEMENT_MODES = []string{NONE, CIRCLE_MODE, DIAGONAL_MODE, BACK_DIAGONAL_MODE}

type BattleMovementState struct {
	hWnd             HWND
	mode             BattleMovementMode
	currentDirection int
}

func (state *BattleMovementState) nextDirection() (nextDirection int) {
	switch state.currentDirection {
	case 0:
		nextDirection = DIRECTION[rand.Intn(2)]
	default:
		nextDirection = state.currentDirection * -1
	}
	state.currentDirection = nextDirection
	return
}

func (state *BattleMovementState) Move() {

	var x, y int
	switch state.mode {
	case CIRCLE_MODE:
		x, y = circle()
	case DIAGONAL_MODE:
		x, y = diagonal(state, false)
	case BACK_DIAGONAL_MODE:
		x, y = diagonal(state, true)
	default:
		x, y = none()
	}

	log.Printf("Handle %d moves to (%d, %d)\n", state.hWnd, x, y)

	sys.LeftClick(state.hWnd, int32(x), int32(y))
}

func circle() (x, y int) {
	xOrigin := GAME_WIDTH / 2
	yOrigin := GAME_HEIGHT / 2

	xOffset := rand.Intn(RADIUS + 1)
	yOffset := int(math.Sqrt(math.Pow(float64(RADIUS), 2) - math.Pow(float64(xOffset), 2)))

	x = xOrigin + xOffset*DIRECTION[rand.Intn(2)]
	y = yOrigin + yOffset*DIRECTION[rand.Intn(2)]
	return
}

func diagonal(state *BattleMovementState, isReverse bool) (x, y int) {
	xOrigin := GAME_WIDTH / 2
	yOrigin := GAME_HEIGHT / 2

	Offset := int(math.Sqrt(math.Pow(float64(RADIUS), 2) / 2))

	direction := state.nextDirection()

	if isReverse {
		x = xOrigin + Offset*direction
		y = yOrigin + Offset*direction
	} else {
		x = xOrigin + Offset*direction
		y = yOrigin + Offset*direction*-1
	}

	return
}

func none() (int, int) {
	return GAME_WIDTH / 2, GAME_HEIGHT / 2
}

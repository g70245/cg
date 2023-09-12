package main

import (
	"fmt"
	"math"
	"math/rand"

	. "github.com/lxn/win"
)

var (
	WIDTH     = 640
	HEIGHT    = 480
	RADIUS    = 100
	DIRECTION = []int{-1, 1}
)

type MovementMode string

const (
	NONE               = "None"
	CIRCLE_MODE        = "Circle"
	DIAGONAL_MODE      = "Diagonal"
	BACK_DIAGONAL_MODE = "Back Diagonal"
)

var MOVEMENT_MODES = []string{NONE, CIRCLE_MODE, DIAGONAL_MODE, BACK_DIAGONAL_MODE}

type MovementState struct {
	hWnd             HWND
	mode             MovementMode
	currentDirection int
}

func (state *MovementState) nextDirection() (nextDirection int) {
	if state.currentDirection == 0 {
		nextDirection = DIRECTION[rand.Intn(2)]
	} else {
		nextDirection = state.currentDirection * -1
	}
	state.currentDirection = nextDirection
	return
}

func (m *MovementState) Move() {

	var x, y int
	switch m.mode {
	case CIRCLE_MODE:
		x, y = circle()
	case DIAGONAL_MODE:
		x, y = diagonal(*m, false)
	case BACK_DIAGONAL_MODE:
		x, y = diagonal(*m, true)
	default:
		x, y = none()
	}

	fmt.Printf("Handle %d moves to (%d, %d)\n", m.hWnd, x, y)

	Act(m.hWnd, int32(x), int32(y), WM_MOUSEMOVE)
	Act(m.hWnd, int32(x), int32(y), WM_LBUTTONDOWN)
	Act(m.hWnd, int32(x), int32(y), WM_LBUTTONUP)
}

func circle() (x, y int) {
	xOrigin := WIDTH / 2
	yOrigin := HEIGHT / 2

	xOffset := rand.Intn(RADIUS + 1)
	yOffset := int(math.Sqrt(math.Pow(float64(RADIUS), 2) - math.Pow(float64(xOffset), 2)))

	x = xOrigin + xOffset*DIRECTION[rand.Intn(2)]
	y = yOrigin + yOffset*DIRECTION[rand.Intn(2)]
	return
}

func diagonal(state MovementState, isReverse bool) (x, y int) {
	xOrigin := WIDTH / 2
	yOrigin := HEIGHT / 2

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
	return WIDTH / 2, HEIGHT / 2
}

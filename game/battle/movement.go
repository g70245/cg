package battle

import (
	"cg/game"
	"cg/game/battle/enums"
	"cg/game/battle/enums/movement"
	"cg/internal"

	"log"
	"math"
	"math/rand"

	"github.com/g70245/win"
)

var (
	MovementModes = enums.GenericEnum[movement.Mode]{List: []movement.Mode{movement.None, movement.DIAGONAL, movement.REVERSED_DIAGONAL, movement.BIASED_DIAGONAL, movement.BIASED_REVERSED_DIAGONAL, movement.HYBRID_DIAGONAL}}
)

const (
	RADIUS     float64 = 120
	BIAS_ANGLE float64 = 30
)

type MovementState struct {
	hWnd   win.HWND
	origin game.GamePos
	Mode   movement.Mode
}

func (s *MovementState) Move() {

	var x, y int32
	switch s.Mode {
	case movement.DIAGONAL:
		x, y = diagonal(RADIUS, s, false, false)
	case movement.REVERSED_DIAGONAL:
		x, y = diagonal(RADIUS, s, true, false)
	case movement.BIASED_DIAGONAL:
		x, y = diagonal(RADIUS, s, false, true)
	case movement.BIASED_REVERSED_DIAGONAL:
		x, y = diagonal(RADIUS, s, true, true)
	case movement.HYBRID_DIAGONAL:
		x, y = diagonal(RADIUS, s, rand.Intn(2) != 0, true)
	default:
		x, y = none()
	}

	log.Printf("Handle %d moves to (%d, %d)\n", s.hWnd, x, y)

	internal.LeftClick(s.hWnd, x, y)
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

func (s *MovementState) nextDirectionAngleMultilpier(isReverse bool) float64 {
	current := game.GetCurrentGamePos(s.hWnd)

	check := diagonalCondition
	if isReverse {
		check = reversedDiagonalCondition
	}

	if check(s.origin, current) {
		return 0
	} else {
		return 1
	}
}

func (s *MovementState) nextBiasMultiplier() float64 {
	current := game.GetCurrentGamePos(s.hWnd)

	switch {
	case current.X > s.origin.X && current.Y <= s.origin.Y:
		return 1
	case current.X > s.origin.X && current.Y > s.origin.Y:
		return -1
	case current.X <= s.origin.X && current.Y <= s.origin.Y:
		return -1
	case current.X <= s.origin.X && current.Y > s.origin.Y:
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

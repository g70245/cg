package production

import (
	"cg/game"
	"cg/internal"
)

var (
	PRODUCTION_INVENTORY_PIVOT = game.CheckTarget{X: 560, Y: 100, Color: COLOR_PR_INVENTORY_PIVOT}
)

func (w *Worker) isInventorySlotFree(px, py int32) bool {
	x := px
	for x < px+30 {
		y := py
		for y < py+30 {
			if internal.GetColor(w.hWnd, x, y) != COLOR_PR_INVENTORY_SLOT_EMPTY {
				return false
			}
			y += 5
		}
		x += 5
	}
	return true
}

func (w *Worker) getInventoryPos() (int32, int32, bool) {
	internal.MoveCursorToNowhere(w.hWnd)
	x := PRODUCTION_INVENTORY_PIVOT.X
	for x <= PRODUCTION_INVENTORY_PIVOT.X+54 {
		y := PRODUCTION_INVENTORY_PIVOT.Y
		for y <= PRODUCTION_INVENTORY_PIVOT.Y+34 {
			if internal.GetColor(w.hWnd, x, y) == PRODUCTION_INVENTORY_PIVOT.Color {
				return x - 4*50 - 30, y + 28, true
			}
			y += 1
		}
		x += 1
	}
	return 0, 0, false
}

func (w *Worker) canProduce(x, y int32) bool {
	return internal.GetColor(w.hWnd, x-270, y+180) == COLOR_PR_PRODUCE_BUTTON
}

func (w *Worker) isProducing(x, y int32) bool {
	return internal.GetColor(w.hWnd, x-110, y+10) != COLOR_PR_NOT_PRODUCING
}

func (w *Worker) isProducingSuccessful(x, y int32) bool {
	return internal.GetColor(w.hWnd, x-166, y+180) == COLOR_PR_PRODUCE_BUTTON
}

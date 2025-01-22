package production

import (
	"cg/game"
	"cg/internal"
)

var (
	PRODUCTION_WINDOW_ITEM_PIVOT = game.CheckTarget{X: 560, Y: 100, Color: COLOR_PR_INVENTORY_PIVOT}
)

func (p *ProductionWorker) isSlotFree(px, py int32) bool {
	x := px
	for x < px+30 {
		y := py
		for y < py+30 {
			if internal.GetColor(p.hWnd, x, y) != COLOR_PR_INVENTORY_SLOT_EMPTY {
				return false
			}
			y += 5
		}
		x += 5
	}
	return true
}

func (p *ProductionWorker) getItemWindowPos() (int32, int32, bool) {
	internal.MoveCursorToNowhere(p.hWnd)
	x := PRODUCTION_WINDOW_ITEM_PIVOT.X
	for x <= PRODUCTION_WINDOW_ITEM_PIVOT.X+54 {
		y := PRODUCTION_WINDOW_ITEM_PIVOT.Y
		for y <= PRODUCTION_WINDOW_ITEM_PIVOT.Y+34 {
			if internal.GetColor(p.hWnd, x, y) == PRODUCTION_WINDOW_ITEM_PIVOT.Color {
				return x - 4*50 - 30, y + 28, true
			}
			y += 1
		}
		x += 1
	}
	return 0, 0, false
}

func (p *ProductionWorker) canProduce(x, y int32) bool {
	return internal.GetColor(p.hWnd, x-270, y+180) == game.COLOR_PR_PRODUCE_BUTTON
}

func (p *ProductionWorker) isProducing(x, y int32) bool {
	return internal.GetColor(p.hWnd, x-110, y+10) != game.COLOR_PR_NOT_PRODUCING
}

func (p *ProductionWorker) isProducingSuccessful(x, y int32) bool {
	return internal.GetColor(p.hWnd, x-166, y+180) == game.COLOR_PR_PRODUCE_BUTTON
}

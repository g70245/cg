package game

import (
	. "cg/internal"
	"time"
)

const (
	DURATION_LOG_PRODUCTION_STATUS = 400 * time.Millisecond
)

var (
	PRODUCTION_WINDOW_ITEM_PIVOT = CheckTarget{560, 100, COLOR_PR_INVENTORY_PIVOT}
)

func (p *ProductionWorker) isSlotFree(px, py int32) bool {
	x := px
	for x < px+30 {
		y := py
		for y < py+30 {
			if GetColor(p.hWnd, x, y) != COLOR_PR_INVENTORY_SLOT_EMPTY {
				return false
			}
			y += 5
		}
		x += 5
	}
	return true
}

func (p *ProductionWorker) canProduce(x, y int32) bool {
	return GetColor(p.hWnd, x-270, y+180) == COLOR_PR_PRODUCE_BUTTON
}

func (p *ProductionWorker) isProducing(x, y int32) bool {
	return GetColor(p.hWnd, x-110, y+10) != COLOR_PR_NOT_PRODUCING
}

func (p *ProductionWorker) isProducingSuccessful(x, y int32) bool {
	return GetColor(p.hWnd, x-166, y+180) == COLOR_PR_PRODUCE_BUTTON
}

func isProductionStatusOK(name, dir string, before time.Duration) bool {
	if dir == "" {
		return false
	}
	return checkWord(dir, 10, before, append(LOG_PRODUCTION_FAILURE, name))
}

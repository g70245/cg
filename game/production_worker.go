package game

import (
	. "cg/internal"
	. "cg/utils"
	"log"
	"time"

	"github.com/g70245/win"
)

const (
	NO_WORKER_NAME = "none"

	DURATION_PRODUCTION_WORKER = 400 * time.Millisecond
	DURATION_PRODUCING         = 800 * time.Millisecond
	DURATION_UNPACKING         = 600 * time.Millisecond
	DURATION_TIDYING_UP        = 300 * time.Millisecond

	DURATION_PRODUCTION_CHECKER_LOG         = 2 * time.Second
	DURATION_PRODUCTION_CHECKER_INVENTORY   = 16 * time.Second
	DURATION_PRODUCTION_CHECKER_AUDIBLE_CUE = 4 * time.Second
)

type ProductionWorker struct {
	hWnd     win.HWND
	name     string
	gameDir  *string
	stopChan chan bool

	GatheringMode bool
	ManualMode    bool

	workerTicker           *time.Ticker
	logCheckerTicker       *time.Ticker
	inventoryCheckerTicker *time.Ticker
	audibleCueTicker       *time.Ticker
}

func NewProductionWorker(hWnd win.HWND, gameDir *string, stopChan chan bool) ProductionWorker {
	newWorkerTicker := time.NewTicker(time.Hour)
	newLogCheckerTicker := time.NewTicker(time.Hour)
	newInventoryCheckerTicker := time.NewTicker(time.Hour)
	newAudibleCueTicker := time.NewTicker(time.Hour)

	newWorkerTicker.Stop()
	newLogCheckerTicker.Stop()
	newInventoryCheckerTicker.Stop()
	newAudibleCueTicker.Stop()

	return ProductionWorker{
		hWnd:                   hWnd,
		name:                   NO_WORKER_NAME,
		gameDir:                gameDir,
		stopChan:               stopChan,
		workerTicker:           newWorkerTicker,
		logCheckerTicker:       newLogCheckerTicker,
		inventoryCheckerTicker: newInventoryCheckerTicker,
		audibleCueTicker:       newAudibleCueTicker,
	}
}

func (p *ProductionWorker) Work() {

	p.workerTicker.Reset(DURATION_PRODUCTION_WORKER)
	p.logCheckerTicker.Reset(DURATION_PRODUCTION_CHECKER_LOG)
	p.inventoryCheckerTicker.Reset(DURATION_PRODUCTION_CHECKER_INVENTORY)
	p.audibleCueTicker.Reset(DURATION_PRODUCTION_CHECKER_AUDIBLE_CUE)

	log.Printf("Handle %d Production started\n", p.hWnd)

	go func() {
		defer p.StopTickers()

		p.Reset()

		for {
			select {
			case <-p.workerTicker.C:
				if !p.GatheringMode {
					p.prepareMaterials()
					p.produce()
					p.tidyInventory()
				}
			case <-p.logCheckerTicker.C:
				if isProductionStatusOK(p.name, *p.gameDir, DURATION_PRODUCTION_CHECKER_LOG) {
					log.Printf("Production %d status check was not passed\n", p.hWnd)
					p.StopTickers()
					Beeper.Play()
				}
			case <-p.inventoryCheckerTicker.C:
				if isInventoryFullWithoutClosingAllWindows(p.hWnd) {
					log.Printf("Production %d inventory is full\n", p.hWnd)
					p.StopTickers()
					Beeper.Play()
				}
			case <-p.audibleCueTicker.C:
				if p.ManualMode {
					p.StopTickers()
					Beeper.Play()
				}
			case <-p.stopChan:
				log.Printf("Handle %d Production ended\n", p.hWnd)
				return
			}
		}
	}()
}

func (p *ProductionWorker) StopTickers() {
	p.workerTicker.Stop()
	p.logCheckerTicker.Stop()
	p.inventoryCheckerTicker.Stop()
	p.audibleCueTicker.Stop()
}

func (p *ProductionWorker) Stop() {
	p.stopChan <- true
	p.ManualMode = true
	Beeper.Stop()
}

func (p *ProductionWorker) Reset() {
	p.ManualMode = false
	Beeper.Stop()
}

func (p *ProductionWorker) prepareMaterials() {

	if p.ManualMode {
		return
	}

	log.Printf("Production %d is preparing\n", p.hWnd)

	defer switchWindow(p.hWnd, KEY_INVENTORY)
	switchWindow(p.hWnd, KEY_INVENTORY)

	nx, ny, ok := getItemWindowPos(p.hWnd)
	if !ok {
		log.Printf("Production %d cannot find the position of item window\n", p.hWnd)
		p.ManualMode = true
		return
	}

	var i int32
	for i = 0; i < 5; i++ {
		if isInventorySlotFree(p.hWnd, nx+i*ITEM_COL_LEN, ny) {
			if isInventorySlotFree(p.hWnd, nx+i*ITEM_COL_LEN, ny+3*ITEM_COL_LEN) {
				log.Printf("Production %d have no materials\n", p.hWnd)
				p.ManualMode = true
				return
			}

			DoubleClick(p.hWnd, nx+i*ITEM_COL_LEN, ny+3*ITEM_COL_LEN)
			time.Sleep(DURATION_UNPACKING)

			if isInventorySlotFree(p.hWnd, nx+i*ITEM_COL_LEN, ny) {
				log.Printf("Production %d cannot unpack material\n", p.hWnd)
				p.ManualMode = true
				return
			}
		}
	}
}

func (p *ProductionWorker) produce() {

	if p.ManualMode {
		return
	}

	log.Printf("Production %d is producing\n", p.hWnd)

	px, py, ok := p.getItemWindowPos()
	if !ok {
		log.Printf("Production %d cannot find the position of production window\n", p.hWnd)
		p.ManualMode = true
		return
	}

	LeftClick(p.hWnd, px-270, py+180)

	var i int32
	for i = 0; i < 5; i++ {
		DoubleClickRepeatedly(p.hWnd, px+i*ITEM_COL_LEN+10, py+10)
		if p.canProduce(px, py) {
			break
		}
	}

	if !p.canProduce(px, py) {
		log.Printf("Production %d out of mana or insufficient materials\n", p.hWnd)
		p.ManualMode = true
		return
	}

	LeftClick(p.hWnd, px-270, py+180)
	if !p.isProducing(px, py) {
		log.Printf("Production %d missed the producing button\n", p.hWnd)
		p.ManualMode = true
		return
	}

	for !p.isProducingSuccessful(px, py) {
		time.Sleep(DURATION_PRODUCING)
	}
}

func (p *ProductionWorker) tidyInventory() {

	if p.ManualMode {
		return
	}

	log.Printf("Production %d is tidying up\n", p.hWnd)

	px, py, ok := p.getItemWindowPos()
	if !ok {
		log.Printf("Production %d cannot find the position of production window\n", p.hWnd)
		p.ManualMode = true
		return
	}

	var j int32
	for j = 1; j <= 2; j++ {
		var i int32
		for i = 4; i > 0; i-- {
			MoveCursorToNowhere(p.hWnd)
			if p.isSlotFree(px+i*ITEM_COL_LEN, py+j*ITEM_COL_LEN) {
				continue
			}
			LeftClick(p.hWnd, px+i*50, py+j*50)
			LeftClick(p.hWnd, px+(i-1)*50, py+j*50)
			time.Sleep(DURATION_TIDYING_UP)
		}
	}
}

func (p *ProductionWorker) SetName(name string) {
	p.name = name
}

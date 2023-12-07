package game

import (
	. "cg/system"
	"log"
	"time"

	. "github.com/g70245/win"
)

const (
	PRODUCTION_WORKER_INTERVAL    = 400
	PRODUCTION_PRODUCING_INTERVAL = 800
	PRODUCTION_UNPACK_INTERVAL    = 600
	PRODUCTION_TIDY_UP_INTERVAL   = 300

	NAME_NONE = "none"

	PRODUCTION_CHECKER_LOG_INTERVAL_SEC       = 6
	PRODUCTION_CHECKER_INVENTORY_INTERVAL_SEC = 16
)

type ProductionWorker struct {
	hWnd     HWND
	logDir   *string
	stopChan chan bool

	Name          string
	GatheringMode bool
	ManualMode    bool

	workerTicker           *time.Ticker
	logCheckerTicker       *time.Ticker
	inventoryCheckerTicker *time.Ticker
}

func CreateProductionWorker(hWnd HWND, logDir *string, stopChan chan bool) ProductionWorker {
	newWorkerTicker := time.NewTicker(time.Hour)
	newLogCheckerTicker := time.NewTicker(time.Hour)
	newInventoryCheckerTicker := time.NewTicker(time.Hour)

	newWorkerTicker.Stop()
	newLogCheckerTicker.Stop()
	newInventoryCheckerTicker.Stop()

	return ProductionWorker{
		hWnd:                   hWnd,
		logDir:                 logDir,
		Name:                   NAME_NONE,
		stopChan:               stopChan,
		workerTicker:           newWorkerTicker,
		logCheckerTicker:       newLogCheckerTicker,
		inventoryCheckerTicker: newInventoryCheckerTicker,
	}
}

func (p *ProductionWorker) Work() {

	p.ManualMode = false
	p.workerTicker.Reset(PRODUCTION_WORKER_INTERVAL * time.Millisecond)
	p.logCheckerTicker.Reset(PRODUCTION_CHECKER_LOG_INTERVAL_SEC * time.Second)
	p.inventoryCheckerTicker.Reset(PRODUCTION_CHECKER_INVENTORY_INTERVAL_SEC * time.Second)

	log.Printf("Handle %d Production started\n", p.hWnd)

	go func() {
		defer p.StopTickers()

		for {
			select {
			case <-p.workerTicker.C:
				if !p.GatheringMode {
					p.prepareMaterials()
					p.produce()
					p.tidyInventory()

					if p.ManualMode {
						p.StopTickers()
						Beeper.Play()
					}
				}
			case <-p.logCheckerTicker.C:
				if checkProductionStatus(p.Name, *p.logDir) {
					log.Printf("Production %d status check was not passed\n", p.hWnd)
					p.StopTickers()
					Beeper.Play()
				}
			case <-p.inventoryCheckerTicker.C:
				if checkInventoryWithoutClosingAllWindows(p.hWnd) {
					log.Printf("Production %d inventory is full\n", p.hWnd)
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
}

func (p *ProductionWorker) Stop() {
	p.stopChan <- true
	Beeper.Stop()
}

func (p *ProductionWorker) prepareMaterials() {

	if p.ManualMode {
		return
	}

	defer leverWindowByShortcutWithoutClosingOtherWindows(p.hWnd, 0x45)
	leverWindowByShortcutWithoutClosingOtherWindows(p.hWnd, 0x45)

	nx, ny, ok := getNSItemWindowPos(p.hWnd)
	if !ok {
		log.Printf("Production %d cannot find the position of item window\n", p.hWnd)
		p.ManualMode = true
		return
	}

	var i int32
	for i = 0; i < 5; i++ {
		if isSlotFree(p.hWnd, nx+i*ITEM_COL_LEN, ny) {
			if isSlotFree(p.hWnd, nx+i*ITEM_COL_LEN, ny+3*ITEM_COL_LEN) {
				log.Printf("Production %d have no materials\n", p.hWnd)
				p.ManualMode = true
				return
			}

			DoubleClick(p.hWnd, nx+i*ITEM_COL_LEN, ny+3*ITEM_COL_LEN)
			time.Sleep(PRODUCTION_UNPACK_INTERVAL * time.Millisecond)

			if isSlotFree(p.hWnd, nx+i*ITEM_COL_LEN, ny) {
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

	px, py, ok := getPRItemWindowPos(p.hWnd)
	if !ok {
		log.Printf("Production %d cannot find the position of production window\n", p.hWnd)
		p.ManualMode = true
		return
	}

	LeftClick(p.hWnd, px-270, py+180)

	var i int32
	for i = 0; i < 5; i++ {
		DoubleClickRepeatedly(p.hWnd, px+i*ITEM_COL_LEN+10, py+10)
		if canProduce(p.hWnd, px, py) {
			break
		}
	}

	if !canProduce(p.hWnd, px, py) {
		log.Printf("Production %d out of mana or insufficient materials\n", p.hWnd)
		p.ManualMode = true
		return
	}

	LeftClick(p.hWnd, px-270, py+180)
	if !isProducing(p.hWnd, px, py) {
		log.Printf("Production %d missed the producing button\n", p.hWnd)
		p.ManualMode = true
		return
	}

	for !isProducingSuccessful(p.hWnd, px, py) {
		time.Sleep(PRODUCTION_PRODUCING_INTERVAL * time.Millisecond)
	}
}

func (p *ProductionWorker) tidyInventory() {

	if p.ManualMode {
		return
	}

	px, py, ok := getPRItemWindowPos(p.hWnd)
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
			if isPRSlotFree(p.hWnd, px+i*ITEM_COL_LEN, py+j*ITEM_COL_LEN) {

				continue
			}
			LeftClick(p.hWnd, px+i*50, py+j*50)
			LeftClick(p.hWnd, px+(i-1)*50, py+j*50)
			time.Sleep(PRODUCTION_TIDY_UP_INTERVAL * time.Millisecond)
		}
	}
}

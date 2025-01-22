package production

import (
	"cg/game"
	"cg/internal"
	"cg/utils"
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

type Worker struct {
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

func NewWorker(hWnd win.HWND, gameDir *string, stopChan chan bool) Worker {
	newWorkerTicker := time.NewTicker(time.Hour)
	newLogCheckerTicker := time.NewTicker(time.Hour)
	newInventoryCheckerTicker := time.NewTicker(time.Hour)
	newAudibleCueTicker := time.NewTicker(time.Hour)

	newWorkerTicker.Stop()
	newLogCheckerTicker.Stop()
	newInventoryCheckerTicker.Stop()
	newAudibleCueTicker.Stop()

	return Worker{
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

func (p *Worker) Work() {

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
				if game.IsProductionStatusOK(p.name, *p.gameDir, DURATION_PRODUCTION_CHECKER_LOG) {
					log.Printf("Production %d status check was not passed\n", p.hWnd)
					p.StopTickers()
					utils.Beeper.Play()
				}
			case <-p.inventoryCheckerTicker.C:
				if game.IsInventoryFullWithoutClosingAllWindows(p.hWnd) {
					log.Printf("Production %d inventory is full\n", p.hWnd)
					p.StopTickers()
					utils.Beeper.Play()
				}
			case <-p.audibleCueTicker.C:
				if p.ManualMode {
					p.StopTickers()
					utils.Beeper.Play()
				}
			case <-p.stopChan:
				log.Printf("Handle %d Production ended\n", p.hWnd)
				return
			}
		}
	}()
}

func (p *Worker) StopTickers() {
	p.workerTicker.Stop()
	p.logCheckerTicker.Stop()
	p.inventoryCheckerTicker.Stop()
	p.audibleCueTicker.Stop()
}

func (p *Worker) Stop() {
	p.stopChan <- true
	p.ManualMode = true
	utils.Beeper.Stop()
}

func (p *Worker) Reset() {
	p.ManualMode = false
	utils.Beeper.Stop()
}

func (p *Worker) prepareMaterials() {

	if p.ManualMode {
		return
	}

	log.Printf("Production %d is preparing\n", p.hWnd)

	defer game.SwitchWindow(p.hWnd, game.KEY_INVENTORY)
	game.SwitchWindow(p.hWnd, game.KEY_INVENTORY)

	nx, ny, ok := game.GetItemWindowPos(p.hWnd)
	if !ok {
		log.Printf("Production %d cannot find the position of item window\n", p.hWnd)
		p.ManualMode = true
		return
	}

	var i int32
	for i = 0; i < 5; i++ {
		if game.IsInventorySlotFree(p.hWnd, nx+i*game.ITEM_COL_LEN, ny) {
			if game.IsInventorySlotFree(p.hWnd, nx+i*game.ITEM_COL_LEN, ny+3*game.ITEM_COL_LEN) {
				log.Printf("Production %d have no materials\n", p.hWnd)
				p.ManualMode = true
				return
			}

			internal.DoubleClick(p.hWnd, nx+i*game.ITEM_COL_LEN, ny+3*game.ITEM_COL_LEN)
			time.Sleep(DURATION_UNPACKING)

			if game.IsInventorySlotFree(p.hWnd, nx+i*game.ITEM_COL_LEN, ny) {
				log.Printf("Production %d cannot unpack material\n", p.hWnd)
				p.ManualMode = true
				return
			}
		}
	}
}

func (p *Worker) produce() {

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

	internal.LeftClick(p.hWnd, px-270, py+180)

	var i int32
	for i = 0; i < 5; i++ {
		internal.DoubleClickRepeatedly(p.hWnd, px+i*game.ITEM_COL_LEN+10, py+10)
		if p.canProduce(px, py) {
			break
		}
	}

	if !p.canProduce(px, py) {
		log.Printf("Production %d out of mana or insufficient materials\n", p.hWnd)
		p.ManualMode = true
		return
	}

	internal.LeftClick(p.hWnd, px-270, py+180)
	if !p.isProducing(px, py) {
		log.Printf("Production %d missed the producing button\n", p.hWnd)
		p.ManualMode = true
		return
	}

	for !p.isProducingSuccessful(px, py) {
		time.Sleep(DURATION_PRODUCING)
	}
}

func (p *Worker) tidyInventory() {

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
			internal.MoveCursorToNowhere(p.hWnd)
			if p.isSlotFree(px+i*game.ITEM_COL_LEN, py+j*game.ITEM_COL_LEN) {
				continue
			}
			internal.LeftClick(p.hWnd, px+i*50, py+j*50)
			internal.LeftClick(p.hWnd, px+(i-1)*50, py+j*50)
			time.Sleep(DURATION_TIDYING_UP)
		}
	}
}

func (p *Worker) SetName(name string) {
	p.name = name
}

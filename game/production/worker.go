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

func (w *Worker) Work() {

	w.workerTicker.Reset(DURATION_PRODUCTION_WORKER)
	w.logCheckerTicker.Reset(DURATION_PRODUCTION_CHECKER_LOG)
	w.inventoryCheckerTicker.Reset(DURATION_PRODUCTION_CHECKER_INVENTORY)
	w.audibleCueTicker.Reset(DURATION_PRODUCTION_CHECKER_AUDIBLE_CUE)

	log.Printf("Handle %d Production started\n", w.hWnd)

	go func() {
		defer w.StopTickers()

		w.Reset()

		for {
			select {
			case <-w.workerTicker.C:
				if !w.GatheringMode {
					w.prepareMaterials()
					w.produce()
					w.tidyInventory()
				}
			case <-w.logCheckerTicker.C:
				if game.IsProductionStatusOK(w.name, *w.gameDir, DURATION_PRODUCTION_CHECKER_LOG) {
					log.Printf("Production %d status check was not passed\n", w.hWnd)
					w.StopTickers()
					utils.Beeper.Play()
				}
			case <-w.inventoryCheckerTicker.C:
				if game.IsInventoryFullWithoutClosingAllWindows(w.hWnd) {
					log.Printf("Production %d inventory is full\n", w.hWnd)
					w.StopTickers()
					utils.Beeper.Play()
				}
			case <-w.audibleCueTicker.C:
				if w.ManualMode {
					w.StopTickers()
					utils.Beeper.Play()
				}
			case <-w.stopChan:
				log.Printf("Handle %d Production ended\n", w.hWnd)
				return
			}
		}
	}()
}

func (w *Worker) StopTickers() {
	w.workerTicker.Stop()
	w.logCheckerTicker.Stop()
	w.inventoryCheckerTicker.Stop()
	w.audibleCueTicker.Stop()
}

func (w *Worker) Stop() {
	w.stopChan <- true
	w.ManualMode = true
	utils.Beeper.Stop()
}

func (w *Worker) Reset() {
	w.ManualMode = false
	utils.Beeper.Stop()
}

func (w *Worker) prepareMaterials() {

	if w.ManualMode {
		return
	}

	log.Printf("Production %d is preparing\n", w.hWnd)

	defer game.SwitchWindow(w.hWnd, game.KEY_INVENTORY)
	game.SwitchWindow(w.hWnd, game.KEY_INVENTORY)

	nx, ny, ok := game.GetItemWindowPos(w.hWnd)
	if !ok {
		log.Printf("Production %d cannot find the position of item window\n", w.hWnd)
		w.ManualMode = true
		return
	}

	var i int32
	for i = 0; i < 5; i++ {
		if game.IsInventorySlotFree(w.hWnd, nx+i*game.ITEM_COL_LEN, ny) {
			if game.IsInventorySlotFree(w.hWnd, nx+i*game.ITEM_COL_LEN, ny+3*game.ITEM_COL_LEN) {
				log.Printf("Production %d has no materials\n", w.hWnd)
				w.ManualMode = true
				return
			}

			internal.DoubleClick(w.hWnd, nx+i*game.ITEM_COL_LEN, ny+3*game.ITEM_COL_LEN)
			time.Sleep(DURATION_UNPACKING)

			if game.IsInventorySlotFree(w.hWnd, nx+i*game.ITEM_COL_LEN, ny) {
				log.Printf("Production %d cannot unpack material\n", w.hWnd)
				w.ManualMode = true
				return
			}
		}
	}
}

func (w *Worker) produce() {

	if w.ManualMode {
		return
	}

	log.Printf("Production %d is producing\n", w.hWnd)

	px, py, ok := w.getInventoryPos()
	if !ok {
		log.Printf("Production %d cannot find the position of production window\n", w.hWnd)
		w.ManualMode = true
		return
	}

	internal.LeftClick(w.hWnd, px-270, py+180)

	var i int32
	for i = 0; i < 5; i++ {
		internal.DoubleClickRepeatedly(w.hWnd, px+i*game.ITEM_COL_LEN+10, py+10)
		if w.canProduce(px, py) {
			break
		}
	}

	if !w.canProduce(px, py) {
		log.Printf("Production %d is out of mana or has insufficient materials\n", w.hWnd)
		w.ManualMode = true
		return
	}

	internal.LeftClick(w.hWnd, px-270, py+180)
	if !w.isProducing(px, py) {
		log.Printf("Production %d missed the producing button\n", w.hWnd)
		w.ManualMode = true
		return
	}

	for !w.isProducingSuccessful(px, py) {
		time.Sleep(DURATION_PRODUCING)
	}
}

func (w *Worker) tidyInventory() {

	if w.ManualMode {
		return
	}

	log.Printf("Production %d is tidying up\n", w.hWnd)

	px, py, ok := w.getInventoryPos()
	if !ok {
		log.Printf("Production %d cannot find the position of production window\n", w.hWnd)
		w.ManualMode = true
		return
	}

	var j int32
	for j = 1; j <= 2; j++ {
		var i int32
		for i = 4; i > 0; i-- {
			internal.MoveCursorToNowhere(w.hWnd)
			if w.isInventorySlotFree(px+i*game.ITEM_COL_LEN, py+j*game.ITEM_COL_LEN) {
				continue
			}
			internal.LeftClick(w.hWnd, px+i*50, py+j*50)
			internal.LeftClick(w.hWnd, px+(i-1)*50, py+j*50)
			time.Sleep(DURATION_TIDYING_UP)
		}
	}
}

func (w *Worker) SetName(name string) {
	w.name = name
}

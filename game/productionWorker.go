package game

import (
	sys "cg/system"
	"log"
	"time"

	. "github.com/g70245/win"
)

const (
	PRODUCTION_WORKER_INTERVAL         = 400
	PRODUCTION_CHECKER_INTERVAL_SECOND = 6
	PRODUCTION_PRODUCING_INTERVAL      = 800
	PRODUCTION_UNPACK_INTERVAL         = 600
	PRODUCTION_TIDY_UP_INTERVAL        = 300
	NAME_NONE                          = "none"
)

type ProductionWorker struct {
	hWnd   HWND
	logDir *string

	Name        string
	IsGathering bool
}

func CreateProductionWorker(hWnd HWND, logDir *string) ProductionWorker {
	return ProductionWorker{
		hWnd:   hWnd,
		logDir: logDir,
		Name:   NAME_NONE,
	}
}

func (p *ProductionWorker) Work(stopChan chan bool) {
	workerTicker := time.NewTicker(PRODUCTION_WORKER_INTERVAL * time.Millisecond)
	logCheckerTicker := time.NewTicker(PRODUCTION_CHECKER_INTERVAL_SECOND * time.Second)

	log.Printf("Handle %d Production started\n", p.hWnd)

	isPlayingBeeper := false
	isSomethingWrong := false

	go func() {
		defer workerTicker.Stop()
		defer logCheckerTicker.Stop()

		for {
			select {
			case <-workerTicker.C:
				if !p.IsGathering && !isSomethingWrong {
					if !p.prepareMaterials() {
						isSomethingWrong = true
						break
					}

					if !p.produce() {
						isSomethingWrong = true
						break
					}

					isSomethingWrong = !p.tidyInventory()
				}
			case <-logCheckerTicker.C:
				if checkProductionStatus(p.Name, *p.logDir) {
					isSomethingWrong = true
					log.Printf("Production %d status check was not passed\n", p.hWnd)
				}
			case <-stopChan:
				return
			default:
				if !isPlayingBeeper && (isSomethingWrong) {
					isPlayingBeeper = sys.PlayBeeper()
				}
				time.Sleep(PRODUCTION_WORKER_INTERVAL * time.Microsecond / 3)
			}
		}
	}()
}

func (p *ProductionWorker) prepareMaterials() bool {
	defer leverWindowByShortcutWithoutClosingOtherWindows(p.hWnd, 0x45)
	leverWindowByShortcutWithoutClosingOtherWindows(p.hWnd, 0x45)

	nx, ny, ok := getNSItemWindowPos(p.hWnd)
	if !ok {
		log.Printf("Production %d cannot find the position of item window\n", p.hWnd)
		return false
	}

	var i int32
	for i = 0; i < 5; i++ {
		if isSlotFree(p.hWnd, nx+i*ITEM_COL_LEN, ny) {
			if isSlotFree(p.hWnd, nx+i*ITEM_COL_LEN, ny+3*ITEM_COL_LEN) {
				log.Printf("Production %d have no materials\n", p.hWnd)
				return false
			}

			sys.DoubleClick(p.hWnd, nx+i*ITEM_COL_LEN, ny+3*ITEM_COL_LEN)
			time.Sleep(PRODUCTION_UNPACK_INTERVAL * time.Millisecond)

			if isSlotFree(p.hWnd, nx+i*ITEM_COL_LEN, ny) {
				log.Printf("Production %d cannot unpack material\n", p.hWnd)
				return false
			}
		}
	}

	return true
}

func (p *ProductionWorker) produce() bool {
	px, py, ok := getPRItemWindowPos(p.hWnd)
	if !ok {
		log.Printf("Production %d cannot find the position of production window\n", p.hWnd)
		return false
	}

	sys.LeftClick(p.hWnd, px-270, py+180)

	var i int32
	for i = 0; i < 5; i++ {
		sys.DoubleClickRepeatedly(p.hWnd, px+i*ITEM_COL_LEN+10, py+10)
		if canProduce(p.hWnd, px, py) {
			break
		}
	}

	if !canProduce(p.hWnd, px, py) {
		log.Printf("Production %d out of mana or insufficient materials\n", p.hWnd)
		return false
	}

	sys.LeftClick(p.hWnd, px-270, py+180)

	if !isProducing(p.hWnd, px, py) {
		log.Printf("Production %d missed the producing button\n", p.hWnd)
		return false
	}

	for !isProducingSuccessful(p.hWnd, px, py) {
		time.Sleep(PRODUCTION_PRODUCING_INTERVAL * time.Millisecond)
	}

	return true
}

func (p *ProductionWorker) tidyInventory() bool {
	px, py, ok := getPRItemWindowPos(p.hWnd)
	if !ok {
		log.Printf("Production %d cannot find the position of production window\n", p.hWnd)
		return false
	}

	var j int32
	for j = 1; j <= 2; j++ {
		var i int32
		for i = 4; i > 0; i-- {
			sys.MoveToNowhere(p.hWnd)
			if isPRSlotFree(p.hWnd, px+i*ITEM_COL_LEN, py+j*ITEM_COL_LEN) {

				continue
			}
			sys.LeftClick(p.hWnd, px+i*50, py+j*50)
			sys.LeftClick(p.hWnd, px+(i-1)*50, py+j*50)
			time.Sleep(PRODUCTION_TIDY_UP_INTERVAL * time.Millisecond)
		}
	}
	return true
}

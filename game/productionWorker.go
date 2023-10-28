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
)

type ProductionWorker struct {
	hWnd   HWND
	logDir *string

	IsGathering bool
}

func CreateProductionWorker(hWnd HWND, logDir *string) ProductionWorker {
	return ProductionWorker{hWnd, logDir, false}
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
				if !p.IsGathering {
					resetAllWindowsPosition(p.hWnd)
				}
			case <-logCheckerTicker.C:
				isSomethingWrong = checkProductionStatus(*p.logDir)
			case <-stopChan:
				return
			default:
				if !isPlayingBeeper && (isSomethingWrong) {
					isPlayingBeeper = sys.PlayBeeper()
					log.Printf("Handle %d is something wrong: %t\n", p.hWnd, isSomethingWrong)
				}
				time.Sleep(PRODUCTION_WORKER_INTERVAL * time.Microsecond / 3)
			}
		}
	}()
}

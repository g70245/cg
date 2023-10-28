package game

import (
	. "github.com/g70245/win"
)

type ProductionWorker struct {
	hWnd   HWND
	logDir *string
}

func CreateProductionWorker(hWnd HWND, logDir *string) ProductionWorker {
	return ProductionWorker{hWnd, logDir}
}

func (p *ProductionWorker) Work(stopChan chan bool) {
	go func() {
		select {
		case <-stopChan:
			return
		}
	}()
}

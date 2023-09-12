package main

import (
	"fmt"
	"time"

	. "github.com/lxn/win"
)

const WORKER_DURATION_MILLIS = 600

var GLOBAL_PARTY_LEAD_HWND string

type Worker struct {
	hWnd             HWND
	movementStrategy MovementStrategy
}

type Workers []Worker

func CreateWorkers(hWnds []HWND) Workers {
	workers := make([]Worker, len(hWnds))
	for _, hWnd := range hWnds {
		workers = append(workers, Worker{hWnd: hWnd})
	}
	return workers
}

func (w Worker) GetHandle() string {
	return fmt.Sprint(w.hWnd)
}

func (ws Workers) GetHandles() []string {
	hWnds := make([]string, 0)
	for _, w := range ws {
		hWnds = append(hWnds, w.GetHandle())
	}
	return hWnds
}

func (w Worker) Work(stopChan chan bool) {
	ticker := time.NewTicker(WORKER_DURATION_MILLIS * time.Millisecond)
	m := MovementState{hWnd: w.hWnd}

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				switch GetScene(w.hWnd) {
				case BATTLE_SCENE:
					fmt.Printf("Handle %s is at BATTLE_SCENE\n", w.GetHandle())
				case NORMAL_SCENE:
					if GLOBAL_PARTY_LEAD_HWND == "" || GLOBAL_PARTY_LEAD_HWND == w.GetHandle() {
						m.Move()
						fmt.Printf("Handle %s is at NORMAL_SCENE\n", w.GetHandle())
					}
				}
			case <-stopChan:
				return
			}
		}
	}()
}

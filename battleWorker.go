package main

import (
	"fmt"
	"time"

	. "github.com/lxn/win"
)

const BATTLE_WORKER_DURATION_MILLIS = 800

type BattleWorker struct {
	hWnd         HWND
	movementMode BattleMovementMode
}

type BattleWorkers []BattleWorker

func CreateBattleWorkers(hWnds []HWND) BattleWorkers {
	var workers []BattleWorker
	for _, hWnd := range hWnds {
		workers = append(workers, BattleWorker{hWnd: hWnd, movementMode: NONE})
	}
	return workers
}

func (w BattleWorker) getHandle() string {
	return fmt.Sprint(w.hWnd)
}

func (w BattleWorker) Work(leadHandle *string, stopChan chan bool) {
	ticker := time.NewTicker(BATTLE_WORKER_DURATION_MILLIS * time.Millisecond)
	m := BattleMovementState{hWnd: w.hWnd, mode: w.movementMode}

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				switch GetScene(w.hWnd) {
				case BATTLE_SCENE:
					fmt.Printf("Handle %s is at BATTLE_SCENE\n", w.getHandle())
				case NORMAL_SCENE:
					if *leadHandle == "" || *leadHandle == w.getHandle() {
						m.Move()
						fmt.Printf("Handle %s is at NORMAL_SCENE\n", w.getHandle())
					}
				}
			case <-stopChan:
				return
			default:
			}
		}
	}()
}

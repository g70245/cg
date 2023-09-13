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

// /*leadHandle == "" || *leadHandle == w.getHandle()
func (w BattleWorker) isLeader(leadHandle string) bool {
	return leadHandle == "" || leadHandle == w.getHandle()
}
func (w BattleWorker) Work(leadHandle *string, stopChan chan bool) {
	ticker := time.NewTicker(BATTLE_WORKER_DURATION_MILLIS * time.Millisecond)
	m := BattleMovementState{hWnd: w.hWnd, mode: w.movementMode}
	b := BattleActionState{hWnd: w.hWnd, state: HUMAN_ATTACK_NOT_YET, humanMode: NORMAL_ATTACK, petMode: NORMAL_ATTACK}

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				switch GetScene(w.hWnd) {
				case BATTLE_SCENE:
					fmt.Printf("Handle %s is in BATTLE_SCENE\n", w.getHandle())
					b.Attack()
				case NORMAL_SCENE:
					if w.isLeader(*leadHandle) {
						fmt.Printf("Handle %s is in NORMAL_SCENE\n", w.getHandle())
						m.Move()
					}
				}
			case <-stopChan:
				return
			default:
			}
		}
	}()
}

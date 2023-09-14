package game

import (
	"fmt"
	"log"
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

func (w BattleWorker) GetHandle() string {
	return fmt.Sprint(w.hWnd)
}

func (w *BattleWorker) SetMovementMode(movementMode BattleMovementMode) {
	w.movementMode = movementMode
}

func (w BattleWorker) doesMove(leadHandle string) bool {
	return (leadHandle == "" || leadHandle == w.GetHandle()) && w.movementMode != NONE
}
func (w BattleWorker) Work(leadHandle *string, stopChan chan bool) {
	ticker := time.NewTicker(BATTLE_WORKER_DURATION_MILLIS * time.Millisecond)
	m := BattleMovementState{hWnd: w.hWnd, mode: w.movementMode}
	b := BattleActionState{
		hWnd:        w.hWnd,
		humanStates: []HumanState{H_A_ATTACK},
		petStates:   []PetState{P_ATTACK},
	}

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				switch getScene(w.hWnd) {
				case BATTLE_SCENE:
					log.Printf("Handle %s is in BATTLE_SCENE\n", w.GetHandle())
					b.Attack()
				case NORMAL_SCENE:
					if w.doesMove(*leadHandle) {
						log.Printf("Handle %s is in NORMAL_SCENE\n", w.GetHandle())
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

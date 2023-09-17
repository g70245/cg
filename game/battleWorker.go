package game

import (
	"fmt"
	"time"

	. "github.com/lxn/win"
)

const (
	BATTLE_WORKER_INTERVAL = 800
	LOG_CHECKER_INTERVAL   = 100
)

type BattleWorker struct {
	hWnd          HWND
	MovementState BattleMovementState
	ActionState   BattleActionState
	logDir        *string
}

type BattleWorkers []BattleWorker

func CreateBattleWorkers(hWnds []HWND, logDir *string) BattleWorkers {
	var workers []BattleWorker
	for _, hWnd := range hWnds {
		workers = append(workers, BattleWorker{
			hWnd: hWnd,
			MovementState: BattleMovementState{
				hWnd: hWnd,
				Mode: NONE,
			},
			ActionState: CreateNewBattleActionState(hWnd),
			logDir:      logDir,
		})
	}
	return workers
}

func (w BattleWorker) GetHandle() string {
	return fmt.Sprint(w.hWnd)
}

func (w BattleWorker) canMove(leadHandle string) bool {
	return (leadHandle == "" || leadHandle == w.GetHandle()) && w.MovementState.Mode != NONE
}

func (w *BattleWorker) Work(leadHandle *string, stopChan chan bool) {
	workerTicker := time.NewTicker(BATTLE_WORKER_INTERVAL * time.Millisecond)
	checkerTicker := time.NewTicker(LOG_CHECKER_INTERVAL * time.Millisecond)

	checkLoopStopChan := make(chan bool, 1)
	isTransmittedChan := make(chan bool, 1)

	if *w.logDir != "" {
		defer checkerTicker.Stop()

		go func() {
			for {
				select {
				case <-checkLoopStopChan:
					return
				case <-checkerTicker.C:
					if isTransmittedToOtherMap(*w.logDir) {
						isTransmittedChan <- true
						return
					}
				default:
					time.Sleep(LOG_CHECKER_INTERVAL * time.Microsecond / 3)
				}
			}
		}()
	}

	go func() {
		defer workerTicker.Stop()
		isTransmitted := false

		for {
			select {
			case <-workerTicker.C:
				switch getScene(w.hWnd) {
				case BATTLE_SCENE:
					if !w.ActionState.Started && !isTransmitted {
						go w.ActionState.Attack()
					}
				case NORMAL_SCENE:
					if w.canMove(*leadHandle) {
						w.MovementState.Move()
					}
				default:
					// do nothing
				}
			case <-stopChan:
				w.ActionState.HandleStartedBySelf()
				return
			case isTransmitted = <-isTransmittedChan:
				checkLoopStopChan <- true
			default:
				time.Sleep(BATTLE_WORKER_INTERVAL * time.Microsecond / 3)
			}
		}
	}()
}

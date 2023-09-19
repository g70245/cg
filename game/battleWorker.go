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
	closeAllWindow(w.hWnd)

	workerTicker := time.NewTicker(BATTLE_WORKER_INTERVAL * time.Millisecond)

	logCheckerStopChan := make(chan bool, 1)
	isTPedChan := make(chan bool, 1)

	if *w.logDir != "" {
		checkerTicker := time.NewTicker(LOG_CHECKER_INTERVAL * time.Millisecond)

		go func() {
			defer checkerTicker.Stop()

			for {
				select {
				case <-logCheckerStopChan:
					return
				case <-checkerTicker.C:
					if isTPedToOtherMap(*w.logDir) {
						isTPedChan <- true
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
		w.ActionState.Enabled = true
		isTPed := false

		for {
			select {
			case <-workerTicker.C:
				switch getScene(w.hWnd) {
				case BATTLE_SCENE:
					if w.ActionState.Enabled {
						w.ActionState.Act()
					}
				case NORMAL_SCENE:
					if w.canMove(*leadHandle) && !isTPed {
						w.MovementState.Move()
					}
				default:
					// do nothing
				}
			case <-stopChan:
				w.ActionState.Enabled = false
				return
			case isTPed = <-isTPedChan:
				logCheckerStopChan <- true
			default:
				time.Sleep(BATTLE_WORKER_INTERVAL * time.Microsecond / 3)
			}
		}
	}()
}

package game

import (
	. "cg/system"
	"log"

	"fmt"
	"time"

	. "github.com/lxn/win"
)

const (
	BATTLE_WORKER_INTERVAL               = 800
	BATTLE_RESULT_DISAPPEARING_TIME      = 2
	LOG_CHECKER_INTERVAL                 = 100
	ITEM_CHECKER_INTERVAL                = 30
	ITEM_CHECKER_WAITING_OTHERS_INTERVAL = 400
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
	packageCheckerTicker := time.NewTicker(ITEM_CHECKER_INTERVAL * time.Second)

	logCheckerStopChan := make(chan bool, 1)
	isTPedChan := make(chan bool, 1)

	if *w.logDir != "" {
		logCheckerTicker := time.NewTicker(LOG_CHECKER_INTERVAL * time.Millisecond)

		go func() {
			defer logCheckerTicker.Stop()

			for {
				select {
				case <-logCheckerStopChan:
					return
				case <-logCheckerTicker.C:
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
		w.ActionState.IsOutOfMana = false
		isTPed := false
		isPlayingBeeper := false
		isPackageFull := false

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
				PlayBeeper()
				logCheckerStopChan <- true
			case <-packageCheckerTicker.C:
				isPackageFull = CheckPackageFull(w.hWnd)
				log.Printf("Handle %d is package full: %t\n", w.hWnd, isPackageFull)
			default:
				if !isPlayingBeeper && (w.ActionState.IsOutOfMana || isPackageFull) {
					isPlayingBeeper = PlayBeeper()
				}
				time.Sleep(BATTLE_WORKER_INTERVAL * time.Microsecond / 3)
			}
		}
	}()
}

func CheckPackageFull(hWnd HWND) bool {
	time.Sleep(BATTLE_RESULT_DISAPPEARING_TIME * time.Second)
	closeAllWindow(hWnd)
	LeftClick(hWnd, GAME_WIDTH/2, GAME_HEIGHT/2)
	openWindowByShortcut(hWnd, 0x45)
	defer closeAllWindow(hWnd)
	defer time.Sleep(ITEM_CHECKER_WAITING_OTHERS_INTERVAL * time.Millisecond)

	if px, py, ok := getNItemWindowPos(hWnd); ok {
		return !isAnyItemColFree(hWnd, px, py)
	}
	return false
}

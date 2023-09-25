package game

import (
	. "cg/system"
	"log"

	"fmt"
	"time"

	. "github.com/lxn/win"
)

const (
	BATTLE_WORKER_INTERVAL                    = 800
	LOG_CHECKER_INTERVAL                      = 100
	INVENTORY_CHECKER_INTERVAL_SECOND         = 60
	INVENTORY_CHECKER_WAITING_OTHERS_INTERVAL = 400
)

type BattleWorker struct {
	hWnd                    HWND
	MovementState           BattleMovementState
	ActionState             BattleActionState
	logDir                  *string
	inventoryCheckerEnabled bool
}

type BattleWorkers []BattleWorker

func CreateBattleWorkers(hWnds []HWND, logDir *string, manaChecker *string) BattleWorkers {
	var workers []BattleWorker
	for _, hWnd := range hWnds {
		workers = append(workers, BattleWorker{
			hWnd: hWnd,
			MovementState: BattleMovementState{
				hWnd: hWnd,
				Mode: NONE,
			},
			ActionState: CreateNewBattleActionState(hWnd, logDir, manaChecker),
			logDir:      logDir,
		})
	}
	return workers
}

func (w BattleWorker) GetHandle() string {
	return fmt.Sprint(w.hWnd)
}

func (w BattleWorker) canMove() bool {
	return w.MovementState.Mode != NONE
}

func (w *BattleWorker) Work(stopChan chan bool) {
	closeAllWindow(w.hWnd)

	workerTicker := time.NewTicker(BATTLE_WORKER_INTERVAL * time.Millisecond)
	inventoryCheckerTicker := time.NewTicker(INVENTORY_CHECKER_INTERVAL_SECOND * time.Second)

	logCheckerStopChan := make(chan bool, 1)
	isTeleportedChan := make(chan bool, 1)
	w.activateLogChecker(logCheckerStopChan, isTeleportedChan)

	go func() {
		defer workerTicker.Stop()
		w.ActionState.Enabled = true
		w.ActionState.isOutOfMana = false
		w.ActionState.isEncounteringBaBy = false
		isTeleported := false
		isPlayingBeeper := false
		isInventoryFull := false

		for {
			select {
			case <-workerTicker.C:
				switch getScene(w.hWnd) {
				case BATTLE_SCENE:
					if w.ActionState.Enabled {
						w.ActionState.Act()
					}
				case NORMAL_SCENE:
					if w.canMove() && !isTeleported && !w.ActionState.isOutOfMana {
						w.MovementState.Move()
					}
				default:
					// do nothing
				}
			case <-stopChan:
				StopBeeper()
				return
			case isTeleported = <-isTeleportedChan:
				PlayBeeper()
				logCheckerStopChan <- true
				log.Println("Has been teleported, need to stop the movement")
			case <-inventoryCheckerTicker.C:
				if w.inventoryCheckerEnabled {
					isInventoryFull = checkInventory(w.hWnd)
					log.Printf("Handle %d is inventory full: %t\n", w.hWnd, isInventoryFull)
				}
			default:
				if !isPlayingBeeper && (w.ActionState.isOutOfMana || isInventoryFull) {
					isPlayingBeeper = PlayBeeper()
				}
				time.Sleep(BATTLE_WORKER_INTERVAL * time.Microsecond / 3)
			}
		}
	}()
}

func (w *BattleWorker) StopInventoryChecker() {
	w.inventoryCheckerEnabled = false
}

func (w *BattleWorker) StartInventoryChecker() {
	w.inventoryCheckerEnabled = true
}

func (w *BattleWorker) activateLogChecker(logCheckerStopChan chan bool, isTeleportedChan chan bool) {
	if *w.logDir != "" {
		logCheckerTicker := time.NewTicker(LOG_CHECKER_INTERVAL * time.Millisecond)

		go func() {
			log.Println("Log Checker enabled")
			defer logCheckerTicker.Stop()

			for {
				select {
				case <-logCheckerStopChan:
					log.Println("Log Checker disabled")
					return
				case <-logCheckerTicker.C:
					if isTeleportedToOtherMap(*w.logDir) {
						isTeleportedChan <- true
						return
					}
				default:
					time.Sleep(LOG_CHECKER_INTERVAL * time.Microsecond / 3)
				}
			}
		}()
	}
}

package game

import (
	sys "cg/system"

	"fmt"
	"log"
	"time"

	. "github.com/g70245/win"
)

const (
	BATTLE_WORKER_INTERVAL                    = 400
	INVENTORY_CHECKER_INTERVAL_SECOND         = 60
	INVENTORY_CHECKER_WAITING_OTHERS_INTERVAL = 400
)

type BattleWorker struct {
	hWnd                    HWND
	MovementState           BattleMovementState
	ActionState             BattleActionState
	logDir                  *string
	manaChecker             *string
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
			manaChecker: manaChecker,
		})
	}
	return workers
}

func (w BattleWorker) GetHandle() string {
	return fmt.Sprint(w.hWnd)
}

func (w *BattleWorker) Work(stopChan chan bool) {
	closeAllWindows(w.hWnd)

	workerTicker := time.NewTicker(BATTLE_WORKER_INTERVAL * time.Millisecond)
	inventoryCheckerTicker := time.NewTicker(INVENTORY_CHECKER_INTERVAL_SECOND * time.Second)

	teleportCheckerStopChan := make(chan bool, 1)
	isTeleportedChan := make(chan bool, 1)
	activateChecker(w.hWnd, teleportCheckerStopChan, isTeleportedChan, w.logDir)

	go func() {
		defer workerTicker.Stop()
		defer inventoryCheckerTicker.Stop()
		defer close(teleportCheckerStopChan)
		defer close(isTeleportedChan)

		w.reset()

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
					if w.MovementState.Mode != NONE && !w.ActionState.isOutOfMana && !w.ActionState.isOutOfHealthWhileCatching && !isTeleported {
						w.MovementState.Move()
					}
				default:
					// do nothing
				}
			case <-stopChan:
				sys.StopBeeper()
				teleportCheckerStopChan <- true
				return
			case isTeleported = <-isTeleportedChan:
				sys.PlayBeeper()
				teleportCheckerStopChan <- true
				log.Printf("Handle %d has been teleported to: %s\n", w.hWnd, getMapName(w.hWnd))
			case <-inventoryCheckerTicker.C:
				if w.inventoryCheckerEnabled {
					isInventoryFull = checkInventory(w.hWnd)
					log.Printf("Handle %d is inventory full: %t\n", w.hWnd, isInventoryFull)
				}
			default:
				if !isPlayingBeeper && (w.ActionState.isOutOfHealthWhileCatching || isInventoryFull) {
					isPlayingBeeper = sys.PlayBeeper()
				}
				time.Sleep(BATTLE_WORKER_INTERVAL * time.Microsecond / 3)
			}
		}
	}()
}

func (w *BattleWorker) reset() {
	w.ActionState.Enabled = true
	w.ActionState.isOutOfHealthWhileCatching = false
	w.ActionState.isOutOfMana = false
	w.ActionState.isEncounteringBaBy = false

	w.MovementState.origin = getCurrentGamePos(w.hWnd)
	log.Printf("Handle %d Auto Battle started at (%.f, %.f)\n", w.hWnd, w.MovementState.origin.x, w.MovementState.origin.y)
}

func (w *BattleWorker) StopInventoryChecker() {
	w.inventoryCheckerEnabled = false
}

func (w *BattleWorker) StartInventoryChecker() {
	w.inventoryCheckerEnabled = true
}

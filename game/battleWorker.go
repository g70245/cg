package game

import (
	sys "cg/system"

	"fmt"
	"log"
	"time"

	. "github.com/g70245/win"
)

const (
	BATTLE_WORKER_INTERVAL             = 400
	CHECKER_INTERVAL                   = 300
	CHECKER_INVENTORY_INTERVAL_SECOND  = 60
	CHECKER_INVENTORY_WAITING_INTERVAL = 400
)

type BattleWorker struct {
	hWnd           HWND
	MovementState  BattleMovementState
	ActionState    BattleActionState
	logDir         *string
	manaChecker    *string
	currentMapName string

	teleportAndResourceCheckerEnabled bool
	inventoryCheckerEnabled           bool
}

type BattleWorkers []BattleWorker

func CreateBattleWorkers(hWnds []HWND, logDir, manaChecker *string) BattleWorkers {
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
	inventoryCheckerTicker := time.NewTicker(CHECKER_INVENTORY_INTERVAL_SECOND * time.Second)
	teleportAndResourceCheckerTicker := time.NewTicker(CHECKER_INTERVAL * time.Millisecond)

	go func() {
		defer workerTicker.Stop()
		defer inventoryCheckerTicker.Stop()
		defer teleportAndResourceCheckerTicker.Stop()

		w.reset()
		log.Printf("Handle %d Auto Battle started at (%.f, %.f)\n", w.hWnd, w.MovementState.origin.x, w.MovementState.origin.y)
		log.Printf("Handle %d Current Location: %s\n", w.hWnd, w.currentMapName)

		isPlayingBeeper := false
		isInventoryFull := false
		isTeleportedOrOutOfResource := false

		for {
			select {
			case <-workerTicker.C:
				switch getScene(w.hWnd) {
				case BATTLE_SCENE:
					if w.ActionState.Enabled {
						w.ActionState.Act()
					}
				case NORMAL_SCENE:
					if w.MovementState.Mode != NONE && !w.ActionState.isOutOfMana && !w.ActionState.isOutOfHealthWhileCatching && !isTeleportedOrOutOfResource && !isInventoryFull {
						w.MovementState.Move()
					}
				default:
					// do nothing
				}
			case <-stopChan:
				sys.StopBeeper()
				return
			case <-inventoryCheckerTicker.C:
				if w.inventoryCheckerEnabled && !isInventoryFull {
					isInventoryFull = checkInventory(w.hWnd)
					log.Printf("Handle %d is inventory full: %t\n", w.hWnd, isInventoryFull)
				}
			case <-teleportAndResourceCheckerTicker.C:
				if w.teleportAndResourceCheckerEnabled && !isTeleportedOrOutOfResource {
					if newMapName := getMapName(w.hWnd); w.currentMapName != newMapName || isTeleported(*w.logDir) {
						isTeleportedOrOutOfResource = true
						log.Printf("Handle %d has been teleported to: %s\n", w.hWnd, getMapName(w.hWnd))
					}
					if isOutOfResource(*w.logDir) {
						isTeleportedOrOutOfResource = true
						log.Printf("Handle %d is out of resource\n", w.hWnd)
					}
				}
			default:
				if !isPlayingBeeper && (w.ActionState.isOutOfHealthWhileCatching || isInventoryFull || isTeleportedOrOutOfResource) {
					isPlayingBeeper = sys.PlayBeeper()
				}
				time.Sleep(BATTLE_WORKER_INTERVAL * time.Microsecond / 3)
			}
		}
	}()
}

func (w *BattleWorker) reset() {
	w.currentMapName = getMapName(w.hWnd)

	w.ActionState.Enabled = true
	w.ActionState.isOutOfHealthWhileCatching = false
	w.ActionState.isOutOfMana = false
	w.ActionState.isEncounteringAnyBaby = false

	w.MovementState.origin = getCurrentGamePos(w.hWnd)
}

func (w *BattleWorker) StopInventoryChecker() {
	w.inventoryCheckerEnabled = false
}

func (w *BattleWorker) StartInventoryChecker() {
	w.inventoryCheckerEnabled = true
}

func (w *BattleWorker) StopTeleportAndResourceChecker() {
	w.teleportAndResourceCheckerEnabled = false
}

func (w *BattleWorker) StartTeleportAndResourceChecker() {
	w.teleportAndResourceCheckerEnabled = true
	w.currentMapName = getMapName(w.hWnd)
}

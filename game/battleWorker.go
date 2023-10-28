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
	BATTLE_CHECKER_INTERVAL                   = 300
	BATTLE_CHECKER_INVENTORY_INTERVAL_SECOND  = 60
	BATTLE_CHECKER_INVENTORY_WAITING_INTERVAL = 400
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

	ActivityCheckerEnabled bool
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

func (b *BattleWorker) Work(stopChan chan bool) {
	closeAllWindows(b.hWnd)

	workerTicker := time.NewTicker(BATTLE_WORKER_INTERVAL * time.Millisecond)
	inventoryCheckerTicker := time.NewTicker(BATTLE_CHECKER_INVENTORY_INTERVAL_SECOND * time.Second)
	teleportAndResourceCheckerTicker := time.NewTicker(BATTLE_CHECKER_INTERVAL * time.Millisecond)

	go func() {
		defer workerTicker.Stop()
		defer inventoryCheckerTicker.Stop()
		defer teleportAndResourceCheckerTicker.Stop()

		b.reset()
		log.Printf("Handle %d Auto Battle started at (%.f, %.f)\n", b.hWnd, b.MovementState.origin.x, b.MovementState.origin.y)
		log.Printf("Handle %d Current Location: %s\n", b.hWnd, b.currentMapName)

		isPlayingBeeper := false
		isInventoryFull := false
		isTeleportedOrOutOfResource := false

		for {
			select {
			case <-workerTicker.C:
				switch getScene(b.hWnd) {
				case BATTLE_SCENE:
					if b.ActionState.Enabled {
						b.ActionState.Act()
					}
				case NORMAL_SCENE:
					if b.MovementState.Mode != NONE && !b.ActionState.isOutOfMana && !b.ActionState.isOutOfHealth && !isTeleportedOrOutOfResource && !isInventoryFull {
						b.MovementState.Move()
					}
				default:
					// do nothing
				}
			case <-stopChan:
				sys.StopBeeper()
				return
			case <-inventoryCheckerTicker.C:
				if b.inventoryCheckerEnabled && !isInventoryFull {
					isInventoryFull = checkInventory(b.hWnd)
					log.Printf("Handle %d is inventory full: %t\n", b.hWnd, isInventoryFull)
				}
			case <-teleportAndResourceCheckerTicker.C:
				if b.teleportAndResourceCheckerEnabled && !isTeleportedOrOutOfResource {
					if newMapName := getMapName(b.hWnd); b.currentMapName != newMapName || isTeleported(*b.logDir) {
						isTeleportedOrOutOfResource = true
						log.Printf("Handle %d has been teleported to: %s\n", b.hWnd, getMapName(b.hWnd))
					}
					if isOutOfResource(*b.logDir) {
						isTeleportedOrOutOfResource = true
						log.Printf("Handle %d is out of resource\n", b.hWnd)
					}
				}
			default:
				if !isPlayingBeeper && (b.ActionState.isOutOfHealth || isInventoryFull || isTeleportedOrOutOfResource) {
					isPlayingBeeper = sys.PlayBeeper()
				}
				time.Sleep(BATTLE_WORKER_INTERVAL * time.Microsecond / 3)
			}
		}
	}()
}

func (b *BattleWorker) reset() {
	b.currentMapName = getMapName(b.hWnd)

	b.ActionState.Enabled = true
	b.ActionState.isOutOfHealth = false
	b.ActionState.isOutOfMana = false
	b.ActionState.ActivityCheckerEnabled = b.ActivityCheckerEnabled

	b.MovementState.origin = getCurrentGamePos(b.hWnd)
}

func (b *BattleWorker) StopInventoryChecker() {
	b.inventoryCheckerEnabled = false
}

func (b *BattleWorker) StartInventoryChecker() {
	b.inventoryCheckerEnabled = true
}

func (b *BattleWorker) StopTeleportAndResourceChecker() {
	b.teleportAndResourceCheckerEnabled = false
}

func (b *BattleWorker) StartTeleportAndResourceChecker() {
	b.teleportAndResourceCheckerEnabled = true
	b.currentMapName = getMapName(b.hWnd)
}

package game

import (
	. "cg/system"

	"fmt"
	"log"
	"time"

	. "github.com/g70245/win"
)

const (
	BATTLE_WORKER_INTERVAL                    = 400
	BATTLE_CHECKER_INTERVAL                   = 300
	BATTLE_CHECKER_INVENTORY_INTERVAL_SEC     = 60
	BATTLE_CHECKER_INVENTORY_WAITING_INTERVAL = 400
)

type BattleWorker struct {
	hWnd            HWND
	logDir          *string
	manaChecker     *string
	isInventoryFull *bool
	stopChan        chan bool

	ActionState   BattleActionState
	MovementState BattleMovementState

	currentMapName string

	TeleportAndResourceCheckerEnabled bool
	InventoryCheckerEnabled           bool
	ActivityCheckerEnabled            bool

	workerTicker                     *time.Ticker
	inventoryCheckerTicker           *time.Ticker
	teleportAndResourceCheckerTicker *time.Ticker
}

type BattleWorkers []BattleWorker

func CreateBattleWorkers(hWnds []HWND, logDir, manaChecker *string, isInventoryFull *bool, stopChan chan bool) BattleWorkers {
	var workers []BattleWorker
	for _, hWnd := range hWnds {
		newWorkerTicker := time.NewTicker(time.Hour)
		newInventoryCheckerTicker := time.NewTicker(time.Hour)
		newTeleportAndResourceCheckerTicker := time.NewTicker(time.Hour)

		newWorkerTicker.Stop()
		newInventoryCheckerTicker.Stop()
		newTeleportAndResourceCheckerTicker.Stop()

		workers = append(workers, BattleWorker{
			hWnd:            hWnd,
			logDir:          logDir,
			manaChecker:     manaChecker,
			isInventoryFull: isInventoryFull,
			stopChan:        stopChan,
			ActionState:     CreateNewBattleActionState(hWnd, logDir, manaChecker),
			MovementState: BattleMovementState{
				hWnd: hWnd,
				Mode: NONE,
			},
			workerTicker:                     newWorkerTicker,
			inventoryCheckerTicker:           newInventoryCheckerTicker,
			teleportAndResourceCheckerTicker: newTeleportAndResourceCheckerTicker,
		})
	}
	return workers
}

func (w BattleWorker) GetHandle() string {
	return fmt.Sprint(w.hWnd)
}

func (b *BattleWorker) Work() {
	closeAllWindows(b.hWnd)

	b.workerTicker.Reset(BATTLE_WORKER_INTERVAL * time.Millisecond)
	b.inventoryCheckerTicker.Reset(BATTLE_CHECKER_INVENTORY_INTERVAL_SEC * time.Second)
	b.teleportAndResourceCheckerTicker.Reset(BATTLE_CHECKER_INTERVAL * time.Millisecond)

	go func() {
		defer b.StopTickers()

		b.reset()
		log.Printf("Handle %d Auto Battle started at (%.f, %.f)\n", b.hWnd, b.MovementState.origin.x, b.MovementState.origin.y)
		log.Printf("Handle %d Current Location: %s\n", b.hWnd, b.currentMapName)

		for {
			select {
			case <-b.workerTicker.C:
				switch getScene(b.hWnd) {
				case BATTLE_SCENE:
					b.ActionState.Act()
				case NORMAL_SCENE:
					if b.MovementState.Mode != NONE {
						if b.ActionState.isOutOfHealth || b.ActionState.isOutOfMana {
							b.StopTickers()
							Beeper.Play()
							break
						}

						b.MovementState.Move()
					}
				default:
					// do nothing
				}
			case <-b.inventoryCheckerTicker.C:
				if b.InventoryCheckerEnabled {
					if b.ActivityCheckerEnabled {
						if checkActivityInventory(b.hWnd) {
							log.Printf("Handle %d inventory is full\n", b.hWnd)
							b.setInventoryStatus(true)
							b.StopTickers()
							Beeper.Play()
						}
					} else if checkInventory(b.hWnd) {
						log.Printf("Handle %d inventory is full\n", b.hWnd)
						b.setInventoryStatus(true)
						b.StopTickers()
						Beeper.Play()
					}
				}
			case <-b.teleportAndResourceCheckerTicker.C:
				if b.TeleportAndResourceCheckerEnabled {
					if newMapName := getMapName(b.hWnd); b.currentMapName != newMapName || isTeleported(*b.logDir) {
						log.Printf("Handle %d has been teleported to: %s\n", b.hWnd, getMapName(b.hWnd))
						b.StopTickers()
						Beeper.Play()
					}
					if isOutOfResource(*b.logDir) {
						log.Printf("Handle %d is out of resource\n", b.hWnd)
						b.StopTickers()
						Beeper.Play()
					}
				}
			case <-b.stopChan:
				return
			}
		}
	}()
}

func (b *BattleWorker) StopTickers() {
	defer b.workerTicker.Stop()
	defer b.inventoryCheckerTicker.Stop()
	defer b.teleportAndResourceCheckerTicker.Stop()
}

func (b *BattleWorker) Stop() {
	b.ActionState.Enabled = false
	b.stopChan <- true
	Beeper.Stop()
}

func (b *BattleWorker) reset() {
	b.currentMapName = getMapName(b.hWnd)
	b.setInventoryStatus(false)

	b.ActionState.Enabled = true
	b.ActionState.isOutOfHealth = false
	b.ActionState.isOutOfMana = false
	b.ActionState.ActivityCheckerEnabled = b.ActivityCheckerEnabled

	b.MovementState.origin = getCurrentGamePos(b.hWnd)
}

func (b *BattleWorker) StopInventoryChecker() {
	b.InventoryCheckerEnabled = false
}

func (b *BattleWorker) StartInventoryChecker() {
	b.InventoryCheckerEnabled = true
}

func (b *BattleWorker) StopTeleportAndResourceChecker() {
	b.TeleportAndResourceCheckerEnabled = false
}

func (b *BattleWorker) StartTeleportAndResourceChecker() {
	b.TeleportAndResourceCheckerEnabled = true
	b.currentMapName = getMapName(b.hWnd)
}

func (b *BattleWorker) setInventoryStatus(isFull bool) {
	*b.isInventoryFull = isFull
}

func (b *BattleWorker) getInventoryStatus() bool {
	return *b.isInventoryFull
}

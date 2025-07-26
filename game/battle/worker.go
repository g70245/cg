package battle

import (
	"cg/game"
	"cg/game/enum/enemyorder"
	"cg/game/enum/movement"
	"cg/utils"
	"sync"

	"fmt"
	"log"
	"time"

	"github.com/g70245/win"
)

const (
	NO_MANA_CHECKER = "None"
)

const (
	DURATION_BATTLE_WORKER            = 400 * time.Millisecond
	DURATION_BATTLE_LAST_ACTION       = 1000 * time.Millisecond
	DURATION_BATTLE_CHECKER_LOG       = 300 * time.Millisecond
	DURATION_BATTLE_CHECKER_INVENTORY = 60 * time.Second
)

type Worker struct {
	hWnd                  win.HWND
	gameDir               *string
	manaChecker           *string
	sharedInventoryStatus *bool
	sharedStopChan        chan bool
	sharedWaitGroup       *sync.WaitGroup

	ActionState   ActionState
	MovementState MovementState

	currentMapName string

	TeleportAndResourceCheckerEnabled bool
	InventoryCheckerEnabled           bool
	ActivityCheckerEnabled            bool
	EnemyOrder                        enemyorder.EnemyOrder

	workerTicker                     *time.Ticker
	inventoryCheckerTicker           *time.Ticker
	teleportAndResourceCheckerTicker *time.Ticker

	isOutOfResource bool
}

type Workers []Worker

func CreateWorkers(games game.Games, gameDir, manaChecker *string, sharedInventoryStatus *bool, sharedStopChan chan bool, sharedWaitGroup *sync.WaitGroup) Workers {
	var workers []Worker
	for _, hWnd := range games.GetHWNDs() {
		newWorkerTicker := time.NewTicker(time.Hour)
		newInventoryCheckerTicker := time.NewTicker(time.Hour)
		newTeleportAndResourceCheckerTicker := time.NewTicker(time.Hour)

		newWorkerTicker.Stop()
		newInventoryCheckerTicker.Stop()
		newTeleportAndResourceCheckerTicker.Stop()

		workers = append(workers, Worker{
			hWnd:                  hWnd,
			gameDir:               gameDir,
			manaChecker:           manaChecker,
			sharedInventoryStatus: sharedInventoryStatus,
			sharedStopChan:        sharedStopChan,
			sharedWaitGroup:       sharedWaitGroup,
			ActionState:           CreateNewBattleActionState(hWnd, gameDir, manaChecker),
			MovementState: MovementState{
				hWnd: hWnd,
				Mode: movement.None,
			},
			workerTicker:                     newWorkerTicker,
			inventoryCheckerTicker:           newInventoryCheckerTicker,
			teleportAndResourceCheckerTicker: newTeleportAndResourceCheckerTicker,
		})
	}
	return workers
}

func (w Worker) GetHandle() win.HWND {
	return w.hWnd
}

func (w Worker) GetHandleString() string {
	return fmt.Sprint(w.hWnd)
}

func (w *Worker) Work() {
	game.CloseAllWindows(w.hWnd)

	w.workerTicker.Reset(DURATION_BATTLE_WORKER)
	w.inventoryCheckerTicker.Reset(DURATION_BATTLE_CHECKER_INVENTORY)
	w.teleportAndResourceCheckerTicker.Reset(DURATION_BATTLE_CHECKER_LOG)

	go func() {
		defer w.StopTickers()

		w.reset()
		log.Printf("Handle %d Auto Battle started at (%.f, %.f)\n", w.hWnd, w.MovementState.origin.X, w.MovementState.origin.Y)
		log.Printf("Handle %d Current Location: %s\n", w.hWnd, w.currentMapName)

		for {
			select {
			case <-w.workerTicker.C:
				switch game.GetScene(w.hWnd) {
				case game.BATTLE_SCENE:
					if w.isGrouping() {
						w.sharedWaitGroup.Add(1)
						w.ActionState.Act()
						w.sharedWaitGroup.Done()
					} else {
						w.ActionState.Act()
					}
				case game.NORMAL_SCENE:
					if w.MovementState.Mode == movement.None {
						break
					}

					if w.isOutOfResource || *w.sharedInventoryStatus || w.ActionState.isOutOfHealth || w.ActionState.isOutOfMana {
						w.StopTickers()
						utils.Beeper.Play()
						break
					}
					if w.isGrouping() {
						w.sharedWaitGroup.Wait()
						w.MovementState.Move()
					} else {
						w.MovementState.Move()
					}
				default:
					// do nothing
				}
			case <-w.inventoryCheckerTicker.C:
				if !w.InventoryCheckerEnabled {
					break
				}

				if w.ActivityCheckerEnabled && isInventoryFullForActivity(w.hWnd) {
					log.Printf("Handle %d inventory is full\n", w.hWnd)
					w.setSharedInventoryStatus(true)
					w.StopTickers()
					utils.Beeper.Play()
				} else if isInventoryFull(w.hWnd) {
					log.Printf("Handle %d inventory is full\n", w.hWnd)
					w.setSharedInventoryStatus(true)
					w.StopTickers()
					utils.Beeper.Play()
				}
			case <-w.teleportAndResourceCheckerTicker.C:
				if !w.TeleportAndResourceCheckerEnabled {
					break
				}

				if newMapName := game.GetMapName(w.hWnd); w.currentMapName != newMapName || game.IsTeleported(*w.gameDir) {
					log.Printf("Handle %d has been teleported to: %s\n", w.hWnd, game.GetMapName(w.hWnd))
					w.StopTickers()
					utils.Beeper.Play()
				}
				if w.isOutOfResource = game.IsOutOfResource(*w.gameDir); w.isOutOfResource {
					log.Printf("Handle %d is out of resource\n", w.hWnd)
					w.StopTickers()
					utils.Beeper.Play()
				}
				if game.IsVerificationTriggered(*w.gameDir) {
					log.Printf("Handle %d triggered the verification\n", w.hWnd)
					utils.Beeper.Play()
				}
			case <-w.sharedStopChan:
				log.Printf("Handle %d Auto Battle ended at (%.f, %.f)\n", w.hWnd, w.MovementState.origin.X, w.MovementState.origin.Y)
				return
			}
		}
	}()
}

func (w *Worker) StopTickers() {
	w.workerTicker.Stop()
	w.inventoryCheckerTicker.Stop()
	w.teleportAndResourceCheckerTicker.Stop()

	time.Sleep(DURATION_BATTLE_LAST_ACTION)
	w.ActionState.Act()
}

func (w *Worker) Stop() {
	w.ActionState.Enabled = false
	w.sharedStopChan <- true
	utils.Beeper.Stop()
}

func (w *Worker) reset() {
	w.currentMapName = game.GetMapName(w.hWnd)
	w.setSharedInventoryStatus(false)

	w.ActionState.Enabled = true
	w.ActionState.isOutOfHealth = false
	w.ActionState.isOutOfMana = false
	w.ActionState.ActivityCheckerEnabled = w.ActivityCheckerEnabled
	w.ActionState.EnemyOrder = w.EnemyOrder

	w.MovementState.origin = game.GetCurrentGamePos(w.hWnd)

	w.isOutOfResource = false
}

func (w *Worker) StopInventoryChecker() {
	w.InventoryCheckerEnabled = false
}

func (w *Worker) StartInventoryChecker() {
	w.InventoryCheckerEnabled = true
}

func (w *Worker) StopTeleportAndResourceChecker() {
	w.TeleportAndResourceCheckerEnabled = false
}

func (w *Worker) StartTeleportAndResourceChecker() {
	w.TeleportAndResourceCheckerEnabled = true
	w.currentMapName = game.GetMapName(w.hWnd)
}

func (w *Worker) setSharedInventoryStatus(isFull bool) {
	if w.isGrouping() {
		*w.sharedInventoryStatus = isFull
	}
}

func (w *Worker) isGrouping() bool {
	return *w.manaChecker != NO_MANA_CHECKER
}

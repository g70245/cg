package game

import (
	. "cg/utils"
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

type BattleWorker struct {
	hWnd                  win.HWND
	gameDir               *string
	manaChecker           *string
	sharedInventoryStatus *bool
	sharedStopChan        chan bool
	sharedWaitGroup       *sync.WaitGroup

	ActionState   BattleActionState
	MovementState BattleMovementState

	currentMapName string

	TeleportAndResourceCheckerEnabled bool
	InventoryCheckerEnabled           bool
	ActivityCheckerEnabled            bool

	workerTicker                     *time.Ticker
	inventoryCheckerTicker           *time.Ticker
	teleportAndResourceCheckerTicker *time.Ticker

	isOutOfResource bool
}

type BattleWorkers []BattleWorker

func CreateBattleWorkers(games Games, gameDir, manaChecker *string, sharedInventoryStatus *bool, sharedStopChan chan bool, sharedWaitGroup *sync.WaitGroup) BattleWorkers {
	var workers []BattleWorker
	for _, hWnd := range games.GetHWNDs() {
		newWorkerTicker := time.NewTicker(time.Hour)
		newInventoryCheckerTicker := time.NewTicker(time.Hour)
		newTeleportAndResourceCheckerTicker := time.NewTicker(time.Hour)

		newWorkerTicker.Stop()
		newInventoryCheckerTicker.Stop()
		newTeleportAndResourceCheckerTicker.Stop()

		workers = append(workers, BattleWorker{
			hWnd:                  hWnd,
			gameDir:               gameDir,
			manaChecker:           manaChecker,
			sharedInventoryStatus: sharedInventoryStatus,
			sharedStopChan:        sharedStopChan,
			sharedWaitGroup:       sharedWaitGroup,
			ActionState:           CreateNewBattleActionState(hWnd, gameDir, manaChecker),
			MovementState: BattleMovementState{
				hWnd: hWnd,
				Mode: None,
			},
			workerTicker:                     newWorkerTicker,
			inventoryCheckerTicker:           newInventoryCheckerTicker,
			teleportAndResourceCheckerTicker: newTeleportAndResourceCheckerTicker,
		})
	}
	return workers
}

func (w BattleWorker) GetHandle() win.HWND {
	return w.hWnd
}

func (w BattleWorker) GetHandleString() string {
	return fmt.Sprint(w.hWnd)
}

func (b *BattleWorker) Work() {
	closeAllWindows(b.hWnd)

	b.workerTicker.Reset(DURATION_BATTLE_WORKER)
	b.inventoryCheckerTicker.Reset(DURATION_BATTLE_CHECKER_INVENTORY)
	b.teleportAndResourceCheckerTicker.Reset(DURATION_BATTLE_CHECKER_LOG)

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
					if b.isGrouping() {
						b.sharedWaitGroup.Add(1)
						b.ActionState.Act()
						b.sharedWaitGroup.Done()
					} else {
						b.ActionState.Act()
					}
				case NORMAL_SCENE:
					if b.MovementState.Mode == None {
						break
					}

					if b.isOutOfResource || *b.sharedInventoryStatus || b.ActionState.isOutOfHealth || b.ActionState.isOutOfMana {
						b.StopTickers()
						Beeper.Play()
						break
					}
					if b.isGrouping() {
						b.sharedWaitGroup.Wait()
						b.MovementState.Move()
					} else {
						b.MovementState.Move()
					}
				default:
					// do nothing
				}
			case <-b.inventoryCheckerTicker.C:
				if !b.InventoryCheckerEnabled {
					break
				}

				if b.ActivityCheckerEnabled && b.isInventoryFullForActivity() {
					log.Printf("Handle %d inventory is full\n", b.hWnd)
					b.setSharedInventoryStatus(true)
					b.StopTickers()
					Beeper.Play()
				} else if isInventoryFull(b.hWnd) {
					log.Printf("Handle %d inventory is full\n", b.hWnd)
					b.setSharedInventoryStatus(true)
					b.StopTickers()
					Beeper.Play()
				}
			case <-b.teleportAndResourceCheckerTicker.C:
				if !b.TeleportAndResourceCheckerEnabled {
					break
				}

				if newMapName := getMapName(b.hWnd); b.currentMapName != newMapName || isTeleported(*b.gameDir) {
					log.Printf("Handle %d has been teleported to: %s\n", b.hWnd, getMapName(b.hWnd))
					b.StopTickers()
					Beeper.Play()
				}
				if b.isOutOfResource = isOutOfResource(*b.gameDir); b.isOutOfResource {
					log.Printf("Handle %d is out of resource\n", b.hWnd)
					b.StopTickers()
					Beeper.Play()
				}
				if isVerificationTriggered(*b.gameDir) {
					log.Printf("Handle %d triggered the verification\n", b.hWnd)
					Beeper.Play()
				}
			case <-b.sharedStopChan:
				log.Printf("Handle %d Auto Battle ended at (%.f, %.f)\n", b.hWnd, b.MovementState.origin.x, b.MovementState.origin.y)
				return
			}
		}
	}()
}

func (b *BattleWorker) StopTickers() {
	b.workerTicker.Stop()
	b.inventoryCheckerTicker.Stop()
	b.teleportAndResourceCheckerTicker.Stop()

	time.Sleep(DURATION_BATTLE_LAST_ACTION)
	b.ActionState.Act()
}

func (b *BattleWorker) Stop() {
	b.ActionState.Enabled = false
	b.sharedStopChan <- true
	Beeper.Stop()
}

func (b *BattleWorker) reset() {
	b.currentMapName = getMapName(b.hWnd)
	b.setSharedInventoryStatus(false)

	b.ActionState.Enabled = true
	b.ActionState.isOutOfHealth = false
	b.ActionState.isOutOfMana = false
	b.ActionState.ActivityCheckerEnabled = b.ActivityCheckerEnabled

	b.MovementState.origin = getCurrentGamePos(b.hWnd)

	b.isOutOfResource = false
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

func (b *BattleWorker) setSharedInventoryStatus(isFull bool) {
	if b.isGrouping() {
		*b.sharedInventoryStatus = isFull
	}
}

func (b *BattleWorker) isGrouping() bool {
	return *b.manaChecker != NO_MANA_CHECKER
}

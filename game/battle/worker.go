package battle

import (
	"cg/game"
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

	ActionState   BattleActionState
	MovementState MovementState

	currentMapName string

	TeleportAndResourceCheckerEnabled bool
	InventoryCheckerEnabled           bool
	ActivityCheckerEnabled            bool

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
				Mode: None,
			},
			workerTicker:                     newWorkerTicker,
			inventoryCheckerTicker:           newInventoryCheckerTicker,
			teleportAndResourceCheckerTicker: newTeleportAndResourceCheckerTicker,
		})
	}
	return workers
}

func (b Worker) GetHandle() win.HWND {
	return b.hWnd
}

func (b Worker) GetHandleString() string {
	return fmt.Sprint(b.hWnd)
}

func (b *Worker) Work() {
	game.CloseAllWindows(b.hWnd)

	b.workerTicker.Reset(DURATION_BATTLE_WORKER)
	b.inventoryCheckerTicker.Reset(DURATION_BATTLE_CHECKER_INVENTORY)
	b.teleportAndResourceCheckerTicker.Reset(DURATION_BATTLE_CHECKER_LOG)

	go func() {
		defer b.StopTickers()

		b.reset()
		log.Printf("Handle %d Auto Battle started at (%.f, %.f)\n", b.hWnd, b.MovementState.origin.X, b.MovementState.origin.Y)
		log.Printf("Handle %d Current Location: %s\n", b.hWnd, b.currentMapName)

		for {
			select {
			case <-b.workerTicker.C:
				switch game.GetScene(b.hWnd) {
				case game.BATTLE_SCENE:
					if b.isGrouping() {
						b.sharedWaitGroup.Add(1)
						b.ActionState.Act()
						b.sharedWaitGroup.Done()
					} else {
						b.ActionState.Act()
					}
				case game.NORMAL_SCENE:
					if b.MovementState.Mode == None {
						break
					}

					if b.isOutOfResource || *b.sharedInventoryStatus || b.ActionState.isOutOfHealth || b.ActionState.isOutOfMana {
						b.StopTickers()
						utils.Beeper.Play()
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

				if b.ActivityCheckerEnabled && isInventoryFullForActivity(b.hWnd) {
					log.Printf("Handle %d inventory is full\n", b.hWnd)
					b.setSharedInventoryStatus(true)
					b.StopTickers()
					utils.Beeper.Play()
				} else if isInventoryFull(b.hWnd) {
					log.Printf("Handle %d inventory is full\n", b.hWnd)
					b.setSharedInventoryStatus(true)
					b.StopTickers()
					utils.Beeper.Play()
				}
			case <-b.teleportAndResourceCheckerTicker.C:
				if !b.TeleportAndResourceCheckerEnabled {
					break
				}

				if newMapName := game.GetMapName(b.hWnd); b.currentMapName != newMapName || game.IsTeleported(*b.gameDir) {
					log.Printf("Handle %d has been teleported to: %s\n", b.hWnd, game.GetMapName(b.hWnd))
					b.StopTickers()
					utils.Beeper.Play()
				}
				if b.isOutOfResource = game.IsOutOfResource(*b.gameDir); b.isOutOfResource {
					log.Printf("Handle %d is out of resource\n", b.hWnd)
					b.StopTickers()
					utils.Beeper.Play()
				}
				if game.IsVerificationTriggered(*b.gameDir) {
					log.Printf("Handle %d triggered the verification\n", b.hWnd)
					utils.Beeper.Play()
				}
			case <-b.sharedStopChan:
				log.Printf("Handle %d Auto Battle ended at (%.f, %.f)\n", b.hWnd, b.MovementState.origin.X, b.MovementState.origin.Y)
				return
			}
		}
	}()
}

func (b *Worker) StopTickers() {
	b.workerTicker.Stop()
	b.inventoryCheckerTicker.Stop()
	b.teleportAndResourceCheckerTicker.Stop()

	time.Sleep(DURATION_BATTLE_LAST_ACTION)
	b.ActionState.Act()
}

func (b *Worker) Stop() {
	b.ActionState.Enabled = false
	b.sharedStopChan <- true
	utils.Beeper.Stop()
}

func (b *Worker) reset() {
	b.currentMapName = game.GetMapName(b.hWnd)
	b.setSharedInventoryStatus(false)

	b.ActionState.Enabled = true
	b.ActionState.isOutOfHealth = false
	b.ActionState.isOutOfMana = false
	b.ActionState.ActivityCheckerEnabled = b.ActivityCheckerEnabled

	b.MovementState.origin = game.GetCurrentGamePos(b.hWnd)

	b.isOutOfResource = false
}

func (b *Worker) StopInventoryChecker() {
	b.InventoryCheckerEnabled = false
}

func (b *Worker) StartInventoryChecker() {
	b.InventoryCheckerEnabled = true
}

func (b *Worker) StopTeleportAndResourceChecker() {
	b.TeleportAndResourceCheckerEnabled = false
}

func (b *Worker) StartTeleportAndResourceChecker() {
	b.TeleportAndResourceCheckerEnabled = true
	b.currentMapName = game.GetMapName(b.hWnd)
}

func (b *Worker) setSharedInventoryStatus(isFull bool) {
	if b.isGrouping() {
		*b.sharedInventoryStatus = isFull
	}
}

func (b *Worker) isGrouping() bool {
	return *b.manaChecker != NO_MANA_CHECKER
}

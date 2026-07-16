package battle

import (
	"cg/game"
	"cg/game/enum/enemy"
	"cg/game/enum/movement"
	"cg/utils"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
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

// ManaChecker is shared by every worker in a battle group. The UI may change
// the selected checker while workers are running, so all access is synchronized.
type ManaChecker struct {
	mu    sync.RWMutex
	value string
}

func NewManaChecker() *ManaChecker {
	return &ManaChecker{value: NO_MANA_CHECKER}
}

func (m *ManaChecker) Get() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.value
}

func (m *ManaChecker) Set(value string) {
	m.mu.Lock()
	m.value = value
	m.mu.Unlock()
}

type Worker struct {
	hWnd                  win.HWND
	gameDir               func() string
	manaChecker           *ManaChecker
	sharedInventoryStatus *atomic.Bool
	sharedStopChan        chan bool
	sharedWaitGroup       *sync.WaitGroup

	configMu         sync.RWMutex
	actionState      ActionState
	movementMode     movement.Mode
	customEnemyOrder []string

	teleportAndResourceCheckerEnabled atomic.Bool
	inventoryCheckerEnabled           atomic.Bool
	activityCheckerEnabled            atomic.Bool
	enabled                           atomic.Bool

	workerTicker                     *time.Ticker
	inventoryCheckerTicker           *time.Ticker
	teleportAndResourceCheckerTicker *time.Ticker
}

type Workers []*Worker

func CreateWorkers(games game.Games, gameDir func() string, manaChecker *ManaChecker, sharedInventoryStatus *atomic.Bool, sharedStopChan chan bool, sharedWaitGroup *sync.WaitGroup) Workers {
	workers := make(Workers, 0, len(games))
	for _, hWnd := range games.GetHWNDs() {
		newWorkerTicker := time.NewTicker(time.Hour)
		newInventoryCheckerTicker := time.NewTicker(time.Hour)
		newTeleportAndResourceCheckerTicker := time.NewTicker(time.Hour)

		newWorkerTicker.Stop()
		newInventoryCheckerTicker.Stop()
		newTeleportAndResourceCheckerTicker.Stop()

		workers = append(workers, &Worker{
			hWnd:                             hWnd,
			gameDir:                          gameDir,
			manaChecker:                      manaChecker,
			sharedInventoryStatus:            sharedInventoryStatus,
			sharedStopChan:                   sharedStopChan,
			sharedWaitGroup:                  sharedWaitGroup,
			actionState:                      CreateNewBattleActionState(hWnd),
			movementMode:                     movement.None,
			workerTicker:                     newWorkerTicker,
			inventoryCheckerTicker:           newInventoryCheckerTicker,
			teleportAndResourceCheckerTicker: newTeleportAndResourceCheckerTicker,
		})
	}
	return workers
}

func (w *Worker) GetHandle() win.HWND {
	return w.hWnd
}

func (w *Worker) GetHandleString() string {
	return fmt.Sprint(w.hWnd)
}

func (w *Worker) Work() {
	game.CloseAllWindows(w.hWnd)

	w.workerTicker.Reset(DURATION_BATTLE_WORKER)
	w.inventoryCheckerTicker.Reset(DURATION_BATTLE_CHECKER_INVENTORY)
	w.teleportAndResourceCheckerTicker.Reset(DURATION_BATTLE_CHECKER_LOG)

	actionState, movementState, customEnemyOrder := w.runtimeSnapshot()
	w.enabled.Store(true)

	go func() {
		defer w.stopTickers(&actionState)

		currentMapName := game.GetMapName(w.hWnd)
		w.sharedInventoryStatus.Store(false)
		actionState.configureRuntime(w.enabled.Load, w.activityCheckerEnabled.Load, w.gameDir, w.manaChecker)
		actionState.reset()

		var enemies []game.CheckTarget
		for _, value := range customEnemyOrder {
			enemies = append(enemies, EnemyEnumMap[enemy.Position(value)])
		}
		actionState.CustomEnemies = enemies
		movementState.origin = game.GetCurrentGamePos(w.hWnd)

		teleportCheckerActive := w.teleportAndResourceCheckerEnabled.Load()
		isOutOfResource := false

		log.Printf("Handle %d Auto Battle started at (%.f, %.f)\n", w.hWnd, movementState.origin.X, movementState.origin.Y)
		log.Printf("Handle %d Current Location: %s\n", w.hWnd, currentMapName)

		for {
			select {
			case <-w.workerTicker.C:
				switch game.GetScene(w.hWnd) {
				case game.BATTLE_SCENE:
					if w.isGrouping() {
						// Magic Baby uses turn-based party battles. A character that
						// returns to the normal scene waits below until every party
						// window leaves battle, preventing the leader from moving into
						// another encounter while teammates are still in battle.
						w.sharedWaitGroup.Add(1)
						func() {
							defer w.sharedWaitGroup.Done()
							actionState.Act()
						}()
					} else {
						actionState.Act()
					}
				case game.NORMAL_SCENE:
					mode := w.MovementMode()
					if mode == movement.None {
						break
					}

					if isOutOfResource || w.sharedInventoryStatus.Load() || actionState.isOutOfHealth || actionState.isOutOfMana {
						w.stopTickers(&actionState)
						utils.Beeper.Play()
						break
					}
					movementState.Mode = mode
					if w.isGrouping() {
						w.sharedWaitGroup.Wait()
					}
					movementState.Move()
				default:
					// do nothing
				}
			case <-w.inventoryCheckerTicker.C:
				if !w.inventoryCheckerEnabled.Load() {
					break
				}

				if w.activityCheckerEnabled.Load() && isInventoryFullForActivity(w.hWnd) {
					log.Printf("Handle %d inventory is full\n", w.hWnd)
					w.setSharedInventoryStatus(true)
					w.stopTickers(&actionState)
					utils.Beeper.Play()
				} else if isInventoryFull(w.hWnd) {
					log.Printf("Handle %d inventory is full\n", w.hWnd)
					w.setSharedInventoryStatus(true)
					w.stopTickers(&actionState)
					utils.Beeper.Play()
				}
			case <-w.teleportAndResourceCheckerTicker.C:
				checkerEnabled := w.teleportAndResourceCheckerEnabled.Load()
				if !checkerEnabled {
					teleportCheckerActive = false
					break
				}
				if !teleportCheckerActive {
					currentMapName = game.GetMapName(w.hWnd)
					teleportCheckerActive = true
					break
				}

				if newMapName := game.GetMapName(w.hWnd); currentMapName != newMapName || game.IsTeleported(w.gameDir()) {
					log.Printf("Handle %d has been teleported to: %s\n", w.hWnd, newMapName)
					w.stopTickers(&actionState)
					utils.Beeper.Play()
				}
				if isOutOfResource = game.IsOutOfResource(w.gameDir()); isOutOfResource {
					log.Printf("Handle %d is out of resource\n", w.hWnd)
					w.stopTickers(&actionState)
					utils.Beeper.Play()
				}
				if game.IsVerificationTriggered(w.gameDir()) {
					log.Printf("Handle %d triggered the verification\n", w.hWnd)
					utils.Beeper.Play()
				}
			case <-w.sharedStopChan:
				log.Printf("Handle %d Auto Battle ended at (%.f, %.f)\n", w.hWnd, movementState.origin.X, movementState.origin.Y)
				return
			}
		}
	}()
}

func (w *Worker) stopTickers(actionState *ActionState) {
	w.workerTicker.Stop()
	w.inventoryCheckerTicker.Stop()
	w.teleportAndResourceCheckerTicker.Stop()

	time.Sleep(DURATION_BATTLE_LAST_ACTION)
	actionState.Act()
}

func (w *Worker) Stop() {
	w.enabled.Store(false)
	w.sharedStopChan <- true
	utils.Beeper.Stop()
}

func (w *Worker) StopInventoryChecker() {
	w.inventoryCheckerEnabled.Store(false)
}

func (w *Worker) StartInventoryChecker() {
	w.inventoryCheckerEnabled.Store(true)
}

func (w *Worker) SetActivityCheckerEnabled(enabled bool) {
	w.activityCheckerEnabled.Store(enabled)
}

func (w *Worker) StopTeleportAndResourceChecker() {
	w.teleportAndResourceCheckerEnabled.Store(false)
}

func (w *Worker) StartTeleportAndResourceChecker() {
	w.teleportAndResourceCheckerEnabled.Store(true)
}

func (w *Worker) SetMovementMode(mode movement.Mode) {
	w.configMu.Lock()
	w.movementMode = mode
	w.configMu.Unlock()
}

func (w *Worker) MovementMode() movement.Mode {
	w.configMu.RLock()
	defer w.configMu.RUnlock()
	return w.movementMode
}

func (w *Worker) SetCustomEnemyOrder(order []string) {
	w.configMu.Lock()
	w.customEnemyOrder = append([]string(nil), order...)
	w.configMu.Unlock()
}

func (w *Worker) CustomEnemyOrder() []string {
	w.configMu.RLock()
	defer w.configMu.RUnlock()
	return append([]string(nil), w.customEnemyOrder...)
}

func (w *Worker) UpdateActionState(update func(*ActionState)) {
	w.configMu.Lock()
	defer w.configMu.Unlock()
	update(&w.actionState)
}

func (w *Worker) ReplaceActionState(actionState ActionState) {
	w.configMu.Lock()
	actionState.hWnd = w.hWnd
	w.actionState = actionState.clone()
	w.configMu.Unlock()
}

func (w *Worker) ActionStateSnapshot() ActionState {
	w.configMu.RLock()
	defer w.configMu.RUnlock()
	return w.actionState.clone()
}

func (w *Worker) runtimeSnapshot() (ActionState, MovementState, []string) {
	w.configMu.RLock()
	defer w.configMu.RUnlock()
	return w.actionState.clone(), MovementState{hWnd: w.hWnd, Mode: w.movementMode}, append([]string(nil), w.customEnemyOrder...)
}

func (w *Worker) setSharedInventoryStatus(isFull bool) {
	if w.isGrouping() {
		w.sharedInventoryStatus.Store(isFull)
	}
}

func (w *Worker) isGrouping() bool {
	return w.manaChecker.Get() != NO_MANA_CHECKER
}

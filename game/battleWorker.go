package game

import (
	"fmt"
	"time"

	. "github.com/lxn/win"
)

const BATTLE_WORKER_DURATION_MILLIS = 800

type BattleWorker struct {
	hWnd          HWND
	MovementState BattleMovementState
	ActionState   BattleActionState
}

type BattleWorkers []BattleWorker

func CreateBattleWorkers(hWnds []HWND) BattleWorkers {
	var workers []BattleWorker
	for _, hWnd := range hWnds {
		workers = append(workers, BattleWorker{
			hWnd: hWnd,
			MovementState: BattleMovementState{
				hWnd: hWnd,
				Mode: NONE,
			},
			ActionState: BattleActionState{
				hWnd:          hWnd,
				HumanStates:   []string{H_A_ATTACK},
				HumanSkillIds: make(map[int]string),
				PetStates:     []string{P_ATTACK},
				PetSkillIds:   make(map[int]string),
			},
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
func (w BattleWorker) Work(leadHandle *string, stopChan chan bool) {
	ticker := time.NewTicker(BATTLE_WORKER_DURATION_MILLIS * time.Millisecond)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				switch getScene(w.hWnd) {
				case BATTLE_SCENE:
					// log.Printf("Handle %s is in BATTLE_SCENE\n", w.GetHandle())
					if !w.ActionState.Started {
						go w.ActionState.Attack()
					}
				case NORMAL_SCENE:
					if w.canMove(*leadHandle) {
						// log.Printf("Handle %s is in NORMAL_SCENE\n", w.GetHandle())
						w.MovementState.Move()
					}
				default:
					// do nothing
				}
			case <-stopChan:
				w.ActionState.HandleStartedBySelf()
				return
			default:
			}
		}
	}()
}

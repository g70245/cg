package battle

import (
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"cg/game/enum/character"
	"cg/game/enum/movement"
)

func TestManaCheckerConcurrentAccess(t *testing.T) {
	checker := NewManaChecker()
	var waitGroup sync.WaitGroup

	for i := 0; i < 50; i++ {
		i := i
		waitGroup.Add(2)
		go func() {
			defer waitGroup.Done()
			checker.Set(fmt.Sprint(i))
		}()
		go func() {
			defer waitGroup.Done()
			_ = checker.Get()
		}()
	}

	waitGroup.Wait()
}

func TestWorkerBeginWorkAllowsOnlyOneConcurrentStart(t *testing.T) {
	worker := &Worker{}
	const attempts = 50

	start := make(chan struct{})
	results := make(chan bool, attempts)
	var waitGroup sync.WaitGroup
	for i := 0; i < attempts; i++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			<-start
			results <- worker.beginWork()
		}()
	}

	close(start)
	waitGroup.Wait()
	close(results)

	successes := 0
	for result := range results {
		if result {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("successful starts = %d, want 1", successes)
	}
}

func TestWorkerCanBeginWorkAgainAfterFinish(t *testing.T) {
	worker := &Worker{}
	if !worker.beginWork() {
		t.Fatal("first beginWork() = false, want true")
	}
	if worker.beginWork() {
		t.Fatal("second beginWork() = true while running, want false")
	}

	worker.finishWork()
	if !worker.beginWork() {
		t.Fatal("beginWork() after finishWork() = false, want true")
	}
}

func TestWorkerActionStateSnapshotIsIndependent(t *testing.T) {
	worker := &Worker{
		hWnd:                  1,
		actionState:           CreateNewBattleActionState(1),
		movementMode:          movement.None,
		manaChecker:           NewManaChecker(),
		sharedInventoryStatus: &atomic.Bool{},
	}

	snapshot := worker.ActionStateSnapshot()
	snapshot.CharacterActions[0].Param = "changed"
	snapshot.PetActions[0].Param = "changed"

	stored := worker.ActionStateSnapshot()
	if stored.CharacterActions[0].Param != "" {
		t.Fatalf("character action snapshot changed stored state: %q", stored.CharacterActions[0].Param)
	}
	if stored.PetActions[0].Param != "" {
		t.Fatalf("pet action snapshot changed stored state: %q", stored.PetActions[0].Param)
	}
}

func TestWorkerConcurrentConfigurationAccess(t *testing.T) {
	checker := NewManaChecker()
	checker.Set("1")
	sharedInventoryStatus := &atomic.Bool{}
	worker := &Worker{
		hWnd:                  1,
		actionState:           CreateNewBattleActionState(1),
		movementMode:          movement.None,
		manaChecker:           checker,
		sharedInventoryStatus: sharedInventoryStatus,
	}

	var waitGroup sync.WaitGroup
	for i := 0; i < 50; i++ {
		i := i
		waitGroup.Add(2)
		go func() {
			defer waitGroup.Done()
			worker.SetMovementMode(movement.DIAGONAL)
			worker.SetCustomEnemyOrder([]string{fmt.Sprintf("T%d", i%5+1)})
			worker.UpdateActionState(func(actionState *ActionState) {
				actionState.ClearCharacterActions()
				actionState.AddCharacterAction(character.Attack)
			})
			worker.SetActivityCheckerEnabled(i%2 == 0)
			worker.setSharedInventoryStatus(i%2 == 0)
		}()
		go func() {
			defer waitGroup.Done()
			_ = worker.MovementMode()
			_ = worker.CustomEnemyOrder()
			_ = worker.ActionStateSnapshot()
			_ = worker.activityCheckerEnabled.Load()
			_ = sharedInventoryStatus.Load()
		}()
	}

	waitGroup.Wait()
}

func TestActionStateJSONRoundTrip(t *testing.T) {
	actionState := CreateNewBattleActionState(1)
	actionState.AddCharacterAction(character.Defend)

	encoded, err := json.Marshal(actionState)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var decoded ActionState
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if len(decoded.CharacterActions) != len(actionState.CharacterActions) {
		t.Fatalf("character action count = %d, want %d", len(decoded.CharacterActions), len(actionState.CharacterActions))
	}
	if len(decoded.PetActions) != len(actionState.PetActions) {
		t.Fatalf("pet action count = %d, want %d", len(decoded.PetActions), len(actionState.PetActions))
	}
}

package battle

import (
	"errors"
	"sync"
	"testing"

	"cg/game"

	"github.com/g70245/win"
)

func TestCharacterHealthRatioForRiding(t *testing.T) {
	tests := []struct {
		steps uint32
		want  float32
	}{
		{steps: 0, want: 0.6},
		{steps: 1, want: 0.6},
		{steps: 149, want: 0.6},
		{steps: 150, want: 0.3},
		{steps: 480, want: 0.3},
	}

	for _, test := range tests {
		if got := characterHealthRatioForRiding(0.6, test.steps); got != test.want {
			t.Fatalf("characterHealthRatioForRiding(0.6, %d) = %v, want %v", test.steps, got, test.want)
		}
	}
}

func TestVitalsMonitorChecksCharacterHealthAndManaWithOneRead(t *testing.T) {
	reads := 0
	monitor := newVitalsMonitorWith(
		[]win.HWND{1},
		func(win.HWND) (game.CharacterStatus, error) {
			reads++
			return game.CharacterStatus{CurrentHP: 80, MaximumHP: 100, CurrentMP: 10, MaximumMP: 100}, nil
		},
		func(win.HWND) (uint32, error) { return 0, nil },
		func(win.HWND, int) (game.PetStatus, error) { return game.PetStatus{}, nil },
	)
	monitor.SetCharacterHealth(true, 0.5)
	monitor.SetCharacterMana(true, 0.2)

	lower, err := monitor.Check()
	if err != nil || !lower {
		t.Fatalf("Check() = (%t, %v), want (true, nil)", lower, err)
	}
	if reads != 1 {
		t.Fatalf("character reads = %d, want 1", reads)
	}
}

func TestVitalsMonitorChecksEveryCharacterHandle(t *testing.T) {
	read := make(map[win.HWND]int)
	monitor := newVitalsMonitorWith(
		[]win.HWND{1, 2},
		func(hWnd win.HWND) (game.CharacterStatus, error) {
			read[hWnd]++
			if hWnd == 2 {
				return game.CharacterStatus{CurrentHP: 40, MaximumHP: 100}, nil
			}
			return game.CharacterStatus{CurrentHP: 80, MaximumHP: 100}, nil
		},
		func(win.HWND) (uint32, error) { return 0, nil },
		func(win.HWND, int) (game.PetStatus, error) { return game.PetStatus{}, nil },
	)
	monitor.SetCharacterHealth(true, 0.5)

	lower, err := monitor.Check()
	if err != nil || !lower {
		t.Fatalf("Check() = (%t, %v), want (true, nil)", lower, err)
	}
	if read[1] != 1 || read[2] != 1 {
		t.Fatalf("character reads = %v, want both handles read once", read)
	}
}

func TestVitalsMonitorUsesHalfCharacterHealthRatioWhileRidingWithReserve(t *testing.T) {
	monitor := newVitalsMonitorWith(
		[]win.HWND{1},
		func(win.HWND) (game.CharacterStatus, error) {
			return game.CharacterStatus{CurrentHP: 40, MaximumHP: 100}, nil
		},
		func(win.HWND) (uint32, error) { return 150, nil },
		func(win.HWND, int) (game.PetStatus, error) { return game.PetStatus{}, nil },
	)
	monitor.SetCharacterHealth(true, 0.6)

	lower, err := monitor.Check()
	if err != nil || lower {
		t.Fatalf("Check() = (%t, %v), want (false, nil)", lower, err)
	}
}

func TestVitalsMonitorDoesNotReadRidingStepsForManaOnly(t *testing.T) {
	monitor := newVitalsMonitorWith(
		[]win.HWND{1},
		func(win.HWND) (game.CharacterStatus, error) {
			return game.CharacterStatus{CurrentMP: 40, MaximumMP: 100}, nil
		},
		func(win.HWND) (uint32, error) { return 0, errors.New("must not be called") },
		func(win.HWND, int) (game.PetStatus, error) { return game.PetStatus{}, nil },
	)
	monitor.SetCharacterMana(true, 0.5)

	lower, err := monitor.Check()
	if err != nil || !lower {
		t.Fatalf("Check() = (%t, %v), want (true, nil)", lower, err)
	}
}

func TestVitalsMonitorChecksOnlyActivePets(t *testing.T) {
	reads := 0
	monitor := newVitalsMonitorWith(
		[]win.HWND{1},
		func(win.HWND) (game.CharacterStatus, error) { return game.CharacterStatus{}, nil },
		func(win.HWND) (uint32, error) { return 0, nil },
		func(_ win.HWND, index int) (game.PetStatus, error) {
			reads++
			if index == 4 {
				return game.PetStatus{CurrentMP: 20, MaximumMP: 100, State: game.PET_STATE_ACTIVE}, nil
			}
			return game.PetStatus{State: 1}, nil
		},
	)
	monitor.SetPetMana(true, 0.5)

	lower, err := monitor.Check()
	if err != nil || !lower {
		t.Fatalf("Check() = (%t, %v), want (true, nil)", lower, err)
	}
	if reads != game.PET_SLOT_COUNT {
		t.Fatalf("pet reads = %d, want %d", reads, game.PET_SLOT_COUNT)
	}
}

func TestVitalsMonitorAllowsNoActivePets(t *testing.T) {
	monitor := newVitalsMonitorWith(
		[]win.HWND{1},
		func(win.HWND) (game.CharacterStatus, error) { return game.CharacterStatus{}, nil },
		func(win.HWND) (uint32, error) { return 0, nil },
		func(win.HWND, int) (game.PetStatus, error) { return game.PetStatus{State: 1}, nil },
	)
	monitor.SetPetHealth(true, 0.5)
	monitor.SetPetMana(true, 0.5)

	lower, err := monitor.Check()
	if err != nil || lower {
		t.Fatalf("Check() = (%t, %v), want (false, nil)", lower, err)
	}
}

func TestVitalsMonitorReturnsReadFailure(t *testing.T) {
	wantErr := errors.New("read failed")
	monitor := newVitalsMonitorWith(
		[]win.HWND{1},
		func(win.HWND) (game.CharacterStatus, error) { return game.CharacterStatus{}, wantErr },
		func(win.HWND) (uint32, error) { return 0, nil },
		func(win.HWND, int) (game.PetStatus, error) { return game.PetStatus{}, nil },
	)
	monitor.SetCharacterMana(true, 0.5)

	if _, err := monitor.Check(); !errors.Is(err, wantErr) {
		t.Fatalf("Check() error = %v, want %v", err, wantErr)
	}
}

func TestVitalsMonitorBlockAllowsOneWinnerAndReset(t *testing.T) {
	monitor := &VitalsMonitor{}
	const attempts = 50

	start := make(chan struct{})
	results := make(chan bool, attempts)
	var waitGroup sync.WaitGroup
	for i := 0; i < attempts; i++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			<-start
			results <- monitor.Block()
		}()
	}
	close(start)
	waitGroup.Wait()
	close(results)

	winners := 0
	for result := range results {
		if result {
			winners++
		}
	}
	if winners != 1 {
		t.Fatalf("Block() winners = %d, want 1", winners)
	}
	if !monitor.IsBlocked() {
		t.Fatal("IsBlocked() = false, want true")
	}
	monitor.Reset()
	if monitor.IsBlocked() {
		t.Fatal("IsBlocked() after Reset() = true, want false")
	}
}

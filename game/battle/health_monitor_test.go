package battle

import (
	"errors"
	"sync"
	"testing"

	"cg/game"

	"github.com/g70245/win"
)

func TestCharacterHealthRatio(t *testing.T) {
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
		if got := characterHealthRatio(0.6, test.steps); got != test.want {
			t.Fatalf("characterHealthRatio(0.6, %d) = %v, want %v", test.steps, got, test.want)
		}
	}
}

func TestHealthMonitorChecksEveryCharacterHandle(t *testing.T) {
	read := make(map[win.HWND]int)
	monitor := newHealthMonitorWith(
		[]win.HWND{1, 2},
		func(hWnd win.HWND) (uint32, uint32, error) {
			read[hWnd]++
			if hWnd == 2 {
				return 40, 100, nil
			}
			return 80, 100, nil
		},
		func(win.HWND) (uint32, error) { return 0, nil },
		func(win.HWND, int) (game.PetStatus, error) { return game.PetStatus{}, nil },
	)
	monitor.SetCharacter(true, 0.5)

	lower, err := monitor.Check()
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !lower {
		t.Fatal("Check() lower = false, want true")
	}
	if read[1] != 1 || read[2] != 1 {
		t.Fatalf("character reads = %v, want both handles read once", read)
	}
}

func TestHealthMonitorUsesHalfCharacterRatioWhileRidingWithReserve(t *testing.T) {
	monitor := newHealthMonitorWith(
		[]win.HWND{1},
		func(win.HWND) (uint32, uint32, error) { return 40, 100, nil },
		func(win.HWND) (uint32, error) { return 150, nil },
		func(win.HWND, int) (game.PetStatus, error) { return game.PetStatus{}, nil },
	)
	monitor.SetCharacter(true, 0.6)

	lower, err := monitor.Check()
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if lower {
		t.Fatal("Check() lower = true, want false at adjusted ratio 0.3")
	}
}

func TestHealthMonitorChecksOnlyActivePets(t *testing.T) {
	read := 0
	monitor := newHealthMonitorWith(
		[]win.HWND{1},
		func(win.HWND) (uint32, uint32, error) { return 0, 0, nil },
		func(win.HWND) (uint32, error) { return 0, nil },
		func(_ win.HWND, index int) (game.PetStatus, error) {
			read++
			if index == 4 {
				return game.PetStatus{CurrentHP: 20, MaximumHP: 100, State: game.PET_STATE_ACTIVE}, nil
			}
			return game.PetStatus{State: 1}, nil
		},
	)
	monitor.SetPet(true, 0.5)

	lower, err := monitor.Check()
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !lower {
		t.Fatal("Check() lower = false, want true")
	}
	if read != game.PET_SLOT_COUNT {
		t.Fatalf("pet reads = %d, want %d", read, game.PET_SLOT_COUNT)
	}
}

func TestHealthMonitorAllowsNoActivePets(t *testing.T) {
	monitor := newHealthMonitorWith(
		[]win.HWND{1},
		func(win.HWND) (uint32, uint32, error) { return 0, 0, nil },
		func(win.HWND) (uint32, error) { return 0, nil },
		func(win.HWND, int) (game.PetStatus, error) { return game.PetStatus{State: 1}, nil },
	)
	monitor.SetPet(true, 0.5)

	lower, err := monitor.Check()
	if err != nil || lower {
		t.Fatalf("Check() = (%t, %v), want (false, nil)", lower, err)
	}
}

func TestHealthMonitorReturnsReadFailure(t *testing.T) {
	wantErr := errors.New("read failed")
	monitor := newHealthMonitorWith(
		[]win.HWND{1},
		func(win.HWND) (uint32, uint32, error) { return 0, 0, wantErr },
		func(win.HWND) (uint32, error) { return 0, nil },
		func(win.HWND, int) (game.PetStatus, error) { return game.PetStatus{}, nil },
	)
	monitor.SetCharacter(true, 0.5)

	if _, err := monitor.Check(); !errors.Is(err, wantErr) {
		t.Fatalf("Check() error = %v, want %v", err, wantErr)
	}
}

func TestHealthMonitorBlockAllowsOneWinnerAndReset(t *testing.T) {
	monitor := &HealthMonitor{}
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

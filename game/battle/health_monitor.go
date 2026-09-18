package battle

import (
	"fmt"
	"math"
	"sync/atomic"

	"cg/game"

	"github.com/g70245/win"
)

const ridingHealthRatioStepThreshold = 150

type healthReader func(win.HWND) (uint32, uint32, error)
type ridingStepsReader func(win.HWND) (uint32, error)
type petStatusReader func(win.HWND, int) (game.PetStatus, error)

type healthMonitorSetting struct {
	enabled atomic.Bool
	ratio   atomic.Uint32
}

func (s *healthMonitorSetting) set(enabled bool, ratio float32) {
	if ratio > 0 {
		s.ratio.Store(math.Float32bits(ratio))
	}
	s.enabled.Store(enabled)
}

func (s *healthMonitorSetting) get() (bool, float32) {
	return s.enabled.Load(), math.Float32frombits(s.ratio.Load())
}

type HealthMonitor struct {
	hWnds []win.HWND

	character healthMonitorSetting
	pet       healthMonitorSetting
	blocked   atomic.Bool

	readCharacterHealth healthReader
	readRidingSteps     ridingStepsReader
	readPetStatus       petStatusReader
}

func NewHealthMonitor(games game.Games) *HealthMonitor {
	return newHealthMonitorWith(games.GetHWNDs(), game.ReadCharacterHealth, game.ReadRidingSteps, game.ReadPetStatus)
}

func newHealthMonitorWith(hWnds []win.HWND, readCharacterHealth healthReader, readRidingSteps ridingStepsReader, readPetStatus petStatusReader) *HealthMonitor {
	return &HealthMonitor{
		hWnds:               append([]win.HWND(nil), hWnds...),
		readCharacterHealth: readCharacterHealth,
		readRidingSteps:     readRidingSteps,
		readPetStatus:       readPetStatus,
	}
}

func (m *HealthMonitor) SetCharacter(enabled bool, ratio float32) {
	m.character.set(enabled, ratio)
}

func (m *HealthMonitor) SetPet(enabled bool, ratio float32) {
	m.pet.set(enabled, ratio)
}

func (m *HealthMonitor) Reset() {
	m.blocked.Store(false)
}

func (m *HealthMonitor) IsBlocked() bool {
	return m.blocked.Load()
}

func (m *HealthMonitor) Block() bool {
	return m.blocked.CompareAndSwap(false, true)
}

func (m *HealthMonitor) Check() (bool, error) {
	if enabled, ratio := m.character.get(); enabled {
		for _, hWnd := range m.hWnds {
			current, maximum, err := m.readCharacterHealth(hWnd)
			if err != nil {
				return false, fmt.Errorf("handle %d: %w", hWnd, err)
			}
			steps, err := m.readRidingSteps(hWnd)
			if err != nil {
				return false, fmt.Errorf("handle %d: %w", hWnd, err)
			}
			lower, err := isHealthRatioLowerThan(current, maximum, characterHealthRatio(ratio, steps))
			if err != nil {
				return false, fmt.Errorf("handle %d character health: %w", hWnd, err)
			}
			if lower {
				return true, nil
			}
		}
	}

	if enabled, ratio := m.pet.get(); enabled {
		for _, hWnd := range m.hWnds {
			for index := 0; index < game.PET_SLOT_COUNT; index++ {
				status, err := m.readPetStatus(hWnd, index)
				if err != nil {
					return false, fmt.Errorf("handle %d: %w", hWnd, err)
				}
				if status.State != game.PET_STATE_ACTIVE {
					continue
				}
				lower, err := isHealthRatioLowerThan(status.CurrentHP, status.MaximumHP, ratio)
				if err != nil {
					return false, fmt.Errorf("handle %d pet %d health: %w", hWnd, index, err)
				}
				if lower {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

func characterHealthRatio(ratio float32, ridingSteps uint32) float32 {
	if ridingSteps >= ridingHealthRatioStepThreshold {
		return ratio / 2
	}
	return ratio
}

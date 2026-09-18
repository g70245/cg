package battle

import (
	"fmt"
	"math"
	"sync/atomic"

	"cg/game"

	"github.com/g70245/win"
)

const ridingHealthRatioStepThreshold = 150

type characterStatusReader func(win.HWND) (game.CharacterStatus, error)
type ridingStepsReader func(win.HWND) (uint32, error)
type petStatusReader func(win.HWND, int) (game.PetStatus, error)

type vitalsMonitorSetting struct {
	enabled atomic.Bool
	ratio   atomic.Uint32
}

func (s *vitalsMonitorSetting) set(enabled bool, ratio float32) {
	if ratio > 0 {
		s.ratio.Store(math.Float32bits(ratio))
	}
	s.enabled.Store(enabled)
}

func (s *vitalsMonitorSetting) get() (bool, float32) {
	return s.enabled.Load(), math.Float32frombits(s.ratio.Load())
}

type VitalsMonitor struct {
	hWnds []win.HWND

	characterHealth vitalsMonitorSetting
	petHealth       vitalsMonitorSetting
	characterMana   vitalsMonitorSetting
	petMana         vitalsMonitorSetting
	blocked         atomic.Bool

	readCharacterStatus characterStatusReader
	readRidingSteps     ridingStepsReader
	readPetStatus       petStatusReader
}

func NewVitalsMonitor(games game.Games) *VitalsMonitor {
	return newVitalsMonitorWith(games.GetHWNDs(), game.ReadCharacterStatus, game.ReadRidingSteps, game.ReadPetStatus)
}

func newVitalsMonitorWith(hWnds []win.HWND, readCharacterStatus characterStatusReader, readRidingSteps ridingStepsReader, readPetStatus petStatusReader) *VitalsMonitor {
	return &VitalsMonitor{
		hWnds:               append([]win.HWND(nil), hWnds...),
		readCharacterStatus: readCharacterStatus,
		readRidingSteps:     readRidingSteps,
		readPetStatus:       readPetStatus,
	}
}

func (m *VitalsMonitor) SetCharacterHealth(enabled bool, ratio float32) {
	m.characterHealth.set(enabled, ratio)
}

func (m *VitalsMonitor) SetPetHealth(enabled bool, ratio float32) {
	m.petHealth.set(enabled, ratio)
}

func (m *VitalsMonitor) SetCharacterMana(enabled bool, ratio float32) {
	m.characterMana.set(enabled, ratio)
}

func (m *VitalsMonitor) SetPetMana(enabled bool, ratio float32) {
	m.petMana.set(enabled, ratio)
}

func (m *VitalsMonitor) Reset() {
	m.blocked.Store(false)
}

func (m *VitalsMonitor) IsBlocked() bool {
	return m.blocked.Load()
}

func (m *VitalsMonitor) Block() bool {
	return m.blocked.CompareAndSwap(false, true)
}

func (m *VitalsMonitor) Check() (bool, error) {
	characterHealthEnabled, characterHealthRatio := m.characterHealth.get()
	characterManaEnabled, characterManaRatio := m.characterMana.get()
	if characterHealthEnabled || characterManaEnabled {
		for _, hWnd := range m.hWnds {
			status, err := m.readCharacterStatus(hWnd)
			if err != nil {
				return false, fmt.Errorf("handle %d: %w", hWnd, err)
			}

			if characterHealthEnabled {
				steps, err := m.readRidingSteps(hWnd)
				if err != nil {
					return false, fmt.Errorf("handle %d: %w", hWnd, err)
				}
				lower, err := isRatioLowerThan(status.CurrentHP, status.MaximumHP, characterHealthRatioForRiding(characterHealthRatio, steps))
				if err != nil {
					return false, fmt.Errorf("handle %d character health: %w", hWnd, err)
				}
				if lower {
					return true, nil
				}
			}

			if characterManaEnabled {
				lower, err := isRatioLowerThan(status.CurrentMP, status.MaximumMP, characterManaRatio)
				if err != nil {
					return false, fmt.Errorf("handle %d character mana: %w", hWnd, err)
				}
				if lower {
					return true, nil
				}
			}
		}
	}

	petHealthEnabled, petHealthRatio := m.petHealth.get()
	petManaEnabled, petManaRatio := m.petMana.get()
	if petHealthEnabled || petManaEnabled {
		for _, hWnd := range m.hWnds {
			for index := 0; index < game.PET_SLOT_COUNT; index++ {
				status, err := m.readPetStatus(hWnd, index)
				if err != nil {
					return false, fmt.Errorf("handle %d: %w", hWnd, err)
				}
				if status.State != game.PET_STATE_ACTIVE {
					continue
				}

				if petHealthEnabled {
					lower, err := isRatioLowerThan(status.CurrentHP, status.MaximumHP, petHealthRatio)
					if err != nil {
						return false, fmt.Errorf("handle %d pet %d health: %w", hWnd, index, err)
					}
					if lower {
						return true, nil
					}
				}

				if petManaEnabled {
					lower, err := isRatioLowerThan(status.CurrentMP, status.MaximumMP, petManaRatio)
					if err != nil {
						return false, fmt.Errorf("handle %d pet %d mana: %w", hWnd, index, err)
					}
					if lower {
						return true, nil
					}
				}
			}
		}
	}

	return false, nil
}

func characterHealthRatioForRiding(ratio float32, ridingSteps uint32) float32 {
	if ridingSteps >= ridingHealthRatioStepThreshold {
		return ratio / 2
	}
	return ratio
}

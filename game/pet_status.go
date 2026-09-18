package game

import (
	"fmt"

	"cg/internal"

	"github.com/g70245/win"
)

const petStateSize = 1

type PetStatus struct {
	CurrentHP uint32
	MaximumHP uint32
	State     uint8
}

func ReadPetStatus(hWnd win.HWND, index int) (PetStatus, error) {
	base, err := petBaseAddress(index)
	if err != nil {
		return PetStatus{}, err
	}

	stateData, err := internal.ReadMemoryAtAddress(hWnd, base+MEMORY_PET_STATE_OFFSET, petStateSize)
	if err != nil {
		return PetStatus{}, fmt.Errorf("read pet %d state: %w", index, err)
	}

	status := PetStatus{State: stateData[0]}
	if status.State != PET_STATE_ACTIVE {
		return status, nil
	}

	healthData, err := internal.ReadMemoryAtAddress(hWnd, base, characterHealthSize)
	if err != nil {
		return PetStatus{}, fmt.Errorf("read pet %d health: %w", index, err)
	}
	status.CurrentHP, status.MaximumHP, err = decodeHealth(healthData)
	if err != nil {
		return PetStatus{}, fmt.Errorf("read pet %d health: %w", index, err)
	}
	return status, nil
}

func petBaseAddress(index int) (uint32, error) {
	if index < 0 || index >= PET_SLOT_COUNT {
		return 0, fmt.Errorf("pet index %d out of range", index)
	}
	return MEMORY_PET_BASE + uint32(index)*MEMORY_PET_STRIDE, nil
}

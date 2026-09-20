package game

import (
	"encoding/binary"
	"fmt"

	"cg/internal"

	"github.com/g70245/win"
)

func HasFlawlessPet(hWnd win.HWND) (bool, error) {
	return hasFlawlessPetWith(func(address uint32, size uint) ([]byte, error) {
		return internal.ReadMemoryAtAddress(hWnd, address, size)
	})
}

func hasFlawlessPetWith(readMemory processMemoryReader) (bool, error) {
	actors, err := readBattleEnemyActors(readMemory)
	if err != nil {
		return false, err
	}

	for index, actor := range actors {
		slot := BATTLE_ENEMY_SLOT_START + index
		if actor == 0 {
			continue
		}

		context, err := readBattleUint32(readMemory, actor+MEMORY_ACTOR_CONTEXT_OFFSET)
		if err != nil {
			return false, fmt.Errorf("read battle slot %d actor context: %w", slot, err)
		}
		if context == 0 {
			continue
		}

		layer, err := readBattleUint32(readMemory, context+MEMORY_CONTEXT_LAYER2_OFFSET)
		if err != nil {
			return false, fmt.Errorf("read battle slot %d layer: %w", slot, err)
		}
		if layer == 0 {
			continue
		}

		layerSize := uint(MEMORY_LAYER_RESOURCE_OFFSET + uint32MemorySize)
		layerData, err := readMemory(layer, layerSize)
		if err != nil {
			return false, fmt.Errorf("read battle slot %d layer data: %w", slot, err)
		}
		if len(layerData) < int(layerSize) {
			return false, fmt.Errorf("read battle slot %d layer data: got %d bytes, want %d", slot, len(layerData), layerSize)
		}

		parent := binary.LittleEndian.Uint32(layerData[MEMORY_LAYER_PARENT_OFFSET:])
		layerSlot := binary.LittleEndian.Uint32(layerData[MEMORY_LAYER_SLOT_OFFSET:])
		resourceID := binary.LittleEndian.Uint32(layerData[MEMORY_LAYER_RESOURCE_OFFSET:])
		if parent == actor && layerSlot == uint32(slot) && resourceID == FLAWLESS_PET_LAYER_RESOURCEID {
			return true, nil
		}
	}

	return false, nil
}

package game

import (
	"encoding/binary"
	"fmt"

	"cg/internal"

	"github.com/g70245/win"
)

const uint32MemorySize = 4

type processMemoryReader func(address uint32, size uint) ([]byte, error)

func HasFlawlessPet(hWnd win.HWND) (bool, error) {
	return hasFlawlessPetWith(func(address uint32, size uint) ([]byte, error) {
		return internal.ReadMemoryAtAddress(hWnd, address, size)
	})
}

func hasFlawlessPetWith(readMemory processMemoryReader) (bool, error) {
	actorTableAddress := uint32(MEMORY_BATTLE_ACTOR_TABLE + BATTLE_ENEMY_SLOT_START*uint32MemorySize)
	actorTableSize := uint(BATTLE_ENEMY_SLOT_COUNT * uint32MemorySize)
	actorTable, err := readMemory(actorTableAddress, actorTableSize)
	if err != nil {
		return false, fmt.Errorf("read enemy actor table: %w", err)
	}
	if len(actorTable) < int(actorTableSize) {
		return false, fmt.Errorf("read enemy actor table: got %d bytes, want %d", len(actorTable), actorTableSize)
	}

	for index := 0; index < BATTLE_ENEMY_SLOT_COUNT; index++ {
		slot := BATTLE_ENEMY_SLOT_START + index
		actor := binary.LittleEndian.Uint32(actorTable[index*uint32MemorySize:])
		if actor == 0 {
			continue
		}

		context, err := readFlawlessPetUint32(readMemory, actor+MEMORY_ACTOR_CONTEXT_OFFSET)
		if err != nil {
			return false, fmt.Errorf("read battle slot %d actor context: %w", slot, err)
		}
		if context == 0 {
			continue
		}

		layer, err := readFlawlessPetUint32(readMemory, context+MEMORY_CONTEXT_LAYER2_OFFSET)
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

func readFlawlessPetUint32(readMemory processMemoryReader, address uint32) (uint32, error) {
	data, err := readMemory(address, uint32MemorySize)
	if err != nil {
		return 0, err
	}
	if len(data) < uint32MemorySize {
		return 0, fmt.Errorf("got %d bytes, want %d", len(data), uint32MemorySize)
	}
	return binary.LittleEndian.Uint32(data), nil
}

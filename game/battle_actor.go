package game

import (
	"encoding/binary"
	"fmt"
)

const uint32MemorySize = 4

type processMemoryReader func(address uint32, size uint) ([]byte, error)

func readBattleEnemyActors(readMemory processMemoryReader) ([]uint32, error) {
	actorTableAddress := uint32(MEMORY_BATTLE_ACTOR_TABLE + BATTLE_ENEMY_SLOT_START*uint32MemorySize)
	actorTableSize := uint(BATTLE_ENEMY_SLOT_COUNT * uint32MemorySize)
	actorTable, err := readMemory(actorTableAddress, actorTableSize)
	if err != nil {
		return nil, fmt.Errorf("read enemy actor table: %w", err)
	}
	if len(actorTable) < int(actorTableSize) {
		return nil, fmt.Errorf("read enemy actor table: got %d bytes, want %d", len(actorTable), actorTableSize)
	}

	actors := make([]uint32, BATTLE_ENEMY_SLOT_COUNT)
	for index := range actors {
		actors[index] = binary.LittleEndian.Uint32(actorTable[index*uint32MemorySize:])
	}
	return actors, nil
}

func readBattleUint32(readMemory processMemoryReader, address uint32) (uint32, error) {
	data, err := readMemory(address, uint32MemorySize)
	if err != nil {
		return 0, err
	}
	if len(data) < uint32MemorySize {
		return 0, fmt.Errorf("got %d bytes, want %d", len(data), uint32MemorySize)
	}
	return binary.LittleEndian.Uint32(data), nil
}

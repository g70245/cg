package game

import (
	"encoding/binary"
	"errors"
	"fmt"
	"testing"
)

type levelOneMemory map[uint32][]byte

func (memory levelOneMemory) read(address uint32, size uint) ([]byte, error) {
	data, ok := memory[address]
	if !ok {
		return nil, fmt.Errorf("address %#x is unavailable", address)
	}
	if len(data) < int(size) {
		return nil, fmt.Errorf("address %#x has %d bytes, want %d", address, len(data), size)
	}
	return data[:size], nil
}

func newLevelOneMemory(levels map[int]uint32) levelOneMemory {
	memory := levelOneMemory{}
	actorTable := make([]byte, BATTLE_ENEMY_SLOT_COUNT*uint32MemorySize)
	for slot, level := range levels {
		actor := uint32(0x03000000 + slot*0x1000)
		binary.LittleEndian.PutUint32(actorTable[(slot-BATTLE_ENEMY_SLOT_START)*uint32MemorySize:], actor)

		levelData := make([]byte, uint32MemorySize)
		binary.LittleEndian.PutUint32(levelData, level)
		memory[actor+MEMORY_ACTOR_LEVEL_OFFSET] = levelData
	}
	memory[MEMORY_BATTLE_ACTOR_TABLE+BATTLE_ENEMY_SLOT_START*uint32MemorySize] = actorTable
	return memory
}

func TestHasLevelOneEnemyWith(t *testing.T) {
	tests := []struct {
		name       string
		memory     levelOneMemory
		want       bool
		wantErr    bool
		readMemory processMemoryReader
	}{
		{
			name:   "returns false for empty enemy slots",
			memory: newLevelOneMemory(nil),
		},
		{
			name:   "finds a level one enemy in a later slot",
			memory: newLevelOneMemory(map[int]uint32{12: 37, 18: 1}),
			want:   true,
		},
		{
			name:   "rejects enemies above level one",
			memory: newLevelOneMemory(map[int]uint32{10: 2, 15: 37}),
		},
		{
			name: "reports actor table read failures",
			readMemory: func(uint32, uint) ([]byte, error) {
				return nil, errors.New("read failed")
			},
			wantErr: true,
		},
		{
			name: "reports actor level read failures",
			readMemory: func(address uint32, size uint) ([]byte, error) {
				actorTableAddress := uint32(MEMORY_BATTLE_ACTOR_TABLE + BATTLE_ENEMY_SLOT_START*uint32MemorySize)
				if address == actorTableAddress {
					actorTable := make([]byte, BATTLE_ENEMY_SLOT_COUNT*uint32MemorySize)
					binary.LittleEndian.PutUint32(actorTable, 0x03000000)
					return actorTable, nil
				}
				return nil, errors.New("read failed")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			readMemory := tt.readMemory
			if readMemory == nil {
				readMemory = tt.memory.read
			}

			got, err := hasLevelOneEnemyWith(readMemory)
			if (err != nil) != tt.wantErr {
				t.Fatalf("hasLevelOneEnemyWith() error = %v, wantErr %t", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("hasLevelOneEnemyWith() = %t, want %t", got, tt.want)
			}
		})
	}
}

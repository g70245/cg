package game

import (
	"encoding/binary"
	"errors"
	"fmt"
	"testing"
)

const (
	testActor   = uint32(0x036CCC00)
	testContext = uint32(0x2514A848)
	testLayer   = uint32(0x036C8070)
)

type flawlessPetMemory map[uint32][]byte

func (memory flawlessPetMemory) read(address uint32, size uint) ([]byte, error) {
	data, ok := memory[address]
	if !ok {
		return nil, fmt.Errorf("address %#x is unavailable", address)
	}
	if len(data) < int(size) {
		return nil, fmt.Errorf("address %#x has %d bytes, want %d", address, len(data), size)
	}
	return data[:size], nil
}

func (memory flawlessPetMemory) putUint32(address uint32, value uint32) {
	data := make([]byte, uint32MemorySize)
	binary.LittleEndian.PutUint32(data, value)
	memory[address] = data
}

func newFlawlessPetMemory(slot int, parent uint32, layerSlot uint32, resourceID uint32) flawlessPetMemory {
	memory := flawlessPetMemory{}
	actorTable := make([]byte, BATTLE_ENEMY_SLOT_COUNT*uint32MemorySize)
	binary.LittleEndian.PutUint32(actorTable[(slot-BATTLE_ENEMY_SLOT_START)*uint32MemorySize:], testActor)
	memory[MEMORY_BATTLE_ACTOR_TABLE+BATTLE_ENEMY_SLOT_START*uint32MemorySize] = actorTable
	memory.putUint32(testActor+MEMORY_ACTOR_CONTEXT_OFFSET, testContext)
	memory.putUint32(testContext+MEMORY_CONTEXT_LAYER2_OFFSET, testLayer)

	layerData := make([]byte, MEMORY_LAYER_RESOURCE_OFFSET+uint32MemorySize)
	binary.LittleEndian.PutUint32(layerData[MEMORY_LAYER_PARENT_OFFSET:], parent)
	binary.LittleEndian.PutUint32(layerData[MEMORY_LAYER_SLOT_OFFSET:], layerSlot)
	binary.LittleEndian.PutUint32(layerData[MEMORY_LAYER_RESOURCE_OFFSET:], resourceID)
	memory[testLayer] = layerData
	return memory
}

func TestHasFlawlessPetWith(t *testing.T) {
	tests := []struct {
		name       string
		memory     flawlessPetMemory
		want       bool
		wantErr    bool
		readMemory processMemoryReader
	}{
		{
			name: "returns false for empty enemy slots",
			memory: flawlessPetMemory{
				MEMORY_BATTLE_ACTOR_TABLE + BATTLE_ENEMY_SLOT_START*uint32MemorySize: make([]byte, BATTLE_ENEMY_SLOT_COUNT*uint32MemorySize),
			},
		},
		{
			name:   "finds a flawless pet in a later enemy slot",
			memory: newFlawlessPetMemory(17, testActor, 17, FLAWLESS_PET_LAYER_RESOURCEID),
			want:   true,
		},
		{
			name:   "rejects a layer left by another actor",
			memory: newFlawlessPetMemory(15, testActor+0x1000, 15, FLAWLESS_PET_LAYER_RESOURCEID),
		},
		{
			name:   "rejects a layer assigned to another battle slot",
			memory: newFlawlessPetMemory(15, testActor, 14, FLAWLESS_PET_LAYER_RESOURCEID),
		},
		{
			name:   "rejects another layer resource",
			memory: newFlawlessPetMemory(15, testActor, 15, 0x12345),
		},
		{
			name: "reports memory read failures",
			readMemory: func(uint32, uint) ([]byte, error) {
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

			got, err := hasFlawlessPetWith(readMemory)
			if (err != nil) != tt.wantErr {
				t.Fatalf("hasFlawlessPetWith() error = %v, wantErr %t", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("hasFlawlessPetWith() = %t, want %t", got, tt.want)
			}
		})
	}
}

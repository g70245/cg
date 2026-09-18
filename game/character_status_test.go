package game

import (
	"encoding/binary"
	"testing"
)

func TestDecodeCharacterStatus(t *testing.T) {
	data := make([]byte, characterStatusSize)
	values := []uint32{2540, 3000, 480, 1079}
	for index, value := range values {
		start := index * xorValueSize
		encodeTestXORValue(data[start:start+xorValueSize], value, uint32(index+1)*0x11111111)
	}

	status, err := decodeCharacterStatus(data)
	if err != nil {
		t.Fatalf("decodeCharacterStatus() error = %v", err)
	}
	if status.CurrentHP != 2540 || status.MaximumHP != 3000 || status.CurrentMP != 480 || status.MaximumMP != 1079 {
		t.Fatalf("decodeCharacterStatus() = %+v", status)
	}
}

func TestDecodeCharacterStatusRejectsShortData(t *testing.T) {
	if _, err := decodeCharacterStatus(make([]byte, characterStatusSize-1)); err == nil {
		t.Fatal("decodeCharacterStatus() error = nil, want an error")
	}
}

func TestDecodeRidingSteps(t *testing.T) {
	data := make([]byte, ridingStepsSize)
	binary.LittleEndian.PutUint32(data, 480)

	steps, err := decodeRidingSteps(data)
	if err != nil {
		t.Fatalf("decodeRidingSteps() error = %v", err)
	}
	if steps != 480 {
		t.Fatalf("decodeRidingSteps() = %d, want 480", steps)
	}
}

func TestDecodeRidingStepsRejectsShortData(t *testing.T) {
	if _, err := decodeRidingSteps(make([]byte, ridingStepsSize-1)); err == nil {
		t.Fatal("decodeRidingSteps() error = nil, want an error")
	}
}

func encodeTestXORValue(data []byte, value, key uint32) {
	binary.LittleEndian.PutUint32(data[4:8], value^key)
	binary.LittleEndian.PutUint32(data[8:12], key)
}

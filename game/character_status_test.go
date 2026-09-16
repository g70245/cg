package game

import (
	"encoding/binary"
	"testing"
)

func TestDecodeCharacterHealth(t *testing.T) {
	data := make([]byte, characterHealthSize)
	encodeTestXORValue(data[:xorValueSize], 2540, 0x12345678)
	encodeTestXORValue(data[xorValueSize:], 3000, 0x87654321)

	current, maximum, err := decodeCharacterHealth(data)
	if err != nil {
		t.Fatalf("decodeCharacterHealth() error = %v", err)
	}
	if current != 2540 || maximum != 3000 {
		t.Fatalf("decodeCharacterHealth() = (%d, %d), want (2540, 3000)", current, maximum)
	}
}

func TestDecodeCharacterHealthRejectsShortData(t *testing.T) {
	if _, _, err := decodeCharacterHealth(make([]byte, characterHealthSize-1)); err == nil {
		t.Fatal("decodeCharacterHealth() error = nil, want an error")
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

package game

import (
	"encoding/binary"
	"fmt"

	"cg/internal"

	"github.com/g70245/win"
)

const (
	xorValueSize        = 16
	characterHealthSize = xorValueSize * 2
	ridingStepsSize     = 4
)

func ReadCharacterHealth(hWnd win.HWND) (uint32, uint32, error) {
	data, err := internal.ReadMemoryAtAddress(hWnd, MEMORY_CHARACTER_HEALTH, characterHealthSize)
	if err != nil {
		return 0, 0, fmt.Errorf("read character health: %w", err)
	}

	current, maximum, err := decodeCharacterHealth(data)
	if err != nil {
		return 0, 0, fmt.Errorf("read character health: %w", err)
	}
	return current, maximum, nil
}

func ReadRidingSteps(hWnd win.HWND) (uint32, error) {
	data, err := internal.ReadMemoryAtAddress(hWnd, MEMORY_RIDING_REMAINING_STEPS, ridingStepsSize)
	if err != nil {
		return 0, fmt.Errorf("read riding steps: %w", err)
	}

	steps, err := decodeRidingSteps(data)
	if err != nil {
		return 0, fmt.Errorf("read riding steps: %w", err)
	}
	return steps, nil
}

func decodeCharacterHealth(data []byte) (uint32, uint32, error) {
	current, maximum, err := decodeHealth(data)
	if err != nil {
		return 0, 0, fmt.Errorf("decode character health: %w", err)
	}
	return current, maximum, nil
}

func decodeHealth(data []byte) (uint32, uint32, error) {
	if len(data) < characterHealthSize {
		return 0, 0, fmt.Errorf("got %d bytes, want %d", len(data), characterHealthSize)
	}

	return decodeXORValue(data[:xorValueSize]), decodeXORValue(data[xorValueSize:characterHealthSize]), nil
}

func decodeXORValue(data []byte) uint32 {
	return binary.LittleEndian.Uint32(data[4:8]) ^ binary.LittleEndian.Uint32(data[8:12])
}

func decodeRidingSteps(data []byte) (uint32, error) {
	if len(data) < ridingStepsSize {
		return 0, fmt.Errorf("decode riding steps: got %d bytes, want %d", len(data), ridingStepsSize)
	}

	return binary.LittleEndian.Uint32(data[:ridingStepsSize]), nil
}

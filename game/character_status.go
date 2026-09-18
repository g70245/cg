package game

import (
	"encoding/binary"
	"fmt"

	"cg/internal"

	"github.com/g70245/win"
)

const (
	xorValueSize        = 16
	characterStatusSize = xorValueSize * 4
	ridingStepsSize     = 4
)

type CharacterStatus struct {
	CurrentHP uint32
	MaximumHP uint32
	CurrentMP uint32
	MaximumMP uint32
}

func ReadCharacterStatus(hWnd win.HWND) (CharacterStatus, error) {
	data, err := internal.ReadMemoryAtAddress(hWnd, MEMORY_CHARACTER_STATUS, characterStatusSize)
	if err != nil {
		return CharacterStatus{}, fmt.Errorf("read character status: %w", err)
	}

	status, err := decodeCharacterStatus(data)
	if err != nil {
		return CharacterStatus{}, fmt.Errorf("read character status: %w", err)
	}
	return status, nil
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

func decodeCharacterStatus(data []byte) (CharacterStatus, error) {
	if len(data) < characterStatusSize {
		return CharacterStatus{}, fmt.Errorf("got %d bytes, want %d", len(data), characterStatusSize)
	}
	return CharacterStatus{
		CurrentHP: decodeXORValue(data[0*xorValueSize : 1*xorValueSize]),
		MaximumHP: decodeXORValue(data[1*xorValueSize : 2*xorValueSize]),
		CurrentMP: decodeXORValue(data[2*xorValueSize : 3*xorValueSize]),
		MaximumMP: decodeXORValue(data[3*xorValueSize : 4*xorValueSize]),
	}, nil
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

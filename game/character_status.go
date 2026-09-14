package game

import (
	"encoding/binary"
	"fmt"

	"cg/internal"

	"github.com/g70245/win"
)

const (
	characterHealthOffset = 0xB4C308
	xorValueSize          = 16
	characterHealthSize   = xorValueSize * 2
)

func ReadCharacterHealth(hWnd win.HWND) (uint32, uint32, error) {
	data, err := internal.ReadMemoryAtModuleOffset(hWnd, characterHealthOffset, characterHealthSize)
	if err != nil {
		return 0, 0, fmt.Errorf("read character health: %w", err)
	}

	current, maximum, err := decodeCharacterHealth(data)
	if err != nil {
		return 0, 0, fmt.Errorf("read character health: %w", err)
	}
	return current, maximum, nil
}

func decodeCharacterHealth(data []byte) (uint32, uint32, error) {
	if len(data) < characterHealthSize {
		return 0, 0, fmt.Errorf("decode character health: got %d bytes, want %d", len(data), characterHealthSize)
	}

	return decodeXORValue(data[:xorValueSize]), decodeXORValue(data[xorValueSize:characterHealthSize]), nil
}

func decodeXORValue(data []byte) uint32 {
	return binary.LittleEndian.Uint32(data[4:8]) ^ binary.LittleEndian.Uint32(data[8:12])
}

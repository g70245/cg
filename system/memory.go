package system

import (
	. "github.com/g70245/win"
)

type Pos struct {
	x, y int
}

func ReadMemory(hWnd HWND, lpBaseAddress uint32, size uint) []byte {
	lpdwProcessId := new(uint32)
	GetWindowThreadProcessId(hWnd, lpdwProcessId)
	readMemoryHandle, _ := OpenProcess(0x1F0FFF, false, uint32(*lpdwProcessId))
	return ReadProcessMemory(readMemoryHandle, lpBaseAddress, size)
}

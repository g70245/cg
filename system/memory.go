package system

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"

	. "github.com/g70245/win"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/transform"
)

func ReadMemory(hWnd HWND, lpBaseAddress uint32, size uint) []byte {
	lpdwProcessId := new(uint32)
	GetWindowThreadProcessId(hWnd, lpdwProcessId)
	readMemoryHandle, _ := OpenProcess(0x1F0FFF, false, uint32(*lpdwProcessId))
	return ReadProcessMemory(readMemoryHandle, lpBaseAddress, size)
}

func ReadMemoryString(hWnd HWND, lpBaseAddress uint32, size uint) string {
	data := ReadMemory(hWnd, lpBaseAddress, size)
	for i, v := range data {
		if v == 0x00 {
			data = data[:i]
			break
		}
	}
	transformReader := transform.NewReader(bytes.NewReader(data), traditionalchinese.Big5.NewDecoder())
	decBytes, _ := io.ReadAll(transformReader)

	return string(decBytes)
}

func ReadMemoryFloat32(hWnd HWND, lpBaseAddress uint32, size uint) float32 {
	data := ReadMemory(hWnd, lpBaseAddress, size)
	return math.Float32frombits(binary.LittleEndian.Uint32(data))
}

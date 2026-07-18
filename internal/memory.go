package internal

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"

	"github.com/g70245/win"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/transform"
)

const processAllAccess = 0x1F0FFF

type memoryOperations struct {
	getWindowThreadProcessID func(win.HWND, *uint32) uint32
	openProcess              func(uint32, bool, uint32) (win.HWND, error)
	readProcessMemory        func(win.HWND, uint32, uint) []byte
	closeHandle              func(win.HANDLE) bool
}

func newMemoryOperations() memoryOperations {
	return memoryOperations{
		getWindowThreadProcessID: win.GetWindowThreadProcessId,
		openProcess:              win.OpenProcess,
		readProcessMemory:        win.ReadProcessMemory,
		closeHandle:              win.CloseHandle,
	}
}

func ReadMemoryString(hWnd win.HWND, lpBaseAddress uint32, size uint) string {
	data := readMemory(hWnd, lpBaseAddress, size)
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

func ReadMemoryFloat32(hWnd win.HWND, lpBaseAddress uint32, size uint) float32 {
	data := readMemory(hWnd, lpBaseAddress, size)
	return math.Float32frombits(binary.LittleEndian.Uint32(data))
}

func ReadMemoryUint32(hWnd win.HWND, lpBaseAddress uint32) uint32 {
	data := readMemory(hWnd, lpBaseAddress, 4)
	return binary.LittleEndian.Uint32(data)
}

func readMemory(hWnd win.HWND, lpBaseAddress uint32, size uint) []byte {
	return readMemoryWith(newMemoryOperations(), hWnd, lpBaseAddress, size)
}

func readMemoryWith(operations memoryOperations, hWnd win.HWND, lpBaseAddress uint32, size uint) []byte {
	processID := new(uint32)
	operations.getWindowThreadProcessID(hWnd, processID)

	readMemoryHandle, _ := operations.openProcess(processAllAccess, false, *processID)
	if readMemoryHandle == 0 {
		return make([]byte, size)
	}
	defer operations.closeHandle(win.HANDLE(readMemoryHandle))

	return operations.readProcessMemory(readMemoryHandle, lpBaseAddress, size)
}

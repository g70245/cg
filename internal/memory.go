package internal

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"

	"github.com/g70245/win"
	"golang.org/x/sys/windows"
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

type moduleMemoryOperations struct {
	getWindowThreadProcessID func(win.HWND, *uint32) uint32
	createSnapshot           func(uint32, uint32) (windows.Handle, error)
	moduleFirst              func(windows.Handle, *windows.ModuleEntry32) error
	openProcess              func(uint32, bool, uint32) (windows.Handle, error)
	readProcessMemory        func(windows.Handle, uintptr, *byte, uintptr, *uintptr) error
	closeHandle              func(windows.Handle) error
}

func newMemoryOperations() memoryOperations {
	return memoryOperations{
		getWindowThreadProcessID: win.GetWindowThreadProcessId,
		openProcess:              win.OpenProcess,
		readProcessMemory:        win.ReadProcessMemory,
		closeHandle:              win.CloseHandle,
	}
}

func newModuleMemoryOperations() moduleMemoryOperations {
	return moduleMemoryOperations{
		getWindowThreadProcessID: win.GetWindowThreadProcessId,
		createSnapshot:           windows.CreateToolhelp32Snapshot,
		moduleFirst:              windows.Module32First,
		openProcess:              windows.OpenProcess,
		readProcessMemory:        windows.ReadProcessMemory,
		closeHandle:              windows.CloseHandle,
	}
}

func ReadMemoryAtModuleOffset(hWnd win.HWND, offset uint32, size uint) ([]byte, error) {
	return readMemoryAtModuleOffsetWith(newModuleMemoryOperations(), hWnd, offset, size)
}

func readMemoryAtModuleOffsetWith(operations moduleMemoryOperations, hWnd win.HWND, offset uint32, size uint) (data []byte, err error) {
	if size == 0 {
		return nil, fmt.Errorf("read module memory: size must be greater than zero")
	}

	var processID uint32
	operations.getWindowThreadProcessID(hWnd, &processID)
	if processID == 0 {
		return nil, fmt.Errorf("read module memory: resolve process ID")
	}

	snapshot, err := operations.createSnapshot(windows.TH32CS_SNAPMODULE|windows.TH32CS_SNAPMODULE32, processID)
	if err != nil {
		return nil, fmt.Errorf("read module memory: create module snapshot: %w", err)
	}
	defer func() {
		if closeErr := operations.closeHandle(snapshot); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("read module memory: close module snapshot: %w", closeErr))
		}
	}()

	module := windows.ModuleEntry32{Size: uint32(windows.SizeofModuleEntry32)}
	if err := operations.moduleFirst(snapshot, &module); err != nil {
		return nil, fmt.Errorf("read module memory: resolve main module: %w", err)
	}

	address := module.ModBaseAddr + uintptr(offset)
	if address < module.ModBaseAddr {
		return nil, fmt.Errorf("read module memory: address overflow")
	}

	process, err := operations.openProcess(windows.PROCESS_QUERY_INFORMATION|windows.PROCESS_VM_READ, false, processID)
	if err != nil {
		return nil, fmt.Errorf("read module memory: open process: %w", err)
	}
	defer func() {
		if closeErr := operations.closeHandle(process); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("read module memory: close process: %w", closeErr))
		}
	}()

	data = make([]byte, size)
	var bytesRead uintptr
	if err := operations.readProcessMemory(process, address, &data[0], uintptr(size), &bytesRead); err != nil {
		return nil, fmt.Errorf("read module memory: read process memory: %w", err)
	}
	if bytesRead != uintptr(size) {
		return nil, fmt.Errorf("read module memory: read %d bytes, want %d", bytesRead, size)
	}

	return data, nil
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

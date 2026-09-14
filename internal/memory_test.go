package internal

import (
	"errors"
	"reflect"
	"testing"
	"unsafe"

	"github.com/g70245/win"
	"golang.org/x/sys/windows"
)

func TestReadMemoryWithManagesProcessHandle(t *testing.T) {
	const (
		windowHandle  = win.HWND(11)
		processHandle = win.HWND(22)
		processID     = uint32(33)
		baseAddress   = uint32(44)
		readSize      = uint(4)
	)

	tests := []struct {
		name       string
		openHandle win.HWND
		openError  error
		wantData   []byte
		wantEvents []string
	}{
		{
			name:       "closes handle after reading memory",
			openHandle: processHandle,
			wantData:   []byte{1, 2, 3, 4},
			wantEvents: []string{"get process ID", "open process", "read memory", "close handle"},
		},
		{
			name:       "does not read or close when opening fails",
			openError:  errors.New("open process"),
			wantData:   []byte{0, 0, 0, 0},
			wantEvents: []string{"get process ID", "open process"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var events []string
			operations := memoryOperations{
				getWindowThreadProcessID: func(hWnd win.HWND, targetProcessID *uint32) uint32 {
					events = append(events, "get process ID")
					if hWnd != windowHandle {
						t.Fatalf("GetWindowThreadProcessId() handle = %d, want %d", hWnd, windowHandle)
					}
					*targetProcessID = processID
					return 1
				},
				openProcess: func(access uint32, inheritHandle bool, actualProcessID uint32) (win.HWND, error) {
					events = append(events, "open process")
					if access != processAllAccess {
						t.Errorf("OpenProcess() access = %#x, want %#x", access, processAllAccess)
					}
					if inheritHandle {
						t.Error("OpenProcess() inherited handle, want false")
					}
					if actualProcessID != processID {
						t.Errorf("OpenProcess() process ID = %d, want %d", actualProcessID, processID)
					}
					return tt.openHandle, tt.openError
				},
				readProcessMemory: func(handle win.HWND, address uint32, size uint) []byte {
					events = append(events, "read memory")
					if handle != processHandle {
						t.Errorf("ReadProcessMemory() handle = %d, want %d", handle, processHandle)
					}
					if address != baseAddress {
						t.Errorf("ReadProcessMemory() address = %d, want %d", address, baseAddress)
					}
					if size != readSize {
						t.Errorf("ReadProcessMemory() size = %d, want %d", size, readSize)
					}
					return tt.wantData
				},
				closeHandle: func(handle win.HANDLE) bool {
					events = append(events, "close handle")
					if handle != win.HANDLE(processHandle) {
						t.Errorf("CloseHandle() handle = %d, want %d", handle, processHandle)
					}
					return true
				},
			}

			got := readMemoryWith(operations, windowHandle, baseAddress, readSize)

			if !reflect.DeepEqual(got, tt.wantData) {
				t.Errorf("readMemoryWith() = %v, want %v", got, tt.wantData)
			}
			if !reflect.DeepEqual(events, tt.wantEvents) {
				t.Errorf("operation order = %v, want %v", events, tt.wantEvents)
			}
		})
	}
}

func TestReadMemoryAtModuleOffsetWith(t *testing.T) {
	const (
		windowHandle  = win.HWND(11)
		processID     = uint32(33)
		snapshot      = windows.Handle(44)
		processHandle = windows.Handle(55)
		moduleBase    = uintptr(0x400000)
		offset        = uint32(0x1234)
		readSize      = uint(4)
	)

	var events []string
	operations := moduleMemoryOperations{
		getWindowThreadProcessID: func(hWnd win.HWND, targetProcessID *uint32) uint32 {
			events = append(events, "get process ID")
			if hWnd != windowHandle {
				t.Fatalf("GetWindowThreadProcessId() handle = %d, want %d", hWnd, windowHandle)
			}
			*targetProcessID = processID
			return 1
		},
		createSnapshot: func(flags uint32, actualProcessID uint32) (windows.Handle, error) {
			events = append(events, "create snapshot")
			wantFlags := uint32(windows.TH32CS_SNAPMODULE | windows.TH32CS_SNAPMODULE32)
			if flags != wantFlags {
				t.Errorf("CreateToolhelp32Snapshot() flags = %#x, want %#x", flags, wantFlags)
			}
			if actualProcessID != processID {
				t.Errorf("CreateToolhelp32Snapshot() process ID = %d, want %d", actualProcessID, processID)
			}
			return snapshot, nil
		},
		moduleFirst: func(actualSnapshot windows.Handle, module *windows.ModuleEntry32) error {
			events = append(events, "get main module")
			if actualSnapshot != snapshot {
				t.Errorf("Module32First() snapshot = %d, want %d", actualSnapshot, snapshot)
			}
			if module.Size != uint32(windows.SizeofModuleEntry32) {
				t.Errorf("Module32First() size = %d, want %d", module.Size, windows.SizeofModuleEntry32)
			}
			module.ModBaseAddr = moduleBase
			return nil
		},
		openProcess: func(access uint32, inheritHandle bool, actualProcessID uint32) (windows.Handle, error) {
			events = append(events, "open process")
			wantAccess := uint32(windows.PROCESS_QUERY_INFORMATION | windows.PROCESS_VM_READ)
			if access != wantAccess {
				t.Errorf("OpenProcess() access = %#x, want %#x", access, wantAccess)
			}
			if inheritHandle {
				t.Error("OpenProcess() inherited handle, want false")
			}
			if actualProcessID != processID {
				t.Errorf("OpenProcess() process ID = %d, want %d", actualProcessID, processID)
			}
			return processHandle, nil
		},
		readProcessMemory: func(handle windows.Handle, address uintptr, data *byte, size uintptr, bytesRead *uintptr) error {
			events = append(events, "read memory")
			if handle != processHandle {
				t.Errorf("ReadProcessMemory() handle = %d, want %d", handle, processHandle)
			}
			if address != moduleBase+uintptr(offset) {
				t.Errorf("ReadProcessMemory() address = %#x, want %#x", address, moduleBase+uintptr(offset))
			}
			if size != uintptr(readSize) {
				t.Errorf("ReadProcessMemory() size = %d, want %d", size, readSize)
			}
			buffer := unsafe.Slice(data, size)
			copy(buffer, []byte{1, 2, 3, 4})
			*bytesRead = size
			return nil
		},
		closeHandle: func(handle windows.Handle) error {
			switch handle {
			case processHandle:
				events = append(events, "close process")
			case snapshot:
				events = append(events, "close snapshot")
			default:
				t.Errorf("CloseHandle() handle = %d, want process or snapshot", handle)
			}
			return nil
		},
	}

	got, err := readMemoryAtModuleOffsetWith(operations, windowHandle, offset, readSize)
	if err != nil {
		t.Fatalf("readMemoryAtModuleOffsetWith() error = %v", err)
	}
	if want := []byte{1, 2, 3, 4}; !reflect.DeepEqual(got, want) {
		t.Errorf("readMemoryAtModuleOffsetWith() = %v, want %v", got, want)
	}
	wantEvents := []string{"get process ID", "create snapshot", "get main module", "open process", "read memory", "close process", "close snapshot"}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Errorf("operation order = %v, want %v", events, wantEvents)
	}
}

func TestReadMemoryAtModuleOffsetWithRejectsShortRead(t *testing.T) {
	operations := moduleMemoryOperations{
		getWindowThreadProcessID: func(win.HWND, *uint32) uint32 { return 1 },
		createSnapshot:           func(uint32, uint32) (windows.Handle, error) { return 1, nil },
		moduleFirst: func(_ windows.Handle, module *windows.ModuleEntry32) error {
			module.ModBaseAddr = 0x400000
			return nil
		},
		openProcess: func(uint32, bool, uint32) (windows.Handle, error) { return 2, nil },
		readProcessMemory: func(windows.Handle, uintptr, *byte, uintptr, *uintptr) error {
			return nil
		},
		closeHandle: func(windows.Handle) error { return nil },
	}

	var processID uint32
	operations.getWindowThreadProcessID = func(_ win.HWND, targetProcessID *uint32) uint32 {
		*targetProcessID = 33
		processID = *targetProcessID
		return 1
	}

	if _, err := readMemoryAtModuleOffsetWith(operations, 11, 0x1234, 4); err == nil {
		t.Fatal("readMemoryAtModuleOffsetWith() error = nil, want a short-read error")
	}
	if processID == 0 {
		t.Fatal("process ID was not resolved")
	}
}

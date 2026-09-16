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

func TestReadMemoryAtAddressWith(t *testing.T) {
	const (
		windowHandle    = win.HWND(11)
		processID       = uint32(33)
		processHandle   = windows.Handle(55)
		expectedAddress = uint32(0xF4C464)
		readSize        = uint(4)
	)

	var events []string
	operations := addressMemoryOperations{
		getWindowThreadProcessID: func(hWnd win.HWND, targetProcessID *uint32) uint32 {
			events = append(events, "get process ID")
			if hWnd != windowHandle {
				t.Fatalf("GetWindowThreadProcessId() handle = %d, want %d", hWnd, windowHandle)
			}
			*targetProcessID = processID
			return 1
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
		readProcessMemory: func(handle windows.Handle, actualAddress uintptr, data *byte, size uintptr, bytesRead *uintptr) error {
			events = append(events, "read memory")
			if handle != processHandle {
				t.Errorf("ReadProcessMemory() handle = %d, want %d", handle, processHandle)
			}
			if actualAddress != uintptr(expectedAddress) {
				t.Errorf("ReadProcessMemory() address = %#x, want %#x", actualAddress, uintptr(expectedAddress))
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
			if handle != processHandle {
				t.Errorf("CloseHandle() handle = %d, want %d", handle, processHandle)
			}
			events = append(events, "close process")
			return nil
		},
	}

	got, err := readMemoryAtAddressWith(operations, windowHandle, expectedAddress, readSize)
	if err != nil {
		t.Fatalf("readMemoryAtAddressWith() error = %v", err)
	}
	if want := []byte{1, 2, 3, 4}; !reflect.DeepEqual(got, want) {
		t.Errorf("readMemoryAtAddressWith() = %v, want %v", got, want)
	}
	wantEvents := []string{"get process ID", "open process", "read memory", "close process"}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Errorf("operation order = %v, want %v", events, wantEvents)
	}
}

func TestReadMemoryAtAddressWithRejectsShortRead(t *testing.T) {
	operations := addressMemoryOperations{
		getWindowThreadProcessID: func(win.HWND, *uint32) uint32 { return 1 },
		openProcess:              func(uint32, bool, uint32) (windows.Handle, error) { return 2, nil },
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

	if _, err := readMemoryAtAddressWith(operations, 11, 0x401234, 4); err == nil {
		t.Fatal("readMemoryAtAddressWith() error = nil, want a short-read error")
	}
	if processID == 0 {
		t.Fatal("process ID was not resolved")
	}
}

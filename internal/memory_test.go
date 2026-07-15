package internal

import (
	"errors"
	"reflect"
	"testing"

	"github.com/g70245/win"
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

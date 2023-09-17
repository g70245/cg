package game

import (
	. "cg/system"
	"time"

	. "github.com/lxn/win"
)

const (
	ACTION_INTERVAL = 120
)

func closeAllWindow(hWnd HWND) {
	KeyCombinationMsg(hWnd, VK_SHIFT, VK_F12)
	time.Sleep(ACTION_INTERVAL * time.Millisecond)
}

func openHumanWindow(hWnd HWND, key uintptr) {
	KeyCombinationMsg(hWnd, VK_SHIFT, VK_F12)
	time.Sleep(ACTION_INTERVAL * time.Millisecond)
	KeyCombinationMsg(hWnd, VK_CONTROL, key)
	time.Sleep(ACTION_INTERVAL * time.Millisecond)
	resetAllWindowPos(hWnd)
}

func resetAllWindowPos(hWnd HWND) {
	KeyCombinationMsg(hWnd, VK_CONTROL, VK_F12)
	time.Sleep(ACTION_INTERVAL * time.Millisecond)
}

func useHumanSkill(hWnd HWND, x, y int32, id, level int) {
	LeftClick(hWnd, x, y+int32((id-1)*13))
	time.Sleep(ACTION_INTERVAL * time.Millisecond)
	LeftClick(hWnd, x, y+int32(level*13))
	time.Sleep(ACTION_INTERVAL * time.Millisecond)
}

func usePetSkill(hWnd HWND, x, y int32, id int) {
	LeftClick(hWnd, x, y+int32((id-1)*13))
	time.Sleep(ACTION_INTERVAL * time.Millisecond)
}

func getLastLineOfLog(logDir string) string {
	path, _, _ := FindLastModifiedFileBefore(logDir, time.Now().Add(10*time.Second))
	return GetLastLineWithSeek(path)
}

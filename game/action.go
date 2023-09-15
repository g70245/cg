package game

import (
	. "cg/system"
	"time"

	. "github.com/lxn/win"
)

func closeAll(hWnd HWND) {
	KeyCombinationMsg(hWnd, VK_SHIFT, VK_F12)
}

func openWindow(hWnd HWND, key uintptr) {
	KeyCombinationMsg(hWnd, VK_SHIFT, VK_F12)
	time.Sleep(100 * time.Millisecond)
	KeyCombinationMsg(hWnd, VK_CONTROL, key)
	time.Sleep(100 * time.Millisecond)
	KeyCombinationMsg(hWnd, VK_CONTROL, VK_F12)
}

func useClickSkillAtIndex() {

}

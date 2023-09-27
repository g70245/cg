package game

import (
	"log"
	"time"

	. "github.com/g70245/win"
)

const (
	TELEPORT_CHECKER_INTERVAL = 300
)

func activateTeleportChecker(hWnd HWND, teleportCheckerStopChan chan bool, isTeleportedChan chan bool) {

	teleportCheckerTicker := time.NewTicker(TELEPORT_CHECKER_INTERVAL * time.Millisecond)

	go func(hWnd HWND) {
		defer teleportCheckerTicker.Stop()

		log.Println("Teleport Checker enabled")
		currentMapName := getMapName(hWnd)
		log.Printf("Handle %d Current Location: %s\n", hWnd, currentMapName)

		for {

			select {
			case <-teleportCheckerStopChan:
				log.Println("Teleport Checker disabled")
				return
			case <-teleportCheckerTicker.C:
				if newMapName := getMapName(hWnd); currentMapName != newMapName {
					isTeleportedChan <- true
					return
				}
			default:
				time.Sleep(TELEPORT_CHECKER_INTERVAL * time.Microsecond / 3)
			}
		}
	}(hWnd)
}

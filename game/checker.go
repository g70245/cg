package game

import (
	"log"
	"time"

	. "github.com/g70245/win"
)

const (
	TELEPORT_CHECKER_INTERVAL = 300
)

func activateCheckers(hWnd HWND, checkerStopChan chan bool, isTeleportedChan chan bool, logDir *string) {

	teleportCheckerTicker := time.NewTicker(TELEPORT_CHECKER_INTERVAL * time.Millisecond)
	logCheckerTicker := time.NewTicker(TELEPORT_CHECKER_INTERVAL * time.Millisecond)

	go func(hWnd HWND) {
		defer teleportCheckerTicker.Stop()
		defer logCheckerTicker.Stop()

		log.Println("Checkers enabled")
		currentMapName := getMapName(hWnd)
		log.Printf("Handle %d Current Location: %s\n", hWnd, currentMapName)

		for {

			select {
			case <-checkerStopChan:
				log.Println("Checkers disabled")
				return
			case <-teleportCheckerTicker.C:
				if newMapName := getMapName(hWnd); currentMapName != newMapName {
					isTeleportedChan <- true
					return
				}
			case <-logCheckerTicker.C:
				if *logDir != "" {

				}
			default:
				time.Sleep(TELEPORT_CHECKER_INTERVAL * time.Microsecond / 3)
			}
		}
	}(hWnd)
}

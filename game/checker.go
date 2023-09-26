package game

import (
	"log"
	"time"
)

func activateTeleportChecker(logDir *string, teleportCheckerStopChan chan bool, isTeleportedChan chan bool) {

	logCheckerTicker := time.NewTicker(LOG_CHECKER_INTERVAL * time.Millisecond)

	go func() {
		log.Println("Log Checker enabled")
		defer logCheckerTicker.Stop()

		for {
			if *logDir == "" {
				continue
			}

			select {
			case <-teleportCheckerStopChan:
				log.Println("Log Checker disabled")
				return
			case <-logCheckerTicker.C:
				if isTeleportedToOtherMap(*logDir) {
					isTeleportedChan <- true
					return
				}
			default:
				time.Sleep(LOG_CHECKER_INTERVAL * time.Microsecond / 3)
			}
		}
	}()
}

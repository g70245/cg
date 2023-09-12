package main

import (
	"fmt"
	"time"
)

const WORKER_DURATION_MILLIS = 600

var GLOBAL_PARTY_LEAD_HWND string

func Worker(window Window, stopChan chan bool) {
	ticker := time.NewTicker(WORKER_DURATION_MILLIS * time.Millisecond)
	m := MovementState{hWnd: window.hWnd}

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				switch GetScene(window.hWnd) {
				case BATTLE_SCENE:
					fmt.Printf("PID %s is at BATTLE_SCENE\n", window.Get())
				case NORMAL_SCENE:
					if GLOBAL_PARTY_LEAD_HWND == "" || GLOBAL_PARTY_LEAD_HWND == window.Get() {
						m.Move()
						fmt.Printf("PID %s is at NORMAL_SCENE\n", window.Get())
					}
				}
			case <-stopChan:
				return
			}
		}
	}()
}

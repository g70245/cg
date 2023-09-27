package system

import (
	"log"
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

var alertPausedChan, alertDoneChan chan bool

func CreateBeeper(path string) {
	go func() {
		alertPausedChan = make(chan bool)
		alertDoneChan = make(chan bool)

		f, err := os.Open(path)
		if err != nil {
			log.Fatal(err)
		}

		streamer, format, err := mp3.Decode(f)
		if err != nil {
			log.Fatal(err)
		}
		defer streamer.Close()

		speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
		ctrl := &beep.Ctrl{Streamer: beep.Loop(-1, streamer), Paused: true}
		speaker.Play(ctrl)

		for {
			select {
			case paused := <-alertPausedChan:
				speaker.Lock()
				ctrl.Paused = paused
				speaker.Unlock()
			case <-alertDoneChan:
				speaker.Lock()
				ctrl.Paused = true
				speaker.Unlock()

				defer speaker.Clear()
				defer speaker.Close()
				return
			}
		}
	}()
}

func PlayBeeper() bool {
	if alertPausedChan != nil {
		alertPausedChan <- false
		return true
	}
	return false
}

func StopBeeper() {
	if alertPausedChan != nil {
		alertPausedChan <- true
	}
}

func CloseBeeper() {
	if alertDoneChan != nil {
		alertDoneChan <- true
		close(alertPausedChan)
		close(alertDoneChan)

		alertPausedChan = nil
		alertDoneChan = nil
	}
}

func IsBeeperReady() bool {
	return alertPausedChan != nil
}

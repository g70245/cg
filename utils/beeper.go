package utils

import (
	"log"
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

type beeper struct {
	isReady              bool
	pausedChan, doneChan chan bool
}

var Beeper beeper

func init() {
	if Beeper == (beeper{}) {
		Beeper = beeper{pausedChan: make(chan bool), doneChan: make(chan bool)}
	}
}

func (b *beeper) Init(path string) {

	go func() {
		b.Close()

		f, err := os.Open(path)
		if err != nil {
			log.Fatal(err)
		}

		streamer, format, err := mp3.Decode(f)
		if err != nil {
			log.Fatal(err)
		}

		speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
		ctrl := &beep.Ctrl{Streamer: beep.Loop(-1, streamer), Paused: true}
		speaker.Play(ctrl)

		b.isReady = true

		defer streamer.Close()
		defer speaker.Clear()

		for {
			select {
			case paused := <-b.pausedChan:
				speaker.Lock()
				ctrl.Paused = paused
				speaker.Unlock()
			case <-b.doneChan:
				speaker.Lock()
				ctrl.Paused = true
				speaker.Unlock()
				return
			}
		}
	}()
}

func (b *beeper) Play() {
	b.pausedChan <- false
}

func (b *beeper) Stop() {
	if b.isReady {
		b.pausedChan <- true
	}
}

func (b *beeper) Close() {
	if b.isReady {
		b.doneChan <- true
		b.isReady = false
	}
}

func (b *beeper) IsReady() bool {
	return b.isReady
}

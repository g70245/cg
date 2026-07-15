package utils

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

type audioSession interface {
	Play()
	Stop()
	Close() error
}

type audioSessionFactory func(string) (audioSession, error)

type beeper struct {
	lifecycleMu sync.Mutex
	mu          sync.RWMutex
	session     audioSession
	newSession  audioSessionFactory
}

var Beeper = newBeeper(newSpeakerSession)

func newBeeper(factory audioSessionFactory) *beeper {
	return &beeper{newSession: factory}
}

func (b *beeper) Init(path string) error {
	b.lifecycleMu.Lock()
	defer b.lifecycleMu.Unlock()

	if err := b.closeCurrentSession(); err != nil {
		return fmt.Errorf("close current alert music: %w", err)
	}

	session, err := b.newSession(path)
	if err != nil {
		return fmt.Errorf("initialize alert music: %w", err)
	}

	b.mu.Lock()
	b.session = session
	b.mu.Unlock()
	return nil
}

func (b *beeper) Play() {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.session != nil {
		b.session.Play()
	}
}

func (b *beeper) Stop() {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.session != nil {
		b.session.Stop()
	}
}

func (b *beeper) Close() error {
	b.lifecycleMu.Lock()
	defer b.lifecycleMu.Unlock()
	return b.closeCurrentSession()
}

func (b *beeper) IsReady() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.session != nil
}

func (b *beeper) closeCurrentSession() error {
	b.mu.Lock()
	session := b.session
	b.session = nil
	b.mu.Unlock()

	if session == nil {
		return nil
	}
	return session.Close()
}

type speakerSession struct {
	streamer beep.StreamSeekCloser
	control  *beep.Ctrl
}

func newSpeakerSession(path string) (audioSession, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open MP3 file %q: %w", path, err)
	}

	streamer, format, err := mp3.Decode(file)
	if err != nil {
		if closeErr := file.Close(); closeErr != nil {
			return nil, fmt.Errorf("decode MP3 file %q: %w (close file: %v)", path, err, closeErr)
		}
		return nil, fmt.Errorf("decode MP3 file %q: %w", path, err)
	}

	if err := speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10)); err != nil {
		if closeErr := streamer.Close(); closeErr != nil {
			return nil, fmt.Errorf("initialize speaker: %w (close MP3 stream: %v)", err, closeErr)
		}
		return nil, fmt.Errorf("initialize speaker: %w", err)
	}

	control := &beep.Ctrl{Streamer: beep.Loop(-1, streamer), Paused: true}
	speaker.Play(control)
	return &speakerSession{streamer: streamer, control: control}, nil
}

func (s *speakerSession) Play() {
	speaker.Lock()
	s.control.Paused = false
	speaker.Unlock()
}

func (s *speakerSession) Stop() {
	speaker.Lock()
	s.control.Paused = true
	speaker.Unlock()
}

func (s *speakerSession) Close() error {
	s.Stop()
	speaker.Clear()
	return s.streamer.Close()
}

package utils

import (
	"errors"
	"sync"
	"testing"
)

type fakeAudioSession struct {
	mu         sync.Mutex
	playCount  int
	stopCount  int
	closeCount int
	closeError error
}

func (s *fakeAudioSession) Play() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.playCount++
}

func (s *fakeAudioSession) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopCount++
}

func (s *fakeAudioSession) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closeCount++
	return s.closeError
}

func (s *fakeAudioSession) counts() (play, stop, close int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.playCount, s.stopCount, s.closeCount
}

func TestBeeperLifecycle(t *testing.T) {
	session := &fakeAudioSession{}
	b := newBeeper(func(path string) (audioSession, error) {
		if path != "alert.mp3" {
			t.Fatalf("factory path = %q, want alert.mp3", path)
		}
		return session, nil
	})

	b.Play()
	b.Stop()
	if b.IsReady() {
		t.Fatal("new beeper is ready, want unconfigured")
	}

	if err := b.Init("alert.mp3"); err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	if !b.IsReady() {
		t.Fatal("initialized beeper is not ready")
	}

	b.Play()
	b.Stop()
	if err := b.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if err := b.Close(); err != nil {
		t.Fatalf("second Close() error = %v", err)
	}
	if b.IsReady() {
		t.Fatal("closed beeper is ready")
	}

	play, stop, close := session.counts()
	if play != 1 || stop != 1 || close != 1 {
		t.Fatalf("session counts = play %d, stop %d, close %d; want 1, 1, 1", play, stop, close)
	}
}

func TestBeeperInitFailure(t *testing.T) {
	wantErr := errors.New("invalid MP3")
	b := newBeeper(func(string) (audioSession, error) {
		return nil, wantErr
	})

	if err := b.Init("invalid.mp3"); !errors.Is(err, wantErr) {
		t.Fatalf("Init() error = %v, want %v", err, wantErr)
	}
	if b.IsReady() {
		t.Fatal("beeper is ready after failed initialization")
	}
}

func TestBeeperReinitializationClosesCurrentSession(t *testing.T) {
	first := &fakeAudioSession{}
	second := &fakeAudioSession{}
	callCount := 0
	b := newBeeper(func(string) (audioSession, error) {
		callCount++
		if callCount == 1 {
			return first, nil
		}
		return second, nil
	})

	if err := b.Init("first.mp3"); err != nil {
		t.Fatal(err)
	}
	if err := b.Init("second.mp3"); err != nil {
		t.Fatal(err)
	}

	_, _, firstClose := first.counts()
	_, _, secondClose := second.counts()
	if firstClose != 1 || secondClose != 0 {
		t.Fatalf("close counts = first %d, second %d; want 1, 0", firstClose, secondClose)
	}
}

func TestBeeperCloseFailureLeavesBeeperUnconfigured(t *testing.T) {
	wantErr := errors.New("close session")
	session := &fakeAudioSession{closeError: wantErr}
	b := newBeeper(func(string) (audioSession, error) {
		return session, nil
	})

	if err := b.Init("alert.mp3"); err != nil {
		t.Fatal(err)
	}
	if err := b.Close(); !errors.Is(err, wantErr) {
		t.Fatalf("Close() error = %v, want %v", err, wantErr)
	}
	if b.IsReady() {
		t.Fatal("beeper is ready after close failure")
	}
}

func TestBeeperConcurrentControls(t *testing.T) {
	session := &fakeAudioSession{}
	b := newBeeper(func(string) (audioSession, error) {
		return session, nil
	})
	if err := b.Init("alert.mp3"); err != nil {
		t.Fatal(err)
	}

	var waitGroup sync.WaitGroup
	for i := 0; i < 25; i++ {
		waitGroup.Add(3)
		go func() {
			defer waitGroup.Done()
			b.Play()
		}()
		go func() {
			defer waitGroup.Done()
			b.Stop()
		}()
		go func() {
			defer waitGroup.Done()
			_ = b.IsReady()
		}()
	}
	waitGroup.Wait()
}

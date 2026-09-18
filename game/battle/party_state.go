package battle

import "sync/atomic"

type PartyState struct {
	enabled atomic.Bool
}

func (s *PartyState) SetEnabled(enabled bool) {
	s.enabled.Store(enabled)
}

func (s *PartyState) Enabled() bool {
	return s.enabled.Load()
}

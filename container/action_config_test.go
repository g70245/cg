package container

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"

	"cg/game/battle"
	"cg/game/enum/character"
)

type testReadCloser struct {
	reader     io.Reader
	closeErr   error
	closeCount int
}

func (r *testReadCloser) Read(data []byte) (int, error) {
	return r.reader.Read(data)
}

func (r *testReadCloser) Close() error {
	r.closeCount++
	return r.closeErr
}

type testWriteCloser struct {
	buffer     bytes.Buffer
	writeErr   error
	shortWrite bool
	closeErr   error
	closeCount int
}

func (w *testWriteCloser) Write(data []byte) (int, error) {
	if w.writeErr != nil {
		return 0, w.writeErr
	}
	if w.shortWrite {
		return len(data) - 1, nil
	}
	return w.buffer.Write(data)
}

func (w *testWriteCloser) Close() error {
	w.closeCount++
	return w.closeErr
}

func TestLoadActionConfiguration(t *testing.T) {
	want := battle.CreateNewBattleActionState(1)
	want.AddCharacterAction(character.Defend)
	data, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	reader := &testReadCloser{reader: bytes.NewReader(data)}

	got, err := loadActionConfiguration(reader)
	if err != nil {
		t.Fatalf("loadActionConfiguration() error = %v", err)
	}
	if len(got.CharacterActions) != len(want.CharacterActions) {
		t.Fatalf("character action count = %d, want %d", len(got.CharacterActions), len(want.CharacterActions))
	}
	if reader.closeCount != 1 {
		t.Fatalf("reader close count = %d, want 1", reader.closeCount)
	}
}

func TestLoadActionConfigurationErrors(t *testing.T) {
	readErr := errors.New("read failed")
	closeErr := errors.New("close failed")
	tests := []struct {
		name       string
		reader     *testReadCloser
		wantPrefix string
	}{
		{
			name:       "read",
			reader:     &testReadCloser{reader: errorReader{err: readErr}},
			wantPrefix: "read file: ",
		},
		{
			name:       "invalid format",
			reader:     &testReadCloser{reader: strings.NewReader("not JSON")},
			wantPrefix: "invalid file format: ",
		},
		{
			name:       "close",
			reader:     &testReadCloser{reader: strings.NewReader("{}"), closeErr: closeErr},
			wantPrefix: "close file: ",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := loadActionConfiguration(test.reader)
			if err == nil || !strings.HasPrefix(err.Error(), test.wantPrefix) {
				t.Fatalf("loadActionConfiguration() error = %v, want prefix %q", err, test.wantPrefix)
			}
			if test.reader.closeCount != 1 {
				t.Fatalf("reader close count = %d, want 1", test.reader.closeCount)
			}
		})
	}
}

func TestSaveActionConfiguration(t *testing.T) {
	want := battle.CreateNewBattleActionState(1)
	want.AddCharacterAction(character.Defend)
	writer := &testWriteCloser{}

	if err := saveActionConfiguration(writer, want); err != nil {
		t.Fatalf("saveActionConfiguration() error = %v", err)
	}
	if writer.closeCount != 1 {
		t.Fatalf("writer close count = %d, want 1", writer.closeCount)
	}

	var got battle.ActionState
	if err := json.Unmarshal(writer.buffer.Bytes(), &got); err != nil {
		t.Fatalf("saved action configuration is invalid: %v", err)
	}
	if len(got.CharacterActions) != len(want.CharacterActions) {
		t.Fatalf("character action count = %d, want %d", len(got.CharacterActions), len(want.CharacterActions))
	}
}

func TestSaveActionConfigurationErrors(t *testing.T) {
	writeErr := errors.New("write failed")
	closeErr := errors.New("close failed")
	tests := []struct {
		name       string
		writer     *testWriteCloser
		wantPrefix string
		wantErr    error
	}{
		{
			name:       "write",
			writer:     &testWriteCloser{writeErr: writeErr},
			wantPrefix: "write file: ",
			wantErr:    writeErr,
		},
		{
			name:       "short write",
			writer:     &testWriteCloser{shortWrite: true},
			wantPrefix: "write file: ",
			wantErr:    io.ErrShortWrite,
		},
		{
			name:       "close",
			writer:     &testWriteCloser{closeErr: closeErr},
			wantPrefix: "close file: ",
			wantErr:    closeErr,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := saveActionConfiguration(test.writer, battle.ActionState{})
			if err == nil || !strings.HasPrefix(err.Error(), test.wantPrefix) {
				t.Fatalf("saveActionConfiguration() error = %v, want prefix %q", err, test.wantPrefix)
			}
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("saveActionConfiguration() error = %v, want wrapped %v", err, test.wantErr)
			}
			if test.writer.closeCount != 1 {
				t.Fatalf("writer close count = %d, want 1", test.writer.closeCount)
			}
		})
	}
}

type errorReader struct {
	err error
}

func (r errorReader) Read([]byte) (int, error) {
	return 0, r.err
}

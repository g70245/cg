package internal

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestGetLastLines(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, dir string)
		lineCount int
		want      []string
		wantErr   bool
		noFiles   bool
	}{
		{
			name:      "missing directory",
			lineCount: 2,
			wantErr:   true,
		},
		{
			name:      "empty directory",
			setup:     func(t *testing.T, dir string) {},
			lineCount: 2,
			wantErr:   true,
			noFiles:   true,
		},
		{
			name: "empty file",
			setup: func(t *testing.T, dir string) {
				writeTestFile(t, dir, "game.log", nil)
			},
			lineCount: 2,
			want:      []string{},
		},
		{
			name: "returns requested lines with CRLF",
			setup: func(t *testing.T, dir string) {
				writeTestFile(t, dir, "game.log", []byte("first\r\nsecond\r\nthird\r\n"))
			},
			lineCount: 2,
			want:      []string{"second", "third"},
		},
		{
			name: "handles final line without newline",
			setup: func(t *testing.T, dir string) {
				writeTestFile(t, dir, "game.log", []byte("first\nsecond"))
			},
			lineCount: 1,
			want:      []string{"second"},
		},
		{
			name: "uses latest modified file",
			setup: func(t *testing.T, dir string) {
				oldPath := writeTestFile(t, dir, "old.log", []byte("old"))
				newPath := writeTestFile(t, dir, "new.log", []byte("new"))
				now := time.Now()
				if err := os.Chtimes(oldPath, now.Add(-time.Minute), now.Add(-time.Minute)); err != nil {
					t.Fatalf("Chtimes(%q): %v", oldPath, err)
				}
				if err := os.Chtimes(newPath, now, now); err != nil {
					t.Fatalf("Chtimes(%q): %v", newPath, err)
				}
			},
			lineCount: 1,
			want:      []string{"new"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			logDir := filepath.Join(root, "Log")
			if tt.setup != nil {
				if err := os.Mkdir(logDir, 0o755); err != nil {
					t.Fatalf("Mkdir(%q): %v", logDir, err)
				}
				tt.setup(t, logDir)
			}

			got, err := GetLastLines(logDir, tt.lineCount)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetLastLines() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.noFiles && !errors.Is(err, ErrNoLogFiles) {
				t.Fatalf("GetLastLines() error = %v, want ErrNoLogFiles", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetLastLines() = %v, want %v", got, tt.want)
			}
		})
	}
}

func writeTestFile(t *testing.T, dir, name string, data []byte) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("WriteFile(%q): %v", path, err)
	}
	return path
}

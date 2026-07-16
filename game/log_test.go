package game

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateLogDirectory(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T) string
		wantReason string
	}{
		{
			name:       "not selected",
			setup:      func(t *testing.T) string { return "" },
			wantReason: "game directory is not selected",
		},
		{
			name:       "missing Log directory",
			setup:      func(t *testing.T) string { return t.TempDir() },
			wantReason: "Log folder is missing or unreadable",
		},
		{
			name: "empty Log directory",
			setup: func(t *testing.T) string {
				root := t.TempDir()
				if err := os.Mkdir(filepath.Join(root, "Log"), 0o755); err != nil {
					t.Fatal(err)
				}
				return root
			},
			wantReason: "no log files were found",
		},
		{
			name: "readable log file",
			setup: func(t *testing.T) string {
				root := t.TempDir()
				logDir := filepath.Join(root, "Log")
				if err := os.Mkdir(logDir, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(logDir, "game.log"), nil, 0o644); err != nil {
					t.Fatal(err)
				}
				return root
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gameDir := tt.setup(t)
			err := ValidateLogDirectory(gameDir)
			if tt.wantReason == "" {
				if err != nil {
					t.Fatalf("ValidateLogDirectory() error = %v, want nil", err)
				}
				return
			}
			if err == nil || err.Error() != tt.wantReason {
				t.Fatalf("ValidateLogDirectory() error = %v, want %q", err, tt.wantReason)
			}
			if gameDir != "" && strings.Contains(err.Error(), gameDir) {
				t.Fatalf("ValidateLogDirectory() error contains game directory %q: %v", gameDir, err)
			}
		})
	}
}

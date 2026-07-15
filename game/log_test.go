package game

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateLogDirectory(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) string
		wantErr bool
	}{
		{
			name:    "not selected",
			setup:   func(t *testing.T) string { return "" },
			wantErr: true,
		},
		{
			name:    "missing Log directory",
			setup:   func(t *testing.T) string { return t.TempDir() },
			wantErr: true,
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
			wantErr: true,
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
			if err := ValidateLogDirectory(tt.setup(t)); (err != nil) != tt.wantErr {
				t.Fatalf("ValidateLogDirectory() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

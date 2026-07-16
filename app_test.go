package main

import (
	"path/filepath"
	"testing"
)

func TestGameDirForUserProfile(t *testing.T) {
	tests := []struct {
		name        string
		userProfile string
		want        string
	}{
		{
			name:        "user profile",
			userProfile: `C:\Users\test-user`,
			want:        filepath.Join(`C:\Users\test-user`, "Documents", "CG"),
		},
		{
			name: "missing user profile",
			want: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := gameDirForUserProfile(test.userProfile); got != test.want {
				t.Fatalf("gameDirForUserProfile(%q) = %q, want %q", test.userProfile, got, test.want)
			}
		})
	}
}

func TestDefaultGameDirUsesUserProfile(t *testing.T) {
	userProfile := `C:\Users\test-user`
	t.Setenv("USERPROFILE", userProfile)

	want := filepath.Join(userProfile, "Documents", "CG")
	if got := defaultGameDir(); got != want {
		t.Fatalf("defaultGameDir() = %q, want %q", got, want)
	}
}

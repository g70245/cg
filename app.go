package main

import (
	"cg/container"
	"os"
	"path/filepath"
)

func gameDirForUserProfile(userProfile string) string {
	if userProfile == "" {
		return ""
	}
	return filepath.Join(userProfile, "Documents", "CG")
}

func defaultGameDir() string {
	return gameDirForUserProfile(os.Getenv("USERPROFILE"))
}

func main() {
	container.App("CG", defaultGameDir(), 960, 320)
}

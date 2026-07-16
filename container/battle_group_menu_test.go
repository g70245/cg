package container

import (
	"cg/game"
	"cg/game/battle"
	"reflect"
	"testing"

	"github.com/g70245/win"
)

func TestManaCheckerOptionsFollowCurrentAliases(t *testing.T) {
	groupGames := game.Games{
		"123": win.HWND(123),
		"456": win.HWND(456),
	}
	allGames := game.Games{
		"1": win.HWND(123),
		"4": win.HWND(456),
	}

	wantOptions := []string{battle.NO_MANA_CHECKER, "1", "4"}
	if got := currentManaCheckerOptions(groupGames, allGames); !reflect.DeepEqual(got, wantOptions) {
		t.Fatalf("mana checker options = %v, want %v", got, wantOptions)
	}

	manaChecker := battle.NewManaChecker()
	manaChecker.Set("456")
	if got := currentManaCheckerAlias(manaChecker, allGames); got != "4" {
		t.Fatalf("mana checker alias = %q, want %q", got, "4")
	}
}

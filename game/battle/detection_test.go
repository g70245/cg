package battle

import (
	"fmt"
	"image"
	"image/color"
	"testing"

	"cg/game"

	"github.com/g70245/win"
)

func TestGetSkillWindowPosFromCapture(t *testing.T) {
	capture := image.NewRGBA(image.Rect(0, 0, game.GAME_WIDTH, game.GAME_HEIGHT))
	wantX := BATTLE_WINDOW_SKILL_FIRST.X + 1
	wantY := BATTLE_WINDOW_SKILL_FIRST.Y + 1
	setBattleCaptureColor(capture, wantX, wantY, BATTLE_WINDOW_SKILL_FIRST.Color)

	gotX, gotY, found := getSkillWindowPosFromCapture(capture)
	if gotX != wantX || gotY != wantY || !found {
		t.Fatalf("getSkillWindowPosFromCapture() = (%d, %d, %t), want (%d, %d, true)", gotX, gotY, found, wantX, wantY)
	}
}

func TestGetInventoryPosFromCapture(t *testing.T) {
	capture := image.NewRGBA(image.Rect(0, 0, game.GAME_WIDTH, game.GAME_HEIGHT))
	pivotX := BATTLE_INVENTORY_MONEY_PIVOT.X + 1
	pivotY := BATTLE_INVENTORY_MONEY_PIVOT.Y + 1
	setBattleCaptureColor(capture, pivotX, pivotY, BATTLE_INVENTORY_MONEY_PIVOT.Color)

	gotX, gotY, found := getInventoryPosFromCapture(capture)
	wantX := pivotX - 78
	wantY := pivotY + 20
	if gotX != wantX || gotY != wantY || !found {
		t.Fatalf("getInventoryPosFromCapture() = (%d, %d, %t), want (%d, %d, true)", gotX, gotY, found, wantX, wantY)
	}
}

func TestBattleWindowPositionSearchHandlesNilCapture(t *testing.T) {
	tests := []struct {
		name   string
		search func(*image.RGBA) (int32, int32, bool)
	}{
		{name: "skill window", search: getSkillWindowPosFromCapture},
		{name: "inventory window", search: getInventoryPosFromCapture},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if x, y, found := test.search(nil); x != 0 || y != 0 || found {
				t.Fatalf("search(nil) = (%d, %d, %t), want (0, 0, false)", x, y, found)
			}
		})
	}
}

func TestGetSelfTargetFromCapturePreservesGroupPriority(t *testing.T) {
	tests := []struct {
		name           string
		characterFirst bool
		want           *game.CheckTarget
	}{
		{name: "characters first", characterFirst: true, want: &allCharacters[0]},
		{name: "pets first", characterFirst: false, want: &allPets[0]},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			capture := image.NewRGBA(image.Rect(0, 0, game.GAME_WIDTH, game.GAME_HEIGHT))
			markSelfTarget(capture, &allCharacters[0], COLOR_BATTLE_SELF_NAME_1)
			markSelfTarget(capture, &allPets[0], COLOR_BATTLE_SELF_NAME_1)

			got, found := getSelfTargetFromCapture(capture, test.characterFirst)
			if got != test.want || !found {
				t.Fatalf("getSelfTargetFromCapture() = (%v, %t), want (%v, true)", got, found, test.want)
			}
		})
	}
}

func TestGetSelfTargetFromCaptureSkipsMissingPlayer(t *testing.T) {
	capture := image.NewRGBA(image.Rect(0, 0, game.GAME_WIDTH, game.GAME_HEIGHT))
	setBattleCaptureColor(capture, allCharacters[0].X+8, allCharacters[0].Y-10, COLOR_BATTLE_SELF_NAME_1)
	markSelfTarget(capture, &allCharacters[1], COLOR_BATTLE_SELF_NAME_1)

	got, found := getSelfTargetFromCapture(capture, true)
	want := &allCharacters[1]
	if got != want || !found {
		t.Fatalf("getSelfTargetFromCapture() = (%v, %t), want (%v, true)", got, found, want)
	}
}

func TestGetSelfTargetFromCaptureRecognizesAllNameColors(t *testing.T) {
	nameColors := []win.COLORREF{COLOR_BATTLE_SELF_NAME_1, COLOR_BATTLE_SELF_NAME_2, COLOR_BATTLE_SELF_NAME_3, COLOR_BATTLE_SELF_NAME_4, COLOR_BATTLE_SELF_NAME_5}

	for _, nameColor := range nameColors {
		t.Run(nameColorTestName(nameColor), func(t *testing.T) {
			capture := image.NewRGBA(image.Rect(0, 0, game.GAME_WIDTH, game.GAME_HEIGHT))
			markSelfTarget(capture, &allCharacters[0], nameColor)

			got, found := getSelfTargetFromCapture(capture, true)
			want := &allCharacters[0]
			if got != want || !found {
				t.Fatalf("getSelfTargetFromCapture() = (%v, %t), want (%v, true)", got, found, want)
			}
		})
	}
}

func TestGetSelfTargetFromCaptureHandlesNilCapture(t *testing.T) {
	if target, found := getSelfTargetFromCapture(nil, true); target != nil || found {
		t.Fatalf("getSelfTargetFromCapture(nil) = (%v, %t), want (nil, false)", target, found)
	}
}

func markSelfTarget(capture *image.RGBA, target *game.CheckTarget, nameColor win.COLORREF) {
	setBattleCaptureColor(capture, target.X, target.Y, COLOR_BATTLE_BLOOD_UPPER)
	setBattleCaptureColor(capture, target.X+8, target.Y-10, nameColor)
}

func nameColorTestName(nameColor win.COLORREF) string {
	return fmt.Sprintf("color %d", nameColor)
}

func setBattleCaptureColor(capture *image.RGBA, x, y int32, colorRef win.COLORREF) {
	capture.SetRGBA(int(x), int(y), color.RGBA{
		R: byte(colorRef),
		G: byte(colorRef >> 8),
		B: byte(colorRef >> 16),
		A: 0xff,
	})
}

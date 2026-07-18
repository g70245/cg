package battle

import (
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

func setBattleCaptureColor(capture *image.RGBA, x, y int32, colorRef win.COLORREF) {
	capture.SetRGBA(int(x), int(y), color.RGBA{
		R: byte(colorRef),
		G: byte(colorRef >> 8),
		B: byte(colorRef >> 16),
		A: 0xff,
	})
}

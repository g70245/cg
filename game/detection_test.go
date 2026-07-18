package game

import (
	"image"
	"image/color"
	"testing"

	"cg/game/items"

	"github.com/g70245/win"
)

func TestGetItemPosFromCapture(t *testing.T) {
	const (
		originX     int32        = 100
		originY     int32        = 120
		targetColor win.COLORREF = items.COLOR_ITEM_POTION
	)

	tests := []struct {
		name        string
		granularity int32
		prepare     func(*image.RGBA)
		wantX       int32
		wantY       int32
		wantFound   bool
	}{
		{
			name:        "finds an item at granularity one",
			granularity: 1,
			prepare: func(capture *image.RGBA) {
				setCaptureColor(capture, originX+1, originY+1, targetColor)
			},
			wantX:     originX + 1,
			wantY:     originY + 1,
			wantFound: true,
		},
		{
			name:        "coarser granularity can skip the item pixel",
			granularity: 2,
			prepare: func(capture *image.RGBA) {
				setCaptureColor(capture, originX+1, originY+1, targetColor)
			},
		},
		{
			name:        "returns the first matching inventory slot",
			granularity: 1,
			prepare: func(capture *image.RGBA) {
				setCaptureColor(capture, originX+ITEM_COL_LEN+4, originY+3, targetColor)
				setCaptureColor(capture, originX+2*ITEM_COL_LEN+4, originY+3, targetColor)
			},
			wantX:     originX + ITEM_COL_LEN + 4,
			wantY:     originY + 3,
			wantFound: true,
		},
		{
			name:        "skips a disabled slot and searches the next slot",
			granularity: 1,
			prepare: func(capture *image.RGBA) {
				setCaptureColor(capture, originX, originY, items.COLOR_ITEM_CAN_NOT_BE_USED)
				setCaptureColor(capture, originX+1, originY, targetColor)
				setCaptureColor(capture, originX+ITEM_COL_LEN+2, originY+2, targetColor)
			},
			wantX:     originX + ITEM_COL_LEN + 2,
			wantY:     originY + 2,
			wantFound: true,
		},
		{
			name:        "returns not found when no item matches",
			granularity: 1,
			prepare:     func(*image.RGBA) {},
		},
		{
			name:        "rejects non-positive granularity",
			granularity: 0,
			prepare: func(capture *image.RGBA) {
				setCaptureColor(capture, originX, originY, targetColor)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			capture := image.NewRGBA(image.Rect(0, 0, GAME_WIDTH, GAME_HEIGHT))
			test.prepare(capture)

			gotX, gotY, gotFound := getItemPosFromCapture(capture, originX, originY, targetColor, test.granularity)
			if gotX != test.wantX || gotY != test.wantY || gotFound != test.wantFound {
				t.Fatalf("getItemPosFromCapture() = (%d, %d, %t), want (%d, %d, %t)", gotX, gotY, gotFound, test.wantX, test.wantY, test.wantFound)
			}
		})
	}
}

func TestGetItemPosFromCaptureHandlesNilCapture(t *testing.T) {
	if x, y, found := getItemPosFromCapture(nil, 0, 0, items.COLOR_ITEM_POTION, 1); x != 0 || y != 0 || found {
		t.Fatalf("getItemPosFromCapture(nil) = (%d, %d, %t), want (0, 0, false)", x, y, found)
	}
}

func setCaptureColor(capture *image.RGBA, x, y int32, colorRef win.COLORREF) {
	capture.SetRGBA(int(x), int(y), color.RGBA{
		R: byte(colorRef),
		G: byte(colorRef >> 8),
		B: byte(colorRef >> 16),
		A: 0xff,
	})
}

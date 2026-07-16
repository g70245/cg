package container

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2/canvas"
	fynetest "fyne.io/fyne/v2/test"
)

func TestTagContainerUsesContentWidth(t *testing.T) {
	testApp := fynetest.NewApp()
	defer testApp.Quit()

	shortText := "A"
	longText := "A much longer action tag"
	tagColor := color.RGBA{0x33, 0x4E, 0x9A, 0xFF}

	shortTag := newTagContainer(shortText, tagColor)
	longTag := newTagContainer(longText, tagColor)

	if got := shortTag.MinSize().Width; got < tagMinimumWidth {
		t.Fatalf("short tag width = %v, want at least %v", got, tagMinimumWidth)
	}
	if got, wantGreaterThan := shortTag.MinSize().Width, canvas.NewText(shortText, color.White).MinSize().Width; got <= wantGreaterThan {
		t.Fatalf("padded tag width = %v, want greater than text width %v", got, wantGreaterThan)
	}
	if got, wantGreaterThan := longTag.MinSize().Width, shortTag.MinSize().Width; got <= wantGreaterThan {
		t.Fatalf("long tag width = %v, want greater than short tag width %v", got, wantGreaterThan)
	}
}

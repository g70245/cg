package internal

import (
	"image"
	"image/color"
	"testing"

	"github.com/g70245/win"
)

const testColorRef = win.COLORREF(0xC7FBFF)

func TestRGBAAreaContainsColor(t *testing.T) {
	capture := image.NewRGBA(image.Rect(0, 0, 4, 4))
	capture.SetRGBA(2, 1, color.RGBA{R: 0xff, G: 0xfb, B: 0xc7, A: 0xff})

	tests := []struct {
		name                                         string
		originX, originY, destinationX, destinationY int32
		want                                         bool
	}{
		{name: "finds COLORREF converted to RGB", originX: 2, originY: 1, destinationX: 2, destinationY: 1, want: true},
		{name: "does not search outside requested area", originX: 0, originY: 0, destinationX: 1, destinationY: 1},
		{name: "clips negative coordinates", originX: -5, originY: -5, destinationX: 2, destinationY: 1, want: true},
		{name: "clips coordinates beyond bounds", originX: 2, originY: 1, destinationX: 10, destinationY: 10, want: true},
		{name: "rejects area outside image", originX: 5, originY: 5, destinationX: 10, destinationY: 10},
		{name: "rejects inverted horizontal range", originX: 2, originY: 0, destinationX: 1, destinationY: 3},
		{name: "rejects inverted vertical range", originX: 0, originY: 2, destinationX: 3, destinationY: 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := RGBAAreaContainsColor(capture, test.originX, test.originY, test.destinationX, test.destinationY, testColorRef)
			if got != test.want {
				t.Fatalf("RGBAAreaContainsColor() = %t, want %t", got, test.want)
			}
		})
	}
}

func TestRGBAAreaContainsColorHandlesNilCapture(t *testing.T) {
	if RGBAAreaContainsColor(nil, 0, 0, 1, 1, testColorRef) {
		t.Fatal("RGBAAreaContainsColor(nil) = true, want false")
	}
}

func TestRGBAFindColor(t *testing.T) {
	capture := image.NewRGBA(image.Rect(0, 0, 4, 4))
	capture.SetRGBA(0, 2, color.RGBA{R: 0xff, G: 0xfb, B: 0xc7, A: 0xff})
	capture.SetRGBA(1, 0, color.RGBA{R: 0xff, G: 0xfb, B: 0xc7, A: 0xff})

	tests := []struct {
		name                                         string
		originX, originY, destinationX, destinationY int32
		granularity                                  int32
		wantX, wantY                                 int32
		wantFound                                    bool
	}{
		{name: "returns the first match in column-major order", originX: 0, originY: 0, destinationX: 3, destinationY: 3, granularity: 1, wantX: 0, wantY: 2, wantFound: true},
		{name: "clips coordinates outside the capture", originX: -2, originY: -2, destinationX: 1, destinationY: 0, granularity: 1, wantX: 1, wantY: 0, wantFound: true},
		{name: "returns not found outside the capture", originX: 5, originY: 5, destinationX: 8, destinationY: 8, granularity: 1},
		{name: "rejects non-positive granularity", originX: 0, originY: 0, destinationX: 3, destinationY: 3, granularity: 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotX, gotY, gotFound := RGBAFindColor(capture, test.originX, test.originY, test.destinationX, test.destinationY, testColorRef, test.granularity)
			if gotX != test.wantX || gotY != test.wantY || gotFound != test.wantFound {
				t.Fatalf("RGBAFindColor() = (%d, %d, %t), want (%d, %d, %t)", gotX, gotY, gotFound, test.wantX, test.wantY, test.wantFound)
			}
		})
	}
}

func TestRGBAFindColorRespectsGranularity(t *testing.T) {
	capture := image.NewRGBA(image.Rect(0, 0, 4, 4))
	capture.SetRGBA(1, 1, color.RGBA{R: 0xff, G: 0xfb, B: 0xc7, A: 0xff})

	if x, y, found := RGBAFindColor(capture, 0, 0, 3, 3, testColorRef, 2); x != 0 || y != 0 || found {
		t.Fatalf("RGBAFindColor() with granularity 2 = (%d, %d, %t), want (0, 0, false)", x, y, found)
	}
	if x, y, found := RGBAFindColor(capture, 0, 0, 3, 3, testColorRef, 1); x != 1 || y != 1 || !found {
		t.Fatalf("RGBAFindColor() with granularity 1 = (%d, %d, %t), want (1, 1, true)", x, y, found)
	}
}

func TestRGBAFindColorHandlesNilCapture(t *testing.T) {
	if x, y, found := RGBAFindColor(nil, 0, 0, 1, 1, testColorRef, 1); x != 0 || y != 0 || found {
		t.Fatalf("RGBAFindColor(nil) = (%d, %d, %t), want (0, 0, false)", x, y, found)
	}
}

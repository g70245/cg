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

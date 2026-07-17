package internal

import (
	"fmt"
	"image"
	"unsafe"

	"github.com/g70245/win"
)

func CaptureClientArea(hWnd win.HWND, x, y, width, height int32) (*image.RGBA, error) {
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("capture client area: width and height must be positive")
	}

	sourceDC := win.GetDC(hWnd)
	if sourceDC == 0 {
		return nil, fmt.Errorf("capture client area: get window device context")
	}
	defer win.ReleaseDC(hWnd, sourceDC)

	memoryDC := win.CreateCompatibleDC(sourceDC)
	if memoryDC == 0 {
		return nil, fmt.Errorf("capture client area: create compatible device context")
	}
	defer win.DeleteDC(memoryDC)

	bitmap := win.CreateCompatibleBitmap(sourceDC, width, height)
	if bitmap == 0 {
		return nil, fmt.Errorf("capture client area: create compatible bitmap")
	}
	defer win.DeleteObject(win.HGDIOBJ(bitmap))

	previousObject := win.SelectObject(memoryDC, win.HGDIOBJ(bitmap))
	if previousObject == 0 {
		return nil, fmt.Errorf("capture client area: select compatible bitmap")
	}
	bitmapSelected := true
	defer func() {
		if bitmapSelected {
			win.SelectObject(memoryDC, previousObject)
		}
	}()

	if !win.BitBlt(memoryDC, 0, 0, width, height, sourceDC, x, y, win.SRCCOPY|win.CAPTUREBLT) {
		return nil, fmt.Errorf("capture client area: copy window pixels")
	}

	if win.SelectObject(memoryDC, previousObject) == 0 {
		return nil, fmt.Errorf("capture client area: restore compatible device context")
	}
	bitmapSelected = false

	bitmapInfo := win.BITMAPINFO{
		BmiHeader: win.BITMAPINFOHEADER{
			BiSize:        uint32(unsafe.Sizeof(win.BITMAPINFOHEADER{})),
			BiWidth:       width,
			BiHeight:      -height,
			BiPlanes:      1,
			BiBitCount:    32,
			BiCompression: win.BI_RGB,
		},
	}

	pixelCount := int64(width) * int64(height)
	if pixelCount > int64(int(^uint(0)>>1))/4 {
		return nil, fmt.Errorf("capture client area: dimensions are too large")
	}

	bgra := make([]byte, int(pixelCount)*4)
	if scanLines := win.GetDIBits(memoryDC, bitmap, 0, uint32(height), &bgra[0], &bitmapInfo, win.DIB_RGB_COLORS); scanLines != height {
		return nil, fmt.Errorf("capture client area: read bitmap pixels: got %d of %d scan lines", scanLines, height)
	}

	capture := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
	for offset := 0; offset < len(bgra); offset += 4 {
		capture.Pix[offset] = bgra[offset+2]
		capture.Pix[offset+1] = bgra[offset+1]
		capture.Pix[offset+2] = bgra[offset]
		capture.Pix[offset+3] = 0xff
	}

	return capture, nil
}

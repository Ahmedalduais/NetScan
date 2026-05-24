//go:build windows

package backend

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"syscall"
	"unsafe"
)

var (
	shell32 = syscall.NewLazyDLL("shell32.dll")
	user32  = syscall.NewLazyDLL("user32.dll")
	gdi32   = syscall.NewLazyDLL("gdi32.dll")

	procExtractIconExW = shell32.NewProc("ExtractIconExW")
	procDestroyIcon    = user32.NewProc("DestroyIcon")
	procGetDC          = user32.NewProc("GetDC")
	procReleaseDC      = user32.NewProc("ReleaseDC")
	procCreateCompatibleDC   = gdi32.NewProc("CreateCompatibleDC")
	procDeleteDC       = gdi32.NewProc("DeleteDC")
	procCreateCompatibleBitmap = gdi32.NewProc("CreateCompatibleBitmap")
	procSelectObject   = gdi32.NewProc("SelectObject")
	procDeleteObject   = gdi32.NewProc("DeleteObject")
	procGetDIBits      = gdi32.NewProc("GetDIBits")
	procDrawIconEx     = user32.NewProc("DrawIconEx")
	procGetObjectW     = gdi32.NewProc("GetObjectW")
)

const (
	iconSize = 32 // 32x32 icon
)

// BITMAPINFOHEADER for 32-bit bitmap
type bitmapInfoHeader struct {
	Size          uint32
	Width         int32
	Height        int32
	Planes        uint16
	BitCount      uint16
	Compression   uint32
	SizeImage     uint32
	XPelsPerMeter int32
	YPelsPerMeter int32
	ClrUsed       uint32
	ClrImportant  uint32
}

type bitmapInfo struct {
	Header bitmapInfoHeader
	Colors [1]uint32
}

// extractIconWindows extracts the application icon from an executable file.
// Returns base64-encoded PNG string.
func extractIconWindows(exePath string) string {
	ptr, err := syscall.UTF16PtrFromString(exePath)
	if err != nil {
		return ""
	}

	var hIcon syscall.Handle

	// Extract large icon (32x32)
	ret, _, _ := procExtractIconExW.Call(
		uintptr(unsafe.Pointer(ptr)),
		0, // first icon
		uintptr(unsafe.Pointer(&hIcon)), // large icon
		0, // small icon (none)
		1, // extract 1 icon
	)

	if ret == 0 || hIcon == 0 {
		return ""
	}
	defer procDestroyIcon.Call(uintptr(hIcon))

	// Get screen DC
	screenDC, _, _ := procGetDC.Call(0)
	if screenDC == 0 {
		return ""
	}
	defer procReleaseDC.Call(0, screenDC)

	// Create compatible DC
	memDC, _, _ := procCreateCompatibleDC.Call(screenDC)
	if memDC == 0 {
		return ""
	}
	defer procDeleteDC.Call(memDC)

	// Create compatible bitmap
	hBitmap, _, _ := procCreateCompatibleBitmap.Call(screenDC, iconSize, iconSize)
	if hBitmap == 0 {
		return ""
	}
	defer procDeleteObject.Call(hBitmap)

	// Select bitmap into DC
	oldObj, _, _ := procSelectObject.Call(memDC, hBitmap)

	// Draw the icon onto the bitmap
	procDrawIconEx.Call(
		memDC,
		0, 0, // x, y
		uintptr(hIcon),
		iconSize, iconSize,
		0,    // frame (animated icon)
		0,    // reserved
		0x0003, // DI_NORMAL | DI_COMPAT
	)

	// Select old object back
	procSelectObject.Call(memDC, oldObj)

	// Read bitmap bits
	img := iconToImage(hBitmap)
	if img == nil {
		return ""
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return ""
	}

	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
}

// iconToImage converts a GDI bitmap handle to a Go image.
func iconToImage(hBitmap uintptr) image.Image {
	// Get bitmap info
	var bmp struct {
		Type      int16
		Width     int32
		Height    int32
		WidthBytes int32
		Planes    uint16
		BitsPixel uint16
	}
	procGetObjectW.Call(hBitmap, unsafe.Sizeof(bmp), uintptr(unsafe.Pointer(&bmp)))

	if bmp.Width <= 0 || bmp.Height <= 0 {
		return nil
	}

	width := int(bmp.Width)
	height := int(bmp.Height)

	// Prepare bitmap info header
	bmpInfo := bitmapInfoHeader{
		Size:         uint32(unsafe.Sizeof(bitmapInfoHeader{})),
		Width:        int32(width),
		Height:       -int32(height), // negative = top-down bitmap
		Planes:       1,
		BitCount:     32, // BGRA
		Compression:  0,  // BI_RGB
	}

	// Allocate pixel buffer
	pixelBuf := make([]byte, width*height*4)

	// Get DIBits
	screenDC, _, _ := procGetDC.Call(0)
	if screenDC != 0 {
		procGetDIBits.Call(
			screenDC,
			hBitmap,
			0,                               // start scan line
			uintptr(height),                 // number of scan lines
			uintptr(unsafe.Pointer(&pixelBuf[0])),
			uintptr(unsafe.Pointer(&bmpInfo)),
			0, // DIB_RGB_COLORS
		)
		procReleaseDC.Call(0, screenDC)
	}

	// Create Go image from BGRA data
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := (y*width + x) * 4
			// BGRA -> RGBA
			img.Pix[y*img.Stride+x*4+0] = pixelBuf[idx+2] // R
			img.Pix[y*img.Stride+x*4+1] = pixelBuf[idx+1] // G
			img.Pix[y*img.Stride+x*4+2] = pixelBuf[idx+0] // B
			img.Pix[y*img.Stride+x*4+3] = pixelBuf[idx+3] // A
		}
	}

	// Trim transparent edges
	return trimTransparent(img)
}

// trimTransparent removes fully transparent rows/columns from the edges.
func trimTransparent(img *image.RGBA) *image.RGBA {
	bounds := img.Bounds()
	if bounds.Empty() {
		return img
	}

	minX, minY := bounds.Max.X, bounds.Max.Y
	maxX, maxY := bounds.Min.X, bounds.Min.Y

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			idx := (y-bounds.Min.Y)*img.Stride + (x-bounds.Min.X)*4
			if img.Pix[idx+3] > 0 { // has alpha
				if x < minX {
					minX = x
				}
				if y < minY {
					minY = y
				}
				if x > maxX {
					maxX = x
				}
				if y > maxY {
					maxY = y
				}
			}
		}
	}

	if minX > maxX || minY > maxY {
		return img // fully transparent
	}

	// Add 1px padding
	if minX > 0 {
		minX--
	}
	if minY > 0 {
		minY--
	}
	if maxX < bounds.Max.X-1 {
		maxX++
	}
	if maxY < bounds.Max.Y-1 {
		maxY++
	}

	cropped := image.NewRGBA(image.Rect(0, 0, maxX-minX+1, maxY-minY+1))
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			srcIdx := (y-bounds.Min.Y)*img.Stride + (x-bounds.Min.X)*4
			dstIdx := (y-minY)*cropped.Stride + (x-minX)*4
			copy(cropped.Pix[dstIdx:dstIdx+4], img.Pix[srcIdx:srcIdx+4])
		}
	}

	// Ensure minimum size
	if cropped.Bounds().Dx() < 16 || cropped.Bounds().Dy() < 16 {
		return img
	}

	return cropped
}

// Ensure unused import is used
var _ = fmt.Sprintf

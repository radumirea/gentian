package painter

import (
	"image"
	"image/color"
)

func PaintHLine(img *image.RGBA, x1, x2, y int, col color.Color) {
	for ; x1 <= x2; x1++ {
		img.Set(x1, y, col)
	}
}

func PaintVLine(img *image.RGBA, x, y1, y2 int, col color.Color) {
	for ; y1 <= y2; y1++ {
		img.Set(x, y1, col)
	}
}

func PaintBorder(img *image.RGBA, col color.Color) {
	PaintHLine(img, 0, img.Rect.Dx()-1, 0, col)
	PaintHLine(img, 0, img.Rect.Dx()-1, img.Rect.Dy()-1, col)
	PaintVLine(img, 0, 0, img.Rect.Dy()-1, col)
	PaintVLine(img, img.Rect.Dx()-1, 0, img.Rect.Dy()-1, col)
}

package utils

import (
	"image"
	"image/draw"
	"path"
)

// ImageToRGBA converts an image.Image to an *image.RGBA
func ImageToRGBA(src image.Image) *image.RGBA {
	b := src.Bounds()
	rgba := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(rgba, rgba.Bounds(), src, b.Min, draw.Src)
	return rgba
}

// SetExtension sets the file extension to nexExt and removes the old one
func SetExtension(filename string, newExt string) string {
	ext := path.Ext(filename)
	return filename[0:len(filename)-len(ext)] + newExt
}

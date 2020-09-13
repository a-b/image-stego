package chunk

import (
	"image"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"os"

	"dennis-tra/image-stego/internal/utils"
)

// OpenImageFile opens the file at the given path and returns the decoded *image.RGBA
func OpenImageFile(filename string) (*image.RGBA, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	err = file.Close()
	if err != nil {
		return nil, err
	}

	return utils.ImageToRGBA(img), nil
}

// SaveImageFile saves the given image data to the given filepath as a PNG image.
func SaveImageFile(filepath string, img image.Image) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	err = png.Encode(file, img)
	if err != nil {
		return err
	}

	return nil
}

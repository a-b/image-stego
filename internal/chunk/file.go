package chunk

import (
	"dennis-tra/image-stego/internal/utils"
	"image"
	"os"
	_ "image/jpeg"
	_ "image/png"
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

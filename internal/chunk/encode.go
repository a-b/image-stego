package chunk

import (
	"dennis-tra/image-stego/internal/utils"
	"encoding/hex"
	"github.com/cbergoon/merkletree"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
	"path"
)

func Encode(filepath string, outdir string) error {
	filename := path.Base(filepath)

	log.Println("Opening image:", filepath)
	rgba, err := OpenImageFile(filepath)
	if err != nil {
		return err
	}

	// copy original image for the overlay
	overlayImage := utils.ImageToRGBA(rgba.SubImage(rgba.Bounds()))

	var list []merkletree.Content

	log.Println("Calculating bounds...")
	bounds := CalculateChunkBounds(rgba)

	log.Println("Building merkle tree...")
	// Build merkle tree
	for _, boundRow := range bounds {
		for _, bound := range boundRow {

			chunk := &Chunk{RGBA: image.NewRGBA(bound)}

			for x := 0; x < bound.Dx(); x++ {
				for y := 0; y < bound.Dy(); y++ {
					original := bound.Min.Add(image.Pt(x, y))
					chunk.Set(original.X, original.Y, rgba.RGBAAt(original.X, original.Y))
				}
			}
			list = append(list, chunk)
		}
	}

	// Create a new Merkle Tree from the list of Content
	t, err := merkletree.NewTree(list)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Merkle Tree Root Hash:")
	log.Println("\t", hex.EncodeToString(t.MerkleRoot()))

	log.Println("Drawing overlay image...")
	// Draw overlay image
	for cx, boundRow := range bounds {
		for cy, bound := range boundRow {

			chunk := &Chunk{RGBA: image.NewRGBA(bound)}

			// Draw mask image
			var clr color.RGBA
			if (cx%2 == 0 && cy%2 == 0) || (cx%2 != 0 && cy%2 != 0) {
				clr = color.RGBA{B: 255, A: 255}
			} else {
				clr = color.RGBA{R: 255, A: 255}
			}

			draw.DrawMask(
				overlayImage,
				chunk.Bounds(),
				&image.Uniform{C: clr},
				image.Point{},
				&image.Uniform{C: color.RGBA{R: 255, G: 255, B: 255, A: 80}},
				image.Point{},
				draw.Over,
			)
		}
	}
	overlay := path.Join(outdir, utils.SetExtension(filename, ".overlay.png"))
	log.Println("Saving overlay image:", overlay)

	out, err := os.Create(overlay)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	err = png.Encode(out, overlayImage)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Encoding Merkle Tree information into LSBs of the image")
	destImg := image.NewRGBA(rgba.Bounds())
	for cx, boundRow := range bounds {
		for cy, _ := range boundRow {
			chunk := list[cx*len(boundRow)+cy].(*Chunk)

			paths, sides, err := t.GetMerklePath(chunk)
			if err != nil {
				log.Fatal(err)
			}

			writeBuffer := []byte{}
			writeBuffer = append(writeBuffer, uint8(len(paths)))
			for i, path := range paths {
				side := uint8(sides[i])
				writeBuffer = append(writeBuffer, side)
				writeBuffer = append(writeBuffer, path...)
			}

			_, err = chunk.Write(writeBuffer)
			if err != nil {
				log.Fatal(err)
			}

			draw.Draw(destImg, chunk.Bounds(), chunk, chunk.Bounds().Min, draw.Src)
		}
	}

	encoded := path.Join(outdir, utils.SetExtension(filename, ".png"))
	log.Println("Saving encoded image:", encoded)
	out, err = os.Create(encoded)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	err = png.Encode(out, destImg)
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

package main

import (
	"encoding/hex"
	"fmt"
	"github.com/cbergoon/merkletree"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
)

func encode(rgba *image.RGBA) {

	overlayImage := imageToRGBA(rgba.SubImage(rgba.Bounds()))

	var list []merkletree.Content

	bounds := chunkBounds(rgba)

	for cx, boundRow := range bounds {
		for cy, bound := range boundRow {

			chunk := &Chunk{RGBA: image.NewRGBA(bound)}

			for x := 0; x < bound.Dx(); x++ {
				for y := 0; y < bound.Dy(); y++ {
					original := bound.Min.Add(image.Pt(x, y))
					chunk.Set(original.X, original.Y, rgba.RGBAAt(original.X, original.Y))
				}
			}
			list = append(list, chunk)

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

	// Create a new Merkle Tree from the list of Content
	t, err := merkletree.NewTree(list)
	if err != nil {
		log.Fatal(err)
	}

	out, err := os.Create("out/overlay.png")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	err = png.Encode(out, overlayImage)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Merkle Tree Root:", hex.EncodeToString(t.MerkleRoot()))
	//---------------------

	ch, err := os.Create("out/chunk-hashes.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	destImg := image.NewRGBA(rgba.Bounds())

	for cx, boundRow := range bounds {
		for cy, _ := range boundRow {
			chunk := list[cx*len(boundRow)+cy].(*Chunk)

			paths, sides, err := t.GetMerklePath(chunk)
			if err != nil {
				log.Fatal(err)
			}

			chunkHash, err := chunk.CalculateHash()
			if err != nil {
				log.Fatal(err)
			}

			_, err = ch.Write([]byte(fmt.Sprintf("%02d_%02d.png,%s\n", cx, cy, hex.EncodeToString(chunkHash))))
			if err != nil {
				log.Fatal(err)
			}

			chunkFile, err := os.Create(fmt.Sprintf("out/chunks/%02d_%02d.png", cx, cy))
			if err != nil {
				log.Fatal(err)
			}
			defer chunkFile.Close()

			chunkImage := image.NewRGBA(chunk.Bounds())
			draw.Draw(chunkImage, chunk.Bounds(), chunk, chunk.Bounds().Min, draw.Src)

			err = png.Encode(chunkFile, chunkImage)
			if err != nil {
				log.Fatal(err)
			}

			writeBuffer := []byte{uint8(len(paths))}
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

	out, err = os.Create("out/encoded.png")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	err = png.Encode(out, destImg)
	if err != nil {
		log.Fatal(err)
	}
}

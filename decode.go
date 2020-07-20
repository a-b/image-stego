package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
)

type Index struct {
	cx int
	cy int
}

func decode(rgba *image.RGBA) {

	bounds := chunkBounds(rgba)

	roots := map[string][]Index{}

	for cx, boundRow := range bounds {
		for cy, bound := range boundRow {

			chunk := &Chunk{RGBA: image.NewRGBA(bound)}

			for x := 0; x < bound.Dx(); x++ {
				for y := 0; y < bound.Dy(); y++ {
					original := bound.Min.Add(image.Pt(x, y))
					chunk.Set(original.X, original.Y, rgba.RGBAAt(original.X, original.Y))
				}
			}

			pathCount := make([]byte, 1)
			_, err := chunk.Read(pathCount)
			if err != nil {
				log.Fatal(err)
			}

			chunkHash, _ := chunk.CalculateHash()
			prevHash := chunkHash
			for i := 0; i < int(pathCount[0]); i++ {
				side := make([]byte, 1)
				data := make([]byte, 32)

				_, err := chunk.Read(side)
				if err != nil {
					//log.Fatal(err)
				}

				_, err = chunk.Read(data)
				if err != nil {
					//log.Fatal(err)
				}

				hsh := sha256.New()

				if side[0] == 0 {
					prevHash = append(data, prevHash...)
				} else {
					prevHash = append(prevHash, data...)
				}

				hsh.Write(prevHash)
				prevHash = hsh.Sum(nil)
			}

			merkleRoot := hex.EncodeToString(prevHash)

			_, exists := roots[merkleRoot]
			if !exists {
				roots[merkleRoot] = []Index{}
			}
			roots[merkleRoot] = append(roots[merkleRoot], Index{cx: cx, cy: cy})
		}
	}

	m := 0
	canonicalMerkleRoot := ""
	for merkleRoot, indices := range roots {
		if len(indices) > m {
			m = len(indices)
			canonicalMerkleRoot = merkleRoot
		}
	}

	if len(roots) == 1 {
		fmt.Println("This image is sane - the merkle root:", canonicalMerkleRoot)
		return
	} else {
		fmt.Println("The merkle root, that appeared multiple times is:", canonicalMerkleRoot)
	}

	overlayImage := imageToRGBA(rgba.SubImage(rgba.Bounds()))

	for merkleRoot, indices := range roots {
		for _, idx := range indices {
			bound := bounds[idx.cx][idx.cy]
			chunk := &Chunk{RGBA: image.NewRGBA(bound)}

			if merkleRoot != canonicalMerkleRoot {
				draw.DrawMask(
					overlayImage,
					chunk.Bounds(),
					&image.Uniform{C: color.RGBA{R: 255, A: 255}},
					image.Point{},
					&image.Uniform{C: color.RGBA{R: 255, G: 255, B: 255, A: 80}},
					image.Point{},
					draw.Over,
				)
			}
		}
	}

	out, err := os.Create("out/decoded-overlay.png")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	err = png.Encode(out, overlayImage)
	if err != nil {
		log.Fatal(err)
	}
}

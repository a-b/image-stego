package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	"log"
)

func decode(rgba *image.RGBA) {

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
					log.Fatal(err)
				}
				_, err = chunk.Read(data)
				if err != nil {
					log.Fatal(err)
				}

				hsh := sha256.New()
				buffer := []byte{}

				if side[0] == 0 {
					buffer = append(buffer, data...)
					buffer = append(buffer, prevHash...)
				} else if side[0] == 1 {
					buffer = append(buffer, prevHash...)
					buffer = append(buffer, data...)
				} else {
					log.Fatal("Unsupported side")
				}

				hsh.Write(buffer)
				prevHash = hsh.Sum(nil)
			}

			fmt.Printf("Chunk (%02d/%02d) (Paths: %d) - %s - %s\n", cx, cy, pathCount[0], hex.EncodeToString(prevHash), hex.EncodeToString(chunkHash))
		}
	}
}

//f790c9c0597a04e3a119bfcbf1d0cfe3b977df7d98a0f54c45d6fc822bc52d52

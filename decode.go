package main

import (
	"encoding/hex"
	"fmt"
	"image"
	"log"
	"os"
)

func Decode() {

	filename := "outimage.png"
	fmt.Println("Operating on", filename)

	fmt.Println("Opening...")
	reader, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Decoding...")
	m, _, err := image.Decode(reader)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Closing...")
	err = reader.Close()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Image file dimensions", m.Bounds().Max)
	totalPixels := m.Bounds().Size().X * m.Bounds().Size().Y
	fmt.Println("Image total pixels", totalPixels)
	lsbBits := totalPixels * 3
	lsbBytes := float32(lsbBits) / 8.0
	fmt.Println("LSB Bits", lsbBits)
	fmt.Println("LSB Bytes", lsbBytes)
	//maxNumHashes := lsbBits / 256
	//fmt.Println("Maximum number of hashes", maxNumHashes)
	//
	//numOfChunks := 1.0
	//
	//for numOfChunks*math.Log2(numOfChunks) < float64(maxNumHashes)  {
	//	numOfChunks++
	//}
	//fmt.Println("Total number of chunks", numOfChunks)
	//

	rgbaImage := imageToRGBA(m)

	chunkCountX := 4
	chunkCountY := 2
	chunkWidth := m.Bounds().Size().X / chunkCountX
	chunkHeight := m.Bounds().Size().Y / chunkCountY

	fmt.Println("Chunk counts: ", chunkCountX, chunkCountY)
	fmt.Println("Chunk dimensions: ", chunkWidth, chunkHeight)

	fmt.Println("Start building merkle tree...")
	for cx := 0; cx < chunkCountX; cx++ {
		for cy := 0; cy < chunkCountY; cy++ {
			fmt.Println("-- Checking chunk at", cx, cy)
			chunk := &Chunk{
				RGBA: image.NewRGBA(image.Rect(0, 0, chunkWidth, chunkHeight)),
			}
			for x := 0; x < chunkWidth; x++ {
				for y := 0; y < chunkHeight; y++ {
					color := rgbaImage.RGBAAt(cx*chunkWidth+x, cy*chunkHeight+y)
					chunk.Set(x, y, color)
				}
			}

			h, err := chunk.LSBHash()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("MerkleRoot Calculated", hex.EncodeToString(h))
		}
	}
}

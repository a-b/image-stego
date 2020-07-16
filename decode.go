package main

import (
	"crypto/sha256"
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

	for cx := 0; cx < chunkCountX; cx++ {
		for cy := 0; cy < chunkCountY; cy++ {
			chunk := &Chunk{
				RGBA: image.NewRGBA(image.Rect(0, 0, chunkWidth, chunkHeight)),
			}

			for x := 0; x < chunkWidth; x++ {
				for y := 0; y < chunkHeight; y++ {
					color := rgbaImage.RGBAAt(cx*chunkWidth+x, cy*chunkHeight+y)
					chunk.Set(x, y, color)
				}
			}


			prevHash, _ := chunk.CalculateHash()
			fmt.Println("-- Checking chunk at", cx, cy, hex.EncodeToString(prevHash))


			buffer := make([]byte, 99)
			_, err = chunk.Read(buffer)
			if err != nil {
				log.Fatal(err)
			}

			pathCount := buffer[0]
			fmt.Println("Read Path length", pathCount)
			if pathCount != 3 {
				continue
			}

			fmt.Println("Chunk hash ", hex.EncodeToString(prevHash))
			if err != nil {
				log.Fatal(err)
			}
			i := 1
			hsh := sha256.New()
			w := []byte{}
			for i+32 < len(buffer) {
				side := buffer[i]
				data := buffer[i : i+32]
				fmt.Println("side", side)

				i += 33

				if side == 1 { // left
					w = append(w, data...)
					w = append(w, prevHash...)
				} else if side == 0 {
					w = append(w, prevHash...)
					w = append(w, data...)
				} else {
					continue
				}
				hsh.Write(w)
				prevHash = hsh.Sum(nil)
			}
			log.Println("MERKLE ROOT", hex.EncodeToString(prevHash))
		}
	}
}

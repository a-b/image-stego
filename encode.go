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

	//chunk := Chunk{RGBA: rgba}
	//
	//hashBitLength := 256
	//merkleSideBitLength := 1
	//maxNumHashes := chunk.LSBCount() / (hashBitLength + merkleSideBitLength)
	//
	//numOfChunks := 1
	//for numOfChunks*int(math.Floor(math.Log2(float64(numOfChunks)))) < maxNumHashes {
	//	numOfChunks++
	//}
	//
	//if numOfChunks%2 != 0 {
	//	numOfChunks -= 1
	//}
	//fmt.Println("Total number of chunks", numOfChunks)
	//
	//// Calculate optimal distribution of chunks along width and height
	//factors := primeFactors(numOfChunks)
	//chunkCountX := factors[len(factors)-1]
	//chunkCountY := factors[len(factors)-2]
	//for i := len(factors) - 3; i >= 0; i-- {
	//	if chunkCountX > chunkCountY {
	//		chunkCountY *= factors[i]
	//	} else {
	//		chunkCountX *= factors[i]
	//	}
	//}
	//
	//if chunkCountX*chunkCountY != numOfChunks {
	//	log.Fatal("AAAH")
	//}
	//
	//fmt.Printf("Divide width in %d parts\n", chunkCountX)
	//fmt.Printf("Divide height in %d parts\n", chunkCountY)
	//
	//// Add clippings (the side length to chunk count ration will likely be rational so we add the remainder to the
	//// side lengths equally.
	//chunkWidth := chunk.Width() / chunkCountX
	//chunkHeight := chunk.Height() / chunkCountY
	//
	//chunkWidthClippings := chunk.Width() % chunkCountX
	//chunkHeightClippings := chunk.Height() % chunkCountY
	//
	//fmt.Println("Start building merkle tree...")
	overlayImage := imageToRGBA(rgba.SubImage(rgba.Bounds()))

	//cxOff := 0
	//cyOff := 0
	var list []merkletree.Content
	//for cx := 0; cx < chunkCountX; cx++ {
	//
	//	cw := chunkWidth
	//	if cx < chunkWidthClippings {
	//		cw += 1
	//		cxOff = 0
	//	} else {
	//		cxOff = chunkWidthClippings
	//	}
	//
	//	for cy := 0; cy < chunkCountY; cy++ {
	//
	//		ch := chunkHeight
	//		if cy < chunkHeightClippings {
	//			ch += 1
	//			cyOff = 0
	//		} else {
	//			cyOff = chunkHeightClippings
	//		}
	//
	//		chunk := &Chunk{RGBA: image.NewRGBA(image.Rect(0, 0, cw, ch))}
	//
	//		for x := 0; x < cw; x++ {
	//			for y := 0; y < ch; y++ {
	//				color := rgba.RGBAAt(cx*cw+x, cy*ch+y)
	//				chunk.Set(x, y, color)
	//			}
	//		}
	//		hash, _ := chunk.CalculateHash()
	//		list = append(list, chunk)
	//
	//		var clr color.RGBA
	//		if (cx%2 == 0 && cy%2 == 0) || (cx%2 != 0 && cy%2 != 0) {
	//			clr = color.RGBA{B: 255, A: 255}
	//		} else {
	//			clr = color.RGBA{R: 255, A: 255}
	//		}
	//
	//		draw.DrawMask(overlayImage, chunk.Bounds().Add(image.Pt(cxOff+cx*cw, cyOff+cy*ch)), &image.Uniform{C: clr}, image.Point{}, &image.Uniform{C: color.RGBA{R: 255, G: 255, B: 255, A: 80}}, image.Point{}, draw.Over)
	//		if cx == 0 {
	//			fmt.Printf("Chunk (%d/%d) size (%dx%d) hash: %s\n", cx, cy, cw, ch, hex.EncodeToString(hash))
	//		}
	//	}
	//}


	bounds := chunkBounds(rgba)

	for cx, boundRow := range bounds {
		for cy, bound := range boundRow {

			chunk := &Chunk{RGBA: image.NewRGBA(bound)}

			for x := 0; x < bound.Dx(); x++ {
				for y := 0; y < bound.Dy(); y++ {
					original := bound.Min.Add(image.Pt(x, y))
					color := rgba.RGBAAt(original.X, original.Y)
					chunk.Set(x, y, color)
				}
			}
			hash, _ := chunk.CalculateHash()
			list = append(list, chunk)

			var clr color.RGBA
			if (cx%2 == 0 && cy%2 == 0) || (cx%2 != 0 && cy%2 != 0) {
				clr = color.RGBA{B: 255, A: 255}
			} else {
				clr = color.RGBA{R: 255, A: 255}
			}

			draw.DrawMask(overlayImage, chunk.Bounds(), &image.Uniform{C: clr}, image.Point{}, &image.Uniform{C: color.RGBA{R: 255, G: 255, B: 255, A: 80}}, image.Point{}, draw.Over)
			if cx == 0 {
				fmt.Printf("Chunk (%d/%d) size (%dx%d) hash: %s\n", cx, cy, bound.Dx(), bound.Dy(), hex.EncodeToString(hash))
			}

		}
	}

	// Create a new Merkle Tree from the list of Content
	t, err := merkletree.NewTree(list)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Merkle Tree Root:", hex.EncodeToString(t.MerkleRoot()))

	out, err := os.Create("overlay.png")
	if err != nil {
		// Handle error
	}
	defer out.Close()

	err = png.Encode(out, overlayImage)
	if err != nil {
		log.Fatal(err)
	}
}

func primeFactors(n int) (pfs []int) {
	// Get the number of 2s that divide n
	for n%2 == 0 {
		pfs = append(pfs, 2)
		n = n / 2
	}

	// n must be odd at this point. so we can skip one element
	// (note i = i + 2)
	for i := 3; i*i <= n; i = i + 2 {
		// while i divides n, append i and divide n
		for n%i == 0 {
			pfs = append(pfs, i)
			n = n / i
		}
	}

	// This condition is to handle the case when n is a prime number
	// greater than 2
	if n > 2 {
		pfs = append(pfs, n)
	}

	return
}

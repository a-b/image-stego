package main

import (
	"image"
	"math"
)

type Grid struct {
	ChunkBounds [][]image.Rectangle
}

func NewGrid(rgba *image.RGBA) *Grid {
	
}

func chunkBounds(rgba *image.RGBA) [][]image.Rectangle {

	chunk := Chunk{RGBA: rgba}

	hashBitLength := 256
	merkleSideBitLength := 1
	maxNumHashes := chunk.LSBCount() / (hashBitLength + merkleSideBitLength)

	numOfChunks := 1
	for numOfChunks*int(math.Floor(math.Log2(float64(numOfChunks)))) < maxNumHashes {
		numOfChunks++
	}

	if numOfChunks%2 != 0 {
		numOfChunks -= 1
	}

	// Calculate optimal distribution of chunks along width and height
	factors := primeFactors(numOfChunks)
	chunkCountX := factors[len(factors)-1]
	chunkCountY := factors[len(factors)-2]
	for i := len(factors) - 3; i >= 0; i-- {
		if chunkCountX > chunkCountY {
			chunkCountY *= factors[i]
		} else {
			chunkCountX *= factors[i]
		}
	}

	// Add clippings (the side length to chunk count ration will likely be rational so we add the remainder to the
	// side lengths equally.
	chunkWidth := chunk.Width() / chunkCountX
	chunkHeight := chunk.Height() / chunkCountY

	chunkWidthClippings := chunk.Width() % chunkCountX
	chunkHeightClippings := chunk.Height() % chunkCountY

	bounds := make([][]image.Rectangle, chunkCountX)
	for i := range bounds {
		bounds[i] = make([]image.Rectangle, chunkCountY)
	}

	cxOff := 0
	cyOff := 0
	for cx := 0; cx < chunkCountX; cx++ {

		cw := chunkWidth
		if cx < chunkWidthClippings {
			cw += 1
			cxOff = 0
		} else {
			cxOff = chunkWidthClippings
		}

		for cy := 0; cy < chunkCountY; cy++ {

			ch := chunkHeight
			if cy < chunkHeightClippings {
				ch += 1
				cyOff = 0
			} else {
				cyOff = chunkHeightClippings
			}

			bounds[cx][cy] = image.Rect(cw, ch, 0, 0).Add(image.Pt(cxOff+cx*cw, cyOff+cy*ch))
		}
	}

	return bounds
}

package main

import (
	"image"
	"math"
)

func chunkBounds(rgba *image.RGBA) [][]image.Rectangle {

	chunk := Chunk{RGBA: rgba}

	chunkCount := 2
	for {
		neededBitsPerChunk := int(math.Ceil(math.Log2(float64(chunkCount))))*265 + 8 // 8 -> path count, 265 -> Merkle tree side
		neededBitsTotal := chunkCount * neededBitsPerChunk

		bitsPerChunk := chunk.LSBCount() / chunkCount
		clippings := chunk.LSBCount() % chunkCount
		bytesPerChunk := (bitsPerChunk - clippings) / 8
		if neededBitsTotal > chunk.LSBCount() || neededBitsPerChunk > bytesPerChunk*8 {
			break
		}

		chunkCount += 2
	}
	chunkCount -= 2

	hashBitLength := 256
	merkleSideBitLength := 1
	maxNumHashes := chunk.LSBCount() / (hashBitLength + merkleSideBitLength)

	numOfChunks := 1
	for numOfChunks*int(math.Ceil(math.Log2(float64(numOfChunks)))) < maxNumHashes {
		numOfChunks++
	}

	if numOfChunks%2 != 0 {
		numOfChunks -= 1
	}

	numOfChunks = chunkCount - 20

	// Calculate optimal distribution of chunks along width and height
	factors := primeFactors(numOfChunks)
	chunkCountX := factors[len(factors)-1]
	chunkCountY := 1
	if len(factors) > 1 {
		chunkCountY = factors[len(factors)-2]
	}
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

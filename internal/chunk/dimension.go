package chunk

import (
	"image"
	"math"
)

const (
	// The number of bits occupied by one SHA256 hash
	HashBitLength = 256

	// The number of bits occupied by the side information of a merkle tree leaf.
	MerkleSideBitLength = 1

	// The number of bits occupied by the information of how many merkle tree leafs are encoded in the chunk.
	PathCountBitLength = 8
)

// CalculateChunkBounds takes the given *image.RGBA and calculates the optimal distribution of image chunks
// to encode the merkle tree data.
//
// The more chunks we anticipate the smaller they become, the more of them are there and the more data needs
// to be encoded in one chunk to store all the merkle leaf data. So there is clearly an optimum of chunk
// count. Basically we want the highest number of chunks where each individual one can still store all
// the necessary merkle information.
//
// The calculation is an iterative process. The calculation starts with the assumption that we want to use
// two chunks to encode the data. First it calculates the required amount of bits to encode all merkle leafs
// within one chunk. Then it calculates the total number of available bits per chunk. In the first iteration
// the number of available bits will be much larger than the required bits.
//
// The number of bytes (not bits) that can be encoded in the chunks may differ because there may be an odd
// number of available bits per chunk, so there is a clipping of bits.
//
// If the amount of required bits exceeds the available least significant bits or exceeds the number of bits
// of the available bytes then we stop and are sure we have found the maximum number of chunks that this
// image can be divided into.
//
// Beware that with one merkle tree leaf hash (256 bits) the side of the leaf node (1 bit) needs to encoded
// and the number of leaf nodes (offset of 8) ase well.
//
// After the number of chunks have been calculated we calculate the maximum number of hashes that could
// be encoded in the whole image
func CalculateChunkBounds(rgba *image.RGBA) [][]image.Rectangle {

	chunk := Chunk{RGBA: rgba}

	// Calculate maximum number of chunks that this image can be divided into taken into account
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

	// Chunk count contains the maximum number of chunks

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
	numOfChunks -= 1

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

	// Add clippings (the side length to chunk count ratio will likely be rational so we add the remainder to the
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

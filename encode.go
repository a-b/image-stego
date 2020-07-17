package main

import (
	"fmt"
	"image"
	"log"
	"math"
)

func encode(rgba *image.RGBA) {

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
	fmt.Println("Total number of chunks", numOfChunks)

	// Calculate optimal distribution of chunks along width an height
	factors := primeFactors(numOfChunks)
	xDivisionCount := factors[len(factors)-1]
	yDivisionCount := factors[len(factors)-2]
	for i := len(factors)-3; i >= 0; i-- {
		if xDivisionCount > yDivisionCount {
			yDivisionCount *= factors[i]
		} else {
			xDivisionCount *= factors[i]
		}
	}

	if xDivisionCount * yDivisionCount != numOfChunks {
		log.Fatal("AAAH")
	}

	fmt.Printf("Divide width in %d parts\n", xDivisionCount)
	fmt.Printf("Divide height in %d parts\n", yDivisionCount)

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

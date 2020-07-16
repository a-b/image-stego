package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"image"
	"io"
	"testing"
)

// ones is a byte with all bits set to one
const ones = 0b11111111

// zeroes is a byte with all bits set to zero
const zeroes = 0b00000000

// blackImage creates an RGBA image with the given width and height
// where all pixels are black. The underlying Pix byte array
// contains w x h x 4 entries.
func blackImage(w, h int) *image.RGBA {
	return image.NewRGBA(image.Rect(0, 0, w, h))
}

// whiteImage creates an RGBA image with the given width and height
// where all pixels are white. The underlying Pix byte array
// contains w x h x 4 entries.
func whiteImage(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for i := range img.Pix {
		img.Pix[i] = 0b11111111
	}
	return img
}

func TestWrite_EmptyInput(t *testing.T) {

	chunk := Chunk{RGBA: blackImage(2, 2)}

	n, err := chunk.Write([]byte{})
	require.NoError(t, err)

	assert.Equal(t, 0, n)
	assert.Equal(t, 0, chunk.off)

	// Test expected bit representation
	for _, p := range chunk.Pix {
		assert.EqualValues(t, 0, p)
	}
}

func TestWrite_SetAllBitsToOne(t *testing.T) {

	chunk := Chunk{RGBA: blackImage(2, 2)}

	n, err := chunk.Write([]byte{ones, ones})
	require.NoError(t, err)

	assert.Equal(t, 2, n)
	assert.Equal(t, 16, chunk.off)

	// Test expected bit representation
	for _, p := range chunk.Pix {
		assert.EqualValues(t, 1, p)
	}
}

func TestWrite_SetAllBitsToOneWithBreak(t *testing.T) {

	chunk := Chunk{RGBA: blackImage(2, 2)}

	n, err := chunk.Write([]byte{ones})
	require.NoError(t, err)

	assert.Equal(t, 1, n)
	assert.Equal(t, 8, chunk.off)

	n, err = chunk.Write([]byte{ones})
	require.NoError(t, err)

	assert.Equal(t, 1, n)
	assert.Equal(t, 16, chunk.off)

	// Test expected bit representation
	for _, p := range chunk.Pix {
		assert.EqualValues(t, 1, p)
	}
}

func TestWrite_SetMixedBits(t *testing.T) {

	chunk := Chunk{RGBA: blackImage(2, 2)}

	n, err := chunk.Write([]byte{0b11110000, 0b00001111})
	require.NoError(t, err)

	assert.Equal(t, 2, n)
	assert.Equal(t, 16, chunk.off)

	// Test expected bit representation
	assert.EqualValues(t, 1, chunk.Pix[0])
	assert.EqualValues(t, 1, chunk.Pix[3])
	assert.EqualValues(t, 0, chunk.Pix[4])
	assert.EqualValues(t, 0, chunk.Pix[11])
	assert.EqualValues(t, 1, chunk.Pix[12])
}

func TestWrite_MoreThanPossible(t *testing.T) {

	chunk := Chunk{RGBA: blackImage(2, 2)}

	n, err := chunk.Write([]byte{ones, ones, ones})
	assert.EqualError(t, err, io.EOF.Error())

	assert.Equal(t, 2, n)
	assert.Equal(t, 16, chunk.off)

	// Test expected bit representation
	assert.EqualValues(t, 1, chunk.Pix[15])
}

func TestWrite_PartialByteWritten(t *testing.T) {

	chunk := Chunk{RGBA: blackImage(1, 3)} // 12 bytes

	n, err := chunk.Write([]byte{ones, ones})
	assert.EqualError(t, err, io.EOF.Error())

	assert.Equal(t, 1, n)
	assert.Equal(t, 8, chunk.off)

	// Test expected bit representation
	assert.EqualValues(t, 1, chunk.Pix[0])
	assert.EqualValues(t, 1, chunk.Pix[7])
	assert.EqualValues(t, 0, chunk.Pix[8])
	assert.EqualValues(t, 0, chunk.Pix[11])
}

func TestRead_MatchingLength(t *testing.T) {
	chunk := Chunk{RGBA: whiteImage(2, 2)} // 16 bytes -> 2 bytes LSB

	buffer := make([]byte, 2)
	n, err := chunk.Read(buffer)
	require.NoError(t, err)

	assert.Equal(t, 2, n)
	for _, b := range buffer {
		assert.EqualValues(t, 255, b)
	}
}

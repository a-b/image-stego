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

// emptyImage creates an RGBA image with 2x2 Pixels where all pixels are black. The underlying
// Pix byte array contains 2x2 x 4 (RGBA) = 16 entries.
func emptyImage(w, h int) *image.RGBA {
	return image.NewRGBA(image.Rect(0, 0, w, h))
}

func TestWrite_EmptyInput(t *testing.T) {

	chunk := Chunk{RGBA: emptyImage(2, 2)}

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

	chunk := Chunk{RGBA: emptyImage(2, 2)}

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

	chunk := Chunk{RGBA: emptyImage(2, 2)}

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

	chunk := Chunk{RGBA: emptyImage(2, 2)}

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

	chunk := Chunk{RGBA: emptyImage(2, 2)}

	n, err := chunk.Write([]byte{ones, ones, ones})
	assert.EqualError(t, err, io.EOF.Error())

	assert.Equal(t, 2, n)
	assert.Equal(t, 16, chunk.off)

	// Test expected bit representation
	assert.EqualValues(t, 1, chunk.Pix[15])
}

func TestWrite_PartialByteWritten(t *testing.T) {

	chunk := Chunk{RGBA: emptyImage(1, 3)} // 12 bytes

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

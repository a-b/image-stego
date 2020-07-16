package main

import (
	"fmt"
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
		img.Pix[i] = ones
	}
	return img
}

func TestChunk_MaxPayloadSize1(t *testing.T) {
	tests := []struct {
		width  int
		height int
		want   int
	}{
		{2, 2, 2},
		{1, 3, 1},
		{100, 100, 5000},
	}
	for _, tt := range tests {
		name := fmt.Sprintf("An image of size %d x %d can hold %d bytes", tt.width, tt.height, tt.want)
		t.Run(name, func(t *testing.T) {
			c := &Chunk{RGBA: whiteImage(tt.width, tt.height)}
			got := c.MaxPayloadSize()
			assert.Equal(t, tt.want, got, "MaxPayloadSize() = %v, want %v", got, tt.want)
		})
	}
}

func TestChunk_WriteEmptyInput(t *testing.T) {

	chunk := Chunk{RGBA: blackImage(2, 2)}

	n, err := chunk.Write([]byte{})
	require.NoError(t, err)

	assert.Equal(t, 0, n)
	assert.Equal(t, 0, chunk.wOff)

	// Test expected bit representation
	for _, p := range chunk.Pix {
		assert.EqualValues(t, 0, p)
	}
}

func TestChunk_WriteSetAllBitsToOne(t *testing.T) {

	chunk := Chunk{RGBA: blackImage(2, 2)}

	n, err := chunk.Write([]byte{ones, ones})
	require.NoError(t, err)

	assert.Equal(t, 2, n)
	assert.Equal(t, 16, chunk.wOff)

	// Test expected bit representation
	for _, p := range chunk.Pix {
		assert.EqualValues(t, 1, p)
	}
}

func TestChunk_WriteSetAllBitsToOneWithBreak(t *testing.T) {

	chunk := Chunk{RGBA: blackImage(2, 2)}

	n, err := chunk.Write([]byte{ones})
	require.NoError(t, err)

	assert.Equal(t, 1, n)
	assert.Equal(t, 8, chunk.wOff)

	n, err = chunk.Write([]byte{ones})
	require.NoError(t, err)

	assert.Equal(t, 1, n)
	assert.Equal(t, 16, chunk.wOff)

	// Test expected bit representation
	for _, p := range chunk.Pix {
		assert.EqualValues(t, 1, p)
	}
}

func TestChunk_WriteSetMixedBits(t *testing.T) {

	chunk := Chunk{RGBA: blackImage(2, 2)}

	n, err := chunk.Write([]byte{0b11110000, 0b00001111})
	require.NoError(t, err)

	assert.Equal(t, 2, n)
	assert.Equal(t, 16, chunk.wOff)

	// Test expected bit representation
	assert.EqualValues(t, 1, chunk.Pix[0])
	assert.EqualValues(t, 1, chunk.Pix[3])
	assert.EqualValues(t, 0, chunk.Pix[4])
	assert.EqualValues(t, 0, chunk.Pix[11])
	assert.EqualValues(t, 1, chunk.Pix[12])
}

func TestChunk_WriteMoreThanPossible(t *testing.T) {

	chunk := Chunk{RGBA: blackImage(2, 2)}

	n, err := chunk.Write([]byte{ones, ones, ones})
	assert.EqualError(t, err, io.EOF.Error())

	assert.Equal(t, 2, n)
	assert.Equal(t, 16, chunk.wOff)

	// Test expected bit representation
	assert.EqualValues(t, 1, chunk.Pix[15])
}

func TestChunk_WritePartialByteWritten(t *testing.T) {

	chunk := Chunk{RGBA: blackImage(1, 3)} // 12 bytes

	n, err := chunk.Write([]byte{ones, ones})
	assert.EqualError(t, err, io.EOF.Error())

	assert.Equal(t, 1, n)
	assert.Equal(t, 8, chunk.wOff)

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
	assert.Equal(t, 16, chunk.rOff)
	for _, b := range buffer {
		assert.EqualValues(t, ones, b)
	}
}

func TestRead_SmallerReadBuffer(t *testing.T) {
	chunk := Chunk{RGBA: whiteImage(2, 2)} // 16 bytes -> 2 bytes LSB

	buffer := make([]byte, 1)
	n, err := chunk.Read(buffer)
	require.NoError(t, err)

	assert.Equal(t, 1, n)
	assert.Equal(t, 8, chunk.rOff)
	for _, b := range buffer {
		assert.EqualValues(t, ones, b)
	}
}

func TestRead_LargerReadBuffer(t *testing.T) {
	chunk := Chunk{RGBA: whiteImage(2, 2)} // 16 bytes -> 2 bytes LSB

	buffer := make([]byte, 3)
	n, err := chunk.Read(buffer)
	require.EqualError(t, err, io.EOF.Error())

	assert.Equal(t, 2, n)
	assert.Equal(t, 16, chunk.rOff)
	assert.EqualValues(t, ones, buffer[0])
	assert.EqualValues(t, ones, buffer[0])
}

func TestRead_PartialReadBuffer(t *testing.T) {
	chunk := Chunk{RGBA: whiteImage(1, 3)} // 12 bytes -> 1.5 bytes LSB

	buffer := make([]byte, 2)
	n, err := chunk.Read(buffer)
	require.EqualError(t, err, io.EOF.Error())

	assert.Equal(t, 1, n)
	assert.Equal(t, 8, chunk.rOff)
	assert.EqualValues(t, ones, buffer[0])
	assert.EqualValues(t, zeroes, buffer[1])
}

func TestReadWrite(t *testing.T) {
	payload := []byte{42, 24}
	chunk := Chunk{RGBA: whiteImage(2, 2)} // 16 bytes -> 2 bytes LSB

	n, err := chunk.Write(payload)
	require.NoError(t, err)
	assert.Equal(t, 2, n)

	parsed := make([]byte, 2)
	n, err = chunk.Read(parsed)
	require.NoError(t, err)
	assert.Equal(t, 2, n)

	assert.EqualValues(t, 42, parsed[0])
	assert.EqualValues(t, 24, parsed[1])
}

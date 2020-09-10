package chunk

import (
	"fmt"
	"image"
	"io"
	"math/rand"
	"testing"

	"dennis-tra/image-stego/pkg/bit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestBitsPerPixelInRange(t *testing.T) {
	assert.GreaterOrEqual(t, BitsPerPixel, 1)
	assert.LessOrEqual(t, BitsPerPixel, 4)
}

func TestChunk_PixelCount(t *testing.T) {
	width := rand.Int() % 100
	height := rand.Int() % 100
	chunk := Chunk{RGBA: blackImage(width, height)}
	assert.Equal(t, width*height, chunk.PixelCount())
}

func TestChunk_LSBCount(t *testing.T) {
	chunk := Chunk{RGBA: blackImage(5, 5)}
	assert.Equal(t, 5*5*BitsPerPixel, chunk.LSBCount())
}

func TestChunk_MaxPayloadSize1(t *testing.T) {
	tests := []struct {
		width  int
		height int
	}{
		{2, 2},
		{1, 3},
		{100, 100},
	}
	for _, tt := range tests {
		want := tt.width * tt.height * BitsPerPixel / 8
		name := fmt.Sprintf("An image of size %d x %d can hold %d bytes", tt.width, tt.height, want)
		t.Run(name, func(t *testing.T) {
			c := &Chunk{RGBA: whiteImage(tt.width, tt.height)}
			got := c.MaxPayloadSize()
			assert.Equal(t, want, got, "MaxPayloadSize() = %v, want %v", got, want)
		})
	}
}

func TestChunk_WriteEmptyInput(t *testing.T) {

	chunk := Chunk{RGBA: blackImage(2, 2)}

	n, err := chunk.Write([]byte{})
	require.NoError(t, err)

	assert.Equal(t, 0, n)

	// Test expected bit representation
	for _, p := range chunk.Pix {
		assert.EqualValues(t, 0, p)
	}
}

func TestChunk_WriteSetAllBitsToOne(t *testing.T) {

	chunk := Chunk{RGBA: blackImage(2, 2)}

	n, err := chunk.Write([]byte{ones})
	require.NoError(t, err)

	assert.Equal(t, 1, n)

	// Test expected bit representation
	for i, p := range chunk.Pix {
		if i >= 8 {
			break
		}
		if (i+1)%4 == 0 && i != 0 {
			assert.EqualValues(t, 0, p)
		} else {
			assert.EqualValues(t, 1, p)
		}
	}
}

func TestChunk_WriteSetMixedBits(t *testing.T) {

	chunk := Chunk{RGBA: blackImage(3, 2)}

	n, err := chunk.Write([]byte{0b11110000, 0b00001111})
	require.NoError(t, err)

	assert.Equal(t, 2, n)

	// Test expected bit representation
	expects := []PixExpect{
		{0, 1},
		{1, 1},
		{2, 1},
		{3, 0},
		{4, 1},
		{5, 0},
		//
		{15, 0},
		{16, 1},
		{17, 1},
		{18, 1},
		{19, 0},
		{20, 1},
	}
	assertPixExpect(t, chunk, expects)
}

func TestChunk_WriteMoreThanPossible(t *testing.T) {

	chunk := Chunk{RGBA: blackImage(3, 2)}

	n, err := chunk.Write([]byte{ones, ones, ones})
	assert.EqualError(t, err, io.EOF.Error())

	assert.Equal(t, 2, n)

	// Test expected bit representation
	assert.EqualValues(t, 1, chunk.Pix[20])
}

func TestChunk_WritePartialByteWritten(t *testing.T) {

	chunk := Chunk{RGBA: blackImage(1, 3)} // 12 bytes

	n, err := chunk.Write([]byte{ones, ones})
	assert.EqualError(t, err, io.EOF.Error())

	assert.Equal(t, 1, n)

	// Test expected bit representation
	expects := []PixExpect{
		{0, 1},
		{1, 1},
		{2, 1},
		{3, 0},
		{4, 1},
		{5, 1},
		{6, 1},
		{7, 0},
		{8, 1},
		{9, 1},
		{10, 0},
		{11, 0},
	}
	assertPixExpect(t, chunk, expects)
}

func TestRead_MatchingLength(t *testing.T) {
	chunk := Chunk{RGBA: whiteImage(4, 6)} // 24 pixel -> 24*3=72 available LSBs -> 72/8 = 9 bytes

	buffer := make([]byte, 9)
	n, err := chunk.Read(buffer)
	require.NoError(t, err)

	assert.Equal(t, 9, n)
	for _, b := range buffer {
		assert.EqualValues(t, ones, b)
	}
}

func TestRead_SmallerReadBuffer(t *testing.T) {
	chunk := Chunk{RGBA: whiteImage(2, 3)} // 6 Pixel -> 6*3=18 available LSBs -> 18/8 = 2.25 bytes

	buffer := make([]byte, 1)
	n, err := chunk.Read(buffer)
	require.NoError(t, err)

	assert.Equal(t, 1, n)
	for _, b := range buffer {
		assert.EqualValues(t, ones, b)
	}
}

func TestRead_LargerReadBuffer(t *testing.T) {
	chunk := Chunk{RGBA: whiteImage(2, 3)} // 6 Pixel -> 6*3=18 available LSBs -> 18/8 = 2.25 bytes

	buffer := make([]byte, 3)
	n, err := chunk.Read(buffer)
	require.EqualError(t, err, io.EOF.Error())

	assert.Equal(t, 2, n)
	assert.EqualValues(t, ones, buffer[0])
	assert.EqualValues(t, ones, buffer[0])
}

func TestRead_PartialReadBuffer(t *testing.T) {
	chunk := Chunk{RGBA: whiteImage(1, 3)} // 3 Pixel -> 3*3=9 available LSBs -> 9/8 = 1 byte

	buffer := make([]byte, 2)
	n, err := chunk.Read(buffer)
	require.EqualError(t, err, io.EOF.Error())

	assert.Equal(t, 1, n)
	assert.EqualValues(t, ones, buffer[0])
	assert.EqualValues(t, zeroes, buffer[1])
}

func TestReadWrite(t *testing.T) {
	payload := []byte{42, 24}
	chunk := Chunk{RGBA: whiteImage(2, 3)} // 6 Pixel -> 6*3=18 available LSBs -> 18/8 = 2.25 byte

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


// PixExpect holds an index and expected bit value.
type PixExpect struct {
	idx int
	bit int
}

func assertPixExpect(t *testing.T, chunk Chunk, expects []PixExpect) {
	for _, e := range expects {
		got := 0
		if bit.GetLSB(chunk.Pix[e.idx]) {
			got = 1
		}
		assert.EqualValues(t, e.bit, got, "Pixel at idx %d has val %d, want: %d", e.idx, chunk.Pix[e.idx], e.bit)
	}
}

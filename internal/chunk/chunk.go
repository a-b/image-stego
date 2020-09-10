package chunk

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"image"
	"io"

	"dennis-tra/image-stego/pkg/bit"
	"github.com/cbergoon/merkletree"
	"github.com/icza/bitio"
)

const (
	// BitsPerPixel says how many colors should be used for LSB encoding.
	// 1 - only R, 2 - R and G, 3 - R, G and B, 4 - R, G, B and A
	BitsPerPixel = 3
)

// Chunk is a wrapper around an image.RGBA struct that keeps track of
// the read and written bits to the LSBs of the image.RGBA
type Chunk struct {
	*image.RGBA
}

// MaxPayloadSize returns the maximum number of bytes that can be written to this chunk
func (c *Chunk) MaxPayloadSize() int {
	return c.LSBCount() / 8
}

// Width is a short hand to return the width in pixels of the chunk
func (c *Chunk) Width() int {
	return c.Bounds().Size().X
}

// Height is a short hand to return the height in pixels of the chunk
func (c *Chunk) Height() int {
	return c.Bounds().Size().Y
}

// PixelCount returns the total number of pixels
func (c *Chunk) PixelCount() int {
	return len(c.Pix) / 4
}

// LSBCount returns the total number of least significant bits available for encoding a message.
// Only the RGB values are considered not the A.
func (c *Chunk) LSBCount() int {
	return c.PixelCount() * BitsPerPixel
}

// CalculateHash calculates the SHA256 hash of the most significant bits. The least
// significant bit (LSB) is not considered in the hash generation as it is used to
// store the derived Merkle leaves.
// Note: From an implementation point of view the LSB is actually considered but
// always overwritten by a 0.
// This method is necessary to conform to the merkletree.Content interface.
func (c *Chunk) CalculateHash() ([]byte, error) {

	h := sha256.New()

	for x := 0; x < c.Width(); x++ {
		for y := 0; y < c.Height(); y++ {

			rgba := c.RGBA.RGBAAt(x, y)

			byt := []byte{
				bit.WithLSB(rgba.R, false),
				bit.WithLSB(rgba.G, false),
				bit.WithLSB(rgba.B, false),
			}
			if _, err := h.Write(byt); err != nil {
				return nil, err
			}
		}
	}

	return h.Sum(nil), nil
}

// Write writes the given bytes to the least significant bits of the chunk.
// It returns the number of bytes written from p and an error if one occurred.
// Consult the io.Writer documentation for the intended behaviour of the function.
// A byte from p is either written completely or not at all to the least significant bits.
// Subsequent calls to write will continue were the last write left off.
func (c *Chunk) Write(p []byte) (n int, err error) {
	r := bitio.NewReader(bytes.NewBuffer(p))

	for i := 0; i < len(p); i++ {

		bitOff := i*8

		// Stop early if there is not enough LSB space left
		if bitOff+7 >= len(c.Pix)-len(c.Pix)/4 {
			return n, io.EOF
		}

		for j := 0; j < 8; j++ {

			bitVal, err := r.ReadBool()
			if err != nil {
				return n, err
			}

			c.Pix[bitOff+j+(bitOff+j)/3] = bit.WithLSB(c.Pix[bitOff+j+(bitOff+j)/3], bitVal)
		}

		// As one byte was written increment the counter
		n += 1
	}

	return n, nil
}

// Read reads the amount of bytes given in p from the LSBs of the image chunk.
func (c *Chunk) Read(p []byte) (n int, err error) {

	b := bytes.NewBuffer(p)
	w := bitio.NewWriter(b)
	b.Reset()
	defer w.Close()

	for i := 0; i < len(p); i++ {

		// calculate current read bit offset: static read offset from potential last run plus idx-var times bits in a byte
		bitOff := i*8

		// Stop early if there are not enough LSBs left
		if bitOff+8+(bitOff+8)/BitsPerPixel > len(c.Pix) {
			return n, io.EOF
		}

		for j := 0; j < 8; j++ {
			err := w.WriteBool(bit.GetLSB(c.Pix[bitOff+j+(bitOff+j)/BitsPerPixel]))
			if err != nil {
				return i, err
			}
		}

		// As one whole byte was read increment the counter
		n += 1
	}

	return n, err
}

// Equals tests for equality of two Contents
func (c *Chunk) Equals(o merkletree.Content) (bool, error) {

	oc, ok := o.(*Chunk) // other chunk
	if !ok {
		return false, errors.New("invalid type casting")
	}

	if oc.Width() != c.Width() || oc.Height() != c.Height() {
		return false, nil
	}

	for x := 0; x < c.Width(); x++ {
		for y := 0; y < c.Height(); y++ {

			thisColor := c.RGBAAt(c.Bounds().Min.X+x, c.Bounds().Min.Y+y)
			otherColor := oc.RGBAAt(c.Bounds().Min.X+x, c.Bounds().Min.Y+y)

			if bit.WithLSB(thisColor.R, false) != bit.WithLSB(otherColor.R, false) {
				return false, nil
			}

			if bit.WithLSB(thisColor.G, false) != bit.WithLSB(otherColor.G, false) {
				return false, nil
			}

			if bit.WithLSB(thisColor.B, false) != bit.WithLSB(otherColor.B, false) {
				return false, nil
			}

			if bit.WithLSB(thisColor.A, false) != bit.WithLSB(otherColor.A, false) {
				return false, nil
			}
		}
	}

	return true, nil
}

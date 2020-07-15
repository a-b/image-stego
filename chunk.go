package main

import (
	"crypto/sha256"
	"errors"
	"github.com/cbergoon/merkletree"
	"image"
	"math/bits"
)

type Chunk struct {
	*image.RGBA
	n int // written bytes
}

// Width is a short hand to return the width in pixels of the chunk
func (c *Chunk) Width() int {
	return c.Bounds().Size().X
}

// Height is a short hand to return the height in pixels of the chunk
func (c *Chunk) Height() int {
	return c.Bounds().Size().Y
}

// CalculateHash calculates the SHA256 hash of the most significant bits. The least
// significant bit (LSB) is not considered in the hash generation as it
// is used to store the derived Merkle leaves.
// Actually the LSB is considered but always overwritten by a 0.
func (c *Chunk) CalculateHash() ([]byte, error) {

	h := sha256.New()

	for x := 0; x < c.Width(); x++ {
		for y := 0; y < c.Height(); y++ {
			// TODO: Evaluate if this should be changed to a bit reader,
			// that really just considers the most significant bits
			// instead of zero-ing the LSBs
			color := c.RGBAAt(x, y)
			byt := []byte{
				WithLSB(color.R, false),
				WithLSB(color.G, false),
				WithLSB(color.B, false),
			}
			if _, err := h.Write(byt); err != nil {
				return nil, err
			}
		}
	}

	return h.Sum(nil), nil
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

			thisColor := c.RGBAAt(x, y)
			otherColor := oc.RGBAAt(x, y)

			if WithLSB(thisColor.R, false) != WithLSB(otherColor.R, false) {
				return false, nil
			}

			if WithLSB(thisColor.G, false) != WithLSB(otherColor.G, false) {
				return false, nil
			}

			if WithLSB(thisColor.B, false) != WithLSB(otherColor.B, false) {
				return false, nil
			}

			if WithLSB(thisColor.A, false) != WithLSB(otherColor.A, false) {
				return false, nil
			}
		}
	}

	return true, nil
}

// Write writes the given bytes to the least significant bits of the chunk.
func (c *Chunk) Write(p []byte) (n int, err error) {
	for payloadByteIdx, payloadByte := range p {
		for payloadBitIdx := uint8(0); payloadBitIdx < 8; payloadBitIdx++ {
			lsb := BitAtIdx(payloadByte, payloadBitIdx)
			pixIdx := n + payloadByteIdx
			c.Pix[pixIdx] = WithLSB(c.Pix[pixIdx], lsb)
		}
		n++
	}
	return n, nil
}

// BitAtIdx returns true if the bit at the given index is 1 and false if it is 0
func BitAtIdx(b byte, i uint8) bool {
	return bits.OnesCount8(byte(1<<i)&b) > 0
}

func (c *Chunk) ReadLSB() ([]byte, error) {
	byteArr := []byte{}
	i := 0
	for i < len(c.Pix) {
		newByte := uint8(0)
		for j := 0; j < 8; j++ {
			if GetLSB(c.Pix[i+j]) {
				newByte = newByte | byte(1<<j)
			}
		}
		byteArr = append(byteArr, newByte)
		i += 8
	}
	return byteArr, nil
}

func (c *Chunk) LSBHash() ([]byte, error) {

	lsbByteArr, err := c.ReadLSB()
	if err != nil {
		return nil, err
	}

	sides := []bool{}
	paths := [][]byte{}
	for i, b := range lsbByteArr {
		if i%32 == 0 {
			if b == 1 {
				sides = append(sides, true)
			} else {
				sides = append(sides, false)
			}
			paths = append(paths, []byte{})
			continue
		}
		paths[(i-1)/32] = append(paths[(i-1)/32], b)
	}

	prevHash, err := c.CalculateHash()
	if err != nil {
		return nil, err
	}

	for i, side := range sides {
		hsh := sha256.New()
		w := []byte{}
		if side { // right
			w = append(w, prevHash...)
			w = append(w, paths[i]...)
		} else { // left
			w = append(w, paths[i]...)
			w = append(w, prevHash...)
		}
		hsh.Write(w)
		prevHash = hsh.Sum(nil)
	}

	return prevHash, nil
}

package main

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"github.com/cbergoon/merkletree"
	"github.com/icza/bitio"
	"image"
	"io"
	"math/bits"
)

type Chunk struct {
	*image.RGBA
	// offset of written bits
	off int
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

// Write writes the given bytes to the least significant bits of the chunk. It returns the number of bytes written from p
// A byte from p is either written completely or not at all.
func (c *Chunk) Write(p []byte) (n int, err error) {
	r := bitio.NewReader(bytes.NewBuffer(p))

	defer func() {
		c.off += n * 8
	}()

	for i := 0; i < len(p); i++ {

		buffer := make([]byte, 8)
		bitOff := c.off+i * 8

		if bitOff+7 >= len(c.Pix) {
			return n, io.EOF
		}

		for j := range buffer {

			bit, err := r.ReadBool()
			if err != nil {
				return n, err
			}

			buffer[j] = WithLSB(c.Pix[bitOff+j], bit)
		}

		for j, b := range buffer {
			c.Pix[bitOff+j] = b
		}

		n += 1
	}
	return n, nil
}

// BitAtIdx returns true if the bit at the given index is 1 and false if it is 0
func BitAtIdx(b byte, i uint8) bool {
	return bits.OnesCount8(byte(1<<i)&b) > 0
}

func (c *Chunk) Read(p []byte) (int, error) {
	b := bytes.NewBuffer(p)
	w := bitio.NewWriter(b)
	i := 0
	for i < b.Len() {
		p := c.Pix[i]
		err := w.WriteBool(GetLSB(p))
		if err != nil {
			return i, err
		}
		i++
	}
	return i, nil
}

func (c *Chunk) LSBHash() ([]byte, error) {
	//
	//lsbByteArr, err := c.ReadLSB()
	//if err != nil {
	//	return nil, err
	//}
	//
	//sides := []bool{}
	//paths := [][]byte{}
	//for i, b := range lsbByteArr {
	//	if i%32 == 0 {
	//		if b == 1 {
	//			sides = append(sides, true)
	//		} else {
	//			sides = append(sides, false)
	//		}
	//		paths = append(paths, []byte{})
	//		continue
	//	}
	//	paths[(i-1)/32] = append(paths[(i-1)/32], b)
	//}

	prevHash, err := c.CalculateHash()
	if err != nil {
		return nil, err
	}
	//
	//for i, side := range sides {
	//	hsh := sha256.New()
	//	w := []byte{}
	//	if side { // right
	//		w = append(w, prevHash...)
	//		w = append(w, paths[i]...)
	//	} else { // left
	//		w = append(w, paths[i]...)
	//		w = append(w, prevHash...)
	//	}
	//	hsh.Write(w)
	//	prevHash = hsh.Sum(nil)
	//}

	return prevHash, nil
}

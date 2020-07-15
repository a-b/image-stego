package main

import (
	"encoding/hex"
	"fmt"
	"github.com/cbergoon/merkletree"
	"image"
	"image/draw"
	"image/png"
	_ "image/png"
	"log"
	"os"
)

func main() {
	filename := "rect.png"
	fmt.Println("Operating on", filename)

	fmt.Println("Opening...")
	reader, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Decoding...")
	m, _, err := image.Decode(reader)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Closing...")
	err = reader.Close()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Image file dimensions", m.Bounds().Max)
	totalPixels := m.Bounds().Size().X * m.Bounds().Size().Y
	fmt.Println("Image total pixels", totalPixels)
	lsbBits := totalPixels * 3
	lsbBytes := float32(lsbBits) / 8.0
	fmt.Println("LSB Bits", lsbBits)
	fmt.Println("LSB Bytes", lsbBytes)
	//maxNumHashes := lsbBits / 256
	//fmt.Println("Maximum number of hashes", maxNumHashes)
	//
	//numOfChunks := 1.0
	//
	//for numOfChunks*math.Log2(numOfChunks) < float64(maxNumHashes)  {
	//	numOfChunks++
	//}
	//fmt.Println("Total number of chunks", numOfChunks)
	//

	rgbaImage := imageToRGBA(m)

	chunkCountX := 4
	chunkCountY := 2
	chunkWidth := m.Bounds().Size().X / chunkCountX
	chunkHeight := m.Bounds().Size().Y / chunkCountY

	fmt.Println("Chunk counts: ", chunkCountX, chunkCountY)
	fmt.Println("Chunk dimensions: ", chunkWidth, chunkHeight)

	fmt.Println("Start building merkle tree...")
	var list []merkletree.Content
	for cx := 0; cx < chunkCountX; cx++ {
		for cy := 0; cy < chunkCountY; cy++ {
			fmt.Println("-- Checking chunk at", cx, cy)
			chunk := &Chunk{
				RGBA: image.NewRGBA(image.Rect(0, 0, chunkWidth, chunkHeight)),
				n:    0,
			}
			for x := 0; x < chunkWidth; x++ {
				for y := 0; y < chunkHeight; y++ {
					color := rgbaImage.RGBAAt(cx*chunkWidth+x, cy*chunkHeight+y)
					chunk.Set(x, y, color)
				}
			}
			hash, _ := chunk.CalculateHash()
			fmt.Println("Chunk hash", hex.EncodeToString(hash))
			list = append(list, chunk)
		}
	}

	// Create a new Merkle Tree from the list of Content
	t, err := merkletree.NewTree(list)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Start writing image...")
	destImg := image.NewRGBA(image.Rect(0, 0, m.Bounds().Dx(), m.Bounds().Dy()))
	for cx := 0; cx < chunkCountX; cx++ {
		for cy := 0; cy < chunkCountY; cy++ {
			chunk := list[cx*chunkCountY+cy].(*Chunk)
			paths, sides, err := t.GetMerklePath(chunk)
			if err != nil {
				log.Fatal(err)
			}

			for i, path := range paths {
				side := uint8(sides[i])
				if _, err := chunk.Write(append([]byte{side}, path...)); err != nil {
					log.Fatal(err)
				}
			}
			draw.Draw(destImg, chunk.Bounds().Add(image.Pt(cx*chunkWidth, cy*chunkHeight)), chunk, image.Point{}, draw.Src)
		}
	}

	f, err := os.Create("outimage.png")
	if err != nil {
		// Handle error
	}
	defer f.Close()

	err = png.Encode(f, destImg)
	if err != nil {
		log.Fatal(err)
	}

	//// Get the Merkle Root of the tree
	//mr := t.MerkleRoot()
	//fmt.Println("MerkleRoot", hex.EncodeToString(mr))
	//
	////Verify the entire tree (hashes for each node) is valid
	//vt, err := t.VerifyTree()
	//if err != nil {
	//	log.Fatal(err)
	//}
	//fmt.Println("Verify Tree: ", vt)
	//
	//fmt.Println("GetMerklePath")
	//
	//vChunk := list[4]
	//paths, indexes, err := t.GetMerklePath(vChunk)
	//prevHash, _ := vChunk.CalculateHash()
	//for i, side := range indexes {
	//	hsh := sha256.New()
	//	w := []byte{}
	//	if side == 1 { // right
	//		w = append(w, prevHash...)
	//		w = append(w, paths[i]...)
	//	} else if side == 0 { // left
	//		w = append(w, paths[i]...)
	//		w = append(w, prevHash...)
	//	}
	//	hsh.Write(w)
	//	prevHash = hsh.Sum(nil)
	//}
	//
	//fmt.Println("MerkleRoot Calculated", hex.EncodeToString(prevHash))

}

//type TestContent struct {
//	hashString string
//}
//
////CalculateHash hashes the values of a TestContent
//func (t TestContent) CalculateHash() ([]byte, error) {
//	return hex.DecodeString(t.hashString)
//}
//
////Equals tests for equality of two Contents
//func (t TestContent) Equals(other merkletree.Content) (bool, error) {
//	return t.hashString == other.(TestContent).hashString, nil
//}

// WithLSB returns the given byte with the least significant bit (LSB) set to
// the given bit value, while true means 1 and false means 0.
func WithLSB(b byte, bit bool) byte {
	if bit {
		return b | 1
	} else {
		return b & 0xFE
	}
}

// imageToRGBA converts image.Image to image.RGBA
func imageToRGBA(src image.Image) *image.RGBA {
	fmt.Println("Converting image to RGBA")
	b := src.Bounds()
	m := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(m, m.Bounds(), src, b.Min, draw.Src)
	return m
}

package main

import (
	"flag"
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
)

func main() {
	decodePtr := flag.Bool("d", false, "Whether to decode the given image file")


	// Once all flags are declared, call `flag.Parse()`
	// to execute the command-line parsing.
	flag.Parse()

	fmt.Println("decode:", *decodePtr)
	fmt.Println("files:", flag.Args())


	for _, filename := range flag.Args() {

		file, err := os.Open(filename)
		if err != nil {
			log.Fatal(err)
		}

		img, _, err := image.Decode(file)
		if err != nil {
			log.Fatal(err)
		}

		err = file.Close()
		if err != nil {
			log.Fatal(err)
		}

		rgba := imageToRGBA(img)
		if *decodePtr {
			decode(rgba)
		} else {
			encode(rgba)
		}
	}

	//if len(os.Args) > 1 && os.Args[1] == "-d" {
	//	Decode()
	//	return
	//}
	//
	//filename := "rect.png"
	//fmt.Println("Operating on", filename)
	//
	//fmt.Println("Opening...")
	//reader, err := os.Open(filename)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//fmt.Println("Decoding...")
	//m, _, err := image.Decode(reader)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//fmt.Println("Closing...")
	//err = reader.Close()
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//fmt.Println("Image file dimensions", m.Bounds().Max)
	//totalPixels := m.Bounds().Size().X * m.Bounds().Size().Y
	//fmt.Println("Image total pixels", totalPixels)
	//lsbBits := totalPixels * 3
	//lsbBytes := float32(lsbBits) / 8.0
	//fmt.Println("LSB Bits", lsbBits)
	//fmt.Println("LSB Bytes", lsbBytes)
	////maxNumHashes := lsbBits / 256
	////fmt.Println("Maximum number of hashes", maxNumHashes)
	////
	////numOfChunks := 1.0
	////
	////for numOfChunks*math.Log2(numOfChunks) < float64(maxNumHashes)  {
	////	numOfChunks++
	////}
	////fmt.Println("Total number of chunks", numOfChunks)
	////
	//
	//rgbaImage := imageToRGBA(m)
	//
	//chunkCountX := 2
	//chunkCountY := 2
	//chunkWidth := m.Bounds().Size().X / chunkCountX
	//chunkHeight := m.Bounds().Size().Y / chunkCountY
	//
	//fmt.Println("Chunk counts: ", chunkCountX, chunkCountY)
	//fmt.Println("Chunk dimensions: ", chunkWidth, chunkHeight)
	//
	//fmt.Println("Start building merkle tree...")
	//var list []merkletree.Content
	//for cx := 0; cx < chunkCountX; cx++ {
	//	for cy := 0; cy < chunkCountY; cy++ {
	//		fmt.Println("-- Checking chunk at", cx, cy)
	//		chunk := &Chunk{
	//			RGBA: image.NewRGBA(image.Rect(0, 0, chunkWidth, chunkHeight)),
	//		}
	//		for x := 0; x < chunkWidth; x++ {
	//			for y := 0; y < chunkHeight; y++ {
	//				color := rgbaImage.RGBAAt(cx*chunkWidth+x, cy*chunkHeight+y)
	//				chunk.Set(x, y, color)
	//			}
	//		}
	//		hash, _ := chunk.CalculateHash()
	//		fmt.Println("Chunk hash", hex.EncodeToString(hash))
	//		list = append(list, chunk)
	//	}
	//}
	//
	//// Create a new Merkle Tree from the list of Content
	//t, err := merkletree.NewTree(list)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//fmt.Println("MERKLE ROOT", hex.EncodeToString(t.MerkleRoot()))
	//
	//fmt.Println("")
	//fmt.Println("")
	//
	//destImg := image.NewRGBA(image.Rect(0, 0, m.Bounds().Dx(), m.Bounds().Dy()))
	//for cx := 0; cx < chunkCountX; cx++ {
	//	for cy := 0; cy < chunkCountY; cy++ {
	//		chunk := list[cx*chunkCountY+cy].(*Chunk)
	//		prevHsh, _ := chunk.CalculateHash()
	//		fmt.Println("-- Writing chunk at", cx, cy, hex.EncodeToString(prevHsh))
	//		paths, sides, err := t.GetMerklePath(chunk)
	//		if err != nil {
	//			log.Fatal(err)
	//		}
	//
	//		writeBuffer := []byte{}
	//		writeBuffer = append(writeBuffer, uint8(len(paths)))
	//		for i, path := range paths {
	//			side := uint8(sides[i])
	//			writeBuffer = append(writeBuffer, side)
	//			writeBuffer = append(writeBuffer, path...)
	//		}
	//
	//		w := new(bytes.Buffer)
	//		err = steganography.EncodeRGBA(w, chunk.RGBA, writeBuffer)
	//		_ = err
	//
	//		i, _, _ := image.Decode(w)
	//		a := imageToRGBA(i)
	//		draw.Draw(destImg, a.Bounds().Add(image.Pt(cx*chunkWidth, cy*chunkHeight)), a, image.Point{}, draw.Src)
	//	}
	//}
	//
	//f, err := os.Create("outimage.png")
	//if err != nil {
	//	// Handle error
	//}
	//defer f.Close()
	//
	//err = png.Encode(f, destImg)
	//if err != nil {
	//	log.Fatal(err)
	//}

}

func decode(filename *image.RGBA) {

}

// imageToRGBA converts image.Image to image.RGBA
func imageToRGBA(src image.Image) *image.RGBA {
	b := src.Bounds()
	rgba := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(rgba, rgba.Bounds(), src, b.Min, draw.Src)
	return rgba
}

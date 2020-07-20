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

// imageToRGBA converts image.Image to image.RGBA
func imageToRGBA(src image.Image) *image.RGBA {
	b := src.Bounds()
	rgba := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(rgba, rgba.Bounds(), src, b.Min, draw.Src)
	return rgba
}

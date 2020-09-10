package chunk

import (
	"crypto/sha256"
	"dennis-tra/image-stego/internal/utils"
	"encoding/hex"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
	"path"
)

func Decode(filepath string) error {

	log.Println("Opening image:", filepath)
	rgba, err := OpenImageFile(filepath)
	if err != nil {
		return err
	}

	log.Println("Calculating bounds...")
	bounds := CalculateChunkBounds(rgba)

	log.Println("Calculating Merkle tree roots for every chunk...")
	roots := map[string][]Index{}

	for cx, boundRow := range bounds {
		for cy, bound := range boundRow {

			chunk := &Chunk{RGBA: image.NewRGBA(bound)}

			for x := 0; x < bound.Dx(); x++ {
				for y := 0; y < bound.Dy(); y++ {
					original := bound.Min.Add(image.Pt(x, y))
					chunk.Set(original.X, original.Y, rgba.RGBAAt(original.X, original.Y))
				}
			}

			pathCount := make([]byte, 1)
			_, err := chunk.Read(pathCount)
			if err != nil {
				return err
			}

			chunkHash, _ := chunk.CalculateHash()
			prevHash := chunkHash
			for i := 0; i < int(pathCount[0]); i++ {
				side := make([]byte, 1)
				data := make([]byte, 32)

				_, err := chunk.Read(side)
				if err != nil {
					//return err
				}

				_, err = chunk.Read(data)
				if err != nil {
					//return err
				}

				hsh := sha256.New()

				if side[0] == 0 {
					prevHash = append(data, prevHash...)
				} else {
					prevHash = append(prevHash, data...)
				}

				hsh.Write(prevHash)
				prevHash = hsh.Sum(nil)
			}

			merkleRoot := hex.EncodeToString(prevHash)

			_, exists := roots[merkleRoot]
			if !exists {
				roots[merkleRoot] = []Index{}
			}
			roots[merkleRoot] = append(roots[merkleRoot], Index{cx: cx, cy: cy})
		}
	}

	m := 0
	canonicalMerkleRoot := ""
	for merkleRoot, indices := range roots {
		if len(indices) > m {
			m = len(indices)
			canonicalMerkleRoot = merkleRoot
		}
	}

	if len(roots) == 1 {
		log.Println("This image has not been tampered with. All chunks have the same Merkle Root:")
		log.Println("\t", canonicalMerkleRoot)
		return nil
	}

	log.Println("Found multiple Merkle Roots. This image has been tampered with. RootHashes:")
	log.Println("Count\tRoot")
	for root, indexes := range roots {
		log.Printf("%07d\t%s\n", len(indexes), root)
	}

	log.Println("Drawing overlay image of altered regions...")
	overlayImage := utils.ImageToRGBA(rgba.SubImage(rgba.Bounds()))
	for merkleRoot, indices := range roots {
		for _, idx := range indices {
			bound := bounds[idx.cx][idx.cy]
			chunk := &Chunk{RGBA: image.NewRGBA(bound)}

			if merkleRoot != canonicalMerkleRoot {
				draw.DrawMask(
					overlayImage,
					chunk.Bounds(),
					&image.Uniform{C: color.RGBA{R: 255, A: 255}},
					image.Point{},
					&image.Uniform{C: color.RGBA{R: 255, G: 255, B: 255, A: 80}},
					image.Point{},
					draw.Over,
				)
			}
		}
	}

	overlay := path.Join(path.Dir(filepath), utils.SetExtension(path.Base(filepath), ".overlay.png"))
	log.Println("Saving overlay image:", overlay)
	out, err := os.Create(overlay)
	if err != nil {
		return err
	}
	defer out.Close()

	err = png.Encode(out, overlayImage)
	if err != nil {
		return err
	}
	return nil
}

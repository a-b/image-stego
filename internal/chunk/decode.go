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

			chunk := &Chunk{
				RGBA: utils.ImageToRGBA(rgba.SubImage(bound)),
			}

			// First byte contains the number of hashes in this chunk (called paths in the merkletree package)
			pathCount := make([]byte, 1)
			_, err := chunk.Read(pathCount)
			if err != nil {
				return err
			}

			chunkHash, _ := chunk.CalculateHash()
			prevHash := chunkHash
			for i := 0; i < int(pathCount[0]); i++ {
				// The order in which the hashes should be concatenated to calculate the composite hash
				side := make([]byte, 1)

				// The hash data for the new composite hash
				data := make([]byte, 32)

				// EOFs can happen if pathCount is wrong due to image manipulation
				// of that specific chunk. pathCount could be way larger than
				// the maximum chunk payload, therefore an EOF can happen.
				_, err := chunk.Read(side)
				if err != nil {
					break
				}

				_, err = chunk.Read(data)
				if err != nil {
					break
				}

				hsh := sha256.New()

				if side[0] == 0 {
					prevHash = append(data, prevHash...)
				} else if side[0] == 1 {
					prevHash = append(prevHash, data...)
				} else {
					break
				}

				hsh.Write(prevHash)
				prevHash = hsh.Sum(nil)
			}

			merkleRoot := hex.EncodeToString(prevHash)

			// persist root hash
			_, exists := roots[merkleRoot]
			if !exists {
				roots[merkleRoot] = []Index{}
			}
			roots[merkleRoot] = append(roots[merkleRoot], Index{cx: cx, cy: cy})
		}
	}

	// Find the root hash that appeared multiple times
	m := 0
	canonicalMerkleRoot := ""
	for merkleRoot, indices := range roots {
		if len(indices) > m {
			m = len(indices)
			canonicalMerkleRoot = merkleRoot
		}
	}

	if len(roots) == 1 {
		log.Println("This image has not been tampered with. All chunks have the same Merkle Root:", canonicalMerkleRoot)
		return nil
	}

	log.Println("Found multiple Merkle Roots. This image has been tampered with! RootHashes:")

	log.Println("Count\tRoot")
	for root, indexes := range roots {
		log.Printf("%5d\t%s\n", len(indexes), root)
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

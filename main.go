package main

import (
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"github.com/jmhobbs/go-paa"
)

func main() {
	var (
		sourceDir = flag.String("source", "./layers", "Source directory containing PAA/PNG files")
		width     = flag.Int("width", 32, "How many tiles wide")
		height    = flag.Int("height", 32, "How many tiles high")
		tileSize  = flag.Int("size", 512, "Size of tiles")
		overlap   = flag.Int("overlap", 64, "Overlap size")
		output    = flag.String("output", "map.png", "Output file name")
	)
	flag.Parse()

	tileEffectiveSize := *tileSize - (*overlap * 2)
	finalSizeX := tileEffectiveSize * *width
	finalSizeY := tileEffectiveSize * *height

	dst := image.NewRGBA(image.Rect(0, 0, finalSizeX, finalSizeY))
	for y := 0; y < *height; y++ {
		for x := 0; x < *width; x++ {
			cmp, err := loadFile(*sourceDir, x, y)
			if err != nil {
				log.Fatalf("error loading PNG for tile (%d, %d): %v", x, y, err)
			}
			draw.Draw(
				dst,
				image.Rect(x*tileEffectiveSize, y*tileEffectiveSize, (x+1)*tileEffectiveSize, (y+1)*tileEffectiveSize),
				cmp,
				image.Point{*overlap, *overlap},
				draw.Src,
			)
		}
	}
	sink, err := os.Create(*output)
	if err != nil {
		log.Fatalf("error creating output file: %v", err)
	}
	defer func() {
		err := sink.Close()
		if err != nil {
			log.Printf("error closing output file: %v", err)
		}
	}()
	err = png.Encode(sink, dst)
	if err != nil {
		log.Fatalf("error writing output image: %v", err)
	}
}

func loadFile(dir string, x, y int) (image.Image, error) {
	img, err := loadPaa(dir, x, y)
	if err == nil {
		return img, nil
	}
	if !os.IsNotExist(err) {
		return nil, err
	}
	return loadPng(dir, x, y)
}

func loadPng(dir string, x, y int) (image.Image, error) {
	filename := fmt.Sprintf(filepath.Join(dir, "S_%03d_%03d_lco.png"), x, y)
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening input file %s: %w", filename, err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("error closing input file %s: %v", filename, err)
		}
	}()
	img, err := png.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("error decoding input file %s: %w", filename, err)
	}
	return img, nil
}

func loadPaa(dir string, x, y int) (image.Image, error) {
	filename := fmt.Sprintf(filepath.Join(dir, "S_%03d_%03d_lco.paa"), x, y)
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening input file %s: %w", filename, err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("error closing input file %s: %v", filename, err)
		}
	}()

	paaImg, err := paa.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("error decoding input file %s: %w", filename, err)
	}

	img, err := paaImg.Mipmaps[0].Image(f)
	if err != nil {
		return nil, fmt.Errorf("error decoding mipmap %s: %w", filename, err)
	}

	if paaImg.SWIZ != nil {
		img = paaImg.SWIZ.Unswizzle(img)
	}

	return img, nil
}

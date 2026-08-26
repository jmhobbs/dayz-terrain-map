package main

import (
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
)

func main() {
	var (
		sourceDir = flag.String("source", "./layers", "Source directory containing PNG files")
		width     = flag.Int("width", 16, "How many tiles wide")
		height    = flag.Int("height", 16, "How many tiles high")
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
			cmp := loadPng(*sourceDir, x, y)
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
		panic(err)
	}
	defer sink.Close()
	png.Encode(sink, dst)
}

func loadPng(dir string, x, y int) image.Image {
	f, err := os.Open(fmt.Sprintf(filepath.Join(dir, "S_%03d_%03d_lco.png"), x, y))
	if err != nil {
		panic(err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		panic(err)
	}
	return img
}

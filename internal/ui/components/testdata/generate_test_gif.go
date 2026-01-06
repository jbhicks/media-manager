//go:build ignore

// This file generates a test GIF file for unit tests.
// Run with: go run generate_test_gif.go
package main

import (
	"image"
	"image/color"
	"image/gif"
	"os"
)

func main() {
	// Create a simple 10x10 animated GIF with 2 frames
	images := make([]*image.Paletted, 2)
	delays := make([]int, 2)

	palette := []color.Color{
		color.RGBA{255, 0, 0, 255},   // Red
		color.RGBA{0, 0, 255, 255},   // Blue
		color.RGBA{255, 255, 255, 0}, // Transparent
	}

	for i := range images {
		images[i] = image.NewPaletted(image.Rect(0, 0, 10, 10), palette)
		for x := 0; x < 10; x++ {
			for y := 0; y < 10; y++ {
				if i == 0 {
					images[i].SetColorIndex(x, y, 0) // Red frame
				} else {
					images[i].SetColorIndex(x, y, 1) // Blue frame
				}
			}
		}
		delays[i] = 50 // 500ms delay
	}

	f, err := os.Create("test.gif")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	err = gif.EncodeAll(f, &gif.GIF{
		Image: images,
		Delay: delays,
	})
	if err != nil {
		panic(err)
	}

	println("Created test.gif")
}

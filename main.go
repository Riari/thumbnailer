package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/h2non/filetype"
	"github.com/muesli/smartcrop"
	"github.com/muesli/smartcrop/nfnt"
	"github.com/nfnt/resize"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func findCrop(img image.Image, w int, h int) image.Rectangle {
	analyzer := smartcrop.NewAnalyzer(nfnt.NewDefaultResizer())
	crop, err := analyzer.FindBestCrop(img, w, h)
	check(err)

	return crop
}

func cropAndScale(img image.Image, crop image.Rectangle, w uint, h uint) image.Image {
	type SubImager interface {
		SubImage(r image.Rectangle) image.Image
	}

	cropped := img.(SubImager).SubImage(crop)
	scaled := resize.Resize(w, h, cropped, resize.Lanczos3)

	return scaled
}

func imageProcessor(width int, height int, inputPath string, outputPath string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		b, err := ioutil.ReadFile(path)

		if filetype.IsImage(b) {
			kind, _ := filetype.Match(b)
			reader := bytes.NewReader(b)
			var img image.Image

			if kind.Extension == "gif" {
				img, _ = gif.Decode(reader)
			} else {
				img, _, _ = image.Decode(reader)
			}

			crop := findCrop(img, width, height)

			fmt.Printf("Selected crop: %+v\n", crop)

			scaledImg := cropAndScale(img, crop, uint(width), uint(height))

			ext := filepath.Ext(info.Name())
			name := strings.TrimSuffix(info.Name(), ext)
			outputName := strings.Join([]string{name, "_thumb.jpg"}, "")

			var pathToCreate = filepath.Join(outputPath, outputName)
			out, err := os.Create(pathToCreate)
			check(err)

			fmt.Printf("Writing %s\n", pathToCreate)

			err = jpeg.Encode(out, scaledImg, &jpeg.Options{Quality: 80})
			check(err)
		}

		return nil
	}
}

func main() {
	cwd, _ := os.Getwd()

	var width = flag.Int("w", 350, "width in px")
	var height = flag.Int("h", 197, "height in px")
	var inputPath = flag.String("in", filepath.Join(cwd, "/input"), "input path")
	var outputPath = flag.String("out", filepath.Join(cwd, "/output"), "output path")
	flag.Parse()

	err := filepath.Walk(*inputPath, imageProcessor(*width, *height, *inputPath, *outputPath))

	check(err)
}

package gatomemes

import (
	"bytes"
	"errors"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
)

var imgbytes []byte

func encodeImage(src image.Image) {
	buffer := new(bytes.Buffer)
	err := png.Encode(buffer, src)
	checkError("encode png: ", err)
	imgbytes = buffer.Bytes()
}

func decodeImage(fileType string, reader io.Reader) (draw.Image, error) {
	switch fileType {
	case "image/png":
		img, err := png.Decode(reader)
		if err != nil {
			return nil, err
		}
		if img, ok := img.(draw.Image); ok {
			return img, nil
		}
	case "image/jpeg":
		img, err := jpeg.Decode(reader)
		if err != nil {
			return nil, err
		}
		if img, ok := img.(*image.YCbCr); ok {
			return jpegToPng(img)
		}
	}
	return nil, errors.New("unsupported format")

}

func openLocalImage(path string) (draw.Image, error) {
	file, err := os.Open(path)
	// checkFatalError(err)
	if err != nil {
		return nil, err
	}

	switch filepath.Ext(path) {
	case ".png":
		return decodeImage("image/png", file)
	case ".jpg":
		return decodeImage("image/jpeg", file)
	}
	return nil, errors.New("unsupported format")
}

func jpegToPng(src *image.YCbCr) (draw.Image, error) {
	var out draw.Image
	out = image.NewNRGBA(src.Bounds())

	for y := 0; y < src.Bounds().Dy(); y++ {
		for x := 0; x < src.Bounds().Dx(); x++ {
			srcColor := src.At(x, y)
			out.Set(x, y, srcColor)
		}
	}
	return out, nil
}

func savePngOnDisk(img image.Image, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	if err = png.Encode(file, img); err != nil {
		return err
	}
	return nil

}

package gatomemes

import (
	"bytes"
	"image"
	"image/draw"
	"image/png"
	"os"
)

func encodePNG(src image.Image) ([]byte, error) {
	buffer := &bytes.Buffer{}
	err := png.Encode(buffer, src)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func decodeImage(data []byte, mimeType string) (draw.Image, error) {
	d, err := newDecoder(data, mimeType)
	if err != nil {
		return nil, err
	}

	img, err := d.decode()
	if err != nil {
		return nil, err
	}

	if img, ok := img.(draw.Image); ok {
		return img, nil
	}

	return nil, unsupportedError
}

func openLocalImage(path string) (draw.Image, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return decodeImage(file, "")
}

func saveImageOnDisk(img image.Image, path string) error {
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

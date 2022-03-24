package gatomemes

import (
	"bytes"
	"errors"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
)

var unsupportedError = errors.New("unsupported format")

type decoder interface {
	decode() (draw.Image, error)
}

type decoderPNG struct {
	r io.Reader
}

func (decoder *decoderPNG) decode() (draw.Image, error) {
	img, err := png.Decode(decoder.r)
	if err != nil {
		return nil, err
	}

	printColorReport(img)

	// TODO: fix png with indexed colors
	if img, ok := img.(draw.Image); ok {
		return img, nil
	}

	return nil, unsupportedError
}

type decoderJPEG struct {
	r io.Reader
}

func (decoder *decoderJPEG) decode() (draw.Image, error) {
	img, err := jpeg.Decode(decoder.r)
	if err != nil {
		return nil, err
	}

	printColorReport(img)

	if img, ok := img.(*image.YCbCr); ok {
		return convertToRGBA(img)
	}

	return nil, unsupportedError
}

func newDecoder(data []byte, mimeType string) (decoder, error) {
	reader := bytes.NewReader(data)

	if mimeType == "" {
		mimeType = http.DetectContentType(data)
	}

	switch mimeType {
	case "image/png":
		return &decoderPNG{r: reader}, nil
	case "image/jpeg":
		return &decoderJPEG{r: reader}, nil
	}

	return nil, unsupportedError
}

func printColorReport(img image.Image) {
	_, ok := img.(*image.Alpha)
	logger.Println("Alpha:   ", ok)
	_, ok = img.(*image.Alpha16)
	logger.Println("Alpha16: ", ok)
	_, ok = img.(*image.CMYK)
	logger.Println("CMYK:    ", ok)
	_, ok = img.(*image.Gray)
	logger.Println("Gray:    ", ok)
	_, ok = img.(*image.Gray16)
	logger.Println("Gray16:  ", ok)
	_, ok = img.(*image.NRGBA)
	logger.Println("NRGBA:   ", ok)
	_, ok = img.(*image.NRGBA64)
	logger.Println("NRGBA64: ", ok)
	_, ok = img.(*image.NYCbCrA)
	logger.Println("NYCbCrA: ", ok)
	_, ok = img.(*image.Paletted)
	logger.Println("Paletted:", ok)
	_, ok = img.(*image.RGBA)
	logger.Println("RGBA:    ", ok)
	_, ok = img.(*image.RGBA64)
	logger.Println("RGBA64:  ", ok)
	_, ok = img.(*image.YCbCr)
	logger.Println("YCbCr:   ", ok)
}

func convertToRGBA(src image.Image) (draw.Image, error) {
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

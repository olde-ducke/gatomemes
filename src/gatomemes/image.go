package gatomemes

import (
	"bytes"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"os"
)

var imgbytes []byte

func encodeImage(dst image.Image) {
	// convert to png//jpeg
	buffer := new(bytes.Buffer)
	err := png.Encode(buffer, dst) //, &jpeg.Options{Quality: 98})
	checkError("encode png: ", err)
	imgbytes = buffer.Bytes()
}

func savePngOnDisk(img image.Image, filename string) {
	f, err := os.Create(filename)
	checkError("os.Create: ", err)
	defer f.Close()
	err = png.Encode(f, img)
	checkError("encode png: ", err)
}

func openPngFromDisk(filename string) {
	var err error
	imgbytes, err = ioutil.ReadFile(filename)
	checkError("openPngFromDisk: ", err)
}

// gets response fro GetNew(), converts it to png and returns result to
// font drawing
func convertRespons(responseBody io.ReadCloser) {
	log.Println("conseguir un nuevo gatito ")
	src := convertJpegToPng(responseBody)
	drawTextOnImage(src.(draw.Image))
}

func convertJpegToPng(data io.ReadCloser) (src image.Image) {
	// decoding jpeg to image.Image
	src, err := jpeg.Decode(data)
	checkError("decode jpeg: ", err)

	// converting jpeg to png, because you can't draw in YCbCr space
	// TODO: find a way to avoid this
	buffer := new(bytes.Buffer)
	err = png.Encode(buffer, src)
	checkError("encode png :", err)

	// decoding png back to image.Image again
	src, err = png.Decode(buffer)
	checkError("decode png :", err)

	return src
}

package gatomemes

import (
	"image"
	"image/draw"
	"io/ioutil"
	"math"
	"os"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type textData struct {
	dbAccessFunc                    func() (lines [2]string)
	context                         freetype.Context
	size                            fixed.Int26_6
	firstLineWidth, secondLineWidth fixed.Int26_6
	text                            [2]string
}

var text textData
var textFont *truetype.Font

// magic numbers for text offset from border, without offset some symbols will reach Max.Y
// bounds of image and some symbol points get cutoff by clipping bounds on X axis
// represented as Int26_6
const textXOffset, textYOffset = 20 << 6, 7 << 6

// ignores dpi, it's 72.0 by default, otherwise the same conversion as in freetype
// for local use without need to call to context for dpi value
func float2fixed(x float64) fixed.Int26_6 {
	return fixed.Int26_6(x * 64.0)
}

func fitTextOnImage(dst draw.Image) {

	lines := text.dbAccessFunc()
	dstWidth := fixed.I(dst.Bounds().Max.X)
	srcHeight := fixed.I(dst.Bounds().Max.Y)

	text.context = *freetype.NewContext()
	text.context.SetFont(textFont)
	text.context.SetDPI(72.0) // default is 72.0, btw
	text.context.SetHinting(font.HintingFull)

	// janky font size adjustement
	//*********************************************
	var line string
	if measureString(textFont, 32.0, lines[0]) > measureString(textFont, 32.0, lines[1]) {
		line = lines[0]
	} else {
		line = lines[1]
	}

	// TODO: add magic scale number instead of several magic numbers
	size := float64(dstWidth.Floor() / 16) // magic number
	for i, delta := 0, size; ; i++ {
		x := dstWidth - measureString(textFont, float2fixed(size), line) - textXOffset
		if x < fixed.I(0) {
			size -= delta
		} else {
			size += delta
		}
		delta /= 2
		//log.Printf("%v %v %v\n", x, size, i)
		if (x >= 3<<6 && x <= 4<<6) || delta < 0.001 {
			break
		}
	}

	text.size = float2fixed(math.Round(size))
	text.context.SetDst(dst)
	text.context.SetClip(dst.Bounds())
	text.context.SetFontSize(size)
	text.firstLineWidth = measureString(textFont, text.size, lines[0])
	text.secondLineWidth = measureString(textFont, text.size, lines[1])

	// 'E' is randomly chosen, all symbols in used font have same metrics
	glyphHeight := textFont.VMetric(text.size, textFont.Index('E')).AdvanceHeight
	glyphTopBearing := textFont.VMetric(text.size, textFont.Index('E')).TopSideBearing

	//*********************************************

	//hacky outline, set black source image and draw two lines 4 times with offsets
	outlineOffset := float2fixed(size / 32.0) // magic number
	text.context.SetSrc(image.Black)
	//TODO: for loop and dot position for font drawing
	_, err := text.context.DrawString(lines[0], fixed.Point26_6{
		X: (dstWidth-text.firstLineWidth)/2 - outlineOffset,
		Y: glyphHeight - outlineOffset + textYOffset,
	})
	checkError("DrawString: ", err)
	_, err = text.context.DrawString(lines[1], fixed.Point26_6{
		X: (dstWidth-text.secondLineWidth)/2 - outlineOffset,
		Y: srcHeight - glyphHeight + glyphTopBearing - outlineOffset - textYOffset,
	})
	checkError("DrawString: ", err)
	_, err = text.context.DrawString(lines[0], fixed.Point26_6{
		X: (dstWidth-text.firstLineWidth)/2 + outlineOffset,
		Y: glyphHeight - outlineOffset + textYOffset,
	})
	checkError("DrawString: ", err)
	_, err = text.context.DrawString(lines[1], fixed.Point26_6{
		X: (dstWidth-text.secondLineWidth)/2 + outlineOffset,
		Y: srcHeight - glyphHeight + glyphTopBearing - outlineOffset - textYOffset,
	})
	checkError("DrawString: ", err)
	_, err = text.context.DrawString(lines[0], fixed.Point26_6{
		X: (dstWidth-text.firstLineWidth)/2 - outlineOffset,
		Y: glyphHeight + outlineOffset + textYOffset,
	})
	checkError("DrawString: ", err)
	_, err = text.context.DrawString(lines[1], fixed.Point26_6{
		X: (dstWidth-text.secondLineWidth)/2 - outlineOffset,
		Y: srcHeight - glyphHeight + glyphTopBearing + outlineOffset - textYOffset,
	})
	checkError("DrawString: ", err)
	_, err = text.context.DrawString(lines[0], fixed.Point26_6{
		X: (dstWidth-text.firstLineWidth)/2 + outlineOffset,
		Y: glyphHeight + outlineOffset + textYOffset,
	})
	checkError("DrawString: ", err)
	_, err = text.context.DrawString(lines[1], fixed.Point26_6{
		X: (dstWidth-text.secondLineWidth)/2 + outlineOffset,
		Y: srcHeight - glyphHeight + glyphTopBearing + outlineOffset - textYOffset,
	})
	checkError("DrawString: ", err)

	// draw two actual lines in white over black borders
	text.context.SetSrc(image.White)
	_, err = text.context.DrawString(lines[0], fixed.Point26_6{
		X: (dstWidth - text.firstLineWidth) / 2,
		Y: glyphHeight + textYOffset,
	})
	checkError("DrawString: ", err)
	_, err = text.context.DrawString(lines[1], fixed.Point26_6{
		X: (dstWidth - text.secondLineWidth) / 2,
		Y: srcHeight - glyphHeight + glyphTopBearing - textYOffset,
	})
	checkError("DrawString: ", err)

	//*********************************************
	encodeImage(dst)
}

// ported from opentype sources
func measureString(f *truetype.Font, scale fixed.Int26_6, s string) (advance fixed.Int26_6) {
	// TODO: check what happens on empty string
	// write as a method?
	var currC, prevC truetype.Index
	for i, c := range s {
		currC = f.Index(c)
		if i > 0 {
			advance += f.Kern(scale, prevC, currC)
		}
		a := f.HMetric(scale, currC).AdvanceWidth
		advance += a
		prevC = currC
	}
	return advance
}

func getSymbolPositions(f *truetype.Font, scale fixed.Int26_6, s string) (positions []fixed.Int26_6) {
	// TODO: same as above, but returns each letter position
	var advance fixed.Int26_6
	var currC, prevC truetype.Index
	for i, c := range s {
		currC = f.Index(c)
		if i > 0 {
			advance += f.Kern(scale, prevC, currC)
		}
		advance += f.HMetric(scale, currC).AdvanceWidth
		positions = append(positions, advance)
		prevC = currC
	}
	return positions
}

func init() {
	// opening font
	var err error
	fontbytes, err := ioutil.ReadFile(os.Getenv("PROJECTFONT"))
	checkError("fontbytes:", err)

	textFont, err = freetype.ParseFont(fontbytes)
	checkError("ParseFont: ", err)
}

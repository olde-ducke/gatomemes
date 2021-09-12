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
	font                            *truetype.Font
	context                         freetype.Context
	xoffset, yoffset, size          fixed.Int26_6
	firstLineWidth, secondLineWidth fixed.Int26_6
	text                            [2]string
}

// magic numbers for text offset from border, without offset some symbols will reach Max.Y
// bounds of image and some symbol points get cutoff by clipping bounds on X axis
var memeText = textData{
	xoffset: fixed.I(20),
	yoffset: fixed.I(7),
}

// ignores dpi, it's 72.0 by default, otherwise the same conversion as in freetype
// for local use without need to call to context for dpi value
func float2fixed(x float64) fixed.Int26_6 {
	return fixed.Int26_6(x * 64.0)
}

func drawTextOnImage(dst draw.Image) {

	lines := memeText.dbAccessFunc()
	dstWidth := fixed.I(dst.Bounds().Max.X)
	srcHeight := fixed.I(dst.Bounds().Max.Y)

	// janky font size adjustement
	//*********************************************
	line := lines[0]
	if len(lines[1]) > len(lines[0]) {
		line = lines[1]
	}
	size := float64(dstWidth.Floor() / 8) // magic number
	for i := 0; ; i++ {
		x := dstWidth - measureString(memeText.font, float2fixed(size), line) - memeText.xoffset
		if x < fixed.I(0) {
			size -= 0.2 // yet another magic number
		} else {
			size += 0.2
		}
		//log.Printf("%v %v %v\n", x, size, i)
		if (x <= memeText.xoffset+fixed.I(1) || i > 100) && x >= fixed.I(0) {
			break
		}
	}
	memeText.size = float2fixed(math.Floor(size))
	memeText.context.SetDst(dst)
	memeText.context.SetClip(dst.Bounds())
	memeText.context.SetFontSize(size)
	memeText.firstLineWidth = measureString(memeText.font, memeText.size, lines[0])
	memeText.secondLineWidth = measureString(memeText.font, memeText.size, lines[1])

	// 'E' is randomly chosen, all symbols in used font have same metrics
	glyph_height := memeText.font.VMetric(memeText.size, memeText.font.Index('E')).AdvanceHeight
	glyph_topside := memeText.font.VMetric(memeText.size, memeText.font.Index('E')).TopSideBearing

	//*********************************************

	//hacky outline, set black source image and draw two lines 4 times with offsets
	outlineOffset := float2fixed(size / 32.0) // magic number
	memeText.context.SetSrc(image.Black)
	//TODO: for loop and dot position for font drawing
	_, err := memeText.context.DrawString(lines[0], fixed.Point26_6{
		X: (dstWidth-memeText.firstLineWidth)/2 - outlineOffset,
		Y: glyph_height - outlineOffset + memeText.yoffset,
	})
	checkError("DrawString: ", err)
	_, err = memeText.context.DrawString(lines[1], fixed.Point26_6{
		X: (dstWidth-memeText.secondLineWidth)/2 - outlineOffset,
		Y: srcHeight - glyph_height + glyph_topside - outlineOffset - memeText.yoffset,
	})
	checkError("DrawString: ", err)
	_, err = memeText.context.DrawString(lines[0], fixed.Point26_6{
		X: (dstWidth-memeText.firstLineWidth)/2 + outlineOffset,
		Y: glyph_height - outlineOffset + memeText.yoffset,
	})
	checkError("DrawString: ", err)
	_, err = memeText.context.DrawString(lines[1], fixed.Point26_6{
		X: (dstWidth-memeText.secondLineWidth)/2 + outlineOffset,
		Y: srcHeight - glyph_height + glyph_topside - outlineOffset - memeText.yoffset,
	})
	checkError("DrawString: ", err)
	_, err = memeText.context.DrawString(lines[0], fixed.Point26_6{
		X: (dstWidth-memeText.firstLineWidth)/2 - outlineOffset,
		Y: glyph_height + outlineOffset + memeText.yoffset,
	})
	checkError("DrawString: ", err)
	_, err = memeText.context.DrawString(lines[1], fixed.Point26_6{
		X: (dstWidth-memeText.secondLineWidth)/2 - outlineOffset,
		Y: srcHeight - glyph_height + glyph_topside + outlineOffset - memeText.yoffset,
	})
	checkError("DrawString: ", err)
	_, err = memeText.context.DrawString(lines[0], fixed.Point26_6{
		X: (dstWidth-memeText.firstLineWidth)/2 + outlineOffset,
		Y: glyph_height + outlineOffset + memeText.yoffset,
	})
	checkError("DrawString: ", err)
	_, err = memeText.context.DrawString(lines[1], fixed.Point26_6{
		X: (dstWidth-memeText.secondLineWidth)/2 + outlineOffset,
		Y: srcHeight - glyph_height + glyph_topside + outlineOffset - memeText.yoffset,
	})
	checkError("DrawString: ", err)

	// draw two actual lines in white over black borders
	memeText.context.SetSrc(image.White)
	_, err = memeText.context.DrawString(lines[0], fixed.Point26_6{
		X: (dstWidth - memeText.firstLineWidth) / 2,
		Y: glyph_height + memeText.yoffset,
	})
	checkError("DrawString: ", err)
	_, err = memeText.context.DrawString(lines[1], fixed.Point26_6{
		X: (dstWidth - memeText.secondLineWidth) / 2,
		Y: srcHeight - glyph_height + glyph_topside - memeText.yoffset,
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

	memeText.font, err = freetype.ParseFont(fontbytes)
	checkError("ParseFont: ", err)

	// setting everything that we can for font drawing, for later use
	memeText.context = *freetype.NewContext()
	memeText.context.SetFont(memeText.font)
	memeText.context.SetDPI(72.0) // default is 72.0, btw
	memeText.context.SetHinting(font.HintingFull)
	memeText.dbAccessFunc = getRandomLines
}

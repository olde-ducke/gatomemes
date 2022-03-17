package gatomemes

import (
	"errors"
	"image"
	"image/color"
	"image/draw"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang/freetype"
	"github.com/golang/freetype/raster"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

var fonts []*truetype.Font

const (
	top = iota
	bottom
	middle
)

type textDrawer struct {
	str          string
	font         *truetype.Font
	fontScale    fixed.Int26_6
	fontColor    color.Color
	dpi          float64
	hinting      font.Hinting
	outlineWidth fixed.Int26_6
	outlineColor color.Color

	//dst  imageDraw
	// mask *image.RGBA

	// painter *raster.RGBAPainter
	// rast    *raster.Rasterizer

	// glyphXPositions []fixed.Int26_6
	glyphCache map[rune]*truetype.GlyphBuf
}

type options struct {
	fontIndex    int
	fontSize     float64
	fontColor    color.Color
	dpi          float64
	hinting      font.Hinting
	outlineWidth float64
	outlineColor color.Color
	distort      bool
}

func (drawer *textDrawer) pointToFixed(f float64) fixed.Int26_6 {
	return fixed.Int26_6(f * float64(drawer.dpi) * (64.0 / 72.0))
}

func (drawer *textDrawer) changeSize(size float64) {
	drawer.fontScale = drawer.pointToFixed(size)

	for _, curr := range drawer.str {
		var buf truetype.GlyphBuf
		err := buf.Load(drawer.font, drawer.fontScale,
			drawer.font.Index(curr), drawer.hinting,
		)

		// in case of an error report error and
		// replace glyph with nothing
		if err != nil {
			log.Println(err)
		}
		drawer.glyphCache[curr] = &buf
	}
}

func newDrawer(text string, opt *options) (*textDrawer, error) {
	if text == "" {
		return nil, errors.New("empty string, nothing to draw")
	}

	if opt == nil {
		opt = &options{
			fontIndex:    0,
			fontSize:     64.0,
			fontColor:    color.White,
			dpi:          72.0,
			hinting:      font.HintingNone,
			outlineWidth: 0.0,
			outlineColor: color.Black,
		}
	} else {
		// clamp options that could break everything
		// TODO: clamp on other side to?
		if opt.fontSize <= 0.0 {
			opt.fontSize = 64.0
		}

		if opt.dpi <= 0.0 {
			opt.dpi = 72.0
		}

		if opt.outlineWidth < 0.0 {
			opt.outlineWidth = 0.0
		}

		if opt.fontColor == nil {
			opt.fontColor = color.White
		}

		if opt.outlineColor == nil {
			opt.outlineColor = color.Black
		}
	}

	drawer := &textDrawer{
		str:          text,
		font:         fonts[opt.fontIndex%len(fonts)],
		fontColor:    opt.fontColor,
		dpi:          opt.dpi,
		hinting:      opt.hinting,
		outlineColor: opt.outlineColor,
	}

	drawer.outlineWidth = drawer.pointToFixed(opt.outlineWidth)
	drawer.glyphCache = make(map[rune]*truetype.GlyphBuf, 0)
	drawer.changeSize(opt.fontSize)

	return drawer, nil
}

// func (drawer *textDrawer) setDst(dst draw.Image) {
//  // TODO: for now draws on empty image,
//  // needs color model converter
//  drawer.dst = dst
//  drawer.mask = image.NewRGBA(dst.Bounds)
//  drawer.painter = raster.NewRGBAPainter(drawer.mask)
//  drawer.rast = raster.NewRasterizer(mask.Bounds().Dx(), mask.Bounds().Dy())
// }

func (drawer *textDrawer) getGlyphPositions() (positions []fixed.Int26_6) {
	var advance fixed.Int26_6
	var hasPrevious bool
	var prev rune
	font := drawer.font
	for _, curr := range drawer.str {
		if hasPrevious {
			kern := font.Kern(drawer.fontScale, font.Index(prev), font.Index(curr))
			kern = (kern + 32) &^ 63
			advance += kern
		}
		advance += drawer.glyphCache[curr].AdvanceWidth
		positions = append(positions, advance)
		prev, hasPrevious = curr, true

	}
	return positions
}

// FIXME: not perfect, pretty sure it sometimes
// gives wrong coords, ported from opentype, which
// uses different approach compared to freetype
func (drawer *textDrawer) measureString() (width fixed.Int26_6, height fixed.Int26_6) {
	var advance, maxY fixed.Int26_6
	var hasPrevious bool
	var prev rune
	font := drawer.font
	for _, curr := range drawer.str {
		if hasPrevious {
			kern := font.Kern(drawer.fontScale, font.Index(prev), font.Index(curr))
			kern = (kern + 32) &^ 63
			advance += kern
		}

		if y := drawer.glyphCache[curr].Bounds.Max.Y; y > maxY {
			maxY = y
		}

		advance += drawer.glyphCache[curr].AdvanceWidth
		prev, hasPrevious = curr, true

	}
	return advance, maxY
}

// NOTE: most of the low-level font handling is ported
// from truetype, which doesn't support text outlining
// (not complete port as stated in doc) and doesn't
// expose data needed for doing that manually
func drawGlyph(str string, opt *options, dst draw.Image, vAlignment int) {
	dstWidth, dstHeight := dst.Bounds().Dx(), dst.Bounds().Dy()
	opt.fontSize = float64(dstWidth / 16) // magic number

	drawer, err := newDrawer(str, opt)
	if err != nil {
		log.Println(err)
		return
	}
	// drawer.setDst(dst)
	textXOffset := drawer.outlineWidth
	if textXOffset < 640 {
		textXOffset = 640 // 10 in fixed.Int26_6
	}

	// FIXME: as broken as ever
	var width, height fixed.Int26_6
	for i, delta := 0, opt.fontSize; ; i++ {
		width, height = drawer.measureString()
		x := fixed.I(dstWidth) - width - textXOffset
		if x <= 0 {
			opt.fontSize -= delta
		} else {
			opt.fontSize += delta
		}
		drawer.changeSize(opt.fontSize)
		delta /= 2
		// log.Printf("%v %v %v\n", x, opt.fontSize, i)
		if /*(x <= 1<<6) ||*/ delta < 0.001 {
			break
		}
	}

	startCoords := fixed.Point26_6{
		X: (fixed.I(dstWidth) - width) / 2,
	}

	switch vAlignment {
	case top:
		startCoords.Y = textXOffset + height + textXOffset/2
	case middle:
		startCoords.Y = (fixed.I(dstHeight) + height) / 2
	default:
		startCoords.Y = fixed.I(dstHeight) - textXOffset - textXOffset/2
	}

	coords := drawer.getGlyphPositions()
	// log.Println(drawer.measureString())
	// test := image.NewRGBA(dst.Bounds())
	// cntx := freetype.NewContext()
	// cntx.SetDPI(opt.dpi)
	// cntx.SetFont(fonts[opt.fontIndex%len(fonts)])
	// cntx.SetClip(test.Bounds())
	// cntx.SetFontSize(opt.fontSize)
	// cntx.SetHinting(opt.hinting)
	// cntx.SetSrc(image.White)
	// cntx.SetDst(test)
	// cntx.DrawString(str, startCoords)
	// savePngOnDisk(test, "img/test2.png")

	offset := startCoords
	i := 0
	//  drawer.rast.Clear()
	mask := image.NewRGBA(dst.Bounds())
	painter := raster.NewRGBAPainter(mask)
	rast := raster.NewRasterizer(mask.Bounds().Dx(), mask.Bounds().Dy())
	rast.Clear()

	var paths []raster.Path
	for _, r := range drawer.str {
		// add basic distortion
		if opt.distort {
			for i, _ := range drawer.glyphCache[r].Points {
				drawer.glyphCache[r].Points[i].X += drawer.pointToFixed(noise())
				drawer.glyphCache[r].Points[i].Y += drawer.pointToFixed(noise())
			}
		}
		// first pass: create path to draw and draw outline
		var path raster.Path
		var e0 int
		// .Ends contain indices of ending points of contours,
		// basically add every contour, that glyph has, to path
		for _, e1 := range drawer.glyphCache[r].Ends {
			path = append(path, createPath(
				drawer.glyphCache[r].Points[e0:e1],
				offset.X,
				offset.Y)...)
			e0 = e1
		}
		// draw outline with strokes if needed
		if drawer.outlineWidth > 0.0 {
			// drawer.painter.SetColor(randomRGB())
			rast.AddStroke(path, drawer.outlineWidth, nil, nil)
		}
		offset.X = startCoords.X + coords[i]
		paths = append(paths, path)
		i++
	}
	// not sure if resetting back to false is needed,
	// non-zero winding prevents nulling of areas
	// with self-intersecting, by default set to false,
	// which doesn't break regular contour drawing
	painter.SetColor(drawer.outlineColor)
	rast.UseNonZeroWinding = true
	rast.Rasterize(painter)
	rast.UseNonZeroWinding = false

	// second pass draw glyph itself
	rast.Clear()
	offset = startCoords
	for i, _ = range paths {

		// drawer.painter.SetColor(drawer.fontColor)
		// drawer.rast.Clear()
		rast.AddPath(paths[i])
		// drawer.rast.Rasterize(drawer.painter)
		offset.X = startCoords.X + coords[i]
	}
	painter.SetColor(drawer.fontColor)
	rast.Rasterize(painter)
	draw.DrawMask(dst, dst.Bounds(), mask, image.Point{0, 0},
		mask, image.Point{0, 0}, draw.Over)

}

func createPath(ps []truetype.Point, dx, dy fixed.Int26_6) (path raster.Path) {
	if len(ps) == 0 {
		return path
	}
	start := fixed.Point26_6{
		X: dx + ps[0].X,
		Y: dy - ps[0].Y,
	}
	others := []truetype.Point(nil)
	if ps[0].Flags&0x01 != 0 {
		others = ps[1:]
	} else {
		last := fixed.Point26_6{
			X: dx + ps[len(ps)-1].X,
			Y: dy - ps[len(ps)-1].Y,
		}
		if ps[len(ps)-1].Flags&0x01 != 0 {
			start = last
			others = ps[:len(ps)-1]
		} else {
			start = fixed.Point26_6{
				X: (start.X + last.X) / 2,
				Y: (start.Y + last.Y) / 2,
			}
			others = ps
		}
	}
	path.Start(start)
	q0, on0 := start, true
	for _, p := range others {
		q := fixed.Point26_6{
			X: dx + p.X,
			Y: dy - p.Y,
		}
		on := p.Flags&0x01 != 0
		if on {
			if on0 {
				path.Add1(q)
			} else {
				path.Add2(q0, q)
			}
		} else {
			if on0 {
				// No-op.
			} else {
				mid := fixed.Point26_6{
					X: (q0.X + q.X) / 2,
					Y: (q0.Y + q.Y) / 2,
				}
				path.Add2(q0, mid)
			}
		}
		q0, on0 = q, on
	}
	if on0 {
		path.Add1(start)
	} else {
		path.Add2(q0, start)
	}
	return path
}

func randomRGBA() color.RGBA {
	return color.RGBA{
		R: randomUint8(),
		G: randomUint8(),
		B: randomUint8(),
		A: randomUint8(),
	}
}

func randomRGB() color.RGBA {
	return color.RGBA{
		R: randomUint8(),
		G: randomUint8(),
		B: randomUint8(),
		A: 255,
	}
}

func extractColor(input string) color.RGBA {
	n, err := strconv.ParseInt(input, 16, 64)
	if err != nil {
		log.Println(err)
		return color.RGBA{0, 0, 0, 255}
	}
	return color.RGBA{
		R: uint8(n >> 24 & 0xff),
		G: uint8(n >> 16 & 0xff),
		B: uint8(n >> 8 & 0xff),
		A: uint8(n & 0xff),
	}
}

func randomUint8() uint8 {
	rand.Seed(time.Now().UnixNano())
	return uint8(rand.Intn(256))
}

func noise() float64 {
	rand.Seed(time.Now().UnixNano())
	return (rand.Float64() - 0.5) * 2.0
}

func init() {
	filenames := strings.Fields(os.Getenv("PROJECTFONT"))
	if len(filenames) == 0 {
		log.Fatal("no fonts found")
	}
	for _, filename := range filenames {
		fontBytes, err := ioutil.ReadFile(filename)
		checkError("init: ", err)
		f, err := freetype.ParseFont(fontBytes)
		checkError("init: ", err)
		fonts = append(fonts, f)
	}
}

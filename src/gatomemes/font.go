package gatomemes

import (
	"image"
	"image/color"
	"image/draw"
	"io/ioutil"
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
	middle
	bottom
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
	glyphCache   map[rune]*truetype.GlyphBuf
}

type Options struct {
	FontIndex      int64
	FontScale      int64
	FontColor      string
	OutlineColor   string
	OutlineScale   int64
	DisableOutline bool
	Distort        bool

	fontSize     float64
	dpi          float64
	hinting      font.Hinting
	outlineWidth float64
}

var defaultOptions = &Options{
	FontIndex:    0,
	FontScale:    2,
	FontColor:    "ffffff",
	OutlineColor: "000000",
	OutlineScale: 1,

	fontSize:     64.0,
	dpi:          72.0,
	hinting:      font.HintingNone,
	outlineWidth: 10.0,
}

func GetFontNames() []string {
	names := make([]string, 0, len(fonts))
	for _, f := range fonts {
		names = append(names, f.Name(truetype.NameIDPostscriptName))
	}
	return names
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
			logger.Println(err)
		}
		drawer.glyphCache[curr] = &buf
	}
}

func filterOptions(opt *Options) *Options {
	if opt == nil {
		return defaultOptions
	}

	opt.FontIndex = opt.FontIndex % int64(len(fonts))

	if opt.FontScale < 1 || opt.FontScale > 4 {
		opt.FontScale = 2
	}

	if opt.FontColor == "" {
		opt.FontColor = "ffffff"
	}

	if opt.OutlineColor == "" {
		opt.OutlineColor = "000000"
	}

	if opt.OutlineScale < 1 || opt.OutlineScale > 4 {
		opt.OutlineScale = 1
	}

	if opt.dpi <= 0.0 || opt.dpi >= 96.0 {
		opt.dpi = 72.0
	}

	return opt
}

func newDrawer(text string, opt *Options) *textDrawer {
	// logger.Println("in newDrawer     font size:", opt.fontSize)
	// logger.Println("in newDrawer outline width:", opt.outlineWidth)
	// FIXME: clamp relative to image size
	if opt.fontSize <= 0.0 {
		opt.fontSize = 64.0
	}

	if opt.fontSize >= 256 {
		opt.fontSize = 256.0
	}

	if opt.outlineWidth <= 0.0 {
		opt.outlineWidth = 10.0
	}

	if opt.outlineWidth >= 120.0 {
		opt.outlineWidth = 120.0
	}

	drawer := &textDrawer{
		str:          text,
		font:         fonts[opt.FontIndex],
		fontColor:    extractColor(opt.FontColor),
		dpi:          opt.dpi,
		hinting:      opt.hinting,
		outlineColor: extractColor(opt.OutlineColor),
	}

	drawer.outlineWidth = drawer.pointToFixed(opt.outlineWidth)
	drawer.glyphCache = make(map[rune]*truetype.GlyphBuf, 0)
	drawer.changeSize(opt.fontSize)
	// logger.Println("intermediate     font size:", drawer.fontScale)
	// logger.Println("intermediate outline width:", drawer.outlineWidth)

	return drawer
}

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
func drawGlyphs(str string, opt *Options, dst draw.Image, vAlignment int) {
	if str == "" {
		// logger.Println("empty string, nothing to draw")
		return
	}

	if dst == nil {
		logger.Println("no image provided")
		return
	}

	opt = filterOptions(opt)

	dstWidth, dstHeight := dst.Bounds().Dx(), dst.Bounds().Dy()
	opt.fontSize = float64(dstWidth*int(opt.FontScale)) / float64(64)        // magic number
	opt.outlineWidth = float64(dstWidth*int(opt.OutlineScale)) / float64(48) // same
	// logger.Println("initial          font size:", opt.fontSize)
	// logger.Println("initial      outline width:", opt.outlineWidth)

	drawer := newDrawer(str, opt)

	textXOffset := drawer.outlineWidth
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
		// logger.Printf("%v %v %v\n", x, opt.fontSize, i)
		if /*(x <= 1<<6) ||*/ delta < 0.001 {
			break
		}
	}
	// logger.Println("final            font size:", drawer.fontScale)
	// logger.Println("final        outline width:", drawer.outlineWidth)

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
	offset := startCoords
	i := 0
	//  drawer.rast.Clear()
	mask := image.NewRGBA(dst.Bounds())
	painter := raster.NewRGBAPainter(mask)
	rast := raster.NewRasterizer(mask.Bounds().Dx(), mask.Bounds().Dy())
	rast.UseNonZeroWinding = true
	rast.Clear()

	var paths []raster.Path
	for _, r := range drawer.str {
		// add basic distortion
		if opt.Distort {
			for i := range drawer.glyphCache[r].Points {
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
		if !opt.DisableOutline {
			rast.AddStroke(path, drawer.outlineWidth, nil, nil)
		}
		offset.X = startCoords.X + coords[i]
		paths = append(paths, path)
		i++
	}
	painter.SetColor(drawer.outlineColor)
	rast.Rasterize(painter)

	// second pass draw glyph itself
	rast.Clear()
	offset = startCoords
	for i = range paths {

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

func randomUint8() uint8 {
	rand.Seed(time.Now().UnixNano())
	return uint8(rand.Intn(256))
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
	if input == "random" {
		return randomRGB()
	}

	n, err := strconv.ParseInt(input, 16, 32)
	if err != nil {
		logger.Println(err)
		return color.RGBA{0, 0, 0, 255}
	}

	return color.RGBA{
		R: uint8(n >> 16 & 0xff),
		G: uint8(n >> 8 & 0xff),
		B: uint8(n & 0xff),
		A: 255,
	}
}

func noise() float64 {
	rand.Seed(time.Now().UnixNano())
	return (rand.Float64() - 0.5) * 2.0
}

func init() {
	filenames := strings.Fields(os.Getenv("APP_FONTS"))
	if len(filenames) == 0 {
		logger.Fatal("no fonts found")
	}

	for _, filename := range filenames {
		fontBytes, err := ioutil.ReadFile(filename)
		if err != nil {
			logger.Fatal("init: ", err)
		}

		f, err := freetype.ParseFont(fontBytes)
		if err != nil {
			logger.Fatal("init: ", err)
		}

		fonts = append(fonts, f)
	}
}

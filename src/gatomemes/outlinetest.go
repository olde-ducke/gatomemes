package gatomemes

import (
	"image"
	"io/ioutil"
	"log"
	"os"

	"github.com/golang/freetype"
	"github.com/golang/freetype/raster"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

func DrawTestOutline() {
	fontbytes, err := ioutil.ReadFile(os.Getenv("PROJECTFONT"))
	checkError("FontPile:", err)

	//contour
	f, err := freetype.ParseFont(fontbytes)
	checkError("parse_font: ", err)

	cntx := freetype.NewContext()

	r := 'I'
	cntx.SetFont(f)
	cntx.SetDPI(72.0) // default is 72.0, btw
	//cntx.SetSrc(image.White)
	cntx.SetHinting(font.HintingFull)

	size := 58.0
	test := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{512, 512}})
	painter := raster.NewRGBAPainter(test)
	rast := raster.NewRasterizer(512, 512)
	rast.SetBounds(512, 512)

	var glyph_buffer truetype.GlyphBuf
	err = glyph_buffer.Load(f, float2fixed(size), f.Index(r), font.HintingFull)
	if err != nil {
		log.Fatal(err)
	}

	prev := fixed.Point26_6{X: 0, Y: 0}
	multiplier := float2fixed(2)
	offset := fixed.Point26_6{X: multiplier, Y: multiplier}
	/*maxX := glyph_buffer.Points[0].X
	maxY := glyph_buffer.Points[0].Y
	for _, point := range glyph_buffer.Points {
		if point.X > maxX {
			maxX = point.X
		}
		if point.Y > maxY {
			maxY = point.Y
		}
	}
	centerPoint := fixed.Point26_6{X: maxX << 6 / 2, Y: maxY << 6 / 2}*/

	for i := range glyph_buffer.Points {
		log.Println("before:", glyph_buffer.Points[i].X, glyph_buffer.Points[i].Y)
		dx := glyph_buffer.Points[i].X - prev.X
		dy := glyph_buffer.Points[i].Y - prev.Y
		if dx > 0 {
			offset.X = multiplier * 1
		}
		if dx < 0 {
			offset.X = multiplier * 1
		}
		if dy > 0 {
			offset.Y = multiplier * -1
		}
		if dy < 0 {
			offset.Y = multiplier * -1
		}
		glyph_buffer.Points[i].X += offset.X
		glyph_buffer.Points[i].Y += offset.Y
		prev.X = glyph_buffer.Points[i].X - offset.X
		prev.Y = glyph_buffer.Points[i].Y - offset.Y
		log.Println("after :", glyph_buffer.Points[i].X, glyph_buffer.Points[i].Y)
	}

	painter.SetColor(image.Black)
	drawContour(rast, glyph_buffer.Points, fixed.I(256), fixed.I(256))
	rast.Clear()
	e0 := 0
	for _, e1 := range glyph_buffer.Ends {
		drawContour(rast, glyph_buffer.Points[e0:e1], fixed.I(256), fixed.I(256))
		e0 = e1
	}
	rast.Rasterize(painter)
	rast.Clear()

	err = glyph_buffer.Load(f, float2fixed(size), f.Index(r), font.HintingFull)
	painter.SetColor(image.White)
	drawContour(rast, glyph_buffer.Points, fixed.I(256), fixed.I(256))
	rast.Clear()
	e0 = 0
	for _, e1 := range glyph_buffer.Ends {
		drawContour(rast, glyph_buffer.Points[e0:e1], fixed.I(256), fixed.I(256))
		e0 = e1
	}
	rast.Rasterize(painter)

	savePngOnDisk(test, "img/test.png")
}

func drawContour(r *raster.Rasterizer, ps []truetype.Point, dx, dy fixed.Int26_6) {
	if len(ps) == 0 {
		return
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
	r.Start(start)
	q0, on0 := start, true
	for _, p := range others {
		q := fixed.Point26_6{
			X: dx + p.X,
			Y: dy - p.Y,
		}
		on := p.Flags&0x01 != 0
		if on {
			if on0 {
				r.Add1(q)
			} else {
				r.Add2(q0, q)
			}
		} else {
			if on0 {
				// No-op.
			} else {
				mid := fixed.Point26_6{
					X: (q0.X + q.X) / 2,
					Y: (q0.Y + q.Y) / 2,
				}
				r.Add2(q0, mid)
			}
		}
		q0, on0 = q, on
	}
	if on0 {
		r.Add1(start)
	} else {
		r.Add2(q0, start)
	}
}

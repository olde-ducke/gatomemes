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
	err = glyph_buffer.Load(f, float2fixed(size), f.Index('Щ'), font.HintingFull)
	if err != nil {
		log.Fatal(err)
	}

	painter.SetColor(image.Black)
	multiplier := fixed.Point26_6{X: float2fixed(1.1), Y: float2fixed(1.1)}
	drawCountour(rast, glyph_buffer.Points, fixed.I(256)-2*multiplier.X, fixed.I(256)+2*multiplier.Y, multiplier)
	rast.Rasterize(painter)
	rast.Clear()
	painter.SetColor(image.White)
	drawNormalCountour(rast, glyph_buffer.Points, fixed.I(256), fixed.I(256))
	rast.Rasterize(painter)
	rast.Clear()

	savePngOnDisk(test, "img/test.png")
}

func drawCountour(rast *raster.Rasterizer, points []truetype.Point, dx, dy fixed.Int26_6, multiplier fixed.Point26_6) {
	startpoint := fixed.Point26_6{
		X: dx + points[0].X.Mul(multiplier.X),
		Y: dy - points[0].Y.Mul(multiplier.Y),
	}
	others := []truetype.Point(nil)
	if points[0].Flags&0x01 != 0 {
		others = points[1:]
	} else {
		lastpoint := fixed.Point26_6{
			X: dx + points[len(points)-1].X.Mul(multiplier.X),
			Y: dy - points[len(points)-1].Y.Mul(multiplier.Y),
		}
		if points[len(points)-1].Flags&0x01 != 0 {
			startpoint = lastpoint
			others = points[:len(points)-1]
		} else {
			startpoint = fixed.Point26_6{
				X: (startpoint.X + lastpoint.X) / 2,
				Y: (startpoint.Y + lastpoint.Y) / 2,
			}
			others = points
		}
	}
	rast.Start(startpoint)
	startposition, on0 := startpoint, true
	for _, p := range others {
		position := fixed.Point26_6{
			X: dx + p.X.Mul(multiplier.X),
			Y: dy - p.Y.Mul(multiplier.Y),
		}
		oncontour := p.Flags&0x01 != 0
		if oncontour {
			if on0 {
				rast.Add1(position)
			} else {
				rast.Add2(startposition, position)
			}
		} else {
			if on0 {
				// No-op.
			} else {
				mid := fixed.Point26_6{
					X: (startposition.X + position.X) / 2,
					Y: (startposition.Y + position.Y) / 2,
				}
				rast.Add2(startposition, mid)
			}
		}
		startposition, on0 = position, oncontour
	}
	// Close the curve.
	if on0 {
		rast.Add1(startpoint)
	} else {
		rast.Add2(startposition, startpoint)
	}
}

func drawNormalCountour(rast *raster.Rasterizer, points []truetype.Point, dx, dy fixed.Int26_6) {
	startpoint := fixed.Point26_6{
		X: dx + points[0].X,
		Y: dy - points[0].Y,
	}
	others := []truetype.Point(nil)
	if points[0].Flags&0x01 != 0 {
		others = points[1:]
	} else {
		lastpoint := fixed.Point26_6{
			X: dx + points[len(points)-1].X,
			Y: dy - points[len(points)-1].Y,
		}
		if points[len(points)-1].Flags&0x01 != 0 {
			startpoint = lastpoint
			others = points[:len(points)-1]
		} else {
			startpoint = fixed.Point26_6{
				X: (startpoint.X + lastpoint.X) / 2,
				Y: (startpoint.Y + lastpoint.Y) / 2,
			}
			others = points
		}
	}
	rast.Start(startpoint)
	startposition, on0 := startpoint, true
	for _, p := range others {
		position := fixed.Point26_6{
			X: dx + p.X,
			Y: dy - p.Y,
		}
		oncontour := p.Flags&0x01 != 0
		if oncontour {
			if on0 {
				rast.Add1(position)
			} else {
				rast.Add2(startposition, position)
			}
		} else {
			if on0 {
				// No-op.
			} else {
				mid := fixed.Point26_6{
					X: (startposition.X + position.X) / 2,
					Y: (startposition.Y + position.Y) / 2,
				}
				rast.Add2(startposition, mid)
			}
		}
		startposition, on0 = position, oncontour
	}
	// Close the curve.
	if on0 {
		rast.Add1(startpoint)
	} else {
		rast.Add2(startposition, startpoint)
	}
}

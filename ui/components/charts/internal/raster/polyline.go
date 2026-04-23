package raster

import (
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	chartcanvas "github.com/wwsheng009/mint/ui/components/charts/internal/canvas"
)

// Point is a discrete plot coordinate on the chart canvas.
type Point struct {
	X int
	Y int
}

// LineGlyphs controls which glyphs are used for each line direction.
type LineGlyphs struct {
	Horizontal rune
	Vertical   rune
	Upward     rune
	Downward   rune
}

// PolylineOptions controls how a polyline is rasterized onto the canvas.
type PolylineOptions struct {
	ShowPoints bool
	PointRune  rune
	Glyphs     LineGlyphs
}

// DefaultLineGlyphs returns the default Unicode line glyphs for charts.
func DefaultLineGlyphs() LineGlyphs {
	return LineGlyphs{
		Horizontal: '─',
		Vertical:   '│',
		Upward:     '╱',
		Downward:   '╲',
	}
}

// DrawPolyline rasterizes a polyline onto the provided canvas.
func DrawPolyline(surface *chartcanvas.RuneCanvas, points []Point, opts PolylineOptions) {
	if surface == nil || len(points) == 0 {
		return
	}

	if opts.PointRune == 0 {
		opts.PointRune = '●'
	}
	if opts.Glyphs == (LineGlyphs{}) {
		opts.Glyphs = DefaultLineGlyphs()
	}

	if len(points) == 1 {
		surface.Set(points[0].X, points[0].Y, opts.PointRune)
		return
	}

	for i := 1; i < len(points); i++ {
		drawSegment(surface, points[i-1], points[i], opts.Glyphs)
	}

	if opts.ShowPoints {
		for _, point := range points {
			surface.Set(point.X, point.Y, opts.PointRune)
		}
	}
}

// DrawPolylineToBuffer rasterizes a polyline directly onto a styled paint buffer.
func DrawPolylineToBuffer(buf *paint.Buffer, points []Point, lineStyle style.Style, opts PolylineOptions) {
	if buf == nil || len(points) == 0 {
		return
	}

	if opts.PointRune == 0 {
		opts.PointRune = '●'
	}
	if opts.Glyphs == (LineGlyphs{}) {
		opts.Glyphs = DefaultLineGlyphs()
	}

	if len(points) == 1 {
		buf.SetCell(points[0].X, points[0].Y, opts.PointRune, lineStyle)
		return
	}

	for i := 1; i < len(points); i++ {
		drawSegmentToBuffer(buf, points[i-1], points[i], lineStyle, opts.Glyphs)
	}

	if opts.ShowPoints {
		for _, point := range points {
			buf.SetCell(point.X, point.Y, opts.PointRune, lineStyle)
		}
	}
}

func drawSegment(surface *chartcanvas.RuneCanvas, start, end Point, glyphs LineGlyphs) {
	path := rasterizeSegment(start, end)
	if len(path) == 0 {
		return
	}

	if len(path) == 1 {
		surface.Set(path[0].X, path[0].Y, glyphs.Horizontal)
		return
	}

	for index, point := range path {
		nextIndex := index + 1
		if nextIndex >= len(path) {
			nextIndex = index
		}

		var glyph rune
		switch {
		case index == len(path)-1:
			glyph = stepGlyph(path[index-1], point, glyphs)
		default:
			glyph = stepGlyph(point, path[nextIndex], glyphs)
		}

		surface.Set(point.X, point.Y, glyph)
	}
}

func drawSegmentToBuffer(buf *paint.Buffer, start, end Point, lineStyle style.Style, glyphs LineGlyphs) {
	path := rasterizeSegment(start, end)
	if len(path) == 0 {
		return
	}

	if len(path) == 1 {
		buf.SetCell(path[0].X, path[0].Y, glyphs.Horizontal, lineStyle)
		return
	}

	for index, point := range path {
		nextIndex := index + 1
		if nextIndex >= len(path) {
			nextIndex = index
		}

		var glyph rune
		switch {
		case index == len(path)-1:
			glyph = stepGlyph(path[index-1], point, glyphs)
		default:
			glyph = stepGlyph(point, path[nextIndex], glyphs)
		}

		buf.SetCell(point.X, point.Y, glyph, lineStyle)
	}
}

func rasterizeSegment(start, end Point) []Point {
	x0, y0 := start.X, start.Y
	x1, y1 := end.X, end.Y

	dx := absInt(x1 - x0)
	dy := -absInt(y1 - y0)
	sx := stepSign(x0, x1)
	sy := stepSign(y0, y1)
	err := dx + dy

	path := make([]Point, 0, maxInt(dx, absInt(dy))+1)
	for {
		path = append(path, Point{X: x0, Y: y0})
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}

	return path
}

func stepGlyph(from, to Point, glyphs LineGlyphs) rune {
	switch {
	case from.Y == to.Y:
		return glyphs.Horizontal
	case from.X == to.X:
		return glyphs.Vertical
	case to.Y < from.Y:
		return glyphs.Upward
	default:
		return glyphs.Downward
	}
}

func stepSign(a, b int) int {
	switch {
	case a < b:
		return 1
	case a > b:
		return -1
	default:
		return 0
	}
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

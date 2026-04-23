package scatterplot

import "github.com/wwsheng009/mint/runtime/style"

// Point describes one 2D point on the scatter plot.
type Point struct {
	X float64
	Y float64
}

// Series describes one scatter plot series.
type Series struct {
	Name   string
	Points []Point
	Style  style.Style
	Glyph  rune
}

func copyPointSlice(src []Point) []Point {
	if len(src) == 0 {
		return nil
	}
	dst := make([]Point, len(src))
	copy(dst, src)
	return dst
}

func copySeriesSlice(src []Series) []Series {
	if len(src) == 0 {
		return nil
	}
	dst := make([]Series, len(src))
	for i, series := range src {
		dst[i] = Series{
			Name:   series.Name,
			Points: copyPointSlice(series.Points),
			Style:  series.Style,
			Glyph:  series.Glyph,
		}
	}
	return dst
}

func pointSlicesEqual(a, b []Point) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func seriesSlicesEqual(a, b []Series) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Name != b[i].Name ||
			a[i].Style != b[i].Style ||
			a[i].Glyph != b[i].Glyph ||
			!pointSlicesEqual(a[i].Points, b[i].Points) {
			return false
		}
	}
	return true
}

func copyFloat64Slice(src []float64) []float64 {
	if len(src) == 0 {
		return nil
	}
	dst := make([]float64, len(src))
	copy(dst, src)
	return dst
}

func float64SlicesEqual(a, b []float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

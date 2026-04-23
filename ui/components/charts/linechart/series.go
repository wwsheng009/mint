package linechart

import "github.com/wwsheng009/mint/runtime/style"

// Series describes one data series in a line chart.
type Series struct {
	Name  string
	Data  []float64
	Style style.Style
}

func copySeriesSlice(src []Series) []Series {
	if len(src) == 0 {
		return nil
	}
	dst := make([]Series, len(src))
	for i, series := range src {
		dst[i] = Series{
			Name:  series.Name,
			Data:  copyFloat64Slice(series.Data),
			Style: series.Style,
		}
	}
	return dst
}

func seriesSlicesEqual(a, b []Series) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Name != b[i].Name || a[i].Style != b[i].Style || !float64SlicesEqual(a[i].Data, b[i].Data) {
			return false
		}
	}
	return true
}

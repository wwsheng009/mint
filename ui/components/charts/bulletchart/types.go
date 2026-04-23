package bulletchart

import (
	"sort"

	"github.com/wwsheng009/mint/runtime/style"
)

// ValueLabelMode controls where the bullet chart value summary is rendered.
type ValueLabelMode int

const (
	ValueLabelModeAuto ValueLabelMode = iota
	ValueLabelModeInline
	ValueLabelModeBelow
)

// Direction controls how default bullet chart semantics map to thresholds.
type Direction int

const (
	DirectionNeutral Direction = iota
	DirectionHigherBetter
	DirectionLowerBetter
)

// QualitativeRange describes one threshold region of a bullet chart.
type QualitativeRange struct {
	Limit int
	Glyph rune
	Style style.Style
}

func defaultQualitativeRanges(max int) []QualitativeRange {
	if max <= 0 {
		max = 100
	}
	return []QualitativeRange{
		{Limit: max / 3, Glyph: '░'},
		{Limit: (max * 2) / 3, Glyph: '▒'},
		{Limit: max, Glyph: '▓'},
	}
}

func normalizeQualitativeRanges(ranges []QualitativeRange, max int) []QualitativeRange {
	if max <= 0 {
		max = 100
	}
	if len(ranges) == 0 {
		return defaultQualitativeRanges(max)
	}

	dst := make([]QualitativeRange, len(ranges))
	copy(dst, ranges)
	for i := range dst {
		if dst[i].Limit < 0 {
			dst[i].Limit = 0
		}
		if dst[i].Limit > max {
			dst[i].Limit = max
		}
		if dst[i].Glyph == 0 {
			dst[i].Glyph = '░'
		}
	}
	sort.Slice(dst, func(i, j int) bool {
		return dst[i].Limit < dst[j].Limit
	})
	return dst
}

func copyQualitativeRangeSlice(src []QualitativeRange, max int) []QualitativeRange {
	if len(src) == 0 {
		return nil
	}
	dst := make([]QualitativeRange, len(src))
	copy(dst, normalizeQualitativeRanges(src, max))
	return dst
}

func qualitativeRangesEqual(a, b []QualitativeRange, max int) bool {
	a = normalizeQualitativeRanges(a, max)
	b = normalizeQualitativeRanges(b, max)
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Limit != b[i].Limit || a[i].Glyph != b[i].Glyph || a[i].Style != b[i].Style {
			return false
		}
	}
	return true
}

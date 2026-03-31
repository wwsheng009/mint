package scale

import "math"

// Linear maps a continuous numeric domain into a discrete integer range.
type Linear struct {
	DomainMin float64
	DomainMax float64
	RangeMin  int
	RangeMax  int
}

// NewLinear creates a new linear scale.
func NewLinear(domainMin, domainMax float64, rangeMin, rangeMax int) Linear {
	return Linear{
		DomainMin: domainMin,
		DomainMax: domainMax,
		RangeMin:  rangeMin,
		RangeMax:  rangeMax,
	}
}

// Map converts a domain value into the configured discrete range.
func (s Linear) Map(value float64) int {
	if s.RangeMin == s.RangeMax {
		return s.RangeMin
	}

	if math.Abs(s.DomainMax-s.DomainMin) < 1e-9 {
		return int(math.Round(float64(s.RangeMin+s.RangeMax) / 2))
	}

	normalized := (value - s.DomainMin) / (s.DomainMax - s.DomainMin)
	if normalized < 0 {
		normalized = 0
	}
	if normalized > 1 {
		normalized = 1
	}

	mapped := float64(s.RangeMin) + normalized*float64(s.RangeMax-s.RangeMin)
	return int(math.Round(mapped))
}

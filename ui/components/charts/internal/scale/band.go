package scale

import "math"

// Band maps ordered categorical indices into discrete x positions.
type Band struct {
	DomainSize int
	RangeMin   int
	RangeMax   int
}

// NewBand creates a new band scale.
func NewBand(domainSize, rangeMin, rangeMax int) Band {
	return Band{
		DomainSize: domainSize,
		RangeMin:   rangeMin,
		RangeMax:   rangeMax,
	}
}

// Position returns the discrete x position for the requested category index.
func (s Band) Position(index int) int {
	if s.DomainSize <= 0 {
		return s.RangeMin
	}
	if s.DomainSize == 1 || s.RangeMin == s.RangeMax {
		return int(math.Round(float64(s.RangeMin+s.RangeMax) / 2))
	}

	if index < 0 {
		index = 0
	}
	if index >= s.DomainSize {
		index = s.DomainSize - 1
	}

	ratio := float64(index) / float64(s.DomainSize-1)
	position := float64(s.RangeMin) + ratio*float64(s.RangeMax-s.RangeMin)
	return int(math.Round(position))
}

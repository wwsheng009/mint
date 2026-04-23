package scatterplot

// ReferenceBand describes one highlighted reference range on an axis.
type ReferenceBand struct {
	Min   float64
	Max   float64
	Label string
}

// NewReferenceBand creates a normalized highlighted reference range.
func NewReferenceBand(min, max float64) ReferenceBand {
	return normalizeReferenceBand(ReferenceBand{Min: min, Max: max})
}

// NewLabeledReferenceBand creates a normalized highlighted reference range with a legend label.
func NewLabeledReferenceBand(min, max float64, label string) ReferenceBand {
	return normalizeReferenceBand(ReferenceBand{
		Min:   min,
		Max:   max,
		Label: label,
	})
}

func normalizeReferenceBand(band ReferenceBand) ReferenceBand {
	if band.Min > band.Max {
		band.Min, band.Max = band.Max, band.Min
	}
	return band
}

func copyReferenceBandSlice(src []ReferenceBand) []ReferenceBand {
	if len(src) == 0 {
		return nil
	}
	dst := make([]ReferenceBand, len(src))
	for i, band := range src {
		dst[i] = normalizeReferenceBand(band)
	}
	return dst
}

func referenceBandSlicesEqual(a, b []ReferenceBand) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if normalizeReferenceBand(a[i]) != normalizeReferenceBand(b[i]) {
			return false
		}
	}
	return true
}

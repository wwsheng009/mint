package scatterplot

// ReferenceLine describes one highlighted reference marker on an axis.
type ReferenceLine struct {
	Value float64
	Label string
}

// NewReferenceLine creates a normalized highlighted reference marker.
func NewReferenceLine(value float64) ReferenceLine {
	return normalizeReferenceLine(ReferenceLine{Value: value})
}

// NewLabeledReferenceLine creates a normalized highlighted reference marker with a legend label.
func NewLabeledReferenceLine(value float64, label string) ReferenceLine {
	return normalizeReferenceLine(ReferenceLine{
		Value: value,
		Label: label,
	})
}

func normalizeReferenceLine(line ReferenceLine) ReferenceLine {
	return line
}

func copyReferenceLineSlice(src []ReferenceLine) []ReferenceLine {
	if len(src) == 0 {
		return nil
	}
	dst := make([]ReferenceLine, len(src))
	for i, line := range src {
		dst[i] = normalizeReferenceLine(line)
	}
	return dst
}

func referenceLineSlicesEqual(a, b []ReferenceLine) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if normalizeReferenceLine(a[i]) != normalizeReferenceLine(b[i]) {
			return false
		}
	}
	return true
}

func referenceLineValues(lines []ReferenceLine) []float64 {
	if len(lines) == 0 {
		return nil
	}
	values := make([]float64, len(lines))
	for i, line := range lines {
		values[i] = line.Value
	}
	return values
}

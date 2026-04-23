package downsample

import "math"

// ResampleNearestFloat64 reduces or expands a float64 series to the requested width
// using nearest-neighbor sampling. It is intentionally simple for the first phase.
func ResampleNearestFloat64(data []float64, width int) []float64 {
	if len(data) == 0 || width <= 0 {
		return nil
	}
	if len(data) == width {
		return copyFloat64Slice(data)
	}

	result := make([]float64, width)
	if width == 1 {
		result[0] = data[len(data)-1]
		return result
	}

	last := float64(len(data) - 1)
	for i := 0; i < width; i++ {
		index := int(math.Round(float64(i) * last / float64(width-1)))
		if index < 0 {
			index = 0
		}
		if index >= len(data) {
			index = len(data) - 1
		}
		result[i] = data[index]
	}
	return result
}

// MinMaxFloat64 returns the minimum and maximum values in a non-empty slice.
func MinMaxFloat64(data []float64) (float64, float64) {
	minVal := data[0]
	maxVal := data[0]
	for _, value := range data[1:] {
		if value < minVal {
			minVal = value
		}
		if value > maxVal {
			maxVal = value
		}
	}
	return minVal, maxVal
}

func copyFloat64Slice(src []float64) []float64 {
	if len(src) == 0 {
		return nil
	}
	dst := make([]float64, len(src))
	copy(dst, src)
	return dst
}

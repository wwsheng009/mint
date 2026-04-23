package linechart

import "math"

// resampleForContinuity uses a small LTTB-style sampler so narrow charts
// keep visually important turning points instead of flattening to nearest picks.
func resampleForContinuity(data []float64, threshold int) []float64 {
	if len(data) == 0 || threshold <= 0 {
		return nil
	}
	if threshold >= len(data) {
		return copyFloat64Slice(data)
	}
	if threshold == 1 {
		return []float64{data[0]}
	}
	if threshold == 2 {
		return []float64{data[0], data[len(data)-1]}
	}

	sampled := make([]float64, 0, threshold)
	sampled = append(sampled, data[0])

	every := float64(len(data)-2) / float64(threshold-2)
	a := 0
	for i := 0; i < threshold-2; i++ {
		rangeStart := int(math.Floor(float64(i)*every)) + 1
		rangeEnd := int(math.Floor(float64(i+1)*every)) + 1
		if rangeEnd >= len(data) {
			rangeEnd = len(data) - 1
		}
		avgRangeStart := int(math.Floor(float64(i+1)*every)) + 1
		avgRangeEnd := int(math.Floor(float64(i+2)*every)) + 1
		if avgRangeEnd >= len(data) {
			avgRangeEnd = len(data)
		}
		if avgRangeStart >= len(data) {
			avgRangeStart = len(data) - 1
		}
		if avgRangeStart >= avgRangeEnd {
			avgRangeEnd = minInt(avgRangeStart+1, len(data))
		}

		avgX := 0.0
		avgY := 0.0
		avgCount := avgRangeEnd - avgRangeStart
		for j := avgRangeStart; j < avgRangeEnd; j++ {
			avgX += float64(j)
			avgY += data[j]
		}
		if avgCount > 0 {
			avgX /= float64(avgCount)
			avgY /= float64(avgCount)
		} else {
			avgX = float64(avgRangeStart)
			avgY = data[avgRangeStart]
		}

		pointAX := float64(a)
		pointAY := data[a]
		maxArea := -1.0
		nextA := rangeStart
		for j := rangeStart; j < rangeEnd; j++ {
			area := math.Abs((pointAX-avgX)*(data[j]-pointAY) - (pointAX-float64(j))*(avgY-pointAY))
			if area > maxArea {
				maxArea = area
				nextA = j
			}
		}
		sampled = append(sampled, data[nextA])
		a = nextA
	}

	sampled = append(sampled, data[len(data)-1])
	return sampled
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

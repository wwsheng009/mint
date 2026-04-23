package paint

// RegionDiffResult contains region-level diff information between two buffers.
type RegionDiffResult struct {
	DirtyRegions []Rect
	HasChanges   bool
	ChangedCells int
}

type rowSegment struct {
	start int
	end   int
}

// RegionDiff computes changed rectangular regions between two buffers.
// It is optimized for partial framebuffer updates in async rendering paths.
func RegionDiff(front, back *Buffer) RegionDiffResult {
	result := RegionDiffResult{
		DirtyRegions: nil,
		HasChanges:   false,
		ChangedCells: 0,
	}

	if back == nil {
		return result
	}
	if front == nil || front.Width != back.Width || front.Height != back.Height {
		full := Rect{X: 0, Y: 0, Width: back.Width, Height: back.Height}
		result.DirtyRegions = []Rect{full}
		result.HasChanges = back.Width > 0 && back.Height > 0
		result.ChangedCells = back.Width * back.Height
		return result
	}

	width := minInt(front.Width, back.Width)
	height := minInt(front.Height, back.Height)
	if width <= 0 || height <= 0 {
		return result
	}

	segmentsByRow := make([][]rowSegment, height)
	for y := 0; y < height; y++ {
		rowFront := front.Cells[y]
		rowBack := back.Cells[y]
		x := 0
		for x < width {
			if IsCellEqual(rowFront[x], rowBack[x]) {
				x++
				continue
			}
			start := x
			for x < width && !IsCellEqual(rowFront[x], rowBack[x]) {
				x++
			}
			segmentsByRow[y] = append(segmentsByRow[y], rowSegment{start: start, end: x})
			result.ChangedCells += (x - start)
		}
	}

	rects := mergeSegmentsToRects(segmentsByRow)
	rects = mergeRects(rects)
	if len(rects) > 0 {
		result.HasChanges = true
		result.DirtyRegions = rects
	}
	return result
}

func mergeSegmentsToRects(rows [][]rowSegment) []Rect {
	if len(rows) == 0 {
		return nil
	}

	rects := make([]Rect, 0, 16)
	active := make(map[rowSegment]int)

	for y := 0; y < len(rows); y++ {
		nextActive := make(map[rowSegment]int, len(rows[y]))

		for _, seg := range rows[y] {
			if idx, ok := active[seg]; ok {
				rects[idx].Height++
				nextActive[seg] = idx
				continue
			}

			rects = append(rects, Rect{
				X:      seg.start,
				Y:      y,
				Width:  seg.end - seg.start,
				Height: 1,
			})
			nextActive[seg] = len(rects) - 1
		}

		active = nextActive
	}

	return rects
}

func mergeRects(rects []Rect) []Rect {
	if len(rects) <= 1 {
		return rects
	}

	merged := make([]Rect, 0, len(rects))
	used := make([]bool, len(rects))

	for i := 0; i < len(rects); i++ {
		if used[i] {
			continue
		}
		current := rects[i]
		used[i] = true

		for {
			changed := false
			for j := 0; j < len(rects); j++ {
				if used[j] {
					continue
				}
				if shouldMergeRect(current, rects[j]) {
					current = unionRect(current, rects[j])
					used[j] = true
					changed = true
				}
			}
			if !changed {
				break
			}
		}

		merged = append(merged, current)
	}

	return merged
}

func shouldMergeRect(a, b Rect) bool {
	ax2 := a.X + a.Width
	ay2 := a.Y + a.Height
	bx2 := b.X + b.Width
	by2 := b.Y + b.Height

	overlapOrTouchX := !(ax2 < b.X || bx2 < a.X)
	overlapOrTouchY := !(ay2 < b.Y || by2 < a.Y)
	return overlapOrTouchX && overlapOrTouchY
}

func unionRect(a, b Rect) Rect {
	x1 := minInt(a.X, b.X)
	y1 := minInt(a.Y, b.Y)
	x2 := maxInt(a.X+a.Width, b.X+b.Width)
	y2 := maxInt(a.Y+a.Height, b.Y+b.Height)
	return Rect{
		X:      x1,
		Y:      y1,
		Width:  x2 - x1,
		Height: y2 - y1,
	}
}

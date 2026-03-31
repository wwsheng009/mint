package axis

import "strings"

// HorizontalLine returns a simple horizontal axis line.
func HorizontalLine(width int) string {
	if width <= 0 {
		return ""
	}
	return strings.Repeat("─", width)
}

// LabelRow renders short labels at discrete positions across a row.
func LabelRow(width int, labels []string, positions []int, fallback rune) string {
	if width <= 0 {
		return ""
	}
	if fallback == 0 {
		fallback = '•'
	}

	row := []rune(strings.Repeat(" ", width))
	limit := len(labels)
	if len(positions) < limit {
		limit = len(positions)
	}
	for i := 0; i < limit; i++ {
		x := positions[i]
		if x < 0 || x >= width {
			continue
		}
		row[x] = LabelRune(labels[i], fallback)
	}
	return string(row)
}

// LabelRune returns the first visible rune of the label or the fallback rune.
func LabelRune(label string, fallback rune) rune {
	label = strings.TrimSpace(label)
	if label == "" {
		return fallback
	}
	for _, r := range label {
		return r
	}
	return fallback
}

// GridRows returns evenly distributed interior row indices for horizontal grid lines.
func GridRows(height, count int) []int {
	if height <= 2 || count <= 0 {
		return nil
	}

	rows := make([]int, 0, count)
	last := -1
	for i := 1; i <= count; i++ {
		row := int(float64(i) * float64(height-1) / float64(count+1))
		if row <= 0 || row >= height-1 || row == last {
			continue
		}
		rows = append(rows, row)
		last = row
	}
	return rows
}

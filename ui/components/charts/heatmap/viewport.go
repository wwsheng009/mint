package heatmap

// Viewport describes the visible row/column window for large matrices.
type Viewport struct {
	RowStart int
	RowCount int
	ColStart int
	ColCount int
}

// NewViewport creates a normalized heatmap viewport.
func NewViewport(rowStart, rowCount, colStart, colCount int) Viewport {
	return normalizeViewport(Viewport{
		RowStart: rowStart,
		RowCount: rowCount,
		ColStart: colStart,
		ColCount: colCount,
	})
}

func normalizeViewport(viewport Viewport) Viewport {
	if viewport.RowStart < 0 {
		viewport.RowStart = 0
	}
	if viewport.ColStart < 0 {
		viewport.ColStart = 0
	}
	if viewport.RowCount < 0 {
		viewport.RowCount = 0
	}
	if viewport.ColCount < 0 {
		viewport.ColCount = 0
	}
	return viewport
}

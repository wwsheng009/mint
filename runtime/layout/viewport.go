package layout

// ScrollViewportMetricsReceiver is implemented by adapters that can receive the
// measured content size and the visible viewport size after layout.
type ScrollViewportMetricsReceiver interface {
	SetScrollViewportMetrics(contentWidth, contentHeight, viewportWidth, viewportHeight int)
}

func intersectRects(a, b Rect) (Rect, bool) {
	x1 := max(a.X, b.X)
	y1 := max(a.Y, b.Y)
	x2 := min(a.X+a.Width, b.X+b.Width)
	y2 := min(a.Y+a.Height, b.Y+b.Height)
	if x1 >= x2 || y1 >= y2 {
		return Rect{}, false
	}
	return Rect{X: x1, Y: y1, Width: x2 - x1, Height: y2 - y1}, true
}

func cloneRectPtr(r Rect) *Rect {
	return &Rect{X: r.X, Y: r.Y, Width: r.Width, Height: r.Height}
}

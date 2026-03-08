package scroll

type VerticalViewport struct {
	Offset      int
	ContentSize int
	ViewSize    int
}

func NewVerticalViewport(contentSize, viewSize, offset int) VerticalViewport {
	v := VerticalViewport{
		Offset:      offset,
		ContentSize: contentSize,
		ViewSize:    viewSize,
	}
	v.Normalize()
	return v
}

func (v *VerticalViewport) Set(contentSize, viewSize, offset int) {
	v.ContentSize = contentSize
	v.ViewSize = viewSize
	v.Offset = offset
	v.Normalize()
}

func (v *VerticalViewport) Normalize() {
	if v.ContentSize < 0 {
		v.ContentSize = 0
	}
	if v.ViewSize < 1 {
		v.ViewSize = 1
	}
	if v.Offset < 0 {
		v.Offset = 0
	}
	maxOffset := v.MaxOffset()
	if v.Offset > maxOffset {
		v.Offset = maxOffset
	}
}

func (v VerticalViewport) MaxOffset() int {
	maxOffset := v.ContentSize - v.ViewSize
	if maxOffset < 0 {
		return 0
	}
	return maxOffset
}

func (v VerticalViewport) IsScrollable() bool {
	return v.ContentSize > v.ViewSize
}

func (v VerticalViewport) VisibleRange() (start, end int) {
	start = v.Offset
	if start < 0 {
		start = 0
	}
	if start > v.ContentSize {
		start = v.ContentSize
	}
	end = start + v.ViewSize
	if end > v.ContentSize {
		end = v.ContentSize
	}
	if end < start {
		end = start
	}
	return start, end
}

func (v *VerticalViewport) ScrollTo(offset int) bool {
	old := v.Offset
	v.Offset = offset
	v.Normalize()
	return old != v.Offset
}

func (v *VerticalViewport) ScrollBy(delta int) bool {
	return v.ScrollTo(v.Offset + delta)
}

func (v *VerticalViewport) PageUp() bool {
	return v.ScrollBy(-v.ViewSize)
}

func (v *VerticalViewport) PageDown() bool {
	return v.ScrollBy(v.ViewSize)
}

func (v *VerticalViewport) EnsureVisible(index int) bool {
	if index < 0 {
		index = 0
	}

	old := v.Offset
	if index < v.Offset {
		v.Offset = index
	} else if index >= v.Offset+v.ViewSize {
		v.Offset = index - v.ViewSize + 1
	}
	v.Normalize()
	return old != v.Offset
}

package paint

// PartialFrameBuffer keeps a mutable frame image that can be updated by regions.
// It is used by AsyncRenderer to avoid full-buffer copies on each submitted frame.
type PartialFrameBuffer struct {
	buffer *Buffer
}

// NewPartialFrameBuffer creates a region-updatable framebuffer.
func NewPartialFrameBuffer(width, height int) *PartialFrameBuffer {
	return &PartialFrameBuffer{
		buffer: NewBuffer(width, height),
	}
}

// Buffer returns the underlying buffer.
func (p *PartialFrameBuffer) Buffer() *Buffer {
	if p == nil {
		return nil
	}
	return p.buffer
}

// Resize resizes and clears the internal buffer.
func (p *PartialFrameBuffer) Resize(width, height int) {
	if p == nil {
		return
	}
	if p.buffer == nil {
		p.buffer = NewBuffer(width, height)
		return
	}
	p.buffer.Reset(width, height)
}

// CopyFrom copies the full source frame into the internal buffer.
func (p *PartialFrameBuffer) CopyFrom(src *Buffer) {
	if p == nil || src == nil {
		return
	}
	ensureBufferSize(&p.buffer, src.Width, src.Height)
	copyBufferRect(p.buffer, src, Rect{X: 0, Y: 0, Width: src.Width, Height: src.Height})
}

// ApplyFrom applies only the specified regions from source to internal buffer.
func (p *PartialFrameBuffer) ApplyFrom(src *Buffer, regions []Rect) {
	if p == nil || src == nil {
		return
	}
	ensureBufferSize(&p.buffer, src.Width, src.Height)
	for _, rect := range regions {
		copyBufferRect(p.buffer, src, rect)
	}
}

func ensureBufferSize(dst **Buffer, width, height int) {
	if *dst == nil {
		*dst = NewBuffer(width, height)
		return
	}
	if (*dst).Width != width || (*dst).Height != height {
		(*dst).Reset(width, height)
	}
}

func copyBufferRect(dst, src *Buffer, rect Rect) {
	if dst == nil || src == nil || rect.Width <= 0 || rect.Height <= 0 {
		return
	}

	startX := rect.X
	startY := rect.Y
	endX := rect.X + rect.Width
	endY := rect.Y + rect.Height

	if startX < 0 {
		startX = 0
	}
	if startY < 0 {
		startY = 0
	}
	if endX > src.Width {
		endX = src.Width
	}
	if endY > src.Height {
		endY = src.Height
	}
	if endX > dst.Width {
		endX = dst.Width
	}
	if endY > dst.Height {
		endY = dst.Height
	}
	if startX >= endX || startY >= endY {
		return
	}

	for y := startY; y < endY; y++ {
		copy(dst.Cells[y][startX:endX], src.Cells[y][startX:endX])
	}
}

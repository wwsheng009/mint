package paint

import (
	"sync"
	"sync/atomic"
	"time"
)

// AsyncRendererOptions configures AsyncRenderer behavior.
type AsyncRendererOptions struct {
	FrameInterval time.Duration
	Output        func(string)
}

// AsyncRendererStats tracks async rendering pipeline counters.
type AsyncRendererStats struct {
	SubmittedFrames int64
	RenderedFrames  int64
	DroppedFrames   int64
	PendingRegions  int
}

// AsyncRenderer provides scheduler-driven rendering with region diff and
// partial framebuffer updates.
type AsyncRenderer struct {
	mu sync.Mutex

	renderer *Renderer
	stage    *PartialFrameBuffer

	outputFn       func(string)
	frameInterval  time.Duration
	ticker         *time.Ticker
	stopCh         chan struct{}
	doneCh         chan struct{}
	started        bool
	pendingRects   []Rect
	pendingFull    bool
	lastSubmission time.Time

	submittedFrames int64
	renderedFrames  int64
	droppedFrames   int64
	stopOnce        sync.Once
}

// NewAsyncRenderer creates a new async renderer.
func NewAsyncRenderer(width, height int, opts AsyncRendererOptions) *AsyncRenderer {
	interval := opts.FrameInterval
	if interval <= 0 {
		interval = 16 * time.Millisecond // ~60 FPS
	}

	output := opts.Output
	if output == nil {
		output = func(string) {}
	}

	return &AsyncRenderer{
		renderer:      NewRenderer(width, height),
		stage:         NewPartialFrameBuffer(width, height),
		outputFn:      output,
		frameInterval: interval,
		stopCh:        make(chan struct{}),
		doneCh:        make(chan struct{}),
		pendingRects:  make([]Rect, 0, 8),
	}
}

// Start starts the scheduler loop.
func (a *AsyncRenderer) Start() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.started {
		return
	}
	a.ticker = time.NewTicker(a.frameInterval)
	a.started = true

	go a.loop()
}

// Stop stops the scheduler loop and flushes pending work.
func (a *AsyncRenderer) Stop() {
	a.mu.Lock()
	if !a.started {
		a.mu.Unlock()
		return
	}
	a.started = false
	a.mu.Unlock()

	a.stopOnce.Do(func() {
		close(a.stopCh)
	})
	<-a.doneCh
}

// Resize resizes async pipeline buffers.
func (a *AsyncRenderer) Resize(width, height int) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.stage.Resize(width, height)
	a.renderer.Resize(width, height)
	a.pendingRects = append(a.pendingRects, Rect{X: 0, Y: 0, Width: width, Height: height})
	a.pendingFull = true
}

// SubmitFrame submits the latest frame snapshot.
// The source frame is consumed synchronously via region copy, so caller can
// safely mutate it after SubmitFrame returns.
func (a *AsyncRenderer) SubmitFrame(frame *Buffer, dirtyHints []Rect, forceFull bool) {
	if frame == nil {
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	atomic.AddInt64(&a.submittedFrames, 1)
	now := time.Now()
	if !a.lastSubmission.IsZero() && now.Sub(a.lastSubmission) < a.frameInterval/4 {
		atomic.AddInt64(&a.droppedFrames, 1)
	}
	a.lastSubmission = now

	stageBuf := a.stage.Buffer()
	if stageBuf == nil || stageBuf.Width != frame.Width || stageBuf.Height != frame.Height {
		a.stage.Resize(frame.Width, frame.Height)
		a.stage.CopyFrom(frame)
		a.pendingFull = true
		a.pendingRects = append(a.pendingRects, Rect{X: 0, Y: 0, Width: frame.Width, Height: frame.Height})
		forceFull = true
	} else if forceFull {
		a.stage.CopyFrom(frame)
		a.pendingFull = true
		a.pendingRects = append(a.pendingRects, Rect{X: 0, Y: 0, Width: frame.Width, Height: frame.Height})
	} else {
		diff := RegionDiff(stageBuf, frame)
		if diff.HasChanges && len(diff.DirtyRegions) > 0 {
			a.stage.ApplyFrom(frame, diff.DirtyRegions)
			a.pendingRects = append(a.pendingRects, diff.DirtyRegions...)
		}
	}

	for _, hint := range dirtyHints {
		if hint.Width > 0 && hint.Height > 0 {
			a.pendingRects = append(a.pendingRects, hint)
		}
	}
}

// FlushNow forces immediate rendering of pending frame changes.
func (a *AsyncRenderer) FlushNow() {
	a.renderPending()
}

// Stats returns async renderer statistics.
func (a *AsyncRenderer) Stats() AsyncRendererStats {
	a.mu.Lock()
	pending := len(a.pendingRects)
	a.mu.Unlock()

	return AsyncRendererStats{
		SubmittedFrames: atomic.LoadInt64(&a.submittedFrames),
		RenderedFrames:  atomic.LoadInt64(&a.renderedFrames),
		DroppedFrames:   atomic.LoadInt64(&a.droppedFrames),
		PendingRegions:  pending,
	}
}

func (a *AsyncRenderer) loop() {
	defer func() {
		a.mu.Lock()
		if a.ticker != nil {
			a.ticker.Stop()
			a.ticker = nil
		}
		a.started = false
		a.mu.Unlock()
		close(a.doneCh)
	}()

	for {
		select {
		case <-a.ticker.C:
			a.renderPending()
		case <-a.stopCh:
			a.renderPending()
			return
		}
	}
}

func (a *AsyncRenderer) renderPending() {
	var output string
	var outputFn func(string)

	a.mu.Lock()
	stageBuf := a.stage.Buffer()
	if stageBuf == nil {
		a.mu.Unlock()
		return
	}

	if !a.pendingFull && len(a.pendingRects) == 0 {
		a.mu.Unlock()
		return
	}

	if a.renderer.GetBackBuffer().Width != stageBuf.Width || a.renderer.GetBackBuffer().Height != stageBuf.Height {
		a.renderer.Resize(stageBuf.Width, stageBuf.Height)
		a.pendingFull = true
	}

	back := a.renderer.GetBackBuffer()

	if a.pendingFull {
		copyBufferRect(back, stageBuf, Rect{X: 0, Y: 0, Width: stageBuf.Width, Height: stageBuf.Height})
		a.renderer.ForceFullRender()
	} else {
		regions := mergeRects(a.pendingRects)
		for _, rect := range regions {
			copyBufferRect(back, stageBuf, rect)
			a.renderer.MarkDirtyRect(rect)
		}
	}

	output = a.renderer.Render()
	outputFn = a.outputFn
	a.pendingRects = a.pendingRects[:0]
	a.pendingFull = false
	a.mu.Unlock()

	if output != "" {
		atomic.AddInt64(&a.renderedFrames, 1)
		outputFn(output)
	}
}

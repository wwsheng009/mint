// Package devtools provides the frame timeline for DevTools.
//
// This file implements the FrameTimeline that provides a timeline view
// of frames with their causal relationships and performance metrics.
// P1-2: 使用环形缓冲区实现，避免 O(n) 切片
package devtools

import (
	"sync"
	"sync/atomic"
	"time"
)

// FrameTimeline provides a timeline view of frames with causal relationships.
// P1-2: 使用环形缓冲区替代普通切片，实现 O(1) 插入和裁剪
type FrameTimeline struct {
	enabled atomic.Uint32

	// P1-2: Ring buffer implementation
	buffer    []FrameEntry     // 预分配的环形缓冲区
	capacity  int              // 缓冲区容量
	writePos  uint32           // 写入位置
	count     uint32           // 当前元素数量
	bufferMu  sync.RWMutex     // 保护 buffer, writePos, count

	// Current frame being built
	currentFrame atomic.Pointer[FrameEntry]
}

// FrameEntry represents a single frame in the timeline.
type FrameEntry struct {
	FrameID     FrameID
	StartTime   time.Time
	EndTime     time.Time
	Duration    time.Duration

	// Causal data
	EventCount   int
	MutationCount int
	LayoutCount  int
	RepaintCount int
	EdgeCount    int

	// Performance metrics
	LayoutTime   time.Duration
	PaintTime    time.Duration
	TotalTime    time.Duration

	// Summary reference
	Summary *FrameSummary
}

// NewFrameTimeline creates a new frame timeline with ring buffer.
func NewFrameTimeline() *FrameTimeline {
	return NewFrameTimelineWithCapacity(100)
}

// NewFrameTimelineWithCapacity creates a frame timeline with specified capacity.
// Capacity must be a power of 2 for efficient ring buffer operation.
func NewFrameTimelineWithCapacity(capacity int) *FrameTimeline {
	// Ensure capacity is a power of 2
	if capacity&(capacity-1) != 0 {
		// Round up to next power of 2
		capacity = 1 << (32 - nlz32(uint32(capacity)))
	}

	ft := &FrameTimeline{
		buffer:       make([]FrameEntry, capacity),
		capacity:     capacity,
		writePos:     0,
		count:        0,
	}
	ft.enabled.Store(0)
	return ft
}

// Enable enables the frame timeline.
func (ft *FrameTimeline) Enable() {
	ft.enabled.Store(1)
}

// Disable disables the frame timeline.
func (ft *FrameTimeline) Disable() {
	ft.enabled.Store(0)
}

// IsEnabled returns whether the frame timeline is enabled.
func (ft *FrameTimeline) IsEnabled() bool {
	return ft.enabled.Load() == 1
}

// BeginFrame starts tracking a new frame.
func (ft *FrameTimeline) BeginFrame(frameID FrameID) *FrameEntry {
	if !ft.IsEnabled() {
		return nil
	}

	entry := &FrameEntry{
		FrameID:   frameID,
		StartTime: time.Now(),
	}
	ft.currentFrame.Store(entry)
	return entry
}

// EndFrame ends tracking the current frame.
// P1-2: 使用环形缓冲区，O(1) 写入
func (ft *FrameTimeline) EndFrame() {
	if !ft.IsEnabled() {
		return
	}

	entry := ft.currentFrame.Load()
	if entry == nil {
		return
	}

	entry.EndTime = time.Now()
	entry.Duration = entry.EndTime.Sub(entry.StartTime)
	entry.TotalTime = entry.Duration

	// Add to ring buffer (O(1) operation)
	ft.bufferMu.Lock()
	pos := ft.writePos % uint32(ft.capacity)
	ft.buffer[pos] = *entry

	// Update position and count
	ft.writePos++
	if ft.count < uint32(ft.capacity) {
		ft.count++
	}
	ft.bufferMu.Unlock()

	// Clear current
	ft.currentFrame.Store(nil)
}

// AttachGraph attaches causal graph data to the current frame entry.
func (ft *FrameTimeline) AttachGraph(summary *FrameSummary) {
	if !ft.IsEnabled() || summary == nil {
		return
	}

	entry := ft.currentFrame.Load()
	if entry == nil {
		return
	}

	entry.Summary = summary
	entry.EventCount = summary.EventCount
	entry.MutationCount = summary.MutationCount
	entry.LayoutCount = summary.LayoutCount
	entry.RepaintCount = summary.RepaintCount
	entry.EdgeCount = summary.EdgeCount
}

// SetLayoutTime sets the layout time for the current frame.
func (ft *FrameTimeline) SetLayoutTime(d time.Duration) {
	if !ft.IsEnabled() {
		return
	}

	entry := ft.currentFrame.Load()
	if entry != nil {
		entry.LayoutTime = d
	}
}

// SetPaintTime sets the paint time for the current frame.
func (ft *FrameTimeline) SetPaintTime(d time.Duration) {
	if !ft.IsEnabled() {
		return
	}

	entry := ft.currentFrame.Load()
	if entry != nil {
		entry.PaintTime = d
	}
}

// GetFrame returns a frame entry by frame ID.
func (ft *FrameTimeline) GetFrame(frameID FrameID) *FrameEntry {
	ft.bufferMu.RLock()
	defer ft.bufferMu.RUnlock()

	// Search in reverse order (most recent first)
	for i := uint32(0); i < ft.count; i++ {
		pos := (ft.writePos - 1 - i) % uint32(ft.capacity)
		if ft.buffer[pos].FrameID == frameID {
			return &ft.buffer[pos]
		}
	}
	return nil
}

// GetAllFrames returns all frame entries in chronological order.
func (ft *FrameTimeline) GetAllFrames() []*FrameEntry {
	ft.bufferMu.RLock()
	defer ft.bufferMu.RUnlock()

	n := ft.count
	frames := make([]*FrameEntry, n)

	for i := uint32(0); i < n; i++ {
		var pos uint32
		if n < uint32(ft.capacity) {
			// Buffer not full yet, read from beginning
			pos = i
		} else {
			// Buffer is full, read from oldest
			pos = (ft.writePos - n + i) % uint32(ft.capacity)
		}
		frames[i] = &ft.buffer[pos]
	}
	return frames
}

// GetLastNFrames returns the last N frame entries.
func (ft *FrameTimeline) GetLastNFrames(n int) []*FrameEntry {
	ft.bufferMu.RLock()
	defer ft.bufferMu.RUnlock()

	count := ft.count
	if n > int(count) {
		n = int(count)
	}

	frames := make([]*FrameEntry, n)
	for i := 0; i < n; i++ {
		pos := (ft.writePos - uint32(n) + uint32(i)) % uint32(ft.capacity)
		frames[i] = &ft.buffer[pos]
	}
	return frames
}

// GetFrameCount returns the number of frames in the timeline.
func (ft *FrameTimeline) GetFrameCount() int {
	ft.bufferMu.RLock()
	defer ft.bufferMu.RUnlock()
	return int(ft.count)
}

// GetFrameByIndex returns a frame entry by index (0 = oldest).
func (ft *FrameTimeline) GetFrameByIndex(index int) *FrameEntry {
	ft.bufferMu.RLock()
	defer ft.bufferMu.RUnlock()

	if index < 0 || index >= int(ft.count) {
		return nil
	}

	var pos uint32
	if ft.count < uint32(ft.capacity) {
		// Buffer not full yet
		pos = uint32(index)
	} else {
		// Buffer is full
		pos = (ft.writePos - ft.count + uint32(index)) % uint32(ft.capacity)
	}
	return &ft.buffer[pos]
}

// GetCurrentFrame returns the current frame entry being built.
func (ft *FrameTimeline) GetCurrentFrame() *FrameEntry {
	return ft.currentFrame.Load()
}

// Clear clears all frame entries.
func (ft *FrameTimeline) Clear() {
	ft.bufferMu.Lock()
	defer ft.bufferMu.Unlock()

	ft.writePos = 0
	ft.count = 0
	ft.currentFrame.Store(nil)
}

// SetMaxFrames sets the maximum number of frames to keep.
// P1-2: 需要重新分配缓冲区
func (ft *FrameTimeline) SetMaxFrames(n int) {
	// Ensure power of 2
	if n&(n-1) != 0 {
		n = 1 << (32 - nlz32(uint32(n)))
	}

	ft.bufferMu.Lock()
	defer ft.bufferMu.Unlock()

	if n == ft.capacity {
		return
	}

	// Create new buffer
	newBuffer := make([]FrameEntry, n)

	// Copy existing entries (up to n)
	oldCount := int(ft.count)
	copyCount := oldCount
	if copyCount > n {
		copyCount = n
	}

	for i := 0; i < copyCount; i++ {
		var oldPos uint32
		if ft.count < uint32(ft.capacity) {
			oldPos = uint32(i)
		} else {
			oldPos = (ft.writePos - ft.count + uint32(i)) % uint32(ft.capacity)
		}
		newBuffer[i] = ft.buffer[oldPos]
	}

	ft.buffer = newBuffer
	ft.capacity = n
	ft.writePos = uint32(copyCount)
	if uint32(n) < ft.count {
		ft.count = uint32(n)
	}
}

// GetStats returns statistics about the frame timeline.
func (ft *FrameTimeline) GetStats() *TimelineStats {
	frames := ft.GetAllFrames()

	stats := &TimelineStats{
		Enabled:      ft.IsEnabled(),
		FrameCount:   len(frames),
		MaxFrames:    ft.capacity,
		CurrentFrame: ft.currentFrame.Load(),
	}

	if len(frames) == 0 {
		return stats
	}

	// Calculate statistics
	var totalDuration, totalLayoutTime, totalPaintTime time.Duration
	var minDuration, maxDuration time.Duration = frames[0].Duration, frames[0].Duration

	for _, f := range frames {
		totalDuration += f.Duration
		totalLayoutTime += f.LayoutTime
		totalPaintTime += f.PaintTime

		if f.Duration < minDuration {
			minDuration = f.Duration
		}
		if f.Duration > maxDuration {
			maxDuration = f.Duration
		}
	}

	stats.AvgDuration = totalDuration / time.Duration(len(frames))
	stats.AvgLayoutTime = totalLayoutTime / time.Duration(len(frames))
	stats.AvgPaintTime = totalPaintTime / time.Duration(len(frames))
	stats.MinDuration = minDuration
	stats.MaxDuration = maxDuration

	// Calculate FPS
	if stats.AvgDuration > 0 {
		stats.AvgFPS = 1000 / stats.AvgDuration.Seconds()
	}

	return stats
}

// GetFrameTimeRange returns frames within a time range.
func (ft *FrameTimeline) GetFrameTimeRange(start, end time.Time) []*FrameEntry {
	frames := ft.GetAllFrames()

	var result []*FrameEntry
	for _, f := range frames {
		if (f.StartTime.Equal(start) || f.StartTime.After(start)) &&
			(f.StartTime.Equal(end) || f.StartTime.Before(end)) {
			result = append(result, f)
		}
	}
	return result
}

// GetSlowFrames returns frames with duration above threshold.
func (ft *FrameTimeline) GetSlowFrames(threshold time.Duration) []*FrameEntry {
	frames := ft.GetAllFrames()

	var result []*FrameEntry
	for _, f := range frames {
		if f.Duration >= threshold {
			result = append(result, f)
		}
	}
	return result
}

// TimelineStats contains statistics about the frame timeline.
type TimelineStats struct {
	Enabled       bool
	FrameCount    int
	MaxFrames     int
	CurrentFrame  *FrameEntry
	AvgDuration   time.Duration
	MinDuration   time.Duration
	MaxDuration   time.Duration
	AvgLayoutTime time.Duration
	AvgPaintTime  time.Duration
	AvgFPS        float64
}

// GetCapacity returns the current buffer capacity.
func (ft *FrameTimeline) GetCapacity() int {
	ft.bufferMu.RLock()
	defer ft.bufferMu.RUnlock()
	return ft.capacity
}

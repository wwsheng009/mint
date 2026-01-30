// Package devtools provides the frame timeline for DevTools.
//
// This file implements the FrameTimeline that provides a timeline view
// of frames with their causal relationships and performance metrics.
package devtools

import (
	"sync"
	"sync/atomic"
	"time"
)

// FrameTimeline provides a timeline view of frames with causal relationships.
type FrameTimeline struct {
	enabled atomic.Uint32

	// Frame history
	frames    []*FrameEntry
	framesMu  sync.RWMutex
	maxFrames int

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

// NewFrameTimeline creates a new frame timeline.
func NewFrameTimeline() *FrameTimeline {
	ft := &FrameTimeline{
		frames:    make([]*FrameEntry, 0, 100),
		maxFrames: 100, // Keep last 100 frames
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

	// Add to history
	ft.framesMu.Lock()
	ft.frames = append(ft.frames, entry)

	// Trim to maxFrames
	if len(ft.frames) > ft.maxFrames {
		ft.frames = ft.frames[1:]
	}
	ft.framesMu.Unlock()

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
	ft.framesMu.RLock()
	defer ft.framesMu.RUnlock()

	for _, f := range ft.frames {
		if f.FrameID == frameID {
			return f
		}
	}
	return nil
}

// GetAllFrames returns all frame entries.
func (ft *FrameTimeline) GetAllFrames() []*FrameEntry {
	ft.framesMu.RLock()
	defer ft.framesMu.RUnlock()

	frames := make([]*FrameEntry, len(ft.frames))
	copy(frames, ft.frames)
	return frames
}

// GetLastNFrames returns the last N frame entries.
func (ft *FrameTimeline) GetLastNFrames(n int) []*FrameEntry {
	ft.framesMu.RLock()
	defer ft.framesMu.RUnlock()

	count := len(ft.frames)
	if n > count {
		n = count
	}

	frames := make([]*FrameEntry, n)
	copy(frames, ft.frames[count-n:])
	return frames
}

// GetFrameCount returns the number of frames in the timeline.
func (ft *FrameTimeline) GetFrameCount() int {
	ft.framesMu.RLock()
	defer ft.framesMu.RUnlock()
	return len(ft.frames)
}

// GetFrameByIndex returns a frame entry by index.
func (ft *FrameTimeline) GetFrameByIndex(index int) *FrameEntry {
	ft.framesMu.RLock()
	defer ft.framesMu.RUnlock()

	if index < 0 || index >= len(ft.frames) {
		return nil
	}
	return ft.frames[index]
}

// GetCurrentFrame returns the current frame entry being built.
func (ft *FrameTimeline) GetCurrentFrame() *FrameEntry {
	return ft.currentFrame.Load()
}

// Clear clears all frame entries.
func (ft *FrameTimeline) Clear() {
	ft.framesMu.Lock()
	defer ft.framesMu.Unlock()

	ft.frames = make([]*FrameEntry, 0, ft.maxFrames)
	ft.currentFrame.Store(nil)
}

// SetMaxFrames sets the maximum number of frames to keep.
func (ft *FrameTimeline) SetMaxFrames(n int) {
	ft.framesMu.Lock()
	defer ft.framesMu.Unlock()

	ft.maxFrames = n

	// Trim if necessary
	if len(ft.frames) > n {
		ft.frames = ft.frames[len(ft.frames)-n:]
	}
}

// GetStats returns statistics about the frame timeline.
func (ft *FrameTimeline) GetStats() *TimelineStats {
	ft.framesMu.RLock()
	defer ft.framesMu.RUnlock()

	stats := &TimelineStats{
		Enabled:     ft.IsEnabled(),
		FrameCount:  len(ft.frames),
		MaxFrames:   ft.maxFrames,
		CurrentFrame: ft.currentFrame.Load(),
	}

	if len(ft.frames) == 0 {
		return stats
	}

	// Calculate statistics
	var totalDuration, totalLayoutTime, totalPaintTime time.Duration
	var minDuration, maxDuration time.Duration = ft.frames[0].Duration, ft.frames[0].Duration

	for _, f := range ft.frames {
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

	stats.AvgDuration = totalDuration / time.Duration(len(ft.frames))
	stats.AvgLayoutTime = totalLayoutTime / time.Duration(len(ft.frames))
	stats.AvgPaintTime = totalPaintTime / time.Duration(len(ft.frames))
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
	ft.framesMu.RLock()
	defer ft.framesMu.RUnlock()

	var result []*FrameEntry
	for _, f := range ft.frames {
		if (f.StartTime.Equal(start) || f.StartTime.After(start)) &&
			(f.StartTime.Equal(end) || f.StartTime.Before(end)) {
			result = append(result, f)
		}
	}
	return result
}

// GetSlowFrames returns frames with duration above threshold.
func (ft *FrameTimeline) GetSlowFrames(threshold time.Duration) []*FrameEntry {
	ft.framesMu.RLock()
	defer ft.framesMu.RUnlock()

	var result []*FrameEntry
	for _, f := range ft.frames {
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

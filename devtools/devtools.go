// Package devtools provides the TUI DevTools main entry point.
//
// This is the main interface for enabling and using DevTools in the Mint TUI Runtime.
// DevTools uses an incremental delta model to track changes with minimal performance impact.
//
// Usage:
//
//	// Initialize DevTools
//	dt := devtools.New()
//	dt.Enable()
//
//	// Collect data after each frame
//	dt.CollectLayout(layoutResult)
//	dt.EndFrame()
//
//	// When done
//	dt.Disable()
package devtools

import (
	"sync/atomic"

	"github.com/wwsheng009/mint/runtime"
)

// DevTools is the main DevTools instance.
// It coordinates all debug data collection and provides a single entry point
// for enabling/disabling debug features.
type DevTools struct {
	// Atomic flag for quick enable/disable check
	enabled uint32

	// Async collector
	asyncCollector *AsyncCollector

	// Output channel for debug messages
	outputCh chan *DebugMessage

	// Debug overlay
	debugOverlay *DebugOverlay

	// Configuration
	config Config
}

// New creates a new DevTools instance with default configuration.
func New() *DevTools {
	return NewWithConfig(DefaultConfig())
}

// NewWithConfig creates a new DevTools instance with the specified configuration.
func NewWithConfig(config Config) *DevTools {
	outputCh := make(chan *DebugMessage, 128)

	dt := &DevTools{
		enabled:        0,
		asyncCollector: NewAsyncCollector(outputCh),
		outputCh:       outputCh,
		config:         config,
	}

	if config.EnableOverlay {
		dt.debugOverlay = NewDebugOverlay(80, 24) // Default size
	}

	return dt
}

// Enable enables DevTools.
// This starts all background goroutines and begins collecting debug data.
func (dt *DevTools) Enable() {
	if atomic.CompareAndSwapUint32(&dt.enabled, 0, 1) {
		if dt.config.EnableMutationTap {
			EnableMutationTap()
		}
		dt.asyncCollector.Start()
	}
}

// Disable disables DevTools.
// This stops all background goroutines and stops collecting debug data.
func (dt *DevTools) Disable() {
	if atomic.CompareAndSwapUint32(&dt.enabled, 1, 0) {
		if dt.config.EnableMutationTap {
			DisableMutationTap()
		}
		dt.asyncCollector.Stop()
	}
}

// IsEnabled returns true if DevTools is enabled.
func (dt *DevTools) IsEnabled() bool {
	return atomic.LoadUint32(&dt.enabled) != 0
}

// CollectLayout collects layout data from the given layout result.
// This should be called after each layout pass.
func (dt *DevTools) CollectLayout(result *runtime.LayoutResult) {
	if !dt.IsEnabled() || result == nil {
		return
	}

	// Emit layout event to event bus
	dt.asyncCollector.GetEventBus().Emit(DebugEvent{
		Type:  EventLayout,
		Frame: int(dt.asyncCollector.GetCurrentFrame()),
	})
}

// CollectRepaint collects repaint data.
// This should be called after each repaint pass.
func (dt *DevTools) CollectRepaint(dirtyRegions []Rect, changedCells, totalCells int) {
	if !dt.IsEnabled() {
		return
	}

	// Emit repaint event to event bus
	dt.asyncCollector.GetEventBus().Emit(DebugEvent{
		Type:  EventRepaint,
		Frame: int(dt.asyncCollector.GetCurrentFrame()),
	})
}

// RecordEvent records a single event.
func (dt *DevTools) RecordEvent(eventType, targetID, phase string, data map[string]interface{}) {
	if !dt.IsEnabled() {
		return
	}
	dt.asyncCollector.GetEventCollector().RecordEvent(eventType, targetID, phase, data)
}

// BeginFrame marks the beginning of a new frame.
func (dt *DevTools) BeginFrame() {
	if !dt.IsEnabled() {
		return
	}
	dt.asyncCollector.BeginFrame()
}

// EndFrame marks the end of the current frame.
func (dt *DevTools) EndFrame() {
	if !dt.IsEnabled() {
		return
	}
	dt.asyncCollector.EndFrame()
}

// GetOutputChannel returns the channel for debug messages.
func (dt *DevTools) GetOutputChannel() <-chan *DebugMessage {
	return dt.outputCh
}

// GetOverlay returns the debug overlay.
func (dt *DevTools) GetOverlay() *DebugOverlay {
	return dt.debugOverlay
}

// Highlight highlights a component with a debug border.
func (dt *DevTools) Highlight(id string, x, y, w, h int) {
	if dt.debugOverlay != nil {
		dt.debugOverlay.Highlight(id, Rect{
			X:      x,
			Y:      y,
			Width:  w,
			Height: h,
		})
	}
}

// ClearOverlay clears the debug overlay.
func (dt *DevTools) ClearOverlay() {
	if dt.debugOverlay != nil {
		dt.debugOverlay.Clear()
	}
}

// GetEventBus returns the event bus for direct event emission.
func (dt *DevTools) GetEventBus() *EventBus {
	return dt.asyncCollector.GetEventBus()
}

// DebugOverlay is a simple debug overlay for highlighting components.
type DebugOverlay struct {
	shown map[string]bool
}

// NewDebugOverlay creates a new debug overlay.
func NewDebugOverlay(width, height int) *DebugOverlay {
	return &DebugOverlay{
		shown: make(map[string]bool),
	}
}

// Highlight highlights a component.
func (o *DebugOverlay) Highlight(id string, rect Rect) {
	o.shown[id] = true
}

// Clear clears all highlights.
func (o *DebugOverlay) Clear() {
	for id := range o.shown {
		delete(o.shown, id)
	}
}

// IsShown returns true if a component is highlighted.
func (o *DebugOverlay) IsShown(id string) bool {
	return o.shown[id]
}

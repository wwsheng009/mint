// Package observation provides intelligent analysis for DevTools.
//
// This file implements the WasteDetector for identifying unnecessary renders.
package observation

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/wwsheng009/mint/devtools"
)

// WasteSeverity indicates the severity level of waste.
type WasteSeverity int

const (
	WasteNone WasteSeverity = iota
	WasteLow    // 10-30% wasted renders
	WasteMedium // 30-50% wasted renders
	WasteHigh   // > 50% wasted renders
)

// String returns the string representation of the severity.
func (w WasteSeverity) String() string {
	switch w {
	case WasteLow:
		return "Low"
	case WasteMedium:
		return "Medium"
	case WasteHigh:
		return "High"
	default:
		return "None"
	}
}

// ComponentState tracks the state of a component for waste detection.
type ComponentState struct {
	LayoutVersion uint32
	ContentHash   uint64
	LastChanged   time.Time
	LastChecked   time.Time
}

// WasteFrame tracks waste metrics for a single frame.
type WasteFrame struct {
	FrameID       devtools.FrameID
	TotalNodes    int
	WastedLayouts int
	WastedPaints  int
	WastedNodes   []devtools.NodeID
	Timestamp     time.Time
}

// WasteReport tracks waste statistics for a component.
type WasteReport struct {
	NodeID        devtools.NodeID
	TotalRenders  uint64
	WastedRenders uint64
	WasteRate     float64
	LastWasteTime time.Time
	Severity      WasteSeverity
	LastUpdated   time.Time

	// Recent history (last 100 renders)
	recentWastes    []bool
	recentPos       int
	recentCount     int
}

// updateWasteRate updates the waste rate based on recent history.
func (wr *WasteReport) updateWasteRate() {
	if wr.recentCount == 0 {
		wr.WasteRate = 0
		return
	}

	wastedCount := 0
	for i := 0; i < wr.recentCount; i++ {
		if wr.recentWastes[i] {
			wastedCount++
		}
	}
	wr.WasteRate = float64(wastedCount) / float64(wr.recentCount) * 100

	// Update severity
	switch {
	case wr.WasteRate < 10:
		wr.Severity = WasteNone
	case wr.WasteRate < 30:
		wr.Severity = WasteLow
	case wr.WasteRate < 50:
		wr.Severity = WasteMedium
	default:
		wr.Severity = WasteHigh
	}
}

// recordWaste records a render event (wasted or not).
func (wr *WasteReport) recordWaste(isWasted bool) {
	if cap(wr.recentWastes) == 0 {
		wr.recentWastes = make([]bool, 100)
	}

	wr.recentWastes[wr.recentPos] = isWasted
	wr.recentPos = (wr.recentPos + 1) % 100
	if wr.recentCount < 100 {
		wr.recentCount++
	}

	wr.TotalRenders++
	if isWasted {
		wr.WastedRenders++
		wr.LastWasteTime = time.Now()
	}

	wr.updateWasteRate()
	wr.LastUpdated = time.Now()
}

// WasteDetector identifies unnecessary renders (waste).
type WasteDetector struct {
	mu     sync.RWMutex
	enabled atomic.Bool

	// Component state tracking
	lastComponentState map[devtools.NodeID]*ComponentState
	wasteReports       map[devtools.NodeID]*WasteReport

	// Frame tracking
	frameBuffer        *ringBuffer[WasteFrame]
	bufferSize         int

	// Configuration
	minWasteThreshold  int           // Minimum consecutive waste frames
	stateTTL           time.Duration // How long to remember state

	// Statistics
	totalRenders       atomic.Uint64
	totalWastedRenders atomic.Uint64

	// Callback
	onWasteDetected    func(*WasteReport)
}

// WasteConfig configures the WasteDetector.
type WasteConfig struct {
	MinWasteThreshold int           // Minimum waste frames to flag
	StateTTL          time.Duration // State time-to-live
	BufferSize        int
}

// DefaultWasteConfig returns the default configuration.
func DefaultWasteConfig() *WasteConfig {
	return &WasteConfig{
		MinWasteThreshold: 3,  // Flag after 3 consecutive wastes
		StateTTL:          5 * time.Minute,
		BufferSize:        128,
	}
}

// NewWasteDetector creates a new waste detector.
func NewWasteDetector(cfg *WasteConfig) *WasteDetector {
	if cfg == nil {
		cfg = DefaultWasteConfig()
	}

	return &WasteDetector{
		lastComponentState: make(map[devtools.NodeID]*ComponentState),
		wasteReports:       make(map[devtools.NodeID]*WasteReport),
		frameBuffer:        newRingBuffer[WasteFrame](cfg.BufferSize),
		bufferSize:         cfg.BufferSize,
		minWasteThreshold:  cfg.MinWasteThreshold,
		stateTTL:           cfg.StateTTL,
	}
}

// Enable enables the waste detector.
func (wd *WasteDetector) Enable() {
	wd.enabled.Store(true)
}

// Disable disables the waste detector.
func (wd *WasteDetector) Disable() {
	wd.enabled.Store(false)
}

// IsEnabled returns whether the detector is enabled.
func (wd *WasteDetector) IsEnabled() bool {
	return wd.enabled.Load()
}

// OnWasteDetected sets a callback for waste detection events.
func (wd *WasteDetector) OnWasteDetected(fn func(*WasteReport)) {
	wd.mu.Lock()
	defer wd.mu.Unlock()
	wd.onWasteDetected = fn
}

// ProcessLayout processes a layout update and detects waste.
func (wd *WasteDetector) ProcessLayout(nodeID devtools.NodeID, layoutVersion uint32, contentHash uint64) {
	if !wd.enabled.Load() {
		return
	}

	wd.totalRenders.Add(1)

	wd.mu.Lock()
	defer wd.mu.Unlock()

	now := time.Now()

	// Get or create component state
	state, exists := wd.lastComponentState[nodeID]
	if !exists {
		state = &ComponentState{
			LayoutVersion: layoutVersion,
			ContentHash:   contentHash,
			LastChanged:   now,
			LastChecked:   now,
		}
		wd.lastComponentState[nodeID] = state
		return // First render is not waste
	}

	// Check for waste (no actual change)
	isWasted := (state.LayoutVersion == layoutVersion) && (state.ContentHash == contentHash)

	// Update state
	state.LastChecked = now
	if !isWasted {
		state.LayoutVersion = layoutVersion
		state.ContentHash = contentHash
		state.LastChanged = now
	}

	// Get or create waste report
	report, exists := wd.wasteReports[nodeID]
	if !exists {
		report = &WasteReport{
			NodeID: nodeID,
		}
		wd.wasteReports[nodeID] = report
	}

	prevSeverity := report.Severity
	report.recordWaste(isWasted)

	if isWasted {
		wd.totalWastedRenders.Add(1)
	}

	// Trigger callback if severity changed or significant waste
	if report.Severity != WasteNone && (report.Severity != prevSeverity || wd.onWasteDetected != nil) {
		if wd.onWasteDetected != nil {
			wd.onWasteDetected(report)
		}
	}
}

// ProcessFrame processes a frame and aggregates waste.
func (wd *WasteDetector) ProcessFrame(entry *devtools.FrameEntry) {
	if !wd.enabled.Load() {
		return
	}

	wd.mu.Lock()
	defer wd.mu.Unlock()

	// Collect wasted nodes for this frame
	wastedNodes := make([]devtools.NodeID, 0)

	for nodeID, report := range wd.wasteReports {
		if report.Severity != WasteNone {
			wastedNodes = append(wastedNodes, nodeID)
		}
	}

	wasteFrame := &WasteFrame{
		FrameID:       entry.FrameID,
		TotalNodes:    len(wd.wasteReports),
		WastedLayouts: len(wastedNodes),
		WastedNodes:   wastedNodes,
		Timestamp:     time.Now(),
	}

	wd.frameBuffer.Push(*wasteFrame)

	// Cleanup old states
	wd.cleanup(time.Now())
}

// GetWasteReport returns the waste report for a component.
func (wd *WasteDetector) GetWasteReport(nodeID devtools.NodeID) *WasteReport {
	wd.mu.RLock()
	defer wd.mu.RUnlock()

	return wd.wasteReports[nodeID]
}

// GetWasteReports returns all waste reports sorted by severity.
func (wd *WasteDetector) GetWasteReports() []*WasteReport {
	wd.mu.RLock()
	defer wd.mu.RUnlock()

	result := make([]*WasteReport, 0, len(wd.wasteReports))
	for _, r := range wd.wasteReports {
		if r.Severity != WasteNone {
			result = append(result, r)
		}
	}

	// Sort by severity (high first) then by waste rate
	for i := 0; i < len(result)-1; i++ {
		for j := 0; j < len(result)-i-1; j++ {
			if result[j].Severity < result[j+1].Severity ||
				(result[j].Severity == result[j+1].Severity && result[j].WasteRate < result[j+1].WasteRate) {
				result[j], result[j+1] = result[j+1], result[j]
			}
		}
	}

	return result
}

// GetFrameWaste returns recent waste by frame.
func (wd *WasteDetector) GetFrameWaste(n int) []WasteFrame {
	return wd.frameBuffer.GetLastN(n)
}

// GetStats returns detector statistics.
func (wd *WasteDetector) GetStats() WasteStats {
	total := wd.totalRenders.Load()
	wasted := wd.totalWastedRenders.Load()

	var rate float64
	if total > 0 {
		rate = float64(wasted) / float64(total) * 100
	}

	wd.mu.RLock()
	reportCount := len(wd.wasteReports)
	wastingComponentCount := 0
	for _, r := range wd.wasteReports {
		if r.Severity != WasteNone {
			wastingComponentCount++
		}
	}
	wd.mu.RUnlock()

	return WasteStats{
		TotalRenders:         total,
		WastedRenders:        wasted,
		WasteRate:            rate,
		TotalComponents:      uint64(reportCount),
		WastingComponentCount: uint64(wastingComponentCount),
	}
}

// cleanup removes stale component states.
func (wd *WasteDetector) cleanup(now time.Time) {
	staleTime := now.Add(-wd.stateTTL)

	for nodeID, state := range wd.lastComponentState {
		if state.LastChecked.Before(staleTime) {
			delete(wd.lastComponentState, nodeID)
		}
	}
}

// Reset clears all waste data.
func (wd *WasteDetector) Reset() {
	wd.mu.Lock()
	defer wd.mu.Unlock()

	wd.lastComponentState = make(map[devtools.NodeID]*ComponentState)
	wd.wasteReports = make(map[devtools.NodeID]*WasteReport)
	wd.frameBuffer.Clear()
	wd.totalRenders.Store(0)
	wd.totalWastedRenders.Store(0)
}

// WasteStats contains statistics about waste detection.
type WasteStats struct {
	TotalRenders          uint64
	WastedRenders         uint64
	WasteRate             float64
	TotalComponents       uint64
	WastingComponentCount uint64
}

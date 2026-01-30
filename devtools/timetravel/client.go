// Package timetravel provides TUI client for time travel debugging.
//
// This file implements the TUI client interface for time travel.
package timetravel

import (
	"fmt"
	"strings"
	"sync"

	"github.com/wwsheng009/mint/devtools"
)

// TimeTravelClient provides a TUI interface for time travel debugging.
type TimeTravelClient struct {
	mu     sync.RWMutex
	mgr    *SnapshotManager
	cursor *TimeTravelCursor
	engine *ReplayEngine

	// UI state
	currentView ViewMode
	selectedFrame devtools.FrameID

	// Display options
	showLayout    bool
	showState     bool
	showCausal    bool
	showDiff      bool
}

// ViewMode represents the current view mode.
type ViewMode int

const (
	// ViewTimeline shows the timeline view.
	ViewTimeline ViewMode = iota
	// ViewSnapshot shows a single snapshot.
	ViewSnapshot
	// ViewDiff shows diff between snapshots.
	ViewDiff
	// ViewReplay shows replay controls.
	ViewReplay
)

// NewTimeTravelClient creates a new time travel client.
func NewTimeTravelClient(mgr *SnapshotManager) *TimeTravelClient {
	cursor := NewTimeTravelCursor(mgr)
	engine := NewReplayEngine(mgr, cursor)

	return &TimeTravelClient{
		mgr:          mgr,
		cursor:       cursor,
		engine:       engine,
		currentView:  ViewTimeline,
		showLayout:   true,
		showState:    true,
		showCausal:   true,
		showDiff:     true,
	}
}

// Render renders the current view to a string.
func (ttc *TimeTravelClient) Render(width, height int) string {
	ttc.mu.RLock()
	defer ttc.mu.RUnlock()

	var builder strings.Builder

	switch ttc.currentView {
	case ViewTimeline:
		builder.WriteString(ttc.renderTimeline(width, height))
	case ViewSnapshot:
		builder.WriteString(ttc.renderSnapshot(width, height))
	case ViewDiff:
		builder.WriteString(ttc.renderDiff(width, height))
	case ViewReplay:
		builder.WriteString(ttc.renderReplay(width, height))
	}

	return builder.String()
}

// renderTimeline renders the timeline view.
func (ttc *TimeTravelClient) renderTimeline(width, height int) string {
	var builder strings.Builder

	// Header
	builder.WriteString("╔════════════════════════════════════════════════╗\n")
	builder.WriteString("║          Time Travel - Timeline View           ║\n")
	builder.WriteString("╠════════════════════════════════════════════════╣\n")

	// Frame info
	info := ttc.cursor.GetInfo()
	builder.WriteString(fmt.Sprintf("║ Frame: %d/%d                                    ║\n",
		info.Index+1, info.TotalFrames))

	// Navigation hints
	builder.WriteString("╠════════════════════════════════════════════════╣\n")
	builder.WriteString("║ [n]ext [p]rev [f]irst [l]ast [j]ump [b]ookmark  ║\n")
	builder.WriteString("║ [s]napshot [d]iff [r]eplay [q]uit              ║\n")
	builder.WriteString("╚════════════════════════════════════════════════╝\n")

	// Current frame details
	if current := ttc.cursor.GetCurrent(); current != nil {
		builder.WriteString(fmt.Sprintf("\nCurrent Frame: %d\n", current.FrameID))
		builder.WriteString(fmt.Sprintf("Timestamp: %s\n", current.Timestamp.Format("15:04:05.000")))

		if current.CausalGraph != nil {
			summary := current.CausalGraph.GetFrameSummary()
			builder.WriteString(fmt.Sprintf("Events: %d | Mutations: %d | Layouts: %d | Repaints: %d\n",
				summary.EventCount, summary.MutationCount,
				summary.LayoutCount, summary.RepaintCount))
		}

		if ttc.showLayout && current.LayoutState != nil {
			builder.WriteString(fmt.Sprintf("\nLayout Nodes: %d\n", len(current.LayoutState.Nodes)))
		}

		if ttc.showState {
			builder.WriteString(fmt.Sprintf("Components: %d\n", len(current.ComponentStates)))
		}
	}

	// Bookmarks
	bookmarks := ttc.cursor.GetBookmarks()
	if len(bookmarks) > 0 {
		builder.WriteString("\nBookmarks: ")
		for i, bm := range bookmarks {
			if i > 0 {
				builder.WriteString(", ")
			}
			builder.WriteString(bm)
		}
		builder.WriteString("\n")
	}

	return builder.String()
}

// renderSnapshot renders the snapshot view.
func (ttc *TimeTravelClient) renderSnapshot(width, height int) string {
	var builder strings.Builder

	// Header
	builder.WriteString("╔════════════════════════════════════════════════╗\n")
	builder.WriteString("║          Time Travel - Snapshot View          ║\n")
	builder.WriteString("╠════════════════════════════════════════════════╣\n")

	current := ttc.cursor.GetCurrent()
	if current == nil {
		builder.WriteString("║ No snapshot selected                           ║\n")
		builder.WriteString("╚════════════════════════════════════════════════╝\n")
		return builder.String()
	}

	builder.WriteString(fmt.Sprintf("║ Frame: %d                                      ║\n", current.FrameID))
	builder.WriteString(fmt.Sprintf("║ Time: %s                               ║\n",
		current.Timestamp.Format("15:04:05.000")))
	builder.WriteString("╠════════════════════════════════════════════════╣\n")

	// Components
	if ttc.showState {
		builder.WriteString("║ Components:                                     ║\n")
		for compID, state := range current.ComponentStates {
			builder.WriteString(fmt.Sprintf("║   [%d] %s (%d fields)                      ║\n",
				compID, state.ComponentName, len(state.State)))
		}
	}

	// Layout
	if ttc.showLayout && current.LayoutState != nil {
		builder.WriteString("╠════════════════════════════════════════════════╣\n")
		builder.WriteString(fmt.Sprintf("║ Layout: %d nodes                               ║\n",
			len(current.LayoutState.Nodes)))
	}

	// Causal graph
	if ttc.showCausal && current.CausalGraph != nil {
		builder.WriteString("╠════════════════════════════════════════════════╣\n")
		summary := current.CausalGraph.GetFrameSummary()
		builder.WriteString(fmt.Sprintf("║ Causal: %d events, %d mutations                 ║\n",
			summary.EventCount, summary.MutationCount))
	}

	builder.WriteString("╚════════════════════════════════════════════════╝\n")

	return builder.String()
}

// renderDiff renders the diff view.
func (ttc *TimeTravelClient) renderDiff(width, height int) string {
	var builder strings.Builder

	// Header
	builder.WriteString("╔════════════════════════════════════════════════╗\n")
	builder.WriteString("║          Time Travel - Diff View               ║\n")
	builder.WriteString("╠════════════════════════════════════════════════╣\n")

	diff := ttc.cursor.GetDiffToNext()
	if diff == nil {
		builder.WriteString("║ No diff available                               ║\n")
		builder.WriteString("╚════════════════════════════════════════════════╝\n")
		return builder.String()
	}

	builder.WriteString(fmt.Sprintf("║ Diff: Frame %d -> %d                          ║\n",
		diff.FromFrame, diff.ToFrame))
	builder.WriteString("╠════════════════════════════════════════════════╣\n")

	// Component changes
	if len(diff.ChangedComponents) > 0 {
		builder.WriteString("║ Component Changes:                              ║\n")
		for _, comp := range diff.ChangedComponents {
			builder.WriteString(fmt.Sprintf("║   [%d] %d changes                               ║\n",
				comp.ComponentID, len(comp.Changes.Modified)))
		}
	}

	// Layout changes
	if len(diff.LayoutChanges) > 0 {
		builder.WriteString("╠════════════════════════════════════════════════╣\n")
		builder.WriteString(fmt.Sprintf("║ Layout Changes: %d                             ║\n",
			len(diff.LayoutChanges)))
	}

	// New events
	if len(diff.NewEvents) > 0 {
		builder.WriteString("╠════════════════════════════════════════════════╣\n")
		builder.WriteString(fmt.Sprintf("║ New Events: %d                                  ║\n",
			len(diff.NewEvents)))
		for _, event := range diff.NewEvents {
			builder.WriteString(fmt.Sprintf("║   %s on %s (%s)                         ║\n",
				event.Type, event.TargetID, event.Phase))
		}
	}

	builder.WriteString("╚════════════════════════════════════════════════╝\n")

	return builder.String()
}

// renderReplay renders the replay view.
func (ttc *TimeTravelClient) renderReplay(width, height int) string {
	var builder strings.Builder

	// Header
	builder.WriteString("╔════════════════════════════════════════════════╗\n")
	builder.WriteString("║          Time Travel - Replay Controls         ║\n")
	builder.WriteString("╠════════════════════════════════════════════════╣\n")

	// Status
	if ttc.engine.IsReplaying() {
		builder.WriteString("║ Status: REPLAYING                                ║\n")
	} else {
		builder.WriteString("║ Status: PAUSED                                  ║\n")
	}

	// Speed
	builder.WriteString(fmt.Sprintf("║ Speed: %.1fx                                     ║\n",
		ttc.engine.GetReplaySpeed()))

	// Current position
	info := ttc.cursor.GetInfo()
	builder.WriteString(fmt.Sprintf("║ Position: %d/%d                                  ║\n",
		info.Index+1, info.TotalFrames))

	builder.WriteString("╠════════════════════════════════════════════════╣\n")
	builder.WriteString("║ [space] play/pause  [+] speed up  [-] slow down  ║\n")
	builder.WriteString("║ [n] next frame     [p] prev frame  [s] stop     ║\n")
	builder.WriteString("║ [b] back to timeline                            ║\n")
	builder.WriteString("╚════════════════════════════════════════════════╝\n")

	// Progress bar
	if info.TotalFrames > 0 {
		progress := float64(info.Index+1) / float64(info.TotalFrames)
		barWidth := 50
		filled := int(progress * float64(barWidth))
		builder.WriteString("\n[")
		for i := 0; i < barWidth; i++ {
			if i < filled {
				builder.WriteString("█")
			} else {
				builder.WriteString("░")
			}
		}
		builder.WriteString(fmt.Sprintf("] %.0f%%\n", progress*100))
	}

	return builder.String()
}

// HandleInput handles user input.
func (ttc *TimeTravelClient) HandleInput(input string) bool {
	ttc.mu.Lock()
	defer ttc.mu.Unlock()

	switch input {
	case "n", "next":
		ttc.cursor.MoveNext()
	case "p", "prev":
		ttc.cursor.MovePrev()
	case "f", "first":
		ttc.cursor.MoveToFirst()
	case "l", "last":
		ttc.cursor.MoveToLast()
	case "j", "jump":
		// Would need additional input for frame ID
		ttc.cursor.MoveToLast()
	case "s", "snapshot":
		ttc.currentView = ViewSnapshot
	case "d", "diff":
		ttc.currentView = ViewDiff
	case "r", "replay":
		ttc.currentView = ViewReplay
	case "b", "back":
		ttc.currentView = ViewTimeline
	case "q", "quit":
		return false
	case "+":
		speed := ttc.engine.GetReplaySpeed()
		ttc.engine.SetReplaySpeed(speed * 1.5)
	case "-":
		speed := ttc.engine.GetReplaySpeed()
		ttc.engine.SetReplaySpeed(speed / 1.5)
	case " ":
		if ttc.engine.IsReplaying() {
			ttc.engine.Stop()
		} else {
			current := ttc.cursor.GetCurrent()
			if current != nil {
				ttc.engine.ReplayFrom(current.FrameID)
			}
		}
	}

	return true
}

// SetView sets the current view mode.
func (ttc *TimeTravelClient) SetView(view ViewMode) {
	ttc.mu.Lock()
	defer ttc.mu.Unlock()
	ttc.currentView = view
}

// GetView returns the current view mode.
func (ttc *TimeTravelClient) GetView() ViewMode {
	ttc.mu.RLock()
	defer ttc.mu.RUnlock()
	return ttc.currentView
}

// ToggleLayout toggles layout display.
func (ttc *TimeTravelClient) ToggleLayout() {
	ttc.mu.Lock()
	defer ttc.mu.Unlock()
	ttc.showLayout = !ttc.showLayout
}

// ToggleState toggles state display.
func (ttc *TimeTravelClient) ToggleState() {
	ttc.mu.Lock()
	defer ttc.mu.Unlock()
	ttc.showState = !ttc.showState
}

// ToggleCausal toggles causal graph display.
func (ttc *TimeTravelClient) ToggleCausal() {
	ttc.mu.Lock()
	defer ttc.mu.Unlock()
	ttc.showCausal = !ttc.showCausal
}

// ToggleDiff toggles diff display.
func (ttc *TimeTravelClient) ToggleDiff() {
	ttc.mu.Lock()
	defer ttc.mu.Unlock()
	ttc.showDiff = !ttc.showDiff
}

// GetCursor returns the time travel cursor.
func (ttc *TimeTravelClient) GetCursor() *TimeTravelCursor {
	ttc.mu.RLock()
	defer ttc.mu.RUnlock()
	return ttc.cursor
}

// GetEngine returns the replay engine.
func (ttc *TimeTravelClient) GetEngine() *ReplayEngine {
	ttc.mu.RLock()
	defer ttc.mu.RUnlock()
	return ttc.engine
}

// GetManager returns the snapshot manager.
func (ttc *TimeTravelClient) GetManager() *SnapshotManager {
	ttc.mu.RLock()
	defer ttc.mu.RUnlock()
	return ttc.mgr
}

// ClientConfig contains configuration for the time travel client.
type ClientConfig struct {
	Width        int
	Height       int
	ShowLayout   bool
	ShowState    bool
	ShowCausal   bool
	ShowDiff     bool
}

// NewClientWithConfig creates a client with the given configuration.
func NewClientWithConfig(mgr *SnapshotManager, config ClientConfig) *TimeTravelClient {
	client := NewTimeTravelClient(mgr)

	client.mu.Lock()
	client.showLayout = config.ShowLayout
	client.showState = config.ShowState
	client.showCausal = config.ShowCausal
	client.showDiff = config.ShowDiff
	client.mu.Unlock()

	return client
}

// ExportState exports the current client state.
func (ttc *TimeTravelClient) ExportState() map[string]interface{} {
	ttc.mu.RLock()
	defer ttc.mu.RUnlock()

	state := make(map[string]interface{})
	state["view"] = int(ttc.currentView)
	state["selectedFrame"] = ttc.selectedFrame
	state["showLayout"] = ttc.showLayout
	state["showState"] = ttc.showState
	state["showCausal"] = ttc.showCausal
	state["showDiff"] = ttc.showDiff

	if info := ttc.cursor.GetInfo(); info != nil {
		state["frameIndex"] = info.Index
		state["totalFrames"] = info.TotalFrames
	}

	return state
}

// ImportState imports client state.
func (ttc *TimeTravelClient) ImportState(state map[string]interface{}) error {
	ttc.mu.Lock()
	defer ttc.mu.Unlock()

	if view, ok := state["view"].(float64); ok {
		ttc.currentView = ViewMode(view)
	}

	if frameID, ok := state["selectedFrame"].(float64); ok {
		ttc.selectedFrame = devtools.FrameID(frameID)
	}

	if showLayout, ok := state["showLayout"].(bool); ok {
		ttc.showLayout = showLayout
	}

	if showState, ok := state["showState"].(bool); ok {
		ttc.showState = showState
	}

	if showCausal, ok := state["showCausal"].(bool); ok {
		ttc.showCausal = showCausal
	}

	if showDiff, ok := state["showDiff"].(bool); ok {
		ttc.showDiff = showDiff
	}

	return nil
}

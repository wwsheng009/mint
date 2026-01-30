// Package testing provides testing utilities for DevTools.
package testing

import (
	"sync"
	"testing"

	"github.com/wwsheng009/mint/devtools"
)

// MockRuntime simulates a Runtime for testing.
type MockRuntime struct {
	mu            sync.Mutex
	layoutResult  *devtools.LayoutResultAdapter
	events        []devtools.EventEntry
	componentTree map[string]interface{}
}

// NewMockRuntime creates a new mock runtime.
func NewMockRuntime() *MockRuntime {
	return &MockRuntime{
		events:        make([]devtools.EventEntry, 0),
		componentTree: make(map[string]interface{}),
	}
}

// Layout returns the mock layout result.
func (m *MockRuntime) Layout() *devtools.LayoutResultAdapter {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.layoutResult
}

// SetLayoutResult sets the mock layout result.
func (m *MockRuntime) SetLayoutResult(lr *devtools.LayoutResultAdapter) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.layoutResult = lr
}

// GetEvents returns all recorded events.
func (m *MockRuntime) GetEvents() []devtools.EventEntry {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.events
}

// AddEvent adds an event to the mock runtime.
func (m *MockRuntime) AddEvent(event devtools.EventEntry) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, event)
}

// ClearEvents clears all recorded events.
func (m *MockRuntime) ClearEvents() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = m.events[:0]
}

// MockDevTools wraps DevTools with test helpers.
type MockDevTools struct {
	*devtools.DevTools
	capturedEvents  []devtools.EventEntry
	capturedLayouts []*devtools.LayoutDelta
	causalGraphs    []*devtools.CausalGraph
	snapshots       []interface{}
	mu              sync.Mutex
}

// NewMockDevTools creates a new mock DevTools instance.
func NewMockDevTools() *MockDevTools {
	dt := devtools.New()
	dt.Enable()

	return &MockDevTools{
		DevTools:        dt,
		capturedEvents:  make([]devtools.EventEntry, 0),
		capturedLayouts: make([]*devtools.LayoutDelta, 0),
		causalGraphs:    make([]*devtools.CausalGraph, 0),
		snapshots:       make([]interface{}, 0),
	}
}

// CaptureEvent captures an event for testing.
func (m *MockDevTools) CaptureEvent(event devtools.EventEntry) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.capturedEvents = append(m.capturedEvents, event)
}

// CaptureLayout captures a layout delta for testing.
func (m *MockDevTools) CaptureLayout(delta *devtools.LayoutDelta) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.capturedLayouts = append(m.capturedLayouts, delta)
}

// GetCapturedEvents returns all captured events.
func (m *MockDevTools) GetCapturedEvents() []devtools.EventEntry {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.capturedEvents
}

// GetCapturedLayouts returns all captured layout deltas.
func (m *MockDevTools) GetCapturedLayouts() []*devtools.LayoutDelta {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.capturedLayouts
}

// AssertCausalChain asserts the causal chain has minimum counts.
func (m *MockDevTools) AssertCausalChain(t *testing.T, minEvents, minMutations int) {
	t.Helper()

	m.mu.Lock()
	graphs := m.causalGraphs
	m.mu.Unlock()

	if len(graphs) == 0 {
		t.Fatal("No causal graphs recorded")
	}

	graph := graphs[len(graphs)-1]
	summary := graph.GetFrameSummary()

	if summary.EventCount < minEvents {
		t.Errorf("Expected at least %d events, got %d", minEvents, summary.EventCount)
	}

	if summary.MutationCount < minMutations {
		t.Errorf("Expected at least %d mutations, got %d", minMutations, summary.MutationCount)
	}
}

// Reset clears all captured data.
func (m *MockDevTools) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.capturedEvents = m.capturedEvents[:0]
	m.capturedLayouts = m.capturedLayouts[:0]
	m.causalGraphs = m.causalGraphs[:0]
	m.snapshots = m.snapshots[:0]
}

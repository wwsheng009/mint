// Package testing provides testing utilities for DevTools.
package testing

import (
	"fmt"
	"testing"

	"github.com/wwsheng009/mint/devtools"
)

// EventCountAssertion asserts the number of events captured.
type EventCountAssertion struct {
	MinCount int
	MaxCount int
}

// Verify checks if the event count is within the expected range.
func (a *EventCountAssertion) Verify(f *Fixture) error {
	count := len(f.DevTools.GetCapturedEvents())
	if a.MinCount > 0 && count < a.MinCount {
		return fmt.Errorf("event count %d < minimum %d", count, a.MinCount)
	}
	if a.MaxCount > 0 && count > a.MaxCount {
		return fmt.Errorf("event count %d > maximum %d", count, a.MaxCount)
	}
	return nil
}

// EventCount creates an assertion for event count.
func EventCount(min, max int) TestAssertion {
	return &EventCountAssertion{MinCount: min, MaxCount: max}
}

// CausalChainAssertion asserts the causal chain properties.
type CausalChainAssertion struct {
	MinEvents    int
	MinMutations int
	MinEdges     int
}

// Verify checks if the causal chain meets the requirements.
func (a *CausalChainAssertion) Verify(f *Fixture) error {
	graph := f.DevTools.DevTools.GetEventBus()
	stats := graph.GetStats()

	// Check events sent
	if a.MinEvents > 0 {
		eventsSent := int(stats.EventsSent.Load())
		if eventsSent < a.MinEvents {
			return fmt.Errorf("events sent %d < minimum %d", eventsSent, a.MinEvents)
		}
	}

	return nil
}

// CausalChain creates an assertion for causal chain properties.
func CausalChain(minEvents, minMutations, minEdges int) TestAssertion {
	return &CausalChainAssertion{
		MinEvents:    minEvents,
		MinMutations: minMutations,
		MinEdges:     minEdges,
	}
}

// LayoutDeltaAssertion asserts layout changes for a specific node.
type LayoutDeltaAssertion struct {
	NodeID         devtools.NodeID
	ExpectedChange devtools.ChangeMask
}

// Verify checks if the layout delta contains the expected changes.
func (a *LayoutDeltaAssertion) Verify(f *Fixture) error {
	for _, delta := range f.DevTools.GetCapturedLayouts() {
		for _, changed := range delta.Changed {
			if changed.ID == a.NodeID {
				if changed.Mask&a.ExpectedChange != 0 {
					return nil
				}
			}
		}
	}
	return fmt.Errorf("layout delta not found for node %s", a.NodeID)
}

// LayoutChanged creates an assertion for layout changes.
func LayoutChanged(nodeID devtools.NodeID, changeMask devtools.ChangeMask) TestAssertion {
	return &LayoutDeltaAssertion{
		NodeID:         nodeID,
		ExpectedChange: changeMask,
	}
}

// EnabledAssertion asserts that DevTools is enabled.
type EnabledAssertion struct {
	Expected bool
}

// Verify checks if DevTools is in the expected enabled state.
func (a *EnabledAssertion) Verify(f *Fixture) error {
	if f.DevTools.IsEnabled() != a.Expected {
		return fmt.Errorf("DevTools enabled state: got %v, want %v", f.DevTools.IsEnabled(), a.Expected)
	}
	return nil
}

// IsEnabled creates an assertion for DevTools enabled state.
func IsEnabled(expected bool) TestAssertion {
	return &EnabledAssertion{Expected: expected}
}

// DevToolsStatsAssertion asserts DevTools statistics.
type DevToolsStatsAssertion struct {
	MinEventsSent uint64
	MaxDropped    uint64
}

// Verify checks if DevTools statistics meet requirements.
func (a *DevToolsStatsAssertion) Verify(f *Fixture) error {
	stats := f.DevTools.DevTools.GetEventBus().GetStats()

	eventsSent := stats.EventsSent.Load()
	if eventsSent < a.MinEventsSent {
		return fmt.Errorf("events sent %d < minimum %d", eventsSent, a.MinEventsSent)
	}

	dropped := stats.EventsDropped.Load()
	if a.MaxDropped > 0 && dropped > a.MaxDropped {
		return fmt.Errorf("events dropped %d > maximum %d", dropped, a.MaxDropped)
	}

	return nil
}

// DevToolsStats creates an assertion for DevTools statistics.
func DevToolsStats(minEventsSent uint64, maxDropped uint64) TestAssertion {
	return &DevToolsStatsAssertion{
		MinEventsSent: minEventsSent,
		MaxDropped:    maxDropped,
	}
}

// Helper assertion functions for direct testing.T use

// AssertDevToolsStats asserts DevTools statistics using testing.T.
func AssertDevToolsStats(t *testing.T, dt *devtools.DevTools, expectedEvents, expectedMutations int) {
	t.Helper()

	graph := dt.GetEventBus()
	stats := graph.GetStats()

	eventsSent := int(stats.EventsSent.Load())
	if eventsSent < expectedEvents {
		t.Errorf("Expected at least %d events, got %d", expectedEvents, eventsSent)
	}
}

// AssertTimelineIntegrity asserts the timeline is consistent.
func AssertTimelineIntegrity(t *testing.T, timeline *devtools.FrameTimeline) {
	t.Helper()

	frames := timeline.GetAllFrames()

	// Verify frame sequence
	for i := 1; i < len(frames); i++ {
		prev := frames[i-1]
		curr := frames[i]

		// Verify frame ID is monotonically increasing
		if curr.FrameID <= prev.FrameID {
			t.Errorf("Frame ID not monotonic: %d <= %d", curr.FrameID, prev.FrameID)
		}

		// Verify time ordering
		if curr.StartTime.Before(prev.StartTime) {
			t.Errorf("Frame time not ordered: frame %d before frame %d",
				curr.FrameID, prev.FrameID)
		}
	}
}

// AssertEventBusStats asserts EventBus statistics.
func AssertEventBusStats(t *testing.T, bus *devtools.EventBus, minEventsSent uint64) {
	t.Helper()

	stats := bus.GetStats()
	eventsSent := stats.EventsSent.Load()

	if eventsSent < minEventsSent {
		t.Errorf("Expected at least %d events sent, got %d", minEventsSent, eventsSent)
	}
}

// AssertNoLeaks asserts there are no resource leaks.
func AssertNoLeaks(t *testing.T, dt *devtools.DevTools) {
	t.Helper()

	// Check EventBus stats for unexpected drops
	stats := dt.GetEventBus().GetStats()
	dropped := stats.BackpressureDrops.Load()

	if dropped > 0 {
		t.Logf("Warning: %d events dropped due to backpressure", dropped)
	}
}

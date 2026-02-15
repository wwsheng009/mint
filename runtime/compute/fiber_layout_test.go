// Package compute provides tests for Fiber-first layout
package compute

import (
	"testing"

	"github.com/wwsheng009/mint/runtime"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// =============================================================================
// Fiber Layout Helper Tests (Phase 1)
// =============================================================================

// TestFiberLayoutMethods verifies Fiber layout helper methods work correctly
func TestFiberLayoutMethods(t *testing.T) {
	// Create a simple Fiber tree
	root := &rtui.Fiber{
		NodeID: 1,
		Child:  &rtui.Fiber{
			NodeID: 2,
			Sibling: &rtui.Fiber{
				NodeID: 3,
				Sibling: &rtui.Fiber{
					NodeID: 4,
				},
			},
		},
	}

	// Test GetChildFibers
	children := root.GetChildFibers()
	if len(children) != 3 {
		t.Errorf("Expected 3 children, got %d", len(children))
	}

	// Test GetChildCount
	count := root.GetChildCount()
	if count != 3 {
		t.Errorf("Expected 3 children from GetChildCount, got %d", count)
	}
}

// TestFiberLayoutFields verifies Fiber layout fields are stored
func TestFiberLayoutFields(t *testing.T) {
	// Create a Fiber with layout properties
	fiber := &rtui.Fiber{
		NodeID:         1,
		LayoutDirection: rtui.DirectionRow,
		LayoutAlign:     rtui.AlignCenter,
		LayoutCrossAlign: rtui.AlignEnd,
		LayoutGap:       2,
		LayoutPadding:    [4]int{1, 2, 3, 4},
		LayoutFlex:      3,
	}

	// Verify GetDirection
	if dir := fiber.GetDirection(); dir != rtui.DirectionRow {
		t.Errorf("Expected Direction=Row, got %v", dir)
	}

	// Verify GetAlign
	if align := fiber.GetAlign(); align != rtui.AlignCenter {
		t.Errorf("Expected Align=Center, got %v", align)
	}

	// Verify GetCrossAlign
	if crossAlign := fiber.GetCrossAlign(); crossAlign != rtui.AlignEnd {
		t.Errorf("Expected CrossAlign=End, got %v", crossAlign)
	}

	// Verify GetGap
	if gap := fiber.GetGap(); gap != 2 {
		t.Errorf("Expected Gap=2, got %d", gap)
	}

	// Verify GetPadding
	padding := fiber.GetPadding()
	expectedPadding := [4]int{1, 2, 3, 4}
	if padding != expectedPadding {
		t.Errorf("Expected Padding=%v, got %v", expectedPadding, padding)
	}

	// Verify GetFlex
	if flex := fiber.GetFlex(); flex != 3 {
		t.Errorf("Expected Flex=3, got %d", flex)
	}
}

// =============================================================================
// Engine Fiber Layout Tests
// =============================================================================

// TestGetLayoutInfoFromFiber verifies layout info extraction from Fiber
func TestGetLayoutInfoFromFiber(t *testing.T) {
	engine := NewEngine()

	// Create a Fiber with layout properties
	fiber := &rtui.Fiber{
		NodeID:         1,
		LayoutDirection: rtui.DirectionRow,
		LayoutAlign:     rtui.AlignCenter,
		LayoutCrossAlign: rtui.AlignEnd,
		LayoutGap:       2,
		LayoutPadding:    [4]int{1, 2, 3, 4},
		LayoutFlex:      3,
	}

	// Extract layout info
	info := engine.getLayoutInfoFromFiber(fiber)

	// Verify
	if info.IsHorizontal != true {
		t.Errorf("Expected IsHorizontal=true, got %v", info.IsHorizontal)
	}
	if info.Align != rtui.AlignCenter {
		t.Errorf("Expected Align=Center, got %v", info.Align)
	}
	if info.CrossAlign != rtui.AlignEnd {
		t.Errorf("Expected CrossAlign=End, got %v", info.CrossAlign)
	}
	if info.Gap != 2 {
		t.Errorf("Expected Gap=2, got %d", info.Gap)
	}
	if info.Flex != 3 {
		t.Errorf("Expected Flex=3, got %d", info.Flex)
	}
	expectedPadding := [4]int{1, 2, 3, 4}
	if info.Padding != expectedPadding {
		t.Errorf("Expected Padding=%v, got %v", expectedPadding, info.Padding)
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

// NewTestConstraints creates BoxConstraints for testing
func NewTestConstraints(minW, minH, maxW, maxH int) runtime.BoxConstraints {
	return runtime.BoxConstraints{
		MinWidth:  minW,
		MinHeight: minH,
		MaxWidth:  maxW,
		MaxHeight: maxH,
	}
}

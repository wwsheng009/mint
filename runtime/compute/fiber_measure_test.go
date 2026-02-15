// Package compute tests Fiber-first layout algorithm.
// Simple test to verify Fiber.MeasureLayout works correctly.
package compute

import (
	"testing"

	"github.com/wwsheng009/mint/runtime"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// TestFiberMeasureLayout tests Fiber.MeasureLayout method directly.
func TestFiberMeasureLayout(t *testing.T) {
	// Create a simple Fiber tree
	// Root (HStack) with one text child

	root := &rtui.Fiber{
		Type:   rtui.VNodeElement,
		Tag:    "hstack",
		NodeID:  1,
		Layer:  rtui.LayerBase,
		// Layout style
		LayoutDirection: rtui.DirectionRow,
		LayoutGap:       1,
		LayoutPadding:    [4]int{0, 0, 0, 0},
	}

	child := &rtui.Fiber{
		Type:         rtui.VNodeText,
		Tag:          "text",
		NodeID:       2,
		Layer:        rtui.LayerBase,
		MemoizedState: "Hi", // Text content
	}

	// Link children
	root.Child = child
	child.Return = root

	// Create measurer
	measurer := &testFiberChildMeasurer{}

	// Test MeasureLayout
	constraints := runtime.BoxConstraints{
		MinWidth:  0,
		MaxWidth:  runtime.Infinity,
		MinHeight: 0,
		MaxHeight: runtime.Infinity,
	}

	measurement := root.MeasureLayout(measurer, constraints)

	t.Logf("Measurement.Size = %v", measurement.Size)
	t.Logf("ChildConstraints = %v", measurement.ChildConstraints)

	// Verify size is not zero
	if measurement.Size.Width == 0 || measurement.Size.Height == 0 {
		t.Errorf("Expected non-zero size, got %v", measurement.Size)
	}

	// Verify child constraints
	if len(measurement.ChildConstraints) != 1 {
		t.Errorf("Expected 1 child constraint, got %d", len(measurement.ChildConstraints))
	}

	childConstraint := measurement.ChildConstraints[0]
	t.Logf("Child constraint: %v", childConstraint)

	// Child should have at least enough width for "Hi" (2 chars)
	if childConstraint.MaxWidth < 2 && childConstraint.MaxWidth != runtime.Infinity {
		t.Errorf("Child constraint too small: %d", childConstraint.MaxWidth)
	}
}

// testFiberChildMeasurer implements runtime.ChildMeasurer for testing.
type testFiberChildMeasurer struct{}

// MeasureChild measures a child Fiber's size.
func (m *testFiberChildMeasurer) MeasureChild(child interface{}, constraints runtime.BoxConstraints) runtime.Size {
	fiber, ok := child.(*rtui.Fiber)
	if !ok || fiber == nil {
		return runtime.Size{Width: 0, Height: 0}
	}

	// For text, width = len(text)
	text, ok := fiber.MemoizedState.(string)
	if !ok || text == "" {
		return runtime.Size{Width: 0, Height: 1}
	}

	runes := []rune(text)
	width := len(runes)
	height := 1

	return runtime.Size{Width: width, Height: height}
}

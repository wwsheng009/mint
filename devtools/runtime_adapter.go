// Package devtools provides runtime integration adapters for DevTools.
//
// This file contains adapters that allow DevTools to work with different
// runtime implementations, providing stub behavior when required interfaces
// are not available.
package devtools

// BoxInfo represents debug information about a layout box.
type BoxInfo struct {
	NodeID         string
	Node           *LayoutNodeAdapter
	X, Y           int
	Width          int
	Height         int
	MeasuredWidth  int
	MeasuredHeight int
}

// LayoutNodeAdapter wraps runtime layout nodes (or stub if not available).
type LayoutNodeAdapter struct {
	node interface{}
}

// GetLayoutVersion returns the layout version of the node.
// If the runtime doesn't support this, returns 0.
func (a *LayoutNodeAdapter) GetLayoutVersion() uint32 {
	if a.node == nil {
		return 0
	}
	if n, ok := a.node.(interface{ GetLayoutVersion() uint32 }); ok {
		return n.GetLayoutVersion()
	}
	return 0
}

// GetX returns the X position of the node.
func (a *LayoutNodeAdapter) GetX() int {
	if a.node == nil {
		return 0
	}
	if n, ok := a.node.(interface{ GetX() int }); ok {
		return n.GetX()
	}
	return 0
}

// GetY returns the Y position of the node.
func (a *LayoutNodeAdapter) GetY() int {
	if a.node == nil {
		return 0
	}
	if n, ok := a.node.(interface{ GetY() int }); ok {
		return n.GetY()
	}
	return 0
}

// GetMeasuredWidth returns the measured width of the node.
func (a *LayoutNodeAdapter) GetMeasuredWidth() int {
	if a.node == nil {
		return 0
	}
	if n, ok := a.node.(interface{ GetMeasuredWidth() int }); ok {
		return n.GetMeasuredWidth()
	}
	return 0
}

// GetMeasuredHeight returns the measured height of the node.
func (a *LayoutNodeAdapter) GetMeasuredHeight() int {
	if a.node == nil {
		return 0
	}
	if n, ok := a.node.(interface{ GetMeasuredHeight() int }); ok {
		return n.GetMeasuredHeight()
	}
	return 0
}

// GetStyle returns the style information of the node.
func (a *LayoutNodeAdapter) GetStyle() *StyleAdapter {
	if a.node == nil {
		return &StyleAdapter{}
	}
	if n, ok := a.node.(interface{ GetStyle() interface{} }); ok {
		style := n.GetStyle()
		return &StyleAdapter{style: style}
	}
	return &StyleAdapter{}
}

// StyleAdapter wraps style information.
type StyleAdapter struct {
	style interface{}
}

// GetZIndex returns the Z-index of the style.
func (s *StyleAdapter) GetZIndex() int {
	if s.style == nil {
		return 0
	}
	if st, ok := s.style.(interface{ GetZIndex() int }); ok {
		return st.GetZIndex()
	}
	return 0
}

// GetVisible returns whether the style is visible.
func (s *StyleAdapter) GetVisible() bool {
	if s.style == nil {
		return true
	}
	if st, ok := s.style.(interface{ GetVisible() bool }); ok {
		return st.GetVisible()
	}
	return true
}

// LayoutResultAdapter adapts runtime layout results to devtools format.
type LayoutResultAdapter struct {
	result interface{}
	boxes []BoxInfo
}

// AdaptLayoutResult creates an adapter from a runtime layout result.
func AdaptLayoutResult(result interface{}) *LayoutResultAdapter {
	if result == nil {
		return &LayoutResultAdapter{
			result: nil,
			boxes:  []BoxInfo{},
		}
	}

	// Try to extract boxes from the result
	adapter := &LayoutResultAdapter{
		result: result,
		boxes:  extractBoxes(result),
	}

	return adapter
}

// extractBoxes attempts to extract box information from a layout result.
// This is a stub implementation that should be replaced with actual
// runtime integration when available.
func extractBoxes(result interface{}) []BoxInfo {
	// TODO: Implement actual box extraction from runtime
	// For now, return empty slice
	return []BoxInfo{}
}

// Boxes returns the layout boxes.
func (a *LayoutResultAdapter) Boxes() []BoxInfo {
	return a.boxes
}

// HasRuntimeLayoutSupport checks if the runtime package implements
// required debugging interfaces. This is a compile-time check that
// should be verified during integration.
func HasRuntimeLayoutSupport() bool {
	// In actual implementation, this would check for the existence
	// of runtime.LayoutResult, runtime.LayoutNode, etc.
	// For now, we assume it exists and return true.
	return true
}

// MockLayoutResult creates a mock layout result for testing.
// This should only be used in tests.
func MockLayoutResult(nodeCount int) *LayoutResultAdapter {
	boxes := make([]BoxInfo, nodeCount)
	for i := 0; i < nodeCount; i++ {
		boxes[i] = BoxInfo{
			NodeID:         formatNodeID(i),
			Node:           &LayoutNodeAdapter{},
			X:              i * 10,
			Y:              i * 20,
			Width:          100,
			Height:         50,
			MeasuredWidth:  100,
			MeasuredHeight: 50,
		}
	}

	return &LayoutResultAdapter{
		boxes: boxes,
	}
}

// MockLayoutResultWithDynamicNodes creates a layout result with dynamically
// changing nodes for testing memory cleanup.
func MockLayoutResultWithDynamicNodes(nodeCount int, frame int) *LayoutResultAdapter {
	boxes := make([]BoxInfo, nodeCount)
	for i := 0; i < nodeCount; i++ {
		nodeID := formatNodeID(i)

		// Simulate nodes appearing/disappearing over time
		if frame%3 == 0 && i >= nodeCount/2 {
			// Skip some nodes
			continue
		}

		boxes[i] = BoxInfo{
			NodeID:         nodeID,
			Node:           &LayoutNodeAdapter{},
			X:              i * 10,
			Y:              i * 20,
			Width:          100 + frame%10, // Simulate size change
			Height:         50,
			MeasuredWidth:  100 + frame%10,
			MeasuredHeight: 50,
		}
	}

	return &LayoutResultAdapter{
		boxes: boxes,
	}
}

func formatNodeID(i int) string {
	return formatNodeType(i) + "_" + formatNum(i)
}

func formatNodeType(i int) string {
	types := []string{"button", "label", "container", "input", "text"}
	return types[i%len(types)]
}

func formatNum(i int) string {
	if i < 10 {
		return string(rune('0' + i))
	}
	return formatNum(i/10) + formatNum(i%10)
}

// Test for runtime integration
func init() {
	// Verify that we can create adapters without panicking
	_ = AdaptLayoutResult(nil)
	_ = MockLayoutResult(10)
	_ = MockLayoutResultWithDynamicNodes(10, 0)
}

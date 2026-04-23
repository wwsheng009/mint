// Package devtools provides runtime integration adapters for DevTools.
//
// This file contains adapters that allow DevTools to work with different
// runtime implementations, providing stub behavior when required interfaces
// are not available.
package devtools

import (
	"github.com/wwsheng009/mint/runtime"
	runtimelayout "github.com/wwsheng009/mint/runtime/layout"
)

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
	boxes  []BoxInfo
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
func extractBoxes(result interface{}) []BoxInfo {
	switch v := result.(type) {
	case *runtime.LayoutResult:
		return extractLegacyRuntimeBoxes(v)
	case runtime.LayoutResult:
		return extractLegacyRuntimeBoxes(&v)
	case *runtimelayout.LayoutResult:
		return extractRuntimeLayoutBoxes(v)
	case runtimelayout.LayoutResult:
		return extractRuntimeLayoutBoxes(&v)
	case LayoutDebugView:
		return extractDebugViewBoxes(v)
	default:
		return []BoxInfo{}
	}
}

type layoutNodeSnapshot struct {
	layoutVersion  uint32
	x              int
	y              int
	measuredWidth  int
	measuredHeight int
	style          styleSnapshot
}

func (s layoutNodeSnapshot) GetLayoutVersion() uint32 {
	return s.layoutVersion
}

func (s layoutNodeSnapshot) GetX() int {
	return s.x
}

func (s layoutNodeSnapshot) GetY() int {
	return s.y
}

func (s layoutNodeSnapshot) GetMeasuredWidth() int {
	return s.measuredWidth
}

func (s layoutNodeSnapshot) GetMeasuredHeight() int {
	return s.measuredHeight
}

func (s layoutNodeSnapshot) GetStyle() interface{} {
	return s.style
}

type styleSnapshot struct {
	zIndex  int
	visible bool
}

func (s styleSnapshot) GetZIndex() int {
	return s.zIndex
}

func (s styleSnapshot) GetVisible() bool {
	return s.visible
}

func extractLegacyRuntimeBoxes(result *runtime.LayoutResult) []BoxInfo {
	if result == nil || len(result.Boxes) == 0 {
		return []BoxInfo{}
	}

	boxes := make([]BoxInfo, 0, len(result.Boxes))
	for i := range result.Boxes {
		boxes = append(boxes, boxInfoFromLegacyRuntimeBox(&result.Boxes[i]))
	}
	return boxes
}

func boxInfoFromLegacyRuntimeBox(box *runtime.LayoutBox) BoxInfo {
	nodeID := legacyRuntimeBoxID(box)
	snapshot := snapshotFromLegacyRuntimeBox(nodeID, box)

	return BoxInfo{
		NodeID:         nodeID,
		Node:           &LayoutNodeAdapter{node: snapshot},
		X:              snapshot.x,
		Y:              snapshot.y,
		Width:          snapshot.measuredWidth,
		Height:         snapshot.measuredHeight,
		MeasuredWidth:  snapshot.measuredWidth,
		MeasuredHeight: snapshot.measuredHeight,
	}
}

func legacyRuntimeBoxID(box *runtime.LayoutBox) string {
	if box == nil {
		return ""
	}
	if box.NodeID != "" {
		return box.NodeID
	}
	if box.Node != nil {
		return box.Node.ID
	}
	return ""
}

func snapshotFromLegacyRuntimeBox(nodeID string, box *runtime.LayoutBox) layoutNodeSnapshot {
	x, y := legacyRuntimeBoxPosition(box)
	width, height := legacyRuntimeBoxSize(box)
	zIndex := 0
	version := syntheticLayoutVersion(nodeID, x, y, width, height, zIndex, true)

	if box != nil {
		zIndex = box.ZIndex
		version = syntheticLayoutVersion(nodeID, x, y, width, height, zIndex, true)
	}

	if box != nil && box.Node != nil {
		zIndex = box.Node.Style.ZIndex
		version = box.Node.GetLayoutVersion()
		if version == 0 {
			version = syntheticLayoutVersion(nodeID, x, y, width, height, zIndex, true)
		}
	}

	return layoutNodeSnapshot{
		layoutVersion:  version,
		x:              x,
		y:              y,
		measuredWidth:  width,
		measuredHeight: height,
		style: styleSnapshot{
			zIndex:  zIndex,
			visible: true,
		},
	}
}

func legacyRuntimeBoxPosition(box *runtime.LayoutBox) (int, int) {
	if box == nil {
		return 0, 0
	}
	if box.Node != nil {
		return runtimeNodePosition(box.Node)
	}
	return box.X, box.Y
}

func runtimeNodePosition(node *runtime.LayoutNode) (int, int) {
	if node == nil {
		return 0, 0
	}
	if node.AbsoluteX != 0 || node.AbsoluteY != 0 || (node.X == 0 && node.Y == 0) {
		return node.AbsoluteX, node.AbsoluteY
	}
	return node.X, node.Y
}

func legacyRuntimeBoxSize(box *runtime.LayoutBox) (int, int) {
	if box == nil {
		return 0, 0
	}
	if box.Node != nil {
		return box.Node.MeasuredWidth, box.Node.MeasuredHeight
	}
	return box.W, box.H
}

func extractRuntimeLayoutBoxes(result *runtimelayout.LayoutResult) []BoxInfo {
	if result == nil {
		return []BoxInfo{}
	}

	if result.Root != nil {
		boxes := make([]BoxInfo, 0, len(result.Boxes))
		appendRuntimeLayoutBoxInfo(&boxes, result.Root)
		return boxes
	}

	if len(result.Boxes) == 0 {
		return []BoxInfo{}
	}

	boxes := make([]BoxInfo, 0, len(result.Boxes))
	for i := range result.Boxes {
		boxes = append(boxes, boxInfoFromRuntimeLayoutBox(&result.Boxes[i]))
	}
	return boxes
}

func appendRuntimeLayoutBoxInfo(boxes *[]BoxInfo, box *runtimelayout.LayoutBox) {
	if box == nil {
		return
	}

	*boxes = append(*boxes, boxInfoFromRuntimeLayoutBox(box))
	for _, child := range box.Children {
		appendRuntimeLayoutBoxInfo(boxes, child)
	}
}

func boxInfoFromRuntimeLayoutBox(box *runtimelayout.LayoutBox) BoxInfo {
	nodeID := runtimeLayoutBoxID(box)
	snapshot := snapshotFromRuntimeLayoutBox(nodeID, box)

	return BoxInfo{
		NodeID:         nodeID,
		Node:           &LayoutNodeAdapter{node: snapshot},
		X:              snapshot.x,
		Y:              snapshot.y,
		Width:          snapshot.measuredWidth,
		Height:         snapshot.measuredHeight,
		MeasuredWidth:  snapshot.measuredWidth,
		MeasuredHeight: snapshot.measuredHeight,
	}
}

func runtimeLayoutBoxID(box *runtimelayout.LayoutBox) string {
	if box == nil {
		return ""
	}
	if box.ID != "" {
		return box.ID
	}
	return box.PropsID
}

func snapshotFromRuntimeLayoutBox(nodeID string, box *runtimelayout.LayoutBox) layoutNodeSnapshot {
	x, y := runtimeLayoutBoxPosition(box)
	width, height := 0, 0
	zIndex := 0

	if box != nil {
		width = box.Width
		height = box.Height
		zIndex = box.ZIndex
	}

	return layoutNodeSnapshot{
		layoutVersion:  syntheticLayoutVersion(nodeID, x, y, width, height, zIndex, true),
		x:              x,
		y:              y,
		measuredWidth:  width,
		measuredHeight: height,
		style: styleSnapshot{
			zIndex:  zIndex,
			visible: true,
		},
	}
}

func runtimeLayoutBoxPosition(box *runtimelayout.LayoutBox) (int, int) {
	if box == nil {
		return 0, 0
	}
	if box.AbsX != 0 || box.AbsY != 0 || (box.X == 0 && box.Y == 0) {
		return box.AbsX, box.AbsY
	}
	return box.X, box.Y
}

func extractDebugViewBoxes(view LayoutDebugView) []BoxInfo {
	if view == nil {
		return []BoxInfo{}
	}

	boxes := make([]BoxInfo, 0)
	view.ForEachBox(func(info LayoutDebugInfo) {
		nodeID := info.ID
		if nodeID == "" {
			nodeID = info.Type
		}

		snapshot := layoutNodeSnapshot{
			layoutVersion:  syntheticLayoutVersion(nodeID, info.X, info.Y, info.Width, info.Height, info.ZIndex, info.Visible),
			x:              info.X,
			y:              info.Y,
			measuredWidth:  info.Width,
			measuredHeight: info.Height,
			style: styleSnapshot{
				zIndex:  info.ZIndex,
				visible: info.Visible,
			},
		}

		boxes = append(boxes, BoxInfo{
			NodeID:         nodeID,
			Node:           &LayoutNodeAdapter{node: snapshot},
			X:              snapshot.x,
			Y:              snapshot.y,
			Width:          snapshot.measuredWidth,
			Height:         snapshot.measuredHeight,
			MeasuredWidth:  snapshot.measuredWidth,
			MeasuredHeight: snapshot.measuredHeight,
		})
	})
	return boxes
}

func syntheticLayoutVersion(nodeID string, x, y, width, height, zIndex int, visible bool) uint32 {
	hash := uint32(2166136261)

	mixByte := func(b byte) {
		hash ^= uint32(b)
		hash *= 16777619
	}
	mixInt := func(v int) {
		u := uint32(v)
		mixByte(byte(u))
		mixByte(byte(u >> 8))
		mixByte(byte(u >> 16))
		mixByte(byte(u >> 24))
	}

	for i := 0; i < len(nodeID); i++ {
		mixByte(nodeID[i])
	}
	mixInt(x)
	mixInt(y)
	mixInt(width)
	mixInt(height)
	mixInt(zIndex)
	if visible {
		mixByte(1)
	} else {
		mixByte(0)
	}

	if hash == 0 {
		return 1
	}
	return hash
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

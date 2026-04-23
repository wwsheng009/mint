package devtools

import (
	"testing"
	"time"

	"github.com/wwsheng009/mint/runtime"
	runtimelayout "github.com/wwsheng009/mint/runtime/layout"
)

func TestAdaptLayoutResultLegacyRuntimeResult(t *testing.T) {
	node := runtime.NewLayoutNode("button_1", runtime.NodeTypeText, runtime.NewStyle().WithZIndex(7))
	node.X = 3
	node.Y = 4
	node.AbsoluteX = 12
	node.AbsoluteY = 34
	node.MeasuredWidth = 80
	node.MeasuredHeight = 20
	node.LayoutVersion = 9

	result := &runtime.LayoutResult{
		Boxes: []runtime.LayoutBox{runtime.NewLayoutBox(node)},
	}

	adapter := AdaptLayoutResult(result)
	boxes := adapter.Boxes()
	if len(boxes) != 1 {
		t.Fatalf("len(boxes) = %d, want 1", len(boxes))
	}

	box := boxes[0]
	if box.NodeID != "button_1" {
		t.Fatalf("NodeID = %q, want %q", box.NodeID, "button_1")
	}
	if box.X != 12 || box.Y != 34 {
		t.Fatalf("position = (%d,%d), want (12,34)", box.X, box.Y)
	}
	if box.Width != 80 || box.Height != 20 {
		t.Fatalf("size = %dx%d, want 80x20", box.Width, box.Height)
	}
	if box.Node == nil {
		t.Fatal("box.Node is nil")
	}
	if got := box.Node.GetLayoutVersion(); got != 9 {
		t.Fatalf("GetLayoutVersion() = %d, want 9", got)
	}
	if got := box.Node.GetStyle().GetZIndex(); got != 7 {
		t.Fatalf("GetZIndex() = %d, want 7", got)
	}
}

func TestAdaptLayoutResultNewLayoutResult(t *testing.T) {
	child := &runtimelayout.LayoutBox{
		ID:      "child",
		PropsID: "submit",
		X:       2,
		Y:       3,
		AbsX:    12,
		AbsY:    13,
		Width:   30,
		Height:  10,
		ZIndex:  2,
	}
	root := &runtimelayout.LayoutBox{
		ID:       "root",
		X:        0,
		Y:        0,
		AbsX:     10,
		AbsY:     10,
		Width:    100,
		Height:   40,
		Children: []*runtimelayout.LayoutBox{child},
	}

	result := &runtimelayout.LayoutResult{
		Root:  root,
		Boxes: []runtimelayout.LayoutBox{*root, *child},
	}

	adapter := AdaptLayoutResult(result)
	boxes := adapter.Boxes()
	if len(boxes) != 2 {
		t.Fatalf("len(boxes) = %d, want 2", len(boxes))
	}

	if boxes[0].NodeID != "root" {
		t.Fatalf("boxes[0].NodeID = %q, want %q", boxes[0].NodeID, "root")
	}
	if boxes[0].X != 10 || boxes[0].Y != 10 {
		t.Fatalf("root position = (%d,%d), want (10,10)", boxes[0].X, boxes[0].Y)
	}
	if boxes[1].NodeID != "child" {
		t.Fatalf("boxes[1].NodeID = %q, want %q", boxes[1].NodeID, "child")
	}
	if boxes[1].X != 12 || boxes[1].Y != 13 {
		t.Fatalf("child position = (%d,%d), want (12,13)", boxes[1].X, boxes[1].Y)
	}
	if got := boxes[1].Node.GetStyle().GetZIndex(); got != 2 {
		t.Fatalf("child z-index = %d, want 2", got)
	}
	if got := boxes[1].Node.GetLayoutVersion(); got == 0 {
		t.Fatal("synthetic layout version should be non-zero")
	}
}

func TestLayoutCollectorCollectsLegacyRuntimeChanges(t *testing.T) {
	deltaCh := make(chan *LayoutDelta, 4)
	collector := NewLayoutCollector(deltaCh)
	collector.Enable()

	node := runtime.NewLayoutNode("node1", runtime.NodeTypeText, runtime.NewStyle().WithZIndex(1))
	node.SetPosition(2, 4)
	node.SetSize(10, 5)

	result1 := &runtime.LayoutResult{
		Boxes: []runtime.LayoutBox{runtime.NewLayoutBox(node)},
	}
	collector.Collect(result1)

	delta1 := waitForLayoutDelta(t, deltaCh)
	if len(delta1.Added) != 1 || delta1.Added[0] != NodeID("node1") {
		t.Fatalf("first delta Added = %#v, want [node1]", delta1.Added)
	}

	node.SetPosition(5, 8)
	node.SetSize(14, 7)

	result2 := &runtime.LayoutResult{
		Boxes: []runtime.LayoutBox{runtime.NewLayoutBox(node)},
	}
	collector.Collect(result2)

	delta2 := waitForLayoutDelta(t, deltaCh)
	if len(delta2.Changed) != 1 {
		t.Fatalf("len(delta2.Changed) = %d, want 1", len(delta2.Changed))
	}

	changed := delta2.Changed[0]
	if changed.ID != NodeID("node1") {
		t.Fatalf("changed.ID = %q, want %q", changed.ID, NodeID("node1"))
	}
	if changed.Mask&ChangeRect == 0 {
		t.Fatalf("changed.Mask = %v, want ChangeRect bit set", changed.Mask)
	}
	if changed.Rect == nil {
		t.Fatal("changed.Rect is nil")
	}
	if changed.Rect.X != 5 || changed.Rect.Y != 8 || changed.Rect.Width != 14 || changed.Rect.Height != 7 {
		t.Fatalf("changed.Rect = %+v, want {X:5 Y:8 Width:14 Height:7}", *changed.Rect)
	}
}

func TestLayoutCollectorCollectsNewLayoutResultChanges(t *testing.T) {
	deltaCh := make(chan *LayoutDelta, 4)
	collector := NewLayoutCollector(deltaCh)
	collector.Enable()

	child := &runtimelayout.LayoutBox{
		ID:     "child",
		X:      2,
		Y:      3,
		AbsX:   12,
		AbsY:   13,
		Width:  30,
		Height: 10,
		ZIndex: 2,
	}
	root := &runtimelayout.LayoutBox{
		ID:       "root",
		X:        0,
		Y:        0,
		AbsX:     0,
		AbsY:     0,
		Width:    100,
		Height:   40,
		Children: []*runtimelayout.LayoutBox{child},
	}

	result1 := &runtimelayout.LayoutResult{
		Root:  root,
		Boxes: []runtimelayout.LayoutBox{*root, *child},
	}
	collector.Collect(result1)

	delta1 := waitForLayoutDelta(t, deltaCh)
	if len(delta1.Added) != 2 {
		t.Fatalf("len(delta1.Added) = %d, want 2", len(delta1.Added))
	}

	child.AbsX = 14
	child.AbsY = 16
	child.Width = 31

	result2 := &runtimelayout.LayoutResult{
		Root:  root,
		Boxes: []runtimelayout.LayoutBox{*root, *child},
	}
	collector.Collect(result2)

	delta2 := waitForLayoutDelta(t, deltaCh)
	if len(delta2.Changed) != 1 {
		t.Fatalf("len(delta2.Changed) = %d, want 1", len(delta2.Changed))
	}

	changed := delta2.Changed[0]
	if changed.ID != NodeID("child") {
		t.Fatalf("changed.ID = %q, want %q", changed.ID, NodeID("child"))
	}
	if changed.Rect == nil {
		t.Fatal("changed.Rect is nil")
	}
	if changed.Rect.X != 14 || changed.Rect.Y != 16 || changed.Rect.Width != 31 || changed.Rect.Height != 10 {
		t.Fatalf("changed.Rect = %+v, want {X:14 Y:16 Width:31 Height:10}", *changed.Rect)
	}
}

func waitForLayoutDelta(t *testing.T, ch <-chan *LayoutDelta) *LayoutDelta {
	t.Helper()

	select {
	case delta := <-ch:
		return delta
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for layout delta")
		return nil
	}
}

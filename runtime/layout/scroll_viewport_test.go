package layout

import "testing"

type testScrollViewportNode struct {
	*MockNode
	viewport       ScrollViewport
	contentWidth   int
	contentHeight  int
	viewportWidth  int
	viewportHeight int
}

func newTestScrollViewportNode(id string, viewport ScrollViewport, children ...Node) *testScrollViewportNode {
	node := &testScrollViewportNode{MockNode: NewMockNode(id, viewport.Width, viewport.Height), viewport: viewport}
	node.children = children
	return node
}

func (n *testScrollViewportNode) Type() string                      { return "pageviewport" }
func (n *testScrollViewportNode) GetScrollViewport() ScrollViewport { return n.viewport }
func (n *testScrollViewportNode) SetScrollViewportMetrics(contentWidth, contentHeight, viewportWidth, viewportHeight int) {
	n.contentWidth = contentWidth
	n.contentHeight = contentHeight
	n.viewportWidth = viewportWidth
	n.viewportHeight = viewportHeight
}

func TestScrollViewportOffsetsChildAndAttachesClip(t *testing.T) {
	child := NewMockNode("child", 20, 8)
	viewport := newTestScrollViewportNode("viewport", ScrollViewport{Enabled: true, Width: 20, Height: 3, ScrollOffset: 2}, child)
	result := NewEngine().Layout(viewport, NewConstraints(0, 80, 0, 24))

	if result.Root == nil {
		t.Fatal("root is nil")
	}
	if result.Root.Width != 20 || result.Root.Height != 3 {
		t.Fatalf("viewport size = %dx%d, want 20x3", result.Root.Width, result.Root.Height)
	}
	if result.Root.Clip == nil || *result.Root.Clip != (Rect{X: 0, Y: 0, Width: 20, Height: 3}) {
		t.Fatalf("viewport clip = %#v", result.Root.Clip)
	}
	if len(result.Root.Children) != 1 {
		t.Fatalf("children len = %d, want 1", len(result.Root.Children))
	}
	childBox := result.Root.Children[0]
	if childBox.Y != -2 {
		t.Fatalf("child Y = %d, want -2", childBox.Y)
	}
	if childBox.Clip == nil || *childBox.Clip != *result.Root.Clip {
		t.Fatalf("child clip = %#v, want root clip", childBox.Clip)
	}
}

func TestScrollViewportHitMapClipsScrolledOutChild(t *testing.T) {
	above := NewMockNode("above", 10, 1)
	visible := NewMockNode("visible", 10, 1)
	viewport := newTestScrollViewportNode("viewport", ScrollViewport{Enabled: true, Width: 10, Height: 1, ScrollOffset: 1}, above, visible)
	result := NewEngine().Layout(viewport, NewConstraints(0, 80, 0, 24))

	if result.HitMap == nil {
		t.Fatal("hitmap is nil")
	}
	if entry := result.HitMap.Get("above"); entry != nil {
		t.Fatalf("above hit entry = %#v, want nil", entry)
	}
	entry := result.HitMap.Get("visible")
	if entry == nil {
		t.Fatal("visible hit entry is nil")
	}
	if entry.Rect != (Rect{X: 0, Y: 0, Width: 10, Height: 1}) {
		t.Fatalf("visible hit rect = %#v", entry.Rect)
	}
}

func TestScrollViewportMetricsUseUnclippedContentHeight(t *testing.T) {
	child := NewMockNode("child", 20, 6)
	viewport := newTestScrollViewportNode("viewport", ScrollViewport{Enabled: true, Width: 20, Height: 3, ScrollOffset: 2}, child)
	NewEngine().Layout(viewport, NewConstraints(0, 80, 0, 24))

	if viewport.contentHeight != 6 {
		t.Fatalf("content height = %d, want unclipped height 6", viewport.contentHeight)
	}
	if viewport.viewportHeight != 3 {
		t.Fatalf("viewport height = %d, want 3", viewport.viewportHeight)
	}
}

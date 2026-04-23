package ui

import (
	"strings"
	"testing"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/framework/component"
	"github.com/wwsheng009/mint/internal/render"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

func TestWrapWithDefaultPortalRootsInjectsHosts(t *testing.T) {
	wrapped := wrapWithDefaultPortalRoots(func() VNode {
		return Text("hello")
	})

	vnode := wrapped()
	children := vnode.Children()
	if len(children) != 4 {
		t.Fatalf("children len = %d, want 4", len(children))
	}

	expect := []struct {
		index int
		id    string
		layer rtui.Layer
	}{
		{0, rtui.DefaultOverlayPortalRootID, rtui.LayerOverlay},
		{1, rtui.DefaultModalPortalRootID, rtui.LayerModal},
		{2, rtui.DefaultTooltipPortalRootID, rtui.LayerTooltip},
	}

	for _, item := range expect {
		props := children[item.index].Props()
		if got, _ := props["portalRootId"].(string); got != item.id {
			t.Fatalf("child %d portalRootId = %q, want %q", item.index, got, item.id)
		}
		if got := children[item.index].GetLayer(); got != item.layer {
			t.Fatalf("child %d layer = %v, want %v", item.index, got, item.layer)
		}
		if got, _ := props["position"].(string); got != "absolute" {
			t.Fatalf("child %d position = %q, want absolute", item.index, got)
		}
	}
}

func TestWrapWithDefaultPortalRoots_DoesNotShiftInitialContentTrace(t *testing.T) {
	wrapped := wrapWithDefaultPortalRoots(func() VNode {
		return VStack(
			Text("hello"),
			Text("world"),
		).SetID("content-root")
	})

	node := render.NewDeclarativeNodeFromFuncWithFiber(wrapped)
	node.SetApp(framework.NewApp())
	node.SetRenderMode(render.RenderModeFiberFirst)

	buf := paint.NewBuffer(20, 8)
	ctx := component.PaintContext{
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 20, Height: 8},
		AvailableWidth:  20,
		AvailableHeight: 8,
	}

	node.Paint(ctx, buf)

	layoutTree := node.GetLayoutTreeString()
	paintableTree := node.GetPaintableTreeString()

	contentBox := findLayoutBoxByPropsID(node.GetLayoutRoot(), "content-root")
	if contentBox == nil {
		t.Fatalf("content-root not found in layout tree\n%s\n%s", layoutTree, paintableTree)
	}

	if contentBox.Y != 0 {
		t.Fatalf("content-root Y = %d, want 0\n%s\n%s", contentBox.Y, layoutTree, paintableTree)
	}

	if strings.Contains(layoutTree, "Pos:0,3") {
		t.Fatalf("layout trace still shows content shifted by portal roots\n%s\n%s", layoutTree, paintableTree)
	}
}

func findLayoutBoxByPropsID(box *layout.LayoutBox, id string) *layout.LayoutBox {
	if box == nil {
		return nil
	}
	if box.PropsID == id {
		return box
	}
	for _, child := range box.Children {
		if found := findLayoutBoxByPropsID(child, id); found != nil {
			return found
		}
	}
	return nil
}

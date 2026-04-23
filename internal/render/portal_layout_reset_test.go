package render

import (
	"testing"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/framework/component"
	"github.com/wwsheng009/mint/runtime/paint"
	rttypes "github.com/wwsheng009/mint/runtime/types"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

func TestDeclarativeNode_PortalLayoutClearsClosedPortalBetweenFrames(t *testing.T) {
	showPopup := true

	app := func() rtui.VNode {
		host := rtui.NewElement("box").SetProps(rtui.Props{
			"portalRootId": "overlay-root",
			"_layer":       rtui.LayerOverlay,
			"position":     "absolute",
			"left":         0,
			"top":          0,
			"width":        1,
			"height":       1,
		})
		children := []rtui.VNode{host, rtui.NewElement("text").SetProp("content", "base")}
		if showPopup {
			portal := rtui.NewElement("box").SetProps(rtui.Props{
				"portalRoot": "overlay-root",
				"position":   rttypes.PositionAbsolute,
				"left":       3,
				"top":        1,
			})
			surface := rtui.NewElement("box").SetProps(rtui.Props{
				"width":  6,
				"height": 2,
				"_layer": rtui.LayerOverlay,
			})
			portal.SetChildren([]rtui.VNode{surface})
			children = append(children, portal)
		}
		return rtui.VStack(children...)
	}

	node := NewDeclarativeNodeFromFuncWithFiber(app)
	node.SetApp(framework.NewApp())
	node.SetRenderMode(RenderModeFiberFirst)

	ctx := component.PaintContext{
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 20, Height: 8},
		AvailableWidth:  20,
		AvailableHeight: 8,
	}

	node.Paint(ctx, paint.NewBuffer(20, 8))
	if got := len(node.GetPortalRoots()); got != 1 {
		t.Fatalf("first frame portal roots = %d, want 1\n%s", got, node.GetPortalTreeString())
	}

	showPopup = false
	node.Paint(ctx, paint.NewBuffer(20, 8))

	if got := len(node.GetPortalRoots()); got != 0 {
		t.Fatalf("second frame portal roots = %d, want 0\n%s", got, node.GetPortalTreeString())
	}
}

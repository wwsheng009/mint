package render

import (
	"testing"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/framework/component"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/paint"
	rttypes "github.com/wwsheng009/mint/runtime/types"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

func TestResolvePortalPositioningSizePrefersExplicitWrapperProps(t *testing.T) {
	width, height := resolvePortalPositioningSize(map[string]interface{}{
		"positioningWidth":  4,
		"positioningHeight": 2,
	}, &layout.LayoutBox{
		Width:  10,
		Height: 6,
	})

	if width != 4 || height != 2 {
		t.Fatalf("resolvePortalPositioningSize() = (%d,%d), want (4,2)", width, height)
	}
}

func TestResolvePortalPositioningSizeFallsBackToLayoutSize(t *testing.T) {
	width, height := resolvePortalPositioningSize(nil, &layout.LayoutBox{
		Width:  10,
		Height: 6,
	})

	if width != 10 || height != 6 {
		t.Fatalf("resolvePortalPositioningSize() = (%d,%d), want (10,6)", width, height)
	}
}

func TestDeclarativeNode_PortalLayoutAnchorPositionUsesExplicitWrapperSize(t *testing.T) {
	app := func() rtui.VNode {
		overlayHost := rtui.NewElement("box")
		overlayHost.SetProps(rtui.Props{
			"portalRootId": "overlay-root",
			"_layer":       rtui.LayerOverlay,
			"width":        1,
			"height":       1,
		})

		topSpacer := rtui.NewElement("box")
		topSpacer.SetProps(rtui.Props{
			"width":  1,
			"height": 1,
		})

		leftSpacer := rtui.NewElement("box")
		leftSpacer.SetProps(rtui.Props{
			"width":  10,
			"height": 1,
		})

		anchor := rtui.NewElement("box")
		anchor.SetID("menu-anchor")
		anchor.SetProps(rtui.Props{
			"width":  4,
			"height": 1,
		})

		anchorRow := rtui.NewElement("hstack")
		anchorRow.SetProps(rtui.Props{
			"direction": rtui.DirectionRow,
			"gap":       0,
		})
		anchorRow.SetChildren([]rtui.VNode{leftSpacer, anchor})

		rootSurface := rtui.NewElement("box")
		rootSurface.SetProps(rtui.Props{
			"width":  4,
			"height": 2,
			"_layer": rtui.LayerOverlay,
		})

		submenuSurface := rtui.NewElement("box")
		submenuSurface.SetProps(rtui.Props{
			"width":  6,
			"height": 2,
			"_layer": rtui.LayerOverlay,
		})

		content := rtui.NewElement("hstack")
		content.SetProps(rtui.Props{
			"direction": rtui.DirectionRow,
			"gap":       0,
		})
		content.SetChildren([]rtui.VNode{rootSurface, submenuSurface})

		portal := rtui.NewElement("box")
		portal.SetProps(rtui.Props{
			"portalRoot":        "overlay-root",
			"position":          rttypes.PositionAbsolute,
			"anchorId":          "menu-anchor",
			"anchor":            rttypes.AnchorTopRight,
			"left":              0,
			"top":               0,
			"positioningWidth":  4,
			"positioningHeight": 2,
		})
		portal.SetChildren([]rtui.VNode{content})

		return rtui.VStack(
			overlayHost,
			topSpacer,
			anchorRow,
			portal,
		)
	}

	node := NewDeclarativeNodeFromFuncWithFiber(app)
	node.SetApp(framework.NewApp())
	node.SetRenderMode(RenderModeFiberFirst)

	ctx := component.PaintContext{
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 30, Height: 10},
		AvailableWidth:  30,
		AvailableHeight: 10,
	}

	node.Paint(ctx, paint.NewBuffer(30, 10))

	portalRoots := node.GetPortalRoots()
	if len(portalRoots) != 1 {
		t.Fatalf("portal roots = %d, want 1\n%s", len(portalRoots), node.GetPortalTreeString())
	}

	root := portalRoots[0]
	if len(root.Children) == 0 {
		t.Fatalf("portal root has no children\n%s", node.GetPortalTreeString())
	}
	if root.Children[0].Width <= 4 {
		t.Fatalf("portal content width = %d, want > 4 to prove subtree is wider than wrapper\n%s", root.Children[0].Width, node.GetPortalTreeString())
	}
	if root.X != 10 {
		t.Fatalf("portal root x = %d, want 10\n%s", root.X, node.GetPortalTreeString())
	}
	if root.Y != 2 {
		t.Fatalf("portal root y = %d, want 2\n%s", root.Y, node.GetPortalTreeString())
	}
}

func TestDeclarativeNode_AnchoredPopupPlacementFallsBackBelowTopEdge(t *testing.T) {
	app := func() rtui.VNode {
		overlayHost := rtui.NewElement("box")
		overlayHost.SetProps(rtui.Props{
			"portalRootId": "overlay-root",
			"_layer":       rtui.LayerOverlay,
			"width":        1,
			"height":       1,
		})

		header := rtui.NewElement("box")
		header.SetProps(rtui.Props{
			"width":  72,
			"height": 1,
		})

		leftSpacer := rtui.NewElement("box")
		leftSpacer.SetProps(rtui.Props{
			"width":  61,
			"height": 1,
		})

		anchor := rtui.NewElement("box")
		anchor.SetID("menu-anchor")
		anchor.SetProps(rtui.Props{
			"width":  11,
			"height": 1,
		})

		anchorRow := rtui.NewElement("hstack")
		anchorRow.SetProps(rtui.Props{
			"direction": rtui.DirectionRow,
			"gap":       0,
			"width":     72,
			"height":    1,
		})
		anchorRow.SetChildren([]rtui.VNode{leftSpacer, anchor})

		content := rtui.NewElement("box")
		content.SetProps(rtui.Props{
			"width":  19,
			"height": 3,
			"_layer": rtui.LayerOverlay,
		})

		portal := rtui.NewElement("box")
		portal.SetProps(rtui.Props{
			"portalRoot":        "overlay-root",
			"position":          rttypes.PositionAbsolute,
			"anchorId":          "menu-anchor",
			"anchor":            rttypes.AnchorTopRight,
			"left":              0,
			"top":               -3,
			"width":             19,
			"height":            3,
			"positioningWidth":  19,
			"positioningHeight": 3,
			"popupPlacement":    "top-end",
			"popupOffsetX":      0,
			"popupOffsetY":      0,
		})
		portal.SetChildren([]rtui.VNode{content})

		return rtui.VStack(
			overlayHost,
			header,
			anchorRow,
			portal,
		)
	}

	node := NewDeclarativeNodeFromFuncWithFiber(app)
	node.SetApp(framework.NewApp())
	node.SetRenderMode(RenderModeFiberFirst)

	ctx := component.PaintContext{
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 72, Height: 18},
		AvailableWidth:  72,
		AvailableHeight: 18,
	}

	node.Paint(ctx, paint.NewBuffer(72, 18))

	portalRoots := node.GetPortalRoots()
	if len(portalRoots) != 1 {
		t.Fatalf("portal roots = %d, want 1\n%s", len(portalRoots), node.GetPortalTreeString())
	}

	root := portalRoots[0]
	if root.X != 53 {
		t.Fatalf("portal root x = %d, want 53\n%s", root.X, node.GetPortalTreeString())
	}
	if root.Y != 3 {
		t.Fatalf("portal root y = %d, want 3 after top-edge fallback\n%s", root.Y, node.GetPortalTreeString())
	}
}

func TestDeclarativeNode_ViewportClampedPopupPositionFitsWithinBottomRightEdge(t *testing.T) {
	app := func() rtui.VNode {
		overlayHost := rtui.NewElement("box")
		overlayHost.SetProps(rtui.Props{
			"portalRootId": "overlay-root",
			"_layer":       rtui.LayerOverlay,
			"width":        1,
			"height":       1,
		})

		body := rtui.NewElement("box")
		body.SetProps(rtui.Props{
			"width":  72,
			"height": 18,
		})

		content := rtui.NewElement("box")
		content.SetProps(rtui.Props{
			"width":  19,
			"height": 4,
			"_layer": rtui.LayerOverlay,
		})

		portal := rtui.NewElement("box")
		portal.SetProps(rtui.Props{
			"portalRoot":           "overlay-root",
			"position":             rttypes.PositionAbsolute,
			"left":                 66,
			"top":                  16,
			"width":                19,
			"height":               4,
			"positioningWidth":     19,
			"positioningHeight":    4,
			"popupClampToViewport": true,
		})
		portal.SetChildren([]rtui.VNode{content})

		return rtui.VStack(
			overlayHost,
			body,
			portal,
		)
	}

	node := NewDeclarativeNodeFromFuncWithFiber(app)
	node.SetApp(framework.NewApp())
	node.SetRenderMode(RenderModeFiberFirst)

	ctx := component.PaintContext{
		Bounds:          paint.Rect{X: 0, Y: 0, Width: 72, Height: 18},
		AvailableWidth:  72,
		AvailableHeight: 18,
	}

	node.Paint(ctx, paint.NewBuffer(72, 18))

	portalRoots := node.GetPortalRoots()
	if len(portalRoots) != 1 {
		t.Fatalf("portal roots = %d, want 1\n%s", len(portalRoots), node.GetPortalTreeString())
	}

	root := portalRoots[0]
	if root.X != 53 {
		t.Fatalf("portal root x = %d, want 53 after right-edge clamp\n%s", root.X, node.GetPortalTreeString())
	}
	if root.Y != 14 {
		t.Fatalf("portal root y = %d, want 14 after bottom-edge clamp\n%s", root.Y, node.GetPortalTreeString())
	}
}

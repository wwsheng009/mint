package popconfirm

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	rttypes "github.com/wwsheng009/mint/runtime/types"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/button"
	textcomp "github.com/wwsheng009/mint/ui/components/text"
)

type fakeInstallerHost struct {
	middlewareCount int
	userData        map[string]interface{}
}

func (h *fakeInstallerHost) AddMiddleware(_ action.ActionMiddleware) {
	h.middlewareCount++
}

func (h *fakeInstallerHost) SetUserData(key string, value interface{}) {
	if h.userData == nil {
		h.userData = map[string]interface{}{}
	}
	h.userData[key] = value
}

func (h *fakeInstallerHost) GetUserData(key string) interface{} {
	if h.userData == nil {
		return nil
	}
	return h.userData[key]
}

func TestNewCapturesButtonIntent(t *testing.T) {
	anchor := button.NewBuilder("Delete").OnPress(intent.FieldChangeIntent{Field: "delete", Value: "1"}).Build().(*button.VNode)
	v := New(anchor)
	if v == nil {
		t.Fatal("New returned nil")
	}
	if v.confirmIntent == nil {
		t.Fatal("expected button press intent to be captured as confirm intent")
	}
	if v.trigger != TriggerClick || v.placement != PlacementTop {
		t.Fatalf("defaults = (%v,%v)", v.trigger, v.placement)
	}
}

func TestChildrenAssignsAnchorIDAndOverridesButtonIntent(t *testing.T) {
	anchor := button.NewBuilder("Delete").Build().(*button.VNode)
	v := New(anchor).SetComponentID("delete.confirm")
	children := v.Children()
	if len(children) != 1 {
		t.Fatalf("children len = %d, want 1", len(children))
	}
	if anchor.ID() != "delete.confirm-anchor" {
		t.Fatalf("anchor ID = %q, want delete.confirm-anchor", anchor.ID())
	}
	if _, ok := anchor.PressIntent().(PopconfirmToggleIntent); !ok {
		t.Fatalf("anchor press intent = %T, want PopconfirmToggleIntent", anchor.PressIntent())
	}
}

func TestHandleActionToggleAndIntentConfirm(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID: "delete.confirm",
		propTitle:       "Delete?",
	})
	var emitted []intent.Intent
	inst.SetIntentEmitter(func(i intent.Intent) {
		emitted = append(emitted, i)
	})
	if !inst.HandleAction(action.NewAction(action.ActionClick)) || !inst.open {
		t.Fatal("click should open popconfirm")
	}
	if !inst.HandleIntent(confirmClickIntent{ComponentID: "delete.confirm"}) {
		t.Fatal("confirm click intent should be handled")
	}
	if inst.open {
		t.Fatal("confirm should close popconfirm")
	}
	found := false
	for _, evt := range emitted {
		if _, ok := evt.(PopconfirmConfirmIntent); ok {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected PopconfirmConfirmIntent to be emitted")
	}
}

func TestHandleIntentRespectsComponentID(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID: "delete.confirm",
		propTitle:       "Delete?",
	})
	if !inst.HandleIntent(OpenWithID("delete.confirm")) {
		t.Fatal("expected matching open intent to be handled")
	}
	if inst.HandleIntent(CloseWithID("other.confirm")) {
		t.Fatal("expected other componentID to be ignored")
	}
}

func TestRuntimeChildrenBuildsPortal(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID:    "delete.confirm",
		propAnchorID:       "delete.confirm-anchor",
		propTitle:          "Delete?",
		propDescription:    "This action cannot be undone.",
		propOpen:           true,
		propOpenControlled: true,
	})
	inst.bounds = [4]int{10, 5, 12, 1}
	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	portal := children[0]
	if portal.GetLayer() != rtui.LayerOverlay {
		t.Fatalf("portal layer = %v, want %v", portal.GetLayer(), rtui.LayerOverlay)
	}
	props := portal.Props()
	if props.GetString("portalRoot") != rtui.DefaultOverlayPortalRootID {
		t.Fatalf("portalRoot = %q, want %q", props.GetString("portalRoot"), rtui.DefaultOverlayPortalRootID)
	}
}

func TestBuildOverlaySurfaceContainsButtons(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID: "delete.confirm",
		propAnchorID:    "delete.confirm-anchor",
		propTitle:       "Delete?",
		propDescription: "Cannot undo",
	})
	surface := inst.buildOverlaySurface()
	if surface == nil {
		t.Fatal("expected overlay surface")
	}
	if !containsVNodeText(surface, "Delete?") || !containsVNodeText(surface, "Cannot undo") {
		t.Fatal("expected title and description in overlay surface")
	}
	if !containsVNodeText(surface, "OK") || !containsVNodeText(surface, "Cancel") {
		t.Fatal("expected action buttons in overlay surface")
	}
}

func TestOverlayBoundsUseSharedPlacementCoordinates(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID: "delete.confirm",
		propTitle:       "Delete?",
	})
	inst.SetBounds(10, 8, 8, 1)

	inst.placement = PlacementTop
	x, y, w, h := inst.overlayBounds()
	if x != 10+(8-w)/2 || y != 8-h-inst.gapRows {
		t.Fatalf("top bounds = (%d,%d,%d,%d)", x, y, w, h)
	}

	inst.placement = PlacementTopRight
	x, y, w, h = inst.overlayBounds()
	if x != 18-w || y != 8-h-inst.gapRows {
		t.Fatalf("top-right bounds = (%d,%d,%d,%d)", x, y, w, h)
	}

	inst.placement = PlacementBottomLeft
	x, y, w, h = inst.overlayBounds()
	if x != 10 || y != 10 {
		t.Fatalf("bottom-left bounds = (%d,%d,%d,%d)", x, y, w, h)
	}
}

func TestRuntimeChildrenPlacementAutoUsesTopAnchorAlignment(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID:    "delete.confirm",
		propAnchorID:       "delete.confirm-anchor",
		propTitle:          "Delete?",
		propPlacement:      PlacementAuto,
		propOpen:           true,
		propOpenControlled: true,
	})
	inst.SetBounds(10, 8, 12, 1)

	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}

	props := children[0].Props()
	if got, _ := props["anchor"].(rttypes.Anchor); got != rttypes.AnchorTopLeft {
		t.Fatalf("anchor = %v, want %v", props["anchor"], rttypes.AnchorTopLeft)
	}
	overlayX, overlayY, _, _ := inst.overlayBounds()
	if got, ok := props["left"].(int); !ok || got != overlayX-inst.bounds[0] {
		t.Fatalf("left = %v, want %d", props["left"], overlayX-inst.bounds[0])
	}
	if got, ok := props["top"].(int); !ok || got != overlayY-inst.bounds[1] {
		t.Fatalf("top = %v, want %d", props["top"], overlayY-inst.bounds[1])
	}
}

func TestRuntimeChildrenPlacementAutoFallsBackBelowWhenTopSpaceInsufficient(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID:    "delete.confirm",
		propAnchorID:       "delete.confirm-anchor",
		propTitle:          "Delete?",
		propPlacement:      PlacementAuto,
		propOpen:           true,
		propOpenControlled: true,
	})
	inst.SetBounds(10, 1, 12, 1)

	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}

	props := children[0].Props()
	if got, _ := props["anchor"].(rttypes.Anchor); got != rttypes.AnchorTopLeft {
		t.Fatalf("anchor = %v, want %v", props["anchor"], rttypes.AnchorTopLeft)
	}
	overlayX, overlayY, _, _ := inst.overlayBounds()
	if got, ok := props["left"].(int); !ok || got != overlayX-inst.bounds[0] {
		t.Fatalf("left = %v, want %d", props["left"], overlayX-inst.bounds[0])
	}
	if got, ok := props["top"].(int); !ok || got != overlayY-inst.bounds[1] {
		t.Fatalf("top = %v, want %d", props["top"], overlayY-inst.bounds[1])
	}
}

func TestResolvePopconfirmPlacementAutoMirrorsPopoverHeuristic(t *testing.T) {
	tests := []struct {
		name    string
		anchor  [4]int
		height  int
		gapRows int
		want    Placement
	}{
		{
			name:    "uses top when anchor has enough room above",
			anchor:  [4]int{10, 9, 8, 1},
			height:  5,
			gapRows: 1,
			want:    PlacementTop,
		},
		{
			name:    "falls back to bottom near top edge",
			anchor:  [4]int{10, 1, 8, 1},
			height:  5,
			gapRows: 1,
			want:    PlacementBottom,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolvePopconfirmPlacement(tt.anchor, PlacementAuto, tt.gapRows, tt.height, [2]int{}); got != tt.want {
				t.Fatalf("resolvePopconfirmPlacement() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOverlayBoundsPlacementTopFallsBackWithinViewport(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID: "delete.confirm",
		propTitle:       "Delete now?",
		propPlacement:   PlacementTop,
	})
	inst.SetBounds(1, 1, 4, 1)
	inst.SetViewportSize(28, 8)

	x, y, _, _ := inst.overlayBounds()
	if y <= inst.bounds[1] {
		t.Fatalf("overlay y = %d, want below anchor after top-edge fallback", y)
	}
	if x < 0 || x >= 28 {
		t.Fatalf("overlay x = %d, want within viewport", x)
	}
}

func TestOverlayBoundsPlacementTopStaysAboveAndShiftsRightNearLeftEdge(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID: "delete.confirm",
		propTitle:       "Delete now?",
		propPlacement:   PlacementTop,
	})
	inst.SetBounds(2, 8, 4, 1)
	inst.SetViewportSize(40, 16)

	x, y, _, _ := inst.overlayBounds()
	if x != 2 {
		t.Fatalf("overlay x = %d, want 2 after top-family fallback near left edge", x)
	}
	if y >= inst.bounds[1] {
		t.Fatalf("overlay y = %d, want above anchor row %d", y, inst.bounds[1])
	}
}

func TestOverlayBoundsPlacementTopRightFallsBelowWithinRightFamilyNearCorner(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID: "delete.confirm",
		propTitle:       "Delete now?",
		propPlacement:   PlacementTopRight,
	})
	inst.SetBounds(34, 1, 4, 1)
	inst.SetViewportSize(40, 10)

	x, y, _, _ := inst.overlayBounds()
	if x != 21 {
		t.Fatalf("overlay x = %d, want 21 after top-right corner fallback", x)
	}
	if y != 3 {
		t.Fatalf("overlay y = %d, want 3 after top-right corner fallback", y)
	}
}

func TestOverlayBoundsPlacementTopLeftFallsBelowWithinLeftFamilyNearCorner(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID: "delete.confirm",
		propTitle:       "Delete now?",
		propPlacement:   PlacementTopLeft,
	})
	inst.SetBounds(2, 1, 4, 1)
	inst.SetViewportSize(40, 10)

	x, y, _, _ := inst.overlayBounds()
	if x != 2 {
		t.Fatalf("overlay x = %d, want 2 after top-left corner fallback", x)
	}
	if y != 3 {
		t.Fatalf("overlay y = %d, want 3 after top-left corner fallback", y)
	}
}

func TestOverlayBoundsPlacementBottomStaysBelowAndShiftsLeftNearRightEdge(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID: "delete.confirm",
		propTitle:       "Delete now?",
		propPlacement:   PlacementBottom,
	})
	inst.SetBounds(34, 8, 4, 1)
	inst.SetViewportSize(40, 16)

	x, y, _, _ := inst.overlayBounds()
	if x != 21 {
		t.Fatalf("overlay x = %d, want 21 after bottom-family fallback near right edge", x)
	}
	if y <= inst.bounds[1] {
		t.Fatalf("overlay y = %d, want below anchor row %d", y, inst.bounds[1])
	}
}

func TestOverlayBoundsPlacementBottomRightFallsAboveWithinRightFamilyNearCorner(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID: "delete.confirm",
		propTitle:       "Delete now?",
		propPlacement:   PlacementBottomRight,
	})
	inst.SetBounds(34, 8, 4, 1)
	inst.SetViewportSize(40, 10)

	x, y, _, _ := inst.overlayBounds()
	if x != 21 {
		t.Fatalf("overlay x = %d, want 21 after bottom-right corner fallback", x)
	}
	if y != 2 {
		t.Fatalf("overlay y = %d, want 2 after bottom-right corner fallback", y)
	}
}

func TestOverlayBoundsPlacementBottomLeftFallsAboveWithinLeftFamilyNearCorner(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID: "delete.confirm",
		propTitle:       "Delete now?",
		propPlacement:   PlacementBottomLeft,
	})
	inst.SetBounds(2, 8, 4, 1)
	inst.SetViewportSize(40, 10)

	x, y, _, _ := inst.overlayBounds()
	if x != 2 {
		t.Fatalf("overlay x = %d, want 2 after bottom-left corner fallback", x)
	}
	if y != 2 {
		t.Fatalf("overlay y = %d, want 2 after bottom-left corner fallback", y)
	}
}

func TestOverlayBoundsPlacementTopRightClampsLeftAndStaysAboveInNarrowViewport(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID: "delete.confirm",
		propTitle:       "Delete now?",
		propPlacement:   PlacementTopRight,
	})
	inst.SetBounds(9, 7, 4, 1)
	inst.SetViewportSize(14, 14)

	x, y, _, _ := inst.overlayBounds()
	if x != 0 {
		t.Fatalf("overlay x = %d, want left-edge clamp to 0 in narrow viewport", x)
	}
	if y != 1 {
		t.Fatalf("overlay y = %d, want 1 while staying above anchor in narrow viewport", y)
	}
}

func TestOverlayBoundsPlacementTopLeftClampsLeftAndStaysAboveInNarrowViewport(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID: "delete.confirm",
		propTitle:       "Delete now?",
		propPlacement:   PlacementTopLeft,
	})
	inst.SetBounds(9, 7, 4, 1)
	inst.SetViewportSize(14, 14)

	x, y, _, _ := inst.overlayBounds()
	if x != 0 {
		t.Fatalf("overlay x = %d, want left-edge clamp to 0 in narrow viewport", x)
	}
	if y != 1 {
		t.Fatalf("overlay y = %d, want 1 while staying above anchor in narrow viewport", y)
	}
}

func TestResolvePopconfirmLayoutPlacementTopLeftClampsBothAxesAndKeepsTopFamilyWhenNothingFits(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID: "delete.confirm",
		propTitle:       "Delete now?",
		propPlacement:   PlacementTopLeft,
	})
	result := resolvePopconfirmLayout([4]int{9, 1, 4, 1}, PlacementTopLeft, inst.overlayWidth(), inst.overlayHeight(), inst.gapRows, [2]int{14, 5})
	if result.X != 0 || result.Y != 0 {
		t.Fatalf("dual-axis top-left result = (%d,%d), want (0,0)", result.X, result.Y)
	}
	if result.Placement != popconfirmPlacementToOverlay(PlacementTopLeft) {
		t.Fatalf("placement = %v, want %v after dual-axis clamp", result.Placement, popconfirmPlacementToOverlay(PlacementTopLeft))
	}
	if !result.Clamped {
		t.Fatal("result should report clamping when no vertical candidate fits")
	}
}

func TestResolvePopconfirmLayoutPlacementTopRightClampsBothAxesAndKeepsTopFamilyWhenNothingFits(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID: "delete.confirm",
		propTitle:       "Delete now?",
		propPlacement:   PlacementTopRight,
	})
	result := resolvePopconfirmLayout([4]int{9, 1, 4, 1}, PlacementTopRight, inst.overlayWidth(), inst.overlayHeight(), inst.gapRows, [2]int{14, 5})
	if result.X != 0 || result.Y != 0 {
		t.Fatalf("dual-axis top-right result = (%d,%d), want (0,0)", result.X, result.Y)
	}
	if result.Placement != popconfirmPlacementToOverlay(PlacementTopRight) {
		t.Fatalf("placement = %v, want %v after dual-axis clamp", result.Placement, popconfirmPlacementToOverlay(PlacementTopRight))
	}
	if !result.Clamped {
		t.Fatal("result should report clamping when no vertical candidate fits")
	}
}

func TestOverlayBoundsPlacementBottomStaysBelowAndShiftsRightNearLeftEdge(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID: "delete.confirm",
		propTitle:       "Delete now?",
		propPlacement:   PlacementBottom,
	})
	inst.SetBounds(2, 8, 4, 1)
	inst.SetViewportSize(40, 16)

	x, y, _, _ := inst.overlayBounds()
	if x != 2 {
		t.Fatalf("overlay x = %d, want 2 after bottom-family fallback near left edge", x)
	}
	if y <= inst.bounds[1] {
		t.Fatalf("overlay y = %d, want below anchor row %d", y, inst.bounds[1])
	}
}

func TestOverlayBoundsPlacementBottomRightClampsLeftAndStaysBelowInNarrowViewport(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID: "delete.confirm",
		propTitle:       "Delete now?",
		propPlacement:   PlacementBottomRight,
	})
	inst.SetBounds(9, 7, 4, 1)
	inst.SetViewportSize(14, 14)

	x, y, _, _ := inst.overlayBounds()
	if x != 0 {
		t.Fatalf("overlay x = %d, want left-edge clamp to 0 in narrow viewport", x)
	}
	if y != 9 {
		t.Fatalf("overlay y = %d, want 9 while staying below anchor in narrow viewport", y)
	}
}

func TestOverlayBoundsPlacementBottomLeftClampsLeftAndStaysBelowInNarrowViewport(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID: "delete.confirm",
		propTitle:       "Delete now?",
		propPlacement:   PlacementBottomLeft,
	})
	inst.SetBounds(9, 7, 4, 1)
	inst.SetViewportSize(14, 14)

	x, y, _, _ := inst.overlayBounds()
	if x != 0 {
		t.Fatalf("overlay x = %d, want left-edge clamp to 0 in narrow viewport", x)
	}
	if y != 9 {
		t.Fatalf("overlay y = %d, want 9 while staying below anchor in narrow viewport", y)
	}
}

func TestResolvePopconfirmLayoutPlacementBottomLeftClampsBothAxesAndKeepsBottomFamilyWhenNothingFits(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID: "delete.confirm",
		propTitle:       "Delete now?",
		propPlacement:   PlacementBottomLeft,
	})
	result := resolvePopconfirmLayout([4]int{9, 1, 4, 1}, PlacementBottomLeft, inst.overlayWidth(), inst.overlayHeight(), inst.gapRows, [2]int{14, 5})
	if result.X != 0 || result.Y != 0 {
		t.Fatalf("dual-axis bottom-left result = (%d,%d), want (0,0)", result.X, result.Y)
	}
	if result.Placement != popconfirmPlacementToOverlay(PlacementBottomLeft) {
		t.Fatalf("placement = %v, want %v after dual-axis clamp", result.Placement, popconfirmPlacementToOverlay(PlacementBottomLeft))
	}
	if !result.Clamped {
		t.Fatal("result should report clamping when no vertical candidate fits")
	}
}

func TestResolvePopconfirmLayoutPlacementBottomRightClampsBothAxesAndKeepsBottomFamilyWhenNothingFits(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID: "delete.confirm",
		propTitle:       "Delete now?",
		propPlacement:   PlacementBottomRight,
	})
	result := resolvePopconfirmLayout([4]int{9, 1, 4, 1}, PlacementBottomRight, inst.overlayWidth(), inst.overlayHeight(), inst.gapRows, [2]int{14, 5})
	if result.X != 0 || result.Y != 0 {
		t.Fatalf("dual-axis bottom-right result = (%d,%d), want (0,0)", result.X, result.Y)
	}
	if result.Placement != popconfirmPlacementToOverlay(PlacementBottomRight) {
		t.Fatalf("placement = %v, want %v after dual-axis clamp", result.Placement, popconfirmPlacementToOverlay(PlacementBottomRight))
	}
	if !result.Clamped {
		t.Fatal("result should report clamping when no vertical candidate fits")
	}
}

func TestOverlayBoundsClampWithinViewportWhenNoCandidateFits(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID: "delete.confirm",
		propTitle:       "Delete this record now?",
		propPlacement:   PlacementBottom,
	})
	inst.SetBounds(1, 4, 4, 1)
	inst.SetViewportSize(16, 8)

	x, _, _, _ := inst.overlayBounds()
	if x != 0 {
		t.Fatalf("overlay x = %d, want left-edge clamp to 0", x)
	}
}

func TestBuildActionRowSupportsVariantsAndFooterLayouts(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID:   "delete.confirm",
		propAnchorID:      "delete.confirm-anchor",
		propTitle:         "Delete?",
		propOkVariant:     button.VariantDanger,
		propCancelVariant: button.VariantSuccess,
		propFooterLayout:  FooterLayoutCenter,
	})

	row := inst.buildActionRow()
	if row.Tag() != "hstack" {
		t.Fatalf("row tag = %q, want hstack", row.Tag())
	}
	if align, ok := row.Props()["align"].(rtui.Align); !ok || align != rtui.AlignCenter {
		t.Fatalf("row align = %v, want AlignCenter", row.Props()["align"])
	}
	buttons := actionRowButtons(t, row)
	if buttons[0].Variant() != button.VariantSuccess || buttons[1].Variant() != button.VariantDanger {
		t.Fatalf("button variants = (%v,%v), want (%v,%v)", buttons[0].Variant(), buttons[1].Variant(), button.VariantSuccess, button.VariantDanger)
	}

	inst.footerLayout = FooterLayoutStretch
	row = inst.buildActionRow()
	buttons = actionRowButtons(t, row)
	if buttons[0].GetFlex() != 1 || buttons[1].GetFlex() != 1 {
		t.Fatalf("button flex = (%d,%d), want (1,1)", buttons[0].GetFlex(), buttons[1].GetFlex())
	}

	inst.footerLayout = FooterLayoutVertical
	row = inst.buildActionRow()
	if row.Tag() != "vstack" {
		t.Fatalf("vertical row tag = %q, want vstack", row.Tag())
	}
}

func TestPopconfirmMiddlewareEscapeClosesTopmostOpenPopconfirm(t *testing.T) {
	popconfirmRegistryGlobal.reset()
	defer popconfirmRegistryGlobal.reset()

	first := NewInstance(rtui.Props{propComponentID: "first.confirm", propTitle: "First?"})
	first.OnMount()
	defer first.Destroy()
	if !first.setOpen(true, TriggerClick) {
		t.Fatal("expected first popconfirm to open")
	}

	second := NewInstance(rtui.Props{propComponentID: "second.confirm", propTitle: "Second?"})
	second.OnMount()
	defer second.Destroy()
	if !second.setOpen(true, TriggerClick) {
		t.Fatal("expected second popconfirm to open")
	}

	middleware := NewMiddleware()
	if next := middleware.Before(action.NewAction(action.ActionCancel)); next != nil {
		t.Fatal("escape should be intercepted when a popconfirm closes")
	}
	if !first.open {
		t.Fatal("older popconfirm should remain open after ESC closes topmost")
	}
	if second.open {
		t.Fatal("topmost popconfirm should close after ESC")
	}
}

func TestPopconfirmMiddlewareClickOutsideClosesOpenPopconfirm(t *testing.T) {
	popconfirmRegistryGlobal.reset()
	defer popconfirmRegistryGlobal.reset()

	inst := NewInstance(rtui.Props{propComponentID: "delete.confirm", propTitle: "Delete?"})
	inst.SetBounds(10, 5, 12, 1)
	inst.OnMount()
	defer inst.Destroy()
	inst.setOpen(true, TriggerClick)

	middleware := NewMiddleware()
	act := action.NewAction(action.ActionClick).WithPayload(runtimemsg.NewMouseMsg(1, 1, runtimemsg.MouseLeft, runtimemsg.MouseActionPress))
	if next := middleware.Before(act); next == nil {
		t.Fatal("outside click should continue dispatch after closing popconfirm")
	}
	if inst.open {
		t.Fatal("popconfirm should close after outside click")
	}
}

func TestPopconfirmMiddlewareLeavesAnchorAndOverlayClicksAlone(t *testing.T) {
	popconfirmRegistryGlobal.reset()
	defer popconfirmRegistryGlobal.reset()

	inst := NewInstance(rtui.Props{
		propComponentID: "delete.confirm",
		propTitle:       "Delete?",
		propDescription: "This action cannot be undone.",
	})
	inst.SetBounds(10, 5, 12, 1)
	inst.OnMount()
	defer inst.Destroy()
	inst.setOpen(true, TriggerClick)

	middleware := NewMiddleware()
	anchorClick := action.NewAction(action.ActionClick).WithPayload(runtimemsg.NewMouseMsg(11, 5, runtimemsg.MouseLeft, runtimemsg.MouseActionPress))
	if next := middleware.Before(anchorClick); next == nil {
		t.Fatal("anchor click should continue dispatch")
	}
	if !inst.open {
		t.Fatal("anchor click should keep popconfirm open")
	}

	x, y, _, _ := inst.overlayBounds()
	overlayClick := action.NewAction(action.ActionClick).WithPayload(runtimemsg.NewMouseMsg(x+1, y+1, runtimemsg.MouseLeft, runtimemsg.MouseActionPress))
	if next := middleware.Before(overlayClick); next == nil {
		t.Fatal("overlay click should continue dispatch")
	}
	if !inst.open {
		t.Fatal("overlay click should keep popconfirm open")
	}
}

func TestInstallAddsMiddlewareOnce(t *testing.T) {
	host := &fakeInstallerHost{}

	Install(host)
	Install(host)

	if host.middlewareCount != 1 {
		t.Fatalf("middlewareCount = %d, want 1", host.middlewareCount)
	}
}

func actionRowButtons(t *testing.T, row rtui.VNode) []*button.VNode {
	t.Helper()
	children := row.Children()
	if len(children) == 0 {
		t.Fatal("expected action buttons")
	}
	result := make([]*button.VNode, 0, len(children))
	for _, child := range children {
		btn, ok := child.(*button.VNode)
		if !ok {
			t.Fatalf("action child = %T, want *button.VNode", child)
		}
		result = append(result, btn)
	}
	return result
}

func containsVNodeText(node rtui.VNode, want string) bool {
	if node == nil {
		return false
	}
	if props := node.Props(); props != nil {
		if props.GetString("content") == want || props.GetString("label") == want {
			return true
		}
	}
	if textNode, ok := node.(*textcomp.VNode); ok && textNode.Content() == want {
		return true
	}
	for _, child := range node.Children() {
		if containsVNodeText(child, want) {
			return true
		}
	}
	return false
}

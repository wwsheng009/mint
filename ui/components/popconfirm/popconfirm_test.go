package popconfirm

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/button"
	textcomp "github.com/wwsheng009/mint/ui/components/text"
)

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

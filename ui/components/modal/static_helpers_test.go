package modal

import (
	"testing"

	runtimeintent "github.com/wwsheng009/mint/runtime/intent"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/button"
)

type testHelperFollowUpIntent struct{}

func (testHelperFollowUpIntent) IntentType() string { return "modal.test.followup" }

func TestConfirmHelper_DefaultFooterAndOpenState(t *testing.T) {
	vnode := Confirm("Delete Item", "This action cannot be undone.").BuildVNode()

	if !vnode.IsOpen() {
		t.Fatal("confirm helper should open the modal by default")
	}
	if vnode.Title() != "Delete Item" {
		t.Fatalf("Title = %q, want %q", vnode.Title(), "Delete Item")
	}
	if vnode.BorderStyle() != "rounded" {
		t.Fatalf("BorderStyle = %q, want %q", vnode.BorderStyle(), "rounded")
	}

	buttons := footerButtons(t, vnode.Footer())
	if len(buttons) != 2 {
		t.Fatalf("footer button count = %d, want 2", len(buttons))
	}

	if got := buttonLabel(buttons[0]); got != "Cancel" {
		t.Fatalf("cancel label = %q, want %q", got, "Cancel")
	}
	if buttons[0].Variant() != button.VariantSecondary {
		t.Fatalf("cancel variant = %v, want %v", buttons[0].Variant(), button.VariantSecondary)
	}
	cancelIntent, ok := buttons[0].PressIntent().(staticActionIntent)
	if !ok || cancelIntent.Action != staticActionCancel {
		t.Fatalf("cancel press intent = %#v, want static cancel intent", buttons[0].PressIntent())
	}

	if got := buttonLabel(buttons[1]); got != "OK" {
		t.Fatalf("confirm label = %q, want %q", got, "OK")
	}
	if buttons[1].Variant() != button.VariantPrimary {
		t.Fatalf("confirm variant = %v, want %v", buttons[1].Variant(), button.VariantPrimary)
	}
	confirmIntent, ok := buttons[1].PressIntent().(staticActionIntent)
	if !ok || confirmIntent.Action != staticActionConfirm {
		t.Fatalf("confirm press intent = %#v, want static confirm intent", buttons[1].PressIntent())
	}
}

func TestStatusHelpers_DefaultTitleStyleAndFooter(t *testing.T) {
	tests := []struct {
		name        string
		builder     *Builder
		wantTitle   string
		wantButton  string
		wantVariant button.Variant
		wantAction  staticAction
	}{
		{name: "info", builder: Info("Heads up"), wantTitle: "Info", wantButton: "OK", wantVariant: button.VariantPrimary, wantAction: staticActionAcknowledge},
		{name: "success", builder: Success("Done"), wantTitle: "Success", wantButton: "OK", wantVariant: button.VariantSuccess, wantAction: staticActionAcknowledge},
		{name: "error", builder: Error("Boom"), wantTitle: "Error", wantButton: "OK", wantVariant: button.VariantDanger, wantAction: staticActionAcknowledge},
		{name: "warning", builder: Warning("Careful"), wantTitle: "Warning", wantButton: "OK", wantVariant: button.VariantSecondary, wantAction: staticActionAcknowledge},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vnode := tt.builder.BuildVNode()
			if vnode.Title() != tt.wantTitle {
				t.Fatalf("Title = %q, want %q", vnode.Title(), tt.wantTitle)
			}

			buttons := footerButtons(t, vnode.Footer())
			if len(buttons) != 1 {
				t.Fatalf("footer button count = %d, want 1", len(buttons))
			}
			if got := buttonLabel(buttons[0]); got != tt.wantButton {
				t.Fatalf("button label = %q, want %q", got, tt.wantButton)
			}
			if buttons[0].Variant() != tt.wantVariant {
				t.Fatalf("button variant = %v, want %v", buttons[0].Variant(), tt.wantVariant)
			}
			pressIntent, ok := buttons[0].PressIntent().(staticActionIntent)
			if !ok || pressIntent.Action != tt.wantAction {
				t.Fatalf("press intent = %#v, want static action %q", buttons[0].PressIntent(), tt.wantAction)
			}
		})
	}
}

func TestHelperButtonsCloseTopmostModalAndEmitFollowUp(t *testing.T) {
	resetRegistry(t)

	rt := runtimeintent.NewRuntime()
	var followUps int
	unregister := runtimeintent.RegisterTypedRuntime(rt, func(_ *runtimeintent.ActionContext, _ testHelperFollowUpIntent) runtimeintent.IntentResult {
		followUps++
		return runtimeintent.HandledResult()
	})
	defer unregister()

	prevRT := rtui.GetGlobalIntentRuntime()
	rtui.SetGlobalIntentRuntime(rt)
	t.Cleanup(func() {
		rtui.SetGlobalIntentRuntime(prevRT)
	})

	lower := NewInstance(rtui.Props{"key": "lower", "isOpen": true})
	upperBuilder := Confirm("Confirm", "Proceed?", WithConfirmIntent(testHelperFollowUpIntent{}))
	upper := upperBuilder.BuildInstance()
	registerForTest(lower)
	registerForTest(upper)

	buttons := footerButtons(t, upperBuilder.BuildVNode().Footer())
	confirmIntent, ok := buttons[1].PressIntent().(staticActionIntent)
	if !ok {
		t.Fatalf("confirm press intent type = %T, want staticActionIntent", buttons[1].PressIntent())
	}

	result := rt.Emit(confirmIntent)
	if result.Error != nil {
		t.Fatalf("Emit returned error: %v", result.Error)
	}
	if !result.Handled {
		t.Fatal("Emit should be handled")
	}
	if !lower.isOpen {
		t.Fatal("lower modal should remain open")
	}
	if upper.isOpen {
		t.Fatal("topmost modal should close")
	}
	if followUps != 1 {
		t.Fatalf("followUp count = %d, want 1", followUps)
	}
}

func TestStaticHelpersCanBeFurtherConfigured(t *testing.T) {
	vnode := Success("Saved").
		Title("Saved Changes").
		Width(60).
		CloseOnBackdrop(false).
		BuildVNode()

	if vnode.Title() != "Saved Changes" {
		t.Fatalf("Title = %q, want %q", vnode.Title(), "Saved Changes")
	}
	if vnode.Width() != 60 {
		t.Fatalf("Width = %d, want 60", vnode.Width())
	}
	if vnode.CloseOnBackdrop() {
		t.Fatal("CloseOnBackdrop should be false after override")
	}

	buttons := footerButtons(t, vnode.Footer())
	if len(buttons) != 1 {
		t.Fatalf("footer button count = %d, want 1", len(buttons))
	}
	if got := buttonLabel(buttons[0]); got != "OK" {
		t.Fatalf("button label = %q, want %q", got, "OK")
	}
}

func footerButtons(t *testing.T, footer rtui.VNode) []*button.VNode {
	t.Helper()
	if footer == nil {
		t.Fatal("footer should not be nil")
	}

	children := footer.Children()
	buttons := make([]*button.VNode, 0, len(children))
	for _, child := range children {
		btn, ok := child.(*button.VNode)
		if !ok {
			t.Fatalf("footer child type = %T, want *button.VNode", child)
		}
		buttons = append(buttons, btn)
	}
	return buttons
}

func buttonLabel(v *button.VNode) string {
	if v == nil {
		return ""
	}
	if label, ok := v.Props()["label"].(string); ok {
		return label
	}
	return ""
}

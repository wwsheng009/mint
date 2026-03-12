package modal

import (
	"strings"
	"testing"

	"github.com/wwsheng009/mint/internal/reconciler"
	"github.com/wwsheng009/mint/internal/render"
	"github.com/wwsheng009/mint/runtime/action"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	"github.com/wwsheng009/mint/runtime/paint"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	newtext "github.com/wwsheng009/mint/ui/components/text"
)

func TestInstanceGetBoxModelReservesHeaderRows(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"title":   "Settings",
		"isOpen":  true,
		"padding": 1,
	})

	boxModel := inst.GetBoxModel()
	if boxModel.Padding.Top != 3 {
		t.Fatalf("top padding = %d, want 3", boxModel.Padding.Top)
	}
	if boxModel.Padding.Left != 1 || boxModel.Padding.Right != 1 || boxModel.Padding.Bottom != 1 {
		t.Fatalf("unexpected modal padding: %+v", boxModel.Padding)
	}
	if boxModel.Border.Label != "" {
		t.Fatalf("border label should be empty when modal paints its own header, got %q", boxModel.Border.Label)
	}
}

func TestModalAdapterUsesInstanceBoxModelAndFlexStyle(t *testing.T) {
	vnode := New().
		SetTitle("Dialog").
		SetOpen(true).
		SetContent(newtext.New("body")).
		SetFooter(newtext.New("footer"))

	fiber := reconciler.CreateFiberFromVNode(vnode)
	adapter := render.NewFiberToNodeAdapterPure(fiber)

	boxModel := adapter.GetBoxModel()
	if boxModel.Padding.Top != 2 {
		t.Fatalf("adapter top padding = %d, want 2", boxModel.Padding.Top)
	}
	if adapter.GetFlexStyle() == nil {
		t.Fatal("adapter should expose modal flex style from the instance")
	}
}

func TestInstancePaintKeepsHeaderRowDisplayWidth(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"title":      "设置中心",
		"isOpen":     true,
		"width":      20,
		"height":     8,
		"closeable":  true,
		"closeOnEsc": true,
	})

	cmds := inst.Paint(4, 6)
	headerRow, ok := findCommandAt(cmds, 4, 7)
	if !ok {
		t.Fatal("header row command not found")
	}
	if got := paint.StringWidth(headerRow.Text); got != 20 {
		t.Fatalf("header row display width = %d, want 20", got)
	}
	if _, ok := findCommandContaining(cmds, 7, "设置中心"); !ok {
		t.Fatal("header title command not found")
	}
	if _, ok := findCommandContaining(cmds, 7, "ESC"); !ok {
		t.Fatal("close hint command not found")
	}
}

func TestModalMiddlewareEscapeClosesTopmostOnly(t *testing.T) {
	resetRegistry(t)

	lower := NewInstance(rtui.Props{"isOpen": true, "closeable": true, "closeOnEsc": true})
	upper := NewInstance(rtui.Props{"isOpen": true, "closeable": true, "closeOnEsc": true})
	registerForTest(lower)
	registerForTest(upper)

	middleware := NewModalMiddleware()
	result := middleware.Before(&action.Action{Type: action.ActionCancel})
	if result != nil {
		t.Fatal("ESC should be intercepted while a modal is open")
	}
	if !lower.isOpen {
		t.Fatal("lower modal should remain open")
	}
	if upper.isOpen {
		t.Fatal("topmost modal should close first")
	}
}

func TestModalMiddlewareSwallowsOutsideClickWhenBackdropCloseDisabled(t *testing.T) {
	resetRegistry(t)

	inst := NewInstance(rtui.Props{
		"isOpen":          true,
		"closeable":       true,
		"closeOnBackdrop": false,
	})
	inst.SetBounds(10, 10, 20, 6)
	registerForTest(inst)

	middleware := NewModalMiddleware()
	result := middleware.Before(&action.Action{
		Type: action.ActionClick,
		Payload: &runtimemsg.MouseMsg{
			X:      0,
			Y:      0,
			Action: runtimemsg.MouseActionPress,
		},
	})

	if result != nil {
		t.Fatal("outside click should be swallowed while modal is open")
	}
	if !inst.isOpen {
		t.Fatal("modal should stay open when backdrop close is disabled")
	}
}

func findCommandAt(cmds []paint.DrawCmd, x, y int) (paint.DrawCmd, bool) {
	for _, cmd := range cmds {
		if cmd.X == x && cmd.Y == y {
			return cmd, true
		}
	}
	return paint.DrawCmd{}, false
}

func findCommandContaining(cmds []paint.DrawCmd, y int, needle string) (paint.DrawCmd, bool) {
	for _, cmd := range cmds {
		if cmd.Y == y && strings.Contains(cmd.Text, needle) {
			return cmd, true
		}
	}
	return paint.DrawCmd{}, false
}

func resetRegistry(t *testing.T) {
	t.Helper()
	previous := globalRegistry
	globalRegistry = &modalRegistry{
		modals: make(map[*Instance]bool),
	}
	t.Cleanup(func() {
		globalRegistry = previous
	})
}

func registerForTest(inst *Instance) {
	globalRegistry.register(inst)
	inst.registered = true
}

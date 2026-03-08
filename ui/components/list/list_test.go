package list

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/action"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

func TestInstance_HandleAction_ScrollWithMousePayload(t *testing.T) {
	rows := []string{"a", "b", "c", "d", "e", "f"}
	inst := NewInstance(rtui.Props{
		"rows":           rows,
		"viewportHeight": 3,
	})

	mouseMsg := runtimemsg.NewMouseMsgWithDelta(0, 0, -1, runtimemsg.MouseActionWheel)
	act := action.NewActionWithPayload(action.ActionScroll, mouseMsg)
	if !inst.HandleAction(act) {
		t.Fatal("ActionScroll with MouseMsg payload should be handled")
	}
	if inst.GetScrollOffset() != 1 {
		t.Fatalf("offset = %d, want 1", inst.GetScrollOffset())
	}
}

func TestInstance_PaintWithBorder_ShowsScrollbarWhenScrollable(t *testing.T) {
	rows := []string{"a", "b", "c", "d", "e", "f"}
	inst := NewInstance(rtui.Props{
		"rows":           rows,
		"viewportHeight": 3,
		"showBorder":     true,
		"showScrollbar":  true,
	})

	cmds := inst.Paint(0, 0)
	width := inst.calculateWidth()
	hasScrollbar := false
	for _, cmd := range cmds {
		if cmd.X == width-1 && (cmd.Text == "█" || cmd.Text == "│") {
			hasScrollbar = true
			break
		}
	}
	if !hasScrollbar {
		t.Fatal("expected vertical scrollbar draw commands")
	}
}

func TestInstance_RowStyleFn_AppliesToRenderedRows(t *testing.T) {
	rows := []string{"first", "second"}
	custom := style.Style{FG: style.Green}
	inst := NewInstance(rtui.Props{
		"rows":          rows,
		"showBorder":    false,
		"showSeparator": false,
		"rowStyleFn": func(_ int, _ string) style.Style {
			return custom
		},
	})

	cmds := inst.Paint(0, 0)
	if len(cmds) < 2 {
		t.Fatalf("cmd count = %d, want >= 2", len(cmds))
	}
	if cmds[0].Style.FG != custom.FG {
		t.Fatalf("first row fg = %q, want %q", cmds[0].Style.FG, custom.FG)
	}
	if cmds[1].Style.FG != custom.FG {
		t.Fatalf("second row fg = %q, want %q", cmds[1].Style.FG, custom.FG)
	}
}

func TestInstance_SetProps_WithoutScrollOffsetPreservesInternalScroll(t *testing.T) {
	rows := []string{"a", "b", "c", "d", "e", "f"}
	inst := NewInstance(rtui.Props{
		"rows":           rows,
		"viewportHeight": 3,
	})
	inst.scrollOffset = 2

	inst.SetProps(rtui.Props{
		"rows":           rows,
		"viewportHeight": 3,
		"showBorder":     true,
	})

	if inst.scrollOffset != 2 {
		t.Fatalf("offset after SetProps = %d, want 2", inst.scrollOffset)
	}
}

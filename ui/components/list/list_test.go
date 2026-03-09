package list

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
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

func TestInstance_HandleAction_ClickSelectsVisibleRowWithHeaderAndBorder(t *testing.T) {
	rows := []string{"first", "second", "third", "fourth"}
	inst := NewInstance(rtui.Props{
		"header":         "Items",
		"rows":           rows,
		"viewportHeight": 3,
		"showBorder":     true,
		"showSeparator":  true,
		"scrollOffset":   1,
	})
	inst.SetBounds(0, 0, 20, 6)

	mouseMsg := runtimemsg.NewMouseMsg(1, 4, runtimemsg.MouseLeft, runtimemsg.MouseActionPress)
	mouseMsg.LocalY = 4
	act := action.NewActionWithPayload(action.ActionClick, mouseMsg)

	if !inst.HandleAction(act) {
		t.Fatal("click on visible row should be handled")
	}
	if inst.GetSelectedIndex() != 2 {
		t.Fatalf("selectedIndex = %d, want 2", inst.GetSelectedIndex())
	}
}

func TestInstance_HandleAction_NavigateDownWorksWhenScrollDisabled(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"rows":        []string{"a", "b", "c"},
		"allowScroll": false,
	})

	if !inst.HandleAction(action.NewAction(action.ActionNavigateDown)) {
		t.Fatal("navigate down should still work when scrolling is disabled")
	}
	if inst.GetSelectedIndex() != 0 {
		t.Fatalf("selectedIndex = %d, want 0", inst.GetSelectedIndex())
	}
}

func TestInstance_HandleAction_NavigateDownEmitsFieldChangeIntent(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"rows":         []string{"a", "b", "c"},
		"changeIntent": intent.BindField("selected_row"),
	})

	var emitted []intent.Intent
	inst.SetIntentEmitter(func(i intent.Intent) {
		emitted = append(emitted, i)
	})

	if !inst.HandleAction(action.NewAction(action.ActionNavigateDown)) {
		t.Fatal("navigate down should be handled")
	}
	if len(emitted) != 1 {
		t.Fatalf("emitted intents = %d, want 1", len(emitted))
	}
	fieldChange, ok := emitted[0].(intent.FieldChangeIntent)
	if !ok {
		t.Fatalf("emitted intent = %T, want intent.FieldChangeIntent", emitted[0])
	}
	if fieldChange.Field != "selected_row" {
		t.Fatalf("field = %q, want selected_row", fieldChange.Field)
	}
	if fieldChange.Value != "0" {
		t.Fatalf("value = %q, want 0", fieldChange.Value)
	}
}

func TestInstance_HandleAction_HomeMovesSelectionEvenWhenScrollOffsetIsZero(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"rows":          []string{"a", "b", "c", "d"},
		"selectedIndex": 2,
		"scrollOffset":  0,
	})

	if !inst.HandleAction(action.NewAction(action.ActionNavigateHome)) {
		t.Fatal("home should move selection to first row")
	}
	if inst.GetSelectedIndex() != 0 {
		t.Fatalf("selectedIndex = %d, want 0", inst.GetSelectedIndex())
	}
}

func TestInstance_SetProps_ClampsSelectedIndexWhenRowsShrink(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"rows":          []string{"a", "b", "c"},
		"selectedIndex": 2,
	})

	inst.SetProps(rtui.Props{
		"rows": []string{"a"},
	})

	if inst.GetSelectedIndex() != 0 {
		t.Fatalf("selectedIndex = %d, want 0", inst.GetSelectedIndex())
	}
}

func TestInstance_PaintWithBorder_UsesBoundsWidthForTruncation(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"rows":       []string{"this is a very long row"},
		"showBorder": true,
	})
	inst.SetBounds(0, 0, 12, 4)

	cmds := inst.Paint(0, 0)
	if len(cmds) == 0 {
		t.Fatal("expected paint commands")
	}
	if cmds[0].Text != "┌──────────┐" {
		t.Fatalf("top border = %q, want %q", cmds[0].Text, "┌──────────┐")
	}
	if cmds[1].Text != "│ this ... │" {
		t.Fatalf("row line = %q, want %q", cmds[1].Text, "│ this ... │")
	}
}

func TestInstance_PaintWithoutBorder_ShowsCheckboxMarkers(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"rows":           []string{"alpha", "beta"},
		"showBorder":     false,
		"showSeparator":  false,
		"selectionMode":  SelectionMultiple,
		"checkedIndices": []int{1},
	})

	cmds := inst.Paint(0, 0)
	if len(cmds) != 2 {
		t.Fatalf("cmd count = %d, want 2", len(cmds))
	}
	if cmds[0].Text != "[ ] alpha" {
		t.Fatalf("first row = %q, want %q", cmds[0].Text, "[ ] alpha")
	}
	if cmds[1].Text != "[x] beta" {
		t.Fatalf("second row = %q, want %q", cmds[1].Text, "[x] beta")
	}
}

func TestInstance_HandleAction_MultiSelectTogglesCheckedIndices(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"rows":          []string{"a", "b", "c"},
		"selectionMode": SelectionMultiple,
		"selectedIndex": 1,
	})

	if !inst.HandleAction(action.NewAction(action.ActionSelect)) {
		t.Fatal("select should be handled in multi-select mode")
	}
	if !equalInts(inst.GetCheckedIndices(), []int{1}) {
		t.Fatalf("checkedIndices = %v, want [1]", inst.GetCheckedIndices())
	}
	if !inst.HandleAction(action.NewAction(action.ActionSelect)) {
		t.Fatal("second select should be handled in multi-select mode")
	}
	if len(inst.GetCheckedIndices()) != 0 {
		t.Fatalf("checkedIndices = %v, want []", inst.GetCheckedIndices())
	}
}

func TestInstance_HandleAction_SingleSelectReplacesCheckedIndex(t *testing.T) {
	rows := []string{"first", "second", "third"}
	inst := NewInstance(rtui.Props{
		"rows":           rows,
		"selectionMode":  SelectionSingle,
		"checkedIndices": []int{0},
		"selectedIndex":  0,
		"showBorder":     true,
		"showSeparator":  true,
		"viewportHeight": 3,
	})
	inst.SetBounds(0, 0, 20, 6)

	mouseMsg := runtimemsg.NewMouseMsg(1, 3, runtimemsg.MouseLeft, runtimemsg.MouseActionPress)
	mouseMsg.LocalY = 3
	act := action.NewActionWithPayload(action.ActionClick, mouseMsg)
	if !inst.HandleAction(act) {
		t.Fatal("click should be handled in single-select mode")
	}
	if inst.GetSelectedIndex() != 2 {
		t.Fatalf("selectedIndex = %d, want 2", inst.GetSelectedIndex())
	}
	if !equalInts(inst.GetCheckedIndices(), []int{2}) {
		t.Fatalf("checkedIndices = %v, want [2]", inst.GetCheckedIndices())
	}
}

func TestInstance_HandleAction_MultiSelectEmitsFieldChangeIntent(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"rows":            []string{"a", "b", "c"},
		"selectionMode":   SelectionMultiple,
		"selectedIndex":   1,
		"selectionIntent": intent.BindField("checked_rows"),
	})

	var emitted []intent.Intent
	inst.SetIntentEmitter(func(i intent.Intent) {
		emitted = append(emitted, i)
	})

	if !inst.HandleAction(action.NewAction(action.ActionSelect)) {
		t.Fatal("select should be handled")
	}
	if len(emitted) != 1 {
		t.Fatalf("emitted intents = %d, want 1", len(emitted))
	}
	fieldChange, ok := emitted[0].(intent.FieldChangeIntent)
	if !ok {
		t.Fatalf("emitted intent = %T, want intent.FieldChangeIntent", emitted[0])
	}
	if fieldChange.Field != "checked_rows" {
		t.Fatalf("field = %q, want checked_rows", fieldChange.Field)
	}
	if fieldChange.Value != "1" {
		t.Fatalf("value = %q, want 1", fieldChange.Value)
	}

	inst.HandleAction(action.NewAction(action.ActionNavigateDown))
	inst.HandleAction(action.NewAction(action.ActionSelect))
	last, ok := emitted[len(emitted)-1].(intent.FieldChangeIntent)
	if !ok {
		t.Fatalf("last emitted intent = %T, want intent.FieldChangeIntent", emitted[len(emitted)-1])
	}
	if last.Value != "1,2" {
		t.Fatalf("value = %q, want 1,2", last.Value)
	}
}

func TestInstance_SetProps_WithoutCheckedIndicesPreservesUncontrolledSelection(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"rows":          []string{"a", "b", "c"},
		"selectionMode": SelectionMultiple,
	})
	inst.checkedIndices = []int{1}

	inst.SetProps(rtui.Props{
		"rows":          []string{"a", "b", "c"},
		"selectionMode": SelectionMultiple,
		"showBorder":    true,
	})

	if !equalInts(inst.GetCheckedIndices(), []int{1}) {
		t.Fatalf("checkedIndices = %v, want [1]", inst.GetCheckedIndices())
	}
}

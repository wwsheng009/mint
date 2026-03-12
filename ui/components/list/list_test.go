package list

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

type bubbleCaptureParent struct {
	intents []intent.Intent
}

func (p *bubbleCaptureParent) Key() string                        { return "bubble-parent" }
func (p *bubbleCaptureParent) SetKey(string)                      {}
func (p *bubbleCaptureParent) Init(rtui.Props)                    {}
func (p *bubbleCaptureParent) Destroy()                           {}
func (p *bubbleCaptureParent) OnMount()                           {}
func (p *bubbleCaptureParent) OnUnmount()                         {}
func (p *bubbleCaptureParent) Parent() interface{}                { return nil }
func (p *bubbleCaptureParent) SetParent(rtui.ComponentInstance)   {}
func (p *bubbleCaptureParent) SetProps(rtui.Props) bool           { return false }
func (p *bubbleCaptureParent) GetProps() rtui.Props               { return rtui.Props{} }
func (p *bubbleCaptureParent) MarkDirty()                         {}
func (p *bubbleCaptureParent) IsDirty() bool                      { return false }
func (p *bubbleCaptureParent) GetContext() *rtui.ComponentContext { return nil }
func (p *bubbleCaptureParent) ClearDirty()                        {}
func (p *bubbleCaptureParent) HandleIntent(i intent.Intent) bool {
	p.intents = append(p.intents, i)
	return true
}

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

func TestBuilder_FluentEnhancements(t *testing.T) {
	vnode := NewBuilder().
		ComponentID("orders.list").
		SearchQuery("warn").
		ShowSearchStats(true).
		MatchStyle(style.Style{FG: style.Yellow}).
		InitialScrollOffset(2).
		SelectedIndexControlled(3).
		InitialCheckedIndices(1, 4).
		MultiSelect().
		BuildVNode()

	if vnode.componentID != "orders.list" {
		t.Fatalf("componentID = %q, want orders.list", vnode.componentID)
	}
	if vnode.scrollOffset != 2 || vnode.scrollOffsetControlled {
		t.Fatalf("scroll state = (%d,%v), want (2,false)", vnode.scrollOffset, vnode.scrollOffsetControlled)
	}
	if vnode.selectedIndex != 3 || !vnode.selectedIndexControlled {
		t.Fatalf("selected state = (%d,%v), want (3,true)", vnode.selectedIndex, vnode.selectedIndexControlled)
	}
	if !equalInts(vnode.checkedIndices, []int{1, 4}) || vnode.checkedIndicesControlled {
		t.Fatalf("checked state = (%v,%v), want ([1 4],false)", vnode.checkedIndices, vnode.checkedIndicesControlled)
	}
	if vnode.selectionMode != SelectionMultiple {
		t.Fatalf("selectionMode = %v, want multi", vnode.selectionMode)
	}
	if vnode.searchQuery != "warn" || !vnode.showSearchStats {
		t.Fatalf("search props = (%q,%v), want (warn,true)", vnode.searchQuery, vnode.showSearchStats)
	}
	if vnode.matchStyle.FG != style.Yellow {
		t.Fatalf("matchStyle fg = %q, want %q", vnode.matchStyle.FG, style.Yellow)
	}
}

func TestInstance_SetProps_SelectedIndexWithoutControlActsAsInitialValue(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"rows":          []string{"a", "b", "c"},
		"selectedIndex": 1,
	})

	if !inst.HandleAction(action.NewAction(action.ActionNavigateDown)) {
		t.Fatal("navigate down should be handled")
	}
	inst.SetProps(rtui.Props{
		"rows":          []string{"a", "b", "c"},
		"selectedIndex": 1,
	})

	if inst.GetSelectedIndex() != 2 {
		t.Fatalf("selectedIndex = %d, want 2", inst.GetSelectedIndex())
	}
}

func TestInstance_EmitStateChangeIntentOnNavigation(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"componentID":    "orders.list",
		"rows":           []string{"10", "20"},
		"viewportHeight": 5,
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
	change, ok := emitted[0].(StateChangeIntent)
	if !ok {
		t.Fatalf("emitted intent = %T, want StateChangeIntent", emitted[0])
	}
	if change.ComponentID != "orders.list" {
		t.Fatalf("componentID = %q, want orders.list", change.ComponentID)
	}
	if change.SelectedIndex != 0 || change.SelectedRow != "10" {
		t.Fatalf("selection = (%d,%q), want (0,%q)", change.SelectedIndex, change.SelectedRow, "10")
	}
	if change.ScrollOffset != 0 || change.VisibleRows != 2 || change.TotalRows != 2 {
		t.Fatalf("viewport = (%d,%d,%d), want (0,2,2)", change.ScrollOffset, change.VisibleRows, change.TotalRows)
	}
}

func TestInstance_EmitStateChangeIntentOnScroll(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"componentID":    "orders.list",
		"rows":           []string{"a", "b", "c", "d", "e"},
		"viewportHeight": 3,
	})

	var emitted []intent.Intent
	inst.SetIntentEmitter(func(i intent.Intent) {
		emitted = append(emitted, i)
	})

	mouseMsg := runtimemsg.NewMouseMsgWithDelta(0, 0, -1, runtimemsg.MouseActionWheel)
	act := action.NewActionWithPayload(action.ActionScroll, mouseMsg)
	if !inst.HandleAction(act) {
		t.Fatal("scroll should be handled")
	}
	if len(emitted) != 1 {
		t.Fatalf("emitted intents = %d, want 1", len(emitted))
	}
	change, ok := emitted[0].(StateChangeIntent)
	if !ok {
		t.Fatalf("emitted intent = %T, want StateChangeIntent", emitted[0])
	}
	if change.ScrollOffset != 1 || change.VisibleRows != 3 || change.ViewportHeight != 3 {
		t.Fatalf("scroll state = (%d,%d,%d), want (1,3,3)", change.ScrollOffset, change.VisibleRows, change.ViewportHeight)
	}
	if change.SelectedIndex != -1 || change.SelectedRow != "" {
		t.Fatalf("selection = (%d,%q), want (-1,\"\")", change.SelectedIndex, change.SelectedRow)
	}
}

func TestInstance_EmitStateChangeIntentOnMultiSelect(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"componentID":    "orders.list",
		"rows":           []string{"a", "b", "c"},
		"selectionMode":  SelectionMultiple,
		"selectedIndex":  1,
		"viewportHeight": 3,
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
	change, ok := emitted[0].(StateChangeIntent)
	if !ok {
		t.Fatalf("emitted intent = %T, want StateChangeIntent", emitted[0])
	}
	if change.SelectionMode != SelectionMultiple {
		t.Fatalf("selectionMode = %v, want multi", change.SelectionMode)
	}
	if !equalInts(change.CheckedIndices, []int{1}) {
		t.Fatalf("checkedIndices = %v, want [1]", change.CheckedIndices)
	}
	if len(change.CheckedRows) != 1 || change.CheckedRows[0] != "b" {
		t.Fatalf("checkedRows = %v, want [b]", change.CheckedRows)
	}
}

func TestInstance_HandleIntent_SelectByIndexRespectsComponentID(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"componentID": "orders.list",
		"rows":        []string{"a", "b", "c"},
	})

	if inst.HandleIntent(SelectByIndex("other.list", 2)) {
		t.Fatal("intent for different componentID should be ignored")
	}
	if inst.GetSelectedIndex() != -1 {
		t.Fatalf("selectedIndex = %d, want -1", inst.GetSelectedIndex())
	}
	if !inst.HandleIntent(SelectByIndex("orders.list", 2)) {
		t.Fatal("matching SelectByIndex intent should be handled")
	}
	if inst.GetSelectedIndex() != 2 {
		t.Fatalf("selectedIndex = %d, want 2", inst.GetSelectedIndex())
	}
}

func TestInstance_HandleIntent_ScrollToAndClearSelection(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"componentID":    "orders.list",
		"rows":           []string{"a", "b", "c", "d", "e"},
		"viewportHeight": 2,
	})

	if !inst.HandleIntent(ScrollTo("orders.list", 3)) {
		t.Fatal("ScrollTo intent should be handled")
	}
	if inst.GetScrollOffset() != 3 {
		t.Fatalf("scrollOffset = %d, want 3", inst.GetScrollOffset())
	}
	if !inst.HandleIntent(SelectByIndex("orders.list", 4)) {
		t.Fatal("SelectByIndex should be handled")
	}
	if !inst.HandleIntent(ClearSelection("orders.list")) {
		t.Fatal("ClearSelection should be handled")
	}
	if inst.GetSelectedIndex() != -1 {
		t.Fatalf("selectedIndex = %d, want -1", inst.GetSelectedIndex())
	}
}

func TestInstance_HandleIntent_ToggleCheckedAndClearChecked(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"componentID":   "orders.list",
		"rows":          []string{"a", "b", "c"},
		"selectionMode": SelectionMultiple,
	})

	if !inst.HandleIntent(ToggleChecked("orders.list", 1)) {
		t.Fatal("ToggleChecked should be handled")
	}
	if inst.GetSelectedIndex() != 1 {
		t.Fatalf("selectedIndex = %d, want 1", inst.GetSelectedIndex())
	}
	if !equalInts(inst.GetCheckedIndices(), []int{1}) {
		t.Fatalf("checkedIndices = %v, want [1]", inst.GetCheckedIndices())
	}
	if !inst.HandleIntent(ClearChecked("orders.list")) {
		t.Fatal("ClearChecked should be handled")
	}
	if len(inst.GetCheckedIndices()) != 0 {
		t.Fatalf("checkedIndices = %v, want []", inst.GetCheckedIndices())
	}
}

func TestInstance_HandleAction_ControlledSelectedIndexPendingSurvivesStaleProps(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"rows":                    []string{"a", "b", "c"},
		"selectedIndex":           0,
		"selectedIndexControlled": true,
	})

	if !inst.HandleAction(action.NewAction(action.ActionNavigateDown)) {
		t.Fatal("navigate down should be handled")
	}
	if inst.GetSelectedIndex() != 1 {
		t.Fatalf("selectedIndex after action = %d, want 1", inst.GetSelectedIndex())
	}

	inst.SetProps(rtui.Props{
		"rows":                    []string{"a", "b", "c"},
		"selectedIndex":           0,
		"selectedIndexControlled": true,
	})
	if inst.GetSelectedIndex() != 1 {
		t.Fatalf("selectedIndex after stale props = %d, want 1", inst.GetSelectedIndex())
	}

	inst.SetProps(rtui.Props{
		"rows":                    []string{"a", "b", "c"},
		"selectedIndex":           1,
		"selectedIndexControlled": true,
	})
	if inst.GetSelectedIndex() != 1 {
		t.Fatalf("selectedIndex after fresh props = %d, want 1", inst.GetSelectedIndex())
	}
}

func TestInstance_HandleAction_ControlledScrollPendingSurvivesStaleProps(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"rows":                   []string{"a", "b", "c", "d", "e"},
		"viewportHeight":         2,
		"scrollOffset":           0,
		"scrollOffsetControlled": true,
	})

	mouseMsg := runtimemsg.NewMouseMsgWithDelta(0, 0, -1, runtimemsg.MouseActionWheel)
	act := action.NewActionWithPayload(action.ActionScroll, mouseMsg)
	if !inst.HandleAction(act) {
		t.Fatal("scroll should be handled")
	}
	if inst.GetScrollOffset() != 1 {
		t.Fatalf("scrollOffset after action = %d, want 1", inst.GetScrollOffset())
	}

	inst.SetProps(rtui.Props{
		"rows":                   []string{"a", "b", "c", "d", "e"},
		"viewportHeight":         2,
		"scrollOffset":           0,
		"scrollOffsetControlled": true,
	})
	if inst.GetScrollOffset() != 1 {
		t.Fatalf("scrollOffset after stale props = %d, want 1", inst.GetScrollOffset())
	}

	inst.SetProps(rtui.Props{
		"rows":                   []string{"a", "b", "c", "d", "e"},
		"viewportHeight":         2,
		"scrollOffset":           1,
		"scrollOffsetControlled": true,
	})
	if inst.GetScrollOffset() != 1 {
		t.Fatalf("scrollOffset after fresh props = %d, want 1", inst.GetScrollOffset())
	}
}

func TestInstance_HandleAction_ControlledCheckedIndicesPendingSurviveStaleProps(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"rows":                     []string{"a", "b", "c"},
		"selectionMode":            SelectionMultiple,
		"selectedIndex":            1,
		"checkedIndices":           []int{0},
		"checkedIndicesControlled": true,
	})

	if !inst.HandleAction(action.NewAction(action.ActionSelect)) {
		t.Fatal("select should be handled")
	}
	if !equalInts(inst.GetCheckedIndices(), []int{0, 1}) {
		t.Fatalf("checkedIndices after action = %v, want [0 1]", inst.GetCheckedIndices())
	}

	inst.SetProps(rtui.Props{
		"rows":                     []string{"a", "b", "c"},
		"selectionMode":            SelectionMultiple,
		"selectedIndex":            1,
		"checkedIndices":           []int{0},
		"checkedIndicesControlled": true,
	})
	if !equalInts(inst.GetCheckedIndices(), []int{0, 1}) {
		t.Fatalf("checkedIndices after stale props = %v, want [0 1]", inst.GetCheckedIndices())
	}

	inst.SetProps(rtui.Props{
		"rows":                     []string{"a", "b", "c"},
		"selectionMode":            SelectionMultiple,
		"selectedIndex":            1,
		"checkedIndices":           []int{0, 1},
		"checkedIndicesControlled": true,
	})
	if !equalInts(inst.GetCheckedIndices(), []int{0, 1}) {
		t.Fatalf("checkedIndices after fresh props = %v, want [0 1]", inst.GetCheckedIndices())
	}
}

func TestInstance_EmitLocalNavigationAndRowSelectIntents(t *testing.T) {
	parent := &bubbleCaptureParent{}
	inst := NewInstance(rtui.Props{
		"componentID": "orders.list",
		"rows":        []string{"a", "b", "c"},
	})
	inst.SetParent(parent)

	if !inst.HandleAction(action.NewAction(action.ActionNavigateDown)) {
		t.Fatal("navigate down should be handled")
	}
	if len(parent.intents) != 2 {
		t.Fatalf("captured intents = %d, want 2", len(parent.intents))
	}
	rowSelect, ok := parent.intents[0].(RowSelectIntent)
	if !ok {
		t.Fatalf("first intent = %T, want RowSelectIntent", parent.intents[0])
	}
	if rowSelect.ComponentID != "orders.list" || rowSelect.SelectedIndex != 0 || rowSelect.SelectedRow != "a" {
		t.Fatalf("row select = %#v, want componentID orders.list index 0 row a", rowSelect)
	}
	navigation, ok := parent.intents[1].(NavigationIntent)
	if !ok {
		t.Fatalf("second intent = %T, want NavigationIntent", parent.intents[1])
	}
	if navigation.ComponentID != "orders.list" || navigation.Direction != "down" || navigation.FromIndex != -1 || navigation.ToIndex != 0 {
		t.Fatalf("navigation = %#v, want componentID orders.list direction down from -1 to 0", navigation)
	}
}

func TestInstance_EmitLocalScrollIntent(t *testing.T) {
	parent := &bubbleCaptureParent{}
	inst := NewInstance(rtui.Props{
		"componentID":    "orders.list",
		"rows":           []string{"a", "b", "c", "d", "e"},
		"viewportHeight": 2,
	})
	inst.SetParent(parent)

	mouseMsg := runtimemsg.NewMouseMsgWithDelta(0, 0, -1, runtimemsg.MouseActionWheel)
	act := action.NewActionWithPayload(action.ActionScroll, mouseMsg)
	if !inst.HandleAction(act) {
		t.Fatal("scroll should be handled")
	}
	if len(parent.intents) != 1 {
		t.Fatalf("captured intents = %d, want 1", len(parent.intents))
	}
	scrollIntent, ok := parent.intents[0].(ScrollIntent)
	if !ok {
		t.Fatalf("intent = %T, want ScrollIntent", parent.intents[0])
	}
	if scrollIntent.ComponentID != "orders.list" || scrollIntent.Offset != 1 || scrollIntent.Delta != 1 || scrollIntent.ViewSize != 2 || scrollIntent.ContentSize != 5 {
		t.Fatalf("scroll intent = %#v, want componentID orders.list offset 1 delta 1 viewSize 2 contentSize 5", scrollIntent)
	}
}

func TestInstance_EmitLocalSelectionChangeIntent(t *testing.T) {
	parent := &bubbleCaptureParent{}
	inst := NewInstance(rtui.Props{
		"componentID":   "orders.list",
		"rows":          []string{"a", "b", "c"},
		"selectionMode": SelectionMultiple,
		"selectedIndex": 1,
	})
	inst.SetParent(parent)

	if !inst.HandleAction(action.NewAction(action.ActionSelect)) {
		t.Fatal("select should be handled")
	}
	found := false
	for _, emitted := range parent.intents {
		selectionChange, ok := emitted.(SelectionChangeIntent)
		if !ok {
			continue
		}
		found = true
		if selectionChange.ComponentID != "orders.list" || selectionChange.SelectionMode != SelectionMultiple {
			t.Fatalf("selection change = %#v, want componentID orders.list multi-select", selectionChange)
		}
		if !equalInts(selectionChange.CheckedIndices, []int{1}) {
			t.Fatalf("checkedIndices = %v, want [1]", selectionChange.CheckedIndices)
		}
		if len(selectionChange.CheckedRows) != 1 || selectionChange.CheckedRows[0] != "b" {
			t.Fatalf("checkedRows = %v, want [b]", selectionChange.CheckedRows)
		}
	}
	if !found {
		t.Fatal("expected SelectionChangeIntent to be bubbled to parent")
	}
}

func TestInstance_SearchQueryFiltersRowsAndAutoSelectsFirstMatch(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"rows":           []string{"alpha", "beta", "gamma", "betamax"},
		"searchQuery":    "beta",
		"showBorder":     false,
		"showSeparator":  false,
		"viewportHeight": 4,
	})

	if inst.GetSelectedIndex() != 1 {
		t.Fatalf("selectedIndex = %d, want 1", inst.GetSelectedIndex())
	}

	cmds := inst.Paint(0, 0)
	foundBeta := false
	foundBetamax := false
	foundAlpha := false
	for _, cmd := range cmds {
		if cmd.Text == "beta" {
			foundBeta = true
		}
		if cmd.Text == "betamax" {
			foundBetamax = true
		}
		if cmd.Text == "alpha" {
			foundAlpha = true
		}
	}
	if !foundBeta || !foundBetamax {
		t.Fatalf("expected filtered rows beta and betamax, got %#v", cmds)
	}
	if foundAlpha {
		t.Fatal("expected alpha to be filtered out")
	}
}

func TestInstance_HandleAction_SearchNextPrev(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"rows":        []string{"alpha", "beta", "gamma", "betamax"},
		"searchQuery": "beta",
	})

	if !inst.HandleAction(action.NewAction(action.ActionSearch).WithPayload("next")) {
		t.Fatal("search next should be handled")
	}
	if inst.GetSelectedIndex() != 3 {
		t.Fatalf("selectedIndex after next = %d, want 3", inst.GetSelectedIndex())
	}
	if !inst.HandleAction(action.NewAction(action.ActionSearch).WithPayload("prev")) {
		t.Fatal("search prev should be handled")
	}
	if inst.GetSelectedIndex() != 1 {
		t.Fatalf("selectedIndex after prev = %d, want 1", inst.GetSelectedIndex())
	}
}

func TestInstance_PaintSearchStats(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"rows":             []string{"alpha", "beta", "betamax"},
		"searchQuery":      "beta",
		"showSearchStats":  true,
		"showBorder":       false,
		"showSeparator":    false,
		"viewportHeight":   3,
		"searchStatsStyle": style.Style{FG: style.Green},
	})

	cmds := inst.Paint(0, 0)
	if len(cmds) < 2 {
		t.Fatalf("cmd count = %d, want >= 2", len(cmds))
	}
	if cmds[0].Text != `Search: "beta" 1/2` {
		t.Fatalf("stats line = %q, want %q", cmds[0].Text, `Search: "beta" 1/2`)
	}
	if cmds[0].Style.FG != style.Green {
		t.Fatalf("stats fg = %q, want %q", cmds[0].Style.FG, style.Green)
	}
}

func TestInstance_EmitLocalSearchStatsIntent(t *testing.T) {
	parent := &bubbleCaptureParent{}
	inst := NewInstance(rtui.Props{
		"componentID": "orders.list",
		"rows":        []string{"alpha", "beta", "betamax"},
	})
	inst.SetParent(parent)

	inst.SetProps(rtui.Props{
		"componentID": "orders.list",
		"rows":        []string{"alpha", "beta", "betamax"},
		"searchQuery": "beta",
	})

	found := false
	for _, emitted := range parent.intents {
		stats, ok := emitted.(SearchStatsIntent)
		if !ok {
			continue
		}
		found = true
		if stats.ComponentID != "orders.list" || stats.Query != "beta" || stats.Total != 2 || stats.Selected != 1 {
			t.Fatalf("search stats = %#v, want componentID orders.list query beta total 2 selected 1", stats)
		}
	}
	if !found {
		t.Fatal("expected SearchStatsIntent to be bubbled to parent")
	}
}

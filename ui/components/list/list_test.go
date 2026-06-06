package list

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/virtuallist"
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

func TestRowItem_Text(t *testing.T) {
	item := Item("Orders").
		WithPrefix("[*]").
		WithDescription("3 pending").
		WithSuffix("(hot)")

	if got := item.Text(); got != "[*] Orders - 3 pending (hot)" {
		t.Fatalf("Text() = %q, want %q", got, "[*] Orders - 3 pending (hot)")
	}
}

func TestVNode_ItemsAndRowsStayInSync(t *testing.T) {
	items := []RowItem{
		Item("Orders").WithPrefix("[*]").WithDescription("3 pending"),
		Item("Invoices").WithSuffix("(new)"),
	}

	vnode := New().
		SetItems(items).
		AddItem(Item("Shipments").WithDescription("ready"))

	if len(vnode.Items()) != 3 {
		t.Fatalf("Items len = %d, want 3", len(vnode.Items()))
	}
	wantRows := []string{
		"[*] Orders - 3 pending",
		"Invoices (new)",
		"Shipments - ready",
	}
	rows := vnode.Rows()
	for i := range wantRows {
		if rows[i] != wantRows[i] {
			t.Fatalf("Rows[%d] = %q, want %q; full=%v", i, rows[i], wantRows[i], rows)
		}
	}
}

func TestBuilder_Items(t *testing.T) {
	vnode := NewBuilder().
		Items([]RowItem{Item("Orders").WithDescription("3 pending")}).
		AddItem(Item("Invoices").WithSuffix("(new)")).
		BuildVNode()

	if len(vnode.Items()) != 2 {
		t.Fatalf("Items len = %d, want 2", len(vnode.Items()))
	}
	if rows := vnode.Rows(); rows[0] != "Orders - 3 pending" || rows[1] != "Invoices (new)" {
		t.Fatalf("Rows = %v, want [Orders - 3 pending Invoices (new)]", rows)
	}
}

func TestBuilder_SortAscendingRows(t *testing.T) {
	vnode := NewBuilder().
		Rows([]string{"zeta", "Alpha", "beta"}).
		SortAscending().
		BuildVNode()

	wantRows := []string{"Alpha", "beta", "zeta"}
	if !equalStrings(vnode.Rows(), wantRows) {
		t.Fatalf("Rows = %v, want %v", vnode.Rows(), wantRows)
	}
	props := vnode.Props()
	if props[propSortRows] != true || props[propSortDescending] != false {
		t.Fatalf("sort props = (%v,%v), want (true,false)", props[propSortRows], props[propSortDescending])
	}
}

func TestBuilder_SortDescendingItemsByTitle(t *testing.T) {
	vnode := NewBuilder().
		Items([]RowItem{
			Item("openai").WithPrefix("[ok]"),
			Item("azure").WithPrefix("[warn]"),
			Item("anthropic").WithPrefix("[ok]"),
		}).
		SortDescending().
		BuildVNode()

	wantRows := []string{
		"[ok] openai",
		"[warn] azure",
		"[ok] anthropic",
	}
	if !equalStrings(vnode.Rows(), wantRows) {
		t.Fatalf("Rows = %v, want %v", vnode.Rows(), wantRows)
	}
}

func TestInstance_SortRowsWorksWithSearchAndSelection(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"rows":             []string{"zeta", "alpha", "beta"},
		"sortRows":         true,
		"sortDescending":   false,
		"searchQuery":      "a",
		"showBorder":       false,
		"showSeparator":    false,
		"viewportHeight":   3,
		"showSearchStats":  true,
		"searchStatsStyle": style.Style{FG: style.Green},
	})

	wantRows := []string{"alpha", "beta", "zeta"}
	if !equalStrings(inst.GetRows(), wantRows) {
		t.Fatalf("rows = %v, want %v", inst.GetRows(), wantRows)
	}
	if inst.GetSelectedIndex() != 0 {
		t.Fatalf("selectedIndex = %d, want first sorted match", inst.GetSelectedIndex())
	}
	cmds := inst.Paint(0, 0)
	if len(cmds) < 4 {
		t.Fatalf("cmd count = %d, want >= 4", len(cmds))
	}
	if cmds[1].Text != "alpha" || cmds[2].Text != "beta" || cmds[3].Text != "zeta" {
		t.Fatalf("rendered sorted rows = %q, %q, %q", cmds[1].Text, cmds[2].Text, cmds[3].Text)
	}
}

func TestVNode_ToVirtualList_MapsRowsAndSelection(t *testing.T) {
	vnode := New().
		SetRows([]string{"alpha", "beta", "betamax", "gamma"}).
		SetSearchQuery("beta").
		SetViewportHeight(2).
		SetInitialScrollOffset(0).
		SetInitialSelectedIndex(2)

	virtual := vnode.ToVirtualList()
	if virtual == nil {
		t.Fatal("Expected ToVirtualList to return a vnode")
	}

	items := virtual.Items()
	wantItems := []string{"beta", "betamax"}
	if len(items) != len(wantItems) {
		t.Fatalf("virtual items len = %d, want %d (%v)", len(items), len(wantItems), items)
	}
	for i := range wantItems {
		if items[i] != wantItems[i] {
			t.Fatalf("virtual items[%d] = %q, want %q", i, items[i], wantItems[i])
		}
	}
	if virtual.SelectedIndex() != 1 {
		t.Fatalf("virtual selected index = %d, want 1", virtual.SelectedIndex())
	}
	if virtual.VisibleCount() != 2 {
		t.Fatalf("virtual visible count = %d, want 2", virtual.VisibleCount())
	}
}

func TestBuilder_BuildVirtualList(t *testing.T) {
	virtual := NewBuilder().
		Items([]RowItem{
			Item("Orders").WithDescription("3 pending"),
			Item("Invoices").WithSuffix("(new)"),
		}).
		ViewportHeight(3).
		BuildVirtualList()

	if virtual == nil {
		t.Fatal("Expected BuildVirtualList to return a vnode")
	}
	if _, ok := interface{}(virtual).(*virtuallist.VNode); !ok {
		t.Fatalf("BuildVirtualList returned %T, want *virtuallist.VNode", virtual)
	}
	items := virtual.Items()
	if len(items) != 2 || items[0] != "Orders - 3 pending" || items[1] != "Invoices (new)" {
		t.Fatalf("virtual items = %v, want [Orders - 3 pending Invoices (new)]", items)
	}
}

func TestBuilder_BuildVirtualBridge(t *testing.T) {
	bridge := NewBuilder().
		Rows([]string{"alpha", "beta", "betamax"}).
		SearchQuery("beta").
		BuildVirtualBridge()

	if bridge == nil || bridge.VNode == nil {
		t.Fatal("Expected BuildVirtualBridge to return a bridge with vnode")
	}
	if len(bridge.SourceIndices) != 2 || bridge.SourceIndices[0] != 1 || bridge.SourceIndices[1] != 2 {
		t.Fatalf("SourceIndices = %v, want [1 2]", bridge.SourceIndices)
	}
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

func TestInstance_Items_RenderAsFlattenedRows(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"items": []RowItem{
			Item("Orders").WithPrefix("[*]").WithDescription("3 pending"),
			Item("Invoices").WithSuffix("(new)"),
		},
		"showBorder":    false,
		"showSeparator": false,
	})

	rows := inst.GetRows()
	wantRows := []string{"[*] Orders - 3 pending", "Invoices (new)"}
	for i := range wantRows {
		if rows[i] != wantRows[i] {
			t.Fatalf("rows[%d] = %q, want %q; full=%v", i, rows[i], wantRows[i], rows)
		}
	}

	cmds := inst.Paint(0, 0)
	if len(cmds) < 2 {
		t.Fatalf("cmd count = %d, want >= 2", len(cmds))
	}
	if cmds[0].Text != wantRows[0] {
		t.Fatalf("first rendered row = %q, want %q", cmds[0].Text, wantRows[0])
	}
	if cmds[1].Text != wantRows[1] {
		t.Fatalf("second rendered row = %q, want %q", cmds[1].Text, wantRows[1])
	}
}

func TestInstance_Items_SelectIntentUsesFlattenedText(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"items": []RowItem{
			Item("Orders").WithPrefix("[*]").WithDescription("3 pending"),
			Item("Invoices").WithSuffix("(new)"),
		},
	})

	if !inst.HandleAction(action.NewAction(action.ActionNavigateDown)) {
		t.Fatal("navigate down should select first item")
	}

	row, ok := inst.GetSelectedRow()
	if !ok {
		t.Fatal("expected a selected row")
	}
	if row != "[*] Orders - 3 pending" {
		t.Fatalf("selected row = %q, want %q", row, "[*] Orders - 3 pending")
	}
}

func TestInstance_ToVirtualList_UsesRuntimeScrollAndSelection(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"rows":           []string{"alpha", "beta", "betamax", "gamma"},
		"searchQuery":    "beta",
		"viewportHeight": 1,
		"selectedIndex":  2,
		"scrollOffset":   1,
	})

	virtual := inst.ToVirtualList()
	if virtual == nil {
		t.Fatal("Expected instance bridge to return a vnode")
	}

	items := virtual.Items()
	wantItems := []string{"beta", "betamax"}
	if len(items) != len(wantItems) {
		t.Fatalf("virtual items len = %d, want %d (%v)", len(items), len(wantItems), items)
	}
	for i := range wantItems {
		if items[i] != wantItems[i] {
			t.Fatalf("virtual items[%d] = %q, want %q", i, items[i], wantItems[i])
		}
	}
	if virtual.ScrollOffset() != 1 {
		t.Fatalf("virtual scroll offset = %d, want 1", virtual.ScrollOffset())
	}
	if virtual.SelectedIndex() != 1 {
		t.Fatalf("virtual selected index = %d, want 1", virtual.SelectedIndex())
	}
}

func TestInstance_ToVirtualBridge_PreservesMatchStyleThroughItemStyleFn(t *testing.T) {
	matchStyle := style.Style{FG: style.Cyan}
	inst := NewInstance(rtui.Props{
		"rows":           []string{"alpha", "beta", "betamax"},
		"searchQuery":    "beta",
		"matchStyle":     matchStyle,
		"viewportHeight": 2,
	})

	bridge := inst.ToVirtualBridge()
	if bridge == nil || bridge.VNode == nil {
		t.Fatal("Expected ToVirtualBridge to return a bridge")
	}

	virtualInst := bridge.VNode.CreateInstance().(*virtuallist.Instance)
	cmds := virtualInst.Paint(0, 0)
	if len(cmds) < 3 {
		t.Fatalf("cmd count = %d, want >= 3", len(cmds))
	}
	if cmds[1].Style.FG != matchStyle.FG {
		t.Fatalf("first virtual row fg = %q, want %q", cmds[1].Style.FG, matchStyle.FG)
	}
}

func TestVirtualBridge_SyncToList_MapsVirtualSelectionBackToSourceRows(t *testing.T) {
	inst := NewInstance(rtui.Props{
		"rows":           []string{"alpha", "beta", "betamax", "gamma"},
		"searchQuery":    "beta",
		"viewportHeight": 1,
	})

	bridge := inst.ToVirtualBridge()
	if !bridge.SyncToList(inst, 1, 1) {
		t.Fatal("Expected SyncToList to report changes")
	}
	if inst.GetScrollOffset() != 1 {
		t.Fatalf("scrollOffset = %d, want 1", inst.GetScrollOffset())
	}
	if inst.GetSelectedIndex() != 2 {
		t.Fatalf("selectedIndex = %d, want 2", inst.GetSelectedIndex())
	}
	row, ok := inst.GetSelectedRow()
	if !ok || row != "betamax" {
		t.Fatalf("selected row = (%q,%v), want (betamax,true)", row, ok)
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

func TestInstance_HandleAction_ClickEmitsLocalRowSelectIntent(t *testing.T) {
	parent := &bubbleCaptureParent{}
	inst := NewInstance(rtui.Props{
		"componentID":    "orders.list",
		"header":         "Items",
		"rows":           []string{"first", "second", "third", "fourth"},
		"viewportHeight": 3,
		"showBorder":     true,
		"showSeparator":  true,
		"scrollOffset":   1,
	})
	inst.SetParent(parent)
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
	if len(parent.intents) != 1 {
		t.Fatalf("captured intents = %d, want 1", len(parent.intents))
	}
	rowSelect, ok := parent.intents[0].(RowSelectIntent)
	if !ok {
		t.Fatalf("captured intent = %T, want RowSelectIntent", parent.intents[0])
	}
	if rowSelect.ComponentID != "orders.list" || rowSelect.SelectedIndex != 2 || rowSelect.SelectedRow != "third" {
		t.Fatalf("row select = %#v, want componentID orders.list index 2 row third", rowSelect)
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

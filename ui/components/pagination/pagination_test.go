package pagination

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

func TestNew(t *testing.T) {
	v := New()
	if v == nil {
		t.Fatal("New returned nil")
	}
	if v.Tag() != "pagination" {
		t.Fatalf("Tag = %q, want pagination", v.Tag())
	}
	if v.pageSize != 10 {
		t.Fatalf("pageSize = %d, want 10", v.pageSize)
	}
}

func TestBuilderFluent(t *testing.T) {
	v := NewBuilder().
		Key("pager").
		ComponentID("orders.pagination").
		Total(120).
		PageSize(20).
		CurrentPage(2).
		MaxButtons(7).
		ShowTotal(false).
		Disabled(true).
		SelectedStyle(style.Style{}.Bold(true)).
		Build()

	if v.Key() != "pager" {
		t.Fatalf("Key = %q, want pager", v.Key())
	}
	if v.componentID != "orders.pagination" || v.total != 120 || v.pageSize != 20 {
		t.Fatalf("unexpected builder values: %+v", v)
	}
	if v.currentPage != 2 || !v.currentPageControlled {
		t.Fatalf("current page state = (%d,%v), want (2,true)", v.currentPage, v.currentPageControlled)
	}
	if v.maxButtons != 7 || v.showTotal || !v.disabled {
		t.Fatalf("unexpected display flags: maxButtons=%d showTotal=%v disabled=%v", v.maxButtons, v.showTotal, v.disabled)
	}
}

func TestInstanceMeasureAndPaint(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTotal:       120,
		propPageSize:    10,
		propCurrentPage: 4,
		propMaxButtons:  5,
	})
	size := inst.Measure(layout.Constraints{MaxWidth: 80, MaxHeight: 1})
	if size.Width <= 0 || size.Height != 1 {
		t.Fatalf("size = %+v, want positive width and height 1", size)
	}
	inst.SetBounds(0, 0, 80, 1)
	cmds := inst.Paint(0, 0)
	text := collectTexts(cmds)
	if text == "" {
		t.Fatal("expected non-empty paint output")
	}
	if text != "Prev 1 … 3 4 [5] 6 7 … 12 Next 120 items" {
		t.Fatalf("paint text = %q", text)
	}
}

func TestInstanceClampsCurrentPage(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTotal:       15,
		propPageSize:    10,
		propCurrentPage: 99,
	})
	if inst.GetCurrentPage() != 1 {
		t.Fatalf("currentPage = %d, want 1", inst.GetCurrentPage())
	}
	if inst.GetPageCount() != 2 {
		t.Fatalf("pageCount = %d, want 2", inst.GetPageCount())
	}
}

func TestInstanceHandleActionEmitsIntents(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID:     "orders.pagination",
		propTotal:           120,
		propPageSize:        10,
		propCurrentPage:     0,
		propPageIntentField: intent.BindField("currentPage"),
	})
	var emitted []intent.Intent
	inst.SetIntentEmitter(func(i intent.Intent) {
		emitted = append(emitted, i)
	})

	if !inst.HandleAction(action.NewAction(action.ActionNavigateRight)) {
		t.Fatal("expected navigate right to change page")
	}
	if inst.GetCurrentPage() != 1 {
		t.Fatalf("currentPage = %d, want 1", inst.GetCurrentPage())
	}
	if len(emitted) != 2 {
		t.Fatalf("emitted len = %d, want 2", len(emitted))
	}
	pageChange, ok := emitted[0].(PageChangeIntent)
	if !ok {
		t.Fatalf("first emitted intent = %T, want PageChangeIntent", emitted[0])
	}
	if pageChange.FromPage != 0 || pageChange.ToPage != 1 || pageChange.PageCount != 12 {
		t.Fatalf("unexpected PageChangeIntent: %+v", pageChange)
	}
	fieldChange, ok := emitted[1].(intent.FieldChangeIntent)
	if !ok {
		t.Fatalf("second emitted intent = %T, want intent.FieldChangeIntent", emitted[1])
	}
	if fieldChange.Field != "currentPage" || fieldChange.Value != "1" {
		t.Fatalf("unexpected field change intent: %+v", fieldChange)
	}
}

func TestInstanceHandleClickSelectsPage(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTotal:       120,
		propPageSize:    10,
		propCurrentPage: 4,
		propMaxButtons:  5,
	})
	if !inst.handleClick(9) {
		t.Fatal("expected click on page token to change page")
	}
	if inst.GetCurrentPage() != 2 {
		t.Fatalf("currentPage = %d, want 2", inst.GetCurrentPage())
	}
}

func TestInstanceHandleMouseActionClick(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propTotal:       30,
		propPageSize:    10,
		propCurrentPage: 0,
	})
	mouse := runtimemsg.NewMouseMsgWithTarget(9, 0, 9, 0, 0, runtimemsg.MouseLeft, runtimemsg.MouseActionPress)
	if !inst.HandleAction(action.NewAction(action.ActionClick).WithPayload(mouse)) {
		t.Fatal("expected click action to be handled")
	}
	if inst.GetCurrentPage() != 1 {
		t.Fatalf("currentPage = %d, want 1", inst.GetCurrentPage())
	}
}

func TestInstanceHandleIntent(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID: "orders.pagination",
		propTotal:       100,
		propPageSize:    10,
	})
	if !inst.HandleIntent(PageChangeWithID("orders.pagination", 0, 3, 10, 10, 100)) {
		t.Fatal("expected PageChangeIntent to be handled")
	}
	if inst.GetCurrentPage() != 3 {
		t.Fatalf("currentPage = %d, want 3", inst.GetCurrentPage())
	}
	if inst.HandleIntent(PageChangeWithID("other.pagination", 0, 4, 10, 10, 100)) {
		t.Fatal("expected intent for another component to be ignored")
	}
}

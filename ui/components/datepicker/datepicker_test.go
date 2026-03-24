package datepicker

import (
	"testing"
	"time"

	runtimeintent "github.com/wwsheng009/mint/runtime/intent"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

func TestNew(t *testing.T) {
	v := New()
	if v == nil {
		t.Fatal("New returned nil")
	}
	if v.Tag() != "datepicker" {
		t.Fatalf("Tag = %q, want datepicker", v.Tag())
	}
	if v.placeholder != defaultPlaceholder {
		t.Fatalf("placeholder = %q, want %q", v.placeholder, defaultPlaceholder)
	}
}

func TestBuilderFluent(t *testing.T) {
	v := NewBuilder().
		Key("ship-date").
		SetID("ship-date-input").
		ComponentID("form.ship_date").
		Width(20).
		Placeholder("Select date").
		Value("2026-04-05").
		Disabled(true).
		BuildVNode()

	if v.Key() != "ship-date" || v.ID() != "ship-date-input" {
		t.Fatalf("key/id = (%q,%q)", v.Key(), v.ID())
	}
	if v.componentID != "form.ship_date" || v.width != 20 {
		t.Fatalf("componentID/width = (%q,%d)", v.componentID, v.width)
	}
	if v.value != "2026-04-05" || !v.valueControlled || !v.disabled {
		t.Fatalf("valueControlled/disabled not set")
	}
}

func TestAppendDraftCommitsValidDateAndEmitsFieldChange(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID: "fixture.datepicker",
		propChangeIntent: runtimeintent.ForField(
			runtimeintent.StateKey[string]("filters.ship_date"),
		),
	})

	var emitted []runtimeintent.Intent
	inst.SetIntentEmitter(func(i runtimeintent.Intent) {
		emitted = append(emitted, i)
	})

	if !inst.appendDraft("2026-04-05") {
		t.Fatal("appendDraft should change state")
	}
	if !inst.hasValue || inst.selectedValue() != "2026-04-05" {
		t.Fatalf("selectedValue = %q", inst.selectedValue())
	}
	if len(emitted) != 2 {
		t.Fatalf("emitted len = %d, want 2", len(emitted))
	}
	fieldChange, ok := emitted[0].(runtimeintent.FieldChangeIntent)
	if !ok {
		t.Fatalf("first emitted = %T, want FieldChangeIntent", emitted[0])
	}
	if fieldChange.Field != "filters.ship_date" || fieldChange.Value != "2026-04-05" {
		t.Fatalf("field change = %+v", fieldChange)
	}
	dateChange, ok := emitted[1].(DateChangeIntent)
	if !ok {
		t.Fatalf("second emitted = %T, want DateChangeIntent", emitted[1])
	}
	if dateChange.Value != "2026-04-05" || dateChange.Day != 5 {
		t.Fatalf("date change = %+v", dateChange)
	}
}

func TestNavigateMonthClampsDay(t *testing.T) {
	inst := NewInstance(rtui.Props{})
	inst.commitDate(time.Date(2026, time.January, 31, 0, 0, 0, 0, time.UTC), false, false)
	inst.openPicker()

	if !inst.navigateMonth(1) {
		t.Fatal("navigateMonth should change highlight")
	}
	if !sameDate(inst.highlightedDate, time.Date(2026, time.February, 28, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("highlightedDate = %s, want 2026-02-28", inst.highlightedDate.Format(dateLayout))
	}
}

func TestRuntimeChildrenOpenAddsPopupPortal(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propKey:      "fixture-date",
		propPickerID: "fixture-datepicker",
	})
	inst.openPicker()

	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	if children[0].ID() != "fixture-datepicker-popup-portal" {
		t.Fatalf("portal ID = %q", children[0].ID())
	}
}

func TestCalendarGridUsesMondayStart(t *testing.T) {
	grid := calendarGrid(time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC))
	if len(grid) != 42 {
		t.Fatalf("grid len = %d, want 42", len(grid))
	}
	if !sameDate(grid[0], time.Date(2026, time.February, 23, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("grid[0] = %s, want 2026-02-23", grid[0].Format(dateLayout))
	}
	if !sameDate(grid[41], time.Date(2026, time.April, 5, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("grid[41] = %s, want 2026-04-05", grid[41].Format(dateLayout))
	}
}

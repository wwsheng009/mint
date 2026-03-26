package timepicker

import (
	"testing"
	"time"

	runtimeintent "github.com/wwsheng009/mint/runtime/intent"
	rttypes "github.com/wwsheng009/mint/runtime/types"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

func TestNew(t *testing.T) {
	v := New()
	if v == nil {
		t.Fatal("New returned nil")
	}
	if v.Tag() != "timepicker" {
		t.Fatalf("Tag = %q, want timepicker", v.Tag())
	}
	if v.placeholder != defaultPlaceholder {
		t.Fatalf("placeholder = %q, want %q", v.placeholder, defaultPlaceholder)
	}
}

func TestBuilderFluent(t *testing.T) {
	v := NewBuilder().
		Key("ship-time").
		SetID("ship-time-input").
		ComponentID("form.ship_time").
		Width(12).
		Placeholder("Select time").
		Value("09:30").
		Disabled(true).
		BuildVNode()

	if v.Key() != "ship-time" || v.ID() != "ship-time-input" {
		t.Fatalf("key/id = (%q,%q)", v.Key(), v.ID())
	}
	if v.componentID != "form.ship_time" || v.width != 12 {
		t.Fatalf("componentID/width = (%q,%d)", v.componentID, v.width)
	}
	if v.value != "09:30" || !v.valueControlled || !v.disabled {
		t.Fatalf("valueControlled/disabled not set")
	}
}

func TestAppendDraftCommitsValidTimeAndEmitsFieldChange(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID: "fixture.timepicker",
		propChangeIntent: runtimeintent.ForField(
			runtimeintent.StateKey[string]("filters.ship_time"),
		),
	})

	var emitted []runtimeintent.Intent
	inst.SetIntentEmitter(func(i runtimeintent.Intent) {
		emitted = append(emitted, i)
	})

	if !inst.appendDraft("09:30") {
		t.Fatal("appendDraft should change state")
	}
	if !inst.hasValue || inst.selectedValue() != "09:30" {
		t.Fatalf("selectedValue = %q", inst.selectedValue())
	}
	if len(emitted) != 2 {
		t.Fatalf("emitted len = %d, want 2", len(emitted))
	}
	fieldChange, ok := emitted[0].(runtimeintent.FieldChangeIntent)
	if !ok {
		t.Fatalf("first emitted = %T, want FieldChangeIntent", emitted[0])
	}
	if fieldChange.Field != "filters.ship_time" || fieldChange.Value != "09:30" {
		t.Fatalf("field change = %+v", fieldChange)
	}
	timeChange, ok := emitted[1].(TimeChangeIntent)
	if !ok {
		t.Fatalf("second emitted = %T, want TimeChangeIntent", emitted[1])
	}
	if timeChange.Value != "09:30" || timeChange.Hour != 9 || timeChange.Minute != 30 {
		t.Fatalf("time change = %+v", timeChange)
	}
}

func TestBlurNormalizesTime(t *testing.T) {
	inst := NewInstance(rtui.Props{})

	if !inst.appendDraft("9:5") {
		t.Fatal("appendDraft should accept partial time")
	}
	inst.applyDraftOnBlur()

	if got := inst.selectedValue(); got != "09:05" {
		t.Fatalf("selectedValue = %q, want 09:05", got)
	}
}

func TestInvalidBlurRestoresLastCommittedValue(t *testing.T) {
	inst := NewInstance(rtui.Props{})
	inst.commitMinutes(9*60+30, false, false)
	inst.draft = "25:61"

	inst.applyDraftOnBlur()

	if got := inst.selectedValue(); got != "09:30" {
		t.Fatalf("selectedValue = %q, want 09:30", got)
	}
	if inst.draft != "09:30" {
		t.Fatalf("draft = %q, want 09:30", inst.draft)
	}
}

func TestRuntimeChildrenOpenAddsPopupPortal(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propKey:      "fixture-time",
		propPickerID: "fixture-timepicker",
	})
	inst.openPicker()

	children := inst.RuntimeChildren()
	if len(children) != 2 {
		t.Fatalf("RuntimeChildren len = %d, want 2", len(children))
	}
	if children[0].ID() != "fixture-timepicker-popup-anchor" {
		t.Fatalf("anchor ID = %q", children[0].ID())
	}
	if children[1].ID() != "fixture-timepicker-popup-portal" {
		t.Fatalf("portal ID = %q", children[1].ID())
	}

	props := children[1].Props()
	if got, _ := props["top"].(int); got != 0 {
		t.Fatalf("portal top = %d, want 0", got)
	}
	if got, _ := props["width"].(int); got != 1 {
		t.Fatalf("portal width = %d, want 1", got)
	}
	if got, _ := props["height"].(int); got != 1 {
		t.Fatalf("portal height = %d, want 1", got)
	}
	if got, _ := props["anchor"].(rttypes.Anchor); got != rttypes.AnchorBottomLeft {
		t.Fatalf("portal anchor = %v, want %v", got, rttypes.AnchorBottomLeft)
	}
}

func TestKeyboardNavigationAdjustsHighlightedTime(t *testing.T) {
	inst := NewInstance(rtui.Props{})
	inst.commitMinutes(9*60+30, false, false)
	inst.openPicker()

	if !inst.stepSegment(1) {
		t.Fatal("stepSegment should change hour when active segment is hour")
	}
	if got := formatTimeValue(inst.highlightedMinutes); got != "10:30" {
		t.Fatalf("highlighted = %q, want 10:30", got)
	}
	if !inst.switchSegment(1) {
		t.Fatal("switchSegment should move to minute segment")
	}
	if !inst.stepSegment(-1) {
		t.Fatal("stepSegment should change minute when active segment is minute")
	}
	if got := formatTimeValue(inst.highlightedMinutes); got != "10:29" {
		t.Fatalf("highlighted = %q, want 10:29", got)
	}
}

func TestCommitHighlightedTime(t *testing.T) {
	inst := NewInstance(rtui.Props{})
	inst.openPicker()
	inst.highlightedMinutes = 14*60 + 45

	if !inst.commitHighlighted() {
		t.Fatal("commitHighlighted should change state")
	}
	if got := inst.selectedValue(); got != "14:45" {
		t.Fatalf("selectedValue = %q, want 14:45", got)
	}
	if inst.open {
		t.Fatal("popup should close after commit")
	}
}

func TestControlledValueSync(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propValueControlled: true,
		propValue:           "08:15",
	})

	inst.SetProps(rtui.Props{
		propValueControlled: true,
		propValue:           "14:20",
	})

	if got := inst.selectedValue(); got != "14:20" {
		t.Fatalf("selectedValue = %q, want 14:20", got)
	}
}

func TestNormalizeBlurTimeValue(t *testing.T) {
	tests := []struct {
		raw  string
		want string
		ok   bool
	}{
		{raw: "9:5", want: "09:05", ok: true},
		{raw: "09:05", want: "09:05", ok: true},
		{raw: "24:00", ok: false},
		{raw: "12:60", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			got, ok := normalizeBlurTimeValue(tt.raw)
			if ok != tt.ok {
				t.Fatalf("ok = %v, want %v", ok, tt.ok)
			}
			if got != tt.want {
				t.Fatalf("got = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCurrentTimeMinutesUsesClock(t *testing.T) {
	prev := nowFunc
	nowFunc = func() time.Time {
		return time.Date(2026, time.March, 18, 21, 7, 0, 0, time.UTC)
	}
	defer func() { nowFunc = prev }()

	if got := currentTimeMinutes(); got != 21*60+7 {
		t.Fatalf("currentTimeMinutes = %d, want %d", got, 21*60+7)
	}
}

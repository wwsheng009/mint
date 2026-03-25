package input

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/action"
	runtimeintent "github.com/wwsheng009/mint/runtime/intent"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/form"
)

func TestVNode_Builder_NumberMode(t *testing.T) {
	vnode := NewBuilder().
		Type(TypeNumber).
		AllowNegative(false).
		AllowDecimal(false).
		Min(1).
		Max(10).
		Step(2).
		BuildTyped()

	if vnode.InputType() != TypeNumber {
		t.Fatalf("InputType = %v, want %v", vnode.InputType(), TypeNumber)
	}
	if vnode.AllowNegative() {
		t.Fatal("AllowNegative should be false")
	}
	if vnode.AllowDecimal() {
		t.Fatal("AllowDecimal should be false")
	}
	if !vnode.HasMin() || vnode.Min() != 1 {
		t.Fatalf("Min = (%t, %v), want (true, 1)", vnode.HasMin(), vnode.Min())
	}
	if !vnode.HasMax() || vnode.Max() != 10 {
		t.Fatalf("Max = (%t, %v), want (true, 10)", vnode.HasMax(), vnode.Max())
	}
	if vnode.Step() != 2 {
		t.Fatalf("Step = %v, want 2", vnode.Step())
	}

	props := vnode.Props()
	if allowNegative, _ := props[propAllowNegative].(bool); allowNegative {
		t.Fatal("propAllowNegative should be false")
	}
	if allowDecimal, _ := props[propAllowDecimal].(bool); allowDecimal {
		t.Fatal("propAllowDecimal should be false")
	}
	if hasMin, _ := props[propHasMin].(bool); !hasMin {
		t.Fatal("propHasMin should be true")
	}
	if min, _ := props[propMin].(float64); min != 1 {
		t.Fatalf("propMin = %v, want 1", min)
	}
	if hasMax, _ := props[propHasMax].(bool); !hasMax {
		t.Fatal("propHasMax should be true")
	}
	if max, _ := props[propMax].(float64); max != 10 {
		t.Fatalf("propMax = %v, want 10", max)
	}
	if step, _ := props[propStep].(float64); step != 2 {
		t.Fatalf("propStep = %v, want 2", step)
	}
}

func TestInstance_InsertText_NumberInputFiltersInvalidCharacters(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propInputType: TypeNumber,
	})

	if !inst.InsertText("-12.3a4..5") {
		t.Fatal("InsertText should accept the valid numeric subset")
	}

	if got := inst.GetValue(); got != "-12.345" {
		t.Fatalf("Value = %q, want %q", got, "-12.345")
	}
	if got := inst.CursorPos(); got != len([]rune("-12.345")) {
		t.Fatalf("CursorPos = %d, want %d", got, len([]rune("-12.345")))
	}
}

func TestInstance_InsertText_NumberInputRejectsInvalidCursorInsertion(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propInputType: TypeNumber,
		propValue:     "12.3",
	})

	inst.SetCursorPos(2)
	if inst.InsertText("-") {
		t.Fatal("InsertText should reject '-' away from the leading position")
	}
	if got := inst.GetValue(); got != "12.3" {
		t.Fatalf("Value = %q, want %q", got, "12.3")
	}
	if got := inst.CursorPos(); got != 2 {
		t.Fatalf("CursorPos = %d, want 2", got)
	}

	inst.SetCursorPos(4)
	if inst.InsertText(".") {
		t.Fatal("InsertText should reject a second decimal point")
	}
	if got := inst.CursorPos(); got != 4 {
		t.Fatalf("CursorPos = %d, want 4", got)
	}
}

func TestInstance_InsertText_NumberInputHonorsNumberModeOptions(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propInputType:     TypeNumber,
		propAllowNegative: false,
		propAllowDecimal:  false,
	})

	if !inst.InsertText("-12.3a4") {
		t.Fatal("InsertText should keep the valid digits")
	}
	if got := inst.GetValue(); got != "1234" {
		t.Fatalf("Value = %q, want %q", got, "1234")
	}
}

func TestInstance_InsertText_NumberInputFiltersBeforeMaxLen(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propInputType: TypeNumber,
		propMaxLen:    3,
	})

	if !inst.InsertText("1a2") {
		t.Fatal("InsertText should allow filtered content within maxLen")
	}
	if got := inst.GetValue(); got != "12" {
		t.Fatalf("Value = %q, want %q", got, "12")
	}

	if !inst.InsertText("b3") {
		t.Fatal("InsertText should allow the remaining valid rune after filtering")
	}
	if got := inst.GetValue(); got != "123" {
		t.Fatalf("Value = %q, want %q", got, "123")
	}
}

func TestInstance_NumberInputBlurNormalizesValue(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "trim leading zeros", raw: "0012", want: "12"},
		{name: "prefix decimal with zero", raw: ".5", want: "0.5"},
		{name: "trim trailing decimal point", raw: "1.", want: "1"},
		{name: "normalize negative decimal", raw: "-.5", want: "-0.5"},
		{name: "normalize negative zero", raw: "-0", want: "0"},
		{name: "clear dangling minus", raw: "-", want: ""},
		{name: "clear dangling decimal", raw: "-.", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := NewInstance(rtui.Props{
				propInputType: TypeNumber,
				propValue:     tt.raw,
			})

			inst.SetFocus(true)
			inst.SetFocus(false)

			if got := inst.GetValue(); got != tt.want {
				t.Fatalf("Value = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInstance_NumberInputBlurClampsMinMax(t *testing.T) {
	tests := []struct {
		name  string
		props rtui.Props
		want  string
	}{
		{
			name: "clamp to max",
			props: rtui.Props{
				propInputType: TypeNumber,
				propValue:     "15",
				propHasMax:    true,
				propMax:       10.0,
			},
			want: "10",
		},
		{
			name: "clamp to min",
			props: rtui.Props{
				propInputType: TypeNumber,
				propValue:     "-5",
				propHasMin:    true,
				propMin:       -2.0,
			},
			want: "-2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := NewInstance(tt.props)
			inst.SetFocus(true)
			inst.SetFocus(false)

			if got := inst.GetValue(); got != tt.want {
				t.Fatalf("Value = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInstance_NumberInputBlurEmitsNormalizedFieldChange(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propInputType:    TypeNumber,
		propValue:        "001.20",
		propChangeIntent: runtimeintent.BindField("amount"),
	})

	var emitted []runtimeintent.Intent
	inst.SetIntentEmitter(func(i runtimeintent.Intent) {
		emitted = append(emitted, i)
	})

	inst.SetFocus(true)
	inst.SetFocus(false)

	if got := inst.GetValue(); got != "1.20" {
		t.Fatalf("Value = %q, want %q", got, "1.20")
	}
	if len(emitted) != 1 {
		t.Fatalf("emitted %d intents, want 1", len(emitted))
	}

	changeIntent, ok := emitted[0].(runtimeintent.FieldChangeIntent)
	if !ok {
		t.Fatalf("intent type = %T, want runtimeintent.FieldChangeIntent", emitted[0])
	}
	if changeIntent.Field != "amount" {
		t.Fatalf("Field = %q, want %q", changeIntent.Field, "amount")
	}
	if changeIntent.Value != "1.20" {
		t.Fatalf("Value = %q, want %q", changeIntent.Value, "1.20")
	}
}

func TestInstance_NumberInputBlurEmitsNormalizedFormIntents(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propInputType:    TypeNumber,
		propValue:        "-.5",
		propChangeIntent: runtimeintent.BindField("amount"),
		propFormID:       "paymentForm",
	})

	var emitted []runtimeintent.Intent
	runtimeintent.SetBubbleTestHook(func(component interface{}, i runtimeintent.Intent) bool {
		emitted = append(emitted, i)
		return true
	})
	defer runtimeintent.SetBubbleTestHook(nil)
	inst.SetIntentEmitter(func(runtimeintent.Intent) {})

	inst.SetFocus(true)
	inst.SetFocus(false)

	if got := inst.GetValue(); got != "-0.5" {
		t.Fatalf("Value = %q, want %q", got, "-0.5")
	}
	if len(emitted) != 2 {
		t.Fatalf("emitted %d intents, want 2", len(emitted))
	}

	changeIntent, ok := emitted[0].(form.FormFieldChangeIntent)
	if !ok {
		t.Fatalf("first intent type = %T, want form.FormFieldChangeIntent", emitted[0])
	}
	if changeIntent.FormID != "paymentForm" || changeIntent.Field != "amount" || changeIntent.Value != "-0.5" {
		t.Fatalf("unexpected change intent: %+v", changeIntent)
	}

	blurIntent, ok := emitted[1].(form.FormFieldBlurIntent)
	if !ok {
		t.Fatalf("second intent type = %T, want form.FormFieldBlurIntent", emitted[1])
	}
	if blurIntent.FormID != "paymentForm" || blurIntent.Field != "amount" || blurIntent.Value != "-0.5" {
		t.Fatalf("unexpected blur intent: %+v", blurIntent)
	}
}

func TestInstance_NumberInputCursorStepActions(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propInputType:    TypeNumber,
		propValue:        "1.5",
		propStep:         0.25,
		propHasMax:       true,
		propMax:          2.0,
		propChangeIntent: runtimeintent.BindField("amount"),
	})

	var emitted []runtimeintent.Intent
	inst.SetIntentEmitter(func(i runtimeintent.Intent) {
		emitted = append(emitted, i)
	})

	if !inst.HandleAction(action.NewAction(action.ActionCursorUp)) {
		t.Fatal("ActionCursorUp should be handled for number input")
	}
	if got := inst.GetValue(); got != "1.75" {
		t.Fatalf("Value after up = %q, want %q", got, "1.75")
	}

	if !inst.HandleAction(action.NewAction(action.ActionCursorDown)) {
		t.Fatal("ActionCursorDown should be handled for number input")
	}
	if got := inst.GetValue(); got != "1.5" {
		t.Fatalf("Value after down = %q, want %q", got, "1.5")
	}

	inst.SetValue("2")
	if inst.HandleAction(action.NewAction(action.ActionCursorUp)) {
		t.Fatal("ActionCursorUp at max should not report a change")
	}
	if got := inst.GetValue(); got != "2" {
		t.Fatalf("Value at max = %q, want %q", got, "2")
	}

	if len(emitted) != 2 {
		t.Fatalf("emitted %d intents, want 2", len(emitted))
	}
}

func TestInstance_NumberInputCursorStepHonorsNonNegative(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propInputType:     TypeNumber,
		propValue:         "0",
		propAllowNegative: false,
		propStep:          1,
	})

	if inst.HandleAction(action.NewAction(action.ActionCursorDown)) {
		t.Fatal("ActionCursorDown should not move below zero when negatives are disabled")
	}
	if got := inst.GetValue(); got != "0" {
		t.Fatalf("Value = %q, want %q", got, "0")
	}
}

func TestInstance_NumberInputVerticalCursorMethods(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propInputType: TypeNumber,
		propValue:     "5",
		propStep:      5.0,
		propHasMax:    true,
		propMax:       10.0,
	})

	if !inst.MoveCursorUp() {
		t.Fatal("MoveCursorUp should step number input upward")
	}
	if got := inst.GetValue(); got != "10" {
		t.Fatalf("Value after MoveCursorUp = %q, want %q", got, "10")
	}

	if !inst.MoveCursorDown() {
		t.Fatal("MoveCursorDown should step number input downward")
	}
	if got := inst.GetValue(); got != "5" {
		t.Fatalf("Value after MoveCursorDown = %q, want %q", got, "5")
	}
}

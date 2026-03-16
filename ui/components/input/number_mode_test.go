package input

import (
	"testing"

	runtimeintent "github.com/wwsheng009/mint/runtime/intent"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/form"
)

func TestVNode_Builder_NumberMode(t *testing.T) {
	vnode := NewBuilder().
		Type(TypeNumber).
		AllowNegative(false).
		AllowDecimal(false).
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

	props := vnode.Props()
	if allowNegative, _ := props[propAllowNegative].(bool); allowNegative {
		t.Fatal("propAllowNegative should be false")
	}
	if allowDecimal, _ := props[propAllowDecimal].(bool); allowDecimal {
		t.Fatal("propAllowDecimal should be false")
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

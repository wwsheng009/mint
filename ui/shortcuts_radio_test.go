package ui

import "testing"

func TestRadioShortcuts(t *testing.T) {
	single := NewRadioBuilder().
		Label("Primary").
		Checked(true).
		Build()
	if single == nil {
		t.Fatal("NewRadioBuilder().Build() returned nil")
	}

	short := Radio("Secondary", false)
	if short == nil {
		t.Fatal("Radio() returned nil")
	}

	group := NewRadioGroupBuilder([]RadioOption{
		NewRadioOption("a", "Option A"),
		NewRadioOption("b", "Option B"),
	}).Selected("b").Build()
	if group == nil {
		t.Fatal("NewRadioGroupBuilder().Build() returned nil")
	}
}

func TestRadioAliases(t *testing.T) {
	option := NewRadioOption("value", "Label")
	if option.Value != "value" || option.Label != "Label" {
		t.Fatalf("NewRadioOption() = %+v", option)
	}

	if RadioVertical == RadioHorizontal {
		t.Fatal("radio orientation aliases should differ")
	}
}

package ui

import "testing"

// TestNewCheckbox tests checkbox creation
func TestNewCheckbox(t *testing.T) {
	checkbox := NewCheckbox()

	if checkbox == nil {
		t.Fatal("NewCheckbox() returned nil")
	}

	if checkbox.Checked() != false {
		t.Errorf("Initial checked = %v, want false", checkbox.Checked())
	}

	if checkbox.Disabled() != false {
		t.Errorf("Initial disabled = %v, want false", checkbox.Disabled())
	}

	if checkbox.Label() != "" {
		t.Errorf("Initial label = %v, want empty", checkbox.Label())
	}
}

// TestCheckboxBuilder tests checkbox builder
func TestCheckboxBuilder(t *testing.T) {
	checkbox := CheckboxBuilder().
		Checked(true).
		Disabled(true).
		Label("Accept terms").
		Build()

	checkboxVNode, ok := checkbox.(*CheckboxVNode)
	if !ok {
		t.Fatal("CheckboxBuilder() did not return *CheckboxVNode")
	}

	if checkboxVNode.Checked() != true {
		t.Errorf("Checked = %v, want true", checkboxVNode.Checked())
	}

	if checkboxVNode.Label() != "Accept terms" {
		t.Errorf("Label = %v, want 'Accept terms'", checkboxVNode.Label())
	}

	if !checkboxVNode.Disabled() {
		t.Error("Disabled should be true")
	}
}

// TestCheckboxSetChecked tests setting checked state
func TestCheckboxSetChecked(t *testing.T) {
	checkbox := NewCheckbox()

	checkbox.SetChecked(true)
	if checkbox.Checked() != true {
		t.Errorf("Checked = %v, want true", checkbox.Checked())
	}

	checkbox.SetChecked(false)
	if checkbox.Checked() != false {
		t.Errorf("Checked = %v, want false", checkbox.Checked())
	}

	// Setting should also update the prop
	if checkbox.Props().GetBool("checked") != false {
		t.Errorf("Prop checked = %v, want false", checkbox.Props().GetBool("checked"))
	}
}

// TestCheckboxToggle tests toggle functionality
func TestCheckboxToggle(t *testing.T) {
	checkbox := NewCheckbox()

	// Toggle from false to true
	result := checkbox.Toggle()
	if result != true {
		t.Errorf("Toggle() returned %v, want true", result)
	}
	if checkbox.Checked() != true {
		t.Errorf("Checked = %v, want true after toggle", checkbox.Checked())
	}

	// Toggle from true to false
	result = checkbox.Toggle()
	if result != false {
		t.Errorf("Toggle() returned %v, want false", result)
	}
	if checkbox.Checked() != false {
		t.Errorf("Checked = %v, want false after toggle", checkbox.Checked())
	}
}

// TestCheckboxLabel tests label setting
func TestCheckboxLabel(t *testing.T) {
	checkbox := NewCheckbox()

	checkbox.SetLabel("Test Label")
	if checkbox.Label() != "Test Label" {
		t.Errorf("Label = %v, want 'Test Label'", checkbox.Label())
	}
}

// TestCheckboxDisabled tests disabled state
func TestCheckboxDisabled(t *testing.T) {
	checkbox := NewCheckbox()

	if checkbox.Disabled() != false {
		t.Error("Initial disabled should be false")
	}

	checkbox.SetDisabled(true)
	if !checkbox.Disabled() {
		t.Error("Disabled should be true after SetDisabled(true)")
	}

	checkbox.SetDisabled(false)
	if checkbox.Disabled() {
		t.Error("Disabled should be false after SetDisabled(false)")
	}
}

// TestCheckboxHandlers tests checkbox event handlers
func TestCheckboxHandlers(t *testing.T) {
	onChangeCalled := false
	onChangeValue := false

	checkbox := CheckboxBuilder().
		OnChange(func(checked bool) {
			onChangeCalled = true
			onChangeValue = checked
		}).
		Build()

	checkboxVNode, ok := checkbox.(*CheckboxVNode)
	if !ok {
		t.Fatal("Build() did not return *CheckboxVNode")
	}

	// Test onChange
	if checkboxVNode.OnChange() == nil {
		t.Error("OnChange should not be nil")
	} else {
		checkboxVNode.OnChange()(true)
		if !onChangeCalled {
			t.Error("OnChange handler was not called")
		}
		if onChangeValue != true {
			t.Errorf("OnChange received %v, want true", onChangeValue)
		}
	}
}

// TestCheckboxFocus tests checkbox focus state
func TestCheckboxFocus(t *testing.T) {
	checkbox := NewCheckbox()

	if checkbox.IsFocused() {
		t.Error("Checkbox should not be focused initially")
	}

	checkbox.SetFocus(true)
	if !checkbox.IsFocused() {
		t.Error("Checkbox should be focused after SetFocus(true)")
	}

	checkbox.SetFocus(false)
	if checkbox.IsFocused() {
		t.Error("Checkbox should not be focused after SetFocus(false)")
	}
}

// BenchmarkCheckboxBuilder benchmarks checkbox builder
func BenchmarkCheckboxBuilder(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CheckboxBuilder().
			Checked(true).
			Label("Test").
			Build()
	}
}

// BenchmarkNewCheckbox benchmarks checkbox creation
func BenchmarkNewCheckbox(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewCheckbox()
	}
}

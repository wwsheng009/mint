package ui

import (
	"testing"
)

// TestNewInput tests input creation
func TestNewInput(t *testing.T) {
	input := NewInput()

	if input == nil {
		t.Fatal("NewInput() returned nil")
	}

	if input.Value() != "" {
		t.Errorf("Initial value = %v, want empty", input.Value())
	}

	if input.Placeholder() != "" {
		t.Errorf("Initial placeholder = %v, want empty", input.Placeholder())
	}

	if input.InputType() != InputTypeText {
		t.Errorf("Initial type = %v, want InputTypeText", input.InputType())
	}
}

// TestInputBuilder tests input builder
func TestInputBuilder(t *testing.T) {
	input := InputBuilder().
		Value("test").
		Placeholder("Enter text").
		Type(InputTypePassword).
		MaxLength(10).
		Disabled(true).
		Build()

	inputVNode, ok := input.(*InputVNode)
	if !ok {
		t.Fatal("InputBuilder() did not return *InputVNode")
	}

	if inputVNode.Value() != "test" {
		t.Errorf("Value = %v, want 'test'", inputVNode.Value())
	}

	if inputVNode.Placeholder() != "Enter text" {
		t.Errorf("Placeholder = %v, want 'Enter text'", inputVNode.Value())
	}

	if inputVNode.InputType() != InputTypePassword {
		t.Errorf("Type = %v, want InputTypePassword", inputVNode.InputType())
	}

	if inputVNode.MaxLength() != 10 {
		t.Errorf("MaxLength = %v, want 10", inputVNode.MaxLength())
	}

	if !inputVNode.Disabled() {
		t.Error("Disabled should be true")
	}
}

// TestInputSetValue tests setting input value
func TestInputSetValue(t *testing.T) {
	input := NewInput()

	input.SetValue("hello")
	if input.Value() != "hello" {
		t.Errorf("Value = %v, want 'hello'", input.Value())
	}

	// Setting value should also update the prop
	if input.Props().GetString("value") != "hello" {
		t.Errorf("Prop value = %v, want 'hello'", input.Props().GetString("value"))
	}
}

// TestInputPassword tests password input masking
func TestInputPassword(t *testing.T) {
	input := InputBuilder().
		Type(InputTypePassword).
		Value("secret").
		Build()

	inputVNode, ok := input.(*InputVNode)
	if !ok {
		t.Fatal("Build() did not return *InputVNode")
	}

	if inputVNode.InputType() != InputTypePassword {
		t.Errorf("Type = %v, want InputTypePassword", inputVNode.InputType())
	}

	if inputVNode.Value() != "secret" {
		t.Errorf("Value should still be 'secret', got %v", inputVNode.Value())
	}
}

// TestInputMaxLength tests max length constraint
func TestInputMaxLength(t *testing.T) {
	input := NewInput()
	input.SetMaxLength(5)

	if input.MaxLength() != 5 {
		t.Errorf("MaxLength = %v, want 5", input.MaxLength())
	}

	input.SetMaxLength(0)
	if input.MaxLength() != 0 {
		t.Errorf("MaxLength = %v, want 0 (no limit)", input.MaxLength())
	}
}

// TestInputHandlers tests input event handlers
func TestInputHandlers(t *testing.T) {
	onChangeCalled := false
	onSubmitCalled := false

	input := InputBuilder().
		OnChange(func(s string) {
			onChangeCalled = true
		}).
		OnSubmit(func() {
			onSubmitCalled = true
		}).
		Build()

	inputVNode, ok := input.(*InputVNode)
	if !ok {
		t.Fatal("Build() did not return *InputVNode")
	}

	// Test onChange
	if inputVNode.OnChange() == nil {
		t.Error("OnChange should not be nil")
	} else {
		inputVNode.OnChange()("test")
		if !onChangeCalled {
			t.Error("OnChange handler was not called")
		}
	}

	// Test onSubmit
	if inputVNode.OnSubmitFunc() == nil {
		t.Error("OnSubmit should not be nil")
	} else {
		inputVNode.OnSubmitFunc()()
		if !onSubmitCalled {
			t.Error("OnSubmit handler was not called")
		}
	}
}

// TestInputFocus tests input focus state
func TestInputFocus(t *testing.T) {
	input := NewInput()

	if input.IsFocused() {
		t.Error("Input should not be focused initially")
	}

	input.SetFocus(true)
	if !input.IsFocused() {
		t.Error("Input should be focused after SetFocus(true)")
	}

	input.SetFocus(false)
	if input.IsFocused() {
		t.Error("Input should not be focused after SetFocus(false)")
	}
}

// TestNewTextarea tests textarea creation
func TestNewTextarea(t *testing.T) {
	textarea := NewTextarea()

	if textarea == nil {
		t.Fatal("NewTextarea() returned nil")
	}

	if textarea.Value() != "" {
		t.Errorf("Initial value = %v, want empty", textarea.Value())
	}

	if textarea.Rows() != 3 {
		t.Errorf("Default rows = %v, want 3", textarea.Rows())
	}

	if textarea.Cols() != 40 {
		t.Errorf("Default cols = %v, want 40", textarea.Cols())
	}
}

// TestTextareaBuilder tests textarea builder
func TestTextareaBuilder(t *testing.T) {
	textarea := TextareaBuilder().
		Value("test content").
		Placeholder("Enter text...").
		Rows(5).
		Cols(60).
		MaxLength(100).
		Disabled(true).
		Build()

	textareaVNode, ok := textarea.(*TextareaVNode)
	if !ok {
		t.Fatal("TextareaBuilder() did not return *TextareaVNode")
	}

	if textareaVNode.Value() != "test content" {
		t.Errorf("Value = %v, want 'test content'", textareaVNode.Value())
	}

	if textareaVNode.Rows() != 5 {
		t.Errorf("Rows = %v, want 5", textareaVNode.Rows())
	}

	if textareaVNode.Cols() != 60 {
		t.Errorf("Cols = %v, want 60", textareaVNode.Cols())
	}

	if textareaVNode.MaxLength() != 100 {
		t.Errorf("MaxLength = %v, want 100", textareaVNode.MaxLength())
	}

	if !textareaVNode.Disabled() {
		t.Error("Disabled should be true")
	}
}

// TestTextareaSetValue tests setting textarea value
func TestTextareaSetValue(t *testing.T) {
	textarea := NewTextarea()

	textarea.SetValue("line1\nline2")
	if textarea.Value() != "line1\nline2" {
		t.Errorf("Value = %v, want 'line1\\nline2'", textarea.Value())
	}
}

// TestInputTypes tests all input types
func TestInputTypes(t *testing.T) {
	types := []InputType{
		InputTypeText,
		InputTypePassword,
		InputTypeNumber,
		InputTypeEmail,
	}

	for _, inputType := range types {
		input := InputBuilder().Type(inputType).Build()
		if input.(*InputVNode).InputType() != inputType {
			t.Errorf("Type = %v, want %v", input.(*InputVNode).InputType(), inputType)
		}
	}
}

// BenchmarkInputBuilder benchmarks input builder
func BenchmarkInputBuilder(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		InputBuilder().
			Value("test").
			Placeholder("placeholder").
			MaxLength(10).
			Build()
	}
}

// BenchmarkNewInput benchmarks input creation
func BenchmarkNewInput(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewInput()
	}
}

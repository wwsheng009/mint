package ui

import "testing"

func TestFormShortcuts(t *testing.T) {
	builder := NewFormBuilder()
	if builder == nil {
		t.Fatal("NewFormBuilder() returned nil")
	}

	formNode := NewForm("loginForm").
		Label("Login").
		Layout(FormHorizontal).
		AddChildren(
			NewFormItem("username", Input("Username")).
				Label("Username").
				Build(),
		)
	if formNode == nil {
		t.Fatal("NewForm() returned nil")
	}
	if formNode.Tag() != "form" {
		t.Fatalf("NewForm().Tag() = %q, want form", formNode.Tag())
	}
	if formNode.Key() != "loginForm" {
		t.Fatalf("NewForm().Key() = %q, want loginForm", formNode.Key())
	}
	if len(formNode.Children()) != 1 {
		t.Fatalf("form children len = %d, want 1", len(formNode.Children()))
	}

	short := Form(NewFormItem("email", Input("Email")).Build())
	if short.Tag() != "form" {
		t.Fatalf("Form().Tag() = %q, want form", short.Tag())
	}
}

func TestInputTypeAliases(t *testing.T) {
	vnode := NewInputBuilder().
		Type(InputNumber).
		Value("42").
		Build()
	if vnode == nil {
		t.Fatal("NewInputBuilder().Build() returned nil")
	}
	if got := vnode.Props()["inputType"]; got != InputNumber {
		t.Fatalf("inputType = %v, want %v", got, InputNumber)
	}
}

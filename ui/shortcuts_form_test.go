package ui

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/ui/components/input"
)

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

func TestFormInputItemShortcuts(t *testing.T) {
	item := FormInputItem(
		"baseURL",
		"Gateway URL",
		"http://127.0.0.1:8080",
		FormInputPlaceholder("http://127.0.0.1:8080"),
		FormInputWidth(48),
		FormInputLabelWidth(14),
		FormInputForForm("loginForm"),
		FormInputLayout(FormInline),
		FormInputValidators(Required(), MinLength(8), MaxLength(128)),
		FormInputHelp("Use the Admin API base URL."),
		FormInputRequired(),
	)
	if item == nil {
		t.Fatal("FormInputItem() returned nil")
	}
	props := item.Props()
	if got := props["field"]; got != "baseURL" {
		t.Fatalf("field = %v, want baseURL", got)
	}
	if got := props["formID"]; got != "loginForm" {
		t.Fatalf("formID = %v, want loginForm", got)
	}
	if got := props["itemLayout"]; got != FormInline {
		t.Fatalf("itemLayout = %v, want %v", got, FormInline)
	}
	if got := props["itemLabelWidth"]; got != 14 {
		t.Fatalf("itemLabelWidth = %v, want 14", got)
	}
	if got := props["help"]; got != "Use the Admin API base URL." {
		t.Fatalf("help = %v, want helper text", got)
	}
	if got := props["required"]; got != true {
		t.Fatalf("required = %v, want true", got)
	}
	validators, ok := props["validators"].([]Validator)
	if !ok {
		t.Fatalf("validators = %T, want []Validator", props["validators"])
	}
	if len(validators) != 3 {
		t.Fatalf("validators len = %d, want 3", len(validators))
	}
	child, ok := props["child"].(*input.VNode)
	if !ok {
		t.Fatalf("child = %T, want *input.VNode", props["child"])
	}
	childProps := child.Props()
	if got := childProps["placeholder"]; got != "http://127.0.0.1:8080" {
		t.Fatalf("placeholder = %v, want http://127.0.0.1:8080", got)
	}
	if got := childProps["width"]; got != 48 {
		t.Fatalf("width = %v, want 48", got)
	}
	if _, ok := childProps["changeIntent"].(intent.FieldBinding); !ok {
		t.Fatalf("changeIntent = %T, want intent.FieldBinding", childProps["changeIntent"])
	}
	if got := childProps["formID"]; got != "loginForm" {
		t.Fatalf("child formID = %v, want loginForm", got)
	}

	horizontal := FormInputItem("captcha", "Captcha", "", FormInputHorizontal(18))
	horizontalProps := horizontal.Props()
	if got := horizontalProps["itemLayout"]; got != FormHorizontal {
		t.Fatalf("horizontal itemLayout = %v, want %v", got, FormHorizontal)
	}
	if got := horizontalProps["itemLabelWidth"]; got != 18 {
		t.Fatalf("horizontal itemLabelWidth = %v, want 18", got)
	}

	formNode := NewForm("loginForm").
		Layout(FormHorizontal).
		AddChild(item)
	if formNode == nil {
		t.Fatal("form with FormInputItem returned nil")
	}
	if len(formNode.Children()) != 1 {
		t.Fatalf("form children len = %d, want 1", len(formNode.Children()))
	}

	password := FormPasswordItem("token", "Token", "", FormInputWidth(64))
	if password == nil {
		t.Fatal("FormPasswordItem() returned nil")
	}
	passwordChild := password.Props()["child"].(*input.VNode)
	if got := passwordChild.Props()["inputType"]; got != InputPassword {
		t.Fatalf("password inputType = %v, want %v", got, InputPassword)
	}

	search := FormSearchItem("query", "Search", "", FormInputWidth(32))
	if search == nil {
		t.Fatal("FormSearchItem() returned nil")
	}
	searchChild := search.Props()["child"].(*input.VNode)
	if got := searchChild.Props()["searchVariant"]; got != true {
		t.Fatalf("searchVariant = %v, want true", got)
	}

	reason := FormOperationReasonItem("actionReason", "maintenance", FormInputWidth(72))
	if reason == nil {
		t.Fatal("FormOperationReasonItem() returned nil")
	}
	reasonProps := reason.Props()
	if got := reasonProps["itemLabel"]; got != "Reason" {
		t.Fatalf("operation reason label = %v, want Reason", got)
	}
	if got := reasonProps["required"]; got != true {
		t.Fatalf("operation reason required = %v, want true", got)
	}
	if got := reasonProps["help"]; got != "Required before preparing any operation; shown in the confirmation dialog and sent where the API supports it." {
		t.Fatalf("operation reason help = %v, want standard helper", got)
	}
	reasonChild := reasonProps["child"].(*input.VNode)
	if got := reasonChild.Props()["placeholder"]; got != "maintenance window, incident response, provider recovery" {
		t.Fatalf("operation reason placeholder = %v, want standard placeholder", got)
	}
	if got := reasonChild.Props()["width"]; got != 72 {
		t.Fatalf("operation reason width = %v, want override width 72", got)
	}
}

func TestFormInputItemCustomOption(t *testing.T) {
	item := FormInputItem("email", "Email", "", func(cfg *FormInputItemConfig) {
		cfg.Placeholder = "name@example.com"
		cfg.Validators = []Validator{Email()}
	})
	props := item.Props()
	child := props["child"].(*input.VNode)
	if got := child.Props()["placeholder"]; got != "name@example.com" {
		t.Fatalf("placeholder = %v, want name@example.com", got)
	}
	validators, ok := props["validators"].([]Validator)
	if !ok || len(validators) != 1 {
		t.Fatalf("validators = %#v, want one validator", props["validators"])
	}
}

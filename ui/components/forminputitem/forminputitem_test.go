package forminputitem

import (
	"testing"

	"github.com/wwsheng009/mint/runtime/intent"
	formcomp "github.com/wwsheng009/mint/ui/components/form"
	"github.com/wwsheng009/mint/ui/components/input"
	"github.com/wwsheng009/mint/ui/components/validation"
)

func TestNew(t *testing.T) {
	item := New(
		"baseURL",
		"Gateway URL",
		"http://127.0.0.1:8080",
		Placeholder("http://127.0.0.1:8080"),
		Width(48),
		ForForm("loginForm"),
		Layout(formcomp.LayoutInline),
		Validators(validation.Required(), validation.MinLength(8), validation.MaxLength(128)),
		Help("Use the Admin API base URL."),
		Required(),
	)
	if item == nil {
		t.Fatal("New() returned nil")
	}
	props := item.Props()
	if got := props["field"]; got != "baseURL" {
		t.Fatalf("field = %v, want baseURL", got)
	}
	if got := props["formID"]; got != "loginForm" {
		t.Fatalf("formID = %v, want loginForm", got)
	}
	if got := props["itemLayout"]; got != formcomp.LayoutInline {
		t.Fatalf("itemLayout = %v, want %v", got, formcomp.LayoutInline)
	}
	if got := props["help"]; got != "Use the Admin API base URL." {
		t.Fatalf("help = %v, want helper text", got)
	}
	if got := props["required"]; got != true {
		t.Fatalf("required = %v, want true", got)
	}
	validators, ok := props["validators"].([]validation.Validator)
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
}

func TestPasswordAndSearch(t *testing.T) {
	password := Password("token", "Token", "", Width(64))
	if password == nil {
		t.Fatal("Password() returned nil")
	}
	passwordChild := password.Props()["child"].(*input.VNode)
	if got := passwordChild.Props()["inputType"]; got != input.TypePassword {
		t.Fatalf("password inputType = %v, want %v", got, input.TypePassword)
	}

	search := Search("query", "Search", "", Width(32))
	if search == nil {
		t.Fatal("Search() returned nil")
	}
	searchChild := search.Props()["child"].(*input.VNode)
	if got := searchChild.Props()["searchVariant"]; got != true {
		t.Fatalf("searchVariant = %v, want true", got)
	}
}

func TestCustomOption(t *testing.T) {
	item := New("email", "Email", "", func(cfg *Config) {
		cfg.Placeholder = "name@example.com"
		cfg.Validators = []validation.Validator{validation.Email()}
	})
	props := item.Props()
	child := props["child"].(*input.VNode)
	if got := child.Props()["placeholder"]; got != "name@example.com" {
		t.Fatalf("placeholder = %v, want name@example.com", got)
	}
	validators, ok := props["validators"].([]validation.Validator)
	if !ok || len(validators) != 1 {
		t.Fatalf("validators = %#v, want one validator", props["validators"])
	}
}

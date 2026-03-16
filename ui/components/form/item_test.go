package form

import (
	"strings"
	"testing"

	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/validation"
)

func TestFormInstance_ValidatorsAffectValidation(t *testing.T) {
	inst := NewInstance(rtui.Props{"key": "profileForm"})
	inst.AddValidator("email", validation.Email())

	inst.HandleIntent(FieldBlur("profileForm", "email", "invalid-email"))

	errText, ok := inst.GetError("email")
	if !ok {
		t.Fatal("expected validation error for email field")
	}
	if !strings.Contains(errText, "请输入有效的邮箱地址") {
		t.Fatalf("expected email validation error, got %q", errText)
	}
	if inst.IsValid() {
		t.Fatal("expected form to be invalid after failing validation")
	}
}

func TestFormItemInheritsParentFormID(t *testing.T) {
	child := rtui.NewElement("field").SetProps(rtui.Props{"key": "username-field"})
	item := NewItem("username", child).
		Label("Username").
		Build()

	formVNode := NewForm("accountForm").AddChild(item)
	children := formVNode.Children()
	if len(children) != 1 {
		t.Fatalf("expected one child, got %d", len(children))
	}

	gotFormID, _ := children[0].Props()[itemPropFormID].(string)
	if gotFormID != "accountForm" {
		t.Fatalf("expected inherited formID accountForm, got %q", gotFormID)
	}
}

func TestFormItemSubscribesAndRendersValidationError(t *testing.T) {
	ResetRegistry()
	defer ResetRegistry()

	formInst := NewInstance(rtui.Props{
		"key":    "profileForm",
		"layout": LayoutInline,
	})
	RegisterForm("profileForm", formInst)
	defer UnregisterForm("profileForm")

	ctx := rtui.NewComponentContext("FormItem")
	updateCount := 0
	ctx.SetScheduleUpdate(func() {
		updateCount++
	})

	rtui.SetCurrentContext(ctx)
	defer rtui.SetCurrentContext(nil)

	child := rtui.NewElement("field").SetProps(rtui.Props{"key": "email-field"})
	item := NewItem("email", child).
		Label("Email").
		ForForm("profileForm").
		Validators(validation.Required(), validation.Email()).
		Build()

	first := renderFormItem(item.Props())
	if err := ctx.FinishRender(); err != nil {
		t.Fatalf("finish render failed: %v", err)
	}

	if first.Tag() != "hstack" {
		t.Fatalf("expected inline form item to render as hstack, got %s", first.Tag())
	}

	childFormID, _ := child.Props()[itemPropFormID].(string)
	if childFormID != "profileForm" {
		t.Fatalf("expected child formID to be injected, got %q", childFormID)
	}

	formInst.HandleIntent(FieldBlur("profileForm", "email", "invalid-email"))
	if updateCount == 0 {
		t.Fatal("expected form item subscription to schedule an update")
	}

	ctx.ResetContext()
	second := renderFormItem(item.Props())
	if err := ctx.FinishRender(); err != nil {
		t.Fatalf("finish render after validation failed: %v", err)
	}

	renderedText := strings.Join(collectText(second), "\n")
	if !strings.Contains(renderedText, "Email") {
		t.Fatalf("expected rendered label, got %q", renderedText)
	}
	if !strings.Contains(renderedText, "请输入有效的邮箱地址") {
		t.Fatalf("expected rendered validation error, got %q", renderedText)
	}

	ctx.CleanupAll()
	if sources := formInst.validatorSources["email"]; len(sources) != 0 {
		t.Fatalf("expected validator source cleanup, got %d remaining sources", len(sources))
	}
}

func TestFormItemResolvesAncestorFormWithoutRegistry(t *testing.T) {
	ResetRegistry()
	defer ResetRegistry()

	formInst := NewInstance(rtui.Props{
		"key":    "profileForm",
		"layout": LayoutInline,
	})

	child := rtui.NewElement("field").SetProps(rtui.Props{"key": "email-field"})
	item := NewItem("email", child).
		Label("Email").
		ForForm("profileForm").
		Validators(validation.Required(), validation.Email()).
		Build()

	owner := rtui.NewBaseComponentInstanceWithProps("FormItem", renderFormItem, item.Props())
	formInst.AddChild(owner)

	ctx := owner.GetContext()
	updateCount := 0
	ctx.SetScheduleUpdate(func() {
		updateCount++
	})

	first := owner.Render()
	if err := ctx.FinishRender(); err != nil {
		t.Fatalf("finish render failed: %v", err)
	}

	if first.Tag() != "hstack" {
		t.Fatalf("expected inline form item to render as hstack via ancestor form, got %s", first.Tag())
	}

	formInst.HandleIntent(FieldBlur("profileForm", "email", "invalid-email"))
	if updateCount == 0 {
		t.Fatal("expected ancestor-backed form item subscription to schedule an update")
	}

	ctx.ResetContext()
	second := owner.Render()
	if err := ctx.FinishRender(); err != nil {
		t.Fatalf("finish render after validation failed: %v", err)
	}

	renderedText := strings.Join(collectText(second), "\n")
	if !strings.Contains(renderedText, "请输入有效的邮箱地址") {
		t.Fatalf("expected rendered validation error via ancestor form, got %q", renderedText)
	}

	ctx.CleanupAll()
	if sources := formInst.validatorSources["email"]; len(sources) != 0 {
		t.Fatalf("expected validator source cleanup, got %d remaining sources", len(sources))
	}
}

func collectText(node rtui.VNode) []string {
	if node == nil {
		return nil
	}

	var values []string
	if provider, ok := node.(interface{ Content() string }); ok {
		if content := provider.Content(); content != "" {
			values = append(values, content)
		}
	}

	for _, child := range node.Children() {
		values = append(values, collectText(child)...)
	}
	return values
}

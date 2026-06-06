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

func TestFormItemHelpAndRequiredRender(t *testing.T) {
	child := rtui.NewElement("field").SetProps(rtui.Props{"key": "endpoint-field"})
	item := NewItem("endpoint", child).
		Label("Gateway URL").
		Help("Use the Admin API base URL.").
		Required(true).
		Build()

	rendered := renderFormItem(item.Props())
	text := strings.Join(collectText(rendered), "\n")
	if !strings.Contains(text, "Gateway URL *") {
		t.Fatalf("expected required label marker, got %q", text)
	}
	if !strings.Contains(text, "Use the Admin API base URL.") {
		t.Fatalf("expected help text, got %q", text)
	}
}

func TestFormItemLabelWidthPropagatesToLabelStyle(t *testing.T) {
	child := rtui.NewElement("field").SetProps(rtui.Props{"key": "endpoint-field"})
	item := NewItem("endpoint", child).
		Label("Gateway URL").
		LabelWidth(18).
		Build()

	rendered := renderFormItem(item.Props())
	if len(rendered.Children()) == 0 {
		t.Fatal("expected rendered form item children")
	}
	label := rendered.Children()[0]
	if got := label.Style().Width; got != 18 {
		t.Fatalf("label style width = %d, want 18", got)
	}
}

func TestFormItemOwnerlessRenderDoesNotResolveRegisteredForm(t *testing.T) {
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

	if first.Tag() != "vstack" {
		t.Fatalf("expected ownerless form item to stay unresolved, got %s", first.Tag())
	}

	childFormID, _ := child.Props()[itemPropFormID].(string)
	if childFormID != "profileForm" {
		t.Fatalf("expected child formID to be injected, got %q", childFormID)
	}
	if updateCount != 1 {
		t.Fatalf("expected ownerless unresolved form item to queue one retry, got %d", updateCount)
	}

	formInst.HandleIntent(FieldBlur("profileForm", "email", "invalid-email"))
	if updateCount != 1 {
		t.Fatalf("expected no subscription-driven update from registry form, got %d updates", updateCount)
	}

	ctx.CleanupAll()
	if sources := formInst.validatorSources["email"]; len(sources) != 0 {
		t.Fatalf("expected ownerless path to avoid registry validator source, got %d remaining sources", len(sources))
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

func TestFormItemAncestorResolutionWinsOverRegistryFallback(t *testing.T) {
	ResetRegistry()
	defer ResetRegistry()

	ancestorForm := NewInstance(rtui.Props{
		"key":    "profileForm",
		"layout": LayoutInline,
	})
	registryForm := NewInstance(rtui.Props{
		"key":    "profileForm",
		"layout": LayoutVertical,
	})
	RegisterForm("profileForm", registryForm)
	defer UnregisterForm("profileForm")

	child := rtui.NewElement("field").SetProps(rtui.Props{"key": "email-field"})
	item := NewItem("email", child).
		Label("Email").
		ForForm("profileForm").
		Validators(validation.Required(), validation.Email()).
		Build()

	owner := rtui.NewBaseComponentInstanceWithProps("FormItem", renderFormItem, item.Props())
	ancestorForm.AddChild(owner)

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
		t.Fatalf("expected ancestor inline layout to win over registry fallback, got %s", first.Tag())
	}

	ancestorForm.HandleIntent(FieldBlur("profileForm", "email", "invalid-email"))
	if updateCount == 0 {
		t.Fatal("expected ancestor-backed form item subscription to schedule an update")
	}

	ctx.CleanupAll()

	if sources := ancestorForm.validatorSources["email"]; len(sources) != 0 {
		t.Fatalf("expected ancestor validator sources to be cleaned up, got %d", len(sources))
	}
	if sources := registryForm.validatorSources["email"]; len(sources) != 0 {
		t.Fatalf("expected registry fallback form to stay untouched, got %d", len(sources))
	}
}

func TestFormItemDoesNotCrossTreeToRegistryWhenOwnerExists(t *testing.T) {
	ResetRegistry()
	defer ResetRegistry()

	registryForm := NewInstance(rtui.Props{
		"key":    "profileForm",
		"layout": LayoutInline,
	})
	RegisterForm("profileForm", registryForm)
	defer UnregisterForm("profileForm")

	child := rtui.NewElement("field").SetProps(rtui.Props{"key": "email-field"})
	item := NewItem("email", child).
		Label("Email").
		ForForm("profileForm").
		Validators(validation.Required(), validation.Email()).
		Build()

	owner := rtui.NewBaseComponentInstanceWithProps("FormItem", renderFormItem, item.Props())

	ctx := owner.GetContext()
	updateCount := 0
	ctx.SetScheduleUpdate(func() {
		updateCount++
	})

	first := owner.Render()
	if err := ctx.FinishRender(); err != nil {
		t.Fatalf("finish render failed: %v", err)
	}

	if first.Tag() != "vstack" {
		t.Fatalf("expected unresolved cross-tree form item to stay vertical, got %s", first.Tag())
	}
	if updateCount != 0 {
		t.Fatalf("expected owner-bound unresolved form item to avoid retry, got %d retries", updateCount)
	}

	ctx.CleanupAll()
	if sources := registryForm.validatorSources["email"]; len(sources) != 0 {
		t.Fatalf("expected registry form to remain untouched, got %d validator sources", len(sources))
	}
}

func TestFormItemOwnerlessResolutionQueuesSingleRetry(t *testing.T) {
	ResetRegistry()
	defer ResetRegistry()

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
		ForForm("missingForm").
		Validators(validation.Required()).
		Build()

	first := renderFormItem(item.Props())
	if err := ctx.FinishRender(); err != nil {
		t.Fatalf("finish render failed: %v", err)
	}

	if first.Tag() != "vstack" {
		t.Fatalf("expected unresolved ownerless form item to stay vertical, got %s", first.Tag())
	}
	if updateCount != 1 {
		t.Fatalf("expected ownerless unresolved form item to queue one retry, got %d", updateCount)
	}
}

func TestFormItemHidesUntouchedErrorsUntilSubmit(t *testing.T) {
	ResetRegistry()
	defer ResetRegistry()

	formInst := NewInstance(rtui.Props{
		"key":    "profileForm",
		"layout": LayoutVertical,
		"values": map[string]interface{}{
			"email": "",
		},
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

	if renderedText := strings.Join(collectText(first), "\n"); strings.Contains(renderedText, "请输入有效的邮箱地址") || strings.Contains(renderedText, "必填") {
		t.Fatalf("expected untouched field errors to stay hidden on first render, got %q", renderedText)
	}
	updateCount = 0

	formInst.HandleIntent(Submit("profileForm", nil))
	if updateCount == 0 {
		t.Fatal("expected submit-driven validate-all to schedule an update")
	}

	ctx.ResetContext()
	second := owner.Render()
	if err := ctx.FinishRender(); err != nil {
		t.Fatalf("finish render after submit failed: %v", err)
	}

	renderedText := strings.Join(collectText(second), "\n")
	if !strings.Contains(renderedText, "Email") {
		t.Fatalf("expected rendered label after submit, got %q", renderedText)
	}
	if !strings.Contains(renderedText, "此字段为必填项") {
		t.Fatalf("expected submit to reveal required validation error, got %q", renderedText)
	}

	ctx.CleanupAll()
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

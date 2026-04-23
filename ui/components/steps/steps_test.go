package steps

import (
	"strings"
	"testing"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/action"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	runtimemsg "github.com/wwsheng009/mint/runtime/msg"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

func TestNew(t *testing.T) {
	v := New([]Item{Step("Login"), Step("Verify")})
	if v == nil {
		t.Fatal("New returned nil")
	}
	if v.Tag() != "steps" {
		t.Fatalf("Tag = %q, want steps", v.Tag())
	}
	if got := len(v.Items()); got != 2 {
		t.Fatalf("Items len = %d, want 2", got)
	}
	if v.Direction() != DirectionHorizontal {
		t.Fatalf("Direction = %v, want horizontal", v.Direction())
	}
}

func TestVNodeSetPropsClonesItems(t *testing.T) {
	source := []Item{Step("Login").WithDescription("original")}
	v := New(nil)
	v.SetProps(rtui.Props{
		propItems:             source,
		propCurrent:           1,
		propCurrentControlled: true,
		propDirection:         DirectionVertical,
	})
	source[0].Title = "mutated"

	items := v.Items()
	if items[0].Title != "Login" {
		t.Fatalf("Items should be cloned, got %q", items[0].Title)
	}
	if v.Current() != 1 {
		t.Fatalf("Current = %d, want 1", v.Current())
	}
	if !v.currentControlled {
		t.Fatal("Current should be controlled")
	}
	if v.Direction() != DirectionVertical {
		t.Fatalf("Direction = %v, want vertical", v.Direction())
	}
}

func TestBuilderFluent(t *testing.T) {
	v := NewBuilder().
		Key("flow").
		ComponentID("checkout.steps").
		Titles("Login", "Verify").
		Item(Step("Deploy").WithDescription("release")).
		Current(1).
		Vertical().
		ProgressDot(true).
		Percent(42).
		OnChange(testChangeIntent{}).
		TitleStyle(style.NewStyle().Bold(true)).
		Build()

	if v.Key() != "flow" {
		t.Fatalf("Key = %q, want flow", v.Key())
	}
	if v.componentID != "checkout.steps" {
		t.Fatalf("componentID = %q, want checkout.steps", v.componentID)
	}
	if got := len(v.Items()); got != 3 {
		t.Fatalf("Items len = %d, want 3", got)
	}
	if v.Current() != 1 || !v.currentControlled {
		t.Fatalf("Current state = (%d,%v), want (1,true)", v.Current(), v.currentControlled)
	}
	if !v.progressDot || v.percent != 42 {
		t.Fatalf("progress props = (%v,%d), want (true,42)", v.progressDot, v.percent)
	}
	if v.Direction() != DirectionVertical {
		t.Fatalf("Direction = %v, want vertical", v.Direction())
	}
}

func TestInstanceMeasureHorizontal(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propItems:          []Item{Step("Login"), Step("Verify"), Step("Deploy")},
		propInitialCurrent: 1,
	})
	size := inst.Measure(layout.Constraints{MaxWidth: 80, MaxHeight: 3})
	wantWidth := paint.StringWidth("[✓] Login ── [2] Verify ── [3] Deploy")
	if size.Width != wantWidth {
		t.Fatalf("Width = %d, want %d", size.Width, wantWidth)
	}
	if size.Height != 1 {
		t.Fatalf("Height = %d, want 1", size.Height)
	}
}

func TestInstancePaintHorizontalRendersStatuses(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propItems: []Item{
			Step("Login"),
			Step("Verify"),
			Step("Deploy").AsError(),
		},
		propInitialCurrent: 1,
	})
	inst.SetBounds(0, 0, 80, 1)
	cmds := inst.Paint(0, 0)
	text := collectText(cmds)

	if text != "[✓] Login ── [2] Verify ── [!] Deploy" {
		t.Fatalf("Paint text = %q", text)
	}
	verifyStyle := styleForText(cmds, " Verify")
	if !verifyStyle.IsBold() {
		t.Fatal("current step title should be bold")
	}
	errorStyle := styleForText(cmds, " Deploy")
	if errorStyle.FG != theme.Error() {
		t.Fatalf("error step fg = %q, want %q", errorStyle.FG, theme.Error())
	}
}

func TestInstancePaintVerticalIncludesDescriptions(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propItems: []Item{
			Step("Login").WithDescription("done"),
			Step("Verify").WithDescription("current"),
		},
		propInitialCurrent: 1,
		propDirection:      DirectionVertical,
	})
	inst.SetBounds(0, 0, 40, 8)
	cmds := inst.Paint(0, 0)

	lines := collectLines(cmds)
	if len(lines) != 5 {
		t.Fatalf("line count = %d, want 5", len(lines))
	}
	if lines[0] != "[✓] Login" {
		t.Fatalf("line 0 = %q", lines[0])
	}
	if !strings.Contains(lines[1], "done") {
		t.Fatalf("line 1 = %q, want description", lines[1])
	}
	if !strings.Contains(lines[2], "│") {
		t.Fatalf("line 2 = %q, want connector", lines[2])
	}
	if lines[3] != "[2] Verify" {
		t.Fatalf("line 3 = %q", lines[3])
	}
	if !strings.Contains(lines[4], "current") {
		t.Fatalf("line 4 = %q, want current description", lines[4])
	}
}

func TestInstancePaintTruncatesToBounds(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propItems:          []Item{Step("Initialize"), Step("Verification"), Step("Deploy")},
		propInitialCurrent: 1,
	})
	inst.SetBounds(0, 0, 18, 1)
	cmds := inst.Paint(0, 0)
	text := collectText(cmds)
	if paint.StringWidth(text) > 18 {
		t.Fatalf("paint width = %d, want <= 18", paint.StringWidth(text))
	}
	if !strings.Contains(text, "…") {
		t.Fatalf("truncated text = %q, want ellipsis", text)
	}
}

func TestInstancePaintHorizontalUsesPercent(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propItems:          []Item{Step("Login"), Step("Verify"), Step("Deploy")},
		propInitialCurrent: 1,
		propPercent:        42,
	})
	inst.SetBounds(0, 0, 80, 1)
	text := collectText(inst.Paint(0, 0))
	if text != "[✓] Login ── [42%] Verify ── [3] Deploy" {
		t.Fatalf("paint text = %q", text)
	}
}

func TestInstancePaintHorizontalUsesProgressDot(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propItems:          []Item{Step("Login"), Step("Verify"), Step("Deploy")},
		propInitialCurrent: 1,
		propProgressDot:    true,
		propPercent:        50,
	})
	inst.SetBounds(0, 0, 80, 1)
	text := collectText(inst.Paint(0, 0))
	if text != "● Login ── ◕ Verify ── ○ Deploy" {
		t.Fatalf("paint text = %q", text)
	}
}

func TestInstanceSetPropsMarksDirty(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propItems: []Item{Step("Login"), Step("Verify")},
	})
	inst.MarkClean()

	changed := inst.SetProps(rtui.Props{
		propItems:     []Item{Step("Login"), Step("Deploy")},
		propDirection: DirectionVertical,
	})
	if !changed {
		t.Fatal("SetProps should report changes")
	}
	if !inst.IsDirty() {
		t.Fatal("SetProps should mark instance dirty")
	}
	if inst.direction != DirectionVertical {
		t.Fatalf("direction = %v, want vertical", inst.direction)
	}
}

func TestInstanceHandleActionNavigateEmitsIntents(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID:        "checkout.steps",
		propItems:              []Item{Step("Cart"), Step("Address"), Step("Pay")},
		propInitialCurrent:     0,
		propCurrentIntentField: intent.BindField("currentStep"),
	})
	var emitted []intent.Intent
	inst.SetIntentEmitter(func(i intent.Intent) {
		emitted = append(emitted, i)
	})

	if !inst.HandleAction(action.NewAction(action.ActionNavigateRight)) {
		t.Fatal("expected navigate right to change step")
	}
	if inst.GetCurrent() != 1 {
		t.Fatalf("current = %d, want 1", inst.GetCurrent())
	}
	if len(emitted) != 2 {
		t.Fatalf("emitted len = %d, want 2", len(emitted))
	}
	change, ok := emitted[0].(StepChangeIntent)
	if !ok {
		t.Fatalf("first emitted intent = %T, want StepChangeIntent", emitted[0])
	}
	if change.ComponentID != "checkout.steps" || change.FromIndex != 0 || change.ToIndex != 1 || change.StepTitle != "Address" {
		t.Fatalf("unexpected StepChangeIntent: %+v", change)
	}
	fieldChange, ok := emitted[1].(intent.FieldChangeIntent)
	if !ok {
		t.Fatalf("second emitted intent = %T, want FieldChangeIntent", emitted[1])
	}
	if fieldChange.Field != "currentStep" || fieldChange.Value != "1" {
		t.Fatalf("unexpected FieldChangeIntent: %+v", fieldChange)
	}
}

func TestInstanceHandleClickHorizontal(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propItems: []Item{Step("Login"), Step("Verify"), Step("Deploy")},
	})
	inst.SetBounds(0, 0, 80, 1)
	text := collectText(inst.Paint(0, 0))
	localX := strings.Index(text, "[2]") + 1
	mouse := runtimemsg.NewMouseMsgWithTarget(localX, 0, localX, 0, 0, runtimemsg.MouseLeft, runtimemsg.MouseActionPress)
	if !inst.HandleAction(action.NewAction(action.ActionClick).WithPayload(mouse)) {
		t.Fatal("expected click to change current step")
	}
	if inst.GetCurrent() != 1 {
		t.Fatalf("current = %d, want 1", inst.GetCurrent())
	}
}

func TestInstanceHandleClickVertical(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propItems: []Item{
			Step("Login").WithDescription("done"),
			Step("Verify").WithDescription("current"),
		},
		propDirection: DirectionVertical,
	})
	inst.SetBounds(0, 0, 40, 8)
	mouse := runtimemsg.NewMouseMsgWithTarget(2, 3, 2, 3, 0, runtimemsg.MouseLeft, runtimemsg.MouseActionPress)
	if !inst.HandleAction(action.NewAction(action.ActionClick).WithPayload(mouse)) {
		t.Fatal("expected vertical click to change current step")
	}
	if inst.GetCurrent() != 1 {
		t.Fatalf("current = %d, want 1", inst.GetCurrent())
	}
}

func TestInstanceHandleIntent(t *testing.T) {
	inst := NewInstance(rtui.Props{
		propComponentID: "checkout.steps",
		propItems:       []Item{Step("Login"), Step("Verify"), Step("Deploy")},
	})
	if !inst.HandleIntent(StepChangeWithID("checkout.steps", 0, 2, 3, "deploy", "Deploy")) {
		t.Fatal("expected StepChangeIntent to be handled")
	}
	if inst.GetCurrent() != 2 {
		t.Fatalf("current = %d, want 2", inst.GetCurrent())
	}
	if inst.HandleIntent(StepChangeWithID("other.steps", 2, 1, 3, "verify", "Verify")) {
		t.Fatal("expected StepChangeIntent for another component to be ignored")
	}
}

func TestItemHelpers(t *testing.T) {
	item := Step("Deploy").
		WithKey("deploy").
		WithDescription("ship it").
		WithIcon("D").
		AsProcess()
	if item.Key != "deploy" || item.Description != "ship it" || item.Icon != "D" || item.Status != StatusProcess {
		t.Fatalf("item helpers lost metadata: %+v", item)
	}
}

func TestVNodeImplementsInterfaces(t *testing.T) {
	var _ rtui.VNode = New(nil)
	var _ rtui.InstanceFactory = New(nil)
}

type testChangeIntent struct{}

func (testChangeIntent) IntentType() string { return "testChange" }

func collectText(cmds []paint.DrawCmd) string {
	var builder strings.Builder
	for _, cmd := range cmds {
		builder.WriteString(cmd.Text)
	}
	return builder.String()
}

func styleForText(cmds []paint.DrawCmd, text string) style.Style {
	for _, cmd := range cmds {
		if cmd.Text == text {
			return cmd.Style
		}
	}
	return style.Style{}
}

func collectLines(cmds []paint.DrawCmd) []string {
	maxY := -1
	for _, cmd := range cmds {
		if cmd.Y > maxY {
			maxY = cmd.Y
		}
	}
	if maxY < 0 {
		return nil
	}
	lines := make([]string, maxY+1)
	for _, cmd := range cmds {
		lines[cmd.Y] += cmd.Text
	}
	return lines
}

package confirmdialog

import (
	"testing"

	"github.com/wwsheng009/mint/internal/reconciler"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/layout"
	"github.com/wwsheng009/mint/runtime/types"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/button"
	"github.com/wwsheng009/mint/ui/components/descriptions"
)

type testIntent struct {
	name string
}

func (i testIntent) IntentType() string { return i.name }

func TestBuilderAndProps(t *testing.T) {
	dialog := NewBuilder().
		Key("confirm-disable").
		Title("Disable Key").
		Message("Disable the selected provider key.").
		Warning("Requests may fail over to another key.").
		Open(true).
		Width(72).
		Height(20).
		Target(Target("provider", "Provider", "openai")).
		Target(SensitiveTarget("key", "Key", "provider-key-demo")).
		ReasonField("actionReason").
		ReasonValue("maintenance").
		ReasonRequired(true).
		ConfirmPhrase("DISABLE", "confirmText", "DISABLE").
		ConfirmText("Disable").
		ConfirmVariant(button.VariantDanger).
		OnConfirm(testIntent{"confirm"}).
		OnCancel(testIntent{"cancel"}).
		BuildVNode()

	if dialog.Key() != "confirm-disable" {
		t.Fatalf("key = %q, want confirm-disable", dialog.Key())
	}
	props := dialog.Props()
	if got := props[propTitle]; got != "Disable Key" {
		t.Fatalf("title = %v, want Disable Key", got)
	}
	if got := props[propOpen]; got != true {
		t.Fatalf("open = %v, want true", got)
	}
	if got := props[propWidth]; got != 72 {
		t.Fatalf("width = %v, want 72", got)
	}
	items := props[propTargetItems].([]TargetItem)
	if len(items) != 2 || !items[1].Sensitive {
		t.Fatalf("target items = %#v", items)
	}
	if got := props[propConfirmVariant]; got != button.VariantDanger {
		t.Fatalf("confirm variant = %v, want danger", got)
	}
	if got := props[propConfirmPhrase]; got != "DISABLE" {
		t.Fatalf("confirm phrase = %v, want DISABLE", got)
	}
	if got := props[propConfirmPhraseField]; got != "confirmText" {
		t.Fatalf("confirm phrase field = %v, want confirmText", got)
	}
	if got := props[propConfirmPhraseValue]; got != "DISABLE" {
		t.Fatalf("confirm phrase value = %v, want DISABLE", got)
	}
	if _, ok := props[propConfirmIntent].(intent.Intent); !ok {
		t.Fatalf("confirm intent = %T, want intent.Intent", props[propConfirmIntent])
	}
}

func TestRuntimeChildrenClosedDialogIsEmpty(t *testing.T) {
	inst := NewBuilder().Key("closed").Open(false).BuildInstance()
	if children := inst.RuntimeChildren(); len(children) != 0 {
		t.Fatalf("RuntimeChildren len = %d, want 0", len(children))
	}
}

func TestInstanceMeasureDoesNotReservePageFlowSpace(t *testing.T) {
	inst := NewBuilder().
		Key("confirm-host").
		Title("Enable Key").
		Open(true).
		Width(82).
		Height(29).
		BuildInstance()

	size := inst.Measure(layout.Constraints{MaxWidth: 80, MaxHeight: 24})
	if size.Width != 0 || size.Height != 0 {
		t.Fatalf("confirmdialog host size = %dx%d, want 0x0", size.Width, size.Height)
	}
	if children := inst.RuntimeChildren(); len(children) != 1 || children[0].Tag() != "modal" {
		t.Fatalf("RuntimeChildren = %#v, want modal child despite zero host measure", children)
	}
}

func TestRuntimeChildrenUseCenteredModalLayer(t *testing.T) {
	inst := NewBuilder().
		Key("confirm-layer").
		Title("Disable Key").
		Message("Disable the selected key.").
		Open(true).
		BuildInstance()

	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	if got := children[0].GetLayer(); got != rtui.LayerModal {
		t.Fatalf("dialog layer = %v, want %v", got, rtui.LayerModal)
	}
	if got := children[0].Tag(); got != "modal" {
		t.Fatalf("dialog tag = %q, want modal", got)
	}
	props := children[0].Props()
	if got := props["isOpen"]; got != true {
		t.Fatalf("modal isOpen = %v, want true", got)
	}
	if got := props["centered"]; got != true {
		t.Fatalf("modal centered = %v, want true", got)
	}
	if got := props["closeable"]; got != false {
		t.Fatalf("modal closeable = %v, want false", got)
	}
	if got := props["closeOnEsc"]; got != false {
		t.Fatalf("modal closeOnEsc = %v, want false", got)
	}
	if got := props["closeOnBackdrop"]; got != false {
		t.Fatalf("modal closeOnBackdrop = %v, want false", got)
	}

	fiber := reconciler.CreateFiberFromVNode(children[0])
	reconciler.SyncPositioningProperties(fiber)
	if fiber.Position != types.PositionFixed {
		t.Fatalf("modal position = %v, want %v", fiber.Position, types.PositionFixed)
	}
	if fiber.Anchor != types.AnchorCenter {
		t.Fatalf("modal anchor = %v, want %v", fiber.Anchor, types.AnchorCenter)
	}
}

func TestRuntimeChildrenBuildsModalWithReasonAndFooter(t *testing.T) {
	inst := NewBuilder().
		Key("confirm-disable").
		Title("Disable Key").
		Message("Disable the selected key.").
		Open(true).
		Target(Target("provider", "Provider", "openai")).
		ReasonField("actionReason").
		ReasonRequired(true).
		ConfirmText("Disable").
		OnConfirm(testIntent{"confirm"}).
		OnCancel(testIntent{"cancel"}).
		BuildInstance()

	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	root := children[0]
	if root.Tag() != "modal" {
		t.Fatalf("root tag = %q, want modal", root.Tag())
	}
	if findVNodeByKey(root, "confirm-disable-reason-input") == nil {
		t.Fatal("reason input not found")
	}
	confirm := findVNodeByKey(root, "confirm-disable-confirm")
	if confirm == nil {
		t.Fatal("confirm button not found")
	}
	if got := confirm.Props()["disabled"]; got != true {
		t.Fatalf("confirm disabled = %v, want true while reason is empty", got)
	}
	cancel := findVNodeByKey(root, "confirm-disable-cancel")
	if cancel == nil {
		t.Fatal("cancel button not found")
	}
}

func TestRuntimeChildrenUseCompactTargetColumnsForLargeOperation(t *testing.T) {
	inst := NewBuilder().
		Key("confirm-enable").
		Title("Enable Key").
		Message("Enable the selected provider key.").
		Open(true).
		Width(82).
		Height(28).
		Target(Target("endpoint", "Endpoint", "http://127.0.0.1:8080")).
		Target(APITarget("POST", "/admin/loadbalancer/providers/{provider}/keys/{key}/enable?group_name={group}")).
		Target(ImpactTarget("moderate", "selected key may receive traffic")).
		Target(Target("group", "Group", "default")).
		Target(Target("provider", "Provider", "openai")).
		Target(SensitiveTarget("key", "Key", "provider-key-demo")).
		ReasonField("actionReason").
		ReasonValue("maintenance").
		ReasonRequired(true).
		ConfirmPhrase("ENABLE", "confirmPhrase", "ENABLE").
		ConfirmText("Enable").
		BuildInstance()

	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	targets := findVNodeByKey(children[0], "confirm-enable-targets")
	if targets == nil {
		t.Fatal("target summary not found")
	}
	props := targets.Props()
	if got := props["column"]; got != 2 {
		t.Fatalf("target summary column = %v, want 2 for large operation", got)
	}
	if got := props["contentWidth"]; got != 21 {
		t.Fatalf("target summary content width = %v, want compact two-column width", got)
	}
}

func TestRuntimeChildrenConstrainTargetsWhenReasonPhraseAndFooterNeedRoom(t *testing.T) {
	inst := NewBuilder().
		Key("confirm-enable").
		Title("Enable Key").
		Message("Enable the selected provider key.").
		Open(true).
		Width(82).
		Height(26).
		Target(Target("endpoint", "Endpoint", "http://127.0.0.1:8080")).
		Target(APITarget("POST", "/admin/loadbalancer/providers/{provider}/keys/{key}/enable?group_name={group}")).
		Target(ImpactTarget("moderate", "selected key may receive traffic")).
		Target(Target("group", "Group", "default")).
		Target(Target("group_state", "Group State", "healthy")).
		Target(Target("provider", "Provider", "openai")).
		Target(Target("provider_state", "Provider State", "healthy")).
		Target(SensitiveTarget("key", "Key", "provider-key-demo")).
		Target(Target("key_state", "Key State", "disabled")).
		Target(Target("key_route_state", "Route State", "rate_limited")).
		ReasonField("actionReason").
		ReasonValue("maintenance").
		ReasonRequired(true).
		ConfirmPhrase("ENABLE", "confirmPhrase", "ENABLE").
		ConfirmText("Enable").
		BuildInstance()

	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	targets := findVNodeByKey(children[0], "confirm-enable-targets")
	if targets == nil {
		t.Fatal("target summary not found")
	}
	items, ok := targets.Props()["items"].([]descriptions.Item)
	if !ok {
		t.Fatalf("target items = %T, want []Item", targets.Props()["items"])
	}
	if len(items) != 4 {
		t.Fatalf("target items len = %d, want 4 critical targets", len(items))
	}
	wantLabels := []string{"Provider", "Key", "Key State", "Route State"}
	for index, want := range wantLabels {
		if items[index].Label != want {
			t.Fatalf("target item[%d] label = %q, want %q", index, items[index].Label, want)
		}
	}
	if findVNodeByKey(children[0], "confirm-enable-reason-input") == nil {
		t.Fatal("reason input should remain rendered")
	}
	if findVNodeByKey(children[0], "confirm-enable-confirm-phrase-input") == nil {
		t.Fatal("confirm phrase input should remain rendered")
	}
	if findVNodeByKey(children[0], "confirm-enable-cancel") == nil || findVNodeByKey(children[0], "confirm-enable-confirm") == nil {
		t.Fatal("footer buttons should remain rendered")
	}
}

func TestRuntimeChildrenConfirmPhraseDisablesUntilExactMatch(t *testing.T) {
	inst := NewBuilder().
		Key("confirm-delete").
		Title("Delete Rule").
		Message("Delete the selected alert rule.").
		Open(true).
		Height(20).
		ConfirmPhrase("DELETE", "confirmPhrase", "delete").
		ConfirmText("Delete").
		OnConfirm(testIntent{"confirm"}).
		BuildInstance()

	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	if findVNodeByKey(children[0], "confirm-delete-confirm-phrase-input") == nil {
		t.Fatal("confirm phrase input not found")
	}
	if got := children[0].Props()["height"]; got != minConfirmPhraseHeight {
		t.Fatalf("dialog height = %v, want min confirm phrase height", got)
	}
	confirm := findVNodeByKey(children[0], "confirm-delete-confirm")
	if confirm == nil {
		t.Fatal("confirm button not found")
	}
	if got := confirm.Props()["disabled"]; got != true {
		t.Fatalf("confirm disabled = %v, want true while phrase mismatches", got)
	}

	changed := inst.SetProps(NewBuilder().
		Key("confirm-delete").
		Open(true).
		ConfirmPhrase("DELETE", "confirmPhrase", "DELETE").
		BuildVNode().
		Props())
	if !changed {
		t.Fatal("SetProps should report changed confirm phrase value")
	}
	children = inst.RuntimeChildren()
	confirm = findVNodeByKey(children[0], "confirm-delete-confirm")
	if got := confirm.Props()["disabled"]; got != false {
		t.Fatalf("confirm disabled = %v, want false after exact phrase match", got)
	}
}

func TestRuntimeChildrenShowsConfirmDisabledReasons(t *testing.T) {
	inst := NewBuilder().
		Key("confirm-disable").
		Title("Disable Key").
		Open(true).
		ReasonLabel("Reason").
		ReasonRequired(true).
		ConfirmPhrase("DISABLE", "confirmPhrase", "disable").
		ConfirmText("Disable").
		OnConfirm(testIntent{"confirm"}).
		BuildInstance()

	children := inst.RuntimeChildren()
	if len(children) != 1 {
		t.Fatalf("RuntimeChildren len = %d, want 1", len(children))
	}
	reason := findVNodeByKey(children[0], "confirm-disable-disabled-reason")
	if reason == nil {
		t.Fatal("disabled reason text not found")
	}
	want := "Enter reason before confirming. Type DISABLE to confirm."
	if got := reason.Props()["content"]; got != want {
		t.Fatalf("disabled reason = %v, want %q", got, want)
	}

	changed := inst.SetProps(NewBuilder().
		Key("confirm-disable").
		Open(true).
		ReasonRequired(true).
		ReasonValue("maintenance").
		ConfirmPhrase("DISABLE", "confirmPhrase", "DISABLE").
		DisabledReason("Target is stale.").
		BuildVNode().
		Props())
	if !changed {
		t.Fatal("SetProps should report changed disabled reason inputs")
	}
	children = inst.RuntimeChildren()
	reason = findVNodeByKey(children[0], "confirm-disable-disabled-reason")
	if got := reason.Props()["content"]; got != "Target is stale." {
		t.Fatalf("external disabled reason = %v, want Target is stale.", got)
	}
}

func TestRuntimeChildrenShowsDisabledReasonWhenExternallyDisabled(t *testing.T) {
	inst := NewBuilder().
		Key("confirm-operation").
		Open(true).
		DisableConfirm(true).
		OnConfirm(testIntent{"confirm"}).
		BuildInstance()

	children := inst.RuntimeChildren()
	reason := findVNodeByKey(children[0], "confirm-operation-disabled-reason")
	if reason == nil {
		t.Fatal("disabled reason text not found")
	}
	if got := reason.Props()["content"]; got != "Confirmation is disabled." {
		t.Fatalf("disabled reason = %v, want fallback external disabled reason", got)
	}
}

func TestNormalizeTargetItems(t *testing.T) {
	items := normalizeTargetItems([]TargetItem{{Key: "dup"}, {Key: "dup"}, {}})
	if items[0].Key != "dup" || items[1].Key != "dup-1" || items[2].Key != "target-2" {
		t.Fatalf("item keys = %q, %q, %q", items[0].Key, items[1].Key, items[2].Key)
	}
}

func TestAPITargetCreatesStandardRouteTarget(t *testing.T) {
	item := APITarget("post", " /admin/runtime/reload ")
	if item.Key != "api" || item.Label != "API" || item.Value != "POST /admin/runtime/reload" || item.Sensitive {
		t.Fatalf("api target = %+v, want standard API route target", item)
	}
}

func TestImpactTargetCreatesStandardImpactTarget(t *testing.T) {
	item := ImpactTarget(" HIGH ", "Runtime state may change immediately.")
	if item.Key != "impact" || item.Label != "Impact" || item.Value != "high: Runtime state may change immediately." || item.Sensitive {
		t.Fatalf("impact target = %+v, want standard operational impact target", item)
	}
}

func findVNodeByKey(node rtui.VNode, key string) rtui.VNode {
	if node == nil {
		return nil
	}
	if node.Key() == key {
		return node
	}
	for _, child := range node.Children() {
		if found := findVNodeByKey(child, key); found != nil {
			return found
		}
	}
	return nil
}

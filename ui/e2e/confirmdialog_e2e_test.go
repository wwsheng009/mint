package e2e

import (
	"testing"
	"time"

	"github.com/wwsheng009/mint/runtime/intent"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
	confirmdialogcomp "github.com/wwsheng009/mint/ui/components/confirmdialog"
)

type confirmDialogTestIntent struct {
	name string
}

func (i confirmDialogTestIntent) IntentType() string { return i.name }

func newConfirmDialogStaticApp(reason string) ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("ConfirmDialog E2E Fixture").Build(),
				confirmdialogcomp.NewDangerOperation(
					"disable-key-confirm",
					"Disable Provider Key",
					"Disable the selected provider key and record an audit reason.",
					"Traffic may fail over to another available key.",
					"Disable",
					"actionReason",
					reason,
					confirmDialogTestIntent{"confirmdialog.confirm"},
					confirmDialogTestIntent{"confirmdialog.cancel"},
					confirmdialogcomp.Target("provider", "Provider", "openai"),
					confirmdialogcomp.SensitiveTarget("key", "Key", "provider-key-demo"),
				).
					Width(76).
					Height(26).
					RequirePhrase("DISABLE", "confirmPhrase", "DISABLE").
					Build(),
			})
	}
}

func newConfirmDialogTallTargetApp(reason string) ui.ComponentFunc {
	return func() ui.VNode {
		return ui.PageStack(
			ui.NewTextBuilder("ConfirmDialog Tall Fixture").Build(),
			confirmdialogcomp.NewDangerOperation(
				"enable-key-confirm",
				"Enable Provider Key",
				"Enable the selected provider key.",
				"The selected key may start receiving traffic again.",
				"Enable",
				"actionReason",
				reason,
				confirmDialogTestIntent{"confirmdialog.confirm"},
				confirmDialogTestIntent{"confirmdialog.cancel"},
				confirmdialogcomp.Target("endpoint", "Endpoint", "http://127.0.0.1:8080"),
				confirmdialogcomp.Target("api", "API", "POST /admin/loadbalancer/providers/{provider}/keys/{key}/enable?group_name={group}"),
				confirmdialogcomp.Target("impact", "Impact", "moderate: selected key may start receiving traffic."),
				confirmdialogcomp.Target("group", "Group", "default"),
				confirmdialogcomp.Target("group_state", "Group State", "healthy"),
				confirmdialogcomp.Target("provider", "Provider", "openai"),
				confirmdialogcomp.Target("provider_state", "Provider State", "healthy"),
				confirmdialogcomp.SensitiveTarget("key", "Key", "provider-key-demo"),
				confirmdialogcomp.Target("key_state", "Key State", "disabled"),
			).
				Width(82).
				Height(28).
				RequirePhrase("ENABLE", "confirmPhrase", "ENABLE").
				Build(),
		)
	}
}

func TestE2EConfirmDialogRendersAndMasksSensitiveTarget(t *testing.T) {
	app, err := Run(newConfirmDialogStaticApp("maintenance"), ui.WithSize(100, 32))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{"Disable Provider Key", "Disable the selected provider key", "Provider", "openai", "Reason *", "maintenance", "Confirmation *", "DISABLE", "Disable", "Cancel"} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatalf("expected %q to be visible: %v", text, err)
		}
	}
	if err := app.AssertVisible(ByText("provider-key-demo")); err == nil {
		t.Fatal("sensitive target value should be masked")
	}
	if err := app.AssertVisible(ByText("masked")); err != nil {
		t.Fatalf("expected masked sensitive placeholder: %v", err)
	}
}

func TestE2EConfirmDialogKeepsFooterVisibleInDefaultHeight(t *testing.T) {
	app, err := Run(newConfirmDialogStaticApp("maintenance"), ui.WithSize(100, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{"Disable Provider Key", "Cancel", "Disable"} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatalf("expected %q to be visible in 24-row viewport: %v\n%s", text, err, app.RenderString())
		}
	}

	bounds, err := app.BoundsOf(ByKey("disable-key-confirm-root"))
	if err != nil {
		t.Fatal(err)
	}
	if bounds.Y < 0 || bounds.Y+bounds.Height > 24 {
		t.Fatalf("dialog bounds = %+v, want within 100x24 viewport\n%s", bounds, app.RenderString())
	}
}

func TestE2EConfirmDialogKeepsFooterVisibleWithTallTargets(t *testing.T) {
	app, err := Run(newConfirmDialogTallTargetApp("maintenance"), ui.WithSize(100, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{"Enable Provider Key", "Cancel", "Enable"} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatalf("expected %q to be visible with tall targets: %v\n%s", text, err, app.RenderString())
		}
	}
	if err := app.AssertVisible(ByText("provider-key-demo")); err == nil {
		t.Fatal("sensitive target value should stay masked in tall target dialog")
	}

	bounds, err := app.BoundsOf(ByKey("enable-key-confirm-root"))
	if err != nil {
		t.Fatal(err)
	}
	if bounds.Y < 0 || bounds.Y+bounds.Height > 24 {
		t.Fatalf("dialog bounds = %+v, want within 100x24 viewport\n%s", bounds, app.RenderString())
	}
}

func TestE2EConfirmDialogConfirmDispatch(t *testing.T) {
	unregisters := make([]func(), 0, 1)
	initFn := func() {
		rt := rtui.GetGlobalIntentRuntime()
		if rt == nil {
			return
		}
		unregisters = append(unregisters,
			rt.Register("confirmdialog.confirm", intent.HandlerFunc(func(_ *intent.ActionContext, _ intent.Intent) intent.IntentResult {
				return intent.HandledResult()
			})),
		)
	}
	defer func() {
		for i := len(unregisters) - 1; i >= 0; i-- {
			unregisters[i]()
		}
	}()

	app, err := Run(newConfirmDialogStaticApp("maintenance"), ui.WithSize(100, 32), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	app.ClearIntentLogs()
	if err := app.Driver().Click(ByKey("disable-key-confirm-confirm")); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitIntent("confirmdialog.confirm", 500*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled("confirmdialog.confirm"); err != nil {
		t.Fatal(err)
	}
}

func TestE2EConfirmDialogCancelDispatch(t *testing.T) {
	unregisters := make([]func(), 0, 1)
	initFn := func() {
		rt := rtui.GetGlobalIntentRuntime()
		if rt == nil {
			return
		}
		unregisters = append(unregisters,
			rt.Register("confirmdialog.cancel", intent.HandlerFunc(func(_ *intent.ActionContext, _ intent.Intent) intent.IntentResult {
				return intent.HandledResult()
			})),
		)
	}
	defer func() {
		for i := len(unregisters) - 1; i >= 0; i-- {
			unregisters[i]()
		}
	}()

	app, err := Run(newConfirmDialogStaticApp("maintenance"), ui.WithSize(100, 32), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	point, err := app.ResolvePoint(ByKey("disable-key-confirm-cancel"))
	if err != nil {
		t.Fatal(err)
	}
	if err := app.AssertHit(point, ByKey("disable-key-confirm-cancel")); err != nil {
		t.Fatalf("cancel button should be the top hit target: %v\n%s", err, app.RenderString())
	}

	app.ClearIntentLogs()
	if err := app.Driver().Click(ByKey("disable-key-confirm-cancel")); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitIntent("confirmdialog.cancel", 500*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled("confirmdialog.cancel"); err != nil {
		t.Fatal(err)
	}
}

func TestE2EConfirmDialogActionPageKeysDispatchByMouse(t *testing.T) {
	unregisters := make([]func(), 0, 2)
	initFn := func() {
		rt := rtui.GetGlobalIntentRuntime()
		if rt == nil {
			return
		}
		unregisters = append(unregisters,
			rt.Register("confirmdialog.confirm", intent.HandlerFunc(func(_ *intent.ActionContext, _ intent.Intent) intent.IntentResult {
				return intent.HandledResult()
			})),
			rt.Register("confirmdialog.cancel", intent.HandlerFunc(func(_ *intent.ActionContext, _ intent.Intent) intent.IntentResult {
				return intent.HandledResult()
			})),
		)
	}
	defer func() {
		for i := len(unregisters) - 1; i >= 0; i-- {
			unregisters[i]()
		}
	}()

	appFn := func() ui.VNode {
		return ui.PageStack(
			ui.NewTextBuilder("Actions Page Fixture").Build(),
			confirmdialogcomp.NewDangerOperation(
				"action.confirm",
				"Reset Group",
				"Reset the selected group runtime state.",
				"This operation affects load balancer runtime state.",
				"Reset Group",
				"actionReason",
				"maintenance",
				confirmDialogTestIntent{"confirmdialog.confirm"},
				confirmDialogTestIntent{"confirmdialog.cancel"},
				confirmdialogcomp.Target("endpoint", "Endpoint", "http://127.0.0.1:8080"),
				confirmdialogcomp.Target("group", "Group", "default"),
			).
				Width(82).
				Height(24).
				RequirePhrase("RESET", "actionConfirmText", "RESET").
				Build(),
		)
	}

	app, err := Run(appFn, ui.WithSize(100, 30), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	confirmPoint, err := app.ResolvePoint(ByKey("action.confirm-confirm"))
	if err != nil {
		t.Fatal(err)
	}
	if err := app.AssertHit(confirmPoint, ByKey("action.confirm-confirm")); err != nil {
		t.Fatalf("confirm button should be the top hit target: %v\n%s", err, app.RenderString())
	}

	app.ClearIntentLogs()
	if err := app.Driver().Click(ByKey("action.confirm-confirm")); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitIntent("confirmdialog.confirm", 500*time.Millisecond); err != nil {
		t.Fatal(err)
	}

	cancelPoint, err := app.ResolvePoint(ByKey("action.confirm-cancel"))
	if err != nil {
		t.Fatal(err)
	}
	if err := app.AssertHit(cancelPoint, ByKey("action.confirm-cancel")); err != nil {
		t.Fatalf("cancel button should be the top hit target: %v\n%s", err, app.RenderString())
	}

	app.ClearIntentLogs()
	if err := app.Driver().Click(ByKey("action.confirm-cancel")); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitIntent("confirmdialog.cancel", 500*time.Millisecond); err != nil {
		t.Fatal(err)
	}
}

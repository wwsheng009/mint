package e2e

import (
	"testing"
	"time"

	"github.com/wwsheng009/mint/runtime/intent"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
)

type formDialogTestIntent struct {
	name string
}

func (i formDialogTestIntent) IntentType() string { return i.name }

func newFormDialogStaticApp(disabled bool) ui.ComponentFunc {
	return newFormDialogStaticAppWithKey("runtime-reload-dialog", disabled, false)
}

func newFormDialogStaticAppWithTargets(disabled bool) ui.ComponentFunc {
	return newFormDialogStaticAppWithKey("runtime-reload-dialog", disabled, true)
}

func newFormDialogStaticAppWithKey(key string, disabled, targets bool) ui.ComponentFunc {
	return func() ui.VNode {
		dialog := ui.NewFormDialogDangerReasonActionBuilder(
			key,
			"Reload Runtime",
			"Reload runtime configuration with an audit reason.",
			key+"-form",
			"reason",
			"maintenance",
			"Reload",
			formDialogTestIntent{"formdialog.submit"},
			formDialogTestIntent{"formdialog.cancel"},
		).
			Width(82).
			Height(18).
			CancelText("Cancel").
			SubmitDisabled(disabled).
			DisabledReason(disabledReason(disabled))
		if targets {
			dialog.Height(22).
				Target(ui.FormDialogTarget("instance", "Instance", "gateway-a")).
				Target(ui.FormDialogSensitiveTarget("token", "Admin token", "agw_example_token"))
		}
		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("FormDialog E2E Fixture").Build(),
				dialog.Build(),
			})
	}
}

func disabledReason(disabled bool) string {
	if !disabled {
		return ""
	}
	return "Reason is required before reload."
}

func TestE2EFormDialogRendersModalForm(t *testing.T) {
	app, err := Run(newFormDialogStaticApp(false), ui.WithSize(100, 30))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{
		"FormDialog E2E Fixture",
		"Reload Runtime",
		"Reload runtime configuration",
		"Reason",
		"maintenance",
		"Reload",
		"Cancel",
	} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatalf("expected %q to be visible: %v", text, err)
		}
	}
}

func TestE2EFormDialogRendersTargetSummary(t *testing.T) {
	app, err := Run(newFormDialogStaticAppWithTargets(false), ui.WithSize(100, 30))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{
		"Instance",
		"gateway-a",
		"Admin token",
		"********",
	} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatalf("expected %q to be visible: %v", text, err)
		}
	}
	if err := app.AssertVisible(ByText("agw_example_token")); err == nil {
		t.Fatal("sensitive target value should be masked")
	}
}

func TestE2EFormDialogSubmitDispatch(t *testing.T) {
	unregisters := make([]func(), 0, 1)
	initFn := func() {
		rt := rtui.GetGlobalIntentRuntime()
		if rt == nil {
			return
		}
		unregisters = append(unregisters,
			rt.Register("formdialog.submit", intent.HandlerFunc(func(_ *intent.ActionContext, _ intent.Intent) intent.IntentResult {
				return intent.HandledResult()
			})),
		)
	}
	defer func() {
		for i := len(unregisters) - 1; i >= 0; i-- {
			unregisters[i]()
		}
	}()

	key := "runtime-reload-submit-dialog"
	app, err := Run(newFormDialogStaticAppWithKey(key, false, false), ui.WithSize(100, 30), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	app.ClearIntentLogs()
	if err := app.Driver().Click(ByKey(key + "-submit")); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitIntent("formdialog.submit", 2*time.Second); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled("formdialog.submit"); err != nil {
		t.Fatal(err)
	}
}

func TestE2EFormDialogCancelDispatch(t *testing.T) {
	unregisters := make([]func(), 0, 1)
	initFn := func() {
		rt := rtui.GetGlobalIntentRuntime()
		if rt == nil {
			return
		}
		unregisters = append(unregisters,
			rt.Register("formdialog.cancel", intent.HandlerFunc(func(_ *intent.ActionContext, _ intent.Intent) intent.IntentResult {
				return intent.HandledResult()
			})),
		)
	}
	defer func() {
		for i := len(unregisters) - 1; i >= 0; i-- {
			unregisters[i]()
		}
	}()

	key := "runtime-reload-cancel-dialog"
	app, err := Run(newFormDialogStaticAppWithKey(key, false, false), ui.WithSize(100, 30), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	point, err := app.ResolvePoint(ByKey(key + "-cancel"))
	if err != nil {
		t.Fatal(err)
	}
	if err := app.AssertHit(point, ByKey(key+"-cancel")); err != nil {
		t.Fatalf("cancel button should be the top hit target: %v\n%s", err, app.RenderString())
	}

	app.ClearIntentLogs()
	if err := app.Driver().Click(ByKey(key + "-cancel")); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitIntent("formdialog.cancel", 2*time.Second); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled("formdialog.cancel"); err != nil {
		t.Fatal(err)
	}
}

func TestE2EFormDialogDisabledReason(t *testing.T) {
	app, err := Run(newFormDialogStaticApp(true), ui.WithSize(100, 30))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("Reason is required before reload.")); err != nil {
		t.Fatal(err)
	}
}

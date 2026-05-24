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
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("FormDialog E2E Fixture").Build(),
				ui.NewFormDialogBuilder().
					Key("runtime-reload-dialog").
					Title("Reload Runtime").
					Description("Reload runtime configuration with an audit reason.").
					FormID("runtime-reload-form").
					Opened().
					Width(82).
					Height(18).
					Children(
						ui.FormInputItem(
							"reason",
							"Reason",
							"maintenance",
							ui.FormInputForForm("runtime-reload-form"),
							ui.FormInputWidth(52),
							ui.FormInputValidators(ui.Required()),
						),
					).
					SubmitText("Reload").
					CancelText("Cancel").
					SubmitVariant(ui.ButtonVariantDanger).
					SubmitDisabled(disabled).
					DisabledReason(disabledReason(disabled)).
					OnSubmit(formDialogTestIntent{"formdialog.submit"}).
					OnCancel(formDialogTestIntent{"formdialog.cancel"}).
					Build(),
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

	app, err := Run(newFormDialogStaticApp(false), ui.WithSize(100, 30), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	app.ClearIntentLogs()
	if err := app.Driver().Click(ByID("runtime-reload-dialog-submit")); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitIntent("formdialog.submit", 500*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled("formdialog.submit"); err != nil {
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

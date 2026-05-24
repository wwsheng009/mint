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
					Height(20).
					Build(),
			})
	}
}

func TestE2EConfirmDialogRendersAndMasksSensitiveTarget(t *testing.T) {
	app, err := Run(newConfirmDialogStaticApp("maintenance"), ui.WithSize(100, 28))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{"Disable Provider Key", "Disable the selected provider key", "Provider", "openai", "Reason *", "maintenance", "Disable", "Cancel"} {
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

	app, err := Run(newConfirmDialogStaticApp("maintenance"), ui.WithSize(100, 28), ui.WithInit(initFn))
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

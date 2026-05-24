package e2e

import (
	"fmt"
	"testing"
	"time"

	"github.com/wwsheng009/mint/runtime/intent"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
	toolbarcomp "github.com/wwsheng009/mint/ui/components/toolbar"
)

type toolbarTestIntent struct {
	name string
}

func (i toolbarTestIntent) IntentType() string { return i.name }

func newToolbarStaticApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Toolbar E2E Fixture").Build(),
				toolbarcomp.NewBuilder().
					Key("ops-toolbar").
					Title("Load Balancer").
					TitleWidth(16).
					Width(88).
					Dense(true).
					Left(toolbarcomp.Text("scope", "group: default")).
					Center(toolbarcomp.Badge("state", "degraded").WithColors("black", "yellow")).
					Right(toolbarcomp.Button("refresh", "Refresh", toolbarTestIntent{"toolbar.refresh"}).Primary()).
					Right(toolbarcomp.Button("reset", "Reset Runtime", toolbarTestIntent{"toolbar.reset"}).Danger().WithDisabled(true)).
					Build(),
				ui.DataTable(
					[]ui.TableColumn{
						{Title: "Provider", Width: 18},
						{Title: "Status", Width: 12},
					},
					[][]string{
						{"openai", "degraded"},
						{"azure", "healthy"},
					},
					ui.DataTablePageSize(5),
					ui.DataTableOperationalStyle(),
				),
			})
	}
}

func TestE2EToolbarRendersAboveTable(t *testing.T) {
	app, err := Run(newToolbarStaticApp(), ui.WithSize(96, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{"Toolbar E2E Fixture", "Load Balancer", "group: default", "degraded", "Refresh", "Reset Runtime", "Provider"} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatalf("expected %q to be visible: %v", text, err)
		}
	}

	toolbarPoint, err := app.ResolvePoint(ByText("Load Balancer"))
	if err != nil {
		t.Fatal(err)
	}
	tablePoint, err := app.ResolvePoint(ByText("Provider"))
	if err != nil {
		t.Fatal(err)
	}
	if toolbarPoint.Y >= tablePoint.Y {
		t.Fatalf("expected toolbar above table, got toolbarY=%d tableY=%d", toolbarPoint.Y, tablePoint.Y)
	}
}

func TestE2EToolbarActionDispatch(t *testing.T) {
	unregisters := make([]func(), 0, 1)
	initFn := func() {
		rt := rtui.GetGlobalIntentRuntime()
		if rt == nil {
			return
		}
		unregisters = append(unregisters,
			rt.Register("toolbar.refresh", intent.HandlerFunc(func(_ *intent.ActionContext, _ intent.Intent) intent.IntentResult {
				return intent.HandledResult()
			})),
		)
	}
	defer func() {
		for i := len(unregisters) - 1; i >= 0; i-- {
			unregisters[i]()
		}
	}()

	app, err := Run(newToolbarStaticApp(), ui.WithSize(96, 24), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	app.ClearIntentLogs()
	if err := app.Driver().Click(ByKey("ops-toolbar-right-refresh")); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitIntent("toolbar.refresh", 500*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled("toolbar.refresh"); err != nil {
		t.Fatal(err)
	}
}

func newToolbarStatusBarApp() ui.ComponentFunc {
	return func() ui.VNode {
		children := []ui.VNode{
			ui.NewTextBuilder("Toolbar StatusBar E2E Fixture").Build(),
			ui.NewTextBuilder("Workspace: operations").Build(),
			ui.NewTextBuilder("Runtime: idle").Build(),
			toolbarcomp.NewBuilder().
				Key("ops-status").
				Title("ai-gateway-manager").
				UseStatusBar(true).
				Left(toolbarcomp.Badge("mode", "ADMIN").WithHelp("Admin operations mode")).
				Center(toolbarcomp.Text("focus", "F2 Load Balancer")).
				Right(toolbarcomp.Button("help", "F10 Help", toolbarTestIntent{"toolbar.help"}).WithHelp("Open contextual help")).
				Build(),
		}
		return ui.NewVStack().SetGap(0).SetChildrenList(children)
	}
}

func TestE2EToolbarStatusBarHelpOverlay(t *testing.T) {
	app, err := Run(newToolbarStatusBarApp(), ui.WithSize(80, 12))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("ai-gateway-manager")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("ADMIN")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Admin operations mode")); err == nil {
		t.Fatal("overlay help should be hidden before hover")
	}

	if err := app.Driver().Move(ByKey("ops-status-left-mode")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("Admin operations mode")); err != nil {
			return fmt.Errorf("overlay help not visible: %w", err)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

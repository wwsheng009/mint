package e2e

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/wwsheng009/mint/framework"
	runtimeintent "github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/runtime/store"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
	selectcomp "github.com/wwsheng009/mint/ui/components/select"
)

type selectFixtureState struct {
	ChangeCount      int
	BackgroundClicks int
}

var selectFixtureSeq int64

type selectChangeIntent struct{ Token string }
type selectBackgroundIntent struct{ Token string }

func (i selectChangeIntent) IntentType() string     { return "e2e.select_change." + i.Token }
func (i selectBackgroundIntent) IntentType() string { return "e2e.select_background." + i.Token }

type selectFixtureMeta struct {
	ChangeIntentType     string
	BackgroundIntentType string
}

func newSelectFixture() (ui.ComponentFunc, func(), *store.Store[selectFixtureState], selectFixtureMeta) {
	fixtureStore := store.NewStore(selectFixtureState{})
	token := fmt.Sprintf("%d", atomic.AddInt64(&selectFixtureSeq, 1))
	changeIntent := selectChangeIntent{Token: token}
	backgroundIntent := selectBackgroundIntent{Token: token}
	meta := selectFixtureMeta{
		ChangeIntentType:     changeIntent.IntentType(),
		BackgroundIntentType: backgroundIntent.IntentType(),
	}

	initFn := func() {
		rt := rtui.GetGlobalIntentRuntime()
		if rt == nil {
			return
		}
		rt.Register(changeIntent.IntentType(), runtimeintent.HandlerFunc(func(_ *runtimeintent.ActionContext, _ runtimeintent.Intent) runtimeintent.IntentResult {
			fixtureStore.Update(func(s selectFixtureState) selectFixtureState {
				s.ChangeCount++
				return s
			})
			return runtimeintent.HandledResult()
		}))
		rt.Register(backgroundIntent.IntentType(), runtimeintent.HandlerFunc(func(_ *runtimeintent.ActionContext, _ runtimeintent.Intent) runtimeintent.IntentResult {
			fixtureStore.Update(func(s selectFixtureState) selectFixtureState {
				s.BackgroundClicks++
				return s
			})
			return runtimeintent.HandledResult()
		}))
	}

	appFn := func() ui.VNode {
		changeCount := ui.UseStoreSelector(fixtureStore, func(s selectFixtureState) int { return s.ChangeCount })
		backgroundClicks := ui.UseStoreSelector(fixtureStore, func(s selectFixtureState) int { return s.BackgroundClicks })

		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Select E2E Fixture").Build(),
				ui.HStackBuilder(
					selectcomp.NewBuilder().
						SetID("country-select").
						OverlayPopup(true).
						CloseOnOutside(true).
						FilterOption(true).
						FilterPlaceholder("type to filter").
						Width(24).
						Selected(0).
						Options([]selectcomp.Option{
							{Value: "us", Label: "United States"},
							{Value: "cn", Label: "China"},
							{Value: "jp", Label: "Japan"},
						}).
						OnChange(changeIntent).
						Build(),
					ui.NewButtonBuilder("Background Action").SetID("background-btn").OnPress(backgroundIntent).Build(),
				).Gap(4).Build(),
				ui.NewTextBuilder(fmt.Sprintf("ChangeCount: %d", changeCount)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("BackgroundClicks: %d", backgroundClicks)).Build(),
			})
	}

	return appFn, initFn, fixtureStore, meta
}

func TestE2ESelectOverlayFilterCommitFlow(t *testing.T) {
	appFn, initFn, fixtureStore, meta := newSelectFixture()
	app, err := Run(appFn,
		ui.WithSize(90, 24),
		ui.WithInit(initFn),
		ui.WithPluginSetup(func(app *framework.App) {
			selectcomp.Install(app)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertFocus(ByID("country-select")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("United States")); err != nil {
		t.Fatal(err)
	}

	if err := app.Driver().Special(platform.KeyEnter); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByID("country-select-popup"))
	}); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	if err := app.Driver().Type("ja"); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("Japan"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("Japan"), StyleExpect{
		HasBold: true,
		Bold:    true,
	}); err != nil {
		t.Fatal(err)
	}

	point, err := app.ResolvePoint(ByText("Japan"))
	if err != nil {
		t.Fatal(err)
	}
	if err := app.AssertHit(point, ByID("country-select-popup")); err != nil {
		t.Fatal(err)
	}

	if err := app.Driver().Click(ByText("Japan")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "key:j"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "mouse:left:release"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertLastIntent(meta.ChangeIntentType); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.ChangeIntentType); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByID("country-select-popup")); err == nil {
			return fmt.Errorf("select popup still visible")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("ChangeCount: 1")); err != nil {
		t.Fatal(err)
	}

	state := fixtureStore.Get()
	if state.ChangeCount != 1 || state.BackgroundClicks != 0 {
		t.Fatalf("unexpected select fixture state after commit: %+v", state)
	}
}

func TestE2ESelectOverlayOutsideClickClosesWithoutBackgroundLeak(t *testing.T) {
	appFn, initFn, fixtureStore, meta := newSelectFixture()
	app, err := Run(appFn,
		ui.WithSize(90, 24),
		ui.WithInit(initFn),
		ui.WithPluginSetup(func(app *framework.App) {
			selectcomp.Install(app)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.Driver().Click(ByID("country-select")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByID("country-select-popup"))
	}); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	if err := app.Driver().Click(ByID("background-btn")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByID("country-select-popup")); err == nil {
			return fmt.Errorf("select popup still visible")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	for _, logEntry := range app.IntentLogs() {
		if logEntry.Type == meta.BackgroundIntentType {
			t.Fatalf("background intent leaked through overlay close: %+v", logEntry)
		}
	}

	state := fixtureStore.Get()
	if state.ChangeCount != 0 || state.BackgroundClicks != 0 {
		t.Fatalf("unexpected select fixture state after outside click: %+v", state)
	}
}

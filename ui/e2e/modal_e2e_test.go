package e2e

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/wwsheng009/mint/framework"
	runtimeintent "github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/store"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/ui/components/modal"
)

type modalFixtureState struct {
	Open             bool
	BackgroundClicks int
	ClosedCount      int
	ConfirmedCount   int
}

var modalFixtureSeq int64

type modalOpenIntent struct{ Token string }
type modalCloseIntent struct{ Token string }
type modalConfirmIntent struct{ Token string }
type modalBackgroundIntent struct{ Token string }

func (i modalOpenIntent) IntentType() string       { return "e2e.modal_open." + i.Token }
func (i modalCloseIntent) IntentType() string      { return "e2e.modal_close." + i.Token }
func (i modalConfirmIntent) IntentType() string    { return "e2e.modal_confirm." + i.Token }
func (i modalBackgroundIntent) IntentType() string { return "e2e.modal_background." + i.Token }

type modalFixtureMeta struct {
	OpenIntentType    string
	CloseIntentType   string
	ConfirmIntentType string
}

func newModalFixture() (ui.ComponentFunc, func(), *store.Store[modalFixtureState], modalFixtureMeta) {
	modalStore := store.NewStore(modalFixtureState{})
	token := fmt.Sprintf("%d", atomic.AddInt64(&modalFixtureSeq, 1))
	openIntent := modalOpenIntent{Token: token}
	closeIntent := modalCloseIntent{Token: token}
	confirmIntent := modalConfirmIntent{Token: token}
	backgroundIntent := modalBackgroundIntent{Token: token}
	meta := modalFixtureMeta{
		OpenIntentType:    openIntent.IntentType(),
		CloseIntentType:   closeIntent.IntentType(),
		ConfirmIntentType: confirmIntent.IntentType(),
	}

	initFn := func() {
		rt := rtui.GetGlobalIntentRuntime()
		if rt == nil {
			return
		}
		rt.Register(openIntent.IntentType(), runtimeintent.HandlerFunc(func(_ *runtimeintent.ActionContext, _ runtimeintent.Intent) runtimeintent.IntentResult {
			modalStore.Update(func(s modalFixtureState) modalFixtureState {
				s.Open = true
				return s
			})
			return runtimeintent.HandledResult()
		}))
		rt.Register(closeIntent.IntentType(), runtimeintent.HandlerFunc(func(_ *runtimeintent.ActionContext, _ runtimeintent.Intent) runtimeintent.IntentResult {
			modalStore.Update(func(s modalFixtureState) modalFixtureState {
				s.Open = false
				s.ClosedCount++
				return s
			})
			return runtimeintent.HandledResult()
		}))
		rt.Register(confirmIntent.IntentType(), runtimeintent.HandlerFunc(func(_ *runtimeintent.ActionContext, _ runtimeintent.Intent) runtimeintent.IntentResult {
			modalStore.Update(func(s modalFixtureState) modalFixtureState {
				s.Open = false
				s.ClosedCount++
				s.ConfirmedCount++
				return s
			})
			return runtimeintent.HandledResult()
		}))
		rt.Register(backgroundIntent.IntentType(), runtimeintent.HandlerFunc(func(_ *runtimeintent.ActionContext, _ runtimeintent.Intent) runtimeintent.IntentResult {
			modalStore.Update(func(s modalFixtureState) modalFixtureState {
				s.BackgroundClicks++
				return s
			})
			return runtimeintent.HandledResult()
		}))
	}

	appFn := func() ui.VNode {
		open := ui.UseStoreSelector(modalStore, func(s modalFixtureState) bool { return s.Open })
		backgroundClicks := ui.UseStoreSelector(modalStore, func(s modalFixtureState) int { return s.BackgroundClicks })
		closedCount := ui.UseStoreSelector(modalStore, func(s modalFixtureState) int { return s.ClosedCount })
		confirmedCount := ui.UseStoreSelector(modalStore, func(s modalFixtureState) int { return s.ConfirmedCount })

		children := []ui.VNode{
			ui.NewButtonBuilder("Open Modal").SetID("open-btn").OnPress(openIntent).Build(),
			ui.NewButtonBuilder("Background Action").SetID("background-btn").OnPress(backgroundIntent).Build(),
			ui.NewTextBuilder(fmt.Sprintf("BackgroundClicks: %d", backgroundClicks)).Build(),
			ui.NewTextBuilder(fmt.Sprintf("ClosedCount: %d", closedCount)).Build(),
			ui.NewTextBuilder(fmt.Sprintf("ConfirmedCount: %d", confirmedCount)).Build(),
		}

		if open {
			modalContent := ui.NewVStack().
				SetGap(1).
				SetChildrenList([]ui.VNode{
					ui.NewTextBuilder("Fixture Modal Body").Build(),
					ui.HStackBuilder(
						ui.NewButtonBuilder("Cancel").SetID("cancel-btn").OnPress(closeIntent).Build(),
						ui.NewButtonBuilder("Confirm").SetID("confirm-btn").OnPress(confirmIntent).Build(),
					).Gap(1).Build(),
				})
			children = append(children, modal.NewBuilder().
				SetID("fixture-modal").
				Title("Fixture Modal").
				Content(modalContent).
				Open(true).
				Width(36).
				Height(10).
				CloseOnEsc(true).
				CloseOnBackdrop(true).
				OnClose(closeIntent).
				Build())
		}

		return ui.NewVStack().
			SetGap(1).
			SetChildrenList(children)
	}

	return appFn, initFn, modalStore, meta
}

func TestE2EModalOverlayHitAndConfirmFlow(t *testing.T) {
	appFn, initFn, modalStore, meta := newModalFixture()
	app, err := Run(appFn,
		ui.WithSize(80, 24),
		ui.WithInit(initFn),
		ui.WithPluginSetup(func(app *framework.App) {
			app.AddMiddleware(modal.NewModalMiddleware())
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertFocus(ByID("open-btn")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	if err := app.Driver().Click(ByID("open-btn")); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitTrace(TraceMatch{Kind: TraceIntentDispatch, Name: meta.OpenIntentType}, 500*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("Fixture Modal Body"))
	}); err != nil {
		t.Fatal(err)
	}

	confirmFiber, err := app.ResolveFiber(ByID("confirm-btn"))
	if err != nil {
		t.Fatal(err)
	}
	confirmPoint, err := app.ResolvePoint(ByID("confirm-btn"))
	if err != nil {
		t.Fatal(err)
	}
	if err := app.AssertHit(confirmPoint, ByID("confirm-btn")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTargetID(ByID("confirm-btn"), confirmFiber.ActionTargetID); err != nil {
		t.Fatal(err)
	}

	if err := app.Driver().Click(ByID("confirm-btn")); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitIntent(meta.ConfirmIntentType, 500*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("Fixture Modal Body")); err == nil {
			return fmt.Errorf("modal body still visible")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	state := modalStore.Get()
	if state.ConfirmedCount != 1 || state.ClosedCount != 1 || state.BackgroundClicks != 0 {
		t.Fatalf("unexpected modal state after confirm: %+v", state)
	}
}

func TestE2EModalBackdropClickClosesWithoutBackgroundLeak(t *testing.T) {
	appFn, initFn, modalStore, meta := newModalFixture()
	app, err := Run(appFn,
		ui.WithSize(80, 24),
		ui.WithInit(initFn),
		ui.WithPluginSetup(func(app *framework.App) {
			app.AddMiddleware(modal.NewModalMiddleware())
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.Driver().Click(ByID("open-btn")); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitIntent(meta.OpenIntentType, 500*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("Fixture Modal Body"))
	}); err != nil {
		t.Fatal(err)
	}

	if err := app.Driver().ClickAt(0, 0); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitIntent(meta.CloseIntentType, 500*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("Fixture Modal Body")); err == nil {
			return fmt.Errorf("modal body still visible")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	state := modalStore.Get()
	if state.ClosedCount != 1 {
		t.Fatalf("closedCount = %d, want 1", state.ClosedCount)
	}
	if state.BackgroundClicks != 0 {
		t.Fatalf("backgroundClicks = %d, want 0", state.BackgroundClicks)
	}
	if err := app.AssertIntentHandled(meta.CloseIntentType); err != nil {
		t.Fatal(err)
	}
}

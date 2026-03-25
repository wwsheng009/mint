package e2e

import (
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/wwsheng009/mint/framework"
	runtimeintent "github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/store"
	rttypes "github.com/wwsheng009/mint/runtime/types"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
	menucomp "github.com/wwsheng009/mint/ui/components/menu"
)

type anchoredMenuActivateIntent struct{ Token string }

func (i anchoredMenuActivateIntent) IntentType() string {
	return "e2e.menu.anchored_activate." + i.Token
}

type contextMenuBackgroundIntent struct{ Token string }

func (i contextMenuBackgroundIntent) IntentType() string {
	return "e2e.menu.context_background." + i.Token
}

type contextMenuFixtureState struct {
	BackgroundClicks int
	MenuOpen         bool
}

type anchoredMenuFixtureMeta struct {
	ActivateIntentType string
}

type contextMenuFixtureMeta struct {
	BackgroundIntentType string
}

var menuFixtureSeq int64

func newAnchoredMenuFixture() (ui.ComponentFunc, func(), func(), anchoredMenuFixtureMeta) {
	token := fmt.Sprintf("%d", atomic.AddInt64(&menuFixtureSeq, 1))
	activateIntent := anchoredMenuActivateIntent{Token: token}
	meta := anchoredMenuFixtureMeta{ActivateIntentType: activateIntent.IntentType()}

	unregisters := make([]func(), 0, 1)
	initFn := func() {
		rt := rtui.GetGlobalIntentRuntime()
		if rt == nil {
			return
		}
		rt.Register(meta.ActivateIntentType, runtimeintent.HandlerFunc(func(_ *runtimeintent.ActionContext, _ runtimeintent.Intent) runtimeintent.IntentResult {
			return runtimeintent.HandledResult()
		}))
	}

	cleanupFn := func() {
		for i := len(unregisters) - 1; i >= 0; i-- {
			if unregisters[i] != nil {
				unregisters[i]()
			}
		}
	}

	appFn := func() ui.VNode {
		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Menu E2E Anchored Fixture").Build(),
				ui.NewHStack().
					SetGap(0).
					SetChildrenList([]ui.VNode{
						ui.NewTextBuilder(strings.Repeat(" ", 24)).Build(),
						ui.NewTextBuilder("Menu Anchor").SetID("menu-anchor").Build(),
					}),
				menucomp.NewPopup([]menucomp.MenuItem{
					menucomp.Action("anchored-action", "Anchored Action", activateIntent),
				}).
					SetID("fixture-anchored-menu").
					AnchorTo("menu-anchor", rttypes.AnchorBottomLeft).
					Placement(menucomp.PlacementBottomEnd).
					Build(),
				ui.NewTextBuilder("Anchored menu footer").Build(),
			})
	}

	return appFn, initFn, cleanupFn, meta
}

func newContextMenuFixture() (ui.ComponentFunc, func(), func(), *store.Store[contextMenuFixtureState], contextMenuFixtureMeta) {
	fixtureStore := store.NewStore(contextMenuFixtureState{MenuOpen: true})
	token := fmt.Sprintf("%d", atomic.AddInt64(&menuFixtureSeq, 1))
	backgroundIntent := contextMenuBackgroundIntent{Token: token}
	meta := contextMenuFixtureMeta{BackgroundIntentType: backgroundIntent.IntentType()}

	unregisters := make([]func(), 0, 2)
	initFn := func() {
		rt := rtui.GetGlobalIntentRuntime()
		if rt == nil {
			return
		}
		unregisters = append(unregisters,
			rt.Register(meta.BackgroundIntentType, runtimeintent.HandlerFunc(func(_ *runtimeintent.ActionContext, _ runtimeintent.Intent) runtimeintent.IntentResult {
				fixtureStore.Update(func(s contextMenuFixtureState) contextMenuFixtureState {
					s.BackgroundClicks++
					return s
				})
				return runtimeintent.HandledResult()
			})),
			runtimeintent.RegisterTypedRuntime(rt, func(_ *runtimeintent.ActionContext, i menucomp.CloseMenuIntent) runtimeintent.IntentResult {
				if i.MenuID != "fixture-context-menu" {
					return runtimeintent.IntentResult{}
				}
				fixtureStore.Update(func(s contextMenuFixtureState) contextMenuFixtureState {
					s.MenuOpen = false
					return s
				})
				return runtimeintent.HandledResult()
			}),
		)
	}

	cleanupFn := func() {
		for i := len(unregisters) - 1; i >= 0; i-- {
			if unregisters[i] != nil {
				unregisters[i]()
			}
		}
	}

	appFn := func() ui.VNode {
		backgroundClicks := ui.UseStoreSelector(fixtureStore, func(s contextMenuFixtureState) int { return s.BackgroundClicks })
		menuOpen := ui.UseStoreSelector(fixtureStore, func(s contextMenuFixtureState) bool { return s.MenuOpen })

		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Menu E2E Context Fixture").Build(),
				ui.NewTextBuilder("BackgroundClicks: " + fmt.Sprintf("%d", backgroundClicks)).Build(),
				ui.NewButtonBuilder("Background Action").SetID("menu-background-btn").OnPress(backgroundIntent).Build(),
				ui.NewTextBuilder("Context menu should close before background click leaks").Build(),
				menucomp.NewContextMenu([]menucomp.MenuItem{
					menucomp.Action("context-action", "Context Action", nil),
				}).
					SetID("fixture-context-menu").
					Open(menuOpen).
					PortalOffset(18, 2).
					Build(),
			})
	}

	return appFn, initFn, cleanupFn, fixtureStore, meta
}

func TestE2EMenuAnchoredPopupPlacementAndActivation(t *testing.T) {
	appFn, initFn, cleanupFn, meta := newAnchoredMenuFixture()
	defer cleanupFn()

	app, err := Run(appFn,
		ui.WithSize(70, 18),
		ui.WithInit(initFn),
		ui.WithPluginSetup(func(app *framework.App) {
			menucomp.Install(app, nil)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("Anchored Action")); err != nil {
		t.Fatal(err)
	}

	anchorBounds, err := app.BoundsOf(ByID("menu-anchor"))
	if err != nil {
		t.Fatal(err)
	}
	popupBounds, err := app.BoundsOf(ByID("fixture-anchored-menu"))
	if err != nil {
		t.Fatal(err)
	}

	if popupBounds.X+popupBounds.Width-1 != anchorBounds.X+anchorBounds.Width {
		t.Fatalf("popup surface right edge = %d, want %d", popupBounds.X+popupBounds.Width-1, anchorBounds.X+anchorBounds.Width)
	}
	if popupBounds.Y != anchorBounds.Y+anchorBounds.Height {
		t.Fatalf("popup top = %d, want %d", popupBounds.Y, anchorBounds.Y+anchorBounds.Height)
	}

	point, err := app.ResolvePoint(ByText("Anchored Action"))
	if err != nil {
		t.Fatal(err)
	}
	if err := app.AssertHit(point, ByID("fixture-anchored-menu")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	if err := app.Driver().Click(ByText("Anchored Action")); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitIntent(meta.ActivateIntentType, 500*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.ActivateIntentType); err != nil {
		t.Fatal(err)
	}
}

func TestE2EMenuContextOutsideClickClosesWithoutBackgroundLeak(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newContextMenuFixture()
	defer cleanupFn()

	app, err := Run(appFn,
		ui.WithSize(60, 18),
		ui.WithInit(initFn),
		ui.WithPluginSetup(func(app *framework.App) {
			menucomp.Install(app, nil)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("Context Action")); err != nil {
		t.Fatal(err)
	}

	bounds, err := app.BoundsOf(ByID("fixture-context-menu"))
	if err != nil {
		t.Fatal(err)
	}
	if bounds.X != 18 || bounds.Y != 2 {
		t.Fatalf("context menu bounds = %v, want origin at (18,2)", bounds)
	}

	app.ClearIntentLogs()
	if err := app.Driver().Click(ByID("menu-background-btn")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByID("fixture-context-menu")); err == nil {
			return fmt.Errorf("context menu still visible")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	for _, logEntry := range app.IntentLogs() {
		if logEntry.Type == meta.BackgroundIntentType {
			t.Fatalf("background intent leaked through menu middleware: %+v", logEntry)
		}
	}

	state := fixtureStore.Get()
	if state.BackgroundClicks != 0 {
		t.Fatalf("background click count = %d, want 0", state.BackgroundClicks)
	}
}

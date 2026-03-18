package e2e

import (
	"fmt"
	"testing"
	"time"

	fwtheme "github.com/wwsheng009/mint/framework/theme"
	runtimeintent "github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/store"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
)

type skeletonFixtureState struct {
	Loading bool
}

type toggleSkeletonFixtureIntent struct{}

func (toggleSkeletonFixtureIntent) IntentType() string { return "e2e.toggle_skeleton_fixture" }

func newSkeletonStaticApp() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Skeleton Static E2E Fixture").Build(),
				ui.NewSkeletonBuilder().
					SetID("fixture-skeleton-static").
					Avatar(true).
					AvatarSize(3).
					Width(26).
					TitleWidth(12).
					ParagraphRows(2).
					ParagraphWidths(18, 10).
					Build(),
			})
	}
}

func newSkeletonToggleFixture() (ui.ComponentFunc, func(), func(), *store.Store[skeletonFixtureState]) {
	fixtureStore := store.NewStore(skeletonFixtureState{Loading: true})
	unregisters := make([]func(), 0, 1)

	initFn := func() {
		rt := rtui.GetGlobalIntentRuntime()
		if rt == nil {
			return
		}
		unregisters = append(unregisters,
			runtimeintent.RegisterTypedRuntime(rt, func(_ *runtimeintent.ActionContext, _ toggleSkeletonFixtureIntent) runtimeintent.IntentResult {
				fixtureStore.Update(func(s skeletonFixtureState) skeletonFixtureState {
					s.Loading = !s.Loading
					return s
				})
				return runtimeintent.HandledResult()
			}),
		)
	}

	cleanupFn := func() {
		for index := len(unregisters) - 1; index >= 0; index-- {
			if unregisters[index] != nil {
				unregisters[index]()
			}
		}
	}

	appFn := func() ui.VNode {
		loading := ui.UseStoreSelector(fixtureStore, func(s skeletonFixtureState) bool { return s.Loading })
		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder(fmt.Sprintf("SkeletonLoading: %t", loading)).Build(),
				ui.NewButtonBuilder("Toggle Skeleton").SetID("toggle-skeleton").OnPress(toggleSkeletonFixtureIntent{}).Build(),
				ui.NewSkeletonBuilder().
					SetID("fixture-skeleton-toggle").
					Loading(loading).
					Avatar(true).
					AvatarShape(ui.SkeletonShapeRound).
					AvatarSize(4).
					TitleWidth(14).
					ParagraphRows(2).
					ParagraphWidths(20, 12).
					Content(
						ui.NewTextBuilder("Loaded profile content").SetID("loaded-profile-content").Build(),
					).
					Build(),
			})
	}

	return appFn, initFn, cleanupFn, fixtureStore
}

func TestE2ESkeletonStaticRenderAndStyles(t *testing.T) {
	app, err := Run(newSkeletonStaticApp(), ui.WithSize(90, 24))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("▓▓▓")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("▓▓▓▓▓▓▓▓▓▓▓▓")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("▓▓▓▓▓▓▓▓▓▓")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText("▓▓▓▓▓▓▓▓▓▓▓▓"), StyleExpect{
		HasFG: true,
		FG:    fwtheme.Muted(),
		HasBG: true,
		BG:    fwtheme.Surface(),
	}); err != nil {
		t.Fatal(err)
	}
}

func TestE2ESkeletonLoadingGateFlow(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore := newSkeletonToggleFixture()
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(90, 24), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("SkeletonLoading: true")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("▓▓▓▓▓▓▓▓▓▓▓▓▓▓")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Loaded profile content")); err == nil {
		t.Fatal("loaded content should be hidden while loading")
	}

	app.ClearIntentLogs()
	if err := app.Driver().Click(ByID("toggle-skeleton")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("Loaded profile content"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("SkeletonLoading: false")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		if err := app.AssertVisible(ByText("▓▓▓▓▓▓▓▓▓▓▓▓▓▓")); err == nil {
			return fmt.Errorf("skeleton title placeholder still visible")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	state := fixtureStore.Get()
	if state.Loading {
		t.Fatalf("expected loading false after first toggle: %+v", state)
	}

	app.ClearIntentLogs()
	if err := app.Driver().Click(ByID("toggle-skeleton")); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("SkeletonLoading: true"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("▓▓▓▓▓▓▓▓▓▓▓▓▓▓"))
	}); err != nil {
		t.Fatal(err)
	}

	state = fixtureStore.Get()
	if !state.Loading {
		t.Fatalf("expected loading true after second toggle: %+v", state)
	}
}

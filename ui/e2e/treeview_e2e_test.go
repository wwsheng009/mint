package e2e

import (
	"fmt"
	"testing"
	"time"

	runtimeintent "github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
	inputcomp "github.com/wwsheng009/mint/ui/components/input"
	treeviewcomp "github.com/wwsheng009/mint/ui/components/treeview"
)

type treeViewFixtureState struct {
	Query           string
	CheckedPaths    string
	SearchTotal     int
	SearchSelected  int
	SearchPageCount int
	LazyLoads       int
	SelectionEvents int
	Reorders        int
	LastReordered   string
}

type treeViewFixtureMeta struct {
	TreeComponentID           string
	QueryField                string
	CheckedPathsField         string
	FieldChangeIntentType     string
	SearchResultsIntentType   string
	LazyLoadIntentType        string
	SelectionChangeIntentType string
	NodeReorderIntentType     string
}

func newTreeViewFixture() (ui.ComponentFunc, func(), func(), *store.Store[treeViewFixtureState], treeViewFixtureMeta) {
	fixtureStore := store.NewStore(treeViewFixtureState{})
	meta := treeViewFixtureMeta{
		TreeComponentID:           "fixture.treeview",
		QueryField:                "tree-query",
		CheckedPathsField:         "tree-checked-paths",
		FieldChangeIntentType:     runtimeintent.FieldChangeIntent{}.IntentType(),
		SearchResultsIntentType:   treeviewcomp.SearchResultsIntent{}.IntentType(),
		LazyLoadIntentType:        treeviewcomp.LazyLoadIntent{}.IntentType(),
		SelectionChangeIntentType: treeviewcomp.SelectionChangeIntent{}.IntentType(),
		NodeReorderIntentType:     treeviewcomp.NodeReorderIntent{}.IntentType(),
	}

	unregisters := make([]func(), 0, 5)
	initFn := func() {
		rt := rtui.GetGlobalIntentRuntime()
		if rt == nil {
			return
		}

		unregisters = append(unregisters,
			runtimeintent.RegisterTypedRuntime(rt, func(_ *runtimeintent.ActionContext, i runtimeintent.FieldChangeIntent) runtimeintent.IntentResult {
				if i.Field != meta.QueryField && i.Field != meta.CheckedPathsField {
					return runtimeintent.IntentResult{}
				}
				fixtureStore.Update(func(s treeViewFixtureState) treeViewFixtureState {
					switch i.Field {
					case meta.QueryField:
						s.Query = i.Value
					case meta.CheckedPathsField:
						s.CheckedPaths = i.Value
					}
					return s
				})
				return runtimeintent.HandledResult()
			}),
			runtimeintent.RegisterTypedRuntime(rt, func(_ *runtimeintent.ActionContext, i treeviewcomp.SearchResultsIntent) runtimeintent.IntentResult {
				if i.ComponentID != meta.TreeComponentID {
					return runtimeintent.IntentResult{}
				}
				fixtureStore.Update(func(s treeViewFixtureState) treeViewFixtureState {
					s.SearchTotal = i.Total
					s.SearchSelected = i.Selected
					s.SearchPageCount = i.PageCount
					return s
				})
				return runtimeintent.HandledResult()
			}),
			runtimeintent.RegisterTypedRuntime(rt, func(_ *runtimeintent.ActionContext, i treeviewcomp.LazyLoadIntent) runtimeintent.IntentResult {
				if i.ComponentID != meta.TreeComponentID {
					return runtimeintent.IntentResult{}
				}
				fixtureStore.Update(func(s treeViewFixtureState) treeViewFixtureState {
					s.LazyLoads++
					return s
				})
				return runtimeintent.HandledResult()
			}),
			runtimeintent.RegisterTypedRuntime(rt, func(_ *runtimeintent.ActionContext, i treeviewcomp.SelectionChangeIntent) runtimeintent.IntentResult {
				if i.ComponentID != meta.TreeComponentID {
					return runtimeintent.IntentResult{}
				}
				fixtureStore.Update(func(s treeViewFixtureState) treeViewFixtureState {
					s.SelectionEvents++
					return s
				})
				return runtimeintent.HandledResult()
			}),
			runtimeintent.RegisterTypedRuntime(rt, func(_ *runtimeintent.ActionContext, i treeviewcomp.NodeReorderIntent) runtimeintent.IntentResult {
				if i.ComponentID != meta.TreeComponentID {
					return runtimeintent.IntentResult{}
				}
				fixtureStore.Update(func(s treeViewFixtureState) treeViewFixtureState {
					s.Reorders++
					s.LastReordered = i.Path
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
		query := ui.UseStoreSelector(fixtureStore, func(s treeViewFixtureState) string { return s.Query })
		lazyLoads := ui.UseStoreSelector(fixtureStore, func(s treeViewFixtureState) int { return s.LazyLoads })
		selectionEvents := ui.UseStoreSelector(fixtureStore, func(s treeViewFixtureState) int { return s.SelectionEvents })
		reorders := ui.UseStoreSelector(fixtureStore, func(s treeViewFixtureState) int { return s.Reorders })
		lastReordered := ui.UseStoreSelector(fixtureStore, func(s treeViewFixtureState) string { return s.LastReordered })

		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("TreeView E2E Fixture").Build(),
				inputcomp.NewBuilder().
					SetID("tree-search").
					Search().
					Placeholder("Filter tree").
					Value(query).
					ForField(runtimeintent.BindField(meta.QueryField)).
					Build(),
				treeviewcomp.NewBuilder().
					SetID("repo-tree").
					ComponentID(meta.TreeComponentID).
					Nodes([]treeviewcomp.TreeNode{
						{Content: "workspace", Path: "workspace", NodeID: 1, NodeType: "folder"},
						{Content: "lazy-assets", Path: "workspace/lazy-assets", NodeID: 2, NodeType: "folder", Indent: 4, Lazy: true},
						{Content: "README.md", Path: "workspace/README.md", NodeID: 3, NodeType: "file", Indent: 4},
					}).
					ExpandLevel(1).
					SelectedIndex(1).
					ViewportHeight(6).
					ShowBorder(true).
					ShowSearchStats(true).
					SearchStatsStyle(style.Style{}.Foreground(style.Cyan).Bold(true)).
					SelectedStyle(style.Style{}.Foreground(style.Black).Background(style.Green).Bold(true)).
					MatchStyle(style.Style{}.Foreground(style.Yellow).Underline(true)).
					SearchQueryControlled(query).
					SearchPageSize(1).
					MultiSelect().
					SelectionForField(runtimeintent.BindField(meta.CheckedPathsField)).
					OnLazyLoadChildren(func(node treeviewcomp.TreeNode) []treeviewcomp.TreeNode {
						if node.Path != "workspace/lazy-assets" {
							return nil
						}
						return []treeviewcomp.TreeNode{
							{Content: "alpha.log", NodeID: 10, NodeType: "file"},
							{Content: "beta.log", NodeID: 11, NodeType: "file"},
						}
					}).
					Build(),
				ui.NewTextBuilder(fmt.Sprintf("LazyLoads: %d", lazyLoads)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("SelectionEvents: %d", selectionEvents)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("Reorders: %d", reorders)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("LastReordered: %s", lastReordered)).Build(),
			})
	}

	return appFn, initFn, cleanupFn, fixtureStore, meta
}

func TestE2ETreeViewSearchLazyAndSelectionFlow(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newTreeViewFixture()
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(90, 18), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertFocus(ByID("tree-search")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByComponentID(meta.TreeComponentID)); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	for _, r := range "lazy" {
		if err := app.Driver().Key(r); err != nil {
			t.Fatal(err)
		}
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText(`Search: "lazy" 1/1`))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertStyle(ByText(`Search: "lazy" 1/1`), StyleExpect{
		HasFG:   true,
		FG:      style.Cyan,
		HasBold: true,
		Bold:    true,
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentSequence(meta.FieldChangeIntentType, meta.SearchResultsIntentType); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.SearchResultsIntentType); err != nil {
		t.Fatal(err)
	}

	state := fixtureStore.Get()
	if state.Query != "lazy" || state.SearchTotal != 1 || state.SearchSelected != 1 || state.SearchPageCount != 1 {
		t.Fatalf("unexpected tree search state after typing: %+v", state)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	for range "lazy" {
		if err := app.Driver().Special(platform.KeyBackspace); err != nil {
			t.Fatal(err)
		}
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("README.md"))
	}); err != nil {
		t.Fatal(err)
	}

	state = fixtureStore.Get()
	if state.Query != "" {
		t.Fatalf("query should be cleared after backspace sequence: %+v", state)
	}

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().Special(platform.KeyTab); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertFocus(ByComponentID(meta.TreeComponentID)); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertFocusTransition(ByID("tree-search"), ByComponentID(meta.TreeComponentID)); err != nil {
		t.Fatal(err)
	}

	if err := app.Driver().Special(platform.KeyRight); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("alpha.log"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "key:Right"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.LazyLoadIntentType); err != nil {
		t.Fatal(err)
	}

	state = fixtureStore.Get()
	if state.LazyLoads != 1 {
		t.Fatalf("expected one lazy load after expand: %+v", state)
	}
	if err := app.AssertVisible(ByText("LazyLoads: 1")); err != nil {
		t.Fatal(err)
	}

	point, err := app.ResolvePoint(ByText("alpha.log"))
	if err != nil {
		t.Fatal(err)
	}
	if err := app.AssertHit(point, ByID("repo-tree")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	if err := app.Driver().Click(ByText("alpha.log")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "mouse:left:release"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentSequence(meta.SelectionChangeIntentType, meta.FieldChangeIntentType); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.SelectionChangeIntentType); err != nil {
		t.Fatal(err)
	}

	state = fixtureStore.Get()
	if state.CheckedPaths != "workspace/lazy-assets/alpha.log" || state.SelectionEvents != 1 {
		t.Fatalf("unexpected tree selection state after click: %+v", state)
	}
	if err := app.AssertVisible(ByText("SelectionEvents: 1")); err != nil {
		t.Fatal(err)
	}
}

type treeViewReorderFixtureState struct {
	Reorders      int
	LastReordered string
}

type treeViewReorderFixtureMeta struct {
	TreeComponentID       string
	NodeReorderIntentType string
}

func newTreeViewReorderFixture() (ui.ComponentFunc, func(), func(), *store.Store[treeViewReorderFixtureState], treeViewReorderFixtureMeta) {
	fixtureStore := store.NewStore(treeViewReorderFixtureState{})
	meta := treeViewReorderFixtureMeta{
		TreeComponentID:       "fixture.treeview.reorder",
		NodeReorderIntentType: treeviewcomp.NodeReorderIntent{}.IntentType(),
	}

	unregisters := make([]func(), 0, 1)
	initFn := func() {
		rt := rtui.GetGlobalIntentRuntime()
		if rt == nil {
			return
		}
		unregisters = append(unregisters, runtimeintent.RegisterTypedRuntime(rt, func(_ *runtimeintent.ActionContext, i treeviewcomp.NodeReorderIntent) runtimeintent.IntentResult {
			if i.ComponentID != meta.TreeComponentID {
				return runtimeintent.IntentResult{}
			}
			fixtureStore.Update(func(s treeViewReorderFixtureState) treeViewReorderFixtureState {
				s.Reorders++
				s.LastReordered = i.Path
				return s
			})
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
		reorders := ui.UseStoreSelector(fixtureStore, func(s treeViewReorderFixtureState) int { return s.Reorders })
		lastReordered := ui.UseStoreSelector(fixtureStore, func(s treeViewReorderFixtureState) string { return s.LastReordered })

		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("TreeView Reorder Fixture").Build(),
				treeviewcomp.NewBuilder().
					SetID("reorder-tree").
					ComponentID(meta.TreeComponentID).
					Nodes([]treeviewcomp.TreeNode{
						{Content: "workspace", Path: "workspace", NodeID: 1, NodeType: "folder"},
						{Content: "alpha", Path: "workspace/alpha", NodeID: 2, NodeType: "folder", Indent: 4},
						{Content: "beta", Path: "workspace/beta", NodeID: 3, NodeType: "folder", Indent: 4},
						{Content: "gamma", Path: "workspace/gamma", NodeID: 4, NodeType: "folder", Indent: 4},
					}).
					ExpandLevel(1).
					SelectedIndex(1).
					ViewportHeight(6).
					ShowBorder(true).
					ShowSearchStats(true).
					SearchStatsStyle(style.Style{}.Foreground(style.Cyan).Bold(true)).
					Reorderable(true).
					Build(),
				ui.NewTextBuilder(fmt.Sprintf("Reorders: %d", reorders)).Build(),
				ui.NewTextBuilder(fmt.Sprintf("LastReordered: %s", lastReordered)).Build(),
			})
	}

	return appFn, initFn, cleanupFn, fixtureStore, meta
}

func TestE2ETreeViewDragReorderFlow(t *testing.T) {
	appFn, initFn, cleanupFn, fixtureStore, meta := newTreeViewReorderFixture()
	defer cleanupFn()

	app, err := Run(appFn, ui.WithSize(90, 18), ui.WithInit(initFn))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertFocus(ByComponentID(meta.TreeComponentID)); err != nil {
		t.Fatal(err)
	}

	alphaBefore, err := app.ResolvePoint(ByText("alpha"))
	if err != nil {
		t.Fatal(err)
	}
	gammaBefore, err := app.ResolvePoint(ByText("gamma"))
	if err != nil {
		t.Fatal(err)
	}
	if alphaBefore.Y >= gammaBefore.Y {
		t.Fatalf("unexpected initial row order: alpha=%+v gamma=%+v", alphaBefore, gammaBefore)
	}
	treeBounds, err := app.BoundsOf(ByComponentID(meta.TreeComponentID))
	if err != nil {
		t.Fatal(err)
	}
	startX := treeBounds.X + 1
	startY := treeBounds.Y + 3
	endY := treeBounds.Y + 5

	app.ClearIntentLogs()
	app.ClearRawInputs()
	if err := app.Driver().DragAt(startX, startY, startX, endY); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(500*time.Millisecond, 20*time.Millisecond, func(app *App) error {
		alphaAfter, err := app.ResolvePoint(ByText("alpha"))
		if err != nil {
			return err
		}
		gammaAfter, err := app.ResolvePoint(ByText("gamma"))
		if err != nil {
			return err
		}
		if alphaAfter.Y <= gammaAfter.Y {
			return fmt.Errorf("alpha did not move below gamma: alpha=%+v gamma=%+v", alphaAfter, gammaAfter)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "mouse:left:move"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertTraceContains(TraceMatch{Kind: TraceRawInput, Name: "mouse:left:release"}); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.NodeReorderIntentType); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Reorders: 1")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("LastReordered: workspace/alpha")); err != nil {
		t.Fatal(err)
	}

	state := fixtureStore.Get()
	if state.Reorders != 1 || state.LastReordered != "workspace/alpha" {
		t.Fatalf("unexpected tree reorder state after drag: %+v", state)
	}
}

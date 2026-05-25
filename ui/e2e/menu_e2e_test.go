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

type submenuFlipActivateIntent struct{ Token string }

func (i submenuFlipActivateIntent) IntentType() string {
	return "e2e.menu.submenu_flip_activate." + i.Token
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

type submenuFlipFixtureMeta struct {
	ActivateIntentType string
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

func newClampedContextMenuFixture() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Menu E2E Clamped Context Fixture").Build(),
				ui.NewTextBuilder("Clamp to bottom-right viewport edge").Build(),
				menucomp.NewContextMenu([]menucomp.MenuItem{
					menucomp.Action("context-clamped", "Clamped Context Action", nil),
				}).
					SetID("fixture-clamped-context-menu").
					PortalOffset(66, 16).
					Build(),
			})
	}
}

func newOperationalMenuPresetFixture() ui.ComponentFunc {
	return func() ui.VNode {
		items := menucomp.Items()
		items = menucomp.AppendGroup(items, "Operations",
			menucomp.RefreshAction(nil),
			menucomp.ReloadRuntimeAction(nil),
			menucomp.ResetRuntimeAction(nil),
			menucomp.ClearCircuitBreakersAction(nil),
			menucomp.DisabledAction("reset-key", "Reset Key", "Select a key first", nil),
		)
		return ui.NewVStack().
			SetGap(1).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Menu E2E Operational Preset Fixture").Build(),
				menucomp.NewPopup(items).
					SetID("fixture-operational-menu").
					ShowDescriptions(true).
					MinWidth(42).
					Build(),
			})
	}
}

func newRightEdgeAnchoredMenuFixture() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Menu E2E Right Edge Fixture").Build(),
				ui.NewHStack().
					SetGap(0).
					SetChildrenList([]ui.VNode{
						ui.NewTextBuilder(strings.Repeat(" ", 61)).Build(),
						ui.NewTextBuilder("Menu Anchor").SetID("menu-right-anchor").Build(),
					}),
				menucomp.NewPopup([]menucomp.MenuItem{
					menucomp.Action("edge-action", "Right Edge Action", nil),
				}).
					SetID("fixture-right-edge-menu").
					AnchorTo("menu-right-anchor", rttypes.AnchorBottomLeft).
					Placement(menucomp.PlacementBottomStart).
					Build(),
				ui.NewTextBuilder("Right edge footer").Build(),
			})
	}
}

func newSubmenuFlipFixture() (ui.ComponentFunc, func(), func(), submenuFlipFixtureMeta) {
	token := fmt.Sprintf("%d", atomic.AddInt64(&menuFixtureSeq, 1))
	activateIntent := submenuFlipActivateIntent{Token: token}
	meta := submenuFlipFixtureMeta{ActivateIntentType: activateIntent.IntentType()}

	unregisters := make([]func(), 0, 1)
	initFn := func() {
		rt := rtui.GetGlobalIntentRuntime()
		if rt == nil {
			return
		}
		unregisters = append(unregisters, rt.Register(meta.ActivateIntentType, runtimeintent.HandlerFunc(func(_ *runtimeintent.ActionContext, _ runtimeintent.Intent) runtimeintent.IntentResult {
			return runtimeintent.HandledResult()
		})))
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
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Menu E2E Submenu Flip Fixture").Build(),
				ui.NewHStack().
					SetGap(0).
					SetChildrenList([]ui.VNode{
						ui.NewTextBuilder(strings.Repeat(" ", 61)).Build(),
						ui.NewTextBuilder("Submenu Anchor").SetID("menu-submenu-anchor").Build(),
					}),
				menucomp.NewPopup([]menucomp.MenuItem{
					menucomp.Submenu("tools", "More Tools",
						menucomp.Action("details", "Tool Details", activateIntent),
						menucomp.Action("inspect", "Inspect Tool", nil),
					),
					menucomp.Action("quit", "Quit", nil),
				}).
					SetID("fixture-submenu-menu").
					AnchorTo("menu-submenu-anchor", rttypes.AnchorBottomLeft).
					Placement(menucomp.PlacementBottomStart).
					ActivePath(0, 0).
					Build(),
				ui.NewTextBuilder("Submenu footer").Build(),
			})
	}

	return appFn, initFn, cleanupFn, meta
}

func newNestedSubmenuFlipFixture() (ui.ComponentFunc, func(), func(), submenuFlipFixtureMeta) {
	token := fmt.Sprintf("%d", atomic.AddInt64(&menuFixtureSeq, 1))
	activateIntent := submenuFlipActivateIntent{Token: token}
	meta := submenuFlipFixtureMeta{ActivateIntentType: activateIntent.IntentType()}

	unregisters := make([]func(), 0, 1)
	initFn := func() {
		rt := rtui.GetGlobalIntentRuntime()
		if rt == nil {
			return
		}
		unregisters = append(unregisters, rt.Register(meta.ActivateIntentType, runtimeintent.HandlerFunc(func(_ *runtimeintent.ActionContext, _ runtimeintent.Intent) runtimeintent.IntentResult {
			return runtimeintent.HandledResult()
		})))
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
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Menu E2E Nested Submenu Flip Fixture").Build(),
				ui.NewHStack().
					SetGap(0).
					SetChildrenList([]ui.VNode{
						ui.NewTextBuilder(strings.Repeat(" ", 61)).Build(),
						ui.NewTextBuilder("Nested Anchor").SetID("menu-nested-submenu-anchor").Build(),
					}),
				menucomp.NewPopup([]menucomp.MenuItem{
					menucomp.Submenu("tools", "More Tools",
						menucomp.Submenu("advanced", "Advanced Tools",
							menucomp.Action("deep-action", "Deep Action", activateIntent),
						),
					),
					menucomp.Action("quit", "Quit", nil),
				}).
					SetID("fixture-nested-submenu-menu").
					AnchorTo("menu-nested-submenu-anchor", rttypes.AnchorBottomLeft).
					Placement(menucomp.PlacementBottomStart).
					ActivePath(0, 0, 0).
					Build(),
				ui.NewTextBuilder("Nested submenu footer").Build(),
			})
	}

	return appFn, initFn, cleanupFn, meta
}

func newCornerNestedSubmenuFixture() (ui.ComponentFunc, func(), func(), submenuFlipFixtureMeta) {
	token := fmt.Sprintf("%d", atomic.AddInt64(&menuFixtureSeq, 1))
	activateIntent := submenuFlipActivateIntent{Token: token}
	meta := submenuFlipFixtureMeta{ActivateIntentType: activateIntent.IntentType()}

	unregisters := make([]func(), 0, 1)
	initFn := func() {
		rt := rtui.GetGlobalIntentRuntime()
		if rt == nil {
			return
		}
		unregisters = append(unregisters, rt.Register(meta.ActivateIntentType, runtimeintent.HandlerFunc(func(_ *runtimeintent.ActionContext, _ runtimeintent.Intent) runtimeintent.IntentResult {
			return runtimeintent.HandledResult()
		})))
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
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Menu E2E Corner Nested Fixture").Build(),
				ui.NewHStack().
					SetGap(0).
					SetChildrenList([]ui.VNode{
						ui.NewTextBuilder(strings.Repeat(" ", 37)).Build(),
						ui.NewTextBuilder("Corner Anchor").SetID("menu-corner-anchor").Build(),
					}),
				menucomp.NewPopup([]menucomp.MenuItem{
					menucomp.Submenu("ops", "Operations",
						menucomp.Submenu("massive", "Extremely Wide Branch Options",
							menucomp.Action("deep", "Ultra Recovery Action Path", activateIntent),
						),
					),
					menucomp.Action("quit", "Quit", nil),
				}).
					SetID("fixture-corner-submenu-menu").
					AnchorTo("menu-corner-anchor", rttypes.AnchorBottomLeft).
					Placement(menucomp.PlacementBottomStart).
					ActivePath(0, 0, 0).
					Build(),
				ui.NewTextBuilder("Corner footer").Build(),
			})
	}

	return appFn, initFn, cleanupFn, meta
}

func newBottomCornerNestedSubmenuFixture() (ui.ComponentFunc, func(), func(), submenuFlipFixtureMeta) {
	token := fmt.Sprintf("%d", atomic.AddInt64(&menuFixtureSeq, 1))
	activateIntent := submenuFlipActivateIntent{Token: token}
	meta := submenuFlipFixtureMeta{ActivateIntentType: activateIntent.IntentType()}

	unregisters := make([]func(), 0, 1)
	initFn := func() {
		rt := rtui.GetGlobalIntentRuntime()
		if rt == nil {
			return
		}
		unregisters = append(unregisters, rt.Register(meta.ActivateIntentType, runtimeintent.HandlerFunc(func(_ *runtimeintent.ActionContext, _ runtimeintent.Intent) runtimeintent.IntentResult {
			return runtimeintent.HandledResult()
		})))
	}

	cleanupFn := func() {
		for i := len(unregisters) - 1; i >= 0; i-- {
			if unregisters[i] != nil {
				unregisters[i]()
			}
		}
	}

	appFn := func() ui.VNode {
		children := []ui.VNode{
			ui.NewTextBuilder("Menu E2E Bottom Corner Nested Fixture").Build(),
		}
		for i := 0; i < 19; i++ {
			children = append(children, ui.NewTextBuilder(" ").Build())
		}
		children = append(children,
			ui.NewHStack().
				SetGap(0).
				SetChildrenList([]ui.VNode{
					ui.NewTextBuilder(strings.Repeat(" ", 61)).Build(),
					ui.NewTextBuilder("Bottom Corner Anchor").SetID("menu-bottom-corner-anchor").Build(),
				}),
			menucomp.NewPopup([]menucomp.MenuItem{
				menucomp.Action("new", "New", nil),
				menucomp.Action("open", "Open", nil),
				menucomp.Action("save", "Save", nil),
				menucomp.Action("export", "Export", nil),
				menucomp.Submenu("tools", "More Tools",
					menucomp.Action("scan", "Scan", nil),
					menucomp.Action("repair", "Repair", nil),
					menucomp.Action("archive", "Archive", nil),
					menucomp.Action("cleanup", "Cleanup", nil),
					menucomp.Submenu("advanced", "Advanced Tools",
						menucomp.Action("deep", "Deep Action", activateIntent),
						menucomp.Action("verify", "Verify", nil),
						menucomp.Action("history", "History", nil),
						menucomp.Action("reindex", "Reindex", nil),
						menucomp.Action("recover", "Recover", nil),
					),
				),
			}).
				SetID("fixture-bottom-corner-submenu-menu").
				AnchorTo("menu-bottom-corner-anchor", rttypes.AnchorTopRight).
				Placement(menucomp.PlacementTopEnd).
				ActivePath(4, 4, 0).
				Build(),
		)
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList(children)
	}

	return appFn, initFn, cleanupFn, meta
}

func newNarrowBottomCornerNestedSubmenuFixture() (ui.ComponentFunc, func(), func(), submenuFlipFixtureMeta) {
	token := fmt.Sprintf("%d", atomic.AddInt64(&menuFixtureSeq, 1))
	activateIntent := submenuFlipActivateIntent{Token: token}
	meta := submenuFlipFixtureMeta{ActivateIntentType: activateIntent.IntentType()}

	unregisters := make([]func(), 0, 1)
	initFn := func() {
		rt := rtui.GetGlobalIntentRuntime()
		if rt == nil {
			return
		}
		unregisters = append(unregisters, rt.Register(meta.ActivateIntentType, runtimeintent.HandlerFunc(func(_ *runtimeintent.ActionContext, _ runtimeintent.Intent) runtimeintent.IntentResult {
			return runtimeintent.HandledResult()
		})))
	}

	cleanupFn := func() {
		for i := len(unregisters) - 1; i >= 0; i-- {
			if unregisters[i] != nil {
				unregisters[i]()
			}
		}
	}

	appFn := func() ui.VNode {
		children := []ui.VNode{
			ui.NewTextBuilder("Menu E2E Narrow Bottom Corner Fixture").Build(),
		}
		for i := 0; i < 19; i++ {
			children = append(children, ui.NewTextBuilder(" ").Build())
		}
		children = append(children,
			ui.NewHStack().
				SetGap(0).
				SetChildrenList([]ui.VNode{
					ui.NewTextBuilder(strings.Repeat(" ", 75)).Build(),
					ui.NewTextBuilder("A").SetID("menu-narrow-bottom-corner-anchor").Build(),
				}),
			menucomp.NewPopup([]menucomp.MenuItem{
				menucomp.Action("new", "New", nil),
				menucomp.Action("open", "Open", nil),
				menucomp.Action("save", "Save", nil),
				menucomp.Action("export", "Export", nil),
				menucomp.Submenu("ops", "Operations",
					menucomp.Action("scan", "Inspect Intermediate Recovery Ledger", nil),
					menucomp.Action("repair", "Repair", nil),
					menucomp.Action("archive", "Archive", nil),
					menucomp.Action("cleanup", "Cleanup", nil),
					menucomp.Submenu("massive", "Branch",
						menucomp.Action("deep", "Ultra Recovery Action Path", activateIntent),
						menucomp.Action("verify", "Verify Recovery Journal", nil),
						menucomp.Action("history", "Inspect Historical Recovery Entries", nil),
						menucomp.Action("reindex", "Reindex Recovery Snapshots", nil),
						menucomp.Action("recover", "Recover Snapshot Chain", nil),
					),
				),
			}).
				SetID("fixture-narrow-bottom-corner-submenu-menu").
				AnchorTo("menu-narrow-bottom-corner-anchor", rttypes.AnchorTopRight).
				Placement(menucomp.PlacementTopEnd).
				ActivePath(4, 4, 0).
				Build(),
		)
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList(children)
	}

	return appFn, initFn, cleanupFn, meta
}

func newMirrorRightAfterLeftClampFixture() (ui.ComponentFunc, func(), func(), submenuFlipFixtureMeta) {
	token := fmt.Sprintf("%d", atomic.AddInt64(&menuFixtureSeq, 1))
	activateIntent := submenuFlipActivateIntent{Token: token}
	meta := submenuFlipFixtureMeta{ActivateIntentType: activateIntent.IntentType()}

	unregisters := make([]func(), 0, 1)
	initFn := func() {
		rt := rtui.GetGlobalIntentRuntime()
		if rt == nil {
			return
		}
		unregisters = append(unregisters, rt.Register(meta.ActivateIntentType, runtimeintent.HandlerFunc(func(_ *runtimeintent.ActionContext, _ runtimeintent.Intent) runtimeintent.IntentResult {
			return runtimeintent.HandledResult()
		})))
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
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Menu E2E Mirror Right Fixture").Build(),
				ui.NewHStack().
					SetGap(0).
					SetChildrenList([]ui.VNode{
						ui.NewTextBuilder(strings.Repeat(" ", 75)).Build(),
						ui.NewTextBuilder("A").SetID("menu-mirror-right-anchor").Build(),
					}),
				menucomp.NewPopup([]menucomp.MenuItem{
					menucomp.Action("new", "New", nil),
					menucomp.Action("open", "Open", nil),
					menucomp.Action("save", "Save", nil),
					menucomp.Action("export", "Export", nil),
					menucomp.Submenu("ops", "Operations",
						menucomp.Action("pad", "Inspect Intermediate Recovery Ledger", nil),
						menucomp.Action("repair", "Repair", nil),
						menucomp.Action("archive", "Archive", nil),
						menucomp.Submenu("branch", "Branch",
							menucomp.Action("node", "Moderately Wide Inner Recovery Journal", nil),
							menucomp.Submenu("pivot", "Pivot",
								menucomp.Action("deep", "Deep Action", activateIntent),
							),
						),
					),
				}).
					SetID("fixture-mirror-right-submenu-menu").
					AnchorTo("menu-mirror-right-anchor", rttypes.AnchorBottomLeft).
					Placement(menucomp.PlacementBottomStart).
					ActivePath(4, 3, 1, 0).
					Build(),
				ui.NewTextBuilder("Mirror right footer").Build(),
			})
	}

	return appFn, initFn, cleanupFn, meta
}

func newMirrorLeftAfterRightClampFixture() (ui.ComponentFunc, func(), func(), submenuFlipFixtureMeta) {
	token := fmt.Sprintf("%d", atomic.AddInt64(&menuFixtureSeq, 1))
	activateIntent := submenuFlipActivateIntent{Token: token}
	meta := submenuFlipFixtureMeta{ActivateIntentType: activateIntent.IntentType()}

	unregisters := make([]func(), 0, 1)
	initFn := func() {
		rt := rtui.GetGlobalIntentRuntime()
		if rt == nil {
			return
		}
		unregisters = append(unregisters, rt.Register(meta.ActivateIntentType, runtimeintent.HandlerFunc(func(_ *runtimeintent.ActionContext, _ runtimeintent.Intent) runtimeintent.IntentResult {
			return runtimeintent.HandledResult()
		})))
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
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Menu E2E Mirror Left Fixture").Build(),
				ui.NewHStack().
					SetGap(0).
					SetChildrenList([]ui.VNode{
						ui.NewTextBuilder("A").SetID("menu-mirror-left-anchor").Build(),
						ui.NewTextBuilder(strings.Repeat(" ", 63)).Build(),
					}),
				menucomp.NewPopup([]menucomp.MenuItem{
					menucomp.Action("new", "New", nil),
					menucomp.Action("open", "Open", nil),
					menucomp.Action("save", "Save", nil),
					menucomp.Action("export", "Export", nil),
					menucomp.Submenu("ops", "Operations",
						menucomp.Action("wide", "Extremely Wide Recovery Workspace Ledger", nil),
						menucomp.Action("repair", "Repair", nil),
						menucomp.Action("archive", "Archive", nil),
						menucomp.Submenu("branch", "Branch",
							menucomp.Action("node", "Moderately Wide Inner Recovery Journal", nil),
							menucomp.Submenu("pivot", "Pivot",
								menucomp.Action("deep", "Deep Action", activateIntent),
							),
						),
					),
				}).
					SetID("fixture-mirror-left-submenu-menu").
					AnchorTo("menu-mirror-left-anchor", rttypes.AnchorBottomLeft).
					Placement(menucomp.PlacementBottomStart).
					ActivePath(4, 3, 1, 0).
					Build(),
				ui.NewTextBuilder("Mirror left footer").Build(),
			})
	}

	return appFn, initFn, cleanupFn, meta
}

func newBottomMirrorLeftAfterRightClampFixture() (ui.ComponentFunc, func(), func(), submenuFlipFixtureMeta) {
	token := fmt.Sprintf("%d", atomic.AddInt64(&menuFixtureSeq, 1))
	activateIntent := submenuFlipActivateIntent{Token: token}
	meta := submenuFlipFixtureMeta{ActivateIntentType: activateIntent.IntentType()}

	unregisters := make([]func(), 0, 1)
	initFn := func() {
		rt := rtui.GetGlobalIntentRuntime()
		if rt == nil {
			return
		}
		unregisters = append(unregisters, rt.Register(meta.ActivateIntentType, runtimeintent.HandlerFunc(func(_ *runtimeintent.ActionContext, _ runtimeintent.Intent) runtimeintent.IntentResult {
			return runtimeintent.HandledResult()
		})))
	}

	cleanupFn := func() {
		for i := len(unregisters) - 1; i >= 0; i-- {
			if unregisters[i] != nil {
				unregisters[i]()
			}
		}
	}

	appFn := func() ui.VNode {
		children := []ui.VNode{
			ui.NewTextBuilder("Menu E2E Bottom Mirror Left Fixture").Build(),
		}
		for i := 0; i < 19; i++ {
			children = append(children, ui.NewTextBuilder(" ").Build())
		}
		children = append(children,
			ui.NewTextBuilder("Bottom Mirror Left Anchor").SetID("menu-bottom-mirror-left-anchor").Build(),
			menucomp.NewPopup([]menucomp.MenuItem{
				menucomp.Action("new", "New", nil),
				menucomp.Action("open", "Open", nil),
				menucomp.Action("save", "Save", nil),
				menucomp.Action("export", "Export", nil),
				menucomp.Submenu("ops", "Operations",
					menucomp.Action("scan", "Scan", nil),
					menucomp.Action("repair", "Repair", nil),
					menucomp.Action("archive", "Archive", nil),
					menucomp.Action("cleanup", "Cleanup", nil),
					menucomp.Submenu("branch", "Branch",
						menucomp.Action("wide", "Extremely Wide Recovery Workspace Ledger", nil),
						menucomp.Action("repair-node", "Repair Recovery Nodes", nil),
						menucomp.Action("audit", "Audit", nil),
						menucomp.Action("history", "Historical Recovery Timeline Browser", nil),
						menucomp.Submenu("pivot", "Pivot",
							menucomp.Action("deep", "Deep Action", activateIntent),
							menucomp.Action("verify", "Verify", nil),
							menucomp.Action("recover", "Recover", nil),
						),
					),
				),
			}).
				SetID("fixture-bottom-mirror-left-submenu-menu").
				AnchorTo("menu-bottom-mirror-left-anchor", rttypes.AnchorTopLeft).
				Placement(menucomp.PlacementTopStart).
				ActivePath(4, 4, 4, 0).
				Build(),
		)
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList(children)
	}

	return appFn, initFn, cleanupFn, meta
}

func newBottomMirrorRightAfterLeftClampFixture() (ui.ComponentFunc, func(), func(), submenuFlipFixtureMeta) {
	token := fmt.Sprintf("%d", atomic.AddInt64(&menuFixtureSeq, 1))
	activateIntent := submenuFlipActivateIntent{Token: token}
	meta := submenuFlipFixtureMeta{ActivateIntentType: activateIntent.IntentType()}

	unregisters := make([]func(), 0, 1)
	initFn := func() {
		rt := rtui.GetGlobalIntentRuntime()
		if rt == nil {
			return
		}
		unregisters = append(unregisters, rt.Register(meta.ActivateIntentType, runtimeintent.HandlerFunc(func(_ *runtimeintent.ActionContext, _ runtimeintent.Intent) runtimeintent.IntentResult {
			return runtimeintent.HandledResult()
		})))
	}

	cleanupFn := func() {
		for i := len(unregisters) - 1; i >= 0; i-- {
			if unregisters[i] != nil {
				unregisters[i]()
			}
		}
	}

	appFn := func() ui.VNode {
		children := []ui.VNode{
			ui.NewTextBuilder("Menu E2E Bottom Mirror Right Fixture").Build(),
		}
		for i := 0; i < 19; i++ {
			children = append(children, ui.NewTextBuilder(" ").Build())
		}
		children = append(children,
			ui.NewHStack().
				SetGap(0).
				SetChildrenList([]ui.VNode{
					ui.NewTextBuilder(strings.Repeat(" ", 75)).Build(),
					ui.NewTextBuilder("A").SetID("menu-bottom-mirror-right-anchor").Build(),
				}),
			menucomp.NewPopup([]menucomp.MenuItem{
				menucomp.Action("new", "New", nil),
				menucomp.Action("open", "Open", nil),
				menucomp.Action("save", "Save", nil),
				menucomp.Action("export", "Export", nil),
				menucomp.Submenu("ops", "Operations",
					menucomp.Action("scan", "Scan", nil),
					menucomp.Action("repair", "Repair", nil),
					menucomp.Action("archive", "Archive", nil),
					menucomp.Action("cleanup", "Cleanup", nil),
					menucomp.Submenu("branch", "Branch",
						menucomp.Action("wide", "Extremely Wide Recovery Workspace Ledger", nil),
						menucomp.Action("repair-node", "Repair Recovery Nodes", nil),
						menucomp.Action("audit", "Audit", nil),
						menucomp.Action("history", "Historical Recovery Timeline Browser", nil),
						menucomp.Submenu("pivot", "Pivot",
							menucomp.Action("deep", "Deep Action", activateIntent),
							menucomp.Action("verify", "Verify", nil),
							menucomp.Action("recover", "Recover", nil),
						),
					),
				),
			}).
				SetID("fixture-bottom-mirror-right-submenu-menu").
				AnchorTo("menu-bottom-mirror-right-anchor", rttypes.AnchorTopRight).
				Placement(menucomp.PlacementTopEnd).
				ActivePath(4, 4, 4, 0).
				Build(),
		)
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList(children)
	}

	return appFn, initFn, cleanupFn, meta
}

func newBottomZigZagMirrorFixture() (ui.ComponentFunc, func(), func(), submenuFlipFixtureMeta) {
	token := fmt.Sprintf("%d", atomic.AddInt64(&menuFixtureSeq, 1))
	activateIntent := submenuFlipActivateIntent{Token: token}
	meta := submenuFlipFixtureMeta{ActivateIntentType: activateIntent.IntentType()}

	unregisters := make([]func(), 0, 1)
	initFn := func() {
		rt := rtui.GetGlobalIntentRuntime()
		if rt == nil {
			return
		}
		unregisters = append(unregisters, rt.Register(meta.ActivateIntentType, runtimeintent.HandlerFunc(func(_ *runtimeintent.ActionContext, _ runtimeintent.Intent) runtimeintent.IntentResult {
			return runtimeintent.HandledResult()
		})))
	}

	cleanupFn := func() {
		for i := len(unregisters) - 1; i >= 0; i-- {
			if unregisters[i] != nil {
				unregisters[i]()
			}
		}
	}

	appFn := func() ui.VNode {
		children := []ui.VNode{
			ui.NewTextBuilder("Menu E2E Bottom ZigZag Fixture").Build(),
		}
		for i := 0; i < 19; i++ {
			children = append(children, ui.NewTextBuilder(" ").Build())
		}
		children = append(children,
			ui.NewTextBuilder("Bottom ZigZag Anchor").SetID("menu-bottom-zigzag-anchor").Build(),
			menucomp.NewPopup([]menucomp.MenuItem{
				menucomp.Action("new", "New", nil),
				menucomp.Action("open", "Open", nil),
				menucomp.Action("save", "Save", nil),
				menucomp.Action("export", "Export", nil),
				menucomp.Submenu("ops", "Operations",
					menucomp.Action("scan", "Scan", nil),
					menucomp.Action("repair", "Repair", nil),
					menucomp.Action("archive", "Archive", nil),
					menucomp.Action("cleanup", "Cleanup", nil),
					menucomp.Submenu("branch", "Branch",
						menucomp.Action("wide", "Extremely Wide Recovery Workspace Ledger", nil),
						menucomp.Action("repair-node", "Repair Recovery Nodes", nil),
						menucomp.Action("audit", "Audit", nil),
						menucomp.Action("history", "Historical Recovery Timeline Browser", nil),
						menucomp.Submenu("pivot", "Pivot",
							menucomp.Action("verify", "Verify", nil),
							menucomp.Action("compare", "Compare", nil),
							menucomp.Action("recover", "Recover", nil),
							menucomp.Submenu("rebound", "Rebound",
								menucomp.Action("deep", "Deep Action", activateIntent),
								menucomp.Action("secondary", "Secondary Recovery Marker", nil),
								menucomp.Action("rollback", "Rollback", nil),
								menucomp.Action("archive-tail", "Archive Tail", nil),
							),
						),
					),
				),
			}).
				SetID("fixture-bottom-zigzag-submenu-menu").
				AnchorTo("menu-bottom-zigzag-anchor", rttypes.AnchorTopLeft).
				Placement(menucomp.PlacementTopStart).
				ActivePath(4, 4, 4, 3, 0).
				Build(),
		)
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList(children)
	}

	return appFn, initFn, cleanupFn, meta
}

func newTopEdgeAnchoredMenuFixture() ui.ComponentFunc {
	return func() ui.VNode {
		return ui.NewVStack().
			SetGap(0).
			SetChildrenList([]ui.VNode{
				ui.NewTextBuilder("Menu E2E Top Edge Fixture").Build(),
				ui.NewHStack().
					SetGap(0).
					SetChildrenList([]ui.VNode{
						ui.NewTextBuilder(strings.Repeat(" ", 61)).Build(),
						ui.NewTextBuilder("Menu Anchor").SetID("menu-top-anchor").Build(),
					}),
				menucomp.NewPopup([]menucomp.MenuItem{
					menucomp.Action("top-action", "Top Edge Action", nil),
				}).
					SetID("fixture-top-edge-menu").
					AnchorTo("menu-top-anchor", rttypes.AnchorTopRight).
					Placement(menucomp.PlacementTopEnd).
					Build(),
				ui.NewTextBuilder("Top edge footer").Build(),
			})
	}
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

	if popupBounds.X+popupBounds.Width-1 != anchorBounds.X+anchorBounds.Width-1 {
		t.Fatalf("popup surface right edge = %d, want %d", popupBounds.X+popupBounds.Width-1, anchorBounds.X+anchorBounds.Width-1)
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

	if err := app.Eventually(1*time.Second, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("Context Action"))
	}); err != nil {
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
	backgroundBounds, err := app.BoundsOf(ByID("menu-background-btn"))
	if err != nil {
		t.Fatal(err)
	}
	clickPoint := Point{X: backgroundBounds.X, Y: backgroundBounds.Y}
	if bounds.Contains(clickPoint.X, clickPoint.Y) {
		clickPoint.X = bounds.X + bounds.Width + 1
		clickPoint.Y = backgroundBounds.Y
	}
	if err := app.Driver().ClickAt(clickPoint.X, clickPoint.Y); err != nil {
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

func TestE2EContextMenuClampKeepsPopupWithinViewport(t *testing.T) {
	app, err := Run(newClampedContextMenuFixture(),
		ui.WithSize(72, 18),
		ui.WithPluginSetup(func(app *framework.App) {
			menucomp.Install(app, nil)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("Clamped Context Action")); err != nil {
		t.Fatal(err)
	}

	bounds, err := app.BoundsOf(ByID("fixture-clamped-context-menu"))
	if err != nil {
		t.Fatal(err)
	}

	if bounds.X+bounds.Width > 72 {
		t.Fatalf("context menu right edge = %d, want <= 72", bounds.X+bounds.Width)
	}
	if bounds.Y+bounds.Height > 18 {
		t.Fatalf("context menu bottom edge = %d, want <= 18", bounds.Y+bounds.Height)
	}
	if bounds.X != 72-bounds.Width {
		t.Fatalf("context menu x = %d, want %d after right-edge clamp", bounds.X, 72-bounds.Width)
	}
	if bounds.Y != 18-bounds.Height {
		t.Fatalf("context menu y = %d, want %d after bottom-edge clamp", bounds.Y, 18-bounds.Height)
	}
}

func TestE2EMenuOperationalPresetsRender(t *testing.T) {
	app, err := Run(newOperationalMenuPresetFixture(), ui.WithSize(96, 16))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	for _, text := range []string{
		"Operations",
		"Refresh",
		"Refresh current data",
		"Reload Runtime",
		"Reload runtime configuration",
		"Reset Runtime",
		"Clear Circuit Breakers",
		"Reset Key",
		"Select a key first",
	} {
		if err := app.AssertVisible(ByText(text)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestE2EMenuAnchoredPopupFallsBackWithinRightViewportEdge(t *testing.T) {
	app, err := Run(newRightEdgeAnchoredMenuFixture(),
		ui.WithSize(72, 18),
		ui.WithPluginSetup(func(app *framework.App) {
			menucomp.Install(app, nil)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("Right Edge Action")); err != nil {
		t.Fatal(err)
	}

	anchorBounds, err := app.BoundsOf(ByID("menu-right-anchor"))
	if err != nil {
		t.Fatal(err)
	}
	popupBounds, err := app.BoundsOf(ByID("fixture-right-edge-menu"))
	if err != nil {
		t.Fatal(err)
	}

	if popupBounds.X+popupBounds.Width > 72 {
		t.Fatalf("popup right edge = %d, want <= 72", popupBounds.X+popupBounds.Width)
	}
	if popupBounds.X+popupBounds.Width-1 != anchorBounds.X+anchorBounds.Width-1 {
		t.Fatalf("popup surface right edge = %d, want %d after bottom-start fallback", popupBounds.X+popupBounds.Width-1, anchorBounds.X+anchorBounds.Width-1)
	}
	if popupBounds.Y != anchorBounds.Y+anchorBounds.Height {
		t.Fatalf("popup top = %d, want %d below anchor", popupBounds.Y, anchorBounds.Y+anchorBounds.Height)
	}
}

func TestE2EMenuAnchoredPopupFallsBackBelowTopViewportEdge(t *testing.T) {
	app, err := Run(newTopEdgeAnchoredMenuFixture(),
		ui.WithSize(72, 18),
		ui.WithPluginSetup(func(app *framework.App) {
			menucomp.Install(app, nil)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("Top Edge Action")); err != nil {
		t.Fatal(err)
	}

	anchorBounds, err := app.BoundsOf(ByID("menu-top-anchor"))
	if err != nil {
		t.Fatal(err)
	}
	popupBounds, err := app.BoundsOf(ByID("fixture-top-edge-menu"))
	if err != nil {
		t.Fatal(err)
	}

	if popupBounds.Y != anchorBounds.Y+anchorBounds.Height {
		t.Fatalf("popup top = %d, want %d after top-edge fallback", popupBounds.Y, anchorBounds.Y+anchorBounds.Height)
	}
	if popupBounds.X+popupBounds.Width > 72 {
		t.Fatalf("popup right edge = %d, want <= 72", popupBounds.X+popupBounds.Width)
	}
	if popupBounds.X+popupBounds.Width-1 != anchorBounds.X+anchorBounds.Width-1 {
		t.Fatalf("popup surface right edge = %d, want %d after top-end fallback", popupBounds.X+popupBounds.Width-1, anchorBounds.X+anchorBounds.Width-1)
	}
}

func TestE2EMenuSubmenuFlipsLeftNearRightViewportEdge(t *testing.T) {
	appFn, initFn, cleanupFn, meta := newSubmenuFlipFixture()
	defer cleanupFn()

	app, err := Run(appFn,
		ui.WithSize(72, 18),
		ui.WithInit(initFn),
		ui.WithPluginSetup(func(app *framework.App) {
			menucomp.Install(app, nil)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.Eventually(1*time.Second, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("More Tools"))
	}); err != nil {
		t.Fatal(err)
	}
	if err := app.Eventually(1*time.Second, 20*time.Millisecond, func(app *App) error {
		return app.AssertVisible(ByText("Tool Details"))
	}); err != nil {
		t.Fatal(err)
	}

	rootPoint, err := app.ResolvePoint(ByText("More Tools"))
	if err != nil {
		t.Fatal(err)
	}
	childPoint, err := app.ResolvePoint(ByText("Tool Details"))
	if err != nil {
		t.Fatal(err)
	}
	if childPoint.X >= rootPoint.X {
		t.Fatalf("submenu child x = %d, want < root item x = %d after left-start flip", childPoint.X, rootPoint.X)
	}

	popupBounds, err := app.BoundsOf(ByID("fixture-submenu-menu"))
	if err != nil {
		t.Fatal(err)
	}
	if popupBounds.X > childPoint.X {
		t.Fatalf("popup hit bounds x = %d, want <= submenu child x = %d", popupBounds.X, childPoint.X)
	}

	if err := app.AssertHit(childPoint, ByID("fixture-submenu-menu")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	if err := app.Driver().Click(ByText("Tool Details")); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitIntent(meta.ActivateIntentType, 500*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.ActivateIntentType); err != nil {
		t.Fatal(err)
	}
}

func TestE2EMenuNestedSubmenusKeepFlippedDirectionNearRightViewportEdge(t *testing.T) {
	appFn, initFn, cleanupFn, meta := newNestedSubmenuFlipFixture()
	defer cleanupFn()

	app, err := Run(appFn,
		ui.WithSize(72, 18),
		ui.WithInit(initFn),
		ui.WithPluginSetup(func(app *framework.App) {
			menucomp.Install(app, nil)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("More Tools")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Advanced Tools")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Deep Action")); err != nil {
		t.Fatal(err)
	}

	rootPoint, err := app.ResolvePoint(ByText("More Tools"))
	if err != nil {
		t.Fatal(err)
	}
	branchPoint, err := app.ResolvePoint(ByText("Advanced Tools"))
	if err != nil {
		t.Fatal(err)
	}
	deepPoint, err := app.ResolvePoint(ByText("Deep Action"))
	if err != nil {
		t.Fatal(err)
	}
	if !(deepPoint.X < branchPoint.X && branchPoint.X < rootPoint.X) {
		t.Fatalf("expected deep=%d < branch=%d < root=%d for continuous left cascade", deepPoint.X, branchPoint.X, rootPoint.X)
	}

	popupBounds, err := app.BoundsOf(ByID("fixture-nested-submenu-menu"))
	if err != nil {
		t.Fatal(err)
	}
	if popupBounds.X > deepPoint.X {
		t.Fatalf("popup hit bounds x = %d, want <= deep submenu x = %d", popupBounds.X, deepPoint.X)
	}

	if err := app.AssertHit(deepPoint, ByID("fixture-nested-submenu-menu")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	if err := app.Driver().Click(ByText("Deep Action")); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitIntent(meta.ActivateIntentType, 500*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.ActivateIntentType); err != nil {
		t.Fatal(err)
	}
}

func TestE2EMenuNestedSubmenusStayUsableInNarrowCornerViewport(t *testing.T) {
	appFn, initFn, cleanupFn, meta := newCornerNestedSubmenuFixture()
	defer cleanupFn()

	app, err := Run(appFn,
		ui.WithSize(48, 12),
		ui.WithInit(initFn),
		ui.WithPluginSetup(func(app *framework.App) {
			menucomp.Install(app, nil)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("Ultra Recovery Action Path")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Extremely Wide Branch Options")); err != nil {
		t.Fatal(err)
	}

	branchPoint, err := app.ResolvePoint(ByText("Extremely Wide Branch Options"))
	if err != nil {
		t.Fatal(err)
	}
	deepPoint, err := app.ResolvePoint(ByText("Ultra Recovery Action Path"))
	if err != nil {
		t.Fatal(err)
	}
	if !(deepPoint.X < branchPoint.X) {
		t.Fatalf("expected deep=%d < branch=%d for left-edge clamp cascade", deepPoint.X, branchPoint.X)
	}

	popupBounds, err := app.BoundsOf(ByID("fixture-corner-submenu-menu"))
	if err != nil {
		t.Fatal(err)
	}
	if popupBounds.X > deepPoint.X {
		t.Fatalf("popup hit bounds x = %d, want <= deep submenu x = %d", popupBounds.X, deepPoint.X)
	}

	if err := app.AssertHit(deepPoint, ByID("fixture-corner-submenu-menu")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	if err := app.Driver().Click(ByText("Ultra Recovery Action Path")); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitIntent(meta.ActivateIntentType, 500*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.ActivateIntentType); err != nil {
		t.Fatal(err)
	}
}

func TestE2EMenuNestedSubmenusClampUpwardNearBottomRightCorner(t *testing.T) {
	appFn, initFn, cleanupFn, meta := newBottomCornerNestedSubmenuFixture()
	defer cleanupFn()

	app, err := Run(appFn,
		ui.WithSize(90, 24),
		ui.WithInit(initFn),
		ui.WithPluginSetup(func(app *framework.App) {
			menucomp.Install(app, nil)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("More Tools")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Advanced Tools")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Deep Action")); err != nil {
		t.Fatal(err)
	}

	rootPoint, err := app.ResolvePoint(ByText("More Tools"))
	if err != nil {
		t.Fatal(err)
	}
	branchPoint, err := app.ResolvePoint(ByText("Advanced Tools"))
	if err != nil {
		t.Fatal(err)
	}
	deepPoint, err := app.ResolvePoint(ByText("Deep Action"))
	if err != nil {
		t.Fatal(err)
	}
	if !(deepPoint.X < branchPoint.X && branchPoint.X < rootPoint.X) {
		t.Fatalf("expected deep=%d < branch=%d < root=%d for bottom-right left cascade", deepPoint.X, branchPoint.X, rootPoint.X)
	}
	if deepPoint.Y >= branchPoint.Y {
		t.Fatalf("deep submenu y = %d, want < branch y = %d after upward clamp", deepPoint.Y, branchPoint.Y)
	}

	popupBounds, err := app.BoundsOf(ByID("fixture-bottom-corner-submenu-menu"))
	if err != nil {
		t.Fatal(err)
	}
	if popupBounds.X > deepPoint.X {
		t.Fatalf("popup hit bounds x = %d, want <= deep submenu x = %d", popupBounds.X, deepPoint.X)
	}
	if popupBounds.Y > deepPoint.Y {
		t.Fatalf("popup hit bounds y = %d, want <= deep submenu y = %d", popupBounds.Y, deepPoint.Y)
	}
	if popupBounds.Y+popupBounds.Height > 24 {
		t.Fatalf("popup bottom edge = %d, want <= 24", popupBounds.Y+popupBounds.Height)
	}

	if err := app.AssertHit(deepPoint, ByID("fixture-bottom-corner-submenu-menu")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	if err := app.Driver().Click(ByText("Deep Action")); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitIntent(meta.ActivateIntentType, 500*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.ActivateIntentType); err != nil {
		t.Fatal(err)
	}
}

func TestE2EMenuNestedSubmenusClampLeftAndUpwardInNarrowBottomCorner(t *testing.T) {
	appFn, initFn, cleanupFn, meta := newNarrowBottomCornerNestedSubmenuFixture()
	defer cleanupFn()

	app, err := Run(appFn,
		ui.WithSize(80, 24),
		ui.WithInit(initFn),
		ui.WithPluginSetup(func(app *framework.App) {
			menucomp.Install(app, nil)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("Ultra Recovery Action Path")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Operations")); err != nil {
		t.Fatal(err)
	}

	rootPoint, err := app.ResolvePoint(ByText("Operations"))
	if err != nil {
		t.Fatal(err)
	}
	deepPoint, err := app.ResolvePoint(ByText("Ultra Recovery Action Path"))
	if err != nil {
		t.Fatal(err)
	}
	if deepPoint.X >= rootPoint.X {
		t.Fatalf("deep submenu x = %d, want < root x = %d after left-edge clamp", deepPoint.X, rootPoint.X)
	}
	expectedUnclampedDeepY := rootPoint.Y + 4
	if deepPoint.Y >= expectedUnclampedDeepY {
		t.Fatalf("deep submenu y = %d, want < %d after upward clamp from root row geometry", deepPoint.Y, expectedUnclampedDeepY)
	}

	popupBounds, err := app.BoundsOf(ByID("fixture-narrow-bottom-corner-submenu-menu"))
	if err != nil {
		t.Fatal(err)
	}
	if popupBounds.X != 0 {
		t.Fatalf("popup hit bounds x = %d, want 0 after left-edge clamp", popupBounds.X)
	}
	if popupBounds.Y > deepPoint.Y {
		t.Fatalf("popup hit bounds y = %d, want <= deep submenu y = %d", popupBounds.Y, deepPoint.Y)
	}
	if popupBounds.Y+popupBounds.Height > 24 {
		t.Fatalf("popup bottom edge = %d, want <= 24", popupBounds.Y+popupBounds.Height)
	}

	if err := app.AssertHit(deepPoint, ByID("fixture-narrow-bottom-corner-submenu-menu")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	if err := app.Driver().Click(ByText("Ultra Recovery Action Path")); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitIntent(meta.ActivateIntentType, 500*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.ActivateIntentType); err != nil {
		t.Fatal(err)
	}
}

func TestE2EMenuNestedSubmenusMirrorRightAfterLeftEdgeClamp(t *testing.T) {
	appFn, initFn, cleanupFn, meta := newMirrorRightAfterLeftClampFixture()
	defer cleanupFn()

	app, err := Run(appFn,
		ui.WithSize(80, 18),
		ui.WithInit(initFn),
		ui.WithPluginSetup(func(app *framework.App) {
			menucomp.Install(app, nil)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("Operations")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Pivot")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Deep Action")); err != nil {
		t.Fatal(err)
	}

	rootPoint, err := app.ResolvePoint(ByText("Operations"))
	if err != nil {
		t.Fatal(err)
	}
	pivotPoint, err := app.ResolvePoint(ByText("Pivot"))
	if err != nil {
		t.Fatal(err)
	}
	deepPoint, err := app.ResolvePoint(ByText("Deep Action"))
	if err != nil {
		t.Fatal(err)
	}
	if pivotPoint.X >= rootPoint.X {
		t.Fatalf("pivot submenu x = %d, want < root x = %d after left-edge clamp", pivotPoint.X, rootPoint.X)
	}
	if deepPoint.X <= pivotPoint.X {
		t.Fatalf("deep submenu x = %d, want > pivot x = %d after mirrored right fallback", deepPoint.X, pivotPoint.X)
	}

	popupBounds, err := app.BoundsOf(ByID("fixture-mirror-right-submenu-menu"))
	if err != nil {
		t.Fatal(err)
	}
	if popupBounds.X != 0 {
		t.Fatalf("popup hit bounds x = %d, want 0 after left-edge clamp branch", popupBounds.X)
	}
	if popupBounds.X+popupBounds.Width <= deepPoint.X {
		t.Fatalf("popup hit bounds right edge = %d, want > deep submenu x = %d", popupBounds.X+popupBounds.Width, deepPoint.X)
	}

	if err := app.AssertHit(deepPoint, ByID("fixture-mirror-right-submenu-menu")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	if err := app.Driver().Click(ByText("Deep Action")); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitIntent(meta.ActivateIntentType, 500*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.ActivateIntentType); err != nil {
		t.Fatal(err)
	}
}

func TestE2EMenuNestedSubmenusMirrorLeftAfterRightEdgeClamp(t *testing.T) {
	appFn, initFn, cleanupFn, meta := newMirrorLeftAfterRightClampFixture()
	defer cleanupFn()

	app, err := Run(appFn,
		ui.WithSize(64, 18),
		ui.WithInit(initFn),
		ui.WithPluginSetup(func(app *framework.App) {
			menucomp.Install(app, nil)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("Operations")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Pivot")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Deep Action")); err != nil {
		t.Fatal(err)
	}

	rootPoint, err := app.ResolvePoint(ByText("Operations"))
	if err != nil {
		t.Fatal(err)
	}
	pivotPoint, err := app.ResolvePoint(ByText("Pivot"))
	if err != nil {
		t.Fatal(err)
	}
	deepPoint, err := app.ResolvePoint(ByText("Deep Action"))
	if err != nil {
		t.Fatal(err)
	}
	if pivotPoint.X <= rootPoint.X {
		t.Fatalf("pivot submenu x = %d, want > root x = %d after right-edge clamp branch", pivotPoint.X, rootPoint.X)
	}
	if deepPoint.X >= pivotPoint.X {
		t.Fatalf("deep submenu x = %d, want < pivot x = %d after mirrored left fallback", deepPoint.X, pivotPoint.X)
	}

	popupBounds, err := app.BoundsOf(ByID("fixture-mirror-left-submenu-menu"))
	if err != nil {
		t.Fatal(err)
	}
	if popupBounds.X != 0 {
		t.Fatalf("popup hit bounds x = %d, want 0 with root anchored at left edge", popupBounds.X)
	}
	if popupBounds.X+popupBounds.Width <= pivotPoint.X {
		t.Fatalf("popup hit bounds right edge = %d, want > pivot submenu x = %d", popupBounds.X+popupBounds.Width, pivotPoint.X)
	}

	if err := app.AssertHit(deepPoint, ByID("fixture-mirror-left-submenu-menu")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	if err := app.Driver().Click(ByText("Deep Action")); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitIntent(meta.ActivateIntentType, 500*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.ActivateIntentType); err != nil {
		t.Fatal(err)
	}
}

func TestE2EMenuNestedSubmenusMirrorLeftAndClampUpwardNearBottomRightCorner(t *testing.T) {
	appFn, initFn, cleanupFn, meta := newBottomMirrorLeftAfterRightClampFixture()
	defer cleanupFn()

	app, err := Run(appFn,
		ui.WithSize(64, 24),
		ui.WithInit(initFn),
		ui.WithPluginSetup(func(app *framework.App) {
			menucomp.Install(app, nil)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("Operations")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Pivot")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Deep Action")); err != nil {
		t.Fatal(err)
	}

	rootPoint, err := app.ResolvePoint(ByText("Operations"))
	if err != nil {
		t.Fatal(err)
	}
	pivotPoint, err := app.ResolvePoint(ByText("Pivot"))
	if err != nil {
		t.Fatal(err)
	}
	deepPoint, err := app.ResolvePoint(ByText("Deep Action"))
	if err != nil {
		t.Fatal(err)
	}
	if pivotPoint.X <= rootPoint.X {
		t.Fatalf("pivot submenu x = %d, want > root x = %d after right-edge clamp branch", pivotPoint.X, rootPoint.X)
	}
	if deepPoint.X >= pivotPoint.X {
		t.Fatalf("deep submenu x = %d, want < pivot x = %d after mirrored left fallback", deepPoint.X, pivotPoint.X)
	}
	if deepPoint.Y >= pivotPoint.Y {
		t.Fatalf("deep submenu y = %d, want < pivot y = %d after upward clamp", deepPoint.Y, pivotPoint.Y)
	}

	popupBounds, err := app.BoundsOf(ByID("fixture-bottom-mirror-left-submenu-menu"))
	if err != nil {
		t.Fatal(err)
	}
	if popupBounds.X != 0 {
		t.Fatalf("popup hit bounds x = %d, want 0 with root anchored at left edge", popupBounds.X)
	}
	if popupBounds.Y > deepPoint.Y {
		t.Fatalf("popup hit bounds y = %d, want <= deep submenu y = %d", popupBounds.Y, deepPoint.Y)
	}
	if popupBounds.X+popupBounds.Width <= pivotPoint.X {
		t.Fatalf("popup hit bounds right edge = %d, want > pivot submenu x = %d", popupBounds.X+popupBounds.Width, pivotPoint.X)
	}
	if popupBounds.Y+popupBounds.Height > 24 {
		t.Fatalf("popup bottom edge = %d, want <= 24", popupBounds.Y+popupBounds.Height)
	}

	if err := app.AssertHit(deepPoint, ByID("fixture-bottom-mirror-left-submenu-menu")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	if err := app.Driver().Click(ByText("Deep Action")); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitIntent(meta.ActivateIntentType, 500*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.ActivateIntentType); err != nil {
		t.Fatal(err)
	}
}

func TestE2EMenuNestedSubmenusMirrorRightAndClampUpwardNearBottomLeftCorner(t *testing.T) {
	appFn, initFn, cleanupFn, meta := newBottomMirrorRightAfterLeftClampFixture()
	defer cleanupFn()

	app, err := Run(appFn,
		ui.WithSize(80, 24),
		ui.WithInit(initFn),
		ui.WithPluginSetup(func(app *framework.App) {
			menucomp.Install(app, nil)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("Operations")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Pivot")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Deep Action")); err != nil {
		t.Fatal(err)
	}

	rootPoint, err := app.ResolvePoint(ByText("Operations"))
	if err != nil {
		t.Fatal(err)
	}
	pivotPoint, err := app.ResolvePoint(ByText("Pivot"))
	if err != nil {
		t.Fatal(err)
	}
	deepPoint, err := app.ResolvePoint(ByText("Deep Action"))
	if err != nil {
		t.Fatal(err)
	}
	if pivotPoint.X >= rootPoint.X {
		t.Fatalf("pivot submenu x = %d, want < root x = %d after left-edge clamp branch", pivotPoint.X, rootPoint.X)
	}
	if deepPoint.X <= pivotPoint.X {
		t.Fatalf("deep submenu x = %d, want > pivot x = %d after mirrored right fallback", deepPoint.X, pivotPoint.X)
	}
	if deepPoint.Y >= pivotPoint.Y {
		t.Fatalf("deep submenu y = %d, want < pivot y = %d after upward clamp", deepPoint.Y, pivotPoint.Y)
	}

	popupBounds, err := app.BoundsOf(ByID("fixture-bottom-mirror-right-submenu-menu"))
	if err != nil {
		t.Fatal(err)
	}
	if popupBounds.X != 0 {
		t.Fatalf("popup hit bounds x = %d, want 0 after left-edge clamp branch", popupBounds.X)
	}
	if popupBounds.Y > deepPoint.Y {
		t.Fatalf("popup hit bounds y = %d, want <= deep submenu y = %d", popupBounds.Y, deepPoint.Y)
	}
	if popupBounds.X+popupBounds.Width <= deepPoint.X {
		t.Fatalf("popup hit bounds right edge = %d, want > deep submenu x = %d", popupBounds.X+popupBounds.Width, deepPoint.X)
	}
	if popupBounds.Y+popupBounds.Height > 24 {
		t.Fatalf("popup bottom edge = %d, want <= 24", popupBounds.Y+popupBounds.Height)
	}

	if err := app.AssertHit(deepPoint, ByID("fixture-bottom-mirror-right-submenu-menu")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	if err := app.Driver().Click(ByText("Deep Action")); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitIntent(meta.ActivateIntentType, 500*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.ActivateIntentType); err != nil {
		t.Fatal(err)
	}
}

func TestE2EMenuNestedSubmenusZigZagMirrorAndClampUpwardNearBottom(t *testing.T) {
	appFn, initFn, cleanupFn, meta := newBottomZigZagMirrorFixture()
	defer cleanupFn()

	app, err := Run(appFn,
		ui.WithSize(80, 24),
		ui.WithInit(initFn),
		ui.WithPluginSetup(func(app *framework.App) {
			menucomp.Install(app, nil)
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	if err := app.AssertVisible(ByText("Operations")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Rebound")); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertVisible(ByText("Deep Action")); err != nil {
		t.Fatal(err)
	}

	reboundPoint, err := app.ResolvePoint(ByText("Rebound"))
	if err != nil {
		t.Fatal(err)
	}
	deepPoint, err := app.ResolvePoint(ByText("Deep Action"))
	if err != nil {
		t.Fatal(err)
	}
	if deepPoint.X <= reboundPoint.X {
		t.Fatalf("deep submenu x = %d, want > rebound x = %d after mirrored right rebound", deepPoint.X, reboundPoint.X)
	}
	if deepPoint.Y >= reboundPoint.Y {
		t.Fatalf("deep submenu y = %d, want < rebound y = %d after upward clamp", deepPoint.Y, reboundPoint.Y)
	}

	popupBounds, err := app.BoundsOf(ByID("fixture-bottom-zigzag-submenu-menu"))
	if err != nil {
		t.Fatal(err)
	}
	if popupBounds.X != 0 {
		t.Fatalf("popup hit bounds x = %d, want 0 after mirrored-left branch reaches viewport edge", popupBounds.X)
	}
	if popupBounds.Y > deepPoint.Y {
		t.Fatalf("popup hit bounds y = %d, want <= deep submenu y = %d", popupBounds.Y, deepPoint.Y)
	}
	if popupBounds.X+popupBounds.Width <= deepPoint.X {
		t.Fatalf("popup hit bounds right edge = %d, want > deep submenu x = %d", popupBounds.X+popupBounds.Width, deepPoint.X)
	}
	if popupBounds.Y+popupBounds.Height > 24 {
		t.Fatalf("popup bottom edge = %d, want <= 24", popupBounds.Y+popupBounds.Height)
	}

	if err := app.AssertHit(deepPoint, ByID("fixture-bottom-zigzag-submenu-menu")); err != nil {
		t.Fatal(err)
	}

	app.ClearIntentLogs()
	if err := app.Driver().Click(ByText("Deep Action")); err != nil {
		t.Fatal(err)
	}
	if err := app.AwaitIntent(meta.ActivateIntentType, 500*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := app.AssertIntentHandled(meta.ActivateIntentType); err != nil {
		t.Fatal(err)
	}
}

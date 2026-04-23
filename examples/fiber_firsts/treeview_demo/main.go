package main

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
	treeviewcomp "github.com/wwsheng009/mint/ui/components/treeview"
)

const treeComponentID = "examples.treeview.showcase"

type AppState struct {
	Nodes              []treeviewcomp.TreeNode
	ExpandedPaths      string
	SearchText         string
	SearchMatchedPaths string
	SearchPending      bool
	SearchRequestID    int
	SearchPageSize     int
	CheckedPaths       string
	SelectionMode      string
	SelectedPath       string
	SelectedContent    string
	SelectedNodeID     int
	ShowBorder         bool
	ShowIcons          bool
	Compact            bool
	ShowLineNums       bool
	ShowScrollbar      bool
	ViewportHeight     int
	AsyncMode          string
	LazyRequests       int
	LazyResults        int
	LastAction         string
}

type ToggleBorderIntent struct{}
type ToggleIconsIntent struct{}
type ToggleCompactIntent struct{}
type ToggleLineNumbersIntent struct{}
type ToggleScrollbarIntent struct{}
type ToggleAsyncModeIntent struct{}
type ClearCheckedIntent struct{}
type ClearSearchIntent struct{}
type ResetDemoIntent struct{}
type ExpandAllDemoIntent struct{}
type CollapseAllDemoIntent struct{}
type SearchNextDemoIntent struct{}
type SearchPrevDemoIntent struct{}
type ResolveSearchIntent struct {
	RequestID    int
	Query        string
	MatchedPaths string
}

type AdjustViewportIntent struct{ Delta int }
type SetSelectionModeIntent struct{ Mode string }

func (ToggleBorderIntent) IntentType() string      { return "TreeDemoToggleBorder" }
func (ToggleIconsIntent) IntentType() string       { return "TreeDemoToggleIcons" }
func (ToggleCompactIntent) IntentType() string     { return "TreeDemoToggleCompact" }
func (ToggleLineNumbersIntent) IntentType() string { return "TreeDemoToggleLineNumbers" }
func (ToggleScrollbarIntent) IntentType() string   { return "TreeDemoToggleScrollbar" }
func (ToggleAsyncModeIntent) IntentType() string   { return "TreeDemoToggleAsyncMode" }
func (ClearCheckedIntent) IntentType() string      { return "TreeDemoClearChecked" }
func (ClearSearchIntent) IntentType() string       { return "TreeDemoClearSearch" }
func (ResetDemoIntent) IntentType() string         { return "TreeDemoReset" }
func (ExpandAllDemoIntent) IntentType() string     { return "TreeDemoExpandAll" }
func (CollapseAllDemoIntent) IntentType() string   { return "TreeDemoCollapseAll" }
func (SearchNextDemoIntent) IntentType() string    { return "TreeDemoSearchNext" }
func (SearchPrevDemoIntent) IntentType() string    { return "TreeDemoSearchPrev" }
func (ResolveSearchIntent) IntentType() string     { return "TreeDemoResolveSearch" }
func (AdjustViewportIntent) IntentType() string    { return "TreeDemoAdjustViewport" }
func (SetSelectionModeIntent) IntentType() string  { return "TreeDemoSetSelectionMode" }

func (ToggleBorderIntent) StayPressed() bool      { return true }
func (ToggleIconsIntent) StayPressed() bool       { return true }
func (ToggleCompactIntent) StayPressed() bool     { return true }
func (ToggleLineNumbersIntent) StayPressed() bool { return true }
func (ToggleScrollbarIntent) StayPressed() bool   { return true }
func (ToggleAsyncModeIntent) StayPressed() bool   { return true }
func (ClearCheckedIntent) StayPressed() bool      { return true }
func (ClearSearchIntent) StayPressed() bool       { return true }
func (ResetDemoIntent) StayPressed() bool         { return true }
func (ExpandAllDemoIntent) StayPressed() bool     { return true }
func (CollapseAllDemoIntent) StayPressed() bool   { return true }
func (SearchNextDemoIntent) StayPressed() bool    { return true }
func (SearchPrevDemoIntent) StayPressed() bool    { return true }
func (ResolveSearchIntent) StayPressed() bool     { return true }
func (AdjustViewportIntent) StayPressed() bool    { return true }
func (SetSelectionModeIntent) StayPressed() bool  { return true }

var demoStore = store.NewStore(newInitialState())
var searchRequestSeq int64

func init() {
	reducer.NewBuilder[AppState]().
		On(intent.FieldChangeIntent{}, func(s AppState, i intent.Intent) AppState {
			change, ok := i.(intent.FieldChangeIntent)
			if !ok {
				return s
			}
			switch change.Field {
			case "searchText":
				s.SearchText = change.Value
				query := strings.TrimSpace(s.SearchText)
				if query == "" {
					s.SearchMatchedPaths = ""
					s.SearchPending = false
					s.SearchRequestID = nextSearchRequestID()
					s.LastAction = "Cleared search query"
					return syncSelectedMetadata(s)
				}
				s = beginAsyncSearch(s, fmt.Sprintf("Searching %q asynchronously", query))
				return syncSelectedMetadata(s)
			case "checkedPaths":
				s.CheckedPaths = normalizeCheckedPaths(change.Value, s.SelectionMode)
				s.LastAction = fmt.Sprintf("Checked %d node(s)", len(splitList(s.CheckedPaths)))
			}
			return syncSelectedMetadata(s)
		}).
		On(treeviewcomp.NodeSelectIntent{}, func(s AppState, i intent.Intent) AppState {
			selected, ok := i.(treeviewcomp.NodeSelectIntent)
			if !ok || selected.ComponentID != treeComponentID {
				return s
			}
			s.SelectedPath = selected.Path
			s.SelectedContent = selected.Content
			s.SelectedNodeID = selected.NodeID
			s.LastAction = fmt.Sprintf("Selected %s", firstNonEmpty(selected.Path, selected.Content))
			return s
		}).
		On(treeviewcomp.NodeExpandIntent{}, func(s AppState, i intent.Intent) AppState {
			change, ok := i.(treeviewcomp.NodeExpandIntent)
			if !ok || change.ComponentID != treeComponentID {
				return s
			}
			path := resolvedNodePath(s.Nodes, change.NodeIndex, change.Path, change.NodeID)
			s.ExpandedPaths = setExpandedPath(s.ExpandedPaths, path, true)
			s.LastAction = fmt.Sprintf("Expanded %s", path)
			return syncSelectedMetadata(s)
		}).
		On(treeviewcomp.NodeCollapseIntent{}, func(s AppState, i intent.Intent) AppState {
			change, ok := i.(treeviewcomp.NodeCollapseIntent)
			if !ok || change.ComponentID != treeComponentID {
				return s
			}
			path := resolvedNodePath(s.Nodes, change.NodeIndex, change.Path, change.NodeID)
			s.ExpandedPaths = setExpandedPath(s.ExpandedPaths, path, false)
			s.LastAction = fmt.Sprintf("Collapsed %s", path)
			return syncSelectedMetadata(s)
		}).
		On(treeviewcomp.LazyLoadIntent{}, func(s AppState, i intent.Intent) AppState {
			load, ok := i.(treeviewcomp.LazyLoadIntent)
			if !ok || load.ComponentID != treeComponentID {
				return s
			}
			path := resolvedNodePath(s.Nodes, load.NodeIndex, load.Path, load.NodeID)
			s.Nodes = markNodeLoading(s.Nodes, path)
			s.LazyRequests++
			s.LastAction = fmt.Sprintf("Lazy request: %s", path)
			return syncSelectedMetadata(s)
		}).
		On(treeviewcomp.LazyLoadSuccessIntent{}, func(s AppState, i intent.Intent) AppState {
			load, ok := i.(treeviewcomp.LazyLoadSuccessIntent)
			if !ok || load.ComponentID != treeComponentID {
				return s
			}
			path := resolvedNodePath(s.Nodes, load.NodeIndex, load.Path, load.NodeID)
			s.Nodes = applyLazySuccess(s.Nodes, path, load.Children, load.Replace)
			s.LazyResults++
			s.LastAction = fmt.Sprintf("Lazy success: %s (%d child nodes)", path, len(load.Children))
			if strings.TrimSpace(s.SearchText) != "" {
				return syncSelectedMetadata(beginAsyncSearch(s, s.LastAction+" -> refresh search"))
			}
			return syncSelectedMetadata(s)
		}).
		On(treeviewcomp.LazyLoadFailureIntent{}, func(s AppState, i intent.Intent) AppState {
			load, ok := i.(treeviewcomp.LazyLoadFailureIntent)
			if !ok || load.ComponentID != treeComponentID {
				return s
			}
			path := resolvedNodePath(s.Nodes, load.NodeIndex, load.Path, load.NodeID)
			s.Nodes = applyLazyFailure(s.Nodes, path, load.Error)
			s.LazyResults++
			s.LastAction = fmt.Sprintf("Lazy failure: %s (%s)", path, strings.TrimSpace(load.Error))
			if strings.TrimSpace(s.SearchText) != "" {
				return syncSelectedMetadata(beginAsyncSearch(s, s.LastAction+" -> refresh search"))
			}
			return syncSelectedMetadata(s)
		}).
		On(ResolveSearchIntent{}, func(s AppState, i intent.Intent) AppState {
			resolved, ok := i.(ResolveSearchIntent)
			if !ok {
				return s
			}
			if resolved.RequestID != s.SearchRequestID || strings.TrimSpace(resolved.Query) != strings.TrimSpace(s.SearchText) {
				return s
			}
			s.SearchPending = false
			s.SearchMatchedPaths = resolved.MatchedPaths
			s.LastAction = fmt.Sprintf("Search resolved: %d match(es)", len(splitList(s.SearchMatchedPaths)))
			return syncSelectedMetadata(s)
		}).
		On(ToggleBorderIntent{}, func(s AppState, i intent.Intent) AppState {
			s.ShowBorder = !s.ShowBorder
			s.LastAction = fmt.Sprintf("Border = %t", s.ShowBorder)
			return s
		}).
		On(ToggleIconsIntent{}, func(s AppState, i intent.Intent) AppState {
			s.ShowIcons = !s.ShowIcons
			s.LastAction = fmt.Sprintf("Icons = %t", s.ShowIcons)
			return s
		}).
		On(ToggleCompactIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Compact = !s.Compact
			s.LastAction = fmt.Sprintf("Compact = %t", s.Compact)
			return s
		}).
		On(ToggleLineNumbersIntent{}, func(s AppState, i intent.Intent) AppState {
			s.ShowLineNums = !s.ShowLineNums
			s.LastAction = fmt.Sprintf("ShowLineNums = %t", s.ShowLineNums)
			return s
		}).
		On(ToggleScrollbarIntent{}, func(s AppState, i intent.Intent) AppState {
			s.ShowScrollbar = !s.ShowScrollbar
			s.LastAction = fmt.Sprintf("Scrollbar = %t", s.ShowScrollbar)
			return s
		}).
		On(ToggleAsyncModeIntent{}, func(s AppState, i intent.Intent) AppState {
			if s.AsyncMode == "error" {
				s.AsyncMode = "success"
			} else {
				s.AsyncMode = "error"
			}
			s.LastAction = fmt.Sprintf("Experiments async mode = %s", strings.ToUpper(s.AsyncMode))
			return s
		}).
		On(ClearCheckedIntent{}, func(s AppState, i intent.Intent) AppState {
			s.CheckedPaths = ""
			s.LastAction = "Cleared checked nodes"
			return s
		}).
		On(ClearSearchIntent{}, func(s AppState, i intent.Intent) AppState {
			s.SearchText = ""
			s.SearchMatchedPaths = ""
			s.SearchPending = false
			s.SearchRequestID = nextSearchRequestID()
			s.LastAction = "Cleared search query"
			return syncSelectedMetadata(s)
		}).
		On(AdjustViewportIntent{}, func(s AppState, i intent.Intent) AppState {
			adjust, ok := i.(AdjustViewportIntent)
			if !ok {
				return s
			}
			s.ViewportHeight = clampInt(s.ViewportHeight+adjust.Delta, 6, 16)
			s.LastAction = fmt.Sprintf("ViewportHeight = %d", s.ViewportHeight)
			return s
		}).
		On(SetSelectionModeIntent{}, func(s AppState, i intent.Intent) AppState {
			modeIntent, ok := i.(SetSelectionModeIntent)
			if !ok {
				return s
			}
			s.SelectionMode = normalizeSelectionMode(modeIntent.Mode)
			s.CheckedPaths = normalizeCheckedPaths(s.CheckedPaths, s.SelectionMode)
			s.LastAction = fmt.Sprintf("SelectionMode = %s", strings.ToUpper(s.SelectionMode))
			return s
		}).
		On(ExpandAllDemoIntent{}, func(s AppState, i intent.Intent) AppState {
			s.ExpandedPaths = strings.Join(allExpandablePaths(s.Nodes), ",")
			s.LastAction = "Expanded all tree folders"
			return syncSelectedMetadata(s)
		}).
		On(CollapseAllDemoIntent{}, func(s AppState, i intent.Intent) AppState {
			s.ExpandedPaths = ""
			s.LastAction = "Collapsed tree to root"
			return syncSelectedMetadata(s)
		}).
		On(SearchNextDemoIntent{}, func(s AppState, i intent.Intent) AppState {
			return navigateMatchInState(s, 1)
		}).
		On(SearchPrevDemoIntent{}, func(s AppState, i intent.Intent) AppState {
			return navigateMatchInState(s, -1)
		}).
		On(ResetDemoIntent{}, func(s AppState, i intent.Intent) AppState {
			return newInitialState()
		}).
		BuildAndRegister(intent.DefaultRegistry(), demoStore)
}

func main() {
	err := ui.Run(App,
		ui.WithWidth(132),
		ui.WithHeight(42),
		ui.WithTitle("Interactive TreeView Demo"),
		ui.WithPluginSetup(func(app *framework.App) {
			app.OnKeyCombo("f2", func() { ui.EmitIntentGlobal(SearchPrevDemoIntent{}) })
			app.OnKeyCombo("f3", func() { ui.EmitIntentGlobal(SearchNextDemoIntent{}) })
			app.OnKeyCombo("f4", func() { ui.EmitIntentGlobal(ExpandAllDemoIntent{}) })
			app.OnKeyCombo("f6", func() { ui.EmitIntentGlobal(CollapseAllDemoIntent{}) })
			app.OnKeyCombo("f7", func() { ui.EmitIntentGlobal(ToggleAsyncModeIntent{}) })
			app.OnKeyCombo("f8", func() { ui.EmitIntentGlobal(SetSelectionModeIntent{Mode: "none"}) })
			app.OnKeyCombo("f9", func() { ui.EmitIntentGlobal(SetSelectionModeIntent{Mode: "single"}) })
			app.OnKeyCombo("f10", func() { ui.EmitIntentGlobal(SetSelectionModeIntent{Mode: "multi"}) })
			app.OnKeyCombo("f11", func() { ui.EmitIntentGlobal(ClearCheckedIntent{}) })
			app.OnKeyCombo("f12", func() { ui.EmitIntentGlobal(ResetDemoIntent{}) })
			app.OnKeyCombo("ctrl+l", func() { ui.EmitIntentGlobal(ClearSearchIntent{}) })
		}),
	)
	if err != nil {
		panic(err)
	}
}

func App() ui.VNode {
	state := demoStore.Get()
	checkedPaths := splitList(state.CheckedPaths)
	expandedPaths := splitList(state.ExpandedPaths)
	searchMatchedPaths := splitList(state.SearchMatchedPaths)
	visible := computeVisibleEntries(state.Nodes, expandedPaths, state.SearchText, searchMatchedPaths, state.SearchPending)
	selectedIndex := visibleIndexByPath(visible, state.SelectedPath)
	matchTotal, matchSelected := matchStats(visible, state.SelectedPath)
	pageResults, matchPage, matchPageCount := matchPageEntries(visible, state.SelectedPath, state.SearchPageSize)

	return ui.NewVStack().
		SetGap(1).
		SetChildrenList([]ui.VNode{
			headerPanel(state, checkedPaths, matchTotal, matchSelected, matchPage, matchPageCount),
			controlsPanel(state),
			ui.HStackBuilder(
				ui.Flex(treePanel(state, checkedPaths, expandedPaths, searchMatchedPaths, selectedIndex), 3),
				ui.Flex(sidebar(state, checkedPaths, expandedPaths, matchTotal, matchSelected, matchPage, matchPageCount, pageResults), 2),
			).Gap(1).Stretch().Build(),
		})
}

func headerPanel(state AppState, checkedPaths []string, matchTotal, matchSelected, matchPage, matchPageCount int) ui.VNode {
	return ui.NewVStack().
		SingleBorder("Interactive TreeView Showcase").
		SetGap(0).
		SetChildrenList([]ui.VNode{
			ui.NewTextBuilder("异步受控搜索 / match 高亮分页 / 外部控制展开与匹配跳转 / 父子勾选联动 / 同步与异步 lazy load / 错误重试").Bold(true).FgColor("bright-cyan").Build(),
			ui.NewTextBuilder(fmt.Sprintf("Search=%q  Pending=%t  Matches=%d/%d  Page=%d/%d  Checked=%d  Viewport=%d  Async=%s",
				strings.TrimSpace(state.SearchText),
				state.SearchPending,
				matchSelected,
				matchTotal,
				matchPage,
				matchPageCount,
				len(checkedPaths),
				state.ViewportHeight,
				strings.ToUpper(state.AsyncMode),
			)).FgColor("bright-white").Build(),
			ui.NewTextBuilder(fmt.Sprintf("Selected=%s  Requests=%d  Results=%d  LastAction=%s",
				displaySelected(state.SelectedPath, state.SelectedContent),
				state.LazyRequests,
				state.LazyResults,
				state.LastAction,
			)).FgColor("yellow").Build(),
		})
}

func controlsPanel(state AppState) ui.VNode {
	return ui.NewVStack().
		SingleBorder("Controls").
		SetGap(0).
		SetChildrenList([]ui.VNode{
			ui.HStackBuilder(
				ui.NewTextBuilder("Search").FgColor("cyan").Build(),
				ui.NewInputBuilder().
					Placeholder("content / path / error").
					Value(state.SearchText).
					Width(34).
					ForField(intent.BindField("searchText")).
					Build(),
				ui.NewButtonBuilder("Clear").OnPress(ClearSearchIntent{}).Build(),
				ui.NewButtonBuilder("Prev Match").OnPress(SearchPrevDemoIntent{}).Build(),
				ui.NewButtonBuilder("Next Match").OnPress(SearchNextDemoIntent{}).Build(),
				ui.NewTextBuilder("Search is resolved asynchronously, then projected back as controlled matches + controlled selected index.").FgColor("bright-black").Build(),
			).Gap(1).Stretch().Build(),
			ui.HStackBuilder(
				ui.NewButtonBuilder("Expand All").OnPress(ExpandAllDemoIntent{}).Build(),
				ui.NewButtonBuilder("Collapse All").OnPress(CollapseAllDemoIntent{}).Build(),
				ui.NewButtonBuilder(fmt.Sprintf("Border:%t", state.ShowBorder)).OnPress(ToggleBorderIntent{}).Build(),
				ui.NewButtonBuilder(fmt.Sprintf("Icons:%t", state.ShowIcons)).OnPress(ToggleIconsIntent{}).Build(),
				ui.NewButtonBuilder(fmt.Sprintf("Compact:%t", state.Compact)).OnPress(ToggleCompactIntent{}).Build(),
				ui.NewButtonBuilder(fmt.Sprintf("IDs:%t", state.ShowLineNums)).OnPress(ToggleLineNumbersIntent{}).Build(),
				ui.NewButtonBuilder(fmt.Sprintf("Scrollbar:%t", state.ShowScrollbar)).OnPress(ToggleScrollbarIntent{}).Build(),
			).Gap(1).Stretch().Build(),
			ui.HStackBuilder(
				ui.NewButtonBuilder("None").OnPress(SetSelectionModeIntent{Mode: "none"}).Build(),
				ui.NewButtonBuilder("Single").OnPress(SetSelectionModeIntent{Mode: "single"}).Build(),
				ui.NewButtonBuilder("Multi").OnPress(SetSelectionModeIntent{Mode: "multi"}).Build(),
				ui.NewButtonBuilder("Clear Checked").OnPress(ClearCheckedIntent{}).Build(),
				ui.NewButtonBuilder("Height -").OnPress(AdjustViewportIntent{Delta: -1}).Build(),
				ui.NewButtonBuilder("Height +").OnPress(AdjustViewportIntent{Delta: 1}).Build(),
				ui.NewButtonBuilder(fmt.Sprintf("Async:%s", strings.ToUpper(state.AsyncMode))).OnPress(ToggleAsyncModeIntent{}).Build(),
				ui.NewButtonBuilder("Reset").OnPress(ResetDemoIntent{}).Build(),
			).Gap(1).Stretch().Build(),
		})
}

func treePanel(state AppState, checkedPaths, expandedPaths, searchMatchedPaths []string, selectedIndex int) ui.VNode {
	builder := treeviewcomp.NewBuilder().
		ComponentID(treeComponentID).
		Nodes(state.Nodes).
		SelectedIndexControlled(selectedIndex).
		ExpandedPaths(expandedPaths...).
		ViewportHeight(state.ViewportHeight).
		SearchMatchPathsControlled(searchMatchedPaths...).
		SearchPending(state.SearchPending).
		SearchPageSize(state.SearchPageSize).
		SearchQueryControlled(state.SearchText).
		ShowSearchStats(true).
		ShowBorder(state.ShowBorder).
		ShowScrollbar(state.ShowScrollbar).
		ShowIcons(state.ShowIcons).
		ShowLineNumbers(state.ShowLineNums).
		Compact(state.Compact).
		MatchStyle(style.Style{}.Foreground(style.Yellow).Bold(true)).
		SearchStatsStyle(style.Style{}.Foreground(style.BrightBlack)).
		SelectedStyle(style.Style{}.Foreground(style.Black).Background(style.BrightCyan).Bold(true)).
		RowStyleFn(demoRowStyle).
		CheckedPaths(checkedPaths...).
		SelectionForField(intent.BindField("checkedPaths")).
		OnLazyLoad(simulateLazyLoad)

	switch selectionModeValue(state.SelectionMode) {
	case treeviewcomp.SelectionSingle:
		builder.SingleSelect()
	case treeviewcomp.SelectionMultiple:
		builder.MultiSelect()
	default:
		builder.SelectionMode(treeviewcomp.SelectionNone)
	}

	return ui.NewVStack().
		SingleBorder("Tree").
		SetGap(0).
		SetChildrenList([]ui.VNode{
			ui.NewTextBuilder("Search input is async-controlled. cmd/ and services/ succeed; experiments/ follows the current async mode and can be retried with r or focused F5.").FgColor("bright-black").Build(),
			builder.Build(),
		})
}

func sidebar(state AppState, checkedPaths, expandedPaths []string, matchTotal, matchSelected, matchPage, matchPageCount int, pageResults []pagedMatchEntry) ui.VNode {
	return ui.NewVStack().
		SetGap(1).
		SetChildrenList([]ui.VNode{
			statePanel(state, checkedPaths, expandedPaths, matchTotal, matchSelected, matchPage, matchPageCount),
			searchResultsPanel(state, pageResults),
			usagePanel(),
			shortcutsPanel(),
		})
}

func statePanel(state AppState, checkedPaths, expandedPaths []string, matchTotal, matchSelected, matchPage, matchPageCount int) ui.VNode {
	checkedPreview := "[]"
	if len(checkedPaths) > 0 {
		checkedPreview = "[" + strings.Join(checkedPaths, ", ") + "]"
	}
	expandedPreview := "[]"
	if len(expandedPaths) > 0 {
		expandedPreview = "[" + strings.Join(expandedPaths, ", ") + "]"
	}

	return ui.NewVStack().
		SingleBorder("State").
		SetGap(0).
		SetChildrenList([]ui.VNode{
			ui.NewTextBuilder(fmt.Sprintf("SelectionMode: %s", strings.ToUpper(state.SelectionMode))).FgColor("cyan").Build(),
			ui.NewTextBuilder(fmt.Sprintf("Matches: %d/%d  Page: %d/%d", matchSelected, matchTotal, matchPage, matchPageCount)).FgColor("yellow").Build(),
			ui.NewTextBuilder(fmt.Sprintf("SearchPending: %t", state.SearchPending)).FgColor("bright-black").Build(),
			ui.NewTextBuilder(fmt.Sprintf("SelectedNodeID: %d", state.SelectedNodeID)).FgColor("bright-white").Build(),
			ui.NewTextBuilder(fmt.Sprintf("Selected: %s", truncateText(displaySelected(state.SelectedPath, state.SelectedContent), 52))).FgColor("bright-white").Build(),
			ui.NewTextBuilder(fmt.Sprintf("Expanded: %s", truncateText(expandedPreview, 52))).FgColor("bright-black").Build(),
			ui.NewTextBuilder(fmt.Sprintf("CheckedPaths: %s", truncateText(checkedPreview, 52))).FgColor("bright-black").Build(),
			ui.NewTextBuilder(fmt.Sprintf("Lazy Requests/Results: %d/%d", state.LazyRequests, state.LazyResults)).FgColor("bright-black").Build(),
			ui.NewTextBuilder(fmt.Sprintf("Async experiments mode: %s", strings.ToUpper(state.AsyncMode))).FgColor(asyncModeStyle(state.AsyncMode)).Build(),
		})
}

func searchResultsPanel(state AppState, pageResults []pagedMatchEntry) ui.VNode {
	children := []ui.VNode{
		ui.NewTextBuilder(fmt.Sprintf("PageSize: %d", state.SearchPageSize)).FgColor("bright-black").Build(),
	}
	switch {
	case strings.TrimSpace(state.SearchText) == "":
		children = append(children, ui.NewTextBuilder("Type in Search to start async match paging.").FgColor("bright-black").Build())
	case state.SearchPending:
		children = append(children, ui.NewTextBuilder("Searching asynchronously...").FgColor("yellow").Build())
	case len(pageResults) == 0:
		children = append(children, ui.NewTextBuilder("No matches on the current query.").FgColor("bright-black").Build())
	default:
		for _, result := range pageResults {
			prefix := "  "
			rowStyle := style.Style{}.Foreground(style.BrightWhite)
			if result.selected {
				prefix = "> "
				rowStyle = style.Style{}.Foreground(style.Black).Background(style.BrightCyan).Bold(true)
			}
			children = append(children, ui.NewTextBuilder(fmt.Sprintf("%s%02d %s", prefix, result.matchIndex, truncateText(result.path, 42))).Style(rowStyle).Build())
		}
	}

	return ui.NewVStack().
		SingleBorder("Search Results").
		SetGap(0).
		SetChildrenList(children)
}

func usagePanel() ui.VNode {
	return ui.NewVStack().
		SingleBorder("Usage").
		SetGap(0).
		SetChildrenList([]ui.VNode{
			ui.NewTextBuilder(`ComponentID("examples.treeview.showcase")`).FgColor("bright-white").Build(),
			ui.NewTextBuilder(`SelectedIndexControlled(selectedIndex)`).FgColor("bright-white").Build(),
			ui.NewTextBuilder(`ExpandedPaths(expandedPaths...)`).FgColor("bright-white").Build(),
			ui.NewTextBuilder(`SearchMatchPathsControlled(matchPaths...)`).FgColor("bright-white").Build(),
			ui.NewTextBuilder(`SearchPending(state.SearchPending) + SearchPageSize(state.SearchPageSize)`).FgColor("bright-white").Build(),
			ui.NewTextBuilder(`SearchQueryControlled(state.SearchText)`).FgColor("bright-white").Build(),
			ui.NewTextBuilder(`SelectionForField(intent.BindField("checkedPaths"))`).FgColor("bright-white").Build(),
			ui.NewTextBuilder(`OnLazyLoad(simulateLazyLoad)`).FgColor("bright-white").Build(),
		})
}

func shortcutsPanel() ui.VNode {
	return ui.NewVStack().
		SingleBorder("Shortcuts").
		SetGap(0).
		SetChildrenList([]ui.VNode{
			ui.NewTextBuilder("Tab 切换搜索框和 TreeView 焦点；TreeView 聚焦后可用 ↑ ↓ ← → / Home End / PageUp PageDown").FgColor("bright-white").Build(),
			ui.NewTextBuilder("Enter / Space 切换 checkbox；父节点勾选会联动整棵子树").FgColor("bright-white").Build(),
			ui.NewTextBuilder("r 或聚焦 TreeView 后按 F5: 重试当前 lazy/error 节点").FgColor("bright-white").Build(),
			ui.NewTextBuilder("F2/F3: 上一/下一匹配（结果页会随当前 match 自动切换）  F4/F6: 展开全部/收起全部  F7: 切换 experiments 成功或失败").FgColor("bright-black").Build(),
			ui.NewTextBuilder("F8/F9/F10: none/single/multi  F11: clear checked  F12: reset  Ctrl+L: clear search").FgColor("bright-black").Build(),
		})
}

func demoRowStyle(index int, node treeviewcomp.TreeNode) style.Style {
	switch {
	case node.LoadError != "":
		return style.Style{}.Foreground(style.BrightRed).Bold(true)
	case node.Loading:
		return style.Style{}.Foreground(style.Yellow)
	case node.Lazy:
		return style.Style{}.Foreground(style.Yellow)
	case node.NodeType == "folder":
		return style.Style{}.Foreground(style.Cyan)
	case strings.HasSuffix(node.Content, ".md"):
		return style.Style{}.Foreground(style.Green)
	case index%2 == 0:
		return style.Style{}.Foreground(style.BrightWhite)
	default:
		return style.Style{}
	}
}

func simulateLazyLoad(node treeviewcomp.TreeNode) {
	path := node.Path
	nodeID := node.NodeID

	go func() {
		delay := 120 * time.Millisecond
		if path == "workspace/services" || path == "workspace/experiments" {
			delay = 450 * time.Millisecond
		}
		time.Sleep(delay)

		mode := demoStore.Get().AsyncMode
		switch path {
		case "workspace/cmd":
			ui.EmitIntentGlobal(treeviewcomp.LazyLoadSuccessWithID(treeComponentID, -1, path, nodeID, []treeviewcomp.TreeNode{
				{Content: "main.go", NodeID: 31, NodeType: "file"},
				{Content: "treeview_demo.go", NodeID: 32, NodeType: "file"},
				{Content: "debug/", NodeID: 33, NodeType: "folder"},
				{Indent: 4, Content: "probe.go", NodeID: 34, NodeType: "file"},
			}))
		case "workspace/services":
			ui.EmitIntentGlobal(treeviewcomp.LazyLoadSuccessWithID(treeComponentID, -1, path, nodeID, []treeviewcomp.TreeNode{
				{Content: "gateway/", NodeID: 41, NodeType: "folder"},
				{Indent: 4, Content: "handler.go", NodeID: 42, NodeType: "file"},
				{Content: "billing/", NodeID: 43, NodeType: "folder"},
				{Indent: 4, Content: "reconcile.go", NodeID: 44, NodeType: "file"},
				{Content: "search/", NodeID: 45, NodeType: "folder"},
				{Indent: 4, Content: "indexer.go", NodeID: 46, NodeType: "file"},
			}))
		case "workspace/experiments":
			if mode == "error" {
				ui.EmitIntentGlobal(treeviewcomp.LazyLoadFailureWithID(treeComponentID, -1, path, nodeID, "simulated backend timeout"))
				return
			}
			ui.EmitIntentGlobal(treeviewcomp.LazyLoadSuccessWithID(treeComponentID, -1, path, nodeID, []treeviewcomp.TreeNode{
				{Content: "drafts/", NodeID: 51, NodeType: "folder"},
				{Indent: 4, Content: "notes.md", NodeID: 52, NodeType: "file"},
				{Content: "retry.log", NodeID: 53, NodeType: "file"},
			}))
		}
	}()
}

func nextSearchRequestID() int {
	return int(atomic.AddInt64(&searchRequestSeq, 1))
}

func beginAsyncSearch(state AppState, action string) AppState {
	query := strings.TrimSpace(state.SearchText)
	state.SearchMatchedPaths = ""
	state.SearchRequestID = nextSearchRequestID()
	if query == "" {
		state.SearchPending = false
		state.LastAction = action
		return state
	}
	state.SearchPending = true
	state.LastAction = action
	dispatchAsyncSearch(query, state.SearchRequestID)
	return state
}

func dispatchAsyncSearch(query string, requestID int) {
	go func() {
		delay := 180 * time.Millisecond
		if strings.Contains(strings.ToLower(query), "go") || strings.Contains(strings.ToLower(query), "service") {
			delay = 320 * time.Millisecond
		}
		time.Sleep(delay)

		nodes := demoStore.Get().Nodes
		matchedPaths := strings.Join(asyncSearchMatches(nodes, query), ",")
		ui.EmitIntentGlobal(ResolveSearchIntent{
			RequestID:    requestID,
			Query:        query,
			MatchedPaths: matchedPaths,
		})
	}()
}

func asyncSearchMatches(nodes []treeviewcomp.TreeNode, query string) []string {
	trimmedQuery := strings.TrimSpace(query)
	if trimmedQuery == "" {
		return nil
	}
	matches := make([]string, 0, len(nodes))
	for _, node := range nodes {
		if node.Path == "" {
			continue
		}
		if nodeMatches(node, trimmedQuery) {
			matches = append(matches, node.Path)
		}
	}
	return matches
}

func newInitialState() AppState {
	state := AppState{
		Nodes:              baseTreeNodes(),
		ExpandedPaths:      "workspace,workspace/pkg,workspace/pkg/treeview,workspace/docs",
		SearchText:         "",
		SearchMatchedPaths: "",
		SearchPending:      false,
		SearchRequestID:    0,
		SearchPageSize:     2,
		CheckedPaths:       "workspace/pkg/treeview/selection.go",
		SelectionMode:      "multi",
		SelectedPath:       "workspace/README.md",
		SelectedContent:    "README.md",
		SelectedNodeID:     2,
		ShowBorder:         true,
		ShowIcons:          true,
		Compact:            false,
		ShowLineNums:       false,
		ShowScrollbar:      true,
		ViewportHeight:     11,
		AsyncMode:          "error",
		LazyRequests:       0,
		LazyResults:        0,
		LastAction:         "Ready",
	}
	return syncSelectedMetadata(state)
}

func baseTreeNodes() []treeviewcomp.TreeNode {
	return []treeviewcomp.TreeNode{
		{Content: "workspace/", Path: "workspace", NodeID: 1, NodeType: "folder"},
		{Indent: 4, Content: "README.md", Path: "workspace/README.md", NodeID: 2, NodeType: "file"},
		{Indent: 4, Content: "cmd/", Path: "workspace/cmd", NodeID: 3, NodeType: "folder", Lazy: true},
		{Indent: 4, Content: "pkg/", Path: "workspace/pkg", NodeID: 4, NodeType: "folder"},
		{Indent: 8, Content: "treeview/", Path: "workspace/pkg/treeview", NodeID: 5, NodeType: "folder"},
		{Indent: 12, Content: "builder.go", Path: "workspace/pkg/treeview/builder.go", NodeID: 6, NodeType: "file"},
		{Indent: 12, Content: "instance.go", Path: "workspace/pkg/treeview/instance.go", NodeID: 7, NodeType: "file"},
		{Indent: 12, Content: "selection.go", Path: "workspace/pkg/treeview/selection.go", NodeID: 8, NodeType: "file"},
		{Indent: 4, Content: "services/", Path: "workspace/services", NodeID: 9, NodeType: "folder", Lazy: true},
		{Indent: 4, Content: "experiments/", Path: "workspace/experiments", NodeID: 10, NodeType: "folder", Lazy: true},
		{Indent: 4, Content: "docs/", Path: "workspace/docs", NodeID: 11, NodeType: "folder"},
		{Indent: 8, Content: "architecture.md", Path: "workspace/docs/architecture.md", NodeID: 12, NodeType: "file"},
		{Indent: 8, Content: "keyboard-map.md", Path: "workspace/docs/keyboard-map.md", NodeID: 13, NodeType: "file"},
	}
}

type demoEntry struct {
	node        treeviewcomp.TreeNode
	depth       int
	path        string
	match       bool
	hasChildren bool
}

type pagedMatchEntry struct {
	matchIndex int
	path       string
	selected   bool
}

func computeVisibleEntries(nodes []treeviewcomp.TreeNode, expandedPaths []string, query string, matchedPaths []string, pending bool) []demoEntry {
	entries := buildEntries(nodes)
	expanded := make(map[string]bool, len(expandedPaths))
	for _, path := range expandedPaths {
		if path != "" {
			expanded[path] = true
		}
	}
	matchesByPath := make(map[string]bool, len(matchedPaths))
	for _, path := range matchedPaths {
		if path != "" {
			matchesByPath[path] = true
		}
	}

	trimmedQuery := strings.TrimSpace(query)
	filterActive := trimmedQuery != "" && !pending
	include := make([]bool, len(entries))
	match := make([]bool, len(entries))
	if filterActive {
		stack := make([]int, 0, 8)
		for idx, entry := range entries {
			for len(stack) > 0 && entries[stack[len(stack)-1]].depth >= entry.depth {
				stack = stack[:len(stack)-1]
			}
			if matchesByPath[entry.path] {
				match[idx] = true
				include[idx] = true
				for _, ancestor := range stack {
					include[ancestor] = true
				}
			}
			if entry.hasChildren {
				stack = append(stack, idx)
			}
		}
	}

	type stackItem struct {
		depth    int
		expanded bool
		visible  bool
	}
	stack := make([]stackItem, 0, 8)
	visible := make([]demoEntry, 0, len(entries))
	for idx, entry := range entries {
		for len(stack) > 0 && stack[len(stack)-1].depth >= entry.depth {
			stack = stack[:len(stack)-1]
		}
		isVisible := true
		if len(stack) > 0 {
			top := stack[len(stack)-1]
			isVisible = top.visible && top.expanded
		}
		if filterActive && !include[idx] {
			isVisible = false
		}
		if isVisible {
			entry.match = filterActive && match[idx]
			visible = append(visible, entry)
		}
		if entry.hasChildren {
			isExpanded := expanded[entry.path]
			if filterActive && include[idx] {
				isExpanded = true
			}
			stack = append(stack, stackItem{depth: entry.depth, expanded: isExpanded, visible: isVisible})
		}
	}
	return visible
}

func matchPageEntries(visible []demoEntry, selectedPath string, pageSize int) ([]pagedMatchEntry, int, int) {
	matches := make([]pagedMatchEntry, 0, len(visible))
	selectedMatch := 0
	for _, entry := range visible {
		if !entry.match {
			continue
		}
		matches = append(matches, pagedMatchEntry{
			matchIndex: len(matches) + 1,
			path:       entry.path,
			selected:   entry.path == selectedPath,
		})
		if entry.path == selectedPath {
			selectedMatch = len(matches)
		}
	}
	if len(matches) == 0 {
		return nil, 0, 0
	}
	if pageSize <= 0 || pageSize > len(matches) {
		pageSize = len(matches)
	}
	pageCount := (len(matches) + pageSize - 1) / pageSize
	page := 1
	if selectedMatch > 0 {
		page = (selectedMatch-1)/pageSize + 1
	}
	start := (page - 1) * pageSize
	end := min(start+pageSize, len(matches))
	return matches[start:end], page, pageCount
}

func buildEntries(nodes []treeviewcomp.TreeNode) []demoEntry {
	entries := make([]demoEntry, len(nodes))
	for i, node := range nodes {
		depth := nodeDepth(node)
		hasDescendants := i+1 < len(nodes) && nodeDepth(nodes[i+1]) > depth
		hasChildren := hasDescendants || node.Lazy
		entries[i] = demoEntry{
			node:        node,
			depth:       depth,
			path:        node.Path,
			hasChildren: hasChildren,
		}
	}
	return entries
}

func navigateMatchInState(s AppState, direction int) AppState {
	if direction == 0 {
		return s
	}
	if strings.TrimSpace(s.SearchText) == "" {
		s.LastAction = "Search navigation ignored: empty query"
		return s
	}
	if s.SearchPending {
		s.LastAction = "Search navigation ignored: async search pending"
		return s
	}
	visible := computeVisibleEntries(s.Nodes, splitList(s.ExpandedPaths), s.SearchText, splitList(s.SearchMatchedPaths), s.SearchPending)
	if len(visible) == 0 {
		return s
	}
	start := visibleIndexByPath(visible, s.SelectedPath)
	if start < 0 {
		start = 0
	}
	for step := 1; step <= len(visible); step++ {
		idx := start + step*direction
		for idx < 0 {
			idx += len(visible)
		}
		idx %= len(visible)
		if visible[idx].match {
			s.SelectedPath = visible[idx].path
			s.LastAction = fmt.Sprintf("Moved to match %s", visible[idx].path)
			return syncSelectedMetadata(s)
		}
	}
	s.LastAction = "No search matches available"
	return s
}

func matchStats(visible []demoEntry, selectedPath string) (total int, selected int) {
	for _, entry := range visible {
		if !entry.match {
			continue
		}
		total++
		if entry.path == selectedPath {
			selected = total
		}
	}
	return total, selected
}

func visibleIndexByPath(entries []demoEntry, path string) int {
	if path == "" {
		if len(entries) == 0 {
			return -1
		}
		return 0
	}
	for index, entry := range entries {
		if entry.path == path {
			return index
		}
	}
	return -1
}

func allExpandablePaths(nodes []treeviewcomp.TreeNode) []string {
	entries := buildEntries(nodes)
	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.hasChildren && entry.path != "" {
			paths = append(paths, entry.path)
		}
	}
	return paths
}

func setExpandedPath(raw, path string, expanded bool) string {
	if path == "" {
		return raw
	}
	paths := splitList(raw)
	index := -1
	for i, candidate := range paths {
		if candidate == path {
			index = i
			break
		}
	}
	if expanded {
		if index >= 0 {
			return strings.Join(paths, ",")
		}
		paths = append(paths, path)
		return strings.Join(paths, ",")
	}
	if index < 0 {
		return strings.Join(paths, ",")
	}
	return strings.Join(append(paths[:index], paths[index+1:]...), ",")
}

func applyLazySuccess(nodes []treeviewcomp.TreeNode, path string, children []treeviewcomp.TreeNode, replace bool) []treeviewcomp.TreeNode {
	index := nodeIndexByPath(nodes, path)
	if index < 0 {
		return nodes
	}
	result := append([]treeviewcomp.TreeNode(nil), nodes...)
	if replace {
		result = removeDescendants(result, index)
	}
	parent := result[index]
	depth := nodeDepth(parent)
	insertAt := index + 1
	for insertAt < len(result) && nodeDepth(result[insertAt]) > depth {
		insertAt++
	}
	normalized := normalizeLazyChildren(parent, children)
	if len(normalized) > 0 {
		result = append(result[:insertAt], append(normalized, result[insertAt:]...)...)
	}
	result[index].Lazy = false
	result[index].Loading = false
	result[index].LoadError = ""
	return result
}

func applyLazyFailure(nodes []treeviewcomp.TreeNode, path, err string) []treeviewcomp.TreeNode {
	index := nodeIndexByPath(nodes, path)
	if index < 0 {
		return nodes
	}
	result := append([]treeviewcomp.TreeNode(nil), nodes...)
	result[index].Loading = false
	result[index].LoadError = err
	return result
}

func markNodeLoading(nodes []treeviewcomp.TreeNode, path string) []treeviewcomp.TreeNode {
	index := nodeIndexByPath(nodes, path)
	if index < 0 {
		return nodes
	}
	result := append([]treeviewcomp.TreeNode(nil), nodes...)
	result[index].Loading = true
	result[index].LoadError = ""
	return result
}

func removeDescendants(nodes []treeviewcomp.TreeNode, index int) []treeviewcomp.TreeNode {
	if index < 0 || index >= len(nodes)-1 {
		return nodes
	}
	depth := nodeDepth(nodes[index])
	end := index + 1
	for end < len(nodes) && nodeDepth(nodes[end]) > depth {
		end++
	}
	return append(nodes[:index+1], nodes[end:]...)
}

func normalizeLazyChildren(parent treeviewcomp.TreeNode, children []treeviewcomp.TreeNode) []treeviewcomp.TreeNode {
	if len(children) == 0 {
		return nil
	}
	parentDepth := nodeDepth(parent)
	parentIndent := parentDepth * 4
	baseIndent := (parentDepth + 1) * 4
	normalized := make([]treeviewcomp.TreeNode, 0, len(children))
	for _, child := range children {
		next := child
		if next.Indent <= parentIndent {
			next.Indent = baseIndent + next.Indent
		} else if next.Indent < baseIndent {
			next.Indent = baseIndent
		}
		if next.Path == "" {
			segment := strings.TrimSuffix(next.Content, "/")
			if segment == "" {
				segment = next.Content
			}
			if parent.Path != "" {
				next.Path = parent.Path + "/" + segment
			} else {
				next.Path = segment
			}
		}
		normalized = append(normalized, next)
	}
	return normalized
}

func syncSelectedMetadata(state AppState) AppState {
	// First, check if the selected path is in the nodes array
	index := nodeIndexByPath(state.Nodes, state.SelectedPath)
	if index < 0 {
		// Path doesn't exist in nodes, reset to first visible node
		return resetToFirstVisible(state)
	}

	// Update content and nodeID
	state.SelectedContent = state.Nodes[index].Content
	state.SelectedNodeID = state.Nodes[index].NodeID

	// Now check if the node is currently visible
	visible := computeVisibleEntries(state.Nodes, splitList(state.ExpandedPaths), state.SearchText, splitList(state.SearchMatchedPaths), state.SearchPending)
	if strings.TrimSpace(state.SearchText) != "" && !state.SearchPending {
		return syncSearchSelection(state, visible)
	}
	visibleIdx := visibleIndexByPath(visible, state.SelectedPath)

	if visibleIdx >= 0 {
		// Node is visible, all good
		return state
	}

	// Node is not visible - find a visible ancestor or fallback to first visible node
	return findVisibleAncestorOrFirst(state, visible)
}

func resetToFirstVisible(state AppState) AppState {
	visible := computeVisibleEntries(state.Nodes, splitList(state.ExpandedPaths), state.SearchText, splitList(state.SearchMatchedPaths), state.SearchPending)
	if len(visible) == 0 {
		// No visible nodes, clear selection
		state.SelectedPath = ""
		state.SelectedContent = ""
		state.SelectedNodeID = 0
		return state
	}

	// Select the first visible node
	state.SelectedPath = visible[0].path
	state.SelectedContent = visible[0].node.Content
	state.SelectedNodeID = visible[0].node.NodeID
	return state
}

func findVisibleAncestorOrFirst(state AppState, visible []demoEntry) AppState {
	if len(visible) == 0 {
		// No visible nodes at all
		state.SelectedPath = ""
		state.SelectedContent = ""
		state.SelectedNodeID = 0
		return state
	}

	// Try to find a visible ancestor
	pathParts := strings.Split(state.SelectedPath, "/")
	for i := len(pathParts) - 1; i >= 0; i-- {
		ancestorPath := strings.Join(pathParts[:i], "/")
		if ancestorPath == "" {
			continue
		}

		// Check if this ancestor is visible
		for _, entry := range visible {
			if entry.path == ancestorPath {
				state.SelectedPath = entry.path
				state.SelectedContent = entry.node.Content
				state.SelectedNodeID = entry.node.NodeID
				return state
			}
		}
	}

	// No visible ancestor found, fallback to first visible node
	state.SelectedPath = visible[0].path
	state.SelectedContent = visible[0].node.Content
	state.SelectedNodeID = visible[0].node.NodeID
	return state
}

func syncSearchSelection(state AppState, visible []demoEntry) AppState {
	if len(visible) == 0 {
		state.SelectedPath = ""
		state.SelectedContent = ""
		state.SelectedNodeID = 0
		return state
	}
	for _, entry := range visible {
		if entry.path == state.SelectedPath && entry.match {
			state.SelectedContent = entry.node.Content
			state.SelectedNodeID = entry.node.NodeID
			return state
		}
	}
	for _, entry := range visible {
		if entry.match {
			state.SelectedPath = entry.path
			state.SelectedContent = entry.node.Content
			state.SelectedNodeID = entry.node.NodeID
			return state
		}
	}
	state.SelectedPath = ""
	state.SelectedContent = ""
	state.SelectedNodeID = 0
	return state
}

func resolvedNodePath(nodes []treeviewcomp.TreeNode, nodeIndex int, path string, nodeID int) string {
	if path != "" {
		return path
	}
	if nodeIndex >= 0 && nodeIndex < len(nodes) {
		return nodes[nodeIndex].Path
	}
	if nodeID != 0 {
		for _, node := range nodes {
			if node.NodeID == nodeID {
				return node.Path
			}
		}
	}
	return ""
}

func nodeIndexByPath(nodes []treeviewcomp.TreeNode, path string) int {
	if path == "" {
		return -1
	}
	for index, node := range nodes {
		if node.Path == path {
			return index
		}
	}
	return -1
}

func nodeMatches(node treeviewcomp.TreeNode, query string) bool {
	if query == "" {
		return false
	}
	lower := strings.ToLower(query)
	if strings.Contains(strings.ToLower(node.Content), lower) {
		return true
	}
	if node.Path != "" && strings.Contains(strings.ToLower(node.Path), lower) {
		return true
	}
	if node.LoadError != "" && strings.Contains(strings.ToLower(node.LoadError), lower) {
		return true
	}
	return false
}

func selectionModeValue(raw string) treeviewcomp.SelectionMode {
	switch normalizeSelectionMode(raw) {
	case "single":
		return treeviewcomp.SelectionSingle
	case "multi":
		return treeviewcomp.SelectionMultiple
	default:
		return treeviewcomp.SelectionNone
	}
}

func normalizeSelectionMode(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "single":
		return "single"
	case "multi", "multiple":
		return "multi"
	default:
		return "none"
	}
}

func splitList(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		values = append(values, value)
	}
	return values
}

func normalizeCheckedPaths(raw, mode string) string {
	paths := splitList(raw)
	switch normalizeSelectionMode(mode) {
	case "none":
		return ""
	case "single":
		if len(paths) == 0 {
			return ""
		}
		return paths[0]
	default:
		return strings.Join(paths, ",")
	}
}

func displaySelected(path, content string) string {
	if path != "" {
		return path
	}
	if content != "" {
		return content
	}
	return "(none)"
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func asyncModeStyle(mode string) string {
	if mode == "error" {
		return "bright-red"
	}
	return "green"
}

func nodeDepth(node treeviewcomp.TreeNode) int {
	if node.Indent <= 0 {
		return 0
	}
	return node.Indent / 4
}

func clampInt(value, minimum, maximum int) int {
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}

func truncateText(text string, max int) string {
	if max <= 0 || len(text) <= max {
		return text
	}
	if max <= 3 {
		return text[:max]
	}
	return text[:max-3] + "..."
}

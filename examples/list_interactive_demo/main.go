package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
)

type AppState struct {
	SelectedIndex   string
	ViewportHeight  int
	ShowBorder      bool
	ShowSeparator   bool
	ShowScrollbar   bool
	SelectionEvents int
	LastAction      string
}

type ToggleBorderIntent struct{}
type ToggleSeparatorIntent struct{}
type ToggleScrollbarIntent struct{}
type ResetDemoIntent struct{}

type AdjustViewportIntent struct {
	Delta int
}

func (ToggleBorderIntent) IntentType() string    { return "ListDemoToggleBorder" }
func (ToggleSeparatorIntent) IntentType() string { return "ListDemoToggleSeparator" }
func (ToggleScrollbarIntent) IntentType() string { return "ListDemoToggleScrollbar" }
func (ResetDemoIntent) IntentType() string       { return "ListDemoReset" }
func (AdjustViewportIntent) IntentType() string  { return "ListDemoAdjustViewport" }

func (ToggleBorderIntent) StayPressed() bool    { return true }
func (ToggleSeparatorIntent) StayPressed() bool { return true }
func (ToggleScrollbarIntent) StayPressed() bool { return true }
func (ResetDemoIntent) StayPressed() bool       { return true }
func (AdjustViewportIntent) StayPressed() bool  { return true }

var demoRows = generateRows()

var demoStore = store.NewStore(AppState{
	SelectedIndex:   "0",
	ViewportHeight:  12,
	ShowBorder:      true,
	ShowSeparator:   true,
	ShowScrollbar:   true,
	SelectionEvents: 0,
	LastAction:      "Ready",
})

func init() {
	reducer.NewBuilder[AppState]().
		On(intent.FieldChangeIntent{}, func(s AppState, i intent.Intent) AppState {
			fieldChange, ok := i.(intent.FieldChangeIntent)
			if !ok {
				return s
			}
			if fieldChange.Field != "selectedIndex" {
				return s
			}
			s.SelectedIndex = fieldChange.Value
			s.SelectionEvents++
			s.LastAction = fmt.Sprintf("Selected row #%s", fieldChange.Value)
			return s
		}).
		On(ToggleBorderIntent{}, func(s AppState, i intent.Intent) AppState {
			s.ShowBorder = !s.ShowBorder
			s.LastAction = fmt.Sprintf("Border = %t", s.ShowBorder)
			return s
		}).
		On(ToggleSeparatorIntent{}, func(s AppState, i intent.Intent) AppState {
			s.ShowSeparator = !s.ShowSeparator
			s.LastAction = fmt.Sprintf("Separator = %t", s.ShowSeparator)
			return s
		}).
		On(ToggleScrollbarIntent{}, func(s AppState, i intent.Intent) AppState {
			s.ShowScrollbar = !s.ShowScrollbar
			s.LastAction = fmt.Sprintf("Scrollbar = %t", s.ShowScrollbar)
			return s
		}).
		On(AdjustViewportIntent{}, func(s AppState, i intent.Intent) AppState {
			adjust, ok := i.(AdjustViewportIntent)
			if !ok {
				return s
			}
			s.ViewportHeight = clamp(s.ViewportHeight+adjust.Delta, 5, 16)
			s.LastAction = fmt.Sprintf("ViewportHeight = %d", s.ViewportHeight)
			return s
		}).
		On(ResetDemoIntent{}, func(s AppState, i intent.Intent) AppState {
			s.SelectedIndex = "0"
			s.ViewportHeight = 10
			s.ShowBorder = true
			s.ShowSeparator = true
			s.ShowScrollbar = true
			s.LastAction = "Reset demo state"
			return s
		}).
		BuildAndRegister(intent.DefaultRegistry(), demoStore)
}

func main() {
	err := ui.Run(App,
		ui.WithWidth(96),
		ui.WithHeight(34),
		ui.WithTitle("Interactive List Demo"),
		ui.WithPluginSetup(func(app *framework.App) {
			app.OnKeyCombo("f2", func() { ui.EmitIntentGlobal(ToggleBorderIntent{}) })
			app.OnKeyCombo("f3", func() { ui.EmitIntentGlobal(ToggleSeparatorIntent{}) })
			app.OnKeyCombo("f4", func() { ui.EmitIntentGlobal(ToggleScrollbarIntent{}) })
			app.OnKeyCombo("f5", func() { ui.EmitIntentGlobal(AdjustViewportIntent{Delta: -1}) })
			app.OnKeyCombo("f6", func() { ui.EmitIntentGlobal(AdjustViewportIntent{Delta: 1}) })
			app.OnKeyCombo("f7", func() { ui.EmitIntentGlobal(ResetDemoIntent{}) })
		}),
	)
	if err != nil {
		panic(err)
	}
}

func App() ui.VNode {
	state := demoStore.Get()
	selected := selectedIndex(state.SelectedIndex)
	page, totalPages := pageInfo(selected, state.ViewportHeight, len(demoRows))
	previewVisible := clamp(min(8, state.ViewportHeight), 5, 8)
	previewOffset := centeredOffset(selected, previewVisible, len(demoRows))

	return ui.NewVStack().
		SetGap(1).
		SetChildrenList([]ui.VNode{
			headerPanel(state, selected, page, totalPages),
			ui.HStackBuilder(
				ui.Flex(leftPane(state, selected, page, totalPages, previewVisible, previewOffset), 3),
				ui.Flex(rightPane(state, selected, page, totalPages, previewVisible, previewOffset), 2),
			).Gap(1).Stretch().Build(),
		})
}

func headerPanel(state AppState, selected, page, totalPages int) ui.VNode {
	return ui.NewVStack().
		SingleBorder("Interactive List").
		SetGap(0).
		SetChildrenList([]ui.VNode{
			ui.NewTextBuilder("长列表 / 滚动条 / 翻页 / 点击选中 / VirtualList 镜像").Bold(true).FgColor("bright-cyan").Build(),
			ui.NewTextBuilder(fmt.Sprintf("Rows=%d  Selected=%d  Page=%d/%d  Viewport=%d  Events=%d",
				len(demoRows), selected, page, totalPages, state.ViewportHeight, state.SelectionEvents)).
				FgColor("bright-white").
				Build(),
			ui.NewTextBuilder(fmt.Sprintf("Border=%t  Separator=%t  Scrollbar=%t  LastAction=%s",
				state.ShowBorder, state.ShowSeparator, state.ShowScrollbar, state.LastAction)).
				FgColor("bright-black").
				Build(),
		})
}

func interactiveList(state AppState, selected int) ui.VNode {
	return ui.List().
		Header("ID    LVL   MODULE    SUMMARY").
		Rows(demoRows).
		ViewportHeight(state.ViewportHeight).
		SelectedIndex(selected).
		ShowBorder(state.ShowBorder).
		ShowSeparator(state.ShowSeparator).
		ShowScrollbar(state.ShowScrollbar).
		SelectedStyle(style.Style{}.Foreground(style.Black).Background(style.BrightCyan).Bold(true)).
		RowStyleFn(rowStyleForDemo).
		ForField(intent.BindField("selectedIndex")).
		Build()
}

func detailPanel(state AppState, selected, page, totalPages int) ui.VNode {
	selectedRow := ""
	if selected >= 0 && selected < len(demoRows) {
		selectedRow = truncateText(demoRows[selected], 44)
	}
	return ui.NewVStack().
		SingleBorder("Selection Details").
		SetGap(0).
		SetChildrenList([]ui.VNode{
			ui.NewTextBuilder("当前 list 已支持：键盘翻页、Home/End、鼠标点击选中、FieldChangeIntent 同步").FgColor("green").Build(),
			ui.NewTextBuilder(fmt.Sprintf("Selected row: %d / %d", selected, len(demoRows)-1)).FgColor("yellow").Build(),
			ui.NewTextBuilder(fmt.Sprintf("Current page: %d / %d", page, totalPages)).FgColor("bright-white").Build(),
			ui.NewTextBuilder(selectedRow).FgColor("bright-black").Build(),
		})
}

func leftPane(state AppState, selected, page, totalPages, previewVisible, previewOffset int) ui.VNode {
	return ui.NewVStack().
		SetGap(1).
		SetChildrenList([]ui.VNode{
			interactiveList(state, selected),
			navigationPanel(selected, page, totalPages, previewVisible, previewOffset),
			helpPanel(),
		})
}

func rightPane(state AppState, selected, page, totalPages, previewVisible, previewOffset int) ui.VNode {
	return ui.NewVStack().
		SetGap(1).
		SetChildrenList([]ui.VNode{
			detailPanel(state, selected, page, totalPages),
			virtualMirrorPanel(selected, previewVisible, previewOffset),
		})
}

func navigationPanel(selected, page, totalPages, previewVisible, previewOffset int) ui.VNode {
	return ui.NewVStack().
		SingleBorder("Navigation").
		SetGap(0).
		SetChildrenList([]ui.VNode{
			ui.NewTextBuilder(fmt.Sprintf("Page=%d/%d  Selected=%d", page, totalPages, selected)).FgColor("bright-white").Build(),
			ui.NewTextBuilder(fmt.Sprintf("Mirror offset=%d  visible=%d", previewOffset, previewVisible)).FgColor("bright-black").Build(),
			ui.NewTextBuilder("↑ ↓ / PageUp PageDown / Home End / 鼠标点击").FgColor("cyan").Build(),
		})
}

func virtualMirrorPanel(selected, visible, offset int) ui.VNode {
	return ui.NewVStack().
		SingleBorder("VirtualList Mirror").
		SetGap(0).
		SetChildrenList([]ui.VNode{
			ui.NewTextBuilder(fmt.Sprintf("镜像当前选择窗口：offset=%d visible=%d", offset, visible)).FgColor("bright-black").Build(),
			ui.NewVirtualListBuilder().
				Items(demoRows).
				ItemCount(len(demoRows)).
				Width(36).
				Height(visible + 2).
				ItemHeight(1).
				VisibleCount(visible).
				ScrollOffset(offset).
				SelectedIndex(selected).
				ListStyle(style.Style{}.Foreground(style.BrightWhite)).
				SelectedStyle(style.Style{}.Foreground(style.Black).Background(style.Yellow).Bold(true)).
				Build(),
		})
}

func helpPanel() ui.VNode {
	return ui.NewVStack().
		SingleBorder("Shortcuts").
		SetGap(0).
		SetChildrenList([]ui.VNode{
			ui.NewTextBuilder("Tab 一次确保焦点落到 list；随后使用 ↑ ↓ PageUp PageDown Home End").FgColor("bright-white").Build(),
			ui.NewTextBuilder("鼠标点击任意可见行会同步选中，并驱动下方 VirtualList 镜像窗口").FgColor("bright-white").Build(),
			ui.NewTextBuilder("F2 border  F3 separator  F4 scrollbar  F5 viewport-  F6 viewport+  F7 reset").FgColor("bright-black").Build(),
		})
}

func generateRows() []string {
	levels := []string{"INFO", "WARN", "ERROR", "DEBUG"}
	modules := []string{"auth", "billing", "cache", "search", "notify", "worker"}
	rows := make([]string, 480)
	for index := range rows {
		level := levels[index%len(levels)]
		module := modules[(index/7)%len(modules)]
		rows[index] = fmt.Sprintf("%04d  %-5s %-8s Request %03d | latency=%3dms | owner=team-%d",
			index,
			level,
			module,
			index%127,
			18+(index*13)%220,
			index%9,
		)
	}
	return rows
}

func rowStyleForDemo(index int, row string) style.Style {
	switch {
	case strings.Contains(row, "ERROR"):
		return style.Style{}.Foreground(style.BrightRed).Bold(true)
	case strings.Contains(row, "WARN"):
		return style.Style{}.Foreground(style.Yellow)
	case strings.Contains(row, "DEBUG"):
		return style.Style{}.Foreground(style.BrightBlack)
	case index%10 == 0:
		return style.Style{}.Foreground(style.Cyan)
	default:
		return style.Style{}
	}
}

func selectedIndex(raw string) int {
	index, err := strconv.Atoi(raw)
	if err != nil {
		return 0
	}
	return clamp(index, 0, len(demoRows)-1)
}

func pageInfo(selected, viewport, total int) (int, int) {
	if total <= 0 {
		return 0, 0
	}
	size := max(1, viewport)
	totalPages := (total + size - 1) / size
	return selected/size + 1, totalPages
}

func centeredOffset(selected, visible, total int) int {
	if total <= visible {
		return 0
	}
	offset := selected - visible/2
	return clamp(offset, 0, total-visible)
}

func min(left, right int) int {
	if left < right {
		return left
	}
	return right
}

func max(left, right int) int {
	if left > right {
		return left
	}
	return right
}

func clamp(value, lower, upper int) int {
	if value < lower {
		return lower
	}
	if value > upper {
		return upper
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

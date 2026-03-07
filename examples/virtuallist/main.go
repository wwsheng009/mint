package main

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/ui"
)

// ============================================================================
// AppState - 定义应用状态
// ============================================================================

type AppState struct {
	Offset   int // 滚动偏移量
	Selected int // 选中的项目索引
}

// ============================================================================
// Intent Types
// ============================================================================

type ScrollUpIntent struct{}
func (ScrollUpIntent) IntentType() string { return "ScrollUp" }
func (ScrollUpIntent) StayPressed() bool  { return true }

type ScrollDownIntent struct{}
func (ScrollDownIntent) IntentType() string { return "ScrollDown" }
func (ScrollDownIntent) StayPressed() bool  { return true }

// ============================================================================
// 生成大型列表数据
// ============================================================================

func generateItems() []string {
	items := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		items[i] = fmt.Sprintf("Item #%d - This is a long description text", i+1)
	}
	return items
}

// ============================================================================
// Store 初始化
// ============================================================================

var virtualListStore = store.NewStore(AppState{
	Offset:   0,
	Selected: -1,
})

// ============================================================================
// Reducer 注册
// ============================================================================

func init() {
	reducer.NewBuilder[AppState]().
		On(ScrollUpIntent{}, func(s AppState, i intent.Intent) AppState {
			newOffset := s.Offset - 5
			if newOffset < 0 {
				newOffset = 0
			}
			s.Offset = newOffset
			return s
		}).
		On(ScrollDownIntent{}, func(s AppState, i intent.Intent) AppState {
			maxOffset := len(generateItems()) - 10
			if maxOffset < 0 {
				maxOffset = 0
			}
			newOffset := s.Offset + 5
			if newOffset > maxOffset {
				newOffset = maxOffset
			}
			s.Offset = newOffset
			return s
		}).
		BuildAndRegister(intent.DefaultRegistry(), virtualListStore)
}

// ============================================================================
// Main
// ============================================================================

func main() {
	items := generateItems()

	ui.Run(func() ui.VNode {
		// ✅ 订阅 offset 和 selected 状态
		offset := ui.UseStoreSelector(virtualListStore, func(s AppState) int { return s.Offset })
		selected := ui.UseStoreSelector(virtualListStore, func(s AppState) int { return s.Selected })

		return ui.VStack(
			ui.NewTextBuilder("Virtual List Demo").Bold(true).FgColor("cyan").Build(),
			ui.Text(""),
			ui.NewTextBuilder(fmt.Sprintf("Items: %d | Offset: %d | Selected: %d",
				len(items), offset, selected)).FgColor("gray").Build(),
			ui.Text(""),
			ui.HStack(
				ui.NewButtonBuilder(" Scroll Up ").
					OnPress(ScrollUpIntent{}).
					Build(),
				ui.NewButtonBuilder(" Scroll Down ").
					OnPress(ScrollDownIntent{}).
					Build(),
			),
			ui.Text(""),
			ui.NewTextBuilder("────────────────────────────").FgColor("blue").Build(),
			ui.Text(""),
			// Virtual list - only renders visible items
			ui.NewVirtualListBuilder().
				Items(items).
				ItemHeight(1).
				VisibleCount(10).
				Height(10).
				ScrollOffset(offset).
				SelectedIndex(selected).
				Build(),
			ui.Text(""),
			ui.NewTextBuilder("────────────────────────────").FgColor("blue").Build(),
			ui.Text(""),
			ui.NewTextBuilder("Tab to buttons, Enter to scroll").FgColor("gray").Build(),
		)
	},
		ui.WithWidth(60),
		ui.WithHeight(25),
		ui.WithTitle("Virtual List Demo (Store 模式)"),
	)
}

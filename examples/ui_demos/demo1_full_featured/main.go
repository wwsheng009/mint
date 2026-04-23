// Demo 1: Full-Featured Demo App (Store 模式)
//
// This demo demonstrates the complete TUI engine architecture, covering:
// - Declarative components
// - State system (Store + Reducer)
// - Layout system (Flex, VStack, HStack, Table)
// - Modal (Layer) - Using Layer system
// - Input with Focus management
// - Theme system with semantic colors
// - Button variants (Primary, Secondary, Danger, Success)
// - Scroll containers
// - VirtualList for large data
// - Event handling
// - Animation
//
// This is an integration acceptance test for the UI Runtime.
//
// Based on: framework/docs/ui/demo/demo1.md

package main

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/internal/log"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
)

// =============================================================================
// AppState - 定义应用状态
// =============================================================================

type AppState struct {
	Count     int    // 按钮点击计数
	ShowModal bool   // 是否显示模态框
	Input     string // 输入框文本
}

// =============================================================================
// Intent Types
// =============================================================================

type OpenModalIntent struct{}

func (OpenModalIntent) IntentType() string { return "OpenModal" }
func (OpenModalIntent) StayPressed() bool  { return true }

type AddCountIntent struct{}

func (AddCountIntent) IntentType() string { return "AddCount" }
func (AddCountIntent) StayPressed() bool  { return true }

type QuitIntent struct{}

func (QuitIntent) IntentType() string { return "Quit" }
func (QuitIntent) StayPressed() bool  { return false }

type CloseModalIntent struct{}

func (CloseModalIntent) IntentType() string { return "CloseModal" }
func (CloseModalIntent) StayPressed() bool  { return false }

// 添加一个内部 Intent 用于处理输入框变化
type SetInputIntent struct {
	Value string
}

func (SetInputIntent) IntentType() string { return "SetInput" }
func (SetInputIntent) StayPressed() bool  { return false }

// =============================================================================
// Store 初始化
// =============================================================================

// 环境变量控制：默认打开 modal 用于调试
var autoOpenModal = os.Getenv("AUTO_OPEN_MODAL") == "true"

var appStore = store.NewStore(AppState{
	Count:     0,
	ShowModal: autoOpenModal,
	Input:     "",
})

// =============================================================================
// Reducer 注册
// =============================================================================

func init() {
	reducer.NewBuilder[AppState]().
		On(OpenModalIntent{}, func(s AppState, i intent.Intent) AppState {
			s.ShowModal = true
			return s
		}).
		On(AddCountIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Count++
			return s
		}).
		On(CloseModalIntent{}, func(s AppState, i intent.Intent) AppState {
			s.ShowModal = false
			return s
		}).
		On(QuitIntent{}, func(s AppState, i intent.Intent) AppState {
			ui.Quit()
			return s
		}).
		On(SetInputIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Input = i.(SetInputIntent).Value
			return s
		}).
		BuildAndRegister(intent.DefaultRegistry(), appStore)
}

// =============================================================================
// Main
// =============================================================================

func main() {
	// Ensure default theme is loaded
	_ = theme.SetTheme("nord")

	// ============================================================
	// 调试环境变量 - 用于 HitTest 和 Modal 点击测试
	// ============================================================
	// 启用文件日志（记录 HitTest、鼠标位置、Modal 居中等）
	os.Setenv("TUI_DEBUG_LOG", "demo1_debug.log")

	// 自动打开 Modal，方便直接测试按钮点击
	os.Setenv("AUTO_OPEN_MODAL", "true")

	// 启用以下环境变量可获取更详细的调试信息：
	os.Setenv("TUI_DEBUG_HITMAP", "true") // HitMap 构建详情
	// os.Setenv("TUI_DEBUG_LAYER", "true")    // Layer 系统调试
	// os.Setenv("TUI_DEBUG_RENDER", "true") // 渲染管线调试
	os.Setenv("TUI_DEBUG_UI", "true") // UI 通用调试

	// ============================================================
	// 运行应用
	// ============================================================

	err := ui.Run(App,
		ui.WithWidth(80),
		ui.WithHeight(24),
		ui.WithTitle("Mint TUI - Full Featured Demo (Store 模式)"),
	)
	if err != nil {
		panic(err)
	}

	// 运行结束后提示日志位置
	log.UILogger.IfEnabled().Debug("=== Debug session ended ===")
	log.UILogger.IfEnabled().Debug("Log file: demo1_debug.log")
	log.UILogger.IfEnabled().Debug("Check for:")
	log.UILogger.IfEnabled().Debug("  - [MOUSE] mouse position and HitTest results")
	log.UILogger.IfEnabled().Debug("  - [LAYER] modal centering and position")
	log.UILogger.IfEnabled().Debug("  - [HITMAP] button bounds entries")
}

// =============================================================================
// App - 根组件
// =============================================================================

func App() ui.VNode {
	// ✅ 订阅存储的状态
	count := ui.UseStoreSelector(appStore, func(s AppState) int { return s.Count })
	showModal := ui.UseStoreSelector(appStore, func(s AppState) bool { return s.ShowModal })
	input := ui.UseStoreSelector(appStore, func(s AppState) string { return s.Input })

	// 生成大型列表用于 VirtualList
	items := make([]string, 100)
	for i := range items {
		items[i] = fmt.Sprintf("Log line #%d", i)
	}

	// 渲染主内容和模态框（当打开时）
	// Layer 系统会处理正确的 z-ordering 和居中
	mainContent := ui.VStackBuilder(
		Header(count),
		MainBody(count, input, items),
		DebugPanel(),
	).Stretch().Build()

	// 如果模态框打开，渲染主内容和模态框
	// LayerManager 会将它们分离到不同的层
	if showModal {
		modalVNode := ConfirmModal()

		result := ui.VStack(
			mainContent,
			// Modal 层 - 自动居中并覆盖主内容
			modalVNode,
		)

		return result
	}

	// 否则只渲染主内容
	return mainContent
}

// Header demonstrates state + layout with Bordered component
// Uses theme colors: PRIMARY for header background, TEXT for text
func Header(count int) ui.VNode {
	headerContent := ui.HStack(
		ui.NewTextBuilder("TUI Engine Demo").
			Style(style.FgBgBold(theme.Text(), theme.Primary())).
			Build(),
		ui.NewTextBuilder("              ").
			Style(style.FgBg(theme.Surface(), theme.Primary())).
			Build(),
		ui.NewButtonBuilder("[Open Modal]").
			Variant(ui.ButtonVariantPrimary). // 使用 Primary variant，默认就有 PRIMARY 背景
			OnPress(OpenModalIntent{}).
			FocusStyle(ui.FocusStyleBracket). // 恢复 Bracket 样式
			Build(),
		ui.NewTextBuilder(" ").
			Style(style.FgBg(theme.Surface(), theme.Primary())).
			Build(),
		ui.NewTextBuilder(fmt.Sprintf("Clicks: %d", count)).
			Style(style.FgBgBold(theme.BG(), theme.Primary())).
			Build(),
	)

	return ui.NewVStack().
		SingleBorder().
		BorderColor(theme.Primary()).
		SetChildrenList([]ui.VNode{headerContent})
}

// MainBody uses VStack/HStack with Bordered components for layout
// Matches the design from framework/docs/ui/demo/demo1.md:
//
//	┌───────────┬──────────────────────────────────────────┐
//	│ Menu      │ [ Input box............................... ] │
//	├───────────┼──────────────────────────────────────────┤
//	│ Add Count │ Log line #0                               │
//	├───────────┼──────────────────────────────────────────┤
//	│ Quit      │ Log line #1                               │
//	├───────────┼──────────────────────────────────────────┤
//	│           │ Log line #2                               │
//	├───────────┼──────────────────────────────────────────┤
//	│           │ Log line #3                               │
//	├───────────┼──────────────────────────────────────────┤
//	│           │ Log line #4                               │
//	├───────────┼──────────────────────────────────────────┤
//	│           │ Log line #5 ...                            │
//	└───────────┴──────────────────────────────────────────┘
func MainBody(count int, input string, items []string) ui.VNode {
	// Left sidebar with menu buttons
	// Uses theme colors: MUTED for menu label, Primary variant for Add Count, Danger variant for Quit
	sidebar := ui.VStackBuilder(
		ui.NewTextBuilder("Menu").
			Style(style.FgBoldUnderline(theme.Muted())).
			Build(),
		ui.NewButtonBuilder("Add Count").
			Variant(ui.ButtonVariantPrimary).
			OnPress(AddCountIntent{}).
			FocusStyle(ui.FocusStyleBracket).
			Build(),
		ui.NewButtonBuilder("Quit").
			Variant(ui.ButtonVariantDanger).
			FocusStyle(ui.FocusStyleBracket).
			OnPress(QuitIntent{}).
			Build(),
	).Stretch().Build()

	// 右侧内容区，带输入框和日志行
	// Uses theme colors: TEXT for labels, MUTED for log lines, BORDER for divider
	// 注意：Input 组件仍然使用 ForField(intent.ForField) 模式，这个组件需要进一步集成 Intent Bubble
	inputBuilder := ui.NewInputBuilder().
		Value(input).
		Placeholder("Type something...").
		Width(30) // Input width (less than panel width)

	// Store 模式下，Input 组件暂时显示值但不直接更新 Store
	// 完整的 Input Intent Bubble 集成是后续任务
	contentArea := ui.VStackBuilder(
		inputBuilder.Build(),
		ui.NewTextBuilder("──────────────────────────────────────").
			Style(style.Foreground(theme.Border())).
			Build(),
		ui.NewTextBuilder(items[0]).
			Style(style.Foreground(theme.Muted())).
			Build(),
		ui.NewTextBuilder(items[1]).
			Style(style.Foreground(theme.Muted())).
			Build(),
		ui.NewTextBuilder(items[2]).
			Style(style.Foreground(theme.Muted())).
			Build(),
		ui.NewTextBuilder(items[3]).
			Style(style.Foreground(theme.Muted())).
			Build(),
		ui.NewTextBuilder(items[4]).
			Style(style.Foreground(theme.Muted())).
			Build(),
		ui.HStack(
			ui.NewTextBuilder(items[5]).
				Style(style.Foreground(theme.Muted())).
				Build(),
			ui.NewTextBuilder(" ...").
				Style(style.FgItalic(theme.Placeholder())).
				Build(),
		),
	).Stretch().Build()

	// 用边框组合侧边栏和内容
	// Uses theme BORDER color for borders
	return ui.HStackBuilder(
		ui.Flex(
			ui.NewVStack().
				SingleBorder().
				BorderColor(theme.Border()).
				SetChildrenList([]ui.VNode{sidebar}),
			1, // Flex factor
		),
		ui.Flex(
			ui.NewVStack().
				SingleBorder().
				BorderColor(theme.Border()).
				SetChildrenList([]ui.VNode{contentArea}),
			1, // Flex factor
		),
	).Gap(0).Build()
}

// ConfirmModal demonstrates Layer + Focus Trap with overlay rendering
// Uses the new Layer system for automatic centering and backdrop
// Uses theme colors: WARNING for modal border, SUCCESS for OK button
func ConfirmModal() ui.VNode {
	// Modal content - the actual dialog box with border
	// Uses theme WARNING color for modal border to indicate caution
	modalBox := ui.NewVStack().
		SingleBorder().
		BorderColor(theme.Warning()).
		SetWidth(40). // Fixed width for the modal
		SetChildrenList([]ui.VNode{
			ui.VStackBuilder(
				ui.Text(""),
				// DEBUG: Line number to verify position
				ui.Text("=== MODAL START ==="),
				// Centered title - use HStack with AlignCenter
				// Uses theme WARNING color for title
				ui.HStackBuilder(
					ui.NewTextBuilder("*** Are you sure? ***").
						Style(style.FgBold(theme.Warning())).
						Build(),
				).Align(ui.AlignCenter).Build(),
				ui.Text(""),
				// Centered buttons - use HStack with AlignCenter
				// Uses theme colors: Secondary for Cancel, Success for OK
				ui.HStackBuilder(
					ui.NewButtonBuilder("[ Cancel ]").
						Variant(ui.ButtonVariantSecondary).
						OnPress(CloseModalIntent{}).
						FocusStyle(ui.FocusStyleBracket).
						Build(),
					ui.Text(" "),
					ui.NewButtonBuilder("[ OK ]").
						Variant(ui.ButtonVariantSuccess).
						FocusStyle(ui.FocusStyleBracket).
						OnPress(CloseModalIntent{}).
						Build(),
				).Align(ui.AlignCenter).Build(),
				ui.Text(""),
				// Centered footer text
				// Uses theme PLACEHOLDER color for hint text
				ui.HStackBuilder(
					ui.NewTextBuilder("Press ESC to close").
						Style(style.Foreground(theme.Placeholder())).
						Build(),
				).Align(ui.AlignCenter).Build(),
				ui.Text(""),
				// DEBUG: End marker
				ui.Text("=== MODAL END ==="),
			).Build(),
		})

	return ui.Modal(modalBox).
		CloseOnESC(true).
		CloseOnBackdropClick(true).
		Build()
}

// =============================================================================
// Debug Helper Functions
// =============================================================================

// DebugPanel 显示屏幕配置和调试信息
func DebugPanel() ui.VNode {
	infoLines := []string{
		"┌─ SCREEN/INFO PANEL ─────────────────────────────────────────────┐",
		"│ Buffer Size: 80x24 (configured via ui.WithWidth/Height)        │",
		"│ Mode: Store + Reducer (已迁移)                                  │",
		"│ Debug Log: demo1_debug.log (check for HitTest details)         │",
		"│                                                                │",
		"│ MODAL BUTTON HITEST VERIFICATION:                              │",
		"│ 1. Modal opens automatically (AUTO_OPEN_MODAL=true)            │",
		"│ 2. Click modal buttons - they increment the counter           │",
		"│ 3. Check demo1_debug.log for:                                  │",
		"│    - Mouse position (X, Y)                                     │",
		"│    - HitTest results (button bounds)                           │",
		"│    - Multiple button overlap detection                         │",
		"│    - Modal centering calculations                              │",
		"│                                                                │",
		"│ EXPECTED BEHAVIOR:                                             │",
		"│ - Modal centered in buffer: Y position depends on buffer size│",
		"│ - If actual terminal > 24 lines, check logs for actual size   │",
		"│ - Input component uses ForField, intent integration pending   │",
		"└────────────────────────────────────────────────────────────────┘",
	}

	return ui.VStackBuilder(
		ui.Text(infoLines[0]),
		ui.Text(infoLines[1]),
		ui.Text(infoLines[2]),
		ui.Text(infoLines[3]),
		ui.Text(infoLines[4]),
		ui.Text(infoLines[5]),
		ui.Text(infoLines[6]),
		ui.Text(infoLines[7]),
		ui.Text(infoLines[8]),
		ui.Text(infoLines[9]),
		ui.Text(infoLines[10]),
		ui.Text(infoLines[11]),
		ui.Text(infoLines[12]),
		ui.Text(infoLines[13]),
		ui.Text(infoLines[14]),
		ui.Text(infoLines[15]),
	).
		Build()
}

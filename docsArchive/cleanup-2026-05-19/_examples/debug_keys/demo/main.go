package main

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/ui"
)

// =============================================================================
// AppState - 定义应用状态
// =============================================================================

type AppState struct {
	ModalOpen        bool // Modal 是否打开
	OverlayVisible   bool // Overlay 是否可见
	InspectorEnabled bool // Inspector 是否启用
	ShowKeys         bool // 是否显示 Keys
	ShowPaths        bool // 是否显示 Paths
	ShowLayers       bool // 是否显示 Layers
}

// =============================================================================
// Intent Types
// =============================================================================

type OpenModalIntent struct{}
func (OpenModalIntent) IntentType() string      { return "OpenModal" }
func (OpenModalIntent) StayPressed() bool       { return true }

type ShowOverlayIntent struct{}
func (ShowOverlayIntent) IntentType() string    { return "ShowOverlay" }
func (ShowOverlayIntent) StayPressed() bool     { return true }

type ToggleInspectorIntent struct{}
func (ToggleInspectorIntent) IntentType() string { return "ToggleInspector" }
func (ToggleInspectorIntent) StayPressed() bool  { return true }

type CloseAllLayersIntent struct{}
func (CloseAllLayersIntent) IntentType() string  { return "CloseAllLayers" }
func (CloseAllLayersIntent) StayPressed() bool   { return true }

type ToggleShowKeysIntent struct{}
func (ToggleShowKeysIntent) IntentType() string  { return "ToggleShowKeys" }
func (ToggleShowKeysIntent) StayPressed() bool   { return true }

type ToggleShowPathsIntent struct{}
func (ToggleShowPathsIntent) IntentType() string { return "ToggleShowPaths" }
func (ToggleShowPathsIntent) StayPressed() bool  { return true }

type ToggleShowLayersIntent struct{}
func (ToggleShowLayersIntent) IntentType() string { return "ToggleShowLayers" }
func (ToggleShowLayersIntent) StayPressed() bool  { return true }

type ModalButtonClickIntent struct{}
func (ModalButtonClickIntent) IntentType() string { return "ModalButtonClick" }
func (ModalButtonClickIntent) StayPressed() bool  { return true }

type CloseModalIntent struct{}
func (CloseModalIntent) IntentType() string       { return "CloseModal" }
func (CloseModalIntent) StayPressed() bool        { return true }

type OverlayButtonClickIntent struct{}
func (OverlayButtonClickIntent) IntentType() string { return "OverlayButtonClick" }
func (OverlayButtonClickIntent) StayPressed() bool  { return true }

// =============================================================================
// Store 初始化
// ============================================================================

var debugKeysStore = store.NewStore(AppState{
	ModalOpen:        false,
	OverlayVisible:   false,
	InspectorEnabled: true,
	ShowKeys:         true,
	ShowPaths:        true,
	ShowLayers:       true,
})

// =============================================================================
// Reducer 注册
// ============================================================================

func init() {
	reducer.NewBuilder[AppState]().
		On(OpenModalIntent{}, func(s AppState, i intent.Intent) AppState {
			s.ModalOpen = true
			return s
		}).
		On(ShowOverlayIntent{}, func(s AppState, i intent.Intent) AppState {
			s.OverlayVisible = true
			return s
		}).
		On(ToggleInspectorIntent{}, func(s AppState, i intent.Intent) AppState {
			s.InspectorEnabled = !s.InspectorEnabled
			return s
		}).
		On(CloseAllLayersIntent{}, func(s AppState, i intent.Intent) AppState {
			s.ModalOpen = false
			s.OverlayVisible = false
			return s
		}).
		On(ToggleShowKeysIntent{}, func(s AppState, i intent.Intent) AppState {
			s.ShowKeys = !s.ShowKeys
			return s
		}).
		On(ToggleShowPathsIntent{}, func(s AppState, i intent.Intent) AppState {
			s.ShowPaths = !s.ShowPaths
			return s
		}).
		On(ToggleShowLayersIntent{}, func(s AppState, i intent.Intent) AppState {
			s.ShowLayers = !s.ShowLayers
			return s
		}).
		On(CloseModalIntent{}, func(s AppState, i intent.Intent) AppState {
			s.ModalOpen = false
			return s
		}).
		On(ModalButtonClickIntent{}, func(s AppState, i intent.Intent) AppState {
			fmt.Println("Modal button clicked!")
			return s
		}).
		On(OverlayButtonClickIntent{}, func(s AppState, i intent.Intent) AppState {
			fmt.Println("Overlay button clicked!")
			return s
		}).
		BuildAndRegister(intent.DefaultRegistry(), debugKeysStore)
}

// =============================================================================
// Main
// =============================================================================

func main() {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("🔑 UI Key Inspector - 交互式调试工具")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("\n这个工具会在 UI 中以 inspector 的形式显示所有 nodes 的 KEY 信息")
	fmt.Println("按 'q' 退出应用")
	fmt.Println("")

	ui.Run(func() ui.VNode {
		// ✅ 订阅存储的状态
		modalOpen := ui.UseStoreSelector(debugKeysStore, func(s AppState) bool { return s.ModalOpen })
		overlayVisible := ui.UseStoreSelector(debugKeysStore, func(s AppState) bool { return s.OverlayVisible })
		inspectorEnabled := ui.UseStoreSelector(debugKeysStore, func(s AppState) bool { return s.InspectorEnabled })
		showKeys := ui.UseStoreSelector(debugKeysStore, func(s AppState) bool { return s.ShowKeys })
		showPaths := ui.UseStoreSelector(debugKeysStore, func(s AppState) bool { return s.ShowPaths })
		showLayers := ui.UseStoreSelector(debugKeysStore, func(s AppState) bool { return s.ShowLayers })

		// Base layer content
		baseContent := ui.VStack(
			ui.NewTextBuilder("🔑 UI Key Inspector 演示").Bold(true).FgColor("cyan").Build(),
			ui.Text(""),
			ui.NewTextBuilder("点击按钮打开不同的 layer，观察 Inspector 中的 KEY 变化").FgColor("gray").Build(),
			ui.Text(""),
			ui.NewButtonBuilder("打开 Modal").
				OnPress(OpenModalIntent{}).
				Build(),
			ui.NewButtonBuilder("显示 Overlay").
				OnPress(ShowOverlayIntent{}).
				Build(),
			ui.NewButtonBuilder("切换 Inspector").
				OnPress(ToggleInspectorIntent{}).
				Build(),
			ui.NewButtonBuilder("关闭所有 Layers").
				OnPress(CloseAllLayersIntent{}).
				Build(),
			ui.Text(""),
			ui.NewTextBuilder("─────────────────────────────────").FgColor("gray").Build(),
			ui.NewTextBuilder("显示选项:").FgColor("cyan").Build(),
			ui.Text(""),
			buildCheckbox("显示 Keys", showKeys, ToggleShowKeysIntent{}),
			buildCheckbox("显示 Paths", showPaths, ToggleShowPathsIntent{}),
			buildCheckbox("显示 Layer 标记", showLayers, ToggleShowLayersIntent{}),
		)

		var modalContent ui.VNode
		if modalOpen {
			modalContent = ui.VStack(
				ui.NewTextBuilder("这是 Modal (LayerModal)").FgColor("cyan").Build(),
				ui.NewTextBuilder("观察 Inspector 中这个节点的 KEY").FgColor("gray").Build(),
				ui.Text(""),
				ui.NewButtonBuilder("Modal 内部的按钮").
					OnPress(ModalButtonClickIntent{}).
					Build(),
				ui.NewButtonBuilder("关闭 Modal").
					OnPress(CloseModalIntent{}).
					Build(),
			)
		}

		var overlayContent ui.VNode
		if overlayVisible {
			overlayContent = ui.VStack(
				ui.NewTextBuilder("这是 Overlay (LayerOverlay)").FgColor("yellow").Build(),
				ui.NewTextBuilder("观察 Inspector 中这个节点的 KEY").FgColor("gray").Build(),
				ui.Text(""),
				ui.NewButtonBuilder("Overlay 按钮").
					OnPress(OverlayButtonClickIntent{}).
					Build(),
			)
		}

		// 创建 Inspector overlay
		var inspectorContent ui.VNode
		if inspectorEnabled {
			inspectorContent = createInspectorOverlay(showKeys, showPaths, showLayers)
		}

		// 组合所有 children
		children := []ui.VNode{baseContent}
		if modalContent != nil {
			children = append(children, modalContent)
		}
		if overlayContent != nil {
			children = append(children, overlayContent)
		}
		if inspectorContent != nil {
			children = append(children, inspectorContent)
		}

		return ui.VStack(children...)
	},
		ui.WithWidth(80),
		ui.WithHeight(40),
		ui.WithTitle("UI Key Inspector (Store 模式)"),
	)
}

// =============================================================================
// Helper Functions
// =============================================================================

// buildCheckbox 创建一个简单的 checkbox (使用 text + button 模拟)
func buildCheckbox(label string, checked bool, toggleIntent intent.Intent) ui.VNode {
	var status string
	if checked {
		status = "[X] "
	} else {
		status = "[ ] "
	}

	return ui.HStack(
		ui.NewTextBuilder(status+label).Build(),
		ui.NewButtonBuilder("切换").
			OnPress(toggleIntent).
			Build(),
	)
}

// createInspectorOverlay 创建 Inspector overlay，显示 KEY 信息
func createInspectorOverlay(showKeys, showPaths, showLayers bool) ui.VNode {
	// TODO: 这里应该获取当前的 VNode 和 Fiber 树
	// 由于当前架构限制，我们创建一个示例演示

	lines := []string{
		"══════════════════════════════════════════",
		"🔑 KEY INSPECTOR",
		"══════════════════════════════════════════",
		"",
		"示例 KEY 信息:",
		"│─ vstack",
		"  │─ text",
		"  │─ button",
		"  │─ vstack [MODAL]",
		"    │─ text",
		"    │─ button",
		"",
		fmt.Sprintf("显示 Keys: %v", showKeys),
		fmt.Sprintf("显示 Paths: %v", showPaths),
		fmt.Sprintf("显示 Layers: %v", showLayers),
		"",
		"注意: 这是一个演示 overlay",
		"要获取真实数据，需要访问",
		"reconciler 的内部状态",
		"══════════════════════════════════════════",
	}

	textNodes := make([]ui.VNode, len(lines))
	for i, line := range lines {
		textNodes[i] = ui.Text(line)
	}

	return ui.VStack(textNodes...)
}

// Demo 2: Runtime Internals Visualization (Store 模式)
//
// This demo visualizes the internal runtime scheduling flow from setState
// to terminal buffer output - the "Total Assembly Diagram" of the engine.
//
// Pipeline:
//   Event → setState → Scheduler → Render → Reconcile → Layout
//   → Layer Merge → Paint → Buffer Diff → Terminal Output
//
// Based on: framework/docs/ui/demo/demo2_inside.md

package main

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/framework/theme"
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
	CurrentPhase   string // 当前阶段: idle, Event, setState, Scheduler, Render, Reconcile, Layout, Paint
	EventCount     int    // 事件计数
	RenderCount    int    // 渲染计数
	BufferUpdates  int    // 缓冲区更新计数
}

// =============================================================================
// Intent Types
// =============================================================================

type SetEventPhaseIntent struct{}
func (SetEventPhaseIntent) IntentType() string { return "SetEventPhase" }
func (SetEventPhaseIntent) StayPressed() bool  { return true }

type SetSetStatePhaseIntent struct{}
func (SetSetStatePhaseIntent) IntentType() string { return "SetSetStatePhase" }
func (SetSetStatePhaseIntent) StayPressed() bool  { return true }

type SetSchedulerPhaseIntent struct{}
func (SetSchedulerPhaseIntent) IntentType() string { return "SetSchedulerPhase" }
func (SetSchedulerPhaseIntent) StayPressed() bool  { return true }

type SetRenderPhaseIntent struct{}
func (SetRenderPhaseIntent) IntentType() string { return "SetRenderPhase" }
func (SetRenderPhaseIntent) StayPressed() bool  { return true }

type SetReconcilePhaseIntent struct{}
func (SetReconcilePhaseIntent) IntentType() string { return "SetReconcilePhase" }
func (SetReconcilePhaseIntent) StayPressed() bool  { return true }

type SetLayoutPhaseIntent struct{}
func (SetLayoutPhaseIntent) IntentType() string { return "SetLayoutPhase" }
func (SetLayoutPhaseIntent) StayPressed() bool  { return true }

type SetPaintPhaseIntent struct{}
func (SetPaintPhaseIntent) IntentType() string { return "SetPaintPhase" }
func (SetPaintPhaseIntent) StayPressed() bool  { return true }

type SetIdlePhaseIntent struct{}
func (SetIdlePhaseIntent) IntentType() string { return "SetIdlePhase" }
func (SetIdlePhaseIntent) StayPressed() bool  { return true }

// =============================================================================
// Store 初始化
// =============================================================================

var runtimeStore = store.NewStore(AppState{
	CurrentPhase:  "idle",
	EventCount:    0,
	RenderCount:   0,
	BufferUpdates: 0,
})

// =============================================================================
// Reducer 注册
// =============================================================================

func init() {
	reducer.NewBuilder[AppState]().
		On(SetEventPhaseIntent{}, func(s AppState, i intent.Intent) AppState {
			s.CurrentPhase = "Event"
			s.EventCount++
			return s
		}).
		On(SetSetStatePhaseIntent{}, func(s AppState, i intent.Intent) AppState {
			s.CurrentPhase = "setState"
			return s
		}).
		On(SetSchedulerPhaseIntent{}, func(s AppState, i intent.Intent) AppState {
			s.CurrentPhase = "Scheduler"
			s.RenderCount++
			return s
		}).
		On(SetRenderPhaseIntent{}, func(s AppState, i intent.Intent) AppState {
			s.CurrentPhase = "Render"
			return s
		}).
		On(SetReconcilePhaseIntent{}, func(s AppState, i intent.Intent) AppState {
			s.CurrentPhase = "Reconcile"
			return s
		}).
		On(SetLayoutPhaseIntent{}, func(s AppState, i intent.Intent) AppState {
			s.CurrentPhase = "Layout"
			return s
		}).
		On(SetPaintPhaseIntent{}, func(s AppState, i intent.Intent) AppState {
			s.CurrentPhase = "Paint"
			s.BufferUpdates++
			return s
		}).
		On(SetIdlePhaseIntent{}, func(s AppState, i intent.Intent) AppState {
			s.CurrentPhase = "idle"
			return s
		}).
		BuildAndRegister(intent.DefaultRegistry(), runtimeStore)
}

// =============================================================================
// Main
// =============================================================================

func main() {
	// Check if layout debug mode is enabled
	if os.Getenv("TUI_UI_DEBUG_LAYOUT") == "true" || os.Getenv("TUI_LAYOUT_DEBUG") == "true" {
		fmt.Println("=== Layout Info Test ===")
		return
	}

	// Initialize theme
	_ = theme.SetTheme("nord")

	err := ui.Run(RuntimeDemo,
		ui.WithWidth(100),
		ui.WithHeight(35),
		ui.WithTitle("Mint TUI - Runtime Internals (Store 模式)"),
	)
	if err != nil {
		panic(err)
	}
}

// =============================================================================
// RuntimeDemo - 可视化运行时管道
// =============================================================================

func RuntimeDemo() ui.VNode {
	// ✅ 订阅存储的状态
	currentPhase := ui.UseStoreSelector(runtimeStore, func(s AppState) string { return s.CurrentPhase })
	eventCount := ui.UseStoreSelector(runtimeStore, func(s AppState) int { return s.EventCount })
	renderCount := ui.UseStoreSelector(runtimeStore, func(s AppState) int { return s.RenderCount })
	bufferUpdates := ui.UseStoreSelector(runtimeStore, func(s AppState) int { return s.BufferUpdates })

	return ui.VStack(
		HeaderPanel(),
		PipelineVisualization(currentPhase),
		StatisticsPanel(eventCount, renderCount, bufferUpdates),
		ControlPanel(),
		ExplanationPanel(currentPhase),
	)
}

// HeaderPanel shows the title with border
func HeaderPanel() ui.VNode {
	// Use HStackBuilder with AlignCenter for true center alignment
	headerContent := ui.HStackBuilder(
		ui.NewTextBuilder("Runtime Scheduling Pipeline Visualization").
			Style(style.FgBold(theme.Text())).
			Build(),
	).
		Gap(0).
		Align(ui.AlignCenter).
		Build()

	// Use FillWidth() to stretch horizontally WITHOUT affecting vertical direction
	// This is the new layout system feature for single-component stretching
	return ui.NewVStack().
		SingleBorder().
		BorderColor(theme.Primary()).
		SetChildrenList([]ui.VNode{headerContent})
	// Note: FillWidth() is handled by parent layout
}

// PipelineVisualization shows the runtime pipeline flow
func PipelineVisualization(currentPhase string) ui.VNode {
	phases := []struct {
		name     string
		color    string
		position int
	}{
		{"Event", "red", 0},
		{"setState", "yellow", 15},
		{"Scheduler", "green", 30},
		{"Render", "blue", 45},
		{"Reconcile", "magenta", 60},
		{"Layout", "cyan", 75},
		{"Paint", "white", 90},
	}

	activeIndex := -1
	for i, p := range phases {
		if p.name == currentPhase {
			activeIndex = i
			break
		}
	}

	return ui.NewVStack().
		SingleBorder().
		BorderColor(theme.Border()).
		SetChildrenList([]ui.VNode{
			ui.VStack(
				buildPipelineLine(phases, activeIndex),
				ui.Text(""),
				buildPipelineArrows(phases, activeIndex),
			),
		})
}

// buildPipelineLine creates the phase boxes
func buildPipelineLine(phases []struct{ name string; color string; position int }, activeIndex int) ui.VNode {
	var result string
	for _, p := range phases {
		spaces := p.position - len(result)
		for i := 0; i < spaces; i++ {
			result += " "
		}
		result += "[" + p.name + "]"
		_ = p // unused but kept for clarity
	}
	return ui.NewTextBuilder(result).
		Style(style.Foreground(theme.Muted())).
		Build()
}

// buildPipelineArrows creates the flow arrows
func buildPipelineArrows(phases []struct{ name string; color string; position int }, activeIndex int) ui.VNode {
	var result string
	for i := range phases {
		if i > 0 {
			result += "       "
		}
		if i < len(phases)-1 {
			isActive := i == activeIndex || (i+1 == activeIndex && activeIndex != -1)
			arrow := "→"
			if isActive {
				arrow = "➤" // 活跃箭头
			}
			result += arrow
		}
	}
	return ui.NewTextBuilder(result).
		Style(style.Foreground(theme.Border())).
		Build()
}

// StatisticsPanel shows runtime statistics
func StatisticsPanel(eventCount int, renderCount int, bufferUpdates int) ui.VNode {
	stats := ui.VStack(
		ui.NewTextBuilder(fmt.Sprintf("Events: %d", eventCount)).
			Style(style.FgBold(theme.Primary())).
			Build(),
		ui.NewTextBuilder(fmt.Sprintf("Renders: %d", renderCount)).
			Style(style.FgBold(theme.Primary())).
			Build(),
		ui.NewTextBuilder(fmt.Sprintf("Buffer Updates: %d", bufferUpdates)).
			Style(style.FgBold(theme.Primary())).
			Build(),
	)

	return ui.NewVStack().
		SingleBorder().
		BorderColor(theme.Primary()).
		SetChildrenList([]ui.VNode{stats})
}

// ControlPanel provides buttons to simulate pipeline phases
func ControlPanel() ui.VNode {
	buttons := ui.VStack(
		ui.HStack(
			ui.NewButtonBuilder("Event").
				OnPress(SetEventPhaseIntent{}).
				Build(),
			ui.Text(" "),
			ui.NewButtonBuilder("setState").
				OnPress(SetSetStatePhaseIntent{}).
				Build(),
			ui.Text(" "),
			ui.NewButtonBuilder("Scheduler").
				OnPress(SetSchedulerPhaseIntent{}).
				Build(),
		),
		ui.HStack(
			ui.NewButtonBuilder("Render").
				OnPress(SetRenderPhaseIntent{}).
				Build(),
			ui.Text(" "),
			ui.NewButtonBuilder("Reconcile").
				OnPress(SetReconcilePhaseIntent{}).
				Build(),
			ui.Text(" "),
			ui.NewButtonBuilder("Layout").
				OnPress(SetLayoutPhaseIntent{}).
				Build(),
		),
		ui.HStack(
			ui.NewButtonBuilder("Paint").
				OnPress(SetPaintPhaseIntent{}).
				Build(),
			ui.Text(" "),
			ui.NewButtonBuilder("Idle").
				OnPress(SetIdlePhaseIntent{}).
				Build(),
		),
	)

	return ui.NewVStack().
		SingleBorder().
		BorderColor(theme.Primary()).
		SetChildrenList([]ui.VNode{buttons})
}

// ExplanationPanel provides phase descriptions
func ExplanationPanel(currentPhase string) ui.VNode {
	descriptions := map[string]string{
		"idle":      "System is idle, waiting for events",
		"Event":     "User event detected (keypress, mouse, resize)",
		"setState":  "State update triggered, creating render request",
		"Scheduler": "Work scheduler processes render requests efficiently",
		"Render":    "Virtual DOM (VNode) tree is created from components",
		"Reconcile": "Diff of old and new VNode trees, generating patch instructions",
		"Layout":    "Calculate positions (x, y, width, height) for all components",
		"Paint":     "Apply styles and write to terminal buffer",
	}

	desc := descriptions[currentPhase]
	if desc == "" {
		desc = "No phase selected"
	}

	explanation := ui.VStack(
		ui.NewTextBuilder("Current Phase:").
			Style(style.FgBold(theme.Warning())).
			Build(),
		ui.NewTextBuilder(currentPhase).
			Style(style.Foreground(theme.Primary())).
			Build(),
		ui.Text(""),
		ui.NewTextBuilder(desc).
			Style(style.Foreground(theme.Muted())).
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("Use buttons above to simulate pipeline flow").
			Style(style.Foreground(theme.Border())).
			Build(),
	)

	return ui.NewVStack().
		SingleBorder().
		BorderColor(theme.Primary()).
		SetChildrenList([]ui.VNode{explanation})
}

package main

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/ui"
)

// =============================================================================
// AppState - Progress 状态
// =============================================================================

type AppState struct {
	Progress int // 进度值 0-100
}

// =============================================================================
// Intent Types
// =============================================================================

type IncrementProgressIntent struct{}
func (IncrementProgressIntent) IntentType() string { return "IncrementProgress" }
func (IncrementProgressIntent) StayPressed() bool  { return true }

type ResetProgressIntent struct{}
func (ResetProgressIntent) IntentType() string    { return "ResetProgress" }
func (ResetProgressIntent) StayPressed() bool     { return true }

// =============================================================================
// Store 初始化
// =============================================================================

var progressStore = store.NewStore(AppState{
	Progress: 0,
})

// =============================================================================
// Reducer 注册
// =============================================================================

func init() {
	reducer.NewBuilder[AppState]().
		On(IncrementProgressIntent{}, func(s AppState, i intent.Intent) AppState {
			if s.Progress < 100 {
				s.Progress += 10
			}
			return s
		}).
		On(ResetProgressIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Progress = 0
			return s
		}).
		BuildAndRegister(intent.DefaultRegistry(), progressStore)
}

// =============================================================================
// Progress Demo Component
// =============================================================================

func ProgressDemo() ui.VNode {
	// ✅ 订阅 progress 状态
	progress := ui.UseStoreSelector(progressStore, func(s AppState) int { return s.Progress })

	return ui.VStack(
		ui.NewTextBuilder("Progress Bar Demo").
			FgColor("cyan").
			Bold(true).
			Build(),
		ui.Text(""),
		// TODO: SpinnerBuilder 暂时不可用
		// ui.NewSpinnerBuilder().
		// 	Message("Loading demo...").
		// 	FgColor("yellow").
		// 	Build(),
		ui.NewTextBuilder("Spinner placeholder...").FgColor("yellow").Build(),
		ui.Text(""),
		ui.NewProgressBuilder().
			Label("Download:").
			Value(progress).
			Max(100).
			Width(30).
			ShowPercent(true).
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("Status:").
			FgColor("bright-black").
			Build(),
		StatusText(progress),
		ui.Text(""),
		ui.HStack(
			ui.NewButtonBuilder("  +10%  ").
				// ✅ 使用自定义 Intent - 由 Reducer 处理
				OnPress(IncrementProgressIntent{}).
				Build(),
			ui.Text(" "),
			ui.NewButtonBuilder("  Reset  ").
				OnPress(ResetProgressIntent{}).
				Build(),
		),
		ui.Text(""),
		ui.NewTextBuilder("Press button to increase progress").
			FgColor("bright-black").
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("q: quit").
			FgColor("bright-black").
			Build(),
	)
}

// =============================================================================
// Status Text Component
// =============================================================================

func StatusText(progress int) ui.VNode {
	switch {
	case progress < 30:
		return ui.NewTextBuilder("  Starting...").
			FgColor("bright-black").
			Build()
	case progress < 70:
		return ui.NewTextBuilder("  In progress...").
			FgColor("yellow").
			Build()
	case progress < 100:
		return ui.NewTextBuilder("  Almost done!").
			FgColor("cyan").
			Build()
	default:
		return ui.NewTextBuilder("  Complete!").
			FgColor("green").
			Bold(true).
			Build()
	}
}

// =============================================================================
// Main
// =============================================================================

func main() {
	err := ui.Run(ProgressDemo,
		ui.WithWidth(50),
		ui.WithHeight(20),
		ui.WithTitle("Progress Demo (Store 模式)"),
	)
	if err != nil {
		panic(err)
	}
}

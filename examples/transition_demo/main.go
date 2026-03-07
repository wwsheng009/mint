// Package main demonstrates Transition Intent pattern for async operations (Store 模式).
//
// This example shows:
//   1. Using Store + Reducer for async state management
//   2. Showing loading states while operation runs
//   3. Updating UI with results when operation completes
//
// Key Concept: Async operations follow this pattern:
//   1. User clicks → Intent emitted → Reducer updates store
//   2. Handler sets loading state → UI shows spinner
//   3. Background work runs in goroutine
//   4. Goroutine updates store with result → UI re-renders
package main

import (
	"fmt"
	"time"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/ui"
	buttoncomp "github.com/wwsheng009/mint/ui/components/button"
)

// =============================================================================
// AppState - 定义应用状态
// =============================================================================

type AppState struct {
	IsLoading  bool   // 是否正在加载
	LastResult string // 加载结果
	Status     string // 加载状态文本
}

// =============================================================================
// Async Operation Intents
// =============================================================================

// LoadDataIntent starts an async data loading operation.
type LoadDataIntent struct {
	Source string // Where to load data from (e.g., "API", "Database", "File")
}

func (LoadDataIntent) IntentType() string { return "LoadData" }
func (LoadDataIntent) StayPressed() bool  { return true }

type SetLoadingIntent struct {
	Loading bool
}

func (SetLoadingIntent) IntentType() string { return "SetLoading" }
func (SetLoadingIntent) StayPressed() bool  { return false } // 内部操作

type SetResultIntent struct {
	Result string
}

func (SetResultIntent) IntentType() string { return "SetResult" }
func (SetResultIntent) StayPressed() bool  { return false } // 内部操作

type SetStatusIntent struct {
	Status string
}

func (SetStatusIntent) IntentType() string { return "SetStatus" }
func (SetStatusIntent) StayPressed() bool  { return false } // 内部操作

// =============================================================================
// Store 初始化
// =============================================================================

var appStore = store.NewStore(AppState{
	IsLoading:  false,
	LastResult: "",
	Status:     "",
})

// =============================================================================
// Reducer 注册
// =============================================================================

func init() {
	reducer.NewBuilder[AppState]().
		On(LoadDataIntent{}, func(s AppState, i intent.Intent) AppState {
			// 设置加载状态
			s.IsLoading = true
			s.Status = fmt.Sprintf("Loading from %s...", i.(LoadDataIntent).Source)

			// 在后台执行异步操作
			go asyncLoadData(i.(LoadDataIntent).Source)

			return s
		}).
		On(SetLoadingIntent{}, func(s AppState, i intent.Intent) AppState {
			s.IsLoading = i.(SetLoadingIntent).Loading
			return s
		}).
		On(SetResultIntent{}, func(s AppState, i intent.Intent) AppState {
			s.LastResult = i.(SetResultIntent).Result
			return s
		}).
		On(SetStatusIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Status = i.(SetStatusIntent).Status
			return s
		}).
		BuildAndRegister(intent.DefaultRegistry(), appStore)
}

// =============================================================================
// 异步加载函数
// =============================================================================

func asyncLoadData(source string) {
	// 模拟异步操作（如 API 调用、数据库查询、文件 I/O）
	// 随机时长 1-3 秒
	duration := time.Duration(1000 + (time.Now().UnixNano() % 2000))
	time.Sleep(duration)

	// 更新 UI 结果
	result := fmt.Sprintf("Data from %s (loaded in %.1fs)", source, duration.Seconds())

	// 通过 Store 更新状态
	appStore.Update(func(s AppState) AppState {
		s.IsLoading = false
		s.LastResult = result
		s.Status = ""
		return s
	})
}

// =============================================================================
// Main Application Component
// =============================================================================

func App() ui.VNode {
	// ✅ 订阅状态
	isLoading := ui.UseStoreSelector(appStore, func(s AppState) bool { return s.IsLoading })
	lastResult := ui.UseStoreSelector(appStore, func(s AppState) string { return s.LastResult })
	status := ui.UseStoreSelector(appStore, func(s AppState) string { return s.Status })

	// 获取加载动画（每 500ms 交替）
	animationFrame := int(time.Now().UnixNano() / 500000000) % 4
	spinners := []string{"|", "/", "-", "\\"}
	spinChar := spinners[animationFrame]

	// 构建 UI
	return ui.VStack(
		// 标题
		ui.NewTextBuilder("╔══════════════════════════════════════╗").
			FgColor("cyan").
			Build(),
		ui.NewTextBuilder("║     Transition Intent Demo           ║").
			FgColor("cyan").
			Build(),
		ui.NewTextBuilder("╚══════════════════════════════════════╝").
			FgColor("cyan").
			Build(),
		ui.Text(""),

		// 描述
		ui.NewTextBuilder("Async Operation Pattern").
			FgColor("yellow").
			Bold(true).
			Build(),
		ui.Text(""),
		ui.Text("Flow: Click → Loading → Background work → Result"),

		ui.Text(""),
		ui.NewTextBuilder("────────────────────────────────────").
			FgColor("bright-black").
			Build(),
		ui.Text(""),

		// 显示加载状态
		func() ui.VNode {
			if isLoading {
				return ui.HStack(
					ui.NewTextBuilder("[").
						FgColor("yellow").
						Build(),
					ui.NewTextBuilder(spinChar).
						FgColor("yellow").
						Bold(true).
						Build(),
					ui.NewTextBuilder("Loading").
						FgColor("yellow").
						Build(),
					ui.Text("] "),
					ui.Text(status),
				)
			}
			return ui.HStack(
				ui.NewTextBuilder("[Ready] ").
					FgColor("green").
					Build(),
				ui.Text("Click a button to load data"),
			)
		}(),

		ui.Text(""),

		// 显示最后的结果
		func() ui.VNode {
			if lastResult != "" {
				return ui.HStack(
					ui.NewTextBuilder("✓ Result:").
						FgColor("bright-black").
						Bold(true).
						Build(),
					ui.Text(lastResult),
				)
			}
			return ui.Text("")
		}(),

		ui.Text(""),
		ui.NewTextBuilder("────────────────────────────────────").
			FgColor("bright-black").
			Build(),
		ui.Text(""),

		// 加载按钮
		ui.NewTextBuilder("Trigger Async Operations:").
			FgColor("gray").
			Build(),
		ui.Text(""),
		ui.HStack(
			buttoncomp.NewBuilder("[API]").
				OnPress(LoadDataIntent{Source: "API Server"}).
				Variant(buttoncomp.VariantPrimary).
				Disabled(isLoading).
				Build(),
			ui.Text(" "),
			buttoncomp.NewBuilder("[DB]").
				OnPress(LoadDataIntent{Source: "Database"}).
				Variant(buttoncomp.VariantSecondary).
				Disabled(isLoading).
				Build(),
			ui.Text(" "),
			buttoncomp.NewBuilder("[File]").
				OnPress(LoadDataIntent{Source: "File System"}).
				Variant(buttoncomp.VariantSecondary).
				Disabled(isLoading).
				Build(),
		),
		ui.Text(""),
		ui.Text(""),
		ui.NewTextBuilder("[Tip] Buttons disabled during load").
			FgColor("gray").
			Build(),
	)
}

// =============================================================================
// Main Function
// =============================================================================

func main() {
	err := ui.Run(App,
		ui.WithWidth(60),
		ui.WithHeight(22),
		ui.WithTitle("Transition Demo (Store 模式)"),
	)
	if err != nil {
		panic(err)
	}
}

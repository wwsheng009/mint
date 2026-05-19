package main

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/ui"
)

// =============================================================================
// AppState - 定义应用状态
// =============================================================================

type AppState struct {
	Count int // 按钮计数
	Text  string
	Checked1 bool
	Checked2 bool
	SelectedIndex int
}

// =============================================================================
// Intent Types
// =============================================================================

type DecrementMouseIntent struct{}
func (DecrementMouseIntent) IntentType() string { return "DecrementMouse" }
func (DecrementMouseIntent) StayPressed() bool  { return true }

type IncrementMouseIntent struct{}
func (IncrementMouseIntent) IntentType() string { return "IncrementMouse" }
func (IncrementMouseIntent) StayPressed() bool  { return true }

// =============================================================================
// Store 初始化
// =============================================================================

var mouseStore = store.NewStore(AppState{
	Count:       0,
	Text:        "",
	Checked1:    false,
	Checked2:    false,
	SelectedIndex: 0,
})

// =============================================================================
// Reducer 注册
// =============================================================================

func init() {
	reducer.NewBuilder[AppState]().
		On(DecrementMouseIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Count--
			return s
		}).
		On(IncrementMouseIntent{}, func(s AppState, i intent.Intent) AppState {
			s.Count++
			return s
		}).
		BuildAndRegister(intent.DefaultRegistry(), mouseStore)
}

// =============================================================================
// MouseInteractionDemo showcases all mouse-supported components
// =============================================================================

func MouseInteractionDemo() ui.VNode {
	// ✅ 订阅状态
	count := ui.UseStoreSelector(mouseStore, func(s AppState) int { return s.Count })
	text := ui.UseStoreSelector(mouseStore, func(s AppState) string { return s.Text })
	checked1 := ui.UseStoreSelector(mouseStore, func(s AppState) bool { return s.Checked1 })
	checked2 := ui.UseStoreSelector(mouseStore, func(s AppState) bool { return s.Checked2 })
	selectedIndex := ui.UseStoreSelector(mouseStore, func(s AppState) int { return s.SelectedIndex })

	return ui.VStack(
		// Header
		ui.NewTextBuilder("╔══════════════════════════════════════════╗").
			FgColor("cyan").
			Build(),
		ui.NewTextBuilder("║     Mouse Interaction Demo               ║").
			FgColor("cyan").
			Build(),
		ui.NewTextBuilder("║     🖱️ Hover & Click to interact          ║").
			FgColor("cyan").
			Build(),
		ui.NewTextBuilder("╚══════════════════════════════════════════╝").
			FgColor("cyan").
			Build(),
		ui.Text(""),
		ui.Text(""),

		// Button Section
		ui.NewTextBuilder("🔘 BUTTONS - Click to interact").
			FgColor("yellow").
			Bold(true).
			Build(),
		ui.Text(""),
		ui.HStack(
			ui.NewButtonBuilder(" [-] ").
				OnPress(DecrementMouseIntent{}).
				Build(),
			ui.Text(" "),
			ui.NewTextBuilder(fmt.Sprintf(" Count: %d ", count)).
				Bold(true).
				FgColor("green").
				Build(),
			ui.Text(" "),
			ui.NewButtonBuilder(" [+] ").
				OnPress(IncrementMouseIntent{}).
				Build(),
		),
		ui.Text(""),

		// Checkbox Section
		ui.NewTextBuilder("☑️ CHECKBOXES - Click to toggle").
			FgColor("yellow").
			Bold(true).
			Build(),
		ui.Text(""),
		ui.NewCheckboxBuilder().
			Label("Enable notifications").
			Checked(checked1).
			Build(), // TODO: integrate with FieldChangeIntent
		ui.NewCheckboxBuilder().
			Label("Accept terms and conditions").
			Checked(checked2).
			Build(), // TODO: integrate with FieldChangeIntent
		ui.Text(""),

		// Input Section
		ui.NewTextBuilder("📝 INPUT - Click to focus, type to edit").
			FgColor("yellow").
			Bold(true).
			Build(),
		ui.Text(""),
		ui.HStack(
			ui.Text("Name: "),
			ui.NewInputBuilder().
				Value(text).
				Placeholder("Hover and click here...").
				Build(), // TODO: integrate with FieldChangeIntent
		),
		ui.Text(""),

		// Select Section
		ui.NewTextBuilder("📋 SELECT - Click to cycle options").
			FgColor("yellow").
			Bold(true).
			Build(),
		ui.Text(""),
		ui.HStack(
			ui.Text("Theme: "),
			ui.NewSelectBuilder().
				Options([]ui.SelectOption{
					ui.NewSelectOption("dark", "Dark"),
					ui.NewSelectOption("light", "Light"),
					ui.NewSelectOption("blue", "Blue"),
					ui.NewSelectOption("green", "Green"),
				}).
				Selected(selectedIndex).
				Build(), // TODO: integrate with FieldChangeIntent
		),
		ui.Text(""),

		// Textarea Section
		ui.NewTextBuilder("📄 TEXTAREA - Click to focus").
			FgColor("yellow").
			Bold(true).
			Build(),
		ui.Text(""),
		ui.NewTextareaBuilder().
			Placeholder("Hover and click to edit multi-line text...").
			Rows(3).
			Cols(40).
			Build(),
		ui.Text(""),

		// Info Section
		ui.NewTextBuilder("─────────────────────────────────────────").
			FgColor("bright-black").
			Build(),
		ui.Text(""),
		ui.NewTextBuilder("💡 TIP: Hover over controls highlights them").
			FgColor("gray").
			Build(),
		ui.NewTextBuilder("💡 TIP: Click buttons/checkboxes to interact").
			FgColor("gray").
			Build(),
		ui.NewTextBuilder("💡 TIP: Use Tab to navigate, Enter to select").
			FgColor("gray").
			Build(),
	)
}

// =============================================================================
// Main
// =============================================================================

func main() {
	ui.Run(MouseInteractionDemo,
		ui.WithWidth(50),
		ui.WithHeight(28),
		ui.WithTitle("Mouse Interaction Demo (Store 模式)"),
	)
}

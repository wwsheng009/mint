// Demo 5: IDE Interface Demo (Store 模式)
//
// This is a comprehensive IDE-level demo that demonstrates:
// - Grid main layout
// - Sidebar Tree (virtual list)
// - Editor (multi-line input with cursor)
// - Tabs
// - Console (real-time log streaming)
// - Modal (find dialog, command palette)
//
// Based on: framework/docs/ui/demo/demo5_ide.md

package main

import (
	"fmt"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/ui"
)

// =============================================================================
// AppState - IDE 应用状态
// =============================================================================

type AppState struct {
	ActiveFile        string // 当前选中的文件
	ShowCommandPalette bool   // 是否显示命令面板
	SelectedTab       string // 当前选中的标签
	EditorContent     string // 编辑器内容
}

// =============================================================================
// Intent Types
// =============================================================================

type ShowCommandPaletteIntent struct{}
func (ShowCommandPaletteIntent) IntentType() string { return "ShowCommandPalette" }
func (ShowCommandPaletteIntent) StayPressed() bool  { return true }

type HideCommandPaletteIntent struct{}
func (HideCommandPaletteIntent) IntentType() string { return "HideCommandPalette" }
func (HideCommandPaletteIntent) StayPressed() bool  { return true }

type SetActiveFileIntent struct {
	Name string
}
func (SetActiveFileIntent) IntentType() string { return "SetActiveFile" }
func (SetActiveFileIntent) StayPressed() bool  { return true }

type SetSelectedTabIntent struct {
	TabID string
}
func (SetSelectedTabIntent) IntentType() string { return "SetSelectedTab" }
func (SetSelectedTabIntent) StayPressed() bool  { return true }

type SetEditorContentIntent struct {
	Content string
}
func (SetEditorContentIntent) IntentType() string { return "SetEditorContent" }
func (SetEditorContentIntent) StayPressed() bool  { return false }

// =============================================================================
// Store 初始化
// =============================================================================

var ideStore = store.NewStore(AppState{
	ActiveFile:        "main.go",
	ShowCommandPalette: false,
	SelectedTab:       "editor",
	EditorContent:     "func main() {\n    ui.Run(App)\n}",
})

// =============================================================================
// Reducer 注册
// =============================================================================

func init() {
	reducer.NewBuilder[AppState]().
		On(ShowCommandPaletteIntent{}, func(s AppState, i intent.Intent) AppState {
			s.ShowCommandPalette = true
			return s
		}).
		On(HideCommandPaletteIntent{}, func(s AppState, i intent.Intent) AppState {
			s.ShowCommandPalette = false
			return s
		}).
		On(SetActiveFileIntent{}, func(s AppState, i intent.Intent) AppState {
			s.ActiveFile = i.(SetActiveFileIntent).Name
			return s
		}).
		On(SetSelectedTabIntent{}, func(s AppState, i intent.Intent) AppState {
			s.SelectedTab = i.(SetSelectedTabIntent).TabID
			return s
		}).
		On(SetEditorContentIntent{}, func(s AppState, i intent.Intent) AppState {
			s.EditorContent = i.(SetEditorContentIntent).Content
			return s
		}).
		BuildAndRegister(intent.DefaultRegistry(), ideStore)
}

// =============================================================================
// Main
// =============================================================================

func main() {
	err := ui.Run(IDEDemo,
		ui.WithWidth(100),
		ui.WithHeight(40),
		ui.WithTitle("Mint TUI - IDE Interface (Store 模式)"),
	)
	if err != nil {
		panic(err)
	}
}

// =============================================================================
// IDEDemo - 根组件
// =============================================================================

func IDEDemo() ui.VNode {
	// ✅ 订阅存储的状态
	activeFile := ui.UseStoreSelector(ideStore, func(s AppState) string { return s.ActiveFile })
	showCommandPalette := ui.UseStoreSelector(ideStore, func(s AppState) bool { return s.ShowCommandPalette })
	selectedTab := ui.UseStoreSelector(ideStore, func(s AppState) string { return s.SelectedTab })
	editorContent := ui.UseStoreSelector(ideStore, func(s AppState) string { return s.EditorContent })

	return ui.VStack(
		MenuBar(),
		MainArea(activeFile, selectedTab, editorContent),
		StatusBar(activeFile, selectedTab),
		// Command Palette Modal
		renderCommandPalette(showCommandPalette),
	)
}

// MenuBar shows the top menu
func MenuBar() ui.VNode {
	return ui.VStack(
		ui.NewTextBuilder("╔══════════════════════════════════════════════════════════════════════════════════════════════════╗").
			FgColor("blue").
			Build(),
		ui.HStack(
			ui.NewTextBuilder("║ ").
				FgColor("blue").
				Build(),
			ui.NewButtonBuilder("[File]").
				BgColor("blue").
				FgColor("white").
				Build(),
			ui.Text(" "),
			ui.NewButtonBuilder("[Edit]").
				BgColor("blue").
				FgColor("white").
				Build(),
			ui.Text(" "),
			ui.NewButtonBuilder("[View]").
				BgColor("blue").
				FgColor("white").
				Build(),
			ui.Text(" "),
			ui.NewButtonBuilder("[Run]").
				BgColor("blue").
				FgColor("white").
				Build(),
			ui.Text("   "),
			ui.NewTextBuilder("Mint IDE").
				Bold(true).
				FgColor("yellow").
				Build(),
			ui.NewTextBuilder("                                                        ").
				Build(),
			ui.NewButtonBuilder("[Ctrl+P] Command Palette").
				BgColor("yellow").
				FgColor("black").
				OnPress(ShowCommandPaletteIntent{}).
				Build(),
			ui.NewTextBuilder("  ║").
				FgColor("blue").
				Build(),
		),
		ui.NewTextBuilder("╠══════════════════════════════════════════════════════════════════════════════════════════════════╣").
			FgColor("blue").
			Build(),
	)
}

// MainArea shows the main content area
func MainArea(
	activeFile string,
	selectedTab string,
	editorContent string,
) ui.VNode {
	return ui.HStack(
		FileExplorer(activeFile),
		ui.Text(" "),
		ContentArea(selectedTab, activeFile, editorContent),
		ui.Text(" "),
	)
}

// FileExplorer shows the file tree
func FileExplorer(activeFile string) ui.VNode {
	files := []struct {
		name     string
		isFolder bool
		indent   int
	}{
		{"src", true, 0},
		{"ui", true, 1},
		{"components.go", false, 2},
		{"layout.go", false, 2},
		{"hooks.go", false, 2},
		{"core", true, 1},
		{"runtime.go", false, 2},
		{"reconciler.go", false, 2},
		{"pkg", true, 0},
		{"main.go", false, 1},
		{"go.mod", false, 0},
	}

	var children []ui.VNode
	children = append(children,
		ui.NewTextBuilder("┌─ Explorer ──────────────────────────┐").
			FgColor("gray").
			Build(),
	)

	for _, f := range files {
		indent := ""
		for i := 0; i < f.indent; i++ {
			indent += "  "
		}
		prefix := "  "
		if f.isFolder {
			prefix = "> "
		}

		isActive := activeFile == f.name
		var item ui.VNode
		if isActive {
			item = ui.NewTextBuilder("│ "+indent+prefix+f.name+" ").
				FgColor("gray").
				BgColor("blue").
				FgColor("white").
				Build()
		} else {
			item = ui.NewTextBuilder("│ "+indent+prefix+f.name).
				FgColor("gray").
				Build()
		}

		// Make clickable
		if !f.isFolder {
			item = ui.HStack(
				item,
				ui.NewButtonBuilder(" ").
					OnPress(SetActiveFileIntent{Name: f.name}).
					Build(),
			)
		}

		children = append(children, item)
	}

	children = append(children,
		ui.NewTextBuilder("│                                   │").
			FgColor("gray").
			Build(),
		ui.NewTextBuilder("└───────────────────────────────────┘").
			FgColor("gray").
			Build(),
	)

	return ui.VStack(children...)
}

// ContentArea shows the main content with tabs
func ContentArea(
	selectedTab string,
	activeFile string,
	editorContent string,
) ui.VNode {
	return ui.VStack(
		TabsBar(selectedTab, activeFile),
		ui.Text(""),
		TabContent(selectedTab, activeFile, editorContent),
	)
}

// TabsBar shows tab navigation
func TabsBar(selectedTab string, activeFile string) ui.VNode {
	tabs := []struct {
		id    string
		label string
	}{
		{"editor", activeFile},
		{"terminal", "Terminal"},
		{"problems", "Problems (0)"},
		{"output", "Output"},
	}

	var children []ui.VNode
	children = append(children,
		ui.NewTextBuilder("┌─").
			FgColor("gray").
			Build(),
	)

	for _, tab := range tabs {
		isActive := selectedTab == tab.id
		var tabBtn ui.VNode
		if isActive {
			tabBtn = ui.NewButtonBuilder(" "+tab.label+" ").
				BgColor("white").
				FgColor("black").
				OnPress(SetSelectedTabIntent{TabID: tab.id}).
				Build()
		} else {
			tabBtn = ui.NewButtonBuilder(" "+tab.label+" ").
				OnPress(SetSelectedTabIntent{TabID: tab.id}).
				Build()
		}
		children = append(children, tabBtn)
	}

	children = append(children,
		ui.NewTextBuilder("────────────────────────────────────────────────────────────────────────────────────────┐").
			FgColor("gray").
			Build(),
	)

	return ui.HStack(children...)
}

// TabContent renders the selected tab content
func TabContent(selectedTab string, activeFile string, editorContent string) ui.VNode {
	switch selectedTab {
	case "editor":
		return Editor(activeFile, editorContent)
	case "terminal":
		return Terminal(activeFile)
	case "problems":
		return Problems(activeFile)
	case "output":
		return Output(activeFile)
	default:
		return ui.Text("Unknown tab")
	}
}

// Editor shows the code editor
func Editor(activeFile string, editorContent string) ui.VNode {
	return ui.VStack(
		ui.HStack(
			ui.NewTextBuilder("│ ").
				FgColor("gray").
				Build(),
			ui.NewTextBuilder(activeFile).
				Bold(true).
				FgColor("yellow").
				Build(),
			ui.NewTextBuilder("                                                        ║").
				FgColor("gray").
				Build(),
		),
		ui.NewTextBuilder("│").FgColor("gray").Build(),
		ui.HStack(
			ui.NewTextBuilder("│ ").
				FgColor("gray").
				Build(),
			ui.NewTextareaBuilder().
				Placeholder("// Code here...").
				Value(editorContent).
				Rows(15).
				Cols(75).
				Build(),
			ui.NewTextBuilder("  ║").
				FgColor("gray").
				Build(),
		),
		ui.NewTextBuilder("└────────────────────────────────────────────────────────────────────────────────────────────┘").
			FgColor("gray").
			Build(),
	)
}

// Terminal shows the terminal output
func Terminal(activeFile string) ui.VNode {
	return ui.VStack(
		ui.HStack(
			ui.NewTextBuilder("│ ").
				FgColor("gray").
				Build(),
			ui.NewTextBuilder(fmt.Sprintf("Terminal - %s", activeFile)).
				Bold(true).
				FgColor("yellow").
				Build(),
			ui.NewTextBuilder("                                ║").
				FgColor("gray").
				Build(),
		),
		ui.NewTextBuilder("│").FgColor("gray").Build(),
		ui.HStack(
			ui.NewTextBuilder("│ ").
				FgColor("gray").
				Build(),
			ui.Text("> Building..."),
			ui.NewTextBuilder("                                                            ║").
				FgColor("gray").
				Build(),
		),
		ui.HStack(
			ui.NewTextBuilder("│ ").
				FgColor("gray").
				Build(),
			ui.NewTextBuilder("> Done in 1.2s").
				FgColor("green").
				Build(),
			ui.NewTextBuilder("                                                              ║").
				FgColor("gray").
				Build(),
		),
		ui.NewTextBuilder("└────────────────────────────────────────────────────────────────────────────────────────────┘").
			FgColor("gray").
			Build(),
	)
}

// Problems shows the problems panel
func Problems(activeFile string) ui.VNode {
	return ui.VStack(
		ui.HStack(
			ui.NewTextBuilder("│ ").
				FgColor("gray").
				Build(),
			ui.NewTextBuilder(fmt.Sprintf("Problems - %s", activeFile)).
				Bold(true).
				FgColor("yellow").
				Build(),
			ui.NewTextBuilder("                                  ║").
				FgColor("gray").
				Build(),
		),
		ui.NewTextBuilder("│").FgColor("gray").Build(),
		ui.HStack(
			ui.NewTextBuilder("│ ").
				FgColor("gray").
				Build(),
			ui.NewTextBuilder("0 errors").
				FgColor("green").
				Build(),
			ui.NewTextBuilder("                                                             ║").
				FgColor("gray").
				Build(),
		),
		ui.HStack(
			ui.NewTextBuilder("│ ").
				FgColor("gray").
				Build(),
			ui.NewTextBuilder("0 warnings").
				FgColor("yellow").
				Build(),
			ui.NewTextBuilder("                                                            ║").
				FgColor("gray").
				Build(),
		),
		ui.NewTextBuilder("└────────────────────────────────────────────────────────────────────────────────────────────┘").
			FgColor("gray").
			Build(),
	)
}

// Output shows the output panel
func Output(activeFile string) ui.VNode {
	return ui.VStack(
		ui.HStack(
			ui.NewTextBuilder("│ ").
				FgColor("gray").
				Build(),
			ui.NewTextBuilder(fmt.Sprintf("Output - %s", activeFile)).
				Bold(true).
				FgColor("yellow").
				Build(),
			ui.NewTextBuilder("                                   ║").
				FgColor("gray").
				Build(),
		),
		ui.NewTextBuilder("│").FgColor("gray").Build(),
		ui.HStack(
			ui.NewTextBuilder("│ ").
				FgColor("gray").
				Build(),
			ui.Text("[INFO] Starting build..."),
			ui.NewTextBuilder("                                                             ║").
				FgColor("gray").
				Build(),
		),
		ui.HStack(
			ui.NewTextBuilder("│ ").
				FgColor("gray").
				Build(),
			ui.NewTextBuilder("[INFO] Build complete").
				FgColor("green").
				Build(),
			ui.NewTextBuilder("                                                              ║").
				FgColor("gray").
				Build(),
		),
		ui.NewTextBuilder("└────────────────────────────────────────────────────────────────────────────────────────────┘").
			FgColor("gray").
			Build(),
	)
}

// StatusBar shows the bottom status bar
func StatusBar(activeFile string, selectedTab string) ui.VNode {
	return ui.VStack(
		ui.NewTextBuilder("╠══════════════════════════════════════════════════════════════════════════════════════════════════╣").
			FgColor("blue").
			Build(),
		ui.HStack(
			ui.NewTextBuilder("║ ").
				FgColor("blue").
				Build(),
			ui.NewTextBuilder(fmt.Sprintf("%s | %s", activeFile, selectedTab)).
				FgColor("white").
				Build(),
			ui.Text("  "),
			ui.NewTextBuilder("Go").
				BgColor("cyan").
				FgColor("black").
				Build(),
			ui.Text("  "),
			ui.NewTextBuilder("Ln 1, Col 1").
				FgColor("white").
				Build(),
			ui.NewTextBuilder("                                                                               ║").
				FgColor("blue").
				Build(),
		),
		ui.NewTextBuilder("╚══════════════════════════════════════════════════════════════════════════════════════════════════╝").
			FgColor("blue").
			Build(),
	)
}

// renderCommandPalette renders the command palette modal
func renderCommandPalette(showModal bool) ui.VNode {
	if !showModal {
		return ui.Fragment()
	}

	// Modal content
	modalBox := ui.VStack(
		ui.NewTextBuilder("╔═════════════════════════════════════════════════╗").
			FgColor("yellow").
			Build(),
		ui.HStack(
			ui.NewTextBuilder("║ ").
				FgColor("yellow").
				Build(),
			ui.NewInputBuilder().
				Placeholder("Type a command...").
				Width(35).
				Build(),
			ui.NewTextBuilder(" ║").
				FgColor("yellow").
				Build(),
		),
		ui.HStack(
			ui.NewTextBuilder("║ ").
				FgColor("yellow").
				Build(),
			ui.NewTextBuilder("Go to file...").
				Build(),
			ui.NewTextBuilder("                                          ║").
				FgColor("yellow").
				Build(),
		),
		ui.HStack(
			ui.NewTextBuilder("║ ").
				FgColor("yellow").
				Build(),
			ui.NewTextBuilder("Go to symbol...").
				Build(),
			ui.NewTextBuilder("                                          ║").
				FgColor("yellow").
				Build(),
		),
		ui.HStack(
			ui.NewTextBuilder("║ ").
				FgColor("yellow").
				Build(),
			ui.NewTextBuilder("Search files...").
				Build(),
			ui.NewTextBuilder("                                          ║").
				FgColor("yellow").
				Build(),
		),
		ui.HStack(
			ui.NewTextBuilder("║ ").
				FgColor("yellow").
				Build(),
			ui.NewTextBuilder("Press ESC to close").
				FgColor("gray").
				Build(),
			ui.NewTextBuilder("                                         ║").
				FgColor("yellow").
				Build(),
		),
		ui.NewTextBuilder("╚═════════════════════════════════════════════════╝").
			FgColor("yellow").
			Build(),
	)

	return ui.Modal(modalBox).
		CloseOnESC(true).
		CloseOnBackdropClick(true).
		Build()
}

// Demo 5: IDE Interface Demo
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

	"github.com/wwsheng009/mint/app"
	"github.com/wwsheng009/mint/ui"
)

// Intent Types
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

// Global setters for handling dynamic intents
// Note: This pattern uses global variables to work around the ui.On deduplication issue
// See docs/architecture/mvp/INTENT_MANAGEMENT_PATTERNS.md for details
var (
	globalSetShowCommandPalette func(bool)
	globalSetActiveFile         func(string)
	globalSetSelectedTab        func(string)
)

func main() {
	// Register intent handlers once at application start
	ui.On(ShowCommandPaletteIntent{}, func() {
		if globalSetShowCommandPalette != nil {
			globalSetShowCommandPalette(true)
		}
	})
	ui.On(HideCommandPaletteIntent{}, func() {
		if globalSetShowCommandPalette != nil {
			globalSetShowCommandPalette(false)
		}
	})

	err := ui.Run(IDEDemo,
		ui.WithWidth(100),
		ui.WithHeight(40),
		ui.WithTitle("Mint TUI - IDE Interface"),
	)
	if err != nil {
		panic(err)
	}
}

// IDEDemo is the root component
func IDEDemo() ui.VNode {
	activeFile, setActiveFile := ui.UseStateString("main.go")
	showCommandPalette, setShowCommandPalette := ui.UseStateBool(false)
	selectedTab, setSelectedTab := ui.UseStateString("editor")
	editorContent, setEditorContent := ui.UseStateString("func main() {\\n    ui.Run(App)\\n}")

	// Update global setters
	globalSetShowCommandPalette = setShowCommandPalette
	globalSetActiveFile = setActiveFile
	globalSetSelectedTab = setSelectedTab

	// Register file selection handler
	ui.On(SetActiveFileIntent{Name: activeFile}, func() {
		if globalSetActiveFile != nil {
			globalSetActiveFile(activeFile)
		}
	})

	// Register tab selection handler
	ui.On(SetSelectedTabIntent{TabID: selectedTab}, func() {
		if globalSetSelectedTab != nil {
			globalSetSelectedTab(selectedTab)
		}
	})

	return ui.VStack(
		MenuBar(setShowCommandPalette),
		MainArea(activeFile, setActiveFile, selectedTab, setSelectedTab, editorContent, setEditorContent),
		StatusBar(activeFile, selectedTab),
		// Command Palette Modal
		renderCommandPalette(showCommandPalette, setShowCommandPalette),
	)
}

// MenuBar shows the top menu
func MenuBar(setShowCommandPalette func(bool)) ui.VNode {
	return ui.VStack(
		app.NewTextBuilder("╔══════════════════════════════════════════════════════════════════════════════════════════════════╗").
			FgColor("blue").
			Build(),
		ui.HStack(
			app.NewTextBuilder("║ ").
				FgColor("blue").
				Build(),
			app.ButtonBuilder("[File]").
				BgColor("blue").
				FgColor("white").
				Build(),
			ui.Text(" "),
			app.ButtonBuilder("[Edit]").
				BgColor("blue").
				FgColor("white").
				Build(),
			ui.Text(" "),
			app.ButtonBuilder("[View]").
				BgColor("blue").
				FgColor("white").
				Build(),
			ui.Text(" "),
			app.ButtonBuilder("[Run]").
				BgColor("blue").
				FgColor("white").
				Build(),
			ui.Text("   "),
			app.NewTextBuilder("Mint IDE").
				Bold(true).
				FgColor("yellow").
				Build(),
			app.NewTextBuilder("                                                        ").
				Build(),
			app.ButtonBuilder("[Ctrl+P] Command Palette").
				BgColor("yellow").
				FgColor("black").
				OnPress(ShowCommandPaletteIntent{}).
				Build(),
			app.NewTextBuilder("  ║").
				FgColor("blue").
				Build(),
		),
		app.NewTextBuilder("╠══════════════════════════════════════════════════════════════════════════════════════════════════╣").
			FgColor("blue").
			Build(),
	)
}

// MainArea shows the main content area
func MainArea(
	activeFile string,
	setActiveFile func(string),
	selectedTab string,
	setSelectedTab func(string),
	editorContent string,
	setEditorContent func(string),
) ui.VNode {
	return ui.HStack(
		FileExplorer(activeFile, setActiveFile),
		ui.Text(" "),
		ContentArea(selectedTab, setSelectedTab, activeFile, editorContent, setEditorContent),
		ui.Text(" "),
	)
}

// FileExplorer shows the file tree
func FileExplorer(activeFile string, setActiveFile func(string)) ui.VNode {
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
		{" reconciler.go", false, 2},
		{"pkg", true, 0},
		{"main.go", false, 1},
		{"go.mod", false, 0},
	}

	var children []ui.VNode
	children = append(children,
		app.NewTextBuilder("┌─ Explorer ──────────────────────────┐").
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
			item = app.NewTextBuilder("│ "+indent+prefix+f.name+" ").
				FgColor("gray").
				BgColor("blue").
				FgColor("white").
				Build()
		} else {
			item = app.NewTextBuilder("│ "+indent+prefix+f.name).
				FgColor("gray").
				Build()
		}

		// Make clickable
		if !f.isFolder {
			item = ui.HStack(
				item,
				app.ButtonBuilder(" ").
					OnPress(SetActiveFileIntent{Name: f.name}).
					Build(),
			)
		}

		children = append(children, item)
	}

	children = append(children,
		app.NewTextBuilder("│                                   │").
			FgColor("gray").
			Build(),
		app.NewTextBuilder("└───────────────────────────────────┘").
			FgColor("gray").
			Build(),
	)

	return ui.VStack(children...)
}

// ContentArea shows the main content with tabs
func ContentArea(
	selectedTab string,
	setSelectedTab func(string),
	activeFile string,
	editorContent string,
	setEditorContent func(string),
) ui.VNode {
	return ui.VStack(
		TabsBar(selectedTab, setSelectedTab, activeFile),
		ui.Text(""),
		TabContent(selectedTab, activeFile, editorContent, setEditorContent),
	)
}

// TabsBar shows tab navigation
func TabsBar(selectedTab string, setSelectedTab func(string), activeFile string) ui.VNode {
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
		app.NewTextBuilder("┌─").
			FgColor("gray").
			Build(),
	)

	for _, tab := range tabs {
		isActive := selectedTab == tab.id
		var tabBtn ui.VNode
		if isActive {
			tabBtn = app.ButtonBuilder(" "+tab.label+" ").
				BgColor("white").
				FgColor("black").
				OnPress(SetSelectedTabIntent{TabID: tab.id}).
				Build()
		} else {
			tabBtn = app.ButtonBuilder(" "+tab.label+" ").
				OnPress(SetSelectedTabIntent{TabID: tab.id}).
				Build()
		}
		children = append(children, tabBtn)
	}

	children = append(children,
		app.NewTextBuilder("────────────────────────────────────────────────────────────────────────────────────────┐").
			FgColor("gray").
			Build(),
	)

	return ui.HStack(children...)
}

// TabContent renders the selected tab content
func TabContent(
	selectedTab string,
	activeFile string,
	editorContent string,
	setEditorContent func(string),
) ui.VNode {
	if selectedTab == "editor" {
		return EditorPanel(activeFile, editorContent, setEditorContent)
	} else if selectedTab == "terminal" {
		return TerminalPanel()
	}
	return app.NewTextBuilder("│ Tab content: " + selectedTab + "                                                                │").
		FgColor("gray").
		Build()
}

// EditorPanel shows the code editor
func EditorPanel(filename string, content string, setContent func(string)) ui.VNode {
	lines := []string{
		"1 │ func main() {",
		"2 │     ui.Run(App,",
		"3 │         ui.WithSize(80, 24),",
		"4 │         ui.WithTitle(\"My App\"),",
		"5 │     )",
		"6 │ }",
	}

	var children []ui.VNode
	children = append(children,
		app.NewTextBuilder("│ ").
			FgColor("gray").
			Build(),
	)

	for i, line := range lines {
		lineNum := fmt.Sprintf("%2d", i+1)
		var lineText ui.VNode
		if i == 0 {
			lineText = app.NewTextBuilder("│ "+lineNum+" │ "+line+"                                                      │").
				FgColor("gray").
				BgColor("blue").
				FgColor("white").
				Build()
		} else {
			lineText = app.NewTextBuilder("│ "+lineNum+" │ "+line+"                                                      │").
				FgColor("gray").
				Build()
		}
		children = append(children, lineText)
	}

	children = append(children,
		app.NewTextBuilder("│                                                                                             │").
			FgColor("gray").
			Build(),
		app.NewTextBuilder("└────────────────────────────────────────────────────────────────────────────────────────────┘").
			FgColor("gray").
			Build(),
	)

	return ui.VStack(children...)
}

// TerminalPanel shows the console output
func TerminalPanel() ui.VNode {
	logs := []string{
		"[INFO] Starting build...",
		"[INFO] Compiling main.go",
		"[SUCCESS] Build complete in 1.2s",
	}

	var children []ui.VNode
	children = append(children,
		app.NewTextBuilder("│ > Terminal                                                                                │").
			FgColor("gray").
			Build(),
		ui.Text(""),
	)

	for _, log := range logs {
		var color string
		if len(log) > 0 {
			switch log[1:5] {
			case "INFO":
				color = "blue"
			case "SUCC":
				color = "green"
			case "WARN":
				color = "yellow"
			case "ERRO":
				color = "red"
			default:
				color = "white"
			}
		}

		children = append(children,
			app.NewTextBuilder("│ "+log).
				FgColor(color).
				Build(),
		)
	}

	children = append(children,
		ui.Text(""),
		app.NewTextBuilder("│ > _                                                                                      │").
			FgColor("gray").
			Build(),
		app.NewTextBuilder("└────────────────────────────────────────────────────────────────────────────────────────────┘").
			FgColor("gray").
			Build(),
	)

	return ui.VStack(children...)
}

// StatusBar shows the status bar
func StatusBar(activeFile, selectedTab string) ui.VNode {
	return ui.VStack(
		app.NewTextBuilder("╠══════════════════════════════════════════════════════════════════════════════════════════════════╣").
			FgColor("blue").
			Build(),
		ui.HStack(
			app.NewTextBuilder("║ ").
				FgColor("blue").
				Build(),
			app.NewTextBuilder("Ln 1, Col 1").
				FgColor("white").
				Build(),
			app.NewTextBuilder("  ").
				Build(),
			app.NewTextBuilder("UTF-8").
				FgColor("green").
				Build(),
			app.NewTextBuilder("  ").
				Build(),
			app.NewTextBuilder("Go").
				BgColor("cyan").
				FgColor("black").
				Build(),
			app.NewTextBuilder("  ").
				Build(),
			app.NewTextBuilder(activeFile).
				FgColor("yellow").
				Build(),
			app.NewTextBuilder("                                                      ").
				Build(),
			app.NewTextBuilder("✓ Ready").
				FgColor("green").
				Build(),
			app.NewTextBuilder("  ║").
				FgColor("blue").
				Build(),
		),
		app.NewTextBuilder("╚══════════════════════════════════════════════════════════════════════════════════════════════════╝").
			FgColor("blue").
			Build(),
	)
}

// renderCommandPalette shows the command palette modal
func renderCommandPalette(show bool, setShow func(bool)) ui.VNode {
	if !show {
		return ui.Text("")
	}

	commands := []string{
		"File: New File",
		"File: Open File",
		"File: Save",
		"File: Save All",
		"Edit: Find",
		"Edit: Replace",
		"View: Toggle Sidebar",
		"Run: Start Debugging",
	}

	var children []ui.VNode
	children = append(children,
		ui.Text(""),
		ui.HStack(
			ui.Text("         "),
			ui.VStack(
				app.NewTextBuilder("╔══════════════════════════════════════════════════╗").
					FgColor("yellow").
					Build(),
				app.NewTextBuilder("║ > Type command...                             ║").
					FgColor("yellow").
					Build(),
				app.NewTextBuilder("╠══════════════════════════════════════════════════╣").
					FgColor("yellow").
					Build(),
			),
		),
	)

	for _, cmd := range commands {
		children = append(children,
			ui.HStack(
				ui.Text("         "),
				app.NewTextBuilder("║ "+cmd).
					FgColor("yellow").
					Build(),
				app.NewTextBuilder("                              ║").
					FgColor("yellow").
					Build(),
			),
		)
	}

	children = append(children,
		ui.HStack(
			ui.Text("         "),
			app.NewTextBuilder("╚══════════════════════════════════════════════════╝").
				FgColor("yellow").
				Build(),
		),
		ui.Text(""),
		app.NewTextBuilder("Press ESC to close").
			FgColor("gray").
			Build(),
	)

	return ui.VStack(
		ui.Text(""),
		ui.Text(""),
		ui.VStack(children...),
	)
}

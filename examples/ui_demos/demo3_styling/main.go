// Demo 3: Styling System (TUI CSS) (Store 模式)
//
// This demo demonstrates the styling system, which is essentially
// CSS Box Model + Terminal color/attribute system + inheritance rules.
//
// Features:
// - Box Model (Padding, Margin, Border)
// - Color (Foreground, Background)
// - Text Attributes (Bold, Italic, Underline)
// - Style Inheritance
// - Theme System
//
// Based on: framework/docs/ui/demo/demo3_with_style.md

package main

import (
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/reducer"
	"github.com/wwsheng009/mint/runtime/store"
	"github.com/wwsheng009/mint/ui"
)

// Intent Types
type SetStylingTabIntent struct {
	TabID string
}
func (SetStylingTabIntent) IntentType() string { return "SetStylingTab" }
func (SetStylingTabIntent) StayPressed() bool  { return true }

// AppState - 定义应用状态
type AppState struct {
	CurrentTab string // 当前标签: colors, attributes, borders...
}

// Store 初始化
var stylingStore = store.NewStore(AppState{
	CurrentTab: "colors",
})

// Reducer 注册
func init() {
	reducer.NewBuilder[AppState]().
		On(SetStylingTabIntent{}, func(s AppState, i intent.Intent) AppState {
			s.CurrentTab = i.(SetStylingTabIntent).TabID
			return s
		}).
		BuildAndRegister(intent.DefaultRegistry(), stylingStore)
}

func main() {
	err := ui.Run(StylingDemo,
		ui.WithWidth(100),
		ui.WithHeight(40),
		ui.WithTitle("Mint TUI - Styling System (Store 模式)"),
	)
	if err != nil {
		panic(err)
	}
}

// StylingDemo is the root component
func StylingDemo() ui.VNode {
	// ✅ 订阅 currentTab 状态
	currentTab := ui.UseStoreSelector(stylingStore, func(s AppState) string { return s.CurrentTab })

	return ui.VStack(
		HeaderPanel(),
		TabNavigation(currentTab),
		ui.Text(""),
		renderTabContent(currentTab),
	)
}

// HeaderPanel shows the title
func HeaderPanel() ui.VNode {
	return ui.VStack(
		ui.NewTextBuilder("╔══════════════════════════════════════════════════════════════════════════════════════════╗").
			FgColor("cyan").
			Build(),
		ui.HStack(
			ui.NewTextBuilder("║ ").
				FgColor("cyan").
				Build(),
			ui.NewTextBuilder("                        Styling System Demo (TUI CSS)").
				Bold(true).
				FgColor("white").
				Build(),
			ui.NewTextBuilder("                                  ║").
				FgColor("cyan").
				Build(),
		),
		ui.NewTextBuilder("╚══════════════════════════════════════════════════════════════════════════════════════════╝").
			FgColor("cyan").
			Build(),
	)
}

// TabNavigation provides tab buttons
func TabNavigation(currentTab string) ui.VNode {
	tabs := []struct {
		id    string
		label string
		color string
	}{
		{"colors", "Colors", "red"},
		{"attributes", "Attributes", "yellow"},
		{"borders", "Borders", "green"},
		{"inheritance", "Inheritance", "blue"},
		{"themes", "Themes", "magenta"},
	}

	var children []ui.VNode
	for _, tab := range tabs {
		isActive := currentTab == tab.id
		var btn ui.VNode
		if isActive {
			btn = ui.NewButtonBuilder("[" + tab.label + "]").
				BgColor(tab.color).
				FgColor("white").
				OnPress(SetStylingTabIntent{TabID: tab.id}).
				Build()
		} else {
			btn = ui.NewButtonBuilder(" " + tab.label + " ").
				FgColor(tab.color).
				OnPress(SetStylingTabIntent{TabID: tab.id}).
				Build()
		}
		children = append(children, btn, ui.Text(" "))
	}

	return ui.HStack(children...)
}

// renderTabContent renders the selected tab content
func renderTabContent(currentTab string) ui.VNode {
	switch currentTab {
	case "colors":
		return ColorsTab()
	case "attributes":
		return AttributesTab()
	case "borders":
		return BordersTab()
	case "inheritance":
		return InheritanceTab()
	case "themes":
		return ThemesTab()
	default:
		return ui.NewTextBuilder("Unknown tab").Build()
	}
}

// ColorsTab demonstrates color system
func ColorsTab() ui.VNode {
	colors := []struct {
		name  string
		color string
	}{
		{"Red", "red"},
		{"Green", "green"},
		{"Blue", "blue"},
		{"Yellow", "yellow"},
		{"Cyan", "cyan"},
		{"Magenta", "magenta"},
		{"White", "white"},
		{"Black", "black"},
		{"Gray", "gray"},
		{"Bright Red", "bright-red"},
		{"Bright Green", "bright-green"},
		{"Bright Blue", "bright-blue"},
		{"Bright Yellow", "bright-yellow"},
		{"Bright Cyan", "bright-cyan"},
		{"Bright Magenta", "bright-magenta"},
		{"Bright White", "bright-white"},
	}

	var children []ui.VNode
	children = append(children,
		ui.NewTextBuilder("┌─ Color Palette ─────────────────────────────────────────────────────────────────────────┐").
			FgColor("gray").
			Build(),
	)

	for i, c := range colors {
		if i%2 == 0 {
			children = append(children,
				ui.HStack(
					ui.NewTextBuilder("│ ").
						FgColor("gray").
						Build(),
					renderColorSwatch(c.name, c.color),
					ui.Text("   "),
					renderColorSwatch(
						func() string {
							if i+1 < len(colors) {
								return colors[i+1].name
							}
							return ""
						}(),
						func() string {
							if i+1 < len(colors) {
								return colors[i+1].color
							}
							return ""
						}(),
					),
					ui.NewTextBuilder(" │").
						FgColor("gray").
						Build(),
				),
			)
		}
	}

	children = append(children,
		ui.NewTextBuilder("└──────────────────────────────────────────────────────────────────────────────────────────────┘").
			FgColor("gray").
			Build(),
	)

	return ui.VStack(children...)
}

// renderColorSwatch renders a color swatch
func renderColorSwatch(name, color string) ui.VNode {
	if name == "" {
		return ui.Text("                          ")
	}
	return ui.HStack(
		ui.NewTextBuilder("■ ").
			FgColor(color).
			Bold(true).
			Build(),
		ui.NewTextBuilder(name).
			FgColor(color).
			Bold(true).
			Build(),
	)
}

// AttributesTab demonstrates text attributes
func AttributesTab() ui.VNode {
	return ui.VStack(
		ui.NewTextBuilder("┌─ Text Attributes ─────────────────────────────────────────────────────────────────────────┐").
			FgColor("gray").
			Build(),
		ui.Text(""),
		ui.HStack(
			ui.NewTextBuilder("│ ").
				FgColor("gray").
				Build(),
			renderAttribute("Normal Text", "", "", false, false, false),
			ui.NewTextBuilder("     │").
				FgColor("gray").
				Build(),
		),
		ui.Text(""),
		ui.HStack(
			ui.NewTextBuilder("│ ").
				FgColor("gray").
				Build(),
			renderAttribute("Bold Text", "", "", true, false, false),
			ui.NewTextBuilder("       │").
				FgColor("gray").
				Build(),
		),
		ui.Text(""),
		ui.HStack(
			ui.NewTextBuilder("│ ").
				FgColor("gray").
				Build(),
			renderAttribute("Italic Text", "", "", false, true, false),
			ui.NewTextBuilder("      │").
				FgColor("gray").
				Build(),
		),
		ui.Text(""),
		ui.HStack(
			ui.NewTextBuilder("│ ").
				FgColor("gray").
				Build(),
			renderAttribute("Underline Text", "", "", false, false, true),
			ui.NewTextBuilder("  │").
				FgColor("gray").
				Build(),
		),
		ui.Text(""),
		ui.HStack(
			ui.NewTextBuilder("│ ").
				FgColor("gray").
				Build(),
			renderAttribute("Bold + Italic", "", "", true, true, false),
			ui.NewTextBuilder("    │").
				FgColor("gray").
				Build(),
		),
		ui.Text(""),
		ui.HStack(
			ui.NewTextBuilder("│ ").
				FgColor("gray").
				Build(),
			renderAttribute("All Styles", "", "", true, true, true),
			ui.NewTextBuilder("         │").
				FgColor("gray").
				Build(),
		),
		ui.Text(""),
		ui.NewTextBuilder("└──────────────────────────────────────────────────────────────────────────────────────────────┘").
			FgColor("gray").
			Build(),
	)
}

// renderAttribute renders a text with attributes
func renderAttribute(text, fg, bg string, bold, italic, underline bool) ui.VNode {
	builder := ui.NewTextBuilder(text)
	if fg != "" {
		builder.FgColor(fg)
	}
	if bg != "" {
		builder.BgColor(bg)
	}
	if bold {
		builder.Bold(true)
	}
	if italic {
		builder.Italic(true)
	}
	if underline {
		builder.Underline(true)
	}
	return builder.Build()
}

// BordersTab demonstrates border styles
func BordersTab() ui.VNode {
	return ui.VStack(
		ui.NewTextBuilder("┌─ Border Styles ────────────────────────────────────────────────────────────────────────────┐").
			FgColor("gray").
			Build(),
		ui.Text(""),
		ui.HStack(
			ui.NewTextBuilder("│ ").
				FgColor("gray").
				Build(),
			ui.Text("Single: ┌───┐"),
			ui.Text("   Double: ╔═══╗"),
			ui.Text("   Rounded: ╭───╮"),
			ui.NewTextBuilder("    │").
				FgColor("gray").
				Build(),
		),
		ui.Text(""),
		ui.HStack(
			ui.NewTextBuilder("│ ").
				FgColor("gray").
				Build(),
			ui.Text("        │   │"),
			ui.Text("           ║   ║"),
			ui.Text("           │   │"),
			ui.NewTextBuilder("    │").
				FgColor("gray").
				Build(),
		),
		ui.Text(""),
		ui.HStack(
			ui.NewTextBuilder("│ ").
				FgColor("gray").
				Build(),
			ui.Text("        └───┘"),
			ui.Text("           ╚═══╝"),
			ui.Text("           ╰───╯"),
			ui.NewTextBuilder("    │").
				FgColor("gray").
				Build(),
		),
		ui.Text(""),
		ui.HStack(
			ui.NewTextBuilder("│ ").
				FgColor("gray").
				Build(),
			ui.Text("Dashed: ┏ ┳ ┓     Thick: ┏━━━┓     Dotted: ╌╌╌╌┐"),
			ui.NewTextBuilder("    │").
				FgColor("gray").
				Build(),
		),
		ui.Text(""),
		ui.NewTextBuilder("└──────────────────────────────────────────────────────────────────────────────────────────────┘").
			FgColor("gray").
			Build(),
	)
}

// InheritanceTab demonstrates style inheritance
func InheritanceTab() ui.VNode {
	return ui.VStack(
		ui.NewTextBuilder("┌─ Style Inheritance ─────────────────────────────────────────────────────────────────────────┐").
			FgColor("gray").
			Build(),
		ui.Text(""),
		ui.HStack(
			ui.NewTextBuilder("│ ").
				FgColor("gray").
				Build(),
			ui.VStack(
				ui.NewTextBuilder("Parent: Blue + Bold").
					FgColor("blue").
					Bold(true).
					Build(),
				ui.HStack(
					ui.Text("  └─ Child 1: Inherits parent style"),
					ui.Text("      "),
				),
				ui.HStack(
					ui.Text("  └─ Child 2: "),
					ui.NewTextBuilder("Overrides to Red").
						FgColor("red").
						Bold(true).
						Build(),
				),
			),
			ui.NewTextBuilder("                                                            │").
				FgColor("gray").
				Build(),
		),
		ui.Text(""),
		ui.NewTextBuilder("│ Inheritance Rules:                                                                           │").
			FgColor("gray").
			Build(),
		ui.HStack(
			ui.NewTextBuilder("│ ").
				FgColor("gray").
				Build(),
			ui.NewTextBuilder("FgColor").
				Bold(true).
				Build(),
			ui.Text(" - Inherited by default"),
			ui.NewTextBuilder("                                            │").
				FgColor("gray").
				Build(),
		),
		ui.HStack(
			ui.NewTextBuilder("│ ").
				FgColor("gray").
				Build(),
			ui.NewTextBuilder("BgColor").
				Bold(true).
				Build(),
			ui.Text(" - Not inherited (must be set explicitly)"),
			ui.NewTextBuilder("                             │").
				FgColor("gray").
				Build(),
		),
		ui.HStack(
			ui.NewTextBuilder("│ ").
				FgColor("gray").
				Build(),
			ui.NewTextBuilder("Bold/Italic/Underline").
				Bold(true).
				Build(),
			ui.Text(" - Inherited by default"),
			ui.NewTextBuilder("                              │").
				FgColor("gray").
				Build(),
		),
		ui.Text(""),
		ui.NewTextBuilder("└──────────────────────────────────────────────────────────────────────────────────────────────┘").
			FgColor("gray").
			Build(),
	)
}

// ThemesTab demonstrates theme system
func ThemesTab() ui.VNode {
	themes := []struct {
		name     string
		primary  string
		secondary string
		bg       string
		fg       string
	}{
		{"Default", "blue", "cyan", "black", "white"},
		{"Dark", "bright-blue", "bright-cyan", "black", "bright-white"},
		{"Ocean", "cyan", "blue", "blue", "white"},
		{"Forest", "green", "bright-green", "green", "white"},
		{"Sunset", "yellow", "red", "yellow", "black"},
		{"Violet", "magenta", "bright-magenta", "magenta", "white"},
	}

	var children []ui.VNode
	children = append(children,
		ui.NewTextBuilder("┌─ Theme System ─────────────────────────────────────────────────────────────────────────────┐").
			FgColor("gray").
			Build(),
		ui.Text(""),
	)

	for _, theme := range themes {
		children = append(children,
			ui.HStack(
				ui.NewTextBuilder("│ ").
					FgColor("gray").
					Build(),
				ui.NewTextBuilder(" "+theme.name+" ").
					BgColor(theme.bg).
					FgColor(theme.fg).
					Build(),
				ui.Text(" "),
				ui.NewButtonBuilder("[Primary]").
					BgColor(theme.primary).
					FgColor("white").
					Build(),
				ui.Text(" "),
				ui.NewButtonBuilder("[Secondary]").
					BgColor(theme.secondary).
					FgColor("white").
					Build(),
				ui.NewTextBuilder("                                                            │").
					FgColor("gray").
					Build(),
			),
		)
	}

	children = append(children,
		ui.NewTextBuilder("└──────────────────────────────────────────────────────────────────────────────────────────────┘").
			FgColor("gray").
			Build(),
	)

	return ui.VStack(children...)
}

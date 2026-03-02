// Package component_fixtures provides common UI component test fixtures
// that can be reused across multiple test files.
//
// This package is intentionally separate from runtime/compute to avoid
// import cycles, since it depends on app, framework/theme, and ui packages.
package component_fixtures

import (
	"fmt"

	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/style"
	"github.com/wwsheng009/mint/ui"
	"github.com/wwsheng009/mint/ui/components/stack"
)

// =============================================================================
// Intent Types
// =============================================================================

type ModalCloseIntent struct{}
func (ModalCloseIntent) IntentType() string { return "ModalClose" }
func (ModalCloseIntent) StayPressed() bool  { return true }

// ComponentFixture represents a testable UI component configuration
type ComponentFixture struct {
	Name        string
	Description string
	Build       func() ui.VNode
}

// Demo1Config holds configuration for Demo1-style components
type Demo1Config struct {
	Count  int
	Input  string
	Items  []string
	Theme  string
	Width  int
	Height int
}

// DefaultDemo1Config returns default configuration for Demo1 components
func DefaultDemo1Config() *Demo1Config {
	return &Demo1Config{
		Count:  0,
		Input:  "",
		Items:  []string{"Log line #0", "Log line #1", "Log line #2", "Log line #3", "Log line #4", "Log line #5"},
		Theme:  "nord",
		Width:  80,
		Height: 24,
	}
}

// Demo1ConfigOption is a functional option for configuring Demo1
type Demo1ConfigOption func(*Demo1Config)

// WithCount sets the count value
func WithCount(count int) Demo1ConfigOption {
	return func(c *Demo1Config) {
		c.Count = count
	}
}

// WithInput sets the input value
func WithInput(input string) Demo1ConfigOption {
	return func(c *Demo1Config) {
		c.Input = input
	}
}

// WithItems sets the items list
func WithItems(items []string) Demo1ConfigOption {
	return func(c *Demo1Config) {
		c.Items = items
	}
}

// WithTheme sets the theme
func WithTheme(themeName string) Demo1ConfigOption {
	return func(c *Demo1Config) {
		c.Theme = themeName
	}
}

// WithSize sets width and height
func WithSize(width, height int) Demo1ConfigOption {
	return func(c *Demo1Config) {
		c.Width = width
		c.Height = height
	}
}

// =============================================================================
// Demo1 Full-Featured Components
// =============================================================================

// BuildDemo1App builds the complete Demo1 app VNode tree
func BuildDemo1App(opts ...Demo1ConfigOption) ui.VNode {
	cfg := DefaultDemo1Config()
	for _, opt := range opts {
		opt(cfg)
	}

	_ = theme.SetTheme(cfg.Theme)

	mainContent := ui.VStackBuilder(
		BuildDemo1Header(cfg.Count),
		BuildDemo1MainBody(cfg.Count, cfg.Input, cfg.Items),
		BuildDemo1DebugPanel(),
	).Stretch().Build()

	return mainContent
}

// BuildDemo1Header builds the header component
func BuildDemo1Header(count int) ui.VNode {
	headerContent := ui.HStack(
		ui.NewTextBuilder("TUI Engine Demo").
			Style(style.FgBgBold(theme.Text(), theme.Primary())).
			Build(),
		ui.NewTextBuilder("              ").
			Style(style.FgBg(theme.Surface(), theme.Primary())).
			Build(),
		ui.NewTextBuilder(" ").
			Style(style.FgBg(theme.Surface(), theme.Primary())).
			Build(),
		ui.NewTextBuilder(fmt.Sprintf("Clicks: %d", count)).
			Style(style.FgBgBold(theme.BG(), theme.Primary())).
			Build(),
	)

	return stack.NewVStack().
		SingleBorder().
		BorderColor(string(theme.Primary())).
		SetChildrenList([]ui.VNode{headerContent})
}

// BuildDemo1MainBody builds the main body layout
func BuildDemo1MainBody(count int, input string, items []string) ui.VNode {
	sidebar := ui.VStackBuilder(
		ui.NewTextBuilder("Menu").
			Style(style.FgBoldUnderline(theme.Muted())).
			Build(),
		ui.NewTextBuilder("Add Count").
			Style(style.FgBold(theme.Primary())).
			Build(),
		ui.NewTextBuilder("Quit").
			Style(style.FgBold(theme.Error())).
			Build(),
	).Stretch().Build()

	contentArea := buildContentArea(input, items)

	return ui.HStackBuilder(
		ui.Flex(
			stack.NewVStack().
				SingleBorder().
				BorderColor(string(theme.Border())).
				SetChildrenList([]ui.VNode{sidebar}),
			1,
		),
		ui.Flex(
			stack.NewVStack().
				SingleBorder().
				BorderColor(string(theme.Border())).
				SetChildrenList([]ui.VNode{contentArea}),
			1,
		),
	).Gap(0).Build()
}

func buildContentArea(input string, items []string) ui.VNode {
	children := make([]ui.VNode, 0, len(items)+3)

	children = append(children,
		ui.NewTextBuilder("[ Input: "+input+" ]").
			Style(style.Foreground(theme.Text())).
			Build(),
		ui.NewTextBuilder("──────────────────────────────────────").
			Style(style.Foreground(theme.Border())).
			Build(),
	)

	for i := 0; i < len(items)-1; i++ {
		children = append(children,
			ui.NewTextBuilder(items[i]).
				Style(style.Foreground(theme.Muted())).
				Build(),
		)
	}

	// Last item with ellipsis
	children = append(children,
		ui.HStack(
			ui.NewTextBuilder(items[len(items)-1]).
				Style(style.Foreground(theme.Muted())).
				Build(),
			ui.NewTextBuilder(" ...").
				Style(style.FgItalic(theme.Placeholder())).
				Build(),
		),
	)

	return ui.VStackBuilder(children...).Stretch().Build()
}

// BuildDemo1ConfirmModal builds a confirmation modal
func BuildDemo1ConfirmModal(onClose func()) ui.VNode {
	// 使用 ui.On 注册 Intent handler（简单场景）
	ui.On(ModalCloseIntent{}, func() {
		onClose()
	})

	modalBox := stack.NewVStack().
		SingleBorder().
		BorderColor(string(theme.Warning())).
		SetWidth(40).
		SetChildrenList([]ui.VNode{
			ui.VStackBuilder(
				ui.Text(""),
				ui.Text("=== MODAL START ==="),
				ui.HStackBuilder(
					ui.NewTextBuilder("*** Are you sure? ***").
						Style(style.FgBold(theme.Warning())).
						Build(),
				).Align(ui.AlignCenter).Build(),
				ui.Text(""),
				ui.HStackBuilder(
					ui.NewButtonBuilder("[ Cancel ]").
						Variant(ui.ButtonVariantSecondary).
						OnPress(ModalCloseIntent{}).
						FocusStyle(ui.FocusStyleBracket).
						Build(),
					ui.Text(" "),
					ui.NewButtonBuilder("[ OK ]").
						Variant(ui.ButtonVariantSuccess).
						FocusStyle(ui.FocusStyleBracket).
						OnPress(ModalCloseIntent{}).
						Build(),
				).Align(ui.AlignCenter).Build(),
				ui.Text(""),
				ui.HStackBuilder(
					ui.NewTextBuilder("Press ESC to close").
						Style(style.Foreground(theme.Placeholder())).
						Build(),
				).Align(ui.AlignCenter).Build(),
				ui.Text(""),
				ui.Text("=== MODAL END ==="),
			).Build(),
		})

	return ui.Modal(modalBox).
		OnClose(onClose).
		CloseOnESC(true).
		CloseOnBackdropClick(true).
		Build()
}

// BuildDemo1DebugPanel builds the debug info panel
func BuildDemo1DebugPanel() ui.VNode {
	infoLines := []string{
		"┌─ SCREEN/INFO PANEL ─────────────────────────────────────────────┐",
		"│ Buffer Size: 80x24 (configured via ui.WithWidth/Height)        │",
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
		"└────────────────────────────────────────────────────────────────┘",
	}

	children := make([]ui.VNode, len(infoLines))
	for i, line := range infoLines {
		children[i] = ui.Text(line)
	}

	return ui.VStackBuilder(children...).Build()
}

// =============================================================================
// Pre-defined Component Fixtures for Testing
// =============================================================================

// StandardFixtures returns commonly used component test fixtures
func StandardFixtures() []ComponentFixture {
	return []ComponentFixture{
		{
			Name:        "demo1_full_app",
			Description: "Complete Demo1 application with header, body, and debug panel",
			Build: func() ui.VNode {
				return BuildDemo1App()
			},
		},
		{
			Name:        "demo1_header",
			Description: "Header component with title and click counter",
			Build: func() ui.VNode {
				return BuildDemo1Header(0)
			},
		},
		{
			Name:        "demo1_main_body",
			Description: "Main body with sidebar and content area",
			Build: func() ui.VNode {
				return BuildDemo1MainBody(0, "", []string{"Item 1", "Item 2", "Item 3"})
			},
		},
		{
			Name:        "demo1_modal",
			Description: "Confirmation modal with OK/Cancel buttons",
			Build: func() ui.VNode {
				return BuildDemo1ConfirmModal(func() {})
			},
		},
		{
			Name:        "simple_vstack",
			Description: "Simple vertical stack with text items",
			Build: func() ui.VNode {
				return ui.VStackBuilder(
					ui.Text("Line 1"),
					ui.Text("Line 2"),
					ui.Text("Line 3"),
				).Build()
			},
		},
		{
			Name:        "simple_hstack",
			Description: "Simple horizontal stack with text items",
			Build: func() ui.VNode {
				return ui.HStackBuilder(
					ui.Text("A"),
					ui.Text("B"),
					ui.Text("C"),
				).Build()
			},
		},
		{
			Name:        "nested_layout",
			Description: "Nested VStack inside HStack",
			Build: func() ui.VNode {
				return ui.HStackBuilder(
					ui.VStackBuilder(
						ui.Text("Left 1"),
						ui.Text("Left 2"),
					).Build(),
					ui.VStackBuilder(
						ui.Text("Right 1"),
						ui.Text("Right 2"),
					).Build(),
				).Build()
			},
		},
		{
			Name:        "bordered_content",
			Description: "Bordered container with content",
			Build: func() ui.VNode {
				return stack.NewVStack().
					SingleBorder().
					BorderColor("blue").
					SetChildrenList([]ui.VNode{ui.Text("Bordered content")})
			},
		},
		{
			Name:        "flex_layout",
			Description: "HStack with flex items",
			Build: func() ui.VNode {
				return ui.HStackBuilder(
					ui.Flex(ui.Text("Fixed"), 0),
					ui.Flex(ui.Text("Flex-1"), 1),
					ui.Flex(ui.Text("Flex-2"), 2),
				).Build()
			},
		},
		{
			Name:        "keyed_items",
			Description: "VStack with keyed items for reconciliation testing",
			Build: func() ui.VNode {
				return ui.VStackBuilder(
					ui.Element("text").Key("item-a").Prop("content", "A").Build(),
					ui.Element("text").Key("item-b").Prop("content", "B").Build(),
					ui.Element("text").Key("item-c").Prop("content", "C").Build(),
				).Build()
			},
		},
	}
}

// GetFixture returns a fixture by name
func GetFixture(name string) *ComponentFixture {
	for _, f := range StandardFixtures() {
		if f.Name == name {
			return &f
		}
	}
	return nil
}

// =============================================================================
// Test Helper Functions
// =============================================================================

// BuildVNodeTree creates a simple VNode tree for testing
func BuildVNodeTree(depth, breadth int) ui.VNode {
	if depth == 0 {
		return ui.Text("leaf")
	}

	children := make([]ui.VNode, breadth)
	for i := 0; i < breadth; i++ {
		children[i] = BuildVNodeTree(depth-1, breadth)
	}

	return ui.VStackBuilder(children...).Build()
}

// BuildKeyedVNodeTree creates a keyed VNode tree for testing
func BuildKeyedVNodeTree(depth, breadth int, prefix string) ui.VNode {
	if depth == 0 {
		return ui.Element("text").
			Key(prefix).
			Prop("content", prefix).
			Build()
	}

	children := make([]ui.VNode, breadth)
	for i := 0; i < breadth; i++ {
		key := fmt.Sprintf("%s-%d", prefix, i)
		children[i] = BuildKeyedVNodeTree(depth-1, breadth, key)
	}

	return ui.VStackBuilder(children...).Key(prefix).Build()
}

// BuildMixedKeyedTree creates a tree with both keyed and non-keyed nodes
func BuildMixedKeyedTree(keyedCount, nonKeyedCount int) ui.VNode {
	children := make([]ui.VNode, 0, keyedCount+nonKeyedCount)

	for i := 0; i < keyedCount; i++ {
		children = append(children,
			ui.Element("text").
				Key(fmt.Sprintf("keyed-%d", i)).
				Prop("content", fmt.Sprintf("Keyed %d", i)).
				Build(),
		)
	}

	for i := 0; i < nonKeyedCount; i++ {
		children = append(children,
			ui.Element("text").
				Prop("content", fmt.Sprintf("Non-keyed %d", i)).
				Build(),
		)
	}

	return ui.VStackBuilder(children...).Build()
}

// CountNodes counts total nodes in a VNode tree
func CountNodes(vnode ui.VNode) int {
	if vnode == nil {
		return 0
	}
	count := 1
	for _, child := range vnode.Children() {
		count += CountNodes(child)
	}
	return count
}

package container

import (
	"github.com/wwsheng009/mint/framework/theme"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui"
)

// Panel is a high-level container component that manages borders, headers, and content layout.
// It simplifies layout by automatically handling height calculations and flex distribution.
type Panel struct {
	*rtui.BorderedNode

	title       string
	header      rtui.VNode
	footer      rtui.VNode
	content     rtui.VNode
	borderStyle string
	padding     int
	width       int
	height      int
	flex        int
}

// PanelBuilder creates a new Panel
func PanelBuilder() *panelBuilder {
	return &panelBuilder{
		panel: &Panel{
			borderStyle: string(theme.Border()), // Default to theme border
		},
	}
}

type panelBuilder struct {
	panel *Panel
}

// Title sets the panel title (displayed in border if supported, or as a header)
func (b *panelBuilder) Title(title string) *panelBuilder {
	b.panel.title = title
	return b
}

// Header sets a custom header component (appears above content)
func (b *panelBuilder) Header(header rtui.VNode) *panelBuilder {
	b.panel.header = header
	return b
}

// Footer sets a custom footer component (appears below content)
func (b *panelBuilder) Footer(footer rtui.VNode) *panelBuilder {
	b.panel.footer = footer
	return b
}

// Content sets the main content of the panel
func (b *panelBuilder) Content(content rtui.VNode) *panelBuilder {
	b.panel.content = content
	return b
}

// Width sets the explicit width
func (b *panelBuilder) Width(w int) *panelBuilder {
	b.panel.width = w
	return b
}

// Height sets the explicit height
func (b *panelBuilder) Height(h int) *panelBuilder {
	b.panel.height = h
	return b
}

// Flex sets the flex factor
func (b *panelBuilder) Flex(f int) *panelBuilder {
	b.panel.flex = f
	return b
}

// BorderStyle sets the border style
func (b *panelBuilder) BorderStyle(style string) *panelBuilder {
	b.panel.borderStyle = style
	return b
}

// Padding sets inner padding
func (b *panelBuilder) Padding(p int) *panelBuilder {
	b.panel.padding = p
	return b
}

// Style sets the base style
func (b *panelBuilder) Style(s style.Style) *panelBuilder {
	// We'll apply this to the Bordered node later
	return b
}

// Build constructs the Panel VNode
func (b *panelBuilder) Build() rtui.VNode {
	// Construct internal layout
	// Structure:
	// Bordered
	//   └── VStack
	//         ├── Header (Fixed)
	//         ├── Content (Flex)
	//         └── Footer (Fixed)

	var children []rtui.VNode

	// 1. Header
	if b.panel.header != nil {
		children = append(children, b.panel.header)
	} else if b.panel.title != "" {
		// Default header from title
		// TODO: Integrate title into border if Bordered supports it
		// For now, add as text
		children = append(children, ui.Text(b.panel.title))
	}

	// 2. Content
	// Wrap content in a box that handles padding if needed
	// CRITICAL: Content must be Flex=1 to fill remaining space inside the panel
	contentNode := b.panel.content
	if contentNode == nil {
		contentNode = ui.Text("")
	}

	// Ensure content expands using Flex helper
	contentNode = rtui.Flex(contentNode, 1)
	children = append(children, contentNode)

	// 3. Footer
	if b.panel.footer != nil {
		children = append(children, b.panel.footer)
	}

	// Create the main container stack
	container := rtui.VStack(children...)

	// Apply padding if requested (by wrapping in another container with padding props)
	// Note: Mint's VStack handles padding via props, we can set it here
	if b.panel.padding > 0 {
		if setter, ok := container.(interface{ SetProp(string, interface{}) }); ok {
			setter.SetProp("padding", b.panel.padding)
		}
	}

	// Create the Bordered wrapper
	borderBuilder := rtui.Bordered().
		Style(b.panel.borderStyle).
		Child(container)

	if b.panel.width > 0 {
		borderBuilder.Width(b.panel.width)
	}
	if b.panel.height > 0 {
		borderBuilder.Height(b.panel.height)
	}
	if b.panel.flex > 0 {
		// Use manual property setting if Flex method doesn't exist on builder
		// or ensure BorderedBuilder has Flex method
		// Assuming we will fix BorderedBuilder
		borderBuilder.Flex(b.panel.flex)
	}

	// Build the final node
	// Note: We return the BorderedNode directly as it is the root of our component
	return borderBuilder.Build()
}

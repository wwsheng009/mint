package ui

import (
	"fmt"
	"os"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/framework/component"
	frameworkevent "github.com/wwsheng009/mint/framework/event"
	"github.com/wwsheng009/mint/runtime/paint"
)

// Option configures the app
type Option func(*Options)

// Options holds app configuration
type Options struct {
	Width          int
	Height         int
	Title          string
	FPS            int
	EnableDevTools bool
}

// WithWidth sets the window width
func WithWidth(width int) Option {
	return func(o *Options) {
		o.Width = width
	}
}

// WithHeight sets the window height
func WithHeight(height int) Option {
	return func(o *Options) {
		o.Height = height
	}
}

// WithTitle sets the window title
func WithTitle(title string) Option {
	return func(o *Options) {
		o.Title = title
	}
}

// WithFPS sets the frame rate limit
func WithFPS(fps int) Option {
	return func(o *Options) {
		o.FPS = fps
	}
}

// appInstance holds the framework app for quit functionality
var appInstance *framework.App

// Run starts the declarative UI application
func Run(app ComponentFunc, opts ...Option) error {
	options := &Options{
		Width:  80,
		Height: 24,
		Title:  "Mint UI App",
		FPS:    60,
	}

	for _, opt := range opts {
		opt(options)
	}

	// Create the framework app
	fwApp := framework.NewApp()
	fwApp.Resize(options.Width, options.Height)
	appInstance = fwApp

	// Initialize theme
	if err := fwApp.InitTheme("dark"); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize theme: %v\n", err)
	}

	// Create the declarative root component
	declarativeRoot := newDeclarativeRoot(app, fwApp)

	// Set as root
	fwApp.SetRoot(declarativeRoot)

	// Run the app
	return fwApp.Run()
}

// declarativeRoot wraps a declarative component to work with the framework
type declarativeRoot struct {
	component.Node
	appFn          ComponentFunc
	ctx            *ComponentContext
	app            *framework.App
	buttons        []*ButtonVNode // Collected buttons for focus management
	focusedIndex   int            // Currently focused button index (-1 = none)
	renderBtnIndex int            // Current button index during rendering
}

// newDeclarativeRoot creates a new declarative root component
func newDeclarativeRoot(fn ComponentFunc, app *framework.App) component.Node {
	return &declarativeRoot{
		appFn:        fn,
		ctx:          newComponentContext("App"),
		app:          app,
		buttons:      make([]*ButtonVNode, 0),
		focusedIndex: -1,
	}
}

// Paint implements component.Paintable
func (d *declarativeRoot) Paint(ctx component.PaintContext, buffer *paint.Buffer) {
	// Reset hook index for re-render
	d.ctx.resetContext()

	// Set current context
	setCurrentContext(d.ctx)

	// Call the root component to get VNode
	vnode := d.appFn()

	// Clear current context
	setCurrentContext(nil)

	// Validate hooks finished correctly
	if err := d.ctx.finishRender(); err != nil {
		fmt.Fprintf(os.Stderr, "Hook validation error: %v\n", err)
		return
	}

	// Clear and collect buttons for focus management
	d.buttons = d.buttons[:0]
	d.collectButtons(vnode)

	// Auto-focus first button if none focused
	if d.focusedIndex < 0 && len(d.buttons) > 0 {
		d.focusedIndex = 0
	}

	// Reset render button index for tracking during render
	d.renderBtnIndex = 0

	// Render the VNode tree to the buffer
	d.renderVNode(vnode, ctx.X, ctx.Y, buffer)
}

// renderVNode recursively renders a VNode to the buffer
func (d *declarativeRoot) renderVNode(node VNode, x, y int, buffer *paint.Buffer) int {
	if node == nil {
		return 0
	}

	currentY := y

	switch n := node.(type) {
	case *TextVNode:
		d.renderText(n, x, currentY, buffer)
		currentY += 1

	case *ElementVNode:
		for _, child := range n.Children() {
			offsetY := d.renderVNode(child, x, currentY, buffer)
			currentY += offsetY
		}

	case *LayoutNode:
		// Apply padding
		padding := n.Padding()
		currentX := x + padding[3]
		currentY = y + padding[0]
		gap := n.Gap()

		for i, child := range n.Children() {
			offsetY := d.renderVNode(child, currentX, currentY, buffer)
			if n.direction == DirectionRow {
				width := d.measureWidth(child)
				currentX += width + gap
			} else {
				currentY += offsetY
				if i < len(n.Children())-1 {
					currentY += gap
				}
			}
		}

	case *ButtonVNode:
		d.renderButton(n, x, currentY, buffer)
		currentY += 1

	case *FragmentVNode:
		for _, child := range n.Children() {
			offsetY := d.renderVNode(child, x, currentY, buffer)
			currentY += offsetY
		}

	case *ComponentVNode:
		// Render the component by calling its function
		rendered := n.Render()
		if rendered != nil {
			offsetY := d.renderVNode(rendered, x, currentY, buffer)
			currentY += offsetY
		}
	}

	return currentY - y
}

// renderText renders a text node
func (d *declarativeRoot) renderText(node *TextVNode, x, y int, buffer *paint.Buffer) {
	// Check bounds
	if y < 0 || y >= buffer.Height {
		return
	}

	content := node.Content()
	nodeStyle := node.Style()

	// Use SetString for simpler text rendering
	buffer.SetString(x, y, content, nodeStyle)
}

// renderButton renders a button node
func (d *declarativeRoot) renderButton(node *ButtonVNode, x, y int, buffer *paint.Buffer) {
	// Check bounds
	if y < 0 || y >= buffer.Height {
		return
	}

	buttonStyle := node.Style()
	if buttonStyle.FG == "" {
		buttonStyle = buttonStyle.Foreground("white")
	}

	// Highlight focused button using index comparison
	if d.renderBtnIndex == d.focusedIndex {
		buttonStyle = buttonStyle.Background("blue").Foreground("white").Bold(true)
	}

	// Increment render button index for next button
	d.renderBtnIndex++

	label := node.Label()
	if label == "" {
		label = " "
	}

	// Draw [ label ]
	fullLabel := "[" + label + "]"
	buffer.SetString(x, y, fullLabel, buttonStyle)
}

// measureWidth measures the width of a node
func (d *declarativeRoot) measureWidth(node VNode) int {
	switch n := node.(type) {
	case *TextVNode:
		return len(n.Content())
	case *ButtonVNode:
		return len(n.Label()) + 2 // +2 for brackets
	default:
		return 10 // Default width
	}
}

// collectButtons recursively collects all buttons from the VNode tree
func (d *declarativeRoot) collectButtons(node VNode) {
	d.collectButtonsTo(node, &d.buttons)
}

// collectButtonsTo recursively collects all buttons into the provided slice
func (d *declarativeRoot) collectButtonsTo(node VNode, buttons *[]*ButtonVNode) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *ButtonVNode:
		if !n.Disabled() {
			*buttons = append(*buttons, n)
		}
	case *ElementVNode:
		for _, child := range n.Children() {
			d.collectButtonsTo(child, buttons)
		}
	case *LayoutNode:
		for _, child := range n.Children() {
			d.collectButtonsTo(child, buttons)
		}
	case *FragmentVNode:
		for _, child := range n.Children() {
			d.collectButtonsTo(child, buttons)
		}
	case *ComponentVNode:
		rendered := n.Render()
		if rendered != nil {
			d.collectButtonsTo(rendered, buttons)
		}
	}
}

// isFocused checks if a button is currently focused
func (d *declarativeRoot) isFocused(btn *ButtonVNode) bool {
	if d.focusedIndex < 0 || d.focusedIndex >= len(d.buttons) {
		return false
	}
	return d.buttons[d.focusedIndex] == btn
}

// HandleEvent implements frameworkevent.Component interface
func (d *declarativeRoot) HandleEvent(ev frameworkevent.Event) bool {
	// Handle keyboard events
	if keyEv, ok := ev.(*frameworkevent.KeyEvent); ok {
		// Check for quit keys: 'q', 'Q'
		if keyEv.Key.Rune == 'q' || keyEv.Key.Rune == 'Q' {
			if d.app != nil {
				d.app.Quit()
			}
			return true
		}

		// Check for Escape key
		if keyEv.Special == frameworkevent.KeyEscape {
			if d.app != nil {
				d.app.Quit()
			}
			return true
		}

		// Check for Ctrl+C
		if keyEv.Key.Rune == 'c' && keyEv.Key.Ctrl {
			if d.app != nil {
				d.app.Quit()
			}
			return true
		}

		// Tab: focus next button
		if keyEv.Special == frameworkevent.KeyTab {
			if len(d.buttons) > 0 {
				if keyEv.Modifiers == frameworkevent.ModShift {
					// Shift+Tab: focus previous
					d.focusedIndex--
					if d.focusedIndex < 0 {
						d.focusedIndex = len(d.buttons) - 1
					}
				} else {
					// Tab: focus next
					d.focusedIndex++
					if d.focusedIndex >= len(d.buttons) {
						d.focusedIndex = 0
					}
				}
				if d.app != nil {
					d.app.MarkDirty()
				}
			}
			return true
		}

		// Enter/Space: trigger focused button click
		if keyEv.Special == frameworkevent.KeyEnter || keyEv.Key.Rune == ' ' {
			if d.focusedIndex >= 0 && d.focusedIndex < len(d.buttons) {
				btn := d.buttons[d.focusedIndex]
				if onClick := btn.OnClick(); onClick != nil {
					onClick()
				}
			}
			return true
		}

		// Left/Right arrow keys for horizontal button navigation
		if keyEv.Special == frameworkevent.KeyLeft {
			if len(d.buttons) > 0 {
				d.focusedIndex--
				if d.focusedIndex < 0 {
					d.focusedIndex = len(d.buttons) - 1
				}
				if d.app != nil {
					d.app.MarkDirty()
				}
			}
			return true
		}

		if keyEv.Special == frameworkevent.KeyRight {
			if len(d.buttons) > 0 {
				d.focusedIndex++
				if d.focusedIndex >= len(d.buttons) {
					d.focusedIndex = 0
				}
				if d.app != nil {
					d.app.MarkDirty()
				}
			}
			return true
		}
	}

	// Mark for re-render on any event
	return true
}

package ui

import (
	"fmt"
	"os"
	"time"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/framework/component"
	frameworkevent "github.com/wwsheng009/mint/framework/event"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/scheduler"
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
		if os.Getenv("TUI_DEBUG_UI") == "true" {
			fmt.Fprintf(os.Stderr, "Failed to initialize theme: %v\n", err)
		}
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
	appFn               ComponentFunc
	ctx                 *ComponentContext
	app                 *framework.App
	instanceManager     *InstanceManager  // Manages persistent component instances
	reconciler          *Reconciler       // Fiber reconciler (env controlled)
	activeComponentKeys []string          // Track active component keys for cleanup
	buttons             []*ButtonVNode   // Collected buttons for focus management
	inputs              []*InputVNode    // Collected inputs for focus management
	textareas           []*TextareaVNode // Collected textareas for focus management
	checkboxes          []*CheckboxVNode // Collected checkboxes for focus management
	selects             []*SelectVNode   // Collected selects for focus management
	focusedIndex        int              // Currently focused element index (-1 = none)
	focusedType         int              // Type: 0=button, 1=input, 2=textarea, 3=checkbox, 4=select

	// Input scheduling - priority-based event processing
	inputQueue      *scheduler.InputQueue
	mouseMoveHandler *scheduler.MouseMoveHandler
	mouseClickHandler *scheduler.MouseClickHandler
	mouseTracker     *scheduler.MouseTracker

	// Interaction state manager - persists hover/focus state across renders
	// Uses InstanceManager with auto-generated keys based on component position
	interactionState map[string]bool // key -> state (hovered, focused, etc.)
}

// newDeclarativeRoot creates a new declarative root component
func newDeclarativeRoot(fn ComponentFunc, app *framework.App) component.Node {
	// Initialize input queue
	inputQueue := scheduler.NewInputQueue()

	// Initialize mouse tracker
	mouseTracker := scheduler.NewMouseTracker()

	// Create mouse handlers with default throttle config
	config := scheduler.DefaultThrottleConfig()

	mouseMoveHandler := scheduler.NewMouseMoveHandler(
		scheduler.MouseHandlerFunc(func(event *scheduler.MouseEvent) {
			// Mouse move events are handled immediately through handleMouseEvent
			if app != nil {
				app.MarkDirty()
			}
		}),
		config,
	)

	mouseClickHandler := scheduler.NewMouseClickHandler(
		scheduler.MouseHandlerFunc(func(event *scheduler.MouseEvent) {
			// Click events are handled immediately through handleMouseEvent
			if app != nil {
				app.MarkDirty()
			}
		}),
		config,
	)

	// Initialize Fiber reconciler if enabled via environment variable
	var reconciler *Reconciler
	enableFiber := os.Getenv("MINT_USE_FIBER") == "true"
	if enableFiber {
		reconciler = NewReconciler(app, fn, ReconcilerConfig{
			TimeBudget:      5 * time.Millisecond,
			EnableProfiling: false,
			EnableFiber:     true,
		})
		// Set render callback for VNode rendering
		reconciler.SetRenderCallback(func(vnode VNode, x, y int, buffer *paint.Buffer) {
			// This will be called during commit phase to render each VNode
			// For now, we'll use the existing renderVNode method
		})
	}

	return &declarativeRoot{
		appFn:               fn,
		ctx:                 newComponentContext("App"),
		app:                 app,
		instanceManager:     NewInstanceManager(),
		reconciler:          reconciler,
		activeComponentKeys: make([]string, 0),
		buttons:             make([]*ButtonVNode, 0),
		inputs:              make([]*InputVNode, 0),
		textareas:           make([]*TextareaVNode, 0),
		checkboxes:          make([]*CheckboxVNode, 0),
		selects:             make([]*SelectVNode, 0),
		focusedIndex:        -1,
		focusedType:         0,
		inputQueue:          inputQueue,
		mouseMoveHandler:    mouseMoveHandler,
		mouseClickHandler:   mouseClickHandler,
		mouseTracker:        mouseTracker,
	}
}

// Paint implements component.Paintable
func (d *declarativeRoot) Paint(ctx component.PaintContext, buffer *paint.Buffer) {
	// Use Fiber reconciler if enabled
	if d.reconciler != nil {
		d.paintWithFiber(ctx, buffer)
		return
	}

	// Legacy direct rendering path
	d.paintLegacy(ctx, buffer)
}

// paintWithFiber renders using the Fiber reconciler
// In Fiber mode:
// - Component functions are called by beginWorkComponent (not by renderVNode)
// - Fiber tree is built and traversed by reconciler
// - renderCallback only renders leaf nodes (no recursion)
func (d *declarativeRoot) paintWithFiber(ctx component.PaintContext, buffer *paint.Buffer) {
	// Clear active component keys for this render
	d.activeComponentKeys = d.activeComponentKeys[:0]

	// Save hover state before render
	hoveredButtons := d.saveHoveredButtons()
	hoveredInputs := d.saveHoveredInputs()
	hoveredCheckboxes := d.saveHoveredCheckboxes()
	hoveredSelects := d.saveHoveredSelects()
	hoveredTextareas := d.saveHoveredTextareas()

	// Reset interactive elements collection
	d.resetInteractiveElements()

	if os.Getenv("TUI_DEBUG_UI") == "true" {
		fmt.Fprintf(os.Stderr, "[paintWithFiber] Reset interactive elements, starting Fiber render\n")
	}

	// Share instance manager with reconciler for consistent state
	d.reconciler.SetInstanceManager(d.instanceManager)

	// Set render callback for Fiber mode
	// This callback is called for each Fiber node during commit phase
	// It should ONLY render the current node, NOT recurse into children
	// (Fiber tree traversal handles children)
	d.reconciler.SetRenderCallback(d.renderVNodeFiber)

	// Render using Fiber reconciler
	// The reconciler will:
	// 1. Call d.appFn() to get root VNode
	// 2. Build Fiber tree (beginWorkComponent expands components)
	// 3. Traverse Fiber tree and call renderCallback for each node
	d.reconciler.Render(ctx, buffer, d.appFn)

	// Restore hover state after render
	d.restoreHoveredButtons(hoveredButtons)
	d.restoreHoveredInputs(hoveredInputs)
	d.restoreHoveredCheckboxes(hoveredCheckboxes)
	d.restoreHoveredSelects(hoveredSelects)
	d.restoreHoveredTextareas(hoveredTextareas)

	// Validate and clamp focusedIndex to valid range
	totalElements := d.getTotalFocusableCount()
	if os.Getenv("TUI_DEBUG_UI") == "true" {
		fmt.Fprintf(os.Stderr, "After render: buttons=%d, totalElements=%d, focusedIndex=%d\n",
			len(d.buttons), totalElements, d.focusedIndex)
	}
	if totalElements == 0 {
		d.focusedIndex = -1
		d.focusedType = -1
	} else if d.focusedIndex < 0 || d.focusedIndex >= totalElements {
		d.focusedIndex = 0
		d.focusedType = d.getFirstElementType()
	} else {
		_, d.focusedType = d.getElementByIndex(d.focusedIndex)
	}
}

// paintLegacy renders using direct VNode traversal (original implementation)
func (d *declarativeRoot) paintLegacy(ctx component.PaintContext, buffer *paint.Buffer) {
	// Reset hook index for re-render
	d.ctx.resetContext()

	// Set current context
	setCurrentContext(d.ctx)

	// Clear active component keys for this render
	d.activeComponentKeys = d.activeComponentKeys[:0]

	// Call the root component to get VNode
	vnode := d.appFn()

	// Clear current context
	setCurrentContext(nil)

	// Validate hooks finished correctly
	if err := d.ctx.finishRender(); err != nil {
		if os.Getenv("TUI_DEBUG_UI") == "true" {
			fmt.Fprintf(os.Stderr, "Hook validation error: %v\n", err)
		}
		return
	}

	// Run effects after render completes
	d.ctx.runEffects()

	// Note: Component instance cleanup is moved to after renderVNode
	// because activeComponentKeys is populated during render, not during appFn()

	// Clear and collect interactive elements for focus management
	// 保存 hover 状态，使用组件在列表中的索引作为key
	// 索引在重新渲染前后是稳定的（组件在树中的位置不变）
	hoveredButtons := make(map[int]bool)
	hoveredInputs := make(map[int]bool)
	hoveredCheckboxes := make(map[int]bool)
	hoveredSelects := make(map[int]bool)
	hoveredTextareas := make(map[int]bool)

	for i, btn := range d.buttons {
		if btn.IsHovered() {
			hoveredButtons[i] = true
		}
	}
	for i, inp := range d.inputs {
		if inp.IsHovered() {
			hoveredInputs[i] = true
		}
	}
	for i, chk := range d.checkboxes {
		if chk.IsHovered() {
			hoveredCheckboxes[i] = true
		}
	}
	for i, sel := range d.selects {
		if sel.IsHovered() {
			hoveredSelects[i] = true
		}
	}
	for i, ta := range d.textareas {
		if ta.IsHovered() {
			hoveredTextareas[i] = true
		}
	}

	// Reset interactive elements collection
	// They will be collected during renderVNode to avoid double-rendering components
	d.resetInteractiveElements()

	// Validate and clamp focusedIndex to valid range
	// Note: totalElements will be computed during renderVNode
	// We'll validate after render completes

	// 渲染 VNode 树到缓冲区（此过程会设置 bounds）
	// Interactive elements are collected during this render pass
	d.renderVNode(vnode, ctx.X, ctx.Y, buffer)

	// Cleanup unused component instances after render completes
	// activeComponentKeys has been populated by renderVNode's component handling
	d.instanceManager.Cleanup(d.activeComponentKeys)

	// Now validate focusedIndex after elements have been collected
	totalElements := d.getTotalFocusableCount()
	if totalElements == 0 {
		d.focusedIndex = -1
		d.focusedType = -1
	} else if d.focusedIndex < 0 || d.focusedIndex >= totalElements {
		d.focusedIndex = 0
		d.focusedType = d.getFirstElementType()
	} else {
		// Update focusedType in case the element type at focusedIndex changed
		_, d.focusedType = d.getElementByIndex(d.focusedIndex)
	}

	// 渲染完成后恢复 hover 状态（使用索引匹配）
	for i, btn := range d.buttons {
		if hoveredButtons[i] {
			btn.SetHovered(true)
		}
	}
	for i, inp := range d.inputs {
		if hoveredInputs[i] {
			inp.SetHovered(true)
		}
	}
	for i, chk := range d.checkboxes {
		if hoveredCheckboxes[i] {
			chk.SetHovered(true)
		}
	}
	for i, sel := range d.selects {
		if hoveredSelects[i] {
			sel.SetHovered(true)
		}
	}
	for i, ta := range d.textareas {
		if hoveredTextareas[i] {
			ta.SetHovered(true)
		}
	}
}

// Helper functions to save/restore hover state
func (d *declarativeRoot) saveHoveredButtons() map[int]bool {
	result := make(map[int]bool)
	for i, btn := range d.buttons {
		if btn.IsHovered() {
			result[i] = true
		}
	}
	return result
}

func (d *declarativeRoot) saveHoveredInputs() map[int]bool {
	result := make(map[int]bool)
	for i, inp := range d.inputs {
		if inp.IsHovered() {
			result[i] = true
		}
	}
	return result
}

func (d *declarativeRoot) saveHoveredCheckboxes() map[int]bool {
	result := make(map[int]bool)
	for i, chk := range d.checkboxes {
		if chk.IsHovered() {
			result[i] = true
		}
	}
	return result
}

func (d *declarativeRoot) saveHoveredSelects() map[int]bool {
	result := make(map[int]bool)
	for i, sel := range d.selects {
		if sel.IsHovered() {
			result[i] = true
		}
	}
	return result
}

func (d *declarativeRoot) saveHoveredTextareas() map[int]bool {
	result := make(map[int]bool)
	for i, ta := range d.textareas {
		if ta.IsHovered() {
			result[i] = true
		}
	}
	return result
}

func (d *declarativeRoot) restoreHoveredButtons(hovered map[int]bool) {
	for i, btn := range d.buttons {
		if hovered[i] {
			btn.SetHovered(true)
		}
	}
}

func (d *declarativeRoot) restoreHoveredInputs(hovered map[int]bool) {
	for i, inp := range d.inputs {
		if hovered[i] {
			inp.SetHovered(true)
		}
	}
}

func (d *declarativeRoot) restoreHoveredCheckboxes(hovered map[int]bool) {
	for i, chk := range d.checkboxes {
		if hovered[i] {
			chk.SetHovered(true)
		}
	}
}

func (d *declarativeRoot) restoreHoveredSelects(hovered map[int]bool) {
	for i, sel := range d.selects {
		if hovered[i] {
			sel.SetHovered(true)
		}
	}
}

func (d *declarativeRoot) restoreHoveredTextareas(hovered map[int]bool) {
	for i, ta := range d.textareas {
		if hovered[i] {
			ta.SetHovered(true)
		}
	}
}

// renderVNodeFiber renders a SINGLE VNode without recursing into children
// Used in Fiber mode where the Fiber tree traversal handles children
// This is called by renderCallback during the commit phase
func (d *declarativeRoot) renderVNodeFiber(node VNode, x, y int, buffer *paint.Buffer) {
	if node == nil {
		return
	}

	if os.Getenv("TUI_DEBUG_UI") == "true" {
		fmt.Fprintf(os.Stderr, "renderVNodeFiber: type=%T, node=%+v\n", node, node)
	}

	switch n := node.(type) {
	case *TextVNode:
		d.renderText(n, x, y, buffer)

	case *ButtonVNode:
		// Collect button for focus management
		if !n.Disabled() {
			n.focusIndex = len(d.buttons)
			d.buttons = append(d.buttons, n)
			if os.Getenv("TUI_DEBUG_UI") == "true" {
				fmt.Fprintf(os.Stderr, "[renderVNodeFiber] Collected button %d: label=%s\n",
					n.focusIndex, n.Label())
			}
		}
		d.renderButton(n, x, y, buffer)

	case *InputVNode:
		// Collect input for focus management
		if !n.Disabled() && !n.ReadOnly() {
			n.focusIndex = len(d.inputs)
			d.inputs = append(d.inputs, n)
		}
		d.renderInput(n, x, y, buffer)

	case *TextareaVNode:
		// Collect textarea for focus management
		if !n.Disabled() {
			n.focusIndex = len(d.textareas)
			d.textareas = append(d.textareas, n)
		}
		d.renderTextarea(n, x, y, buffer)

	case *CheckboxVNode:
		// Collect checkbox for focus management
		if !n.Disabled() {
			n.focusIndex = len(d.checkboxes)
			d.checkboxes = append(d.checkboxes, n)
		}
		d.renderCheckbox(n, x, y, buffer)

	case *SelectVNode:
		// Collect select for focus management
		if !n.Disabled() {
			n.focusIndex = len(d.selects)
			d.selects = append(d.selects, n)
		}
		d.renderSelect(n, x, y, buffer)

	case *ProgressVNode:
		d.renderProgress(n, x, y, buffer)

	case *SpinnerVNode:
		d.renderSpinner(n, x, y, buffer)

	case *TableVNode:
		d.renderTable(n, x, y, buffer)

	// Container nodes and ComponentVNode are handled by Fiber tree traversal
	// We don't need to render them here as their children will be rendered separately
	case *ElementVNode, *LayoutNode, *GridVNode, *FragmentVNode, *ComponentVNode, *AbsoluteVNode, *VirtualListVNode:
		// Skip - children are rendered by Fiber tree traversal
	}
}

// renderVNode recursively renders a VNode to the buffer (legacy mode)
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

		// Track the maximum height for DirectionRow
		maxHeight := 0
		for i, child := range n.Children() {
			offsetY := d.renderVNode(child, currentX, currentY, buffer)
			if n.direction == DirectionRow {
				// Track maximum height for row layout
				if offsetY > maxHeight {
					maxHeight = offsetY
				}
				width := d.measureWidth(child)
				currentX += width + gap
			} else {
				currentY += offsetY
				if i < len(n.Children())-1 {
					currentY += gap
				}
			}
		}

		// For DirectionRow, add the maximum child height to currentY
		if n.direction == DirectionRow {
			currentY += maxHeight
		}

	case *GridVNode:
		// Apply padding
		padding := n.Padding()
		gap := n.Gap()
		currentX := x + padding[3]
		currentY = y + padding[0]

		// Get available width/height
		availableWidth := buffer.Width - padding[1] - padding[3]
		availableHeight := buffer.Height - padding[0] - padding[2]

		// Calculate column widths and row heights
		colWidths := n.CalculateColumnWidths(availableWidth)
		rowHeights := n.CalculateRowHeights(availableHeight)

		// Track occupied cells for spanning
		occupied := make(map[[2]int]bool)

		// Render each cell
		for _, cell := range n.Cells() {
			// Check if this position is already occupied
			key := [2]int{cell.Row, cell.Col}
			if occupied[key] {
				continue
			}

			// Mark this cell and any spanned cells as occupied
			for r := cell.Row; r < cell.Row+cell.RowSpan && r < len(rowHeights); r++ {
				for c := cell.Col; c < cell.Col+cell.ColSpan && c < len(colWidths); c++ {
					occupied[[2]int{r, c}] = true
				}
			}

			// Calculate cell position and size
			cellX := currentX
			for c := 0; c < cell.Col && c < len(colWidths); c++ {
				cellX += colWidths[c]
				if c < len(colWidths)-1 {
					cellX += gap[0]
				}
			}

			cellY := currentY
			for r := 0; r < cell.Row && r < len(rowHeights); r++ {
				cellY += rowHeights[r]
				if r < len(rowHeights)-1 {
					cellY += gap[1]
				}
			}

			// Calculate cell width and height (considering spans)
			cellWidth := 0
			for c := cell.Col; c < cell.Col+cell.ColSpan && c < len(colWidths); c++ {
				cellWidth += colWidths[c]
				if c < cell.Col+cell.ColSpan-1 && c < len(colWidths)-1 {
					cellWidth += gap[0]
				}
			}

			cellHeight := 0
			for r := cell.Row; r < cell.Row+cell.RowSpan && r < len(rowHeights); r++ {
				cellHeight += rowHeights[r]
				if r < cell.Row+cell.RowSpan-1 && r < len(rowHeights)-1 {
					cellHeight += gap[1]
				}
			}

			// Render the child (clip to cell bounds if needed)
			d.renderVNode(cell.Child, cellX, cellY, buffer)
		}

		// Calculate total height
		for _, h := range rowHeights {
			currentY += h
		}
		currentY += padding[2]

	case *ButtonVNode:
		// Collect button for focus management
		if !n.Disabled() {
			n.focusIndex = len(d.buttons)
			d.buttons = append(d.buttons, n)
		}
		d.renderButton(n, x, currentY, buffer)
		currentY += 1

	case *InputVNode:
		// Collect input for focus management
		if !n.Disabled() && !n.ReadOnly() {
			n.focusIndex = len(d.inputs)
			d.inputs = append(d.inputs, n)
		}
		d.renderInput(n, x, currentY, buffer)
		currentY += 1

	case *TextareaVNode:
		// Collect textarea for focus management
		if !n.Disabled() {
			n.focusIndex = len(d.textareas)
			d.textareas = append(d.textareas, n)
		}
		// Textarea can be multiple lines
		height := n.Rows()
		d.renderTextarea(n, x, currentY, buffer)
		currentY += height

	case *CheckboxVNode:
		// Collect checkbox for focus management
		if !n.Disabled() {
			n.focusIndex = len(d.checkboxes)
			d.checkboxes = append(d.checkboxes, n)
		}
		d.renderCheckbox(n, x, currentY, buffer)
		currentY += 1

	case *ProgressVNode:
		d.renderProgress(n, x, currentY, buffer)
		currentY += 1

	case *SpinnerVNode:
		d.renderSpinner(n, x, currentY, buffer)
		currentY += 1

	case *SelectVNode:
		// Collect select for focus management
		if !n.Disabled() {
			n.focusIndex = len(d.selects)
			d.selects = append(d.selects, n)
		}
		d.renderSelect(n, x, currentY, buffer)
		currentY += 1

	case *TableVNode:
		height := d.renderTable(n, x, currentY, buffer)
		currentY += height

	case *AbsoluteVNode:
		// Calculate position relative to container
		absX, absY := n.CalculatePosition(buffer.Width, buffer.Height)
		// Render child at absolute position
		d.renderVNode(n.Child(), x+absX, y+absY, buffer)
		// Absolute doesn't affect flow

	case *VirtualListVNode:
		// Virtual list rendering
		if n.RenderItem() != nil && len(n.Items()) > 0 {
			start, end := n.GetVisibleRange()
			renderY := y

			// Render visible items
			for i := start; i < end; i++ {
				if i >= 0 && i < len(n.Items()) {
					item := n.Items()[i]
					itemVNode := n.RenderItem()(item)
					offsetY := d.renderVNode(itemVNode, x, renderY, buffer)
					renderY += offsetY
				}
			}
			currentY = y + n.ListHeight()
		}

	case *TooltipVNode:
		// Render the content, and optionally the tooltip text
		d.renderVNode(n.Content(), x, y, buffer)
		currentY += 1
		// Tooltip could be rendered as an overlay if visible

	case *ToastVNode:
		// Render toast if visible
		if n.IsVisible() {
			// Choose color based on type
			color := "blue"
			switch n.ToastType() {
			case ToastSuccess:
				color = "green"
			case ToastWarning:
				color = "yellow"
			case ToastError:
				color = "red"
			}
			// Render toast message
			toastText := NewTextBuilder(n.Message()).FgColor(color).Build()
			if tv, ok := toastText.(*TextVNode); ok {
				d.renderText(tv, x, y, buffer)
			}
			currentY += 1
		}

	case *FragmentVNode:
		for _, child := range n.Children() {
			offsetY := d.renderVNode(child, x, currentY, buffer)
			currentY += offsetY
		}

	case *ComponentVNode:
		// Use instance manager for persistent component state
		componentKey := n.Key()
		if componentKey == "" {
			// Generate a stable key from component name if none provided
			componentKey = "component:" + n.Name()
		}

		// Track this component as active
		d.activeComponentKeys = append(d.activeComponentKeys, componentKey)

		// Get or create component instance
		instance := d.instanceManager.GetOrCreate(componentKey, func() ComponentInstance {
			if n.fnWithProps != nil {
				return NewBaseComponentInstanceWithProps(componentKey, n.fnWithProps, n.Props())
			}
			return NewBaseComponentInstance(componentKey, n.fn)
		})

		// Update props if they changed
		if n.Props() != nil {
			instance.SetProps(n.Props())
		}

		// Render using the persistent instance
		// This ensures hooks state is preserved across renders
		oldContext := getCurrentContext()
		setCurrentContext(instance.GetContext())

		rendered := instance.Render()

		setCurrentContext(oldContext)

		// Run effects for this component
		if rendered != nil {
			instance.GetContext().runEffects()

			// Render the component's output
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

	// Highlight focused button using focusIndex
	isFocused := d.focusedType == 0 && node.focusIndex == d.focusedIndex
	if isFocused {
		buttonStyle = buttonStyle.Background("blue").Foreground("white").Bold(true)
	}

	// Highlight hovered button
	if node.IsHovered() && !isFocused {
		buttonStyle = buttonStyle.Background("cyan").Foreground("black").Bold(true)
	}

	label := node.Label()
	if label == "" {
		label = " "
	}

	// Draw [ label ]
	fullLabel := "[" + label + "]"

	// // DEBUG: 检查最终样式
	// fmt.Fprintf(os.Stderr, "STYLE: label=%s, focused=%v, BG=%q, FG=%q, Bold=%v\n",
	// 	label, isFocused, buttonStyle.BG, buttonStyle.FG, buttonStyle.IsBold())

	buffer.SetString(x, y, fullLabel, buttonStyle)

	// Track bounds for mouse hit testing
	width := len(fullLabel)
	node.SetBounds(x, y, width, 1)
}

// renderInput renders an input node
func (d *declarativeRoot) renderInput(node *InputVNode, x, y int, buffer *paint.Buffer) {
	// Check bounds
	if y < 0 || y >= buffer.Height {
		return
	}

	// Update focus state by using focusIndex instead of pointer comparison
	isFocused := d.focusedType == 1 && node.focusIndex >= 0 && node.focusIndex == d.focusedIndex
	node.SetFocus(isFocused)

	inputStyle := node.Style()
	if inputStyle.FG == "" {
		inputStyle = inputStyle.Foreground("white")
	}

	// Highlight focused input
	if node.IsFocused() {
		inputStyle = inputStyle.Background("blue").Foreground("white").Bold(true)
	}

	// Highlight hovered input
	if node.IsHovered() && !isFocused {
		inputStyle = inputStyle.Background("cyan").Foreground("black").Bold(true)
	}

	// Get display value (or placeholder if empty)
	displayValue := node.Value()
	if displayValue == "" {
		displayValue = node.Placeholder()
		if displayValue == "" {
			displayValue = " "
		}
	}

	// Mask password input
	if node.InputType() == InputTypePassword && len(node.Value()) > 0 {
		displayValue = ""
		for range node.Value() {
			displayValue += "*"
		}
	}

	// Draw input with brackets: [ value     ]
	// Reserve space for cursor if focused
	maxWidth := 20 // Default input width
	if node.MaxLength() > 0 && node.MaxLength() < 20 {
		maxWidth = node.MaxLength()
	}

	// Truncate if too long
	if len(displayValue) > maxWidth {
		displayValue = displayValue[:maxWidth]
	}

	padding := maxWidth - len(displayValue)
	if padding < 0 {
		padding = 0
	}

	// Create padding string
	paddingStr := ""
	for i := 0; i < padding; i++ {
		paddingStr += " "
	}

	fullInput := "[" + displayValue + paddingStr + "]"
	buffer.SetString(x, y, fullInput, inputStyle)

	// Track bounds for mouse hit testing
	width := len(fullInput)
	node.SetBounds(x, y, width, 1)
}

// renderTextarea renders a textarea node
func (d *declarativeRoot) renderTextarea(node *TextareaVNode, x, y int, buffer *paint.Buffer) {
	// Update focus state by using focusIndex instead of pointer comparison
	isFocused := d.focusedType == 2 && node.focusIndex >= 0 && node.focusIndex == d.focusedIndex
	node.SetFocus(isFocused)

	rows := node.Rows()
	if rows < 1 {
		rows = 1
	}

	cols := node.Cols()
	if cols < 1 {
		cols = 20
	}

	for row := 0; row < rows; row++ {
		renderY := y + row
		if renderY >= buffer.Height {
			break
		}

		textareaStyle := node.Style()
		if textareaStyle.FG == "" {
			textareaStyle = textareaStyle.Foreground("white")
		}

		// Highlight focused textarea
		if node.IsFocused() {
			textareaStyle = textareaStyle.Background("blue").Foreground("white").Bold(true)
		}

		// Highlight hovered textarea (only on first row)
		if node.IsHovered() && row == 0 && !isFocused {
			textareaStyle = textareaStyle.Background("cyan").Foreground("black").Bold(true)
		}

		// Simple multi-line rendering
		value := node.Value()
		if value == "" {
			value = node.Placeholder()
			if value == "" {
				value = " "
			}
		}

		// Truncate if too long
		if len(value) > cols {
			value = value[:cols]
		}

		// Pad if too short
		for len(value) < cols {
			value += " "
		}

		// Draw row
		buffer.SetString(x, renderY, "["+value+"]", textareaStyle)
	}

	// Track bounds for mouse hit testing (full textarea area)
	width := cols + 2 // +2 for brackets
	node.SetBounds(x, y, width, rows)
}

// renderCheckbox renders a checkbox node
func (d *declarativeRoot) renderCheckbox(node *CheckboxVNode, x, y int, buffer *paint.Buffer) {
	// Check bounds
	if y < 0 || y >= buffer.Height {
		return
	}

	// Update focus state by using focusIndex instead of pointer comparison
	isFocused := d.focusedType == 3 && node.focusIndex >= 0 && node.focusIndex == d.focusedIndex
	node.SetFocus(isFocused)

	checkboxStyle := node.Style()
	if checkboxStyle.FG == "" {
		checkboxStyle = checkboxStyle.Foreground("white")
	}

	// Highlight focused checkbox
	if node.IsFocused() {
		checkboxStyle = checkboxStyle.Background("blue").Foreground("white").Bold(true)
	}

	// Highlight hovered checkbox
	if node.IsHovered() && !isFocused {
		checkboxStyle = checkboxStyle.Background("cyan").Foreground("black").Bold(true)
	}

	// Checkbox character: [X] for checked, [ ] for unchecked
	boxChar := " "
	if node.Checked() {
		boxChar = "X"
	}

	// Draw [X] or [ ] with label
	display := "[" + boxChar + "]"
	if node.Label() != "" {
		display = display + " " + node.Label()
	}

	buffer.SetString(x, y, display, checkboxStyle)

	// Track bounds for mouse hit testing
	width := len(display)
	node.SetBounds(x, y, width, 1)
}

// renderProgress renders a progress bar node
func (d *declarativeRoot) renderProgress(node *ProgressVNode, x, y int, buffer *paint.Buffer) {
	// Check bounds
	if y < 0 || y >= buffer.Height {
		return
	}

	progressStyle := node.Style()
	if progressStyle.FG == "" {
		progressStyle = progressStyle.Foreground("white")
	}

	// Calculate progress
	percent := node.Percent()
	width := node.Width()
	if width < 10 {
		width = 10
	}

	// Calculate filled width
	filledWidth := (percent * width) / 100
	if filledWidth > width {
		filledWidth = width
	}

	// Build progress bar: [████░░░░] 75%
	bar := ""
	for i := 0; i < width; i++ {
		if i < filledWidth {
			bar += "█"
		} else {
			bar += "░"
		}
	}

	display := "[" + bar + "]"
	if node.ShowPercent() {
		display += fmt.Sprintf(" %d%%", percent)
	}
	if node.Label() != "" {
		display = node.Label() + " " + display
	}

	buffer.SetString(x, y, display, progressStyle)
}

// renderSpinner renders a spinner node
func (d *declarativeRoot) renderSpinner(node *SpinnerVNode, x, y int, buffer *paint.Buffer) {
	// Check bounds
	if y < 0 || y >= buffer.Height {
		return
	}

	spinnerStyle := node.Style()
	if spinnerStyle.FG == "" {
		spinnerStyle = spinnerStyle.Foreground("cyan")
	}

	frame := node.CurrentFrame()
	display := frame + " " + node.Message()

	buffer.SetString(x, y, display, spinnerStyle)
}

// renderSelect renders a select dropdown node
func (d *declarativeRoot) renderSelect(node *SelectVNode, x, y int, buffer *paint.Buffer) {
	// Check bounds
	if y < 0 || y >= buffer.Height {
		return
	}

	// Update focus state by using focusIndex instead of pointer comparison
	isFocused := d.focusedType == 4 && node.focusIndex >= 0 && node.focusIndex == d.focusedIndex
	node.SetFocus(isFocused)

	selectStyle := node.Style()
	if selectStyle.FG == "" {
		selectStyle = selectStyle.Foreground("white")
	}

	// Highlight focused select
	if node.IsFocused() {
		selectStyle = selectStyle.Background("blue").Foreground("white").Bold(true)
	}

	// Highlight hovered select
	if node.IsHovered() && !isFocused {
		selectStyle = selectStyle.Background("cyan").Foreground("black").Bold(true)
	}

	// Get display value
	display := node.SelectedLabel()
	if display == "" {
		display = "Select..."
	}

	// Calculate width based on longest option
	maxLen := len(display)
	for _, opt := range node.Options() {
		if len(opt.Label) > maxLen {
			maxLen = len(opt.Label)
		}
	}
	if maxLen < 10 {
		maxLen = 10
	}

	// Draw select box: [ option     ▼ ]
	padding := maxLen - len(display)
	if padding < 1 {
		padding = 1
	}
	paddingStr := ""
	for i := 0; i < padding; i++ {
		paddingStr += " "
	}
	fullSelect := "[" + display + paddingStr + "▼]"
	buffer.SetString(x, y, fullSelect, selectStyle)

	// Track bounds for mouse hit testing
	width := len(fullSelect)
	node.SetBounds(x, y, width, 1)
}

// renderTable renders a table node
func (d *declarativeRoot) renderTable(node *TableVNode, x, y int, buffer *paint.Buffer) int {
	// Return height (header + rows)
	height := 0
	if y+height >= buffer.Height {
		return 1
	}

	// Get base style
	baseStyle := node.Style()
	if baseStyle.FG == "" {
		baseStyle = baseStyle.Foreground("white")
	}

	// Render header
	headerStyle := baseStyle
	if node.HeaderStyle().FG != "" {
		headerStyle = node.HeaderStyle()
	} else {
		headerStyle = headerStyle.Bold(true).Foreground("cyan")
	}

	headerX := x
	for _, col := range node.Columns() {
		if headerX >= buffer.Width {
			break
		}
		title := col.Title
		if col.Width > 0 && len(title) > col.Width {
			title = title[:col.Width]
		}
		buffer.SetString(headerX, y+height, title, headerStyle)
		headerX += len(title) + 2 // +2 for separator
	}
	height++

	// Render separator line
	if y+height < buffer.Height {
		sepX := x
		for _, col := range node.Columns() {
			if sepX >= buffer.Width {
				break
			}
			width := col.Width
			if width <= 0 {
				width = len(col.Title)
			}
			sep := ""
			for i := 0; i < width+1; i++ {
				sep += "─"
			}
			buffer.SetString(sepX, y+height, sep, baseStyle.Foreground("bright-black"))
			sepX += width + 2
		}
		height++
	}

	// Render rows
	for _, row := range node.Rows() {
		if y+height >= buffer.Height {
			break
		}
		rowX := x
		for colIdx, cell := range row {
			if colIdx >= len(node.Columns()) {
				break
			}
			if rowX >= buffer.Width {
				break
			}
			col := node.Columns()[colIdx]
			display := cell
			if col.Width > 0 && len(display) > col.Width {
				display = display[:col.Width]
			}
			buffer.SetString(rowX, y+height, display, baseStyle)
			rowX += len(display) + 2
		}
		height++
	}

	return height
}

// measureWidth measures the width of a node
func (d *declarativeRoot) measureWidth(node VNode) int {
	switch n := node.(type) {
	case *TextVNode:
		return len(n.Content())
	case *ButtonVNode:
		return len(n.Label()) + 2 // +2 for brackets
	case *InputVNode:
		return 22 // [ + 20 chars + ]
	case *TextareaVNode:
		return n.Cols() + 2 // [ + content + ]
	case *CheckboxVNode:
		width := 4 // [X] or [ ]
		if n.Label() != "" {
			width += 1 + len(n.Label()) // space + label
		}
		return width
	case *ProgressVNode:
		width := n.Width() + 2 // [ + bar + ]
		if n.ShowPercent() {
			width += 5 // space + "100%"
		}
		if n.Label() != "" {
			width += len(n.Label()) + 1 // label + space
		}
		return width
	case *SpinnerVNode:
		return 2 + len(n.Message()) // frame + space + message
	case *SelectVNode:
		// Calculate based on options
		maxLen := 10 // minimum
		for _, opt := range n.Options() {
			if len(opt.Label) > maxLen {
				maxLen = len(opt.Label)
			}
		}
		return maxLen + 5 // [ + content + padding + ▼ ]
	case *TableVNode:
		// Calculate based on columns
		width := 0
		for _, col := range n.Columns() {
			colWidth := col.Width
			if colWidth <= 0 {
				colWidth = len(col.Title)
			}
			width += colWidth + 2 // +2 for separator
		}
		return width
	default:
		return 10 // Default width
	}
}

// collectInteractiveElements recursively collects all interactive elements
func (d *declarativeRoot) collectInteractiveElements(node VNode) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *ButtonVNode:
		if !n.Disabled() {
			n.focusIndex = len(d.buttons) // Set focus index before adding
			d.buttons = append(d.buttons, n)
		}
	case *InputVNode:
		if !n.Disabled() && !n.ReadOnly() {
			n.focusIndex = len(d.inputs) // Set focus index before adding
			d.inputs = append(d.inputs, n)
		}
	case *TextareaVNode:
		if !n.Disabled() {
			n.focusIndex = len(d.textareas) // Set focus index before adding
			d.textareas = append(d.textareas, n)
		}
	case *CheckboxVNode:
		if !n.Disabled() {
			n.focusIndex = len(d.checkboxes) // Set focus index before adding
			d.checkboxes = append(d.checkboxes, n)
		}
	case *SelectVNode:
		if !n.Disabled() {
			n.focusIndex = len(d.selects) // Set focus index before adding
			d.selects = append(d.selects, n)
		}
	case *ElementVNode:
		for _, child := range n.Children() {
			d.collectInteractiveElements(child)
		}
	case *LayoutNode:
		for _, child := range n.Children() {
			d.collectInteractiveElements(child)
		}
	case *GridVNode:
		for _, cell := range n.Cells() {
			d.collectInteractiveElements(cell.Child)
		}
	case *AbsoluteVNode:
		d.collectInteractiveElements(n.Child())
	case *VirtualListVNode:
		// Virtual list items might be interactive, but they're dynamically rendered
		// We'll collect from the actual rendered items during the render phase
	case *FragmentVNode:
		for _, child := range n.Children() {
			d.collectInteractiveElements(child)
		}
	case *ComponentVNode:
		// Use instance manager for persistent component state (same as renderVNode)
		componentKey := n.Key()
		if componentKey == "" {
			// Generate a stable key from component name if none provided
			componentKey = "component:" + n.Name()
		}

		// Track this component as active
		d.activeComponentKeys = append(d.activeComponentKeys, componentKey)

		// Get or create component instance
		instance := d.instanceManager.GetOrCreate(componentKey, func() ComponentInstance {
			if n.FnWithProps() != nil {
				return NewBaseComponentInstanceWithProps(componentKey, n.FnWithProps(), n.Props())
			}
			return NewBaseComponentInstance(componentKey, n.Fn())
		})

		// Update props if they changed
		if n.Props() != nil {
			instance.SetProps(n.Props())
		}

		// Render using the persistent instance with proper context setup
		oldContext := getCurrentContext()
		setCurrentContext(instance.GetContext())

		rendered := instance.Render()

		setCurrentContext(oldContext)

		// Recursively collect interactive elements from the rendered output
		if rendered != nil {
			d.collectInteractiveElements(rendered)
		}
	}
}

// resetInteractiveElements clears all interactive element collections
func (d *declarativeRoot) resetInteractiveElements() {
	d.buttons = d.buttons[:0]
	d.inputs = d.inputs[:0]
	d.textareas = d.textareas[:0]
	d.checkboxes = d.checkboxes[:0]
	d.selects = d.selects[:0]
}

// getFirstElementType returns the type of the first available element
func (d *declarativeRoot) getFirstElementType() int {
	if len(d.buttons) > 0 {
		return 0 // Button
	}
	if len(d.inputs) > 0 {
		return 1 // Input
	}
	if len(d.textareas) > 0 {
		return 2 // Textarea
	}
	if len(d.checkboxes) > 0 {
		return 3 // Checkbox
	}
	if len(d.selects) > 0 {
		return 4 // Select
	}
	return 0 // Default to button
}

// getTotalFocusableCount returns the total number of focusable elements
func (d *declarativeRoot) getTotalFocusableCount() int {
	return len(d.buttons) + len(d.inputs) + len(d.textareas) + len(d.checkboxes) + len(d.selects)
}

// getElementByIndex returns the element at the given focus index
func (d *declarativeRoot) getElementByIndex(index int) (interface{}, int) {
	buttonCount := len(d.buttons)
	inputCount := len(d.inputs)
	textareaCount := len(d.textareas)
	checkboxCount := len(d.checkboxes)

	if index < buttonCount {
		return d.buttons[index], 0
	}
	index -= buttonCount

	if index < inputCount {
		return d.inputs[index], 1
	}
	index -= inputCount

	if index < textareaCount {
		return d.textareas[index], 2
	}
	index -= textareaCount

	if index < checkboxCount {
		return d.checkboxes[index], 3
	}
	index -= checkboxCount

	selectCount := len(d.selects)
	if index < selectCount {
		return d.selects[index], 4
	}

	// Index out of bounds, return first available element or nil
	if len(d.buttons) > 0 {
		return d.buttons[0], 0
	}
	return nil, -1
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
	if os.Getenv("TUI_DEBUG_UI") == "true" {
		fmt.Fprintf(os.Stderr, "[HandleEvent] Called: Type=%v", ev.Type())
		if keyEv, ok := ev.(*frameworkevent.KeyEvent); ok {
			fmt.Fprintf(os.Stderr, ", Rune=%c, Special=%v, Modifiers=%v",
				keyEv.Key.Rune, keyEv.Special, keyEv.Modifiers)
		}
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "[HandleEvent] State: focusedIndex=%d, totalElements=%d, buttons=%d\n",
			d.focusedIndex, d.getTotalFocusableCount(), len(d.buttons))
	}

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

		// Get the focused element
		totalElements := d.getTotalFocusableCount()
		if totalElements == 0 {
			return false
		}

		// Handle character input for focused inputs
		if keyEv.Key.Rune != 0 && !keyEv.Key.Ctrl && !keyEv.Key.Alt {
			// Regular character input
			elem, elemType := d.getElementByIndex(d.focusedIndex)
			if elemType == 1 { // Input
				input := elem.(*InputVNode)
				// Don't check IsFocused() - we already know this is the focused element
				if !input.ReadOnly() && !input.Disabled() {
					d.handleInputChar(input, keyEv.Key.Rune)
					return true
				}
			} else if elemType == 3 && keyEv.Key.Rune == ' ' { // Checkbox + Space
				checkbox := elem.(*CheckboxVNode)
				newState := checkbox.Toggle()
				if onChange := checkbox.OnChange(); onChange != nil {
					onChange(newState)
				}
				if d.app != nil {
					d.app.MarkDirty()
				}
				return true
			}
			return false
		}

		// Handle special keys
		switch keyEv.Special {
		case frameworkevent.KeyTab:
			// Tab: focus next element
			if os.Getenv("TUI_DEBUG_UI") == "true" {
				fmt.Fprintf(os.Stderr, "[HandleEvent] KeyTab: focusedIndex=%d -> ", d.focusedIndex)
			}
			if keyEv.Modifiers == frameworkevent.ModShift {
				d.focusedIndex--
				if d.focusedIndex < 0 {
					d.focusedIndex = totalElements - 1
				}
			} else {
				d.focusedIndex++
				if d.focusedIndex >= totalElements {
					d.focusedIndex = 0
				}
			}
			d.updateFocusedType()
			if d.app != nil {
				d.app.MarkDirty()
			}
			if os.Getenv("TUI_DEBUG_UI") == "true" {
				fmt.Fprintf(os.Stderr, "%d\n", d.focusedIndex)
			}
			return true

		case frameworkevent.KeyBackspace:
			// Handle backspace for focused input
			elem, elemType := d.getElementByIndex(d.focusedIndex)
			if elemType == 1 { // Input
				input := elem.(*InputVNode)
				// Don't check IsFocused() - we already know this is the focused element
				if !input.ReadOnly() && !input.Disabled() {
					d.handleInputBackspace(input)
					return true
				}
			}
			return false

		case frameworkevent.KeyEnter:
			// Enter key behavior
			if os.Getenv("TUI_DEBUG_UI") == "true" {
				fmt.Fprintf(os.Stderr, "[HandleEvent] KeyEnter: focusedIndex=%d, totalElements=%d, buttons=%d\n",
					d.focusedIndex, totalElements, len(d.buttons))
			}
			elem, elemType := d.getElementByIndex(d.focusedIndex)
			if os.Getenv("TUI_DEBUG_UI") == "true" {
				fmt.Fprintf(os.Stderr, "[HandleEvent] getElementByIndex(%d) -> elemType=%d\n", d.focusedIndex, elemType)
			}
			if elemType == 1 { // Input
				input := elem.(*InputVNode)
				// Don't check IsFocused() - we already know this is the focused element
				// Trigger submit handler
				if onSubmit := input.OnSubmitFunc(); onSubmit != nil {
					onSubmit()
				}
				return true
			} else if elemType == 0 { // Button
				btn := elem.(*ButtonVNode)
				if os.Getenv("TUI_DEBUG_UI") == "true" {
					fmt.Fprintf(os.Stderr, "[HandleEvent] Triggering button: label=%s, hasOnClick=%v\n",
						btn.Label(), btn.OnClick() != nil)
				}
				if onClick := btn.OnClick(); onClick != nil {
					if os.Getenv("TUI_DEBUG_UI") == "true" {
						fmt.Fprintf(os.Stderr, "[HandleEvent] About to call onClick, ptr=%p\n", onClick)
					}
					// Wrap in defer to catch any panics
					func() {
						defer func() {
							if r := recover(); r != nil {
								if os.Getenv("TUI_DEBUG_UI") == "true" {
									fmt.Fprintf(os.Stderr, "[HandleEvent] PANIC in onClick: %v\n", r)
								}
							}
						}()
						onClick()
					}()
					if os.Getenv("TUI_DEBUG_UI") == "true" {
						fmt.Fprintf(os.Stderr, "[HandleEvent] Returned from onClick\n")
					}
				}
				// Mark dirty to trigger re-render after button click
				if d.app != nil {
					d.app.MarkDirty()
				}
				return true
			} else if elemType == 3 { // Checkbox
				checkbox := elem.(*CheckboxVNode)
				newState := checkbox.Toggle()
				if onChange := checkbox.OnChange(); onChange != nil {
					onChange(newState)
				}
				if d.app != nil {
					d.app.MarkDirty()
				}
				return true
			} else if elemType == 4 { // Select
				selectNode := elem.(*SelectVNode)
				// Cycle to next option
				nextIdx := selectNode.Selected() + 1
				if nextIdx >= len(selectNode.Options()) {
					nextIdx = 0
				}
				selectNode.SetSelected(nextIdx)
				if onChange := selectNode.OnChange(); onChange != nil {
					onChange(selectNode.SelectedValue())
				}
				if d.app != nil {
					d.app.MarkDirty()
				}
				return true
			}

		case frameworkevent.KeyLeft:
			// Left arrow - navigate to previous element
			d.focusedIndex--
			if d.focusedIndex < 0 {
				d.focusedIndex = totalElements - 1
			}
			d.updateFocusedType()
			if d.app != nil {
				d.app.MarkDirty()
			}
			return true

		case frameworkevent.KeyRight:
			// Right arrow - navigate to next element
			d.focusedIndex++
			if d.focusedIndex >= totalElements {
				d.focusedIndex = 0
			}
			d.updateFocusedType()
			if d.app != nil {
				d.app.MarkDirty()
			}
			return true

		case frameworkevent.KeyUp:
			// Up arrow - for select, cycle to previous option
			elem, elemType := d.getElementByIndex(d.focusedIndex)
			if elemType == 4 { // Select
				selectNode := elem.(*SelectVNode)
				nextIdx := selectNode.Selected() - 1
				if nextIdx < -1 {
					nextIdx = len(selectNode.Options()) - 1
				}
				selectNode.SetSelected(nextIdx)
				if onChange := selectNode.OnChange(); onChange != nil {
					onChange(selectNode.SelectedValue())
				}
				if d.app != nil {
					d.app.MarkDirty()
				}
				return true
			}

		case frameworkevent.KeyDown:
			// Down arrow - for select, cycle to next option
			elem, elemType := d.getElementByIndex(d.focusedIndex)
			if elemType == 4 { // Select
				selectNode := elem.(*SelectVNode)
				nextIdx := selectNode.Selected() + 1
				if nextIdx >= len(selectNode.Options()) {
					nextIdx = 0
				}
				selectNode.SetSelected(nextIdx)
				if onChange := selectNode.OnChange(); onChange != nil {
					onChange(selectNode.SelectedValue())
				}
				if d.app != nil {
					d.app.MarkDirty()
				}
				return true
			}
		}
	}

	// Handle mouse events
	if mouseEv, ok := ev.(*frameworkevent.MouseEvent); ok {
		return d.handleMouseEvent(mouseEv)
	}

	// Mark for re-render on any event
	return true
}

// updateFocusedType updates the focused type based on current index
// and sets the isFocused flag on the actual elements
func (d *declarativeRoot) updateFocusedType() {
	elem, elemType := d.getElementByIndex(d.focusedIndex)
	d.focusedType = elemType

	// Clear focus from all elements that support it
	for _, input := range d.inputs {
		input.SetFocus(false)
	}
	for _, checkbox := range d.checkboxes {
		checkbox.SetFocus(false)
	}
	for _, sel := range d.selects {
		sel.SetFocus(false)
	}
	for _, textarea := range d.textareas {
		textarea.SetFocus(false)
	}

	// Set focus on the currently focused element
	if elem != nil {
		switch e := elem.(type) {
		case *InputVNode:
			e.SetFocus(true)
		case *CheckboxVNode:
			e.SetFocus(true)
		case *SelectVNode:
			e.SetFocus(true)
		case *TextareaVNode:
			e.SetFocus(true)
		}
	}
}

// handleInputChar handles character input for an input field
func (d *declarativeRoot) handleInputChar(input *InputVNode, ch rune) {
	currentValue := input.Value()
	maxLength := input.MaxLength()

	// Check max length
	if maxLength > 0 && len(currentValue) >= maxLength {
		return
	}

	// Append character
	newValue := currentValue + string(ch)
	input.SetValue(newValue)

	// Trigger change handler
	if onChange := input.OnChange(); onChange != nil {
		onChange(newValue)
	}

	// Mark dirty for re-render
	if d.app != nil {
		d.app.MarkDirty()
	}
}

// handleInputBackspace handles backspace for an input field
func (d *declarativeRoot) handleInputBackspace(input *InputVNode) {
	currentValue := input.Value()
	if len(currentValue) == 0 {
		return
	}

	// Remove last character
	newValue := currentValue[:len(currentValue)-1]
	input.SetValue(newValue)

	// Trigger change handler
	if onChange := input.OnChange(); onChange != nil {
		onChange(newValue)
	}

	// Mark dirty for re-render
	if d.app != nil {
		d.app.MarkDirty()
	}
}

// handleMouseEvent handles mouse events with priority-based scheduling
func (d *declarativeRoot) handleMouseEvent(ev *frameworkevent.MouseEvent) bool {
	x, y := ev.X, ev.Y

	// Update mouse tracker
	d.mouseTracker.UpdatePosition(x, y)

	// Convert button to int
	button := int(ev.Button)

	// Route events through handlers for throttling/debouncing
	switch ev.Type() {
	case frameworkevent.EventMouseMove:
		// Use throttled handler for mouse move
		d.mouseMoveHandler.Handle(x, y)
		// Still dispatch for immediate feedback
		return d.dispatchMouseEvent(ev, x, y)

	case frameworkevent.EventMousePress:
		// Update button state in tracker
		d.mouseTracker.UpdateButton(button, true)
		// Always handle press immediately for responsiveness
		return d.dispatchMouseEvent(ev, x, y)

	case frameworkevent.EventMouseRelease:
		// Update button state in tracker
		d.mouseTracker.UpdateButton(button, false)
		// Always handle release immediately for responsiveness
		return d.dispatchMouseEvent(ev, x, y)

	case frameworkevent.EventClick:
		// Use debounced handler for clicks
		d.mouseClickHandler.Handle(x, y, button)
		// Handle click immediately too for responsiveness
		return d.dispatchMouseEvent(ev, x, y)

	default:
		// Handle other events normally
		return d.dispatchMouseEvent(ev, x, y)
	}
}

// dispatchMouseEvent dispatches the mouse event to all interactive elements
func (d *declarativeRoot) dispatchMouseEvent(ev *frameworkevent.MouseEvent, x, y int) bool {
	handled := false

	// Check if mouse is over any button
	for _, btn := range d.buttons {
		if btn.ContainsPoint(x, y) {
			// Mouse is over this button
			if ev.Type() == frameworkevent.EventMouseEnter || ev.Type() == frameworkevent.EventMouseMove {
				if !btn.IsHovered() {
					btn.SetHovered(true)
					if d.app != nil {
						d.app.MarkDirty()
					}
				}
				// Dispatch mouse enter event to button
				if btn.HandleEvent(ev) {
					handled = true
				}
			} else if ev.Type() == frameworkevent.EventMousePress || ev.Type() == frameworkevent.EventMouseRelease || ev.Type() == frameworkevent.EventClick {
				// Dispatch mouse press/release/click to button
				// In Fiber mode, isHovered might not be set, so handle clicks directly
				if btn.HandleEvent(ev) {
					if d.app != nil {
						d.app.MarkDirty()
					}
					handled = true
				} else {
					// Button didn't handle it (likely due to isHovered=false in Fiber mode)
					// Handle click directly for buttons at this position
					onClick := btn.OnClick()
					if onClick != nil {
						onClick()
						if d.app != nil {
							d.app.MarkDirty()
						}
						handled = true
					}
				}
			}
		} else {
			// Mouse is not over this button
			if btn.IsHovered() {
				btn.SetHovered(false)
				if d.app != nil {
					d.app.MarkDirty()
				}
			}
		}
	}

	// Check if mouse is over any checkbox
	for _, chk := range d.checkboxes {
		if chk.ContainsPoint(x, y) {
			// Mouse is over this checkbox
			if ev.Type() == frameworkevent.EventMouseEnter || ev.Type() == frameworkevent.EventMouseMove {
				if !chk.IsHovered() {
					chk.SetHovered(true)
					if d.app != nil {
						d.app.MarkDirty()
					}
				}
				// Dispatch mouse enter event to checkbox
				if chk.HandleEvent(ev) {
					handled = true
				}
			} else if ev.Type() == frameworkevent.EventMousePress || ev.Type() == frameworkevent.EventMouseRelease || ev.Type() == frameworkevent.EventClick {
				// Dispatch mouse press/release/click to checkbox
				if chk.HandleEvent(ev) {
					if d.app != nil {
						d.app.MarkDirty()
					}
					handled = true
				}
			}
		} else {
			// Mouse is not over this checkbox
			if chk.IsHovered() {
				chk.SetHovered(false)
				if d.app != nil {
					d.app.MarkDirty()
				}
			}
		}
	}

	// Check if mouse is over any input
	for _, inp := range d.inputs {
		if inp.ContainsPoint(x, y) {
			// Mouse is over this input
			if ev.Type() == frameworkevent.EventMouseEnter || ev.Type() == frameworkevent.EventMouseMove {
				if !inp.IsHovered() {
					inp.SetHovered(true)
					if d.app != nil {
						d.app.MarkDirty()
					}
				}
				// Dispatch mouse enter event to input
				if inp.HandleEvent(ev) {
					handled = true
				}
			} else if ev.Type() == frameworkevent.EventMousePress || ev.Type() == frameworkevent.EventMouseRelease || ev.Type() == frameworkevent.EventClick {
				// Dispatch mouse press/release/click to input
				if inp.HandleEvent(ev) {
					if d.app != nil {
						d.app.MarkDirty()
					}
					return true
				}
			}
		} else {
			// Mouse is not over this input
			if inp.IsHovered() {
				inp.SetHovered(false)
				if d.app != nil {
					d.app.MarkDirty()
				}
			}
		}
	}

	// Check if mouse is over any select
	for _, sel := range d.selects {
		if sel.ContainsPoint(x, y) {
			// Mouse is over this select
			if ev.Type() == frameworkevent.EventMouseEnter || ev.Type() == frameworkevent.EventMouseMove {
				if !sel.IsHovered() {
					sel.SetHovered(true)
					if d.app != nil {
						d.app.MarkDirty()
					}
				}
				// Dispatch mouse enter event to select
				if sel.HandleEvent(ev) {
					handled = true
				}
			} else if ev.Type() == frameworkevent.EventMousePress || ev.Type() == frameworkevent.EventMouseRelease || ev.Type() == frameworkevent.EventClick {
				// Dispatch mouse press/release/click to select
				if sel.HandleEvent(ev) {
					if d.app != nil {
						d.app.MarkDirty()
					}
					handled = true
				}
			}
		} else {
			// Mouse is not over this select
			if sel.IsHovered() {
				sel.SetHovered(false)
				if d.app != nil {
					d.app.MarkDirty()
				}
			}
		}
	}

	// Check if mouse is over any textarea
	for _, ta := range d.textareas {
		if ta.ContainsPoint(x, y) {
			// Mouse is over this textarea
			if ev.Type() == frameworkevent.EventMouseEnter || ev.Type() == frameworkevent.EventMouseMove {
				if !ta.IsHovered() {
					ta.SetHovered(true)
					if d.app != nil {
						d.app.MarkDirty()
					}
				}
				// Dispatch mouse enter event to textarea
				if ta.HandleEvent(ev) {
					handled = true
				}
			} else if ev.Type() == frameworkevent.EventMousePress || ev.Type() == frameworkevent.EventMouseRelease || ev.Type() == frameworkevent.EventClick {
				// Dispatch mouse press/release/click to textarea
				if ta.HandleEvent(ev) {
					if d.app != nil {
						d.app.MarkDirty()
					}
					handled = true
				}
			}
		} else {
			// Mouse is not over this textarea
			if ta.IsHovered() {
				ta.SetHovered(false)
				if d.app != nil {
					d.app.MarkDirty()
				}
			}
		}
	}

	// Handle mouse leave event - clear all hover states
	if ev.Type() == frameworkevent.EventMouseLeave {
		for _, btn := range d.buttons {
			if btn.IsHovered() {
				btn.SetHovered(false)
				btn.HandleEvent(ev)
			}
		}
		for _, chk := range d.checkboxes {
			if chk.IsHovered() {
				chk.SetHovered(false)
				chk.HandleEvent(ev)
			}
		}
		for _, inp := range d.inputs {
			if inp.IsHovered() {
				inp.SetHovered(false)
				inp.HandleEvent(ev)
			}
		}
		for _, sel := range d.selects {
			if sel.IsHovered() {
				sel.SetHovered(false)
				sel.HandleEvent(ev)
			}
		}
		for _, ta := range d.textareas {
			if ta.IsHovered() {
				ta.SetHovered(false)
				ta.HandleEvent(ev)
			}
		}
		if d.app != nil {
			d.app.MarkDirty()
		}
		return true
	}

	return handled
}

// =============================================================================
// Getter methods for testing/debugging
// =============================================================================

// GetButtons returns the collected buttons (for testing)
func (d *declarativeRoot) GetButtons() []*ButtonVNode {
	return d.buttons
}

// GetInputs returns the collected inputs (for testing)
func (d *declarativeRoot) GetInputs() []*InputVNode {
	return d.inputs
}

// GetTextareas returns the collected textareas (for testing)
func (d *declarativeRoot) GetTextareas() []*TextareaVNode {
	return d.textareas
}

// GetCheckboxes returns the collected checkboxes (for testing)
func (d *declarativeRoot) GetCheckboxes() []*CheckboxVNode {
	return d.checkboxes
}

// GetSelects returns the collected selects (for testing)
func (d *declarativeRoot) GetSelects() []*SelectVNode {
	return d.selects
}

// GetFocusedIndex returns the current focused element index (for testing)
func (d *declarativeRoot) GetFocusedIndex() int {
	return d.focusedIndex
}

// GetFocusedType returns the current focused element type (for testing)
func (d *declarativeRoot) GetFocusedType() int {
	return d.focusedType
}

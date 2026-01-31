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
	appFn            ComponentFunc
	ctx              *ComponentContext
	app              *framework.App
	buttons          []*ButtonVNode   // Collected buttons for focus management
	inputs           []*InputVNode    // Collected inputs for focus management
	textareas        []*TextareaVNode // Collected textareas for focus management
	checkboxes       []*CheckboxVNode // Collected checkboxes for focus management
	selects          []*SelectVNode   // Collected selects for focus management
	focusedIndex     int              // Currently focused element index (-1 = none)
	focusedType      int              // Type: 0=button, 1=input, 2=textarea, 3=checkbox, 4=select
	renderBtnIndex   int              // Current button index during rendering
	renderInputIndex int              // Current input index during rendering
}

// newDeclarativeRoot creates a new declarative root component
func newDeclarativeRoot(fn ComponentFunc, app *framework.App) component.Node {
	return &declarativeRoot{
		appFn:       fn,
		ctx:         newComponentContext("App"),
		app:         app,
		buttons:     make([]*ButtonVNode, 0),
		inputs:      make([]*InputVNode, 0),
		textareas:   make([]*TextareaVNode, 0),
		checkboxes:  make([]*CheckboxVNode, 0),
		selects:     make([]*SelectVNode, 0),
		focusedIndex: -1,
		focusedType: 0,
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

	// Run effects after render completes
	d.ctx.runEffects()

	// Clear and collect interactive elements for focus management
	d.buttons = d.buttons[:0]
	d.inputs = d.inputs[:0]
	d.textareas = d.textareas[:0]
	d.checkboxes = d.checkboxes[:0]
	d.selects = d.selects[:0]
	d.collectInteractiveElements(vnode)

	// Auto-focus first element if none focused
	if d.focusedIndex < 0 {
		d.focusedIndex = 0
		d.focusedType = d.getFirstElementType()
	}

	// Reset render indices for tracking during render
	d.renderBtnIndex = 0
	d.renderInputIndex = 0

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

	case *InputVNode:
		d.renderInput(n, x, currentY, buffer)
		currentY += 1

	case *TextareaVNode:
		// Textarea can be multiple lines
		height := n.Rows()
		d.renderTextarea(n, x, currentY, buffer)
		currentY += height

	case *CheckboxVNode:
		d.renderCheckbox(n, x, currentY, buffer)
		currentY += 1

	case *ProgressVNode:
		d.renderProgress(n, x, currentY, buffer)
		currentY += 1

	case *SpinnerVNode:
		d.renderSpinner(n, x, currentY, buffer)
		currentY += 1

	case *SelectVNode:
		d.renderSelect(n, x, currentY, buffer)
		currentY += 1

	case *TableVNode:
		height := d.renderTable(n, x, currentY, buffer)
		currentY += height

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

// renderInput renders an input node
func (d *declarativeRoot) renderInput(node *InputVNode, x, y int, buffer *paint.Buffer) {
	// Check bounds
	if y < 0 || y >= buffer.Height {
		return
	}

	// Update focus state
	node.SetFocus(d.renderInputIndex == d.focusedIndex && d.focusedType == 1)
	d.renderInputIndex++

	inputStyle := node.Style()
	if inputStyle.FG == "" {
		inputStyle = inputStyle.Foreground("white")
	}

	// Highlight focused input
	if node.IsFocused() {
		inputStyle = inputStyle.Background("blue").Foreground("white").Bold(true)
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

	fullInput := "[" + displayValue + string(make([]byte, padding)) + "]"
	buffer.SetString(x, y, fullInput, inputStyle)
}

// renderTextarea renders a textarea node
func (d *declarativeRoot) renderTextarea(node *TextareaVNode, x, y int, buffer *paint.Buffer) {
	// Update focus state
	node.SetFocus(d.renderInputIndex == d.focusedIndex && d.focusedType == 2)
	d.renderInputIndex++

	rows := node.Rows()
	if rows < 1 {
		rows = 1
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

		// Simple multi-line rendering
		value := node.Value()
		if value == "" {
			value = node.Placeholder()
			if value == "" {
				value = " "
			}
		}

		// Draw row
		buffer.SetString(x, renderY, "["+value+"]", textareaStyle)
	}
}

// renderCheckbox renders a checkbox node
func (d *declarativeRoot) renderCheckbox(node *CheckboxVNode, x, y int, buffer *paint.Buffer) {
	// Check bounds
	if y < 0 || y >= buffer.Height {
		return
	}

	// Update focus state using same index system as inputs
	node.SetFocus(d.renderInputIndex == d.focusedIndex && d.focusedType == 3)
	d.renderInputIndex++

	checkboxStyle := node.Style()
	if checkboxStyle.FG == "" {
		checkboxStyle = checkboxStyle.Foreground("white")
	}

	// Highlight focused checkbox
	if node.IsFocused() {
		checkboxStyle = checkboxStyle.Background("blue").Foreground("white").Bold(true)
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

	// Update focus state
	node.SetFocus(d.renderInputIndex == d.focusedIndex && d.focusedType == 4)
	d.renderInputIndex++

	selectStyle := node.Style()
	if selectStyle.FG == "" {
		selectStyle = selectStyle.Foreground("white")
	}

	// Highlight focused select
	if node.IsFocused() {
		selectStyle = selectStyle.Background("blue").Foreground("white").Bold(true)
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
	fullSelect := "[" + display + string(make([]byte, padding)) + "▼]"
	buffer.SetString(x, y, fullSelect, selectStyle)
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
			d.buttons = append(d.buttons, n)
		}
	case *InputVNode:
		if !n.Disabled() && !n.ReadOnly() {
			d.inputs = append(d.inputs, n)
		}
	case *TextareaVNode:
		if !n.Disabled() {
			d.textareas = append(d.textareas, n)
		}
	case *CheckboxVNode:
		if !n.Disabled() {
			d.checkboxes = append(d.checkboxes, n)
		}
	case *SelectVNode:
		if !n.Disabled() {
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
	case *FragmentVNode:
		for _, child := range n.Children() {
			d.collectInteractiveElements(child)
		}
	case *ComponentVNode:
		rendered := n.Render()
		if rendered != nil {
			d.collectInteractiveElements(rendered)
		}
	}
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

	return d.selects[index], 4
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
				if input.IsFocused() && !input.ReadOnly() && !input.Disabled() {
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
			return true

		case frameworkevent.KeyBackspace:
			// Handle backspace for focused input
			elem, elemType := d.getElementByIndex(d.focusedIndex)
			if elemType == 1 { // Input
				input := elem.(*InputVNode)
				if input.IsFocused() && !input.ReadOnly() && !input.Disabled() {
					d.handleInputBackspace(input)
					return true
				}
			}
			return false

		case frameworkevent.KeyEnter:
			// Enter key behavior
			elem, elemType := d.getElementByIndex(d.focusedIndex)
			if elemType == 1 { // Input
				input := elem.(*InputVNode)
				if input.IsFocused() {
					// Trigger submit handler
					if onSubmit := input.OnSubmitFunc(); onSubmit != nil {
						onSubmit()
					}
					return true
				}
			} else if elemType == 0 { // Button
				btn := elem.(*ButtonVNode)
				if onClick := btn.OnClick(); onClick != nil {
					onClick()
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

	// Mark for re-render on any event
	return true
}

// updateFocusedType updates the focused type based on current index
func (d *declarativeRoot) updateFocusedType() {
	_, elemType := d.getElementByIndex(d.focusedIndex)
	d.focusedType = elemType
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

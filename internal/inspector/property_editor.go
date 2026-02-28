package inspector

import (
	"fmt"
	"strconv"
	"strings"

	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/runtime"
	"github.com/wwsheng009/mint/runtime/style"
)

// PropertyEdit represents a property change
type PropertyEdit struct {
	Element   rtui.VNode
	Property  string
	OldValue  interface{}
	NewValue  interface{}
	Applied   bool
}

// PropertyEditor allows editing VNode properties
type PropertyEditor struct {
	history []PropertyEdit
	enabled bool
}

// NewPropertyEditor creates a new property editor
func NewPropertyEditor() *PropertyEditor {
	return &PropertyEditor{
		history: make([]PropertyEdit, 0),
		enabled: true,
	}
}

// Enable enables property editing
func (pe *PropertyEditor) Enable() {
	pe.enabled = true
}

// Disable disables property editing
func (pe *PropertyEditor) Disable() {
	pe.enabled = false
}

// IsEnabled returns whether editing is enabled
func (pe *PropertyEditor) IsEnabled() bool {
	return pe.enabled
}

// EditFlex edits the flex value of a VNode
func (pe *PropertyEditor) EditFlex(vnode rtui.VNode, newFlex int) error {
	if !pe.enabled {
		return fmt.Errorf("property editor is disabled")
	}

	// Get old value
	oldFlex := pe.getFlexValue(vnode)

	// Apply new value
	err := pe.setFlexValue(vnode, newFlex)
	if err != nil {
		return err
	}

	// Record edit
	pe.history = append(pe.history, PropertyEdit{
		Element:   vnode,
		Property:  "Flex",
		OldValue:  oldFlex,
		NewValue:  newFlex,
		Applied:   true,
	})

	return nil
}

// EditConstraints edits the constraints of a VNode
func (pe *PropertyEditor) EditConstraints(vnode rtui.VNode, newConstraints runtime.BoxConstraints) error {
	if !pe.enabled {
		return fmt.Errorf("property editor is disabled")
	}

	// Get old value
	oldConstraints := pe.getConstraintsValue(vnode)

	// Apply new value
	err := pe.setConstraintsValue(vnode, newConstraints)
	if err != nil {
		return err
	}

	// Record edit
	pe.history = append(pe.history, PropertyEdit{
		Element:   vnode,
		Property:  "Constraints",
		OldValue:  oldConstraints,
		NewValue:  newConstraints,
		Applied:   true,
	})

	return nil
}

// EditPadding edits the padding of a VNode
func (pe *PropertyEditor) EditPadding(vnode rtui.VNode, newPadding int) error {
	if !pe.enabled {
		return fmt.Errorf("property editor is disabled")
	}

	// Get old value
	oldPadding := pe.getPaddingValue(vnode)

	// Apply new value
	err := pe.setPaddingValue(vnode, newPadding)
	if err != nil {
		return err
	}

	// Record edit
	pe.history = append(pe.history, PropertyEdit{
		Element:   vnode,
		Property:  "Padding",
		OldValue:  oldPadding,
		NewValue:  newPadding,
		Applied:   true,
	})

	return nil
}

// EditStyle edits the style of a VNode
func (pe *PropertyEditor) EditStyle(vnode rtui.VNode, newStyle style.Style) error {
	if !pe.enabled {
		return fmt.Errorf("property editor is disabled")
	}

	// Get old value
	oldStyle := pe.getStyleValue(vnode)

	// Apply new value
	err := pe.setStyleValue(vnode, newStyle)
	if err != nil {
		return err
	}

	// Record edit
	pe.history = append(pe.history, PropertyEdit{
		Element:   vnode,
		Property:  "Style",
		OldValue:  oldStyle,
		NewValue:  newStyle,
		Applied:   true,
	})

	return nil
}

// GetHistory returns edit history
func (pe *PropertyEditor) GetHistory() []PropertyEdit {
	return pe.history
}

// ClearHistory clears edit history
func (pe *PropertyEditor) ClearHistory() {
	pe.history = make([]PropertyEdit, 0)
}

// FormatHistory formats edit history as text
func (pe *PropertyEditor) FormatHistory() string {
	if len(pe.history) == 0 {
		return "No edits applied"
	}

	var lines []string
	lines = append(lines, "┌─ Edit History ─────────────────────────────────┐")

	for i, edit := range pe.history {
		if i >= 20 { // Limit history display
			lines = append(lines, fmt.Sprintf("│ ... and %d more edits                        │",
				len(pe.history)-20))
			break
		}

		// Get element type
		info := ExtractElementInfo(edit.Element)
		elementType := info.Type

		lines = append(lines, fmt.Sprintf("│ [%d] %s.%s                          │",
			i+1, truncateString(elementType, 15), edit.Property))
		lines = append(lines, fmt.Sprintf("│     %s → %s%-29s │",
			formatValue(edit.OldValue), formatValue(edit.NewValue), ""))
	}

	lines = append(lines, "└────────────────────────────────────────────────┘")

	return joinLines(lines)
}

// Helper methods for getting/setting values

func (pe *PropertyEditor) getFlexValue(vnode rtui.VNode) int {
	if props := vnode.Props(); props != nil {
		if flex := props.GetInt("flex"); flex > 0 {
			return flex
		}
	}
	return 0
}

func (pe *PropertyEditor) setFlexValue(vnode rtui.VNode, value int) error {
	if props := vnode.Props(); props != nil {
		props.Set("flex", value)
		return nil
	}
	return fmt.Errorf("element does not support flex property")
}

func (pe *PropertyEditor) getConstraintsValue(vnode rtui.VNode) runtime.BoxConstraints {
	if constrainted, ok := vnode.(interface{ GetConstraints() runtime.BoxConstraints }); ok {
		return constrainted.GetConstraints()
	}
	return runtime.BoxConstraints{}
}

func (pe *PropertyEditor) setConstraintsValue(vnode rtui.VNode, value runtime.BoxConstraints) error {
	if constrainted, ok := vnode.(interface{ SetConstraints(runtime.BoxConstraints) }); ok {
		constrainted.SetConstraints(value)
		return nil
	}
	return fmt.Errorf("element does not support constraints property")
}

func (pe *PropertyEditor) getPaddingValue(vnode rtui.VNode) int {
	if props := vnode.Props(); props != nil {
		if padding := props.GetInt("padding"); padding >= 0 {
			return padding
		}
	}
	return 0
}

func (pe *PropertyEditor) setPaddingValue(vnode rtui.VNode, value int) error {
	if props := vnode.Props(); props != nil {
		props.Set("padding", value)
		return nil
	}
	return fmt.Errorf("element does not support padding property")
}

func (pe *PropertyEditor) getStyleValue(vnode rtui.VNode) style.Style {
	if styled, ok := vnode.(interface{ GetStyle() style.Style }); ok {
		return styled.GetStyle()
	}
	return style.Style{}
}

func (pe *PropertyEditor) setStyleValue(vnode rtui.VNode, value style.Style) error {
	// VNode interface already has SetStyle method, no type assertion needed
	vnode.SetStyle(value)
	return nil
}

// formatValue formats a value for display
func formatValue(value interface{}) string {
	switch v := value.(type) {
	case int:
		return strconv.Itoa(v)
	case float64:
		return strconv.FormatFloat(v, 'f', 2, 64)
	case string:
		if len(v) > 20 {
			return truncateString(v, 20)
		}
		return v
	case runtime.BoxConstraints:
		return fmt.Sprintf("Constraints{MinW:%d, MaxW:%d}", v.MinWidth, v.MaxWidth)
	case style.Style:
		return "Style{...}"
	default:
		return fmt.Sprintf("%v", v)
	}
}

// ParseValue parses a string value into the appropriate type
func ParseValue(valueStr string, targetType string) (interface{}, error) {
	valueStr = strings.TrimSpace(valueStr)

	switch targetType {
	case "int":
		return strconv.Atoi(valueStr)
	case "flex":
		return strconv.Atoi(valueStr)
	case "padding":
		return strconv.Atoi(valueStr)
	case "string":
		return valueStr, nil
	default:
		return nil, fmt.Errorf("unsupported target type: %s", targetType)
	}
}

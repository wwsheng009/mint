package collapse

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

const (
	propAccordion         = "accordion"
	propActiveKeys        = "activeKeys"
	propActiveKeysControl = "activeKeysControlled"
	propBordered          = "bordered"
	propChangeIntent      = "changeIntent"
	propChangeIntentField = "changeIntentField"
	propComponentID       = "componentID"
	propContentStyle      = "contentStyle"
	propDisabled          = "disabled"
	propGhost             = "ghost"
	propHeaderStyle       = "headerStyle"
	propActiveHeaderStyle = "activeHeaderStyle"
	propInitialActiveKeys = "initialActiveKeys"
	propItems             = "items"
	propKey               = "key"
	propStyle             = "style"
	propWidth             = "width"
)

// Item describes one collapsible panel.
type Item struct {
	Key      string
	Header   string
	Content  rtui.VNode
	Extra    rtui.VNode
	Disabled bool
}

// Section creates a collapse item with header and content.
func Section(header string, content rtui.VNode) Item {
	return Item{
		Header:  header,
		Content: content,
	}
}

// WithKey sets the panel key.
func (i Item) WithKey(key string) Item {
	i.Key = key
	return i
}

// WithExtra adds header-side extra content.
func (i Item) WithExtra(extra rtui.VNode) Item {
	i.Extra = extra
	return i
}

// WithDisabled marks the panel as disabled.
func (i Item) WithDisabled(disabled bool) Item {
	i.Disabled = disabled
	return i
}

// VNode is the declarative description of a Collapse component.
type VNode struct {
	*rtui.ElementVNode

	key                  string
	componentID          string
	items                []Item
	accordion            bool
	activeKeys           []string
	initialActiveKeys    []string
	activeKeysControlled bool
	disabled             bool
	bordered             bool
	ghost                bool
	width                int
	collapseStyle        style.Style
	headerStyle          style.Style
	activeHeaderStyle    style.Style
	contentStyle         style.Style
	changeIntent         intent.Intent
	changeIntentField    intent.FieldIntent
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// New creates a new Collapse VNode.
func New(items []Item) *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("collapse"),
		items:        normalizeItems(items),
		bordered:     true,
	}
}

func (v *VNode) Key() string { return v.key }

func (v *VNode) SetKey(key string) rtui.VNode {
	v.key = key
	return v
}

func (v *VNode) Tag() string { return "collapse" }

func (v *VNode) Style() style.Style { return v.collapseStyle }

func (v *VNode) SetStyle(s style.Style) rtui.VNode {
	v.collapseStyle = s
	return v
}

func (v *VNode) Children() []rtui.VNode { return nil }

func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode { return v }

func (v *VNode) GetLayer() rtui.Layer { return rtui.LayerBase }

func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode { return v }

func (v *VNode) Props() rtui.Props {
	return rtui.Props{
		propAccordion:         v.accordion,
		propActiveHeaderStyle: v.activeHeaderStyle,
		propActiveKeys:        cloneStrings(v.activeKeys),
		propActiveKeysControl: v.activeKeysControlled,
		propBordered:          v.bordered,
		propChangeIntent:      v.changeIntent,
		propChangeIntentField: v.changeIntentField,
		propComponentID:       v.componentID,
		propContentStyle:      v.contentStyle,
		propDisabled:          v.disabled,
		propGhost:             v.ghost,
		propHeaderStyle:       v.headerStyle,
		propInitialActiveKeys: cloneStrings(v.initialActiveKeys),
		propItems:             cloneItems(v.items),
		propKey:               v.key,
		propStyle:             v.collapseStyle,
		propWidth:             v.width,
	}
}

func (v *VNode) SetProps(props rtui.Props) rtui.VNode {
	if key, ok := props[propKey].(string); ok {
		v.key = key
	}
	if componentID, ok := props[propComponentID].(string); ok {
		v.componentID = componentID
	}
	if items, ok := props[propItems].([]Item); ok {
		v.items = normalizeItems(items)
	}
	if accordion, ok := props[propAccordion].(bool); ok {
		v.accordion = accordion
	}
	if activeKeys, ok := props[propActiveKeys].([]string); ok {
		v.activeKeys = cloneStrings(activeKeys)
	}
	if initialActiveKeys, ok := props[propInitialActiveKeys].([]string); ok {
		v.initialActiveKeys = cloneStrings(initialActiveKeys)
	}
	if activeKeysControlled, ok := props[propActiveKeysControl].(bool); ok {
		v.activeKeysControlled = activeKeysControlled
	}
	if disabled, ok := props[propDisabled].(bool); ok {
		v.disabled = disabled
	}
	if bordered, ok := props[propBordered].(bool); ok {
		v.bordered = bordered
	}
	if ghost, ok := props[propGhost].(bool); ok {
		v.ghost = ghost
	}
	if width, ok := props[propWidth].(int); ok {
		v.width = width
	}
	if s, ok := props[propStyle].(style.Style); ok {
		v.collapseStyle = s
	}
	if s, ok := props[propHeaderStyle].(style.Style); ok {
		v.headerStyle = s
	}
	if s, ok := props[propActiveHeaderStyle].(style.Style); ok {
		v.activeHeaderStyle = s
	}
	if s, ok := props[propContentStyle].(style.Style); ok {
		v.contentStyle = s
	}
	if changeIntent, ok := props[propChangeIntent].(intent.Intent); ok {
		v.changeIntent = changeIntent
	}
	if changeIntentField, ok := props[propChangeIntentField].(intent.FieldIntent); ok {
		v.changeIntentField = changeIntentField
	}
	return v
}

func (v *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(v.Props())
}

// SetComponentID sets the event routing component ID.
func (v *VNode) SetComponentID(id string) *VNode {
	v.componentID = id
	return v
}

// SetItems replaces all panels.
func (v *VNode) SetItems(items []Item) *VNode {
	v.items = normalizeItems(items)
	return v
}

// AddItem appends a panel.
func (v *VNode) AddItem(item Item) *VNode {
	v.items = normalizeItems(append(v.items, item))
	return v
}

// SetAccordion toggles accordion mode.
func (v *VNode) SetAccordion(accordion bool) *VNode {
	v.accordion = accordion
	return v
}

// SetActiveKeys sets controlled expanded panel keys.
func (v *VNode) SetActiveKeys(keys []string) *VNode {
	v.activeKeys = cloneStrings(keys)
	v.activeKeysControlled = true
	return v
}

// SetInitialActiveKeys sets uncontrolled initial expanded keys.
func (v *VNode) SetInitialActiveKeys(keys []string) *VNode {
	v.initialActiveKeys = cloneStrings(keys)
	return v
}

// SetDisabled disables the entire component.
func (v *VNode) SetDisabled(disabled bool) *VNode {
	v.disabled = disabled
	return v
}

// SetBordered toggles panel borders.
func (v *VNode) SetBordered(bordered bool) *VNode {
	v.bordered = bordered
	return v
}

// SetGhost toggles ghost style.
func (v *VNode) SetGhost(ghost bool) *VNode {
	v.ghost = ghost
	return v
}

// SetWidth sets the panel width.
func (v *VNode) SetWidth(width int) *VNode {
	v.width = width
	return v
}

// SetHeaderStyle sets the header button style.
func (v *VNode) SetHeaderStyle(s style.Style) *VNode {
	v.headerStyle = s
	return v
}

// SetActiveHeaderStyle sets the expanded header style.
func (v *VNode) SetActiveHeaderStyle(s style.Style) *VNode {
	v.activeHeaderStyle = s
	return v
}

// SetContentStyle sets the body wrapper style.
func (v *VNode) SetContentStyle(s style.Style) *VNode {
	v.contentStyle = s
	return v
}

// SetChangeIntent sets the optional custom change intent.
func (v *VNode) SetChangeIntent(i intent.Intent) *VNode {
	v.changeIntent = i
	return v
}

// SetChangeIntentField sets the field binding for active keys.
func (v *VNode) SetChangeIntentField(i intent.FieldIntent) *VNode {
	v.changeIntentField = i
	return v
}

// Items returns the configured items.
func (v *VNode) Items() []Item { return cloneItems(v.items) }

// ActiveKeys returns the declarative active keys.
func (v *VNode) ActiveKeys() []string { return cloneStrings(v.activeKeys) }

func cloneItems(items []Item) []Item {
	if len(items) == 0 {
		return nil
	}
	cloned := make([]Item, len(items))
	copy(cloned, items)
	return cloned
}

func cloneStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	cloned := make([]string, len(values))
	copy(cloned, values)
	return cloned
}

func normalizeItems(items []Item) []Item {
	if len(items) == 0 {
		return nil
	}

	cloned := cloneItems(items)
	seen := make(map[string]int, len(cloned))
	for index := range cloned {
		key := strings.TrimSpace(cloned[index].Key)
		if key == "" {
			key = fmt.Sprintf("panel-%d", index)
		}
		base := key
		if count, exists := seen[base]; exists {
			count++
			seen[base] = count
			key = fmt.Sprintf("%s-%d", base, count)
		} else {
			seen[base] = 0
		}
		cloned[index].Key = key
	}
	return cloned
}

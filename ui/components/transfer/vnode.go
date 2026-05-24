package transfer

import (
	"strings"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

const (
	defaultListHeight = 8
	defaultListWidth  = 24
	defaultValueSep   = ","

	propChangeIntent         = "changeIntent"
	propComponentID          = "componentID"
	propFormID               = "formID"
	propItems                = "items"
	propKey                  = "key"
	propListHeight           = "listHeight"
	propListWidth            = "listWidth"
	propOperations           = "operations"
	propSearchable           = "searchable"
	propSearchControlled     = "searchControlled"
	propSearchPlaceholders   = "searchPlaceholders"
	propSourceSearch         = "sourceSearch"
	propStyle                = "style"
	propTargetKeys           = "targetKeys"
	propTargetKeysControlled = "targetKeysControlled"
	propTargetSearch         = "targetSearch"
	propTitles               = "titles"
	propWidth                = "width"
)

var (
	defaultTitles             = [2]string{"Source", "Target"}
	defaultOperations         = [2]string{">", "<"}
	defaultSearchPlaceholders = [2]string{"Search source", "Search target"}
)

// Item describes a single transfer row.
type Item struct {
	Key         string
	Title       string
	Description string
	Disabled    bool
}

// VNode is the declarative description of a Transfer component.
type VNode struct {
	*rtui.ElementVNode

	key                  string
	componentID          string
	items                []Item
	titles               [2]string
	operations           [2]string
	searchable           bool
	searchControlled     bool
	searchPlaceholders   [2]string
	sourceSearch         string
	targetKeys           []string
	targetKeysControlled bool
	targetSearch         string
	listWidth            int
	listHeight           int
	width                int
	rootStyle            style.Style
	changeIntent         intent.Intent
	formID               string
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// New creates a new Transfer VNode.
func New() *VNode {
	return &VNode{
		ElementVNode:         rtui.NewElement("transfer"),
		titles:               defaultTitles,
		operations:           defaultOperations,
		searchPlaceholders:   defaultSearchPlaceholders,
		listWidth:            defaultListWidth,
		listHeight:           defaultListHeight,
		targetKeysControlled: false,
	}
}

// NewItem creates a transfer item with the provided key and title.
func NewItem(key, title string) Item {
	return Item{Key: key, Title: title}
}

func (v *VNode) Key() string { return v.key }

func (v *VNode) SetKey(key string) rtui.VNode {
	v.key = key
	return v
}

// ID returns the explicit business ID, or falls back to key.
func (v *VNode) ID() string {
	if id := v.ElementVNode.ID(); id != "" {
		return id
	}
	return v.key
}

func (v *VNode) SetID(id string) rtui.VNode {
	v.ElementVNode.SetID(id)
	return v
}

func (v *VNode) Tag() string { return "transfer" }

func (v *VNode) Style() style.Style { return v.rootStyle }

func (v *VNode) SetStyle(s style.Style) rtui.VNode {
	v.rootStyle = s
	return v
}

func (v *VNode) Children() []rtui.VNode { return nil }

func (v *VNode) SetChildren(children []rtui.VNode) rtui.VNode { return v }

func (v *VNode) GetLayer() rtui.Layer { return rtui.LayerBase }

func (v *VNode) SetLayer(l rtui.Layer) rtui.VNode { return v }

func (v *VNode) Props() rtui.Props {
	return rtui.Props{
		propChangeIntent:         v.changeIntent,
		propComponentID:          v.componentID,
		propFormID:               v.formID,
		propItems:                cloneItems(v.items),
		propKey:                  v.key,
		propListHeight:           v.listHeight,
		propListWidth:            v.listWidth,
		propOperations:           v.operations,
		propSearchable:           v.searchable,
		propSearchControlled:     v.searchControlled,
		propSearchPlaceholders:   v.searchPlaceholders,
		propSourceSearch:         v.sourceSearch,
		propStyle:                v.rootStyle,
		propTargetKeys:           append([]string(nil), v.targetKeys...),
		propTargetKeysControlled: v.targetKeysControlled,
		propTargetSearch:         v.targetSearch,
		propTitles:               v.titles,
		propWidth:                v.width,
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
		v.items = cloneItems(items)
	}
	if titles, ok := props[propTitles].([2]string); ok {
		v.titles = titles
	}
	if operations, ok := props[propOperations].([2]string); ok {
		v.operations = operations
	}
	if searchable, ok := props[propSearchable].(bool); ok {
		v.searchable = searchable
	}
	if controlled, ok := props[propSearchControlled].(bool); ok {
		v.searchControlled = controlled
	}
	if placeholders, ok := props[propSearchPlaceholders].([2]string); ok {
		v.searchPlaceholders = placeholders
	}
	if sourceSearch, ok := props[propSourceSearch].(string); ok {
		v.sourceSearch = strings.TrimSpace(sourceSearch)
	}
	if targetKeys, ok := props[propTargetKeys].([]string); ok {
		v.targetKeys = append([]string(nil), targetKeys...)
	}
	if controlled, ok := props[propTargetKeysControlled].(bool); ok {
		v.targetKeysControlled = controlled
	}
	if targetSearch, ok := props[propTargetSearch].(string); ok {
		v.targetSearch = strings.TrimSpace(targetSearch)
	}
	if listWidth, ok := props[propListWidth].(int); ok {
		v.listWidth = listWidth
	}
	if listHeight, ok := props[propListHeight].(int); ok {
		v.listHeight = listHeight
	}
	if width, ok := props[propWidth].(int); ok {
		v.width = width
	}
	if rootStyle, ok := props[propStyle].(style.Style); ok {
		v.rootStyle = rootStyle
	}
	if changeIntent, ok := props[propChangeIntent].(intent.Intent); ok {
		v.changeIntent = changeIntent
	}
	if formID, ok := props[propFormID].(string); ok {
		v.formID = formID
	}
	return v
}

func (v *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(v.Props())
}

func (v *VNode) SetComponentID(id string) *VNode {
	v.componentID = id
	return v
}

func (v *VNode) SetItems(items []Item) *VNode {
	v.items = cloneItems(items)
	return v
}

func (v *VNode) AddItem(item Item) *VNode {
	v.items = append(v.items, item)
	return v
}

func (v *VNode) SetTitles(source, target string) *VNode {
	v.titles = [2]string{source, target}
	return v
}

func (v *VNode) SetOperations(toTarget, toSource string) *VNode {
	v.operations = [2]string{toTarget, toSource}
	return v
}

func (v *VNode) SetSearchable(searchable bool) *VNode {
	v.searchable = searchable
	return v
}

func (v *VNode) SetSearchPlaceholders(source, target string) *VNode {
	v.searchPlaceholders = [2]string{source, target}
	return v
}

func (v *VNode) SetSearchValues(source, target string) *VNode {
	v.sourceSearch = strings.TrimSpace(source)
	v.targetSearch = strings.TrimSpace(target)
	v.searchControlled = true
	return v
}

func (v *VNode) SetInitialSearchValues(source, target string) *VNode {
	v.sourceSearch = strings.TrimSpace(source)
	v.targetSearch = strings.TrimSpace(target)
	v.searchControlled = false
	return v
}

func (v *VNode) SetTargetKeys(keys []string) *VNode {
	v.targetKeys = append([]string(nil), keys...)
	v.targetKeysControlled = true
	return v
}

func (v *VNode) SetInitialTargetKeys(keys []string) *VNode {
	v.targetKeys = append([]string(nil), keys...)
	v.targetKeysControlled = false
	return v
}

func (v *VNode) SetListWidth(width int) *VNode {
	v.listWidth = width
	return v
}

func (v *VNode) SetListHeight(height int) *VNode {
	v.listHeight = height
	return v
}

func (v *VNode) SetWidth(width int) *VNode {
	v.width = width
	return v
}

func (v *VNode) SetStyleProps(s style.Style) *VNode {
	v.rootStyle = s
	return v
}

func (v *VNode) SetChangeIntent(changeIntent intent.Intent) *VNode {
	v.changeIntent = changeIntent
	return v
}

func (v *VNode) SetFormID(formID string) *VNode {
	v.formID = formID
	return v
}

func cloneItems(items []Item) []Item {
	if len(items) == 0 {
		return nil
	}
	cloned := make([]Item, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		key := strings.TrimSpace(item.Key)
		if key == "" {
			key = strings.TrimSpace(item.Title)
		}
		if key == "" {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		title := strings.TrimSpace(item.Title)
		if title == "" {
			title = key
		}
		cloned = append(cloned, Item{
			Key:         key,
			Title:       title,
			Description: strings.TrimSpace(item.Description),
			Disabled:    item.Disabled,
		})
	}
	if len(cloned) == 0 {
		return nil
	}
	return cloned
}

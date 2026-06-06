package descriptions

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	emptycomp "github.com/wwsheng009/mint/ui/components/empty"
	panelcomp "github.com/wwsheng009/mint/ui/components/panel"
	textcomp "github.com/wwsheng009/mint/ui/components/text"
)

const defaultDetailPanelEmptyHintScopeWidth = 96

const (
	propBordered     = "bordered"
	propColon        = "colon"
	propColumn       = "column"
	propContentStyle = "contentStyle"
	propExtra        = "extra"
	propEmptyText    = "emptyText"
	propItems        = "items"
	propKey          = "key"
	propContentWidth = "contentWidth"
	propLabelWidth   = "labelWidth"
	propLabelStyle   = "labelStyle"
	propLayout       = "layout"
	propMaskText     = "maskText"
	propStyle        = "style"
	propTitle        = "title"
	propTitleStyle   = "titleStyle"
	propWidth        = "width"
)

// Layout controls how each description item is rendered.
type Layout int

const (
	LayoutHorizontal Layout = iota
	LayoutVertical
)

// Item describes a label/content pair in Descriptions.
type Item struct {
	Key          string
	Label        string
	Content      rtui.VNode
	Value        interface{}
	HasValue     bool
	Span         int
	LabelWidth   int
	ContentWidth int
	EmptyText    string
	Sensitive    bool
	MaskText     string
	LabelStyle   style.Style
	ContentStyle style.Style
}

// Entry creates a descriptions item from label and content.
func Entry(label string, content rtui.VNode) Item {
	return Item{
		Label:   label,
		Content: content,
		Span:    1,
	}
}

// Field creates a text descriptions item.
func Field(label, value string) Item {
	item := Entry(label, textcomp.New(value))
	item.Value = value
	item.HasValue = true
	return item
}

// Value creates a descriptions item from an arbitrary value.
func Value(label string, value interface{}) Item {
	return Item{
		Label:    label,
		Value:    value,
		HasValue: true,
		Span:     1,
	}
}

// FallbackValue creates a descriptions item and replaces nil or blank values with fallback.
func FallbackValue(label string, value interface{}, fallback string) Item {
	return Value(label, descriptionValueText(value, fallback))
}

// CompactValue creates a descriptions item with a compact display-width-bounded value.
func CompactValue(label string, value interface{}, maxWidth int) Item {
	return CompactFallbackValue(label, value, "-", maxWidth)
}

// CompactFallbackValue creates a descriptions item with fallback and display-width truncation.
func CompactFallbackValue(label string, value interface{}, fallback string, maxWidth int) Item {
	return Value(label, compactDescriptionText(descriptionValueText(value, fallback), maxWidth))
}

// CountValue creates a descriptions item for non-negative operational counts.
func CountValue(label string, count int) Item {
	return Value(label, textcomp.IntText(count))
}

// RatioValue creates a descriptions item for available/total operational counts.
func RatioValue(label string, available, total int) Item {
	return Value(label, textcomp.RatioText(available, total))
}

// MaskedValue creates a descriptions item with a partially masked value.
func MaskedValue(label string, value interface{}, visiblePrefix, visibleSuffix int) Item {
	return MaskedFallbackValue(label, value, "-", visiblePrefix, visibleSuffix)
}

// MaskedFallbackValue creates a partially masked descriptions item with fallback.
func MaskedFallbackValue(label string, value interface{}, fallback string, visiblePrefix, visibleSuffix int) Item {
	text := descriptionValueText(value, fallback)
	if strings.TrimSpace(text) == strings.TrimSpace(fallback) {
		return Value(label, text)
	}
	return Value(label, textcomp.MaskSensitive(text, visiblePrefix, visibleSuffix))
}

// StateValue creates a descriptions item with semantically colored operational state text.
func StateValue(label string, value interface{}, state string) Item {
	text := descriptionValueText(value, "-")
	state = descriptionValueText(state, text)
	return Entry(label, textcomp.State(text, state))
}

// BoolStateValue creates a descriptions item from a boolean with custom text and state semantics.
func BoolStateValue(label string, value bool, trueText, falseText, trueState, falseState string) Item {
	if value {
		return StateValue(label, descriptionValueText(trueText, "yes"), descriptionValueText(trueState, trueText))
	}
	return StateValue(label, descriptionValueText(falseText, "no"), descriptionValueText(falseState, falseText))
}

// EnabledValue creates a descriptions item for enabled/disabled boolean state.
func EnabledValue(label string, enabled bool) Item {
	return BoolStateValue(label, enabled, "enabled", "disabled", "enabled", "disabled")
}

// SensitiveField creates a masked descriptions item for secrets and tokens.
func SensitiveField(label string, value interface{}) Item {
	return Value(label, value).WithSensitive(true)
}

// Panel creates a standard titled details panel with a Descriptions body and optional actions.
func Panel(key, title string, width, labelWidth, contentWidth int, items []Item, actions ...rtui.VNode) rtui.VNode {
	details := NewBuilder().
		Key(descriptionsPanelDetailsKey(key)).
		Column(1).
		LabelWidth(labelWidth).
		ContentWidth(contentWidth).
		EmptyText("-").
		Items(items).
		Build()

	content := details
	if len(actions) > 0 {
		children := make([]rtui.VNode, 0, len(actions)+1)
		children = append(children, details)
		children = append(children, actions...)
		content = rtui.VStack(children...)
	}

	builder := panelcomp.NewBuilder().
		Title(title).
		Single().
		Width(width).
		Content(content)
	if strings.TrimSpace(key) != "" {
		builder.Key(key)
	}
	return builder.Build()
}

// ContextStripConfig describes a compact selected-object context row for details panels.
type ContextStripConfig struct {
	Key          string
	Width        int
	Column       int
	LabelWidth   int
	ContentWidth int
	Items        []Item
}

// ContextStrip creates a compact descriptions node for selected-object identity and state.
func ContextStrip(cfg ContextStripConfig) rtui.VNode {
	column := cfg.Column
	if column <= 0 {
		column = len(cfg.Items)
	}
	if column <= 0 {
		column = 1
	}
	if column > 4 {
		column = 4
	}
	key := strings.TrimSpace(cfg.Key)
	if key == "" {
		key = "descriptions.context"
	}
	return NewBuilder().
		Key(key).
		Column(column).
		Width(cfg.Width).
		LabelWidth(cfg.LabelWidth).
		ContentWidth(cfg.ContentWidth).
		EmptyText("-").
		Items(cfg.Items).
		Build()
}

// PanelWithContext creates a details panel with a compact context strip before details and actions.
func PanelWithContext(key, title string, width, labelWidth, contentWidth int, context ContextStripConfig, items []Item, actions ...rtui.VNode) rtui.VNode {
	details := NewBuilder().
		Key(descriptionsPanelDetailsKey(key)).
		Column(1).
		LabelWidth(labelWidth).
		ContentWidth(contentWidth).
		EmptyText("-").
		Items(items).
		Build()

	if strings.TrimSpace(context.Key) == "" {
		context.Key = descriptionsPanelContextKey(key)
	}
	if context.Width <= 0 {
		context.Width = width
	}

	children := make([]rtui.VNode, 0, len(actions)+2)
	children = append(children, ContextStrip(context), details)
	children = append(children, actions...)

	builder := panelcomp.NewBuilder().
		Title(title).
		Single().
		Width(width).
		Content(rtui.VStack(children...))
	if strings.TrimSpace(key) != "" {
		builder.Key(key)
	}
	return builder.Build()
}

// DetailPanelConfig describes a semantic selected-object detail panel.
type DetailPanelConfig struct {
	Key          string
	Title        string
	Width        int
	LabelWidth   int
	ContentWidth int
	EmptyWhen    bool
	EmptyText    string
	EmptyHint    string
	Context      ContextStripConfig
	Items        []Item
	Actions      []rtui.VNode
}

// DetailPanel creates a semantic details panel with a context strip, details, and actions.
func DetailPanel(cfg DetailPanelConfig) rtui.VNode {
	if cfg.EmptyWhen {
		return detailPanelEmpty(cfg)
	}
	return PanelWithContext(cfg.Key, cfg.Title, cfg.Width, cfg.LabelWidth, cfg.ContentWidth, cfg.Context, cfg.Items, cfg.Actions...)
}

// DetailPanelEmptyHint creates a compact recovery hint for an empty detail
// panel, appending a normalized "Scope: ..." summary when scope parts are
// present.
func DetailPanelEmptyHint(action string, parts ...textcomp.KeyValuePart) string {
	return DetailPanelEmptyHintWithScopeWidth(action, defaultDetailPanelEmptyHintScopeWidth, parts...)
}

// DetailPanelEmptyHintWithScopeWidth creates a compact recovery hint and limits
// the scope summary to scopeWidth display cells when scopeWidth is positive.
func DetailPanelEmptyHintWithScopeWidth(action string, scopeWidth int, parts ...textcomp.KeyValuePart) string {
	action = textcomp.FirstNonEmptyText(action)
	scopeParts := make([]textcomp.KeyValuePart, 0, len(parts))
	for _, part := range parts {
		value := textcomp.FirstNonEmptyText(part.Value)
		if value == "" {
			continue
		}
		scopeParts = append(scopeParts, textcomp.KeyValuePart{
			Label: part.Label,
			Value: value,
		})
	}

	scope := textcomp.KeyValueSummaryText(scopeParts...)
	if scope == "-" {
		scope = ""
	}
	if scope != "" && scopeWidth > 0 {
		scope = textcomp.CompactText(scope, scopeWidth)
	}
	if scope == "" {
		return action
	}
	if action == "" {
		return "Scope: " + scope
	}
	return action + " Scope: " + scope
}

func detailPanelEmpty(cfg DetailPanelConfig) rtui.VNode {
	emptyText := strings.TrimSpace(cfg.EmptyText)
	if emptyText == "" {
		emptyText = "No selection available."
	}
	children := make([]rtui.VNode, 0, len(cfg.Actions)+1)
	children = append(children, emptycomp.NewBuilder().Description(emptyText).Build())
	if hint := strings.TrimSpace(cfg.EmptyHint); hint != "" {
		children = append(children, textcomp.Subtle(hint))
	}
	children = append(children, cfg.Actions...)

	content := children[0]
	if len(children) > 1 {
		content = rtui.VStack(children...)
	}

	builder := panelcomp.NewBuilder().
		Title(cfg.Title).
		Single().
		Width(cfg.Width).
		Content(content)
	if strings.TrimSpace(cfg.Key) != "" {
		builder.Key(cfg.Key)
	}
	return builder.Build()
}

// EmptyPanel creates a standard titled details panel for unavailable diagnostics.
func EmptyPanel(key, title string, width int, emptyText string) rtui.VNode {
	if strings.TrimSpace(emptyText) == "" {
		emptyText = "details unavailable"
	}
	builder := panelcomp.NewBuilder().
		Title(title).
		Single().
		Width(width).
		Content(emptycomp.NewBuilder().Description(emptyText).Build())
	if strings.TrimSpace(key) != "" {
		builder.Key(key)
	}
	return builder.Build()
}

// WithKey sets the item key.
func (i Item) WithKey(key string) Item {
	i.Key = key
	return i
}

// WithSpan sets the column span.
func (i Item) WithSpan(span int) Item {
	i.Span = span
	return i
}

// WithLabelWidth sets a fixed label width for this item.
func (i Item) WithLabelWidth(width int) Item {
	i.LabelWidth = width
	return i
}

// WithContentWidth sets a fixed content width for this item.
func (i Item) WithContentWidth(width int) Item {
	i.ContentWidth = width
	return i
}

// WithEmptyText sets the placeholder used when this item has an empty value.
func (i Item) WithEmptyText(emptyText string) Item {
	i.EmptyText = emptyText
	return i
}

// WithSensitive toggles masking for this item.
func (i Item) WithSensitive(sensitive bool) Item {
	i.Sensitive = sensitive
	return i
}

// WithMaskText sets the masked display text for this item.
func (i Item) WithMaskText(maskText string) Item {
	i.MaskText = maskText
	return i
}

// WithLabelStyle sets the item label style.
func (i Item) WithLabelStyle(s style.Style) Item {
	i.LabelStyle = s
	return i
}

// WithContentStyle sets the item content style.
func (i Item) WithContentStyle(s style.Style) Item {
	i.ContentStyle = s
	return i
}

// VNode is the declarative description of a Descriptions component.
type VNode struct {
	*rtui.ElementVNode

	key          string
	title        string
	extra        rtui.VNode
	items        []Item
	column       int
	bordered     bool
	colon        bool
	layout       Layout
	width        int
	labelWidth   int
	contentWidth int
	emptyText    string
	maskText     string
	rootStyle    style.Style
	titleStyle   style.Style
	labelStyle   style.Style
	contentStyle style.Style
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// New creates a new Descriptions VNode.
func New(items []Item) *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("descriptions"),
		items:        normalizeItems(items),
		column:       3,
		bordered:     false,
		colon:        true,
		layout:       LayoutHorizontal,
		emptyText:    "-",
		maskText:     "****",
	}
}

func (v *VNode) Key() string { return v.key }

func (v *VNode) SetKey(key string) rtui.VNode {
	v.key = key
	return v
}

func (v *VNode) Tag() string { return "descriptions" }

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
		propBordered:     v.bordered,
		propColon:        v.colon,
		propColumn:       v.column,
		propContentStyle: v.contentStyle,
		propContentWidth: v.contentWidth,
		propEmptyText:    v.emptyText,
		propExtra:        v.extra,
		propItems:        cloneItems(v.items),
		propKey:          v.key,
		propLabelWidth:   v.labelWidth,
		propLabelStyle:   v.labelStyle,
		propLayout:       v.layout,
		propMaskText:     v.maskText,
		propStyle:        v.rootStyle,
		propTitle:        v.title,
		propTitleStyle:   v.titleStyle,
		propWidth:        v.width,
	}
}

func (v *VNode) SetProps(props rtui.Props) rtui.VNode {
	if key, ok := props[propKey].(string); ok {
		v.key = key
	}
	if title, ok := props[propTitle].(string); ok {
		v.title = title
	}
	if extra, ok := props[propExtra].(rtui.VNode); ok {
		v.extra = extra
	}
	if items, ok := props[propItems].([]Item); ok {
		v.items = normalizeItems(items)
	}
	if column, ok := props[propColumn].(int); ok {
		v.column = column
	}
	if bordered, ok := props[propBordered].(bool); ok {
		v.bordered = bordered
	}
	if colon, ok := props[propColon].(bool); ok {
		v.colon = colon
	}
	if layout, ok := props[propLayout].(Layout); ok {
		v.layout = layout
	}
	if width, ok := props[propWidth].(int); ok {
		v.width = width
	}
	if labelWidth, ok := props[propLabelWidth].(int); ok {
		v.labelWidth = labelWidth
	}
	if contentWidth, ok := props[propContentWidth].(int); ok {
		v.contentWidth = contentWidth
	}
	if emptyText, ok := props[propEmptyText].(string); ok {
		v.emptyText = emptyText
	}
	if maskText, ok := props[propMaskText].(string); ok {
		v.maskText = maskText
	}
	if s, ok := props[propStyle].(style.Style); ok {
		v.rootStyle = s
	}
	if s, ok := props[propTitleStyle].(style.Style); ok {
		v.titleStyle = s
	}
	if s, ok := props[propLabelStyle].(style.Style); ok {
		v.labelStyle = s
	}
	if s, ok := props[propContentStyle].(style.Style); ok {
		v.contentStyle = s
	}
	return v
}

func (v *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(v.Props())
}

// SetTitle sets the optional title text.
func (v *VNode) SetTitle(title string) *VNode {
	v.title = title
	return v
}

// SetExtra sets the optional header-side node.
func (v *VNode) SetExtra(extra rtui.VNode) *VNode {
	v.extra = extra
	return v
}

// SetItems replaces all items.
func (v *VNode) SetItems(items []Item) *VNode {
	v.items = normalizeItems(items)
	return v
}

// AddItem appends an item.
func (v *VNode) AddItem(item Item) *VNode {
	v.items = normalizeItems(append(v.items, item))
	return v
}

// SetColumn sets the target column count.
func (v *VNode) SetColumn(column int) *VNode {
	v.column = column
	return v
}

// SetBordered toggles the bordered presentation.
func (v *VNode) SetBordered(bordered bool) *VNode {
	v.bordered = bordered
	return v
}

// SetColon toggles label colon rendering.
func (v *VNode) SetColon(colon bool) *VNode {
	v.colon = colon
	return v
}

// SetLayout sets the item layout mode.
func (v *VNode) SetLayout(layout Layout) *VNode {
	v.layout = layout
	return v
}

// SetWidth sets a preferred width.
func (v *VNode) SetWidth(width int) *VNode {
	v.width = width
	return v
}

// SetLabelWidth sets the default fixed label width.
func (v *VNode) SetLabelWidth(width int) *VNode {
	v.labelWidth = width
	return v
}

// SetContentWidth sets the default fixed content width.
func (v *VNode) SetContentWidth(width int) *VNode {
	v.contentWidth = width
	return v
}

// SetEmptyText sets the placeholder for empty values.
func (v *VNode) SetEmptyText(emptyText string) *VNode {
	v.emptyText = emptyText
	return v
}

// SetMaskText sets the default masked text for sensitive values.
func (v *VNode) SetMaskText(maskText string) *VNode {
	v.maskText = maskText
	return v
}

// SetTitleStyle sets the title style.
func (v *VNode) SetTitleStyle(s style.Style) *VNode {
	v.titleStyle = s
	return v
}

// SetLabelStyle sets the default label style.
func (v *VNode) SetLabelStyle(s style.Style) *VNode {
	v.labelStyle = s
	return v
}

// SetContentStyle sets the default content style.
func (v *VNode) SetContentStyle(s style.Style) *VNode {
	v.contentStyle = s
	return v
}

// Items returns the configured items.
func (v *VNode) Items() []Item { return cloneItems(v.items) }

func cloneItems(items []Item) []Item {
	if len(items) == 0 {
		return nil
	}
	cloned := make([]Item, len(items))
	copy(cloned, items)
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
			key = fmt.Sprintf("item-%d", index)
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
		if cloned[index].Span < 1 {
			cloned[index].Span = 1
		}
		if cloned[index].LabelWidth < 0 {
			cloned[index].LabelWidth = 0
		}
		if cloned[index].ContentWidth < 0 {
			cloned[index].ContentWidth = 0
		}
	}
	return cloned
}

func descriptionsPanelDetailsKey(key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return "descriptions-panel.details"
	}
	return key + ".details"
}

func descriptionsPanelContextKey(key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return "descriptions-panel.context"
	}
	return key + ".context"
}

func descriptionValueText(value interface{}, fallback string) string {
	fallback = strings.TrimSpace(fallback)
	if fallback == "" {
		fallback = "-"
	}
	if value == nil {
		return fallback
	}
	text := strings.TrimSpace(fmt.Sprint(value))
	if text == "" {
		return fallback
	}
	return text
}

func compactDescriptionText(text string, maxWidth int) string {
	if maxWidth <= 0 || paint.StringWidth(text) <= maxWidth {
		return text
	}
	if maxWidth <= 3 {
		return trimDescriptionText(text, maxWidth)
	}
	prefix := strings.TrimRight(trimDescriptionText(text, maxWidth-3), " ")
	if prefix == "" {
		return "..."
	}
	return prefix + "..."
}

func trimDescriptionText(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	var builder strings.Builder
	width := 0
	for _, r := range text {
		runeWidth := paint.RuneWidth(r)
		if runeWidth <= 0 {
			continue
		}
		if width+runeWidth > maxWidth {
			break
		}
		builder.WriteRune(r)
		width += runeWidth
	}
	return builder.String()
}

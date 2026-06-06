// Package filterbar provides a Fiber-first filter toolbar for data pages.
package filterbar

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/button"
	selectcomp "github.com/wwsheng009/mint/ui/components/select"
)

const (
	propActions    = "actions"
	propFields     = "fields"
	propGap        = "gap"
	propKey        = "key"
	propLabelWidth = "labelWidth"
	propRowGap     = "rowGap"
	propRootGap    = "rootGap"
	propStyle      = "style"
	propSummary    = "summary"
	propTitle      = "title"
	propWidth      = "width"
	propWrap       = "wrap"
)

// FieldKind controls which control a filter field renders.
type FieldKind string

const (
	FieldText   FieldKind = "text"
	FieldSearch FieldKind = "search"
	FieldSelect FieldKind = "select"
	FieldCustom FieldKind = "custom"
)

// Option aliases Select options for filter fields.
type Option = selectcomp.Option

// Field describes one filter control.
type Field struct {
	Key              string
	Label            string
	Kind             FieldKind
	Value            string
	Placeholder      string
	Width            int
	LabelWidth       int
	Options          []Option
	SelectedIndex    int
	HasSelectedIndex bool
	FieldName        string
	ChangeIntent     intent.Intent
	SubmitIntent     intent.Intent
	Disabled         bool
	Custom           rtui.VNode
}

// Action describes a command button in the filter bar.
type Action struct {
	Key            string
	Label          string
	PressIntent    intent.Intent
	Variant        button.Variant
	Disabled       bool
	DisabledReason string
	Width          int
}

// VNode is the declarative description of a FilterBar.
type VNode struct {
	*rtui.ElementVNode

	key        string
	title      string
	summary    string
	fields     []Field
	actions    []Action
	width      int
	gap        int
	rowGap     int
	rootGap    int
	wrap       bool
	labelWidth int
	rootStyle  style.Style
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// New creates a FilterBar VNode.
func New() *VNode {
	return &VNode{
		ElementVNode: rtui.NewElement("filterbar"),
		gap:          1,
		rowGap:       1,
		rootGap:      0,
		wrap:         false,
		labelWidth:   0,
	}
}

func (v *VNode) Key() string { return v.key }

func (v *VNode) SetKey(key string) rtui.VNode {
	v.key = key
	return v
}

// ID returns the explicit business ID, or falls back to the key.
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

func (v *VNode) Tag() string { return "filterbar" }

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
		propActions:    cloneActions(v.actions),
		propFields:     cloneFields(v.fields),
		propGap:        v.gap,
		propKey:        v.key,
		propLabelWidth: v.labelWidth,
		propRowGap:     v.rowGap,
		propRootGap:    v.rootGap,
		propStyle:      v.rootStyle,
		propSummary:    v.summary,
		propTitle:      v.title,
		propWidth:      v.width,
		propWrap:       v.wrap,
	}
}

func (v *VNode) SetProps(props rtui.Props) rtui.VNode {
	v.key = getStringProp(props, propKey, v.key)
	v.title = normalizeInlineText(getStringProp(props, propTitle, v.title))
	v.summary = normalizeInlineText(getStringProp(props, propSummary, v.summary))
	v.fields = normalizeFields(getFieldsProp(props, v.fields))
	v.actions = normalizeActions(getActionsProp(props, v.actions))
	v.width = getIntProp(props, propWidth, v.width)
	v.gap = getIntProp(props, propGap, v.gap)
	v.rowGap = getIntProp(props, propRowGap, v.rowGap)
	v.rootGap = getIntProp(props, propRootGap, v.rootGap)
	v.wrap = getBoolProp(props, propWrap, v.wrap)
	v.labelWidth = getIntProp(props, propLabelWidth, v.labelWidth)
	v.rootStyle = getStyleProp(props, propStyle, v.rootStyle)
	v.normalize()
	return v
}

func (v *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(v.Props())
}

func (v *VNode) SetTitle(title string) *VNode {
	v.title = normalizeInlineText(title)
	return v
}

func (v *VNode) SetSummary(summary string) *VNode {
	v.summary = normalizeInlineText(summary)
	return v
}

func (v *VNode) SetFields(fields []Field) *VNode {
	v.fields = normalizeFields(fields)
	return v
}

func (v *VNode) AddField(field Field) *VNode {
	v.fields = normalizeFields(append(v.fields, field))
	return v
}

func (v *VNode) SetActions(actions []Action) *VNode {
	v.actions = normalizeActions(actions)
	return v
}

func (v *VNode) AddAction(action Action) *VNode {
	v.actions = normalizeActions(append(v.actions, action))
	return v
}

func (v *VNode) SetWidth(width int) *VNode {
	v.width = width
	v.normalize()
	return v
}

func (v *VNode) SetGap(gap int) *VNode {
	v.gap = gap
	v.normalize()
	return v
}

func (v *VNode) SetRowGap(rowGap int) *VNode {
	v.rowGap = rowGap
	v.normalize()
	return v
}

func (v *VNode) SetRootGap(rootGap int) *VNode {
	v.rootGap = rootGap
	v.normalize()
	return v
}

func (v *VNode) SetWrap(wrap bool) *VNode {
	v.wrap = wrap
	return v
}

func (v *VNode) SetLabelWidth(width int) *VNode {
	v.labelWidth = width
	v.normalize()
	return v
}

func (v *VNode) SetRootStyle(s style.Style) *VNode {
	v.rootStyle = s
	return v
}

func (v *VNode) Fields() []Field { return cloneFields(v.fields) }

func (v *VNode) Actions() []Action { return cloneActions(v.actions) }

func (v *VNode) normalize() {
	v.fields = normalizeFields(v.fields)
	v.actions = normalizeActions(v.actions)
	if v.width < 0 {
		v.width = 0
	}
	if v.gap < 0 {
		v.gap = 0
	}
	if v.rowGap < 0 {
		v.rowGap = 0
	}
	if v.rootGap < 0 {
		v.rootGap = 0
	}
	if v.labelWidth < 0 {
		v.labelWidth = 0
	}
}

// Search creates a search field.
func Search(key, label, value string) Field {
	return Field{Key: key, Label: label, Kind: FieldSearch, Value: value}
}

// Text creates a text filter field.
func Text(key, label, value string) Field {
	return Field{Key: key, Label: label, Kind: FieldText, Value: value}
}

// Select creates a select filter field.
func Select(key, label string, options []Option) Field {
	return Field{Key: key, Label: label, Kind: FieldSelect, Options: cloneOptions(options), SelectedIndex: -1}
}

// Custom creates a custom filter field.
func Custom(key, label string, node rtui.VNode) Field {
	return Field{Key: key, Label: label, Kind: FieldCustom, Custom: node}
}

func (f Field) WithKey(key string) Field {
	f.Key = key
	return f
}

func (f Field) WithLabel(label string) Field {
	f.Label = label
	return f
}

func (f Field) WithPlaceholder(placeholder string) Field {
	f.Placeholder = placeholder
	return f
}

func (f Field) WithWidth(width int) Field {
	f.Width = width
	return f
}

func (f Field) WithLabelWidth(width int) Field {
	f.LabelWidth = width
	return f
}

func (f Field) WithOptions(options []Option) Field {
	f.Options = cloneOptions(options)
	return f
}

func (f Field) WithSelectedIndex(index int) Field {
	f.SelectedIndex = index
	f.HasSelectedIndex = true
	return f
}

func (f Field) ForField(fieldName string) Field {
	f.FieldName = fieldName
	return f
}

func (f Field) OnChange(changeIntent intent.Intent) Field {
	f.ChangeIntent = changeIntent
	return f
}

func (f Field) OnSubmit(submitIntent intent.Intent) Field {
	f.SubmitIntent = submitIntent
	return f
}

func (f Field) WithDisabled(disabled bool) Field {
	f.Disabled = disabled
	return f
}

// Button creates an action button.
func Button(key, label string, pressIntent intent.Intent) Action {
	return Action{Key: key, Label: label, PressIntent: pressIntent}
}

// RefreshAction creates a primary refresh action with a standard loading disabled reason.
func RefreshAction(pressIntent intent.Intent, busy bool, loadingReason string) Action {
	action := Button("refresh", "Refresh", pressIntent).Primary()
	if busy {
		return action.WithDisabledReason(defaultLoadingReason(loadingReason))
	}
	return action
}

// ResetAction creates a secondary action that resets filters through the
// application-owned reducer.
func ResetAction(pressIntent intent.Intent, busy bool, loadingReason string) Action {
	action := Button("reset", "Reset", pressIntent).Secondary()
	if busy {
		return action.WithDisabledReason(defaultLoadingReason(loadingReason))
	}
	return action
}

// ResetActionWhenChanged creates a secondary Reset action that is disabled when
// the current filter state already matches the application defaults.
func ResetActionWhenChanged(pressIntent intent.Intent, changed bool, busy bool, loadingReason string) Action {
	action := ResetAction(pressIntent, busy, loadingReason)
	if busy || changed {
		return action
	}
	return action.WithDisabledReason("Nothing to reset.")
}

// ClearFieldAction creates a secondary action that clears a bound field.
func ClearFieldAction(fieldName, value string, busy bool, loadingReason string) Action {
	return ClearFieldActionWithLabel("clear", "Clear", fieldName, value, busy, loadingReason)
}

// ClearFieldActionWithLabel creates a secondary action that clears a bound
// field using caller-provided key and label. It is useful when a multi-field
// FilterBar needs separate clear actions for multiple search fields.
func ClearFieldActionWithLabel(key, label, fieldName, value string, busy bool, loadingReason string) Action {
	key = strings.TrimSpace(key)
	if key == "" {
		key = "clear"
	}
	label = strings.TrimSpace(label)
	if label == "" {
		label = "Clear"
	}
	fieldName = strings.TrimSpace(fieldName)
	action := Button(key, label, intent.FieldChangeIntent{Field: fieldName, Value: ""}).Secondary()
	if busy {
		return action.WithDisabledReason(defaultLoadingReason(loadingReason))
	}
	if fieldName == "" {
		return action.WithDisabledReason("No field is bound.")
	}
	if strings.TrimSpace(value) == "" {
		return action.WithDisabledReason("Nothing to clear.")
	}
	return action
}

// SearchRefresh creates a standard data-page filter bar with one search field and a refresh action.
func SearchRefresh(key, summary string, width, labelWidth int, search Field, refreshIntent intent.Intent, busy bool, loadingReason string) rtui.VNode {
	return NewBuilder().
		Key(key).
		Summary(summary).
		Width(width).
		LabelWidth(labelWidth).
		Field(search).
		Action(RefreshAction(refreshIntent, busy, loadingReason)).
		Build()
}

// SearchRefreshClear creates a standard one-search filter bar with Refresh and Clear actions.
func SearchRefreshClear(key, summary string, width, labelWidth int, search Field, refreshIntent intent.Intent, busy bool, loadingReason string) rtui.VNode {
	return NewBuilder().
		Key(key).
		Summary(summary).
		Width(width).
		LabelWidth(labelWidth).
		Field(search).
		Action(RefreshAction(refreshIntent, busy, loadingReason)).
		Action(ClearFieldAction(search.FieldName, search.Value, busy, loadingReason)).
		Build()
}

// SearchActions creates a standard one-search filter bar with contextual actions.
func SearchActions(key, summary string, width, labelWidth int, search Field, actions ...Action) rtui.VNode {
	return NewBuilder().
		Key(key).
		Summary(summary).
		Width(width).
		LabelWidth(labelWidth).
		Field(search).
		Actions(actions).
		Build()
}

// FieldsRefresh creates a standard multi-field data-page filter bar with Refresh.
func FieldsRefresh(key, summary string, width, labelWidth int, fields []Field, refreshIntent intent.Intent, busy bool, loadingReason string) rtui.VNode {
	return NewBuilder().
		Key(key).
		Summary(summary).
		Width(width).
		LabelWidth(labelWidth).
		Wrap(true).
		Fields(fields).
		Action(RefreshAction(refreshIntent, busy, loadingReason)).
		Build()
}

// FieldsRefreshReset creates a standard multi-field data-page filter bar with
// Refresh and Reset actions. The reset intent is owned by the application so it
// can restore business defaults such as status/source values and page state.
func FieldsRefreshReset(key, summary string, width, labelWidth int, fields []Field, refreshIntent, resetIntent intent.Intent, busy bool, loadingReason string) rtui.VNode {
	return NewBuilder().
		Key(key).
		Summary(summary).
		Width(width).
		LabelWidth(labelWidth).
		Wrap(true).
		Fields(fields).
		Action(RefreshAction(refreshIntent, busy, loadingReason)).
		Action(ResetAction(resetIntent, busy, loadingReason)).
		Build()
}

// FieldsRefreshResetWhenChanged creates a standard multi-field data-page filter
// bar with Refresh and a Reset action that is disabled until filters differ
// from the application-defined defaults.
func FieldsRefreshResetWhenChanged(key, summary string, width, labelWidth int, fields []Field, refreshIntent, resetIntent intent.Intent, resetChanged bool, busy bool, loadingReason string) rtui.VNode {
	return NewBuilder().
		Key(key).
		Summary(summary).
		Width(width).
		LabelWidth(labelWidth).
		Wrap(true).
		Fields(fields).
		Action(RefreshAction(refreshIntent, busy, loadingReason)).
		Action(ResetActionWhenChanged(resetIntent, resetChanged, busy, loadingReason)).
		Build()
}

func (a Action) WithVariant(variant button.Variant) Action {
	a.Variant = variant
	return a
}

func (a Action) Primary() Action {
	a.Variant = button.VariantPrimary
	return a
}

func (a Action) Secondary() Action {
	a.Variant = button.VariantSecondary
	return a
}

func (a Action) Danger() Action {
	a.Variant = button.VariantDanger
	return a
}

func (a Action) WithDisabled(disabled bool) Action {
	a.Disabled = disabled
	return a
}

func (a Action) WithDisabledReason(reason string) Action {
	a.Disabled = true
	a.DisabledReason = normalizeInlineText(reason)
	return a
}

// WithDisabledReasonIf disables the action with reason when condition is true.
// When the action already has a disabled reason, the new reason is appended so
// chained condition checks expose all matching prerequisites.
func (a Action) WithDisabledReasonIf(condition bool, reason string) Action {
	if !condition {
		return a
	}
	if normalizeInlineText(a.DisabledReason) != "" {
		return a.WithDisabledReason(JoinDisabledReasons(a.DisabledReason, reason))
	}
	return a.WithDisabledReason(reason)
}

// WithDisabledReasons disables the action when at least one non-empty reason is
// supplied and joins all reasons into one concise explanation.
func (a Action) WithDisabledReasons(reasons ...string) Action {
	joined := JoinDisabledReasons(reasons...)
	if joined == "" {
		return a
	}
	return a.WithDisabledReason(joined)
}

// JoinDisabledReasons normalizes, filters, and joins multiple disabled reasons.
func JoinDisabledReasons(reasons ...string) string {
	parts := make([]string, 0, len(reasons))
	for _, reason := range reasons {
		reason = normalizeInlineText(reason)
		if reason == "" {
			continue
		}
		parts = append(parts, reason)
	}
	return strings.Join(parts, " ")
}

func (a Action) WithWidth(width int) Action {
	a.Width = width
	return a
}

func normalizeFields(fields []Field) []Field {
	if len(fields) == 0 {
		return nil
	}
	normalized := cloneFields(fields)
	seen := make(map[string]int, len(normalized))
	for index := range normalized {
		key := strings.TrimSpace(normalized[index].Key)
		if key == "" {
			key = fmt.Sprintf("field-%d", index)
		}
		base := key
		if count, exists := seen[base]; exists {
			count++
			seen[base] = count
			key = fmt.Sprintf("%s-%d", base, count)
		} else {
			seen[base] = 0
		}
		normalized[index].Key = key
		if normalized[index].Kind == "" {
			normalized[index].Kind = FieldText
		}
		if normalized[index].Kind != FieldSearch &&
			normalized[index].Kind != FieldSelect &&
			normalized[index].Kind != FieldCustom {
			normalized[index].Kind = FieldText
		}
		if normalized[index].Width < 0 {
			normalized[index].Width = 0
		}
		if normalized[index].LabelWidth < 0 {
			normalized[index].LabelWidth = 0
		}
		normalized[index].Options = cloneOptions(normalized[index].Options)
		if !normalized[index].HasSelectedIndex && normalized[index].Kind == FieldSelect {
			normalized[index].SelectedIndex = -1
		}
	}
	return normalized
}

func normalizeActions(actions []Action) []Action {
	if len(actions) == 0 {
		return nil
	}
	normalized := cloneActions(actions)
	seen := make(map[string]int, len(normalized))
	for index := range normalized {
		key := strings.TrimSpace(normalized[index].Key)
		if key == "" {
			key = fmt.Sprintf("action-%d", index)
		}
		base := key
		if count, exists := seen[base]; exists {
			count++
			seen[base] = count
			key = fmt.Sprintf("%s-%d", base, count)
		} else {
			seen[base] = 0
		}
		normalized[index].Key = key
		if normalized[index].Width < 0 {
			normalized[index].Width = 0
		}
		normalized[index].DisabledReason = normalizeInlineText(normalized[index].DisabledReason)
	}
	return normalized
}

func cloneFields(fields []Field) []Field {
	if len(fields) == 0 {
		return nil
	}
	cloned := make([]Field, len(fields))
	copy(cloned, fields)
	for index := range cloned {
		cloned[index].Options = cloneOptions(cloned[index].Options)
	}
	return cloned
}

func cloneActions(actions []Action) []Action {
	if len(actions) == 0 {
		return nil
	}
	cloned := make([]Action, len(actions))
	copy(cloned, actions)
	return cloned
}

func cloneOptions(options []Option) []Option {
	if len(options) == 0 {
		return nil
	}
	cloned := make([]Option, len(options))
	copy(cloned, options)
	return cloned
}

func normalizeInlineText(content string) string {
	return strings.TrimSpace(strings.NewReplacer("\r\n", " ", "\n", " ", "\r", " ", "\t", " ").Replace(content))
}

func defaultLoadingReason(reason string) string {
	reason = normalizeInlineText(reason)
	if reason == "" {
		return "Data is loading."
	}
	return reason
}

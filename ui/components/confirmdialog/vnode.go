// Package confirmdialog provides an operation-grade confirmation dialog.
package confirmdialog

import (
	"fmt"
	"strings"

	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/style"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/ui/components/button"
)

const (
	propCancelIntent      = "cancelIntent"
	propCancelText        = "cancelText"
	propConfirmIntent     = "confirmIntent"
	propConfirmText       = "confirmText"
	propConfirmVariant    = "confirmVariant"
	propDisableConfirm    = "disableConfirm"
	propDisabledReason    = "disabledReason"
	propHeight            = "height"
	propKey               = "key"
	propMessage           = "message"
	propOpen              = "open"
	propReasonField       = "reasonField"
	propReasonLabel       = "reasonLabel"
	propReasonPlaceholder = "reasonPlaceholder"
	propReasonRequired    = "reasonRequired"
	propReasonValue       = "reasonValue"
	propStyle             = "style"
	propTargetItems       = "targetItems"
	propTitle             = "title"
	propWarning           = "warning"
	propWidth             = "width"
)

// TargetItem describes one target summary field.
type TargetItem struct {
	Key       string
	Label     string
	Value     string
	Sensitive bool
}

// VNode is the declarative description of a ConfirmDialog.
type VNode struct {
	*rtui.ElementVNode

	key               string
	title             string
	message           string
	warning           string
	open              bool
	width             int
	height            int
	targetItems       []TargetItem
	reasonLabel       string
	reasonValue       string
	reasonField       string
	reasonPlaceholder string
	reasonRequired    bool
	confirmText       string
	cancelText        string
	confirmVariant    button.Variant
	disableConfirm    bool
	disabledReason    string
	confirmIntent     intent.Intent
	cancelIntent      intent.Intent
	rootStyle         style.Style
}

var (
	_ rtui.VNode           = (*VNode)(nil)
	_ rtui.InstanceFactory = (*VNode)(nil)
)

// New creates a ConfirmDialog VNode.
func New() *VNode {
	return &VNode{
		ElementVNode:      rtui.NewElement("confirmdialog"),
		width:             68,
		height:            18,
		reasonLabel:       "Reason",
		reasonPlaceholder: "Audit reason",
		confirmText:       "Confirm",
		cancelText:        "Cancel",
		confirmVariant:    button.VariantDanger,
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

func (v *VNode) Tag() string { return "confirmdialog" }

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
		propCancelIntent:      v.cancelIntent,
		propCancelText:        v.cancelText,
		propConfirmIntent:     v.confirmIntent,
		propConfirmText:       v.confirmText,
		propConfirmVariant:    v.confirmVariant,
		propDisableConfirm:    v.disableConfirm,
		propDisabledReason:    v.disabledReason,
		propHeight:            v.height,
		propKey:               v.key,
		propMessage:           v.message,
		propOpen:              v.open,
		propReasonField:       v.reasonField,
		propReasonLabel:       v.reasonLabel,
		propReasonPlaceholder: v.reasonPlaceholder,
		propReasonRequired:    v.reasonRequired,
		propReasonValue:       v.reasonValue,
		propStyle:             v.rootStyle,
		propTargetItems:       cloneTargetItems(v.targetItems),
		propTitle:             v.title,
		propWarning:           v.warning,
		propWidth:             v.width,
	}
}

func (v *VNode) SetProps(props rtui.Props) rtui.VNode {
	v.key = getStringProp(props, propKey, v.key)
	v.title = getStringProp(props, propTitle, v.title)
	v.message = getStringProp(props, propMessage, v.message)
	v.warning = getStringProp(props, propWarning, v.warning)
	v.open = getBoolProp(props, propOpen, v.open)
	v.width = getIntProp(props, propWidth, v.width)
	v.height = getIntProp(props, propHeight, v.height)
	v.targetItems = normalizeTargetItems(getTargetItemsProp(props, propTargetItems, v.targetItems))
	v.reasonLabel = getStringProp(props, propReasonLabel, v.reasonLabel)
	v.reasonValue = getStringProp(props, propReasonValue, v.reasonValue)
	v.reasonField = getStringProp(props, propReasonField, v.reasonField)
	v.reasonPlaceholder = getStringProp(props, propReasonPlaceholder, v.reasonPlaceholder)
	v.reasonRequired = getBoolProp(props, propReasonRequired, v.reasonRequired)
	v.confirmText = getStringProp(props, propConfirmText, v.confirmText)
	v.cancelText = getStringProp(props, propCancelText, v.cancelText)
	v.confirmVariant = getButtonVariantProp(props, propConfirmVariant, v.confirmVariant)
	v.disableConfirm = getBoolProp(props, propDisableConfirm, v.disableConfirm)
	v.disabledReason = getStringProp(props, propDisabledReason, v.disabledReason)
	v.confirmIntent = getIntentProp(props, propConfirmIntent, v.confirmIntent)
	v.cancelIntent = getIntentProp(props, propCancelIntent, v.cancelIntent)
	v.rootStyle = getStyleProp(props, propStyle, v.rootStyle)
	v.normalize()
	return v
}

func (v *VNode) CreateInstance() rtui.ComponentInstance {
	return NewInstance(v.Props())
}

func (v *VNode) SetTitle(title string) *VNode {
	v.title = title
	return v
}

func (v *VNode) SetMessage(message string) *VNode {
	v.message = message
	return v
}

func (v *VNode) SetWarning(warning string) *VNode {
	v.warning = warning
	return v
}

func (v *VNode) SetOpen(open bool) *VNode {
	v.open = open
	return v
}

func (v *VNode) SetWidth(width int) *VNode {
	v.width = width
	v.normalize()
	return v
}

func (v *VNode) SetHeight(height int) *VNode {
	v.height = height
	v.normalize()
	return v
}

func (v *VNode) SetTargetItems(items []TargetItem) *VNode {
	v.targetItems = normalizeTargetItems(items)
	return v
}

func (v *VNode) AddTargetItem(item TargetItem) *VNode {
	v.targetItems = normalizeTargetItems(append(v.targetItems, item))
	return v
}

func (v *VNode) SetReasonLabel(label string) *VNode {
	v.reasonLabel = label
	return v
}

func (v *VNode) SetReasonValue(value string) *VNode {
	v.reasonValue = value
	return v
}

func (v *VNode) SetReasonField(field string) *VNode {
	v.reasonField = field
	return v
}

func (v *VNode) SetReasonPlaceholder(placeholder string) *VNode {
	v.reasonPlaceholder = placeholder
	return v
}

func (v *VNode) SetReasonRequired(required bool) *VNode {
	v.reasonRequired = required
	return v
}

func (v *VNode) SetConfirmText(text string) *VNode {
	v.confirmText = text
	return v
}

func (v *VNode) SetCancelText(text string) *VNode {
	v.cancelText = text
	return v
}

func (v *VNode) SetConfirmVariant(variant button.Variant) *VNode {
	v.confirmVariant = variant
	return v
}

func (v *VNode) SetDisableConfirm(disabled bool) *VNode {
	v.disableConfirm = disabled
	return v
}

func (v *VNode) SetDisabledReason(reason string) *VNode {
	v.disabledReason = reason
	return v
}

func (v *VNode) SetConfirmIntent(i intent.Intent) *VNode {
	v.confirmIntent = i
	return v
}

func (v *VNode) SetCancelIntent(i intent.Intent) *VNode {
	v.cancelIntent = i
	return v
}

func (v *VNode) SetRootStyle(s style.Style) *VNode {
	v.rootStyle = s
	return v
}

func (v *VNode) normalize() {
	v.targetItems = normalizeTargetItems(v.targetItems)
	if v.width < 0 {
		v.width = 0
	}
	if v.height < 0 {
		v.height = 0
	}
}

// Target creates a target summary item.
func Target(key, label, value string) TargetItem {
	return TargetItem{Key: key, Label: label, Value: value}
}

// SensitiveTarget creates a masked target summary item.
func SensitiveTarget(key, label, value string) TargetItem {
	return TargetItem{Key: key, Label: label, Value: value, Sensitive: true}
}

func (i TargetItem) WithKey(key string) TargetItem {
	i.Key = key
	return i
}

func (i TargetItem) WithLabel(label string) TargetItem {
	i.Label = label
	return i
}

func (i TargetItem) WithValue(value string) TargetItem {
	i.Value = value
	return i
}

func (i TargetItem) WithSensitive(sensitive bool) TargetItem {
	i.Sensitive = sensitive
	return i
}

func normalizeTargetItems(items []TargetItem) []TargetItem {
	if len(items) == 0 {
		return nil
	}
	normalized := cloneTargetItems(items)
	seen := make(map[string]int, len(normalized))
	for index := range normalized {
		key := strings.TrimSpace(normalized[index].Key)
		if key == "" {
			key = fmt.Sprintf("target-%d", index)
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
		normalized[index].Label = normalizeConfirmText(normalized[index].Label)
		normalized[index].Value = normalizeConfirmText(normalized[index].Value)
	}
	return normalized
}

func cloneTargetItems(items []TargetItem) []TargetItem {
	if len(items) == 0 {
		return nil
	}
	cloned := make([]TargetItem, len(items))
	copy(cloned, items)
	return cloned
}

func normalizeConfirmText(content string) string {
	return strings.NewReplacer("\r\n", " ", "\n", " ", "\r", " ", "\t", " ").Replace(content)
}

package form

import (
	"reflect"

	"github.com/wwsheng009/mint/framework/theme"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	textcomp "github.com/wwsheng009/mint/ui/components/text"
	"github.com/wwsheng009/mint/ui/components/validation"
)

const (
	itemPropMarker     = "_formItem"
	itemPropChild      = "child"
	itemPropField      = "field"
	itemPropFormID     = "formID"
	itemPropLabel      = "itemLabel"
	itemPropLayout     = "itemLayout"
	itemPropValidators = "validators"
)

type itemModel struct {
	key        string
	field      string
	formID     string
	label      string
	layout     FormLayout
	child      rtui.VNode
	validators []validation.Validator
}

type itemRuntimeState struct {
	sourceID    string
	formID      string
	field       string
	validators  []validation.Validator
	unsubscribe func()
	retryQueued bool
}

// ItemBuilder provides a fluent API for building FormItem wrappers.
type ItemBuilder struct {
	node *rtui.ComponentVNode
}

// NewItem creates a FormItem wrapper around a field component.
func NewItem(field string, child rtui.VNode) *ItemBuilder {
	component := rtui.NewComponentWithProps("FormItem", renderFormItem)
	component.SetKey(field)
	component.SetProps(rtui.Props{
		itemPropMarker: true,
		itemPropField:  field,
		itemPropChild:  child,
	})
	return &ItemBuilder{node: component}
}

// Item creates a FormItem builder.
func Item(field string, child rtui.VNode) *ItemBuilder {
	return NewItem(field, child)
}

// Key sets the wrapper key.
func (b *ItemBuilder) Key(key string) *ItemBuilder {
	b.node.SetKey(key)
	return b
}

// Label sets the field label shown by the wrapper.
func (b *ItemBuilder) Label(label string) *ItemBuilder {
	props := cloneItemProps(b.node)
	props[itemPropLabel] = label
	b.node.SetProps(props)
	return b
}

// ForForm explicitly associates the item with a form.
func (b *ItemBuilder) ForForm(formID string) *ItemBuilder {
	props := cloneItemProps(b.node)
	props[itemPropFormID] = formID
	b.node.SetProps(props)
	return b
}

// Layout overrides the parent form layout for this item.
func (b *ItemBuilder) Layout(layout FormLayout) *ItemBuilder {
	props := cloneItemProps(b.node)
	props[itemPropLayout] = layout
	b.node.SetProps(props)
	return b
}

// Validators registers validators for the wrapped field.
func (b *ItemBuilder) Validators(validators ...validation.Validator) *ItemBuilder {
	props := cloneItemProps(b.node)
	props[itemPropValidators] = append([]validation.Validator(nil), validators...)
	b.node.SetProps(props)
	return b
}

// Build returns the wrapped VNode.
func (b *ItemBuilder) Build() rtui.VNode {
	return b.node
}

func cloneItemProps(node *rtui.ComponentVNode) rtui.Props {
	if node == nil {
		return rtui.Props{}
	}
	props := node.Props()
	if props == nil {
		return rtui.Props{}
	}
	return props.Clone()
}

func renderFormItem(props rtui.Props) rtui.VNode {
	model := itemModelFromProps(props)
	state, ctx, hook := currentItemRuntimeState()
	syncItemRuntimeState(ctx, hook, state, model)

	child := bindFormItemChild(model.child, model.formID)
	layout, errorText := resolveItemView(model)

	switch layout {
	case LayoutHorizontal:
		return renderHorizontalItem(model.label, child, errorText)
	case LayoutInline:
		return renderInlineItem(model.label, child, errorText)
	default:
		return renderVerticalItem(model.label, child, errorText)
	}
}

func itemModelFromProps(props rtui.Props) itemModel {
	model := itemModel{
		key:    "",
		field:  "",
		formID: "",
		label:  "",
		layout: "",
	}
	if props == nil {
		return model
	}
	if value, ok := props["key"].(string); ok {
		model.key = value
	}
	if value, ok := props[itemPropField].(string); ok {
		model.field = value
	}
	if value, ok := props[itemPropFormID].(string); ok {
		model.formID = value
	}
	if value, ok := props[itemPropLabel].(string); ok {
		model.label = value
	}
	if value, ok := props[itemPropLayout].(FormLayout); ok {
		model.layout = value
	} else if value, ok := props[itemPropLayout].(string); ok {
		model.layout = FormLayout(value)
	}
	if value, ok := props[itemPropChild].(rtui.VNode); ok {
		model.child = value
	}
	if value, ok := props[itemPropValidators].([]validation.Validator); ok {
		model.validators = append([]validation.Validator(nil), value...)
	}
	return model
}

func currentItemRuntimeState() (*itemRuntimeState, *rtui.ComponentContext, *rtui.Hook) {
	ctx := rtui.GetCurrentContext()
	if ctx == nil {
		return &itemRuntimeState{}, nil, nil
	}

	if err := ctx.Validator.ValidateHookCall(rtui.HookRef); err != nil {
		panic(err)
	}

	hook := ctx.GetOrCreateHook(rtui.HookRef)
	if state, ok := hook.Value.(*itemRuntimeState); ok && state != nil {
		return state, ctx, hook
	}

	state := &itemRuntimeState{}
	hook.Value = state
	return state, ctx, hook
}

func syncItemRuntimeState(
	ctx *rtui.ComponentContext,
	hook *rtui.Hook,
	state *itemRuntimeState,
	model itemModel,
) {
	if state == nil {
		return
	}

	sourceID := model.key
	if ctx != nil && ctx.ComponentID != "" {
		sourceID = ctx.ComponentID
	}
	if sourceID == "" && model.field != "" {
		sourceID = "form-item:" + model.field
	}

	if hook != nil {
		hook.Cleanup = func() {
			cleanupItemRuntimeState(state)
		}
	}

	identityChanged := state.sourceID != sourceID || state.formID != model.formID || state.field != model.field
	if identityChanged {
		cleanupItemRuntimeState(state)
		state.sourceID = sourceID
		state.formID = model.formID
		state.field = model.field
		state.retryQueued = model.formID != ""
	}

	if model.formID == "" || model.field == "" || sourceID == "" {
		return
	}

	formInst := GetForm(model.formID)
	if formInst == nil {
		if ctx != nil && state.retryQueued {
			state.retryQueued = false
			ctx.ScheduleUpdate()
		}
		return
	}
	state.retryQueued = false

	if state.unsubscribe == nil {
		state.unsubscribe = formInst.Subscribe(func(changedField string) {
			if changedField != "" && changedField != model.field {
				return
			}
			if ctx != nil {
				ctx.ScheduleUpdate()
			}
		})
	}

	if !reflect.DeepEqual(state.validators, model.validators) {
		formInst.setValidatorSource(model.field, sourceID, model.validators)
		state.validators = append([]validation.Validator(nil), model.validators...)
	}
}

func cleanupItemRuntimeState(state *itemRuntimeState) {
	if state == nil {
		return
	}
	if state.unsubscribe != nil {
		state.unsubscribe()
		state.unsubscribe = nil
	}
	if state.formID != "" && state.field != "" && state.sourceID != "" {
		if formInst := GetForm(state.formID); formInst != nil {
			formInst.clearValidatorSource(state.field, state.sourceID)
		}
	}
	state.validators = nil
}

func resolveItemView(model itemModel) (FormLayout, string) {
	layout := model.layout
	errorText := ""
	if model.formID != "" && model.field != "" {
		if formInst := GetForm(model.formID); formInst != nil {
			if layout == "" {
				layout = formInst.Layout()
			}
			if err, ok := formInst.GetError(model.field); ok {
				errorText = err
			}
		}
	}
	return normalizeLayout(layout), errorText
}

func decorateFormChild(formID string, child rtui.VNode) rtui.VNode {
	if child == nil || formID == "" {
		return child
	}
	props := child.Props()
	if props == nil {
		return child
	}
	if marked, ok := props[itemPropMarker].(bool); !ok || !marked {
		return child
	}

	cloned := props.Clone()
	if currentFormID, _ := cloned[itemPropFormID].(string); currentFormID == "" {
		cloned[itemPropFormID] = formID
		child.SetProps(cloned)
	}
	return child
}

func bindFormItemChild(child rtui.VNode, formID string) rtui.VNode {
	if child == nil || formID == "" {
		return child
	}
	props := child.Props()
	if props == nil {
		props = rtui.Props{}
	} else {
		props = props.Clone()
	}
	props[itemPropFormID] = formID
	child.SetProps(props)
	return child
}

func renderVerticalItem(label string, child rtui.VNode, errorText string) rtui.VNode {
	children := make([]rtui.VNode, 0, 3)
	if label != "" {
		children = append(children, buildItemLabel(label))
	}
	if child != nil {
		children = append(children, child)
	}
	if errorText != "" {
		children = append(children, buildItemError(errorText))
	}
	return rtui.VStackBuilder(children...).Gap(0).Build()
}

func renderHorizontalItem(label string, child rtui.VNode, errorText string) rtui.VNode {
	rowChildren := make([]rtui.VNode, 0, 2)
	if label != "" {
		rowChildren = append(rowChildren, buildItemLabel(label))
	}
	if child != nil {
		rowChildren = append(rowChildren, rtui.Flex(child, 1))
	}

	row := rtui.HStackBuilder(rowChildren...).Gap(1).AlignCross(rtui.AlignStart).Build()
	if errorText == "" {
		return row
	}
	return rtui.VStackBuilder(row, buildItemError(errorText)).Gap(0).Build()
}

func renderInlineItem(label string, child rtui.VNode, errorText string) rtui.VNode {
	children := make([]rtui.VNode, 0, 3)
	if label != "" {
		children = append(children, buildItemLabel(label))
	}
	if child != nil {
		children = append(children, child)
	}
	if errorText != "" {
		children = append(children, buildItemError(errorText))
	}
	return rtui.HStackBuilder(children...).Gap(1).AlignCross(rtui.AlignCenter).Build()
}

func buildItemLabel(label string) rtui.VNode {
	return textcomp.New(label).
		Foreground(theme.Muted()).
		Bold(true)
}

func buildItemError(errorText string) rtui.VNode {
	return textcomp.New(errorText).
		Foreground(theme.Error())
}

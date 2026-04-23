package form

import (
	fcontext "github.com/wwsheng009/mint/runtime/context"
	rtui "github.com/wwsheng009/mint/runtime/ui"
)

// Context Key for Form (Phase 2)
const FormContextKey fcontext.ContextKey = "github.com/wwsheng009/mint/ui/components/form:form"

// =============================================================================
// Form Context Interface
// =============================================================================

// FormContext 提供表单数据和方法给子组件访问
// 子组件可以通过 UseContext 获取这个接口，然后与表单交互
type FormContext interface {
	// GetValue 获取字段的值
	GetValue(field string) (interface{}, bool)

	// SetValue 设置字段的值
	SetValue(field string, value interface{})

	// GetValues 获取所有字段值
	GetValues() map[string]interface{}

	// SetValues 批量设置字段值
	SetValues(values map[string]interface{})

	// GetError 获取字段的验证错误
	GetError(field string) (string, bool)

	// GetErrors 获取所有验证错误
	GetErrors() map[string]string

	// IsValid 返回表单是否有效
	IsValid() bool

	// IsSubmitting 返回表单是否正在提交
	IsSubmitting() bool

	// HasSubmitted 返回表单是否至少提交过一次。
	HasSubmitted() bool

	// GetSubmitCount 返回表单提交次数。
	GetSubmitCount() int

	// IsFieldTouched 返回字段是否已被访问（blur）
	IsFieldTouched(field string) bool

	// GetTouchedFields 返回所有已访问字段，按字段名稳定排序。
	GetTouchedFields() []string

	// IsFieldSubmitted 返回字段是否已参与提交/validate-all。
	IsFieldSubmitted(field string) bool

	// GetSubmittedFields 返回所有已提交字段，按字段名稳定排序。
	GetSubmittedFields() []string

	// IsFieldDirty 返回字段值是否相对初始值发生变化
	IsFieldDirty(field string) bool

	// GetDirtyFields 返回所有脏字段，按字段名稳定排序。
	GetDirtyFields() []string
}

// =============================================================================
// Form Context Implementation
// =============================================================================

// formContextImpl 是 FormContext 的实现
// 它持有对 Form Instance 的引用
type formContextImpl struct {
	form *Instance
}

// GetValue 获取字段的值
func (c *formContextImpl) GetValue(field string) (interface{}, bool) {
	return c.form.GetValue(field)
}

// SetValue 设置字段的值
func (c *formContextImpl) SetValue(field string, value interface{}) {
	c.form.SetValue(field, value)
}

// GetValues 获取所有字段值
func (c *formContextImpl) GetValues() map[string]interface{} {
	return c.form.GetValues()
}

// SetValues 批量设置字段值
func (c *formContextImpl) SetValues(values map[string]interface{}) {
	c.form.SetValues(values)
}

// GetError 获取字段的验证错误
func (c *formContextImpl) GetError(field string) (string, bool) {
	return c.form.GetError(field)
}

// GetErrors 获取所有验证错误
func (c *formContextImpl) GetErrors() map[string]string {
	return c.form.GetErrors()
}

// IsValid 返回表单是否有效
func (c *formContextImpl) IsValid() bool {
	return c.form.IsValid()
}

// IsSubmitting 返回表单是否正在提交
func (c *formContextImpl) IsSubmitting() bool {
	return c.form.IsSubmitting()
}

// HasSubmitted 返回表单是否至少提交过一次。
func (c *formContextImpl) HasSubmitted() bool {
	return c.form.HasSubmitted()
}

// GetSubmitCount 返回表单提交次数。
func (c *formContextImpl) GetSubmitCount() int {
	return c.form.GetSubmitCount()
}

// IsFieldTouched 返回字段是否已被访问（blur）
func (c *formContextImpl) IsFieldTouched(field string) bool {
	return c.form.IsFieldTouched(field)
}

// GetTouchedFields 返回所有已访问字段。
func (c *formContextImpl) GetTouchedFields() []string {
	return c.form.GetTouchedFields()
}

// IsFieldSubmitted 返回字段是否已参与提交/validate-all。
func (c *formContextImpl) IsFieldSubmitted(field string) bool {
	return c.form.IsFieldSubmitted(field)
}

// GetSubmittedFields 返回所有已提交字段。
func (c *formContextImpl) GetSubmittedFields() []string {
	return c.form.GetSubmittedFields()
}

// IsFieldDirty 返回字段值是否相对初始值发生变化
func (c *formContextImpl) IsFieldDirty(field string) bool {
	return c.form.IsFieldDirty(field)
}

// GetDirtyFields 返回所有脏字段。
func (c *formContextImpl) GetDirtyFields() []string {
	return c.form.GetDirtyFields()
}

// =============================================================================
// Context Helper Methods
// =============================================================================

// newFormContext 创建一个新的 FormContext 实例
func newFormContext(form *Instance) FormContext {
	if form == nil {
		return nil
	}
	return &formContextImpl{form: form}
}

// GetFormContext resolves a FormContext from the current render owner's
// instance tree only. It never performs implicit registry lookup.
// When formID is empty, it resolves the nearest ancestor Form.
func GetFormContext(formID string) FormContext {
	ctx := rtui.GetCurrentContext()
	if ctx == nil {
		return nil
	}
	owner := ctx.OwnerInstance()
	if owner == nil {
		return nil
	}
	return newFormContext(resolveFormFromOwner(owner, formID))
}

func resolveFormFromOwner(owner rtui.ComponentInstance, formID string) *Instance {
	if owner == nil {
		return nil
	}
	return resolveFormAncestor(owner, formID)
}

func resolveFormAncestor(owner rtui.ComponentInstance, formID string) *Instance {
	var current interface{} = owner
	for hops := 0; current != nil && hops < 64; hops++ {
		if formInst, ok := current.(*Instance); ok {
			if formID == "" || formInst.Key() == formID {
				return formInst
			}
		}

		treeNode, ok := current.(interface{ Parent() interface{} })
		if !ok {
			return nil
		}
		current = treeNode.Parent()
	}
	return nil
}

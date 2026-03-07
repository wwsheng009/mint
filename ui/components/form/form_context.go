package form

import (
	"sync"

	fcontext "github.com/wwsheng009/mint/runtime/context"
)

// Context Key for Form (Phase 2)
const FormContextKey fcontext.ContextKey = "github.com/wwsheng009/mint/ui/components/form:form"

// =============================================================================
// Global Form Registry (for context access)
// =============================================================================

// formRegistry stores active form instances by their formID.
// This allows child components to access form data via formID without
// needing a direct parent reference through the tree.
var (
	formRegistry = make(map[string]*Instance)
	formMu       sync.RWMutex
)

// RegisterForm registers a form instance with the given formID.
func RegisterForm(formID string, form *Instance) {
	formMu.Lock()
	defer formMu.Unlock()
	formRegistry[formID] = form
}

// UnregisterForm unregisters a form instance.
func UnregisterForm(formID string) {
	formMu.Lock()
	defer formMu.Unlock()
	delete(formRegistry, formID)
}

// GetForm returns the form instance for the given formID.
func GetForm(formID string) *Instance {
	formMu.RLock()
	defer formMu.RUnlock()
	return formRegistry[formID]
}

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

// =============================================================================
// Context Helper Methods
// =============================================================================

// newFormContext 创建一个新的 FormContext 实例
func newFormContext(form *Instance) FormContext {
	return &formContextImpl{form: form}
}

// GetFormContext 通过 formID 获取 FormContext
// 这是一个便捷方法，供子组件使用
// 返回的 FormContext 是一个临时的包装器，每次访问都会创建
func GetFormContext(formID string) FormContext {
	form := GetForm(formID)
	if form == nil {
		return nil
	}
	return newFormContext(form)
}

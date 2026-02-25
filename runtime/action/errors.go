package action

import (
	"errors"
	"fmt"
)

// =============================================================================
// Action Errors
// =============================================================================
// Structured action handling errors for debugging and logging

// ErrorType 错误类型
type ErrorType string

const (
	// Target-related errors
	ErrTargetNotFound        ErrorType = "target_not_found"         // 目标组件未找到
	ErrTargetDisabled        ErrorType = "target_disabled"           // 目标组件已禁用
	ErrTargetNotInteractable ErrorType = "target_not_interactable"  // 目标组件不可交互
	ErrTargetExists          ErrorType = "target_exists"            // 目标已存在

	// Payload-related errors
	ErrInvalidPayload      ErrorType = "invalid_payload"           // 无效的 Payload
	ErrMissingPayload      ErrorType = "missing_payload"           // 缺少必需的 Payload
	ErrPayloadTypeMismatch ErrorType = "payload_type_mismatch"     // Payload 类型不匹配

	// Action-related errors
	ErrActionNotSupported ErrorType = "action_not_supported"      // 组件不支持此 Action
	ErrActionNotAllowed   ErrorType = "action_not_allowed"        // 当前状态不允许此操作
	ErrActionFailed       ErrorType = "action_failed"             // Action 执行失败
	ErrActionUnregistered ErrorType = "action_unregistered"       // Action 未注册

	// System errors
	ErrDispatchFailed ErrorType = "dispatch_failed"               // 分发失败
	ErrTimeout        ErrorType = "timeout"                       // 操作超时
	ErrPanicRecovered  ErrorType = "panic_recovered"              // Panic 已恢复

	// Scope errors
	ErrScopeNotFound ErrorType = "scope_not_found"                // Scope 未找到
	ErrScopeCycle    ErrorType = "scope_cycle"                    // Scope 循环
)

// Error Action 处理错误
type Error struct {
	// Type 错误类型
	Type ErrorType

	// Message 错误消息
	Message string

	// Action 触发错误的 Action
	Action *Action

	// Target 目标组件 ID（string）
	Target string

	// TargetID 目标组件 ID（uint64）
	TargetID uint64

	// ComponentType 目标组件类型
	ComponentType string

	// Details 详细信息
	Details map[string]interface{}
}

// NewError 创建错误
func NewError(errorType ErrorType, message string, a *Action) *Error {
	return &Error{
		Type:    errorType,
		Message: message,
		Action:  a,
		Details: make(map[string]interface{}),
	}
}

// WithTarget 设置目标（string）
func (e *Error) WithTarget(target string) *Error {
	e.Target = target
	return e
}

// WithTargetID 设置目标 ID（uint64）
func (e *Error) WithTargetID(targetID uint64) *Error {
	e.TargetID = targetID
	return e
}

// WithComponentType 设置组件类型
func (e *Error) WithComponentType(typ string) *Error {
	e.ComponentType = typ
	return e
}

// WithDetail 添加详情
func (e *Error) WithDetail(key string, value interface{}) *Error {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// WithDetails 添加多个详情
func (e *Error) WithDetails(details map[string]interface{}) *Error {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	for k, v := range details {
		e.Details[k] = v
	}
	return e
}

// Error 实现 error 接口
func (e *Error) Error() string {
	target := ""
	if e.Target != "" {
		target = e.Target
	} else if e.TargetID != 0 {
		target = fmt.Sprintf("%d", e.TargetID)
	}

	if target != "" {
		return fmt.Sprintf("[%s] %s (action: %s, target: %s)",
			e.Type, e.Message, e.Action.Type, target)
	}
	return fmt.Sprintf("[%s] %s (action: %s)", e.Type, e.Message, e.Action.Type)
}

// String 返回详细的错误字符串
func (e *Error) String() string {
	s := e.Error()
	if len(e.Details) > 0 {
		s += ", details: {"
		first := true
		for k, v := range e.Details {
			if !first {
				s += ", "
			}
			s += fmt.Sprintf("%s: %v", k, v)
			first = false
		}
		s += "}"
	}
	if e.ComponentType != "" {
		s += fmt.Sprintf(", component_type: %s", e.ComponentType)
	}
	return s
}

// =============================================================================
// 预定义错误构造器
// =============================================================================

// NewErrTargetNotFound 创建目标未找到错误
func NewErrTargetNotFound(target string, a *Action) *Error {
	return NewError(ErrTargetNotFound,
		fmt.Sprintf("target component not found: %s", target), a).
		WithTarget(target)
}

// NewErrTargetNotFoundByID 创建目标未找到错误（uint64 ID）
func NewErrTargetNotFoundByID(targetID uint64, a *Action) *Error {
	return NewError(ErrTargetNotFound,
		fmt.Sprintf("target component not found: %d", targetID), a).
		WithTargetID(targetID)
}

// NewErrTargetDisabled 创建目标已禁用错误
func NewErrTargetDisabled(target string, a *Action) *Error {
	return NewError(ErrTargetDisabled,
		fmt.Sprintf("target component is disabled: %s", target), a).
		WithTarget(target)
}

// NewErrInvalidPayload 创建无效 Payload 错误
func NewErrInvalidPayload(expectedType string, a *Action) *Error {
	actualType := "nil"
	if a.Payload != nil {
		actualType = fmt.Sprintf("%T", a.Payload)
	}
	return NewError(ErrInvalidPayload,
		fmt.Sprintf("invalid payload type: expected %s, got %s", expectedType, actualType), a).
		WithDetail("expected_type", expectedType).
		WithDetail("actual_type", actualType)
}

// NewErrMissingPayload 创建缺少 Payload 错误
func NewErrMissingPayload(a *Action) *Error {
	return NewError(ErrMissingPayload,
		"action requires payload but none provided", a)
}

// NewErrActionNotSupported 创建不支持 Action 错误
func NewErrActionNotSupported(componentType, actionType string, a *Action) *Error {
	return NewError(ErrActionNotSupported,
		fmt.Sprintf("component %s does not support action: %s", componentType, actionType), a).
		WithComponentType(componentType).
		WithDetail("action_type", actionType)
}

// NewErrActionNotAllowed 创建不允许 Action 错误
func NewErrActionNotAllowed(reason string, a *Action) *Error {
	return NewError(ErrActionNotAllowed,
		fmt.Sprintf("action not allowed: %s", reason), a).
		WithDetail("reason", reason)
}

// NewErrActionFailed 创建 Action 执行失败错误
func NewErrActionFailed(reason string, a *Action) *Error {
	return NewError(ErrActionFailed,
		fmt.Sprintf("action execution failed: %s", reason), a).
		WithDetail("failure_reason", reason)
}

// =============================================================================
// Payload 验证辅助函数
// =============================================================================

// ValidateStringPayload 验证字符串 Payload
func ValidateStringPayload(a *Action) (string, *Error) {
	if a.Payload == nil {
		return "", NewErrMissingPayload(a)
	}
	if s, ok := a.Payload.(string); ok {
		return s, nil
	}
	return "", NewErrInvalidPayload("string", a)
}

// ValidateRunePayload 验证 rune Payload
func ValidateRunePayload(a *Action) (rune, *Error) {
	if a.Payload == nil {
		return 0, NewErrMissingPayload(a)
	}
	if r, ok := a.Payload.(rune); ok {
		return r, nil
	}
	return 0, NewErrInvalidPayload("rune", a)
}

// ValidateIntPayload 验证 int Payload
func ValidateIntPayload(a *Action) (int, *Error) {
	if a.Payload == nil {
		return 0, NewErrMissingPayload(a)
	}
	if i, ok := a.Payload.(int); ok {
		return i, nil
	}
	return 0, NewErrInvalidPayload("int", a)
}

// ValidateMapPayload 验证 map Payload
func ValidateMapPayload(a *Action) (map[string]interface{}, *Error) {
	if a.Payload == nil {
		return nil, NewErrMissingPayload(a)
	}
	if m, ok := a.Payload.(map[string]interface{}); ok {
		return m, nil
	}
	return nil, NewErrInvalidPayload("map[string]interface{}", a)
}

// ValidateBoolPayload 验证 bool Payload
func ValidateBoolPayload(a *Action) (bool, *Error) {
	if a.Payload == nil {
		return false, NewErrMissingPayload(a)
	}
	if b, ok := a.Payload.(bool); ok {
		return b, nil
	}
	return false, NewErrInvalidPayload("bool", a)
}

// =============================================================================
// 错误辅助函数
// =============================================================================

// IsErrTarget checks if error is target-related
func IsErrTarget(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Type == ErrTargetNotFound ||
			e.Type == ErrTargetDisabled ||
			e.Type == ErrTargetNotInteractable
	}
	return false
}

// IsErrPayload checks if error is payload-related
func IsErrPayload(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Type == ErrInvalidPayload ||
			e.Type == ErrMissingPayload ||
			e.Type == ErrPayloadTypeMismatch
	}
	return false
}

// IsErrAction checks if error is action-related
func IsErrAction(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Type == ErrActionNotSupported ||
			e.Type == ErrActionNotAllowed ||
			e.Type == ErrActionFailed
	}
	return false
}

// Wrap wraps an error with Action context
func Wrap(action *Action, err error) *Error {
	if err == nil {
		return nil
	}
	if e, ok := err.(*Error); ok {
		return e
	}
	return NewError(ErrActionFailed, err.Error(), action).WithDetail("original_error", err.Error())
}

// Unwrap extracts the original error if wrapped
func Unwrap(err error) error {
	if e, ok := err.(*Error); ok {
		if originalErr, exists := e.Details["original_error"]; exists {
			if s, ok := originalErr.(string); ok {
				// Can't convert back to original error type, return as string
				return errors.New(s)
			}
		}
	}
	return err
}

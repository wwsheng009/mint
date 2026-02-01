// sandbox/errors.go - 错误定义
package sandbox

import "errors"

var (
	// 生命周期错误
	ErrInvalidTransition = errors.New("sandbox: invalid state transition")
	ErrNotInitialized    = errors.New("sandbox: not initialized")
	ErrAlreadyRunning    = errors.New("sandbox: already running")
	ErrNotRunning        = errors.New("sandbox: not running")

	// 事件注入错误
	ErrInjectionNotAllowed = errors.New("sandbox: event injection not allowed")
	ErrInvalidStrategy     = errors.New("sandbox: invalid injection strategy")
	ErrQueueFull           = errors.New("sandbox: event queue full")
	ErrQueueEmpty          = errors.New("sandbox: event queue empty")

	// 快照错误
	ErrSnapshotNotFound = errors.New("sandbox: snapshot not found")
	ErrSnapshotCorrupt  = errors.New("sandbox: snapshot data corrupted")
	ErrRestoreFailed    = errors.New("sandbox: restore failed")

	// 配置错误
	ErrInvalidConfig = errors.New("sandbox: invalid configuration")

	// 断言错误
	ErrAssertionFailed = errors.New("sandbox: assertion failed")
	ErrTimeout         = errors.New("sandbox: operation timeout")
)

// AssertionError 断言错误详情
type AssertionError struct {
	Message  string
	Expected interface{}
	Actual   interface{}
	Selector string
}

// Error 返回错误信息
func (e *AssertionError) Error() string {
	if e.Expected != nil || e.Actual != nil {
		return e.Message + ": expected " + stringify(e.Expected) + ", got " + stringify(e.Actual)
	}
	return e.Message
}

// stringify 将值转换为字符串表示
func stringify(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

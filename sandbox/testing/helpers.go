// testing/helpers.go - 测试辅助函数
package testing

import (
	"time"
)

// Assert 断言辅助函数
type Assert struct {
	t       T
	failed  bool
}

// T 测试接口 (最小化依赖)
type T interface {
	Helper()
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	FailNow()
}

// NewAssert 创建断言器
func NewAssert(t T) *Assert {
	t.Helper()
	return &Assert{t: t}
}

// Equal 断言相等
func (a *Assert) Equal(expected, actual interface{}) {
	a.t.Helper()
	if expected != actual {
		a.t.Errorf("expected %v, got %v", expected, actual)
		a.failed = true
	}
}

// NotEqual 断言不相等
func (a *Assert) NotEqual(notExpected, actual interface{}) {
	a.t.Helper()
	if notExpected == actual {
		a.t.Errorf("expected not %v, got %v", notExpected, actual)
		a.failed = true
	}
}

// True 断言为真
func (a *Assert) True(value bool, msg ...string) {
	a.t.Helper()
	if !value {
		if len(msg) > 0 {
			a.t.Error(msg[0])
		} else {
			a.t.Error("expected true, got false")
		}
		a.failed = true
	}
}

// False 断言为假
func (a *Assert) False(value bool, msg ...string) {
	a.t.Helper()
	if value {
		if len(msg) > 0 {
			a.t.Error(msg[0])
		} else {
			a.t.Error("expected false, got true")
		}
		a.failed = true
	}
}

// Nil 断言为nil
func (a *Assert) Nil(value interface{}) {
	a.t.Helper()
	if value != nil {
		a.t.Errorf("expected nil, got %v", value)
		a.failed = true
	}
}

// NotNil 断言不为nil
func (a *Assert) NotNil(value interface{}) {
	a.t.Helper()
	if value == nil {
		a.t.Error("expected not nil, got nil")
		a.failed = true
	}
}

// NoError 断言没有错误
func (a *Assert) NoError(err error) {
	a.t.Helper()
	if err != nil {
		a.t.Errorf("expected no error, got %v", err)
		a.failed = true
	}
}

// Error 断言有错误
func (a *Assert) Error(err error) {
	a.t.Helper()
	if err == nil {
		a.t.Error("expected error, got nil")
		a.failed = true
	}
}

// Failed 是否失败
func (a *Assert) Failed() bool {
	return a.failed
}

// Eventually 最终断言 (带重试)
func (a *Assert) Eventually(condition func() bool, timeout time.Duration, checkInterval time.Duration) {
	a.t.Helper()
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		if checkInterval > 0 {
			time.Sleep(checkInterval)
		}
	}

	a.t.Error("eventually condition not met within timeout")
	a.failed = true
}

// EventuallyTimeout 默认超时时间
const EventuallyTimeout = 5 * time.Second

// EventuallyCheckInterval 默认检查间隔
const EventuallyCheckInterval = 100 * time.Millisecond

// Retry 重试函数
func Retry(fn func() error, maxAttempts int, interval time.Duration) error {
	var lastErr error
	for i := 0; i < maxAttempts; i++ {
		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
		}
		if interval > 0 && i < maxAttempts-1 {
			time.Sleep(interval)
		}
	}
	return lastErr
}

// WaitFor 等待条件
func WaitFor(condition func() bool, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

// Must 必须成功，否则panic
func Must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

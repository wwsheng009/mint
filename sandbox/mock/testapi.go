// mock/testapi.go - 测试辅助API
package mock

import (
	"time"

	"github.com/wwsheng009/mint/runtime/platform"
)

// TestHelper 测试辅助器
type TestHelper struct {
	sandbox *MockSandbox
	errors  []error
}

// NewTestHelper 创建测试辅助器
func NewTestHelper(sb *MockSandbox) *TestHelper {
	return &TestHelper{
		sandbox: sb,
		errors:  make([]error, 0),
	}
}

// Errors 返回所有错误
func (th *TestHelper) Errors() []error {
	return th.errors
}

// HasErrors 检查是否有错误
func (th *TestHelper) HasErrors() bool {
	return len(th.errors) > 0
}

// ClearErrors 清除错误
func (th *TestHelper) ClearErrors() {
	th.errors = th.errors[:0]
}

// ==============================================================================
// Action Methods (链式调用，返回 *TestHelper)
// ==============================================================================

// Type 输入文本
func (th *TestHelper) Type(text string) *TestHelper {
	if err := th.sandbox.InjectString(text); err != nil {
		th.errors = append(th.errors, err)
	}
	return th
}

// TypeFast 快速输入文本（无延迟）
func (th *TestHelper) TypeFast(text string) *TestHelper {
	if err := th.sandbox.InjectString(text); err != nil {
		th.errors = append(th.errors, err)
	}
	return th
}

// Press 按下按键
func (th *TestHelper) Press(key platform.SpecialKey) *TestHelper {
	if err := th.sandbox.InjectSpecialKey(key); err != nil {
		th.errors = append(th.errors, err)
	}
	return th
}

// PressKey 按下字符键
func (th *TestHelper) PressKey(key rune) *TestHelper {
	if err := th.sandbox.InjectKey(key); err != nil {
		th.errors = append(th.errors, err)
	}
	return th
}

// Click 点击
func (th *TestHelper) Click(x, y int) *TestHelper {
	if err := th.sandbox.InjectMouse(x, y, platform.MouseLeft, platform.MousePress); err != nil {
		th.errors = append(th.errors, err)
	}
	return th
}

// Tab 按 Tab 键
func (th *TestHelper) Tab() *TestHelper {
	return th.Press(platform.KeyTab)
}

// Enter 按 Enter 键
func (th *TestHelper) Enter() *TestHelper {
	return th.Press(platform.KeyEnter)
}

// Escape 按 Escape 键
func (th *TestHelper) Escape() *TestHelper {
	return th.Press(platform.KeyEscape)
}

// Process 处理所有事件
func (th *TestHelper) Process() *TestHelper {
	if err := th.sandbox.ProcessEvents(); err != nil {
		th.errors = append(th.errors, err)
	}
	return th
}

// Wait 等待一段时间
func (th *TestHelper) Wait(d time.Duration) *TestHelper {
	time.Sleep(d)
	return th
}

// ==============================================================================
// Assertion Methods
// ==============================================================================

// AssertRender 断言渲染包含文本
func (th *TestHelper) AssertRender(text string) *TestHelper {
	if err := th.sandbox.AssertRender(text); err != nil {
		th.errors = append(th.errors, err)
	}
	return th
}

// AssertNotRender 断言渲染不包含文本
func (th *TestHelper) AssertNotRender(text string) *TestHelper {
	if err := th.sandbox.AssertNotRender(text); err != nil {
		th.errors = append(th.errors, err)
	}
	return th
}

// ==============================================================================
// Result Method
// ==============================================================================

// Result 返回测试结果
type TestResult struct {
	Errors []error
}

// Result 完成链式调用并返回结果
func (th *TestHelper) Result() TestResult {
	return TestResult{
		Errors: th.errors,
	}
}

// OK 检查是否成功 (无错误)
func (r TestResult) OK() bool {
	return len(r.Errors) == 0
}

// Error 返回第一个错误
func (r TestResult) Error() error {
	if len(r.Errors) == 0 {
		return nil
	}
	return r.Errors[0]
}

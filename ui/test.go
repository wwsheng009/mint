// ui/test.go - UI层测试集成
package ui

import (
	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/sandbox"
	"github.com/wwsheng009/mint/sandbox/mock"
)

// TestApp 测试应用包装器
type TestApp struct {
	sandbox *mock.MockSandbox
	app     ComponentFunc // 实际应用
	root    *declarativeRoot // 渲染器
}

// testConfig 测试配置
type testConfig struct {
	width  int
	height int
}

// TestOption 测试选项类型
type TestOption func(*testConfig)

// TestRun 运行测试应用
func TestRun(app interface{}, opts ...TestOption) (*TestApp, error) {
	config := &testConfig{
		width:  80,
		height: 24,
	}

	for _, opt := range opts {
		opt(config)
	}

	sb := mock.New(config.width, config.height)

	if err := sb.Initialize(nil); err != nil {
		return nil, err
	}

	// 创建渲染器（不需要 framework.App）
	var appFn ComponentFunc
	if fn, ok := app.(ComponentFunc); ok {
		appFn = fn
	} else if fn, ok := app.(func() VNode); ok {
		appFn = ComponentFunc(fn)
	} else {
		// 如果不是有效的应用函数，仍然返回 TestApp（用于沙箱测试）
		appFn = nil
	}

	var root *declarativeRoot
	if appFn != nil {
		root = newDeclarativeRoot(appFn, nil).(*declarativeRoot)
	}

	testApp := &TestApp{
		sandbox: sb,
		app:     appFn,
		root:    root,
	}

	// 初始渲染
	if root != nil {
		testApp.render()
	}

	return testApp, nil
}

// render 渲染应用到沙箱缓冲区
func (ta *TestApp) render() {
	if ta.root == nil || ta.sandbox == nil {
		return
	}

	// 获取缓冲区
	buffer := ta.sandbox.Buffer()
	if buffer == nil {
		return
	}

	// 重置上下文并调用应用函数
	ta.root.ctx.resetContext()
	setCurrentContext(ta.root.ctx)

	// 获取 VNode 并渲染
	vnode := ta.root.appFn()

	// 简化的渲染：只渲染到 buffer（不处理布局）
	ta.root.renderVNode(vnode, 0, 0, buffer)

	setCurrentContext(nil)
}

// TestRunWithConfig 使用自定义配置运行测试应用
func TestRunWithConfig(app interface{}, config *sandbox.Config) (*TestApp, error) {
	sb := mock.New(config.Width, config.Height)

	if err := sb.Initialize(config); err != nil {
		return nil, err
	}

	// 创建渲染器
	var appFn ComponentFunc
	if fn, ok := app.(ComponentFunc); ok {
		appFn = fn
	} else if fn, ok := app.(func() VNode); ok {
		appFn = ComponentFunc(fn)
	} else {
		appFn = nil
	}

	var root *declarativeRoot
	if appFn != nil {
		root = newDeclarativeRoot(appFn, nil).(*declarativeRoot)
	}

	testApp := &TestApp{
		sandbox: sb,
		app:     appFn,
		root:    root,
	}

	if root != nil {
		testApp.render()
	}

	return testApp, nil
}

// Close 关闭测试应用
func (ta *TestApp) Close() error {
	return ta.sandbox.Close()
}

// Sandbox 获取沙箱
func (ta *TestApp) Sandbox() *mock.MockSandbox {
	return ta.sandbox
}

// Helper 获取测试辅助器
func (ta *TestApp) Helper() *mock.TestHelper {
	return ta.sandbox.Helper()
}

// SetEventHandler 设置事件处理器（渲染后）
func (ta *TestApp) SetEventHandler(handler sandbox.EventHandler) {
	// 包装处理器，在处理事件后重新渲染
	wrappedHandler := func(event platform.RawInput) error {
		// 调用用户提供的处理器（如果有）
		if handler != nil {
			return handler(event)
		}
		return nil
	}
	ta.sandbox.SetEventHandler(sandbox.EventHandler(wrappedHandler))
}

// ClickButton 直接点击按钮（不通过事件）
func (ta *TestApp) ClickButton(index int) error {
	if ta.root == nil {
		return nil
	}

	if index < 0 || index >= len(ta.root.buttons) {
		return nil
	}

	button := ta.root.buttons[index]
	if onClick := button.OnClick(); onClick != nil {
		onClick()
		// 重新渲染
		ta.render()
	}
	return nil
}

// ClickButtonByLabel 点击指定标签的按钮
func (ta *TestApp) ClickButtonByLabel(label string) error {
	if ta.root == nil {
		return nil
	}

	for i, btn := range ta.root.buttons {
		if btn.Label() == label {
			return ta.ClickButton(i)
		}
	}
	return nil
}

// ==============================================================================
// Test Options
// ==============================================================================

// TestWithWidth 设置测试宽度 (避免与 ui.WithWidth 冲突)
func TestWithWidth(w int) TestOption {
	return func(c *testConfig) {
		c.width = w
	}
}

// TestWithHeight 设置测试高度 (避免与 ui.WithHeight 冲突)
func TestWithHeight(h int) TestOption {
	return func(c *testConfig) {
		c.height = h
	}
}

// TestWithSize 设置测试尺寸
func TestWithSize(w, h int) TestOption {
	return func(c *testConfig) {
		c.width = w
		c.height = h
	}
}

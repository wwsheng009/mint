// ui/test.go - UI层测试集成
package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/sandbox"
	"github.com/wwsheng009/mint/sandbox/mock"
)

// TestApp 测试应用包装器
// 提供简化的测试环境，不需要完整的 framework.App
type TestApp struct {
	sandbox       *mock.MockSandbox
	app           ComponentFunc
	ctx           *ComponentContext
	buttons       []*ButtonVNode // 收集的按钮
	focusedIndex  int            // 当前焦点索引
}

// testConfig 测试配置
type testConfig struct {
	width  int
	height int
}

// TestOption 测试选项类型
type TestOption func(*testConfig)

// TestRun 运行测试应用
// 创建一个测试环境，初始化 Context 和 MockSandbox
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

	// 解析应用函数
	var appFn ComponentFunc
	if fn, ok := app.(ComponentFunc); ok {
		appFn = fn
	} else if fn, ok := app.(func() VNode); ok {
		appFn = ComponentFunc(fn)
	} else {
		appFn = nil
	}

	// 创建测试 Context（不需要 framework.App）
	var ctx *ComponentContext
	if appFn != nil {
		ctx = newComponentContext("TestApp")
	}

	testApp := &TestApp{
		sandbox:      sb,
		app:          appFn,
		ctx:          ctx,
		buttons:      make([]*ButtonVNode, 0),
		focusedIndex: -1,
	}

	// 设置事件处理器
	testApp.setupEventHandler()

	// 初始渲染以触发 hooks 初始化
	if appFn != nil && ctx != nil {
		testApp.Render()
	}

	return testApp, nil
}

// setupEventHandler 设置 Sandbox 的事件处理器
func (ta *TestApp) setupEventHandler() {
	ta.sandbox.SetEventHandler(func(event platform.RawInput) error {
		if os.Getenv("TUI_DEBUG_UI") == "true" {
			fmt.Fprintf(os.Stderr, "[TestApp] Received event: Type=%d, Key=%c, Special=%d\n",
				event.Type, event.Key, event.Special)
		}

		// 处理键盘事件
		if event.Type == platform.InputKeyPress {
			return ta.handleKeyEvent(event)
		}

		// 处理鼠标事件
		if event.Type == platform.InputMouse {
			return ta.handleMouseEvent(event)
		}

		return nil
	})
}

// handleKeyEvent 处理键盘事件
func (ta *TestApp) handleKeyEvent(event platform.RawInput) error {
	// Tab: 切换焦点
	if event.Special == platform.KeyTab {
		if len(ta.buttons) > 0 {
			if event.Modifiers&platform.ModShift != 0 {
				// Shift+Tab: 上一个
				ta.focusedIndex--
				if ta.focusedIndex < 0 {
					ta.focusedIndex = len(ta.buttons) - 1
				}
			} else {
				// Tab: 下一个
				ta.focusedIndex++
				if ta.focusedIndex >= len(ta.buttons) {
					ta.focusedIndex = 0
				}
			}
			ta.Render()
			if os.Getenv("TUI_DEBUG_UI") == "true" {
				fmt.Fprintf(os.Stderr, "[TestApp] Tab: focusedIndex=%d\n", ta.focusedIndex)
			}
		}
		return nil
	}

	// Enter: 触发焦点按钮的点击
	if event.Special == platform.KeyEnter {
		if ta.focusedIndex >= 0 && ta.focusedIndex < len(ta.buttons) {
			button := ta.buttons[ta.focusedIndex]
			if os.Getenv("TUI_DEBUG_UI") == "true" {
				fmt.Fprintf(os.Stderr, "[TestApp] Enter: triggering button %d (label=%s)\n",
					ta.focusedIndex, button.Label())
			}
			if onClick := button.OnClick(); onClick != nil {
				onClick()
			}
			ta.Render()
		}
		return nil
	}

	return nil
}

// handleMouseEvent 处理鼠标事件
func (ta *TestApp) handleMouseEvent(event platform.RawInput) error {
	// 检查是否点击了某个按钮
	if event.MouseAction == platform.MousePress {
		for i, btn := range ta.buttons {
			if btn.ContainsPoint(event.MouseX, event.MouseY) {
				if os.Getenv("TUI_DEBUG_UI") == "true" {
					fmt.Fprintf(os.Stderr, "[TestApp] Mouse click on button %d (label=%s)\n",
						i, btn.Label())
				}
				if onClick := btn.OnClick(); onClick != nil {
					onClick()
				}
				ta.focusedIndex = i
				ta.Render()
				return nil
			}
		}
	}
	return nil
}

// collectButtons 从 VNode 中收集按钮
func (ta *TestApp) collectButtons(node VNode) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *ButtonVNode:
		// 收集非禁用按钮
		if !n.Disabled() {
			n.focusIndex = len(ta.buttons)
			ta.buttons = append(ta.buttons, n)
		}
	case *ElementVNode:
		for _, child := range n.Children() {
			ta.collectButtons(child)
		}
	case *LayoutNode:
		for _, child := range n.Children() {
			ta.collectButtons(child)
		}
	case *FragmentVNode:
		for _, child := range n.Children() {
			ta.collectButtons(child)
		}
	}
}

// Render 渲染应用
// 设置当前 Context 并调用应用函数
func (ta *TestApp) Render() {
	if ta.app == nil || ta.ctx == nil {
		return
	}

	// 清空按钮列表（每次渲染重新收集）
	ta.buttons = ta.buttons[:0]

	// 重置 hook index 并设置为当前 context
	ta.ctx.ResetContext()
	SetCurrentContext(ta.ctx)

	// 调用应用函数获取 VNode
	vnode := ta.app()

	// 收集按钮（用于事件处理）
	ta.collectButtons(vnode)

	// 清除当前 context
	SetCurrentContext(nil)

	if os.Getenv("TUI_DEBUG_UI") == "true" {
		fmt.Fprintf(os.Stderr, "[TestApp] Render: collected %d buttons\n", len(ta.buttons))
	}
}

// GetContext 获取测试用的 ComponentContext
// 用于调试和检查 Hooks 状态
func (ta *TestApp) GetContext() *ComponentContext {
	return ta.ctx
}

// GetButtons 获取收集的按钮
func (ta *TestApp) GetButtons() []*ButtonVNode {
	return ta.buttons
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

// TriggerButtonClick 直接触发按钮点击（通过索引）
func (ta *TestApp) TriggerButtonClick(buttonIndex int) {
	if buttonIndex < 0 || buttonIndex >= len(ta.buttons) {
		return
	}

	button := ta.buttons[buttonIndex]
	if onClick := button.OnClick(); onClick != nil {
		if os.Getenv("TUI_DEBUG_UI") == "true" {
			fmt.Fprintf(os.Stderr, "[TestApp] TriggerButtonClick: button %d (label=%s)\n",
				buttonIndex, button.Label())
		}
		onClick()
		ta.Render()
	}
}

// TriggerButtonClickByLabel 通过标签触发按钮点击
func (ta *TestApp) TriggerButtonClickByLabel(label string) {
	for i, btn := range ta.buttons {
		if btn.Label() == label {
			ta.TriggerButtonClick(i)
			return
		}
	}
}

// ==============================================================================
// Test Options
// ==============================================================================

// TestWithWidth 设置测试宽度
func TestWithWidth(w int) TestOption {
	return func(c *testConfig) {
		c.width = w
	}
}

// TestWithHeight 设置测试高度
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

// ==============================================================================
// TestRunWithConfig 使用自定义配置运行测试应用
// ==============================================================================

func TestRunWithConfig(app interface{}, config *sandbox.Config) (*TestApp, error) {
	width := config.Width
	height := config.Height
	if width == 0 {
		width = 80
	}
	if height == 0 {
		height = 24
	}
	return TestRun(app, TestWithSize(width, height))
}


// ============================================================================
// 测试模式 - RunTest
// ============================================================================

// TestableApp 可测试的应用包装器
type TestableApp struct {
	fwApp *framework.App
	root  *declarativeRoot
	opts  *Options
}

// RunTest 运行可测试的应用（在后台运行，支持事件注入）
// 注意：此函数仅用于测试
func RunTest(app ComponentFunc, opts ...Option) (*TestableApp, error) {
	options := &Options{
		Width:  80,
		Height: 24,
		Title:  "Mint UI Test",
		FPS:    60,
	}

	for _, opt := range opts {
		opt(options)
	}

	// Create the framework app
	fwApp := framework.NewApp()
	fwApp.Resize(options.Width, options.Height)

	// Initialize theme (optional, don't fail on error)
	fwApp.InitTheme("dark")

	// Create the declarative root component
	declarativeNode := newDeclarativeRoot(app, fwApp)

	// Type assert to get the concrete type for testing
	declarativeRoot, ok := declarativeNode.(*declarativeRoot)
	if !ok {
		return nil, fmt.Errorf("failed to get declarativeRoot")
	}

	// Set as root
	fwApp.SetRoot(declarativeNode)

	// Run the app in background (Run will call Init)
	go func() {
		fwApp.Run()
	}()

	// Wait for app to start running
	for i := 0; i < 100; i++ {
		if fwApp.GetState() == framework.StateRunning {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	return &TestableApp{
		fwApp: fwApp,
		root:  declarativeRoot,
		opts:  options,
	}, nil
}

// InjectKey 注入字符键
func (ta *TestableApp) InjectKey(key rune) error {
	raw := platform.RawInput{
		Type: platform.InputKeyPress,
		Key:  key,
	}
	return ta.fwApp.InjectEvent(raw)
}

// InjectSpecialKey 注入特殊按键
func (ta *TestableApp) InjectSpecialKey(key platform.SpecialKey) error {
	raw := platform.RawInput{
		Type:    platform.InputKeyPress,
		Special: key,
	}
	return ta.fwApp.InjectEvent(raw)
}

// InjectKeyWithMod 注入带修饰符的按键
func (ta *TestableApp) InjectKeyWithMod(key rune, mod platform.KeyModifier) error {
	raw := platform.RawInput{
		Type:      platform.InputKeyPress,
		Key:       key,
		Modifiers: mod,
	}
	return ta.fwApp.InjectEvent(raw)
}

// InjectMouse 注入鼠标事件
func (ta *TestableApp) InjectMouse(x, y int, button platform.MouseButton, action platform.MouseAction) error {
	raw := platform.RawInput{
		Type:        platform.InputMouse,
		MouseX:      x,
		MouseY:      y,
		MouseButton: button,
		MouseAction: action,
	}
	return ta.fwApp.InjectEvent(raw)
}

// InjectString 注入字符串（逐字符注入）
func (ta *TestableApp) InjectString(text string) error {
	for _, r := range text {
		if err := ta.InjectKey(r); err != nil {
			return err
		}
		time.Sleep(10 * time.Millisecond) // 短暂延迟避免事件丢失
	}
	return nil
}

// GetBuffer 获取当前渲染缓冲区
func (ta *TestableApp) GetBuffer() *paint.Buffer {
	return ta.fwApp.GetRenderer().GetBackBuffer()
}

// GetRenderString 获取渲染输出字符串
func (ta *TestableApp) GetRenderString() string {
	buf := ta.GetBuffer()
	if buf == nil {
		return ""
	}

	var sb strings.Builder
	for y := 0; y < buf.Height; y++ {
		for x := 0; x < buf.Width; x++ {
			cell := buf.Cells[y][x]
			if cell.IsContinuation {
				continue
			}
			if cell.Cluster == "" {
				sb.WriteRune(' ')
			} else {
				sb.WriteString(cell.Cluster)
			}
		}
		if y < buf.Height-1 {
			sb.WriteRune('\n')
		}
	}
	return sb.String()
}

// AssertRender 断言渲染输出包含指定文本
func (ta *TestableApp) AssertRender(text string) error {
	rendered := ta.GetRenderString()
	if !strings.Contains(rendered, text) {
		return fmt.Errorf("render does not contain expected text: %q\nactual:\n%s", text, rendered)
	}
	return nil
}

// AssertNotRender 断言渲染输出不包含指定文本
func (ta *TestableApp) AssertNotRender(text string) error {
	rendered := ta.GetRenderString()
	if strings.Contains(rendered, text) {
		return fmt.Errorf("render contains unexpected text: %q\nactual:\n%s", text, rendered)
	}
	return nil
}

// Close 关闭测试应用
func (ta *TestableApp) Close() error {
	return ta.fwApp.Close()
}

// GetFrameworkApp 获取 framework.App（用于高级测试场景）
func (ta *TestableApp) GetFrameworkApp() *framework.App {
	return ta.fwApp
}

// GetDeclarativeRoot 获取 declarativeRoot（用于调试）
func (ta *TestableApp) GetDeclarativeRoot() *declarativeRoot {
	return ta.root
}

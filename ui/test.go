// ui/test.go - UI层测试集成
package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/wwsheng009/mint/framework"
	"github.com/wwsheng009/mint/internal/render"
	"github.com/wwsheng009/mint/runtime/intent"
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/platform"
	rtui "github.com/wwsheng009/mint/runtime/ui"
	"github.com/wwsheng009/mint/sandbox"
	"github.com/wwsheng009/mint/sandbox/mock"
)

// =============================================================================
// 旧版测试 API (简单测试)
// =============================================================================

// TestApp 测试应用包装器
// 提供简化的测试环境，不需要完整的 framework.App
type TestApp struct {
	sandbox *mock.MockSandbox
	app     ComponentFunc
	ctx     *ComponentContext
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
		ctx = NewComponentContextForRoot()
	}

	testApp := &TestApp{
		sandbox: sb,
		app:     appFn,
		ctx:     ctx,
	}

	// 设置事件处理器（简单记录）
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
		return nil
	})
}

// Render 渲染应用
func (ta *TestApp) Render() {
	if ta.app == nil || ta.ctx == nil {
		return
	}

	// 重置 hook index 并设置为当前 context
	ta.ctx.ResetContext()
	SetCurrentContext(ta.ctx)

	// 调用应用函数获取 VNode
	_ = ta.app()

	// 清除当前 context
	SetCurrentContext(nil)
}

// GetContext 获取测试用的 ComponentContext
func (ta *TestApp) GetContext() *ComponentContext {
	return ta.ctx
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

// =============================================================================
// Test Options
// =============================================================================

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

// TestRunWithConfig 使用自定义配置运行测试应用
func TestRunWithConfig(app interface{}, config interface{}) (*TestApp, error) {
	// Try to use *sandbox.Config if provided
	var width, height int
	if sbConfig, ok := config.(*sandbox.Config); ok {
		width = sbConfig.Width
		height = sbConfig.Height
	} else {
		width = 80
		height = 24
	}
	return TestRun(app, TestWithSize(width, height))
}

// ============================================================================
// 新版测试 API - RunTest (完整框架支持)
// ============================================================================

// TestableApp 可测试的应用包装器
// 使用完整的 framework.App，支持事件注入
type TestableApp struct {
	fwApp   *framework.App
	root    *render.DeclarativeNode
	opts    *Options
	sandbox *mock.MockSandbox
}

// RunTest 运行可测试的应用（在后台运行，支持事件注入）
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

	// Set global appInstance
	appInstance = fwApp

	// Initialize Intent Runtime (required for event handling)
	intentRuntime := intent.NewRuntime()
	intent.SetupBuiltinHandlers(intentRuntime) // Register built-in intent handlers
	rtui.SetGlobalIntentRuntime(intentRuntime)

	// Call initialization function if provided (e.g., for registering Intent Handlers)
	if options.InitFunc != nil {
		options.InitFunc()
	}

	// Create the declarative root component with Fiber reconciler enabled
	// Fiber is now the default and required for persistent component instances and event handlers
	declarativeNode := render.NewDeclarativeNodeFromFuncWithFiber(app, fwApp)

	// Pass Intent Runtime to declarative node for component context
	render.SetDeclarativeNodeIntentRuntime(declarativeNode, intentRuntime)

	// Set as root
	fwApp.SetRoot(declarativeNode)

	// Run the app in background (Run will call Init)
	go func() {
		fwApp.Run()
	}()

	// Wait for app to start running
	for i := 0; i < 200; i++ {
		if fwApp.GetState() == framework.StateRunning {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	return &TestableApp{
		fwApp:   fwApp,
		root:    declarativeNode,
		opts:    options,
		sandbox: nil, // RunTest doesn't use MockSandbox
	}, nil
}

// RunTestWithSandbox 使用 MockSandbox 作为事件源进行测试
func RunTestWithSandbox(app ComponentFunc, opts ...Option) (*TestableApp, error) {
	options := &Options{
		Width:  80,
		Height: 24,
		Title:  "Mint UI Test",
		FPS:    60,
	}

	for _, opt := range opts {
		opt(options)
	}

	// Create MockSandbox for testing
	sb := mock.New(options.Width, options.Height)
	if err := sb.Initialize(nil); err != nil {
		return nil, err
	}

	// Create the framework app
	fwApp := framework.NewApp()
	fwApp.Resize(options.Width, options.Height)

	// Initialize theme (optional, don't fail on error)
	fwApp.InitTheme("dark")

	// Set global appInstance
	appInstance = fwApp

	// Initialize Intent Runtime (required for event handling)
	intentRuntime := intent.NewRuntime()
	intent.SetupBuiltinHandlers(intentRuntime) // Register built-in intent handlers
	rtui.SetGlobalIntentRuntime(intentRuntime)

	// Call initialization function if provided (e.g., for registering Intent Handlers)
	if options.InitFunc != nil {
		options.InitFunc()
	}

	// Create the declarative root component with Fiber reconciler enabled
	// Fiber is now the default and required for persistent component instances and event handlers
	declarativeNode := render.NewDeclarativeNodeFromFuncWithFiber(app, fwApp)

	// Pass Intent Runtime to declarative node for component context
	render.SetDeclarativeNodeIntentRuntime(declarativeNode, intentRuntime)

	// Set as root
	fwApp.SetRoot(declarativeNode)

	// Run the app in background (Run will call Init)
	go func() {
		fwApp.Run()
	}()

	// Wait for app to start running
	for i := 0; i < 200; i++ {
		if fwApp.GetState() == framework.StateRunning {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	return &TestableApp{
		fwApp:   fwApp,
		root:    declarativeNode,
		opts:    options,
		sandbox: sb,
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

// InjectSpecialKeyWithMod 注入带修饰符的特殊按键
func (ta *TestableApp) InjectSpecialKeyWithMod(key platform.SpecialKey, mod platform.KeyModifier) error {
	raw := platform.RawInput{
		Type:      platform.InputKeyPress,
		Special:   key,
		Modifiers: mod,
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
	renderer := ta.fwApp.GetRenderer()
	front := renderer.GetFrontBuffer()
	back := renderer.GetBackBuffer()

	// Check if front buffer has been rendered to
	hasContent := false
	if front != nil && front.Height > 0 && len(front.Cells) > 0 {
		for y := 0; y < front.Height; y++ {
			for x := 0; x < front.Width; x++ {
				if front.Cells[y][x].Cluster != "" && front.Cells[y][x].Cluster != " " {
					hasContent = true
					break
				}
			}
			if hasContent {
				break
			}
		}
	}

	if hasContent {
		return front
	}
	return back
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

// DumpBuffer 打印渲染输出到标准输出，用于调试和验证
func (ta *TestableApp) DumpBuffer() {
	fmt.Printf("===== Buffer Dump (%dx%d) =====\n", ta.GetBuffer().Width, ta.GetBuffer().Height)
	fmt.Print(ta.GetRenderString())
	fmt.Printf("===== End of Buffer =====\n")
}

// SaveBufferToFile 将渲染输出保存到文件
func (ta *TestableApp) SaveBufferToFile(path string) error {
	return os.WriteFile(path, []byte(ta.GetRenderString()), 0644)
}

// Close 关闭测试应用
func (ta *TestableApp) Close() error {
	if ta.sandbox != nil {
		ta.sandbox.Close()
	}
	return ta.fwApp.Close()
}

// GetFrameworkApp 获取 framework.App（用于高级测试场景）
func (ta *TestableApp) GetFrameworkApp() *framework.App {
	return ta.fwApp
}

// GetDeclarativeRoot 获取声明式根节点（用于调试）
func (ta *TestableApp) GetDeclarativeRoot() *render.DeclarativeNode {
	return ta.root
}

// GetSandbox 获取 MockSandbox（仅在使用 RunTestWithSandbox 创建时可用）
func (ta *TestableApp) GetSandbox() *mock.MockSandbox {
	return ta.sandbox
}

// GetFocusedIndex 获取当前焦点元素索引
func (ta *TestableApp) GetFocusedIndex() int {
	if ta.root == nil {
		return -1
	}
	return ta.root.GetFocusedIndex()
}

// GetFocusedType 获取当前焦点元素类型
func (ta *TestableApp) GetFocusedType() int {
	if ta.root == nil {
		return 0
	}
	return ta.root.GetFocusedType()
}

// GetButtons 获取按钮列表
func (ta *TestableApp) GetButtons() []interface{} {
	if ta.root == nil {
		return []interface{}{}
	}
	buttons := ta.root.GetButtons()
	result := make([]interface{}, len(buttons))
	for i, b := range buttons {
		result[i] = b
	}
	return result
}

// GetInputs 获取输入框列表
func (ta *TestableApp) GetInputs() []interface{} {
	if ta.root == nil {
		return []interface{}{}
	}
	inputs := ta.root.GetInputs()
	result := make([]interface{}, len(inputs))
	for i, inp := range inputs {
		result[i] = inp
	}
	return result
}

// GetTextareas 获取文本域列表
func (ta *TestableApp) GetTextareas() []interface{} {
	if ta.root == nil {
		return []interface{}{}
	}
	textareas := ta.root.GetTextareas()
	result := make([]interface{}, len(textareas))
	for i, ta := range textareas {
		result[i] = ta
	}
	return result
}

// GetCheckboxes 获取复选框列表
func (ta *TestableApp) GetCheckboxes() []interface{} {
	if ta.root == nil {
		return []interface{}{}
	}
	checkboxes := ta.root.GetCheckboxes()
	result := make([]interface{}, len(checkboxes))
	for i, cb := range checkboxes {
		result[i] = cb
	}
	return result
}

// GetSelects 获取选择框列表
func (ta *TestableApp) GetSelects() []interface{} {
	if ta.root == nil {
		return []interface{}{}
	}
	selects := ta.root.GetSelects()
	result := make([]interface{}, len(selects))
	for i, s := range selects {
		result[i] = s
	}
	return result
}

// GetFocusManager 获取焦点管理器
func (ta *TestableApp) GetFocusManager() interface{} {
	if ta.root == nil {
		return nil
	}
	return ta.root.GetFocusManager()
}

// ForceRender forces a render to ensure the VNode tree is built
// This is needed in Fiber mode where GetButtons() depends on Paint() being called first
func (ta *TestableApp) ForceRender() {
	// Trigger a render by calling the framework's render method
	if ta.fwApp != nil {
		ta.fwApp.ForceRenderNow()
	}
}

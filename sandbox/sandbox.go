// sandbox/sandbox.go - 核心接口定义
package sandbox

import (
	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/platform"
)

// Sandbox 沙箱核心接口
type Sandbox interface {
	// ========================================================================
	// 生命周期
	// ========================================================================

	// Initialize 初始化沙箱
	Initialize(config *Config) error

	// Run 运行沙箱主循环
	Run() error

	// Pause 暂停沙箱
	Pause() error

	// Resume 恢复沙箱
	Resume() error

	// Close 关闭沙箱并释放资源
	Close() error

	// ========================================================================
	// 状态查询
	// ========================================================================

	// State 获取当前状态
	State() State

	// Type 获取沙箱类型
	Type() SandboxType

	// Config 获取配置
	Config() *Config

	// ========================================================================
	// 缓冲区操作
	// ========================================================================

	// Buffer 获取渲染缓冲区
	Buffer() *paint.Buffer

	// SetBuffer 设置渲染缓冲区
	SetBuffer(buf *paint.Buffer)

	// Resize 调整缓冲区大小
	Resize(width, height int)

	// Size 获取当前尺寸
	Size() (width, height int)
}

// EventSource 事件源接口 (用于真实环境)
type EventSource interface {
	// Events 返回事件通道
	Events() <-chan platform.RawInput

	// Start 启动事件读取
	Start() error

	// Stop 停止事件读取
	Stop() error
}

// EventSink 事件注入接口 (用于测试环境)
type EventSink interface {
	// SetEventHandler 设置事件处理器
	SetEventHandler(handler EventHandler)

	// Inject 注入单个事件
	Inject(event platform.RawInput) error

	// InjectKey 注入按键事件
	InjectKey(key rune) error

	// InjectSpecialKey 注入特殊按键
	InjectSpecialKey(key platform.SpecialKey) error

	// InjectKeyWithMod 注入带修饰符的按键
	InjectKeyWithMod(key rune, mod platform.KeyModifier) error

	// InjectMouse 注入鼠标事件
	InjectMouse(x, y int, button platform.MouseButton, action platform.MouseAction) error

	// InjectResize 注入窗口调整事件
	InjectResize(width, height int) error

	// InjectString 注入字符串 (转换为按键序列)
	InjectString(text string) error

	// ProcessEvents 处理所有待处理事件
	ProcessEvents() error
}

// Snapshotter 快照接口
type Snapshotter interface {
	// Snapshot 创建快照
	Snapshot(level SnapshotLevel, tags ...string) (*Snapshot, error)

	// Restore 恢复快照
	Restore(snap *Snapshot) error

	// ListSnapshots 列出所有快照
	ListSnapshots() []*SnapshotMetadata
}

// TestSandbox 测试沙箱接口 (组合接口)
type TestSandbox interface {
	Sandbox
	EventSink
	Snapshotter

	// IsMock 是否为模拟沙箱
	IsMock() bool

	// AssertRender 断言渲染输出包含指定文本
	AssertRender(text string) error

	// AssertNotRender 断言渲染输出不包含指定文本
	AssertNotRender(text string) error

	// RenderString 获取渲染输出字符串
	RenderString() string

	// Helper 获取测试辅助器
	Helper() interface{}
}

// Renderer 渲染器接口 (由 engine 实现，避免循环依赖)
type Renderer interface {
	// Render 渲染到缓冲区
	Render(buf *paint.Buffer) error
}

// EventDispatcher 事件分发接口 (由 engine 实现)
type EventDispatcher interface {
	// Dispatch 分发事件
	Dispatch(event platform.RawInput) error
}

// EventHandler 事件处理函数类型
type EventHandler func(event platform.RawInput) error

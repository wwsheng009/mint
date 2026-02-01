// mock/sandbox.go - 模拟沙箱实现
package mock

import (
	"strings"
	"sync"
	"time"

	"github.com/wwsheng009/mint/runtime/paint"
	"github.com/wwsheng009/mint/runtime/platform"
	"github.com/wwsheng009/mint/sandbox"
	"github.com/wwsheng009/mint/sandbox/adapter"
)

// MockSandbox 模拟沙箱
type MockSandbox struct {
	mu sync.RWMutex

	lifecycle *sandbox.Lifecycle
	config    *sandbox.Config
	buffer    *paint.Buffer

	// 事件系统
	injector *sandbox.EventInjector
	queue    *BoundedQueue
	recorder *sandbox.EventRecorder

	// 快照
	snapMgr *sandbox.SnapshotManager

	// 事件处理
	eventHandler sandbox.EventHandler
}

// New 创建模拟沙箱
func New(width, height int) *MockSandbox {
	config := sandbox.MockConfig()
	config.Width = width
	config.Height = height

	ms := &MockSandbox{
		lifecycle: sandbox.NewLifecycle(),
		config:    config,
		buffer:    paint.NewBuffer(width, height),
		injector:  sandbox.NewEventInjector(sandbox.InjectAllowed),
		queue:     NewBoundedQueue(DefaultQueueConfig()),
		recorder:  sandbox.NewEventRecorder(config.Event.RecordMaxLen),
		snapMgr:   sandbox.NewSnapshotManager(config.Snapshot.MaxCount),
	}

	ms.injector.SetRecorder(ms.recorder)

	return ms
}

// ==============================================================================
// Sandbox Interface
// ==============================================================================

// Initialize 初始化沙箱
func (ms *MockSandbox) Initialize(config *sandbox.Config) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if config != nil {
		ms.config = config
		ms.buffer = paint.NewBuffer(config.Width, config.Height)
	}

	return ms.lifecycle.Transition(sandbox.StateInitialized)
}

// Run 运行沙箱主循环
func (ms *MockSandbox) Run() error {
	return ms.lifecycle.Transition(sandbox.StateRunning)
}

// Pause 暂停沙箱
func (ms *MockSandbox) Pause() error {
	return ms.lifecycle.Transition(sandbox.StatePaused)
}

// Resume 恢复沙箱
func (ms *MockSandbox) Resume() error {
	return ms.lifecycle.Transition(sandbox.StateRunning)
}

// Close 关闭沙箱
func (ms *MockSandbox) Close() error {
	return ms.lifecycle.Transition(sandbox.StateStopped)
}

// State 获取当前状态
func (ms *MockSandbox) State() sandbox.State {
	return ms.lifecycle.State()
}

// Type 获取沙箱类型
func (ms *MockSandbox) Type() sandbox.SandboxType {
	return sandbox.TypeMock
}

// Config 获取配置
func (ms *MockSandbox) Config() *sandbox.Config {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.config
}

// Buffer 获取渲染缓冲区
func (ms *MockSandbox) Buffer() *paint.Buffer {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.buffer
}

// SetBuffer 设置渲染缓冲区
func (ms *MockSandbox) SetBuffer(buf *paint.Buffer) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.buffer = buf
}

// Resize 调整缓冲区大小
func (ms *MockSandbox) Resize(width, height int) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.buffer = paint.NewBuffer(width, height)
	ms.config.Width = width
	ms.config.Height = height
}

// Size 获取当前尺寸
func (ms *MockSandbox) Size() (int, int) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.config.Width, ms.config.Height
}

// ==============================================================================
// EventSink Interface
// ==============================================================================

// SetEventHandler 设置事件处理器
func (ms *MockSandbox) SetEventHandler(handler sandbox.EventHandler) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.eventHandler = handler
	ms.injector.SetHandler(handler)
}

// Inject 注入单个事件
func (ms *MockSandbox) Inject(event platform.RawInput) error {
	if err := ms.queue.Push(event); err != nil {
		return err
	}
	return ms.injector.Inject(event)
}

// InjectKey 注入按键事件
func (ms *MockSandbox) InjectKey(key rune) error {
	event := adapter.BuildKeyEvent(key)
	return ms.Inject(event)
}

// InjectSpecialKey 注入特殊按键
func (ms *MockSandbox) InjectSpecialKey(key platform.SpecialKey) error {
	event := adapter.BuildSpecialKeyEvent(key)
	return ms.Inject(event)
}

// InjectKeyWithMod 注入带修饰符的按键
func (ms *MockSandbox) InjectKeyWithMod(key rune, mod platform.KeyModifier) error {
	event := platform.RawInput{
		Type:      platform.InputKeyPress,
		Key:       key,
		Modifiers: mod,
		Timestamp: time.Now(),
	}
	return ms.Inject(event)
}

// InjectMouse 注入鼠标事件
func (ms *MockSandbox) InjectMouse(x, y int, button platform.MouseButton, action platform.MouseAction) error {
	event := adapter.BuildMouseEvent(x, y, button, action)
	return ms.Inject(event)
}

// InjectResize 注入窗口调整事件
func (ms *MockSandbox) InjectResize(width, height int) error {
	event := adapter.BuildResizeEvent(width, height)
	return ms.Inject(event)
}

// InjectString 注入字符串 (转换为按键序列)
func (ms *MockSandbox) InjectString(text string) error {
	for _, r := range text {
		if err := ms.InjectKey(r); err != nil {
			return err
		}
	}
	return nil
}

// ProcessEvents 处理所有待处理事件
func (ms *MockSandbox) ProcessEvents() error {
	for !ms.queue.IsEmpty() {
		event, err := ms.queue.Pop()
		if err != nil {
			break
		}
		if ms.eventHandler != nil {
			if err := ms.eventHandler(event); err != nil {
				return err
			}
		}
	}
	return nil
}

// ==============================================================================
// Snapshotter Interface
// ==============================================================================

// Snapshot 创建快照
func (ms *MockSandbox) Snapshot(level sandbox.SnapshotLevel, tags ...string) (*sandbox.Snapshot, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	return ms.snapMgr.Create(level, ms.buffer, ms.recorder.Events(), nil, tags...)
}

// Restore 恢复快照
func (ms *MockSandbox) Restore(snap *sandbox.Snapshot) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if snap.Buffer != nil {
		ms.buffer = paint.NewBuffer(snap.Buffer.Width, snap.Buffer.Height)
		for y := 0; y < snap.Buffer.Height; y++ {
			copy(ms.buffer.Cells[y], snap.Buffer.Cells[y])
		}
	}

	return nil
}

// ListSnapshots 列出所有快照
func (ms *MockSandbox) ListSnapshots() []*sandbox.SnapshotMetadata {
	return ms.snapMgr.List()
}

// ==============================================================================
// TestSandbox Interface
// ==============================================================================

// IsMock 是否为模拟沙箱
func (ms *MockSandbox) IsMock() bool {
	return true
}

// AssertRender 断言渲染输出包含指定文本
func (ms *MockSandbox) AssertRender(text string) error {
	rendered := ms.RenderString()
	if !strings.Contains(rendered, text) {
		return &sandbox.AssertionError{
			Message:  "render does not contain expected text",
			Expected: text,
			Actual:   rendered,
		}
	}
	return nil
}

// AssertNotRender 断言渲染输出不包含指定文本
func (ms *MockSandbox) AssertNotRender(text string) error {
	rendered := ms.RenderString()
	if strings.Contains(rendered, text) {
		return &sandbox.AssertionError{
			Message:  "render contains unexpected text",
			Expected: "not " + text,
			Actual:   rendered,
		}
	}
	return nil
}

// RenderString 获取渲染输出字符串
func (ms *MockSandbox) RenderString() string {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if ms.buffer == nil {
		return ""
	}

	var sb strings.Builder
	for y := 0; y < ms.buffer.Height; y++ {
		for x := 0; x < ms.buffer.Width; x++ {
			cell := ms.buffer.Cells[y][x]
			if cell.IsContinuation {
				continue
			}
			if cell.Cluster == "" {
				sb.WriteRune(' ')
			} else {
				sb.WriteString(cell.Cluster)
			}
		}
		if y < ms.buffer.Height-1 {
			sb.WriteRune('\n')
		}
	}
	return sb.String()
}

// Helper 获取测试辅助器
func (ms *MockSandbox) Helper() *TestHelper {
	return NewTestHelper(ms)
}

// QueueStats 返回队列统计
func (ms *MockSandbox) QueueStats() QueueStats {
	return ms.queue.Stats()
}
